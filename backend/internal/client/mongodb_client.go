package client

import (
	"backend/internal/config"
	"backend/internal/entity"
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDBClient struct {
	client  *mongo.Client
	URI     string
	loggers *entity.Loggers
}

type MongoDBClient interface {
	Connect() error
	Disconnect() error
	InsertOne(database, collection string, document interface{}) (*mongo.InsertOneResult, error)
	InsertMany(database, collection string, documents []interface{}) (*mongo.InsertManyResult, error)
	BatchInsert(database, collection string, documents []interface{}, batchSize int) error
	GetData(database, collection string) ([]interface{}, error)
	GetAllCountData(database string, colls []string) (map[string]int64, error)
	ProcessDataPredictionAssessments(database string) (map[string][]interface{}, error)
}

func NewMongoDBClient(loggers *entity.Loggers) MongoDBClient {
	dbcredentials, _, err := config.DBCredentials()
	if err != nil {
		loggers.ErrorLogger.Printf("Error al obtener las credenciales de la base de datos: %v", err)
		return nil
	}
	return &mongoDBClient{
		URI:     dbcredentials.URI,
		loggers: loggers,
	}
}

func (m *mongoDBClient) ProcessDataPredictionAssessments(database string) (map[string][]interface{}, error) {
	// Implement the logic for processing data prediction assessments here
	return nil, nil
}

func (m *mongoDBClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	clientOptions := options.Client().
		ApplyURI(m.URI).
		SetSocketTimeout(200 * time.Second).
		SetConnectTimeout(60 * time.Second).
		SetMaxConnIdleTime(5 * time.Minute).
		SetMaxPoolSize(100)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		m.loggers.ErrorLogger.Printf("Error al conectar con MongoDB: %v", err)
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		m.loggers.ErrorLogger.Printf("Error al verificar la conexión con MongoDB: %v", err)
		return err
	}

	m.client = client
	m.loggers.InfoLogger.Println("Conectado a MongoDB")
	return nil
}

func (m *mongoDBClient) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		m.loggers.ErrorLogger.Printf("Error al desconectar de MongoDB: %v", err)
		return err
	}

	m.loggers.InfoLogger.Println("Desconectado de MongoDB")
	return nil
}

func (m *mongoDBClient) InsertOne(database, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	col := m.client.Database(database).Collection(collection)
	result, err := col.InsertOne(ctx, document)
	if err != nil {
		m.loggers.ErrorLogger.Printf("Error al insertar el documento: %v", err)
		return nil, err
	}

	m.loggers.InfoLogger.Printf("Documento insertado con ID: %v", result.InsertedID)
	return result, nil
}

func (m *mongoDBClient) InsertMany(database, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	col := m.client.Database(database).Collection(collection)
	result, err := col.InsertMany(ctx, documents)
	if err != nil {
		m.loggers.ErrorLogger.Fatalf("Error al insertar documentos: %v", err)
		return nil, err
	}

	return result, nil
}

func (m *mongoDBClient) BatchInsert(database, collection string, documents []interface{}, batchSize int) error {
	var wg sync.WaitGroup
	docCount := len(documents)
	batches := (docCount + batchSize - 1) / batchSize

	maxGoroutines := 10
	guard := make(chan struct{}, maxGoroutines)

	for i := 0; i < batches; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > docCount {
			end = docCount
		}

		guard <- struct{}{}
		wg.Add(1)

		go func(batch []interface{}) {
			defer wg.Done()
			defer func() { <-guard }()

			if _, err := m.InsertMany(database, collection, batch); err != nil {
				m.loggers.ErrorLogger.Printf("Error al insertar lote: %v", err)
			}
		}(documents[start:end])
	}

	wg.Wait()
	return nil
}
func (m *mongoDBClient) GetAllCountData(database string, colls []string) (map[string]int64, error) {
	data := make(map[string]int64)
	var mu sync.Mutex // Mutex para evitar condiciones de carrera
	var wg sync.WaitGroup
	var firstErr error

	for _, coll := range colls {
		wg.Add(1)
		go func(coll string) {
			defer wg.Done()
			count, err := m.GetCount(database, coll)
			if err != nil {
				m.loggers.ErrorLogger.Printf("Error al obtener los datos de la colección %v: %v", coll, err)
				if firstErr == nil {
					firstErr = err
				}
				return
			}

			mu.Lock()
			data[coll] = count
			mu.Unlock()
		}(coll)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	return data, nil
}

func (m *mongoDBClient) GetCount(database, collection string) (int64, error) {
	var count int64
	var err error

	for retries := 0; retries < 3; retries++ { // Reintenta hasta 3 veces
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Aumenta el tiempo de espera
		defer cancel()

		col := m.client.Database(database).Collection(collection)
		count, err = col.EstimatedDocumentCount(ctx)
		if err == nil {
			return count, nil
		}
		m.loggers.ErrorLogger.Printf("Error al obtener el número de documentos, reintentando... %v", err)
		time.Sleep(2 * time.Second) // Espera antes de reintentar
	}

	return 0, err // Retorna el error si fallan todos los reintentos
}

func (m *mongoDBClient) GetData(database, collection string) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err := m.collectionExists(database, collection)
	if err != nil {
		m.loggers.ErrorLogger.Printf("Error coleccion inexistente: %v", err)
		return nil, err
	}

	col := m.client.Database(database).Collection(collection)
	opts := options.Find().SetBatchSize(int32(viper.GetInt(config.BatchSize))) // Establece el tamaño del lote
	cursor, err := col.Find(ctx, bson.D{}, opts)
	if err != nil {
		m.loggers.ErrorLogger.Printf("Error al obtener los datos: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []interface{}
	for cursor.Next(ctx) {
		var result interface{}
		if err := cursor.Decode(&result); err != nil {
			m.loggers.ErrorLogger.Printf("Error al decodificar el documento: %v", err)
			return nil, err
		}
		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		m.loggers.ErrorLogger.Printf("Error al procesar los documentos: %v", err)
		return nil, err
	}

	return results, nil
}

func (m *mongoDBClient) collectionExists(database, collection string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := m.client.Database(database)
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return false, err
	}

	for _, coll := range collections {
		if coll == collection {
			return true, nil
		}
	}

	return false, nil
}

func (m *mongoDBClient) ProcessStudentAssessmentsPredictions(database string) ([]entity.ProcessedPredictionAssessmentResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	db := m.client.Database(database)
	studentAssessmentCollection := db.Collection("studentAssessment")
	predictionAssessmentCollection := db.Collection("prediction_assessments")
	batchSize := 500

	// Configurar consulta con proyección para obtener solo los campos necesarios
	opts := options.Find().SetProjection(bson.M{"idstudent": 1, "score": 1, "idassessment": 1})

	cursor, err := studentAssessmentCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		log.Printf("Error al obtener los datos de studentAssessment: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Verificar si hay datos en el cursor
	if !cursor.Next(ctx) {
		log.Println("No se encontraron documentos en la colección studentAssessment")
		return nil, nil
	}

	// Variables para almacenar los datos y procesarlos por lotes
	batch := make([]bson.M, 0, batchSize)
	var allPredictions []entity.ProcessedPredictionAssessmentResult // Almacenar todas las predicciones para devolverlas al frontend

	// Volver a procesar todos los documentos
	cursor.Rewind() // Volver al inicio del cursor
	for cursor.Next(ctx) {
		var assessment bson.M
		if err := cursor.Decode(&assessment); err != nil {
			log.Printf("Error al decodificar evaluación: %v", err)
			continue
		}
		log.Printf("Procesando evaluación: %v", assessment) // Verifica el contenido del documento
		batch = append(batch, assessment)

		// Procesar el batch cuando alcance el tamaño adecuado
		if len(batch) == batchSize {
			predictions := m.processAssessmentBatch(ctx, batch, predictionAssessmentCollection)
			allPredictions = append(allPredictions, predictions...)
			batch = batch[:0] // Reiniciar el batch
		}
	}

	// Procesar cualquier lote restante
	if len(batch) > 0 {
		predictions := m.processAssessmentBatch(ctx, batch, predictionAssessmentCollection)
		allPredictions = append(allPredictions, predictions...)
	}

	// Verificar si las predicciones fueron generadas
	if len(allPredictions) == 0 {
		log.Println("No se generaron predicciones.")
	} else {
		log.Printf("Total de predicciones generadas: %d", len(allPredictions))
	}

	// Devolver las predicciones generadas para validarlas en el frontend
	return allPredictions, nil
}

// Función para procesar un batch de predicciones basado en los datos de studentAssessment
func (m *mongoDBClient) processAssessmentBatch(ctx context.Context, assessments []bson.M, collection *mongo.Collection) []entity.ProcessedPredictionAssessmentResult {
	batchPredictions := []interface{}{}
	var processedResults []entity.ProcessedPredictionAssessmentResult

	for _, assessment := range assessments {
		studentIDRaw, ok := assessment["idstudent"]
		if !ok {
			log.Printf("Warning: 'idstudent' no encontrado en el documento: %v", assessment)
			continue
		}
		assessmentIDRaw, ok := assessment["idassessment"]
		if !ok {
			log.Printf("Warning: 'idassessment' no encontrado en el documento: %v", assessment)
			continue
		}
		scoreRaw, ok := assessment["score"]
		if !ok {
			log.Printf("Warning: 'score' no encontrado en el documento: %v", assessment)
			continue
		}

		// Convertir los campos
		studentID, err := m.convertStudentID(studentIDRaw)
		if err != nil {
			log.Printf("Warning: 'idstudent' tiene un tipo no reconocido: %v", studentIDRaw)
			continue
		}
		assessmentID, err := m.convertAssessmentID(assessmentIDRaw)
		if err != nil {
			log.Printf("Warning: 'idassessment' tiene un tipo no reconocido: %v", assessmentIDRaw)
			continue
		}
		score, err := m.convertScore(scoreRaw)
		if err != nil {
			log.Printf("Warning: 'score' tiene un tipo no reconocido: %v", scoreRaw)
			continue
		}

		// Crear predicción basada en el historial del estudiante
		predictedScore := m.calculatePredictedScore(studentID, score)

		prediction := &entity.PredictionAssessment{
			StudentID:      studentID,
			AssessmentID:   assessmentID,
			PredictedScore: predictedScore,
			PredictionDate: time.Now(),
		}

		batchPredictions = append(batchPredictions, prediction)

		// Almacenar las predicciones para devolverlas al frontend
		processedResults = append(processedResults, entity.ProcessedPredictionAssessmentResult{
			StudentID:      studentID,
			AssessmentID:   assessmentID,
			PredictedScore: predictedScore,
		})
	}

	// Inserta las predicciones en la colección
	if len(batchPredictions) > 0 {
		result, err := collection.InsertMany(ctx, batchPredictions)
		if err != nil {
			log.Printf("Error al insertar predicciones: %v", err)
		} else {
			log.Printf("Insertados %d documentos", len(result.InsertedIDs))
		}
	}

	return processedResults
}

// Función para calcular el puntaje predicho basado en los puntajes anteriores
func (m *mongoDBClient) calculatePredictedScore(studentID int, currentScore float64) float64 {
	// Aquí puedes personalizar el modelo predictivo
	// Un modelo simple basado en el puntaje actual
	return currentScore * 1.05 // Ejemplo de una pequeña mejora como predicción
}

// Función para convertir el ID de la evaluación
func (m *mongoDBClient) convertAssessmentID(assessmentIDRaw interface{}) (int, error) {
	switch v := assessmentIDRaw.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("tipo no reconocido para idassessment: %T", v)
	}
}

// Función para convertir el puntaje
func (m *mongoDBClient) convertScore(scoreRaw interface{}) (float64, error) {
	switch v := scoreRaw.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		return 0.0, fmt.Errorf("tipo no reconocido para score: %T", v)
	}
}

// Función para convertir el ID del estudiante a un entero
func (m *mongoDBClient) convertStudentID(studentIDRaw interface{}) (int, error) {
	switch v := studentIDRaw.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string: // En caso de que el ID del estudiante sea una cadena, intenta convertirlo
		id, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("error al convertir idstudent: %v", err)
		}
		return id, nil
	default:
		return 0, fmt.Errorf("tipo no reconocido para idstudent: %T", v)
	}
}
