package app

import (
	"backend/internal/entity"
	"backend/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type app struct {
	service service.Service
}
type App interface {
	ConfigRoutes(*echo.Echo)
	LoadBatchData(echo.Context) error
	DownloadData(echo.Context) error
	GetData(c echo.Context) error
	GetAllCountData(c echo.Context) error
	ProcessDataPredictionAssessments(c echo.Context) error
}

func NewApp(service service.Service) App {
	return &app{
		service: service,
	}
}
func (a *app) ConfigRoutes(e *echo.Echo) {
	e.GET("/api_backend/load_data", a.LoadBatchData)
	e.GET("/api_backend/download_data", a.DownloadData)
	e.GET("/api_backend/get_files", a.GetFiles)
	e.GET("/api_backend/get_data/:collection", a.GetData)
	e.POST("/api_backend/get_all_data", a.GetAllCountData)
	e.POST("/api_backend/process_data_prediction_assessments", a.ProcessDataPredictionAssessments)
}
func (a *app) LoadBatchData(c echo.Context) error {
	err := a.service.LoadBatchData()
	if err != nil {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Load Data)",
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, entity.ResponseGeneric{
		Status:  "Success",
		Message: "Data loaded successfully",
	})
}

func (a *app) DownloadData(c echo.Context) error {
	err := a.service.DownloadData()
	if err != nil {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Download Data)",
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, entity.ResponseGeneric{
		Status:  "Success",
		Message: "Data downloaded successfully",
	})
}
func (a *app) GetFiles(c echo.Context) error {
	files, err := a.service.GetFiles()
	if err != nil {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Download Data)",
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, files)
}
func (a *app) GetData(c echo.Context) error {
	collectionName := c.Param("collection")

	if collectionName == "" {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Download Data)",
			Message: "Missing collection_name or filter parameter",
		})
	}
	data, err := a.service.GetData(collectionName)
	if err != nil {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Download Data)",
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, data)
}

func (a *app) GetAllCountData(c echo.Context) error {
	reqBody := new(entity.CollectionsRequest)
	if err := c.Bind(reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Invalid request payload)",
			Message: err.Error(),
		})
	}
	data, err := a.service.GetAllCountData(reqBody.Collections)
	if err != nil {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Getting Data)",
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, data)
}
func (a *app) ProcessDataPredictionAssessments(c echo.Context) error {
	data, err := a.service.ProcessDataPredictionAssessments()
	if err != nil {
		return c.JSON(http.StatusBadRequest, entity.ResponseGeneric{
			Status:  "Failed (Getting Data)",
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, data)
}
