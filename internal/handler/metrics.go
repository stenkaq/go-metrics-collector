package handler

import (
	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/service"
	"net/http"
	"strconv"
	"strings"
)

type MetricsHandler interface {
	UpdateMetric(w http.ResponseWriter, r *http.Request)
}

type metricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(s service.MetricsService) MetricsHandler {
	return &metricsHandler{service: s}
}

func (h *metricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Разрешен только POST метод", 400)
		return
	}

	rawPath := r.URL.Path
	splitPath := strings.Split(rawPath, "/")

	if len(splitPath) < 5 {
		http.Error(w, "Недостаточно параметров", 400)
		return
	}

	metricsData := map[string]string{
		"type": splitPath[2],
		"name": splitPath[3],
		"val":  splitPath[4],
	}

	if strings.TrimSpace(metricsData["name"]) == "" {
		http.Error(w, "Пустое имя метрики", http.StatusNotFound)
		return
	}

	switch metricsData["type"] {
	case models.Counter:
		parsedVal, err := strconv.ParseInt(metricsData["val"], 10, 64)
		if err != nil {
			http.Error(w, "Неверное значение метрики", http.StatusBadRequest)
			return
		}

		h.service.UpdateCounterMetricValue(service.UpdateCounterMetricValueParams{
			Type:  metricsData["type"],
			Name:  metricsData["name"],
			Value: parsedVal,
		})
	case models.Gauge:
		parsedVal, err := strconv.ParseFloat(metricsData["val"], 64)
		if err != nil {
			http.Error(w, "Неверное значение метрики", http.StatusBadRequest)
			return
		}

		h.service.UpdateGaugeMetricValue(service.UpdateGaugeMetricValueParams{
			Type:  metricsData["type"],
			Name:  metricsData["name"],
			Value: parsedVal,
		})
	default:
		http.Error(w, "Неизвестный тип метрики", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
}
