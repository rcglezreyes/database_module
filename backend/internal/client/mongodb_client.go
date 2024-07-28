package client

import (
	"backend/internal/config"
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type mongoDBClient struct {
	client *mongo.Client
	URI    string
}

type MongoDBClient interface {
	Connect() error
	Disconnect() error
	InsertOne(database, collection string, document interface{}) (*mongo.InsertOneResult, error)
	InsertMany(database, collection string, documents []interface{}) (*mongo.InsertManyResult, error)
	BatchInsert(database, collection string, documents []interface{}, batchSize int) error
}

func NewMongoDBClient() MongoDBClient {
	dbcredentials, err := config.DBCredentials()
	if err != nil {
		return &mongoDBClient{
			URI: dbcredentials.URI,
		}
	}
	return nil
}

func (m *mongoDBClient) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(m.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		errorLogger.Printf("Error al conectar con MongoDB: %v", err)
		return err
	}

	// Verifica la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		errorLogger.Printf("Error al verificar la conexión con MongoDB: %v", err)
		return err
	}

	m.client = client
	infoLogger.Println("Conectado a MongoDB")
	return nil
}

func (m *mongoDBClient) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		errorLogger.Printf("Error al desconectar de MongoDB: %v", err)
		return err
	}

	infoLogger.Println("Desconectado de MongoDB")
	return nil
}

func (m *mongoDBClient) InsertOne(database, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := m.client.Database(database).Collection(collection)
	result, err := col.InsertOne(ctx, document)
	if err != nil {
		errorLogger.Printf("Error al insertar el documento: %v", err)
		return nil, err
	}

	infoLogger.Printf("Documento insertado con ID: %v", result.InsertedID)
	return result, nil
}

func (m *mongoDBClient) InsertMany(database, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := m.client.Database(database).Collection(collection)
	result, err := col.InsertMany(ctx, documents)
	if err != nil {
		errorLogger.Printf("Error al insertar documentos: %v", err)
		return nil, err
	}

	infoLogger.Printf("Documentos insertados con IDs: %v", result.InsertedIDs)
	return result, nil
}

func (m *mongoDBClient) BatchInsert(database, collection string, documents []interface{}, batchSize int) error {
	var wg sync.WaitGroup
	docCount := len(documents)
	batches := (docCount + batchSize - 1) / batchSize // número de lotes necesarios

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
				errorLogger.Printf("Error al insertar lote: %v", err)
			}
		}(documents[start:end])
	}

	wg.Wait()
	return nil
}
