package app

import (
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/router"
	"go-metrics-collector/internal/service"

	"github.com/gin-gonic/gin"
)

func NewMetricsRouter() *gin.Engine {
	metricsRepository := repository.NewMetricsRepository()
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	return router.New(
		router.Handlers{
			Metrics: metricsHandler,
		},
	)
}
