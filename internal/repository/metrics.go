package repository

import (
	models "go-metrics-collector/internal/model"
	"sync"
)

type MetricsRepository interface {
	Save(metric models.Metrics)
	Get(mType string, name string) (models.Metrics, bool)
	GetMetrics() []models.Metrics
	IncrementCounter(name string, delta int64)
}

type metricsRepository struct {
	memStorage memStorage
}

type memStorage struct {
	mu      sync.RWMutex
	metrics map[string]models.Metrics
}

func NewMetricsRepository() MetricsRepository {
	return &metricsRepository{
		memStorage: memStorage{
			metrics: make(map[string]models.Metrics),
		},
	}
}

func (s *metricsRepository) Save(metric models.Metrics) {
	s.memStorage.mu.Lock()
	defer s.memStorage.mu.Unlock()

	s.memStorage.metrics[compoundKey(metric.MType, metric.ID)] = metric
}

func (s *metricsRepository) Get(mType string, name string) (models.Metrics, bool) {
	s.memStorage.mu.RLock()
	defer s.memStorage.mu.RUnlock()

	val, exists := s.memStorage.metrics[compoundKey(mType, name)]

	return val, exists
}

func (s *metricsRepository) GetMetrics() []models.Metrics {
	s.memStorage.mu.RLock()
	defer s.memStorage.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(s.memStorage.metrics))
	for _, metric := range s.memStorage.metrics {
		metrics = append(metrics, metric)
	}

	return metrics
}

func (s *metricsRepository) IncrementCounter(name string, delta int64) {
	s.memStorage.mu.Lock()
	defer s.memStorage.mu.Unlock()

	key := compoundKey(models.Counter, name)
	metric, exists := s.memStorage.metrics[key]

	if exists && metric.Delta != nil {
		delta += *metric.Delta
	}

	metric.ID = name
	metric.MType = models.Counter
	metric.Delta = &delta

	s.memStorage.metrics[key] = metric
}

func compoundKey(mType string, name string) string {
	return mType + "-" + name
}
