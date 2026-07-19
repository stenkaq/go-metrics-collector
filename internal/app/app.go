package app

import (
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/router"
	"go-metrics-collector/internal/service"
	"net/http"
)

func New() http.Handler {
	metricsRepository := repository.NewMetricsRepository()
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	return router.New(
		router.Handlers{
			Metrics: metricsHandler,
		},
	)
}
