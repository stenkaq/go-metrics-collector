package service

import (
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
	GetMetrics() []models.Metrics
	Restore(metrics []models.Metrics)
}

type metricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(r repository.MetricsRepository) MetricsService {
	return &metricsService{repository: r}
}

func (s *metricsService) UpdateGaugeMetricValue(params UpdateGaugeMetricValueParams) {
	value := params.Value

	s.repository.Save(models.Metrics{
		ID:    params.Name,
		MType: models.Gauge,
		Value: &value,
	})
}

func (s *metricsService) UpdateCounterMetricValue(params UpdateCounterMetricValueParams) {
	s.repository.IncrementCounter(params.Name, params.Value)
}

func (s *metricsService) GetMetrics() []models.Metrics {
	return s.repository.GetMetrics()
}

func (s *metricsService) GetMetric(mType string, name string) (models.Metrics, bool) {
	return s.repository.Get(mType, name)
}

func (s *metricsService) Restore(metrics []models.Metrics) {
	for _, metric := range metrics {
		s.repository.Save(metric)
	}
}
