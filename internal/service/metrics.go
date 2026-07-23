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
	GetMetrics() map[string]models.Metrics
}

type metricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(r repository.MetricsRepository) MetricsService {
	return &metricsService{repository: r}
}

func (s *metricsService) UpdateGaugeMetricValue(params UpdateGaugeMetricValueParams) {
	metric, _ := s.GetMetric(models.Gauge, params.Name)
	value := params.Value

	metric.ID = params.Name
	metric.MType = models.Gauge
	metric.Delta = nil
	metric.Value = &value

	s.saveMetric(metric)
}

func (s *metricsService) UpdateCounterMetricValue(params UpdateCounterMetricValueParams) {
	s.repository.IncrementCounter(repository.IncrementCounterParams{
		Value: params.Value,
		Name:  params.Name,
	})
}

func (s *metricsService) GetMetrics() map[string]models.Metrics {
	return s.repository.GetMetrics()
}

func (s *metricsService) GetMetric(mType string, name string) (models.Metrics, bool) {
	key := s.getCompoundKey(mType, name)

	metrics, exists := s.repository.GetByKey(key)

	return metrics, exists
}

func (s *metricsService) saveMetric(metric models.Metrics) {
	key := s.getCompoundKey(metric.MType, metric.ID)

	s.repository.Save(key, metric)
}

func (s *metricsService) getCompoundKey(mType string, name string) string {
	return mType + "-" + name
}
