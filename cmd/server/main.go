package main

import (
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/router"
	"go-metrics-collector/internal/service"
	"net/http"
)

func main() {
	metricsRepository := repository.NewMetricsRepository()
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	httpHandler := router.New(router.Handlers{
		Metrics: metricsHandler,
	})

	http.ListenAndServe("localhost:8080", httpHandler)
}
