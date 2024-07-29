package main

import (
	"backend/internal/app"
	"backend/internal/client"
	"backend/internal/config"
	"backend/internal/entity"
	"backend/internal/model"
	"backend/internal/service"
	"fmt"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func main() {
	loggers := &entity.Loggers{
		InfoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
	if err, isConfigurable := config.ConfigEnv(); !isConfigurable {
		loggers.ErrorLogger.Fatalf(err.Error())
		os.Exit(1)
	} else {
		e := echo.New()
		// Middleware
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))
		e.Use(middleware.Recover())
		e.Use(middleware.CORS())
		//Client
		client := client.NewMongoDBClient(loggers)
		// Conectar a MongoDB
		if err := client.Connect(); err != nil {
			loggers.ErrorLogger.Fatalf("Error al conectar a MongoDB: %v", err)
		}
		defer client.Disconnect()
		//Model
		model := model.NewModel(client, loggers)
		//Service
		service := service.NewService(model, loggers)
		//app
		application := app.NewApp(service)
		application.ConfigRoutes(e)
		//Starting server
		server := fmt.Sprintf(":%v", viper.GetString(config.APP_PORT))
		e.Logger.Fatal(e.Start(server))
	}
}
