package client

import (
	"backend/internal/config"
	"backend/internal/entity"
	"context"
	"sync"
	"time"

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
}

func NewMongoDBClient(loggers *entity.Loggers) MongoDBClient {
	dbcredentials, _, err := config.DBCredentials()
	if err != nil {
		loggers.ErrorLogger.Fatalf("Error al obtener las credenciales de la base de datos: %v", err)
		return nil
	}
	return &mongoDBClient{
		URI:     dbcredentials.URI,
		loggers: loggers,
	}
}

func (m *mongoDBClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(m.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		m.loggers.ErrorLogger.Fatalf("Error al conectar con MongoDB: %v", err)
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		m.loggers.ErrorLogger.Fatalf("Error al verificar la conexión con MongoDB: %v", err)
		return err
	}

	m.client = client
	m.loggers.InfoLogger.Println("Conectado a MongoDB")
	return nil
}

func (m *mongoDBClient) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		m.loggers.ErrorLogger.Printf("Error al desconectar de MongoDB: %v", err)
		return err
	}

	m.loggers.InfoLogger.Println("Desconectado de MongoDB")
	return nil
}

func (m *mongoDBClient) InsertOne(database, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := m.client.Database(database).Collection(collection)
	result, err := col.InsertOne(ctx, document)
	if err != nil {
		m.loggers.ErrorLogger.Fatalf("Error al insertar el documento: %v", err)
		return nil, err
	}

	m.loggers.InfoLogger.Printf("Documento insertado con ID: %v", result.InsertedID)
	return result, nil
}

func (m *mongoDBClient) InsertMany(database, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	for i := 0; i < batches; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > docCount {
			end = docCount
		}

		wg.Add(1)
		go func(batch []interface{}) {
			defer wg.Done()
			if _, err := m.InsertMany(database, collection, batch); err != nil {
				m.loggers.ErrorLogger.Printf("Error al insertar lote: %v", err)
			}
		}(documents[start:end])
	}

	wg.Wait()
	return nil
}
