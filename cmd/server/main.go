package main

import (
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/service"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	metricsRepository := repository.NewMetricsRepository()
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	mux.HandleFunc("/update/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsHandler.UpdateMetric(w, r)
	}))

	http.ListenAndServe("localhost:8080", mux)
}
