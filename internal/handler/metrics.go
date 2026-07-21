package handler

import (
	"fmt"
	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/service"
	"html"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type MetricsHandler interface {
	UpdateMetric(w http.ResponseWriter, mType, name, value string)
	GetMetric(w http.ResponseWriter, mType, name string)
	GetMetrics(w http.ResponseWriter)
}

type metricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(s service.MetricsService) MetricsHandler {
	return &metricsHandler{service: s}
}

func (h *metricsHandler) UpdateMetric(w http.ResponseWriter, mType, name, value string) {
	if strings.TrimSpace(name) == "" {
		http.Error(w, "Пустое имя метрики", http.StatusNotFound)
		return
	}

	switch mType {
	case models.Counter:
		parsedVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "Неверное значение метрики", http.StatusBadRequest)
			return
		}

		h.service.UpdateCounterMetricValue(service.UpdateCounterMetricValueParams{
			Type:  mType,
			Name:  name,
			Value: parsedVal,
		})
	case models.Gauge:
		parsedVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "Неверное значение метрики", http.StatusBadRequest)
			return
		}

		h.service.UpdateGaugeMetricValue(service.UpdateGaugeMetricValueParams{
			Type:  mType,
			Name:  name,
			Value: parsedVal,
		})
	default:
		http.Error(w, "Неизвестный тип метрики", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "text/plain")
}

func (h *metricsHandler) GetMetric(w http.ResponseWriter, mType, name string) {
	metric, exists := h.service.GetMetric(mType, name)

	if !exists {
		http.Error(w, "Неизвестная метрика", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, "<!doctype html><html><body><ul>")

	value := ""
	if metric.MType == models.Counter {
		value = fmt.Sprint(*metric.Delta)
	} else {
		value = fmt.Sprint(*metric.Value)
	}

	fmt.Fprintf(w, "%s: %s", html.EscapeString(name), html.EscapeString(value))

	fmt.Fprint(w, "</ul></body></html>")
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
