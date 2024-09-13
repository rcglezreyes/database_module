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
	ProcessDataPredictionAssessments(database string) ([]entity.ProcessedPredictionAssessmentResult, error)
	ProcessDataVlePredictions(database string) ([]entity.ProcessedPredictionVleResult, error)
	GetScoreDistributionPredictionAssessments(database string) ([]entity.ScoreRangePredictionAssessments, error)
	GetAveragePredictedScoreByAssessmentType(database string) ([]entity.AssessmentTypeAverage, error)
	GetStudentCountByAssessmentID(database string) ([]entity.AssessmentStudentCount, error)
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

// PREDICTIONS ASSESSMENTS

func (m *mongoDBClient) ProcessDataPredictionAssessments(database string) ([]entity.ProcessedPredictionAssessmentResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Println("database: ", database)

	db := m.client.Database(database)

	fmt.Println("db: ", db)
	studentAssessmentCollection := db.Collection("studentAssessment")
	predictionAssessmentCollection := db.Collection("prediction_assessments")
	batchSize := 5000
	fmt.Println("Procesando datos de studentAssessment...")
	
	opts := options.Find().SetProjection(bson.M{"idstudent": 1, "score": 1, "idassessment": 1})

	cursor, err := studentAssessmentCollection.Find(ctx, bson.M{}, opts)
	fmt.Println("Cursor: ", cursor)
	if err != nil {
		log.Fatalf("Error al obtener los datos de studentAssessment: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Variables para almacenar los datos y procesarlos por lotes
	batch := make([]bson.M, 0, batchSize)
	var allPredictions []entity.ProcessedPredictionAssessmentResult

	// Iterar sobre los documentos
	for cursor.Next(ctx) {
		var assessment bson.M
		if err := cursor.Decode(&assessment); err != nil {
			log.Fatalf("Error al decodificar evaluación: %v", err)
		}
		log.Printf("Procesando evaluación: %v", assessment)
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

	// Comprobar si se generaron predicciones
	if len(allPredictions) == 0 {
		log.Println("No se generaron predicciones.")
	} else {
		log.Printf("Total de predicciones generadas: %d", len(allPredictions))
	}

	return allPredictions, nil
}

func (m *mongoDBClient) processAssessmentBatch(ctx context.Context, assessments []bson.M, collection *mongo.Collection) []entity.ProcessedPredictionAssessmentResult {
	batchPredictions := []interface{}{}
	var processedResults []entity.ProcessedPredictionAssessmentResult

	for _, assessment := range assessments {
		studentIDRaw, ok := assessment["idstudent"]
		if !ok {
			log.Fatalf("Error: 'idstudent' no encontrado en el documento: %v", assessment)
		}
		assessmentIDRaw, ok := assessment["idassessment"]
		if !ok {
			log.Fatalf("Error: 'idassessment' no encontrado en el documento: %v", assessment)
		}
		scoreRaw, ok := assessment["score"]
		if !ok {
			log.Fatalf("Error: 'score' no encontrado en el documento: %v", assessment)
		}

		// Convertir los campos
		studentID, err := m.convertStudentID(studentIDRaw)
		if err != nil {
			log.Fatalf("Error al convertir 'idstudent': %v", err)
		}
		assessmentID, err := m.convertAssessmentID(assessmentIDRaw)
		if err != nil {
			log.Fatalf("Error al convertir 'idassessment': %v", err)
		}
		score, err := m.convertScore(scoreRaw)
		if err != nil {
			log.Fatalf("Error al convertir 'score': %v", err)
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
			log.Fatalf("Error al insertar predicciones: %v", err)
		} else {
			log.Printf("Insertados %d documentos", len(result.InsertedIDs))
		}
	}

	return processedResults
}

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
		log.Fatalf("Tipo no reconocido para idassessment: %T", v)
		return 0, fmt.Errorf("tipo no reconocido para idassessment: %T", v)
	}
}

func (m *mongoDBClient) calculatePredictedScore(studentID int, currentScore float64) float64 {
    // Simulando una predicción basada en el historial de puntuaciones previas del estudiante
    predictedScore := currentScore * 1.05 // Aumentamos el score en un 5% como ejemplo

    log.Printf("Predicción calculada para el estudiante %d: %f (score actual: %f)", studentID, predictedScore, currentScore)
    return predictedScore
}


func (m *mongoDBClient) convertScore(scoreRaw interface{}) (float64, error) {
	switch v := scoreRaw.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		log.Fatalf("Tipo no reconocido para score: %T", v)
		return 0.0, fmt.Errorf("tipo no reconocido para score: %T", v)
	}
}

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
			log.Fatalf("Error al convertir idstudent: %v", err)
			return 0, err
		}
		return id, nil
	default:
		log.Fatalf("Tipo no reconocido para idstudent: %T", v)
		return 0, fmt.Errorf("tipo no reconocido para idstudent: %T", v)
	}
}

// Predictions VLE 

func (m *mongoDBClient) ProcessDataVlePredictions(database string) ([]entity.ProcessedPredictionVleResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Conexión a la base de datos y colección
	db := m.client.Database(database)
	studentVleCollection := db.Collection("studentVle")
	predictionsCollection := db.Collection("prediction_vle")
	batchSize := 5000 // Tamaño del batch optimizado

	// Opciones de búsqueda para limitar los campos devueltos
	opts := options.Find().SetProjection(bson.M{"id_student": 1, "resource_type": 1, "sum_click": 1})

	// Cursor para procesar los documentos
	cursor, err := studentVleCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("error al obtener los datos de studentVle: %w", err)
	}
	defer cursor.Close(ctx)

	batch := make([]bson.M, 0, batchSize)
	var wg sync.WaitGroup
	var totalProcessed int // Contador de documentos procesados
	var processedResults []entity.ProcessedPredictionVleResult // Resultados procesados

	for cursor.Next(ctx) {
		var interaction bson.M
		if err := cursor.Decode(&interaction); err != nil {
			log.Printf("error al decodificar interacción: %v", err)
			continue
		}
		batch = append(batch, interaction)

		if len(batch) == batchSize {
			wg.Add(1)
			go func(b []bson.M) {
				defer wg.Done()
				results, err := m.processAndStoreVleBatch(ctx, b, predictionsCollection)
				if err != nil {
					log.Printf("error al procesar y guardar el batch: %v", err)
				} else {
					totalProcessed += len(b)
					processedResults = append(processedResults, results...)
				}
			}(batch)
			batch = make([]bson.M, 0, batchSize)
		}
	}

	// Procesar el último lote restante
	if len(batch) > 0 {
		wg.Add(1)
		go func(b []bson.M) {
			defer wg.Done()
			results, err := m.processAndStoreVleBatch(ctx, b, predictionsCollection)
			if err != nil {
				log.Printf("error al procesar y guardar el lote final: %v", err)
			} else {
				totalProcessed += len(b)
				processedResults = append(processedResults, results...)
			}
		}(batch)
	}

	wg.Wait()

	// Verificar si se produjeron errores durante el procesamiento
	if err := cursor.Err(); err != nil {
		return processedResults, fmt.Errorf("errores encontrados durante el procesamiento de documentos: %w", err)
	}

	return processedResults, nil
}

// Función para procesar un batch de interacciones del VLE y almacenar predicciones en MongoDB
func (m *mongoDBClient) processAndStoreVleBatch(ctx context.Context, vleBatch []bson.M, predictionsCollection *mongo.Collection) ([]entity.ProcessedPredictionVleResult, error) {
	var processedResults []entity.ProcessedPredictionVleResult

	for _, interaction := range vleBatch {
		studentIDRaw, ok := interaction["id_student"]
		if !ok {
			log.Printf("Warning: 'id_student' no encontrado en el documento: %v", interaction)
			continue
		}
		resourceType, ok := interaction["resource_type"].(string)
		if !ok {
			log.Printf("Warning: 'resource_type' no encontrado o tiene un tipo no válido en el documento: %v", interaction)
			continue
		}
		clicksRaw, ok := interaction["sum_click"]
		if !ok {
			log.Printf("Warning: 'sum_click' no encontrado en el documento: %v", interaction)
			continue
		}

		// Convertir los campos
		studentID, err := m.convertStudentID(studentIDRaw)
		if err != nil {
			log.Printf("Warning: 'id_student' tiene un tipo no reconocido: %v", studentIDRaw)
			continue
		}
		clicks, err := m.convertClicks(clicksRaw)
		if err != nil {
			log.Printf("Warning: 'sum_click' tiene un tipo no reconocido: %v", clicksRaw)
			continue
		}

		// Calcular el puntaje predicho
		predictedScore := m.calculatePredictedScoreStudentVle(studentID, resourceType, clicks)

		// Almacenar el resultado en el array
		processedResults = append(processedResults, entity.ProcessedPredictionVleResult{
			StudentID:      studentID,
			PredictedScore: predictedScore,
		})
	}

	// Insertar las predicciones procesadas en la colección `prediction_vle`
	if len(processedResults) > 0 {
		insertData := make([]interface{}, len(processedResults))
		for i, result := range processedResults {
			insertData[i] = result
		}

		_, err := predictionsCollection.InsertMany(ctx, insertData)
		if err != nil {
			return nil, fmt.Errorf("error al insertar las predicciones en MongoDB: %w", err)
		}

		log.Printf("Insertadas %d predicciones en la colección", len(processedResults))
	}

	return processedResults, nil
}

// Función para calcular el puntaje predicho basado en las interacciones con el VLE
func (m *mongoDBClient) calculatePredictedScoreStudentVle(studentID int, resourceType string, clicks int) float64 {
	// Pesos basados en el tipo de interacción
	interactionWeights := map[string]float64{
		"forum":       1.2, // Mayor peso para interacciones con foros
		"quiz":        1.5, // Quiz tiene mayor relevancia para el éxito
		"resource":    1.0, // Interacciones normales
		"assignment":  1.8, // Asignaciones tienen gran importancia
	}

	weight, exists := interactionWeights[resourceType]
	if !exists {
		weight = 1.0 // Si no se encuentra el tipo de recurso, usar peso por defecto
	}

	// Calcular el puntaje predicho (ejemplo básico)
	predictedScore := float64(clicks) * weight
	log.Printf("Predicción calculada para el estudiante %d: %f (tipo de recurso: %s, clics: %d)", studentID, predictedScore, resourceType, clicks)
	return predictedScore
}

// Función para convertir el número de clics
func (m *mongoDBClient) convertClicks(clicksRaw interface{}) (int, error) {
	switch v := clicksRaw.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("tipo no reconocido para sum_click: %T", v)
	}
}

// Charts

func (m *mongoDBClient) GetScoreDistributionPredictionAssessments(database string) ([]entity.ScoreRangePredictionAssessments, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Conexión a la base de datos y colección
	db := m.client.Database(database)
	predictionCollection := db.Collection("prediction_assessments")

	// Definir los rangos de puntuación
	ranges := []struct {
		Lower float64
		Upper float64
		Label string
	}{
		{0, 59, "Menos de 60"},
		{60, 69, "60 a 70"},
		{70, 79, "70 a 80"},
		{80, 89, "80 a 90"},
		{90, 100, "Más de 90"},
	}

	// Crear un mapa para almacenar los conteos
	scoreCounts := make(map[string]int)
	for _, r := range ranges {
		scoreCounts[r.Label] = 0
	}

	// Obtener todos los documentos de la colección
	cursor, err := predictionCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error al obtener los documentos: %w", err)
	}
	defer cursor.Close(ctx)

	// Procesar los documentos y clasificarlos en los rangos correspondientes
	for cursor.Next(ctx) {
		var result struct {
			PredictedScore float64 `bson:"predicted_score"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Printf("error al decodificar documento: %v", err)
			continue
		}

		// Clasificar el resultado en el rango correspondiente
		for _, r := range ranges {
			if result.PredictedScore >= r.Lower && result.PredictedScore <= r.Upper {
				scoreCounts[r.Label]++
				break
			}
		}
	}

	// Verificar si se produjeron errores durante la iteración del cursor
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("error durante la iteración del cursor: %w", err)
	}

	// Completar los resultados con rangos que no tengan estudiantes
	var results []entity.ScoreRangePredictionAssessments
	for _, r := range ranges {
		results = append(results, entity.ScoreRangePredictionAssessments{
			Range:        r.Label,
			StudentCount: scoreCounts[r.Label],
		})
	}

	return results, nil
}

// Average by types

// Función principal para obtener el promedio de puntajes predichos por tipo de evaluación
func (m *mongoDBClient) GetAveragePredictedScoreByAssessmentType(database string) ([]entity.AssessmentTypeAverage, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    db := m.client.Database(database)
    predictionCollection := db.Collection("prediction_assessments")
    assessmentCollection := db.Collection("assessments")

    // Crear índices si no existen
    err := createIndexes(ctx, predictionCollection, assessmentCollection)
    if err != nil {
        return nil, fmt.Errorf("error al crear índices: %w", err)
    }

    // Obtener los tipos de evaluación
    assessmentTypes, err := getDistinctAssessmentTypes(ctx, assessmentCollection)
    if err != nil {
        return nil, fmt.Errorf("error al obtener tipos de evaluación: %w", err)
    }

    log.Printf("Tipos de evaluación obtenidos: %v", assessmentTypes)

    // Variables para goroutines y sincronización
    var wg sync.WaitGroup
    var mu sync.Mutex
    results := []entity.AssessmentTypeAverage{}
    var aggregateErr error

    // Tamaño del batch
    const batchSize = 5000

    processBatch := func(types []string) {
        defer wg.Done()

        // Obtener IDs de evaluaciones por tipo
        assessmentIDs := getAssessmentIDsForTypes(ctx, types, assessmentCollection)
        if len(assessmentIDs) == 0 {
            return
        }

        // Pipeline optimizado
        pipeline := mongo.Pipeline{
            // Filtrar por assessment_ids específicos
            bson.D{{"$match", bson.M{"assessment_id": bson.M{"$in": assessmentIDs}}}},
            // Unir con la colección de assessments
            bson.D{{"$lookup", bson.M{
                "from":         "assessments",
                "localField":   "assessment_id",
                "foreignField": "idassessment",
                "as":           "assessment_info",
            }}},
            // Descomponer la información del assessment
            bson.D{{"$unwind", "$assessment_info"}},
            // Agrupar por tipo de evaluación y calcular el promedio
            bson.D{{"$group", bson.M{
                "_id":           "$assessment_info.assessmenttype",
                "average_score": bson.M{"$avg": "$predicted_score"},
            }}},
            // Ordenar por promedio
            bson.D{{"$sort", bson.M{"average_score": -1}}},
            // Proyectar los campos requeridos
            bson.D{{"$project", bson.M{
                "_id":            0,
                "assessment_type": "$_id",
                "average_score":  "$average_score",
            }}},
        }

        cursorOpts := options.Aggregate().SetBatchSize(batchSize)
        cursor, err := predictionCollection.Aggregate(ctx, pipeline, cursorOpts)
        if err != nil {
            mu.Lock()
            if aggregateErr == nil {
                aggregateErr = fmt.Errorf("error al ejecutar la agregación: %w", err)
            }
            mu.Unlock()
            return
        }
        defer cursor.Close(ctx)

        var localResults []entity.AssessmentTypeAverage
        for cursor.Next(ctx) {
            var result entity.AssessmentTypeAverage
            if err := cursor.Decode(&result); err != nil {
                log.Printf("Error al decodificar resultado: %v", err)
                continue
            }
            localResults = append(localResults, result)
        }

        if err := cursor.Err(); err != nil {
            mu.Lock()
            if aggregateErr == nil {
                aggregateErr = fmt.Errorf("error durante la iteración del cursor: %w", err)
            }
            mu.Unlock()
            return
        }

        // Agregar los resultados locales a los globales
        mu.Lock()
        results = append(results, localResults...)
        mu.Unlock()
    }

    // Procesar en batches
    for i := 0; i < len(assessmentTypes); i += batchSize {
        end := i + batchSize
        if end > len(assessmentTypes) {
            end = len(assessmentTypes)
        }
        batch := assessmentTypes[i:end]

        wg.Add(1)
        go processBatch(batch)
    }

    wg.Wait()

    if aggregateErr != nil {
        return nil, aggregateErr
    }

    return results, nil
}


// Función para obtener IDs de evaluación para un conjunto de tipos
func getAssessmentIDsForTypes(ctx context.Context, types []string, collection *mongo.Collection) []interface{} {
    var ids []interface{}
    cursor, err := collection.Find(ctx, bson.D{{"assessmenttype", bson.D{{"$in", types}}}}, options.Find().SetProjection(bson.D{{"idassessment", 1}}))
    if err != nil {
        log.Printf("error al obtener IDs de evaluación para tipos: %v", err)
        return ids
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var result struct{ IDAssessment interface{} `bson:"idassessment"` }
        if err := cursor.Decode(&result); err != nil {
            log.Printf("error al decodificar ID de evaluación: %v", err)
            continue
        }
        ids = append(ids, result.IDAssessment)
    }

    if err := cursor.Err(); err != nil {
        log.Printf("error durante la iteración del cursor para IDs: %v", err)
    }

    return ids
}


// Función para obtener tipos de evaluación distintos
func getDistinctAssessmentTypes(ctx context.Context, collection *mongo.Collection) ([]string, error) {
    cursor, err := collection.Distinct(ctx, "assessmenttype", bson.D{})
    if err != nil {
        return nil, fmt.Errorf("error al obtener tipos de evaluación: %w", err)
    }

    // Convertir cursor a un slice de strings
    var types []string
    for _, item := range cursor {
        if str, ok := item.(string); ok {
            types = append(types, str)
        } else {
            return nil, fmt.Errorf("tipo de evaluación inesperado: %v", item)
        }
    }

    return types, nil
}

// Función para crear índices en las colecciones
func createIndexes(ctx context.Context, predictionCollection, assessmentCollection *mongo.Collection) error {
    // Índice en prediction_assessments
    _, err := predictionCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{"assessment_id", 1}},
        Options: options.Index().SetName("index_assessment_id"),
    })
    if err != nil {
        return fmt.Errorf("error al crear índice en prediction_assessments: %w", err)
    }
    log.Println("Índice creado en prediction_assessments para 'assessment_id'")

    // Índice en assessments
    _, err = assessmentCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{"assessmenttype", 1}},
        Options: options.Index().SetName("index_assessmenttype"),
    })
    if err != nil {
        return fmt.Errorf("error al crear índice en assessments: %w", err)
    }
    log.Println("Índice creado en assessments para 'assessmenttype'")

    return nil
}


// Students by Assessment

func (m *mongoDBClient) GetStudentCountByAssessmentID(database string) ([]entity.AssessmentStudentCount, error) {
	// Contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	numWorkers := 10

	// Conexión a la base de datos y colección
	collection := m.client.Database(database).Collection("prediction_assessments")

	// Crear índices si no existen
	err := createIndexesCounts(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("error al crear índices: %w", err)
	}

	// Canal para recibir los resultados de las goroutines
	resultChan := make(chan entity.AssessmentStudentCount, 1000)
	var wg sync.WaitGroup

	// Pipeline de agregación
	pipeline := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", "$assessment_id"},
			{"student_count", bson.D{{"$addToSet", "$student_id"}}},
		}}},
		{{"$project", bson.D{
			{"_id", 1},
			{"student_count", bson.D{{"$size", "$student_count"}}},
		}}},
	}

	// Ejecutar la agregación con batch
	cursor, err := collection.Aggregate(ctx, pipeline, options.Aggregate().SetBatchSize(5000))
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar agregación: %w", err)
	}
	defer cursor.Close(ctx)

	// Goroutine para manejar los resultados en paralelo
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processResults(ctx, cursor, resultChan)
		}()
	}

	// Esperar a que todas las goroutines terminen
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Recopilar resultados
	var results []entity.AssessmentStudentCount
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

// Función para procesar resultados de la agregación
func processResults(ctx context.Context, cursor *mongo.Cursor, resultChan chan<- entity.AssessmentStudentCount) {
	for cursor.Next(ctx) {
		var result entity.AssessmentStudentCount
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Error al decodificar el resultado: %v", err)
			continue
		}
		resultChan <- result
	}
	if err := cursor.Err(); err != nil {
		log.Printf("Error en el cursor: %v", err)
	}
}

// Crear índice para optimizar las consultas
func createIndexesCounts(ctx context.Context, collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"assessment_id", 1}, {"student_id", 1}},
		Options: options.Index().SetName("index_assessment_student"),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("error al crear índice: %w", err)
	}
	log.Println("Índice creado en prediction_assessments para 'assessment_id' y 'student_id'")
	return nil
}



