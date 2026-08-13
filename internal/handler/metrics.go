package handler

import (
	"fmt"
	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/service"
	"html"
	"net/http"
	"sort"
)

type MetricsHandler interface {
	Update(w http.ResponseWriter, params MetricsUpdateParams)
	GetMetric(w http.ResponseWriter, mType, name string)
	GetMetrics(w http.ResponseWriter)
}

type MetricsUpdateParams struct {
	ID    string
	MType string
	Delta int64
	Value float64
}

type metricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(s service.MetricsService) MetricsHandler {
	return &metricsHandler{service: s}
}

func (h *metricsHandler) Update(w http.ResponseWriter, params MetricsUpdateParams) {
	switch params.MType {
	case models.Counter:
		h.service.UpdateCounterMetricValue(service.UpdateCounterMetricValueParams{
			Type:  params.MType,
			Name:  params.ID,
			Value: params.Delta,
		})
	case models.Gauge:
		h.service.UpdateGaugeMetricValue(service.UpdateGaugeMetricValueParams{
			Type:  params.MType,
			Name:  params.ID,
			Value: params.Value,
		})
	default:
		http.Error(w, "Неизвестный тип метрики", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
}

func (h *metricsHandler) GetMetric(w http.ResponseWriter, mType, name string) {
	metric, exists := h.service.GetMetric(mType, name)

	if !exists {
		http.Error(w, "Неизвестная метрика", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	value := ""
	if metric.MType == models.Counter {
		value = fmt.Sprint(*metric.Delta)
	} else {
		value = fmt.Sprint(*metric.Value)
	}

	fmt.Fprintf(w, "%s", html.EscapeString(value))
}

func (h *metricsHandler) GetMetrics(w http.ResponseWriter) {
	metrics := h.service.GetMetrics()

	names := make([]string, 0, len(metrics))
	for name := range metrics {
		names = append(names, name)
	}
	sort.Strings(names)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, "<!doctype html><html><body><ul>")

	for _, name := range names {
		value := ""
		metric := metrics[name]

		if metric.MType == models.Counter {
			value = fmt.Sprint(*metric.Delta)
		} else {
			value = fmt.Sprint(*metric.Value)
		}

		fmt.Fprintf(
			w,
			"<li>%s: %s</li>",
			html.EscapeString(name),
			html.EscapeString(value),
		)
	}

	fmt.Fprint(w, "</ul></body></html>")
}
