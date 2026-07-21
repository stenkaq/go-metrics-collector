package repository

import (
	models "go-metrics-collector/internal/model"
	"sync"
)

type MetricsRepository interface {
	Save(key string, params models.Metrics)
	GetByKey(key string) (models.Metrics, bool)
	GetMetrics() map[string]models.Metrics
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

func (s *metricsRepository) Save(key string, params models.Metrics) {
	s.memStorage.mu.Lock()
	defer s.memStorage.mu.Unlock()

	s.memStorage.metrics[key] = params
}

func (s *metricsRepository) GetByKey(key string) (models.Metrics, bool) {
	s.memStorage.mu.RLock()
	defer s.memStorage.mu.RUnlock()

	val, exists := s.memStorage.metrics[key]

	return val, exists
}

func (s *metricsRepository) GetMetrics() map[string]models.Metrics {
	s.memStorage.mu.RLock()
	defer s.memStorage.mu.RUnlock()

	metrics := make(map[string]models.Metrics, len(s.memStorage.metrics))
	for key, metric := range s.memStorage.metrics {
		metrics[key] = metric
	}

	return metrics
}
