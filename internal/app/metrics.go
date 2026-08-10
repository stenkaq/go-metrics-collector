package app

import (
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/router"
	"go-metrics-collector/internal/service"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewMetricsRouter() *gin.Engine {
	metricsRepository := repository.NewMetricsRepository()
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Print("Не удалось создать zap логер")
		panic(err)
	}
	defer logger.Sync()

	return router.New(
		router.Handlers{
			Metrics: metricsHandler,
		},
		logger,
	)
}
