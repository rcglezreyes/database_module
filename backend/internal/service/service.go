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
	GetFiles() ([]*entity.FileInfo, error)
	GetData(collection string) ([]interface{}, error)
	GetAllCountData(collections []string) (map[string]int64, error)
	ProcessDataPredictionAssessments() ([]entity.ProcessedPredictionAssessmentResult, error)
	ProcessDataVlePredictions() ([]entity.ProcessedPredictionVleResult, error)
	GetScoreDistributionPredictionAssessments() ([]entity.ScoreRangePredictionAssessments, error)
	GetAveragePredictedScoreByAssessmentType() ([]entity.AssessmentTypeAverage, error)
	GetStudentCountByAssessmentID() ([]entity.AssessmentStudentCount, error)
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
func (s *service) GetFiles() ([]*entity.FileInfo, error) {
	return s.model.GetFiles()
}
func (s *service) GetData(collection string) ([]interface{}, error) {
	return s.model.GetData(collection)
}
func (s *service) GetAllCountData(collections []string) (map[string]int64, error) {
	return s.model.GetAllCountData(collections)
}
func (s *service) ProcessDataPredictionAssessments() ([]entity.ProcessedPredictionAssessmentResult, error) {
	return s.model.ProcessDataPredictionAssessments()
}
func (s *service) ProcessDataVlePredictions() ([]entity.ProcessedPredictionVleResult, error) {
	return s.model.ProcessDataVlePredictions()
}
func (s *service) GetScoreDistributionPredictionAssessments() ([]entity.ScoreRangePredictionAssessments, error) {
	return s.model.GetScoreDistributionPredictionAssessments()
}
func (s *service) GetAveragePredictedScoreByAssessmentType() ([]entity.AssessmentTypeAverage, error) {
	return s.model.GetAveragePredictedScoreByAssessmentType()
}

func (s *service) GetStudentCountByAssessmentID() ([]entity.AssessmentStudentCount, error) {
	return s.model.GetStudentCountByAssessmentID()
}
