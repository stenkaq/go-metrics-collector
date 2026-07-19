package main

import (
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/service"
	"net/http"
	"path"
)

func main() {
	mux := http.NewServeMux()

	metricsRepository := repository.NewMetricsRepository()
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	mux.HandleFunc("/update/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsHandler.UpdateMetric(w, r)
	}))

	muxWithMiddleware := rejectUncleanPath(mux)
	http.ListenAndServe("localhost:8080", muxWithMiddleware)
}

func rejectUncleanPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := path.Clean(r.URL.Path)

		if cleanPath != r.URL.Path && cleanPath+"/" != r.URL.Path {
			http.Error(w, "Неверный путь", http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
