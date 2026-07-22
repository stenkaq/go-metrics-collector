package service

import (
	"sync"

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
	updateMu   sync.Mutex
}

func NewMetricsService(r repository.MetricsRepository) MetricsService {
	return &metricsService{repository: r}
}

func (s *metricsService) UpdateGaugeMetricValue(params UpdateGaugeMetricValueParams) {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	metric, _ := s.GetMetric(models.Gauge, params.Name)
	value := params.Value

	metric.ID = params.Name
	metric.MType = models.Gauge
	metric.Delta = nil
	metric.Value = &value

	s.saveMetric(metric)
}

func (s *metricsService) UpdateCounterMetricValue(params UpdateCounterMetricValueParams) {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	metric, exists := s.GetMetric(models.Counter, params.Name)
	delta := params.Value
	if exists && metric.Delta != nil {
		delta += *metric.Delta
	}

	metric.ID = params.Name
	metric.MType = models.Counter
	metric.Delta = &delta
	metric.Value = nil

	s.saveMetric(metric)
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
