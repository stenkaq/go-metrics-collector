package handler

import (
	"fmt"
	"html"
	"net/http"
	"sort"
	"strconv"
	"strings"

	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/service"

	"github.com/gin-gonic/gin"
)

type MetricsHandler interface {
	UpdateMetric(c *gin.Context)
	UpdateMetricV2(c *gin.Context)
	GetMetric(c *gin.Context)
	GetMetricV2(c *gin.Context)
	GetMetrics(c *gin.Context)
}

type metricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(s service.MetricsService) MetricsHandler {
	return &metricsHandler{service: s}
}

func (h *metricsHandler) UpdateMetricV2(c *gin.Context) {
	var metric models.Metrics

	if err := c.ShouldBindJSON(&metric); err != nil {
		http.Error(c.Writer, "Ошибка при чтении тела запроса", http.StatusBadRequest)
		return
	}

	if metric.ID == "" || metric.MType == "" || (metric.Delta != nil && metric.Value != nil) {
		http.Error(c.Writer, "Неверные значения полей в теле запроса", http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			http.Error(c.Writer, "Неверные значения полей в теле запроса", http.StatusBadRequest)
			return
		}

		if err := h.service.UpdateCounterMetricValue(c.Request.Context(), service.UpdateCounterMetricValueParams{
			Type:  metric.MType,
			Name:  metric.ID,
			Value: *metric.Delta,
		}); err != nil {
			http.Error(c.Writer, "Не удалось сохранить метрику", http.StatusInternalServerError)
			return
		}
	case models.Gauge:
		if metric.Value == nil {
			http.Error(c.Writer, "Неверные значения полей в теле запроса", http.StatusBadRequest)
			return
		}

		if err := h.service.UpdateGaugeMetricValue(c.Request.Context(), service.UpdateGaugeMetricValueParams{
			Type:  metric.MType,
			Name:  metric.ID,
			Value: *metric.Value,
		}); err != nil {
			http.Error(c.Writer, "Не удалось сохранить метрику", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(c.Writer, "Неизвестный тип метрики", http.StatusBadRequest)
		return
	}

	c.Header("Content-Type", "text/plain")
}

func (h *metricsHandler) UpdateMetric(c *gin.Context) {
	mType := c.Param("type")
	name := c.Param("name")
	value := c.Param("value")

	if strings.TrimSpace(name) == "" {
		http.Error(c.Writer, "Пустое имя метрики", http.StatusNotFound)
		return
	}

	switch mType {
	case models.Counter:
		parsedVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(c.Writer, "Неверное значение метрики", http.StatusBadRequest)
			return
		}

		if err := h.service.UpdateCounterMetricValue(c.Request.Context(), service.UpdateCounterMetricValueParams{
			Type:  mType,
			Name:  name,
			Value: parsedVal,
		}); err != nil {
			http.Error(c.Writer, "Не удалось сохранить метрику", http.StatusInternalServerError)
			return
		}
	case models.Gauge:
		parsedVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(c.Writer, "Неверное значение метрики", http.StatusBadRequest)
			return
		}

		if err := h.service.UpdateGaugeMetricValue(c.Request.Context(), service.UpdateGaugeMetricValueParams{
			Type:  mType,
			Name:  name,
			Value: parsedVal,
		}); err != nil {
			http.Error(c.Writer, "Не удалось сохранить метрику", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(c.Writer, "Неизвестный тип метрики", http.StatusBadRequest)
		return
	}

	c.Header("Content-Type", "text/plain")
}

func (h *metricsHandler) GetMetricV2(c *gin.Context) {
	var request models.Metrics

	if err := c.ShouldBindJSON(&request); err != nil {
		http.Error(c.Writer, "Ошибка при чтении тела запроса", http.StatusBadRequest)
		return
	}

	metric, exists, err := h.service.GetMetric(c.Request.Context(), request.MType, request.ID)
	if err != nil {
		http.Error(c.Writer, "Не удалось прочитать метрику", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(c.Writer, "Неизвестная метрика", http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, metric)
}

func (h *metricsHandler) GetMetric(c *gin.Context) {
	metric, exists, err := h.service.GetMetric(c.Request.Context(), c.Param("type"), c.Param("name"))
	if err != nil {
		http.Error(c.Writer, "Не удалось прочитать метрику", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(c.Writer, "Неизвестная метрика", http.StatusNotFound)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(c.Writer, html.EscapeString(formatValue(metric)))
}

func (h *metricsHandler) GetMetrics(c *gin.Context) {
	metrics, err := h.service.GetMetrics(c.Request.Context())
	if err != nil {
		http.Error(c.Writer, "Не удалось прочитать метрики", http.StatusInternalServerError)
		return
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	c.Header("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(c.Writer, "<!doctype html><html><body><ul>")

	for _, metric := range metrics {
		fmt.Fprintf(
			c.Writer,
			"<li>%s: %s</li>",
			html.EscapeString(metric.ID),
			html.EscapeString(formatValue(metric)),
		)
	}

	fmt.Fprint(c.Writer, "</ul></body></html>")
}

func formatValue(metric models.Metrics) string {
	if metric.MType == models.Counter {
		return fmt.Sprint(*metric.Delta)
	}

	return fmt.Sprint(*metric.Value)
}
