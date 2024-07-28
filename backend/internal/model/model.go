package model

import (
	"backend/internal/client"
	"encoding/csv"
	"os"
)

type model struct {
	client *client.MongoDBClient
}
type Model interface {
	ProcessCSVInBatches(filePath string, batchSize int, processBatch func([][]string) error) error
}

func NewModel(client *client.MongoDBClient) Model {
	return &model{
		client: client,
	}
}

func (m *model) ProcessCSVInBatches(filePath string, batchSize int, processBatch func([][]string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var batch [][]string
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				if len(batch) > 0 {
					if err := processBatch(batch); err != nil {
						return err
					}
				}
				break
			} else {
				return err
			}
		}
		batch = append(batch, record)
		if len(batch) >= batchSize {
			if err := processBatch(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	return nil
}
