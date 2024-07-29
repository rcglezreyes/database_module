package service

import (
	"backend/internal/entity"
	"backend/internal/model"
)

type service struct {
	model   model.Model
	loggers *entity.Loggers
}
type Service interface {
	LoadBatchData() error
	DownloadData() error
}

func NewService(model model.Model, loggers *entity.Loggers) Service {
	return &service{
		model:   model,
		loggers: loggers,
	}
}
func (s *service) LoadBatchData() error {
	return s.model.LoadBatchData()
}
func (s *service) DownloadData() error {
	return s.model.DownloadData()
}
