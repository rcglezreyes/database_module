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
}

func NewApp(service service.Service) App {
	return &app{
		service: service,
	}
}
func (a *app) ConfigRoutes(e *echo.Echo) {
	e.GET("/load_data", a.LoadBatchData)
	e.GET("/download_data", a.DownloadData)
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
