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
}

type metricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(r repository.MetricsRepository) MetricsService {
	return &metricsService{repository: r}
}

func (s *metricsService) UpdateGaugeMetricValue(params UpdateGaugeMetricValueParams) {
	key := s.getCompoundKey(params.Type, params.Name)
	metrics, exists := s.repository.GetByKey(key)

	metrics.Value = &params.Value

	if !exists {
		metrics.MType = models.Gauge
	}

	s.repository.Save(key, metrics)

	fmt.Printf("Сохранена метрика: %s, тип: %s, значение: %v\n", params.Name, metrics.MType, *metrics.Value)
}

func (s *metricsService) UpdateCounterMetricValue(params UpdateCounterMetricValueParams) {
	key := s.getCompoundKey(params.Type, params.Name)
	metrics, exists := s.repository.GetByKey(key)

	if exists {
		*metrics.Delta += params.Value
	} else {
		metrics.Delta = &params.Value
		metrics.MType = models.Counter
	}

	s.repository.Save(key, metrics)

	fmt.Printf("Сохранена метрика: %s, тип: %s, дельта: %v\n", params.Name, metrics.MType, *metrics.Delta)
}

func (s *metricsService) getCompoundKey(mType string, name string) string {
	return fmt.Sprintf("%s-%s", mType, name)
}
