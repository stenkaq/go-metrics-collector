package service

import (
	"fmt"
	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/repository"
)

type UpdateGaugeMetricValueParams struct {
	Type  string
	Name  string
	Value float64
}

type UpdateCounterMetricValueParams struct {
	Type  string
	Name  string
	Value int64
}

type MetricsService interface {
	UpdateGaugeMetricValue(params UpdateGaugeMetricValueParams)
	UpdateCounterMetricValue(params UpdateCounterMetricValueParams)
	GetMetric(mType string, name string) (models.Metrics, bool)
	GetMetrics() map[string]models.Metrics
}

type metricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(r repository.MetricsRepository) MetricsService {
	return &metricsService{repository: r}
}

func (s *metricsService) UpdateGaugeMetricValue(params UpdateGaugeMetricValueParams) {
	metrics, exists := s.GetMetric(params.Type, params.Name)

	metrics.Value = &params.Value

	if !exists {
		metrics.MType = models.Gauge
	}

	s.saveMetric(params.Name, metrics)

	fmt.Printf("Сохранена метрика: %s, тип: %s, значение: %v\n", params.Name, metrics.MType, *metrics.Value)
}

func (s *metricsService) UpdateCounterMetricValue(params UpdateCounterMetricValueParams) {
	metrics, exists := s.GetMetric(params.Type, params.Name)

	if exists {
		*metrics.Delta += params.Value
	} else {
		metrics.Delta = &params.Value
		metrics.MType = models.Counter
	}

	s.saveMetric(params.Name, metrics)

	fmt.Printf("Сохранена метрика: %s, тип: %s, дельта: %v\n", params.Name, metrics.MType, *metrics.Delta)
}

func (s *metricsService) GetMetrics() map[string]models.Metrics {
	return s.repository.GetMetrics()
}

func (s *metricsService) GetMetric(mType string, name string) (models.Metrics, bool) {
	key := s.getCompoundKey(mType, name)

	metrics, exists := s.repository.GetByKey(key)

	return metrics, exists
}

func (s *metricsService) saveMetric(name string, metrics models.Metrics) {
	key := s.getCompoundKey(metrics.MType, name)

	s.repository.Save(key, metrics)
}

func (s *metricsService) getCompoundKey(mType string, name string) string {
	return fmt.Sprintf("%s-%s", mType, name)
}
