package router

import (
	"go-metrics-collector/internal/handler"
	"net/http"
	"path"
)

type Handlers struct {
	Metrics handler.MetricsHandler
}

func New(handlers Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /update/", handlers.Metrics.UpdateMetric)

	return rejectUncleanPath(mux)
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
