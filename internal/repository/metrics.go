package repository

import (
	models "go-metrics-collector/internal/model"
)

type MetricsRepository interface {
	Save(key string, params models.Metrics)
	GetByKey(key string) (models.Metrics, bool)
}

type metricsRepository struct {
	memStorage memStorage
}

type memStorage struct {
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
	s.memStorage.metrics[key] = params
}

func (s *metricsRepository) GetByKey(key string) (models.Metrics, bool) {
	val, exists := s.memStorage.metrics[key]

	return val, exists
}
