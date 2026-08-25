package repository

import (
	"context"
	"sync"

	models "go-metrics-collector/internal/model"
)

type MetricsRepository interface {
	Save(ctx context.Context, metric models.Metrics) error
	Get(ctx context.Context, mType string, name string) (models.Metrics, bool, error)
	GetMetrics(ctx context.Context) ([]models.Metrics, error)
	IncrementCounter(ctx context.Context, name string, delta int64) error
}

var _ MetricsRepository = (*metricsRepository)(nil)

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

func (s *metricsRepository) Save(_ context.Context, metric models.Metrics) error {
	s.memStorage.mu.Lock()
	defer s.memStorage.mu.Unlock()

	s.memStorage.metrics[compoundKey(metric.MType, metric.ID)] = metric

	return nil
}

func (s *metricsRepository) Get(_ context.Context, mType string, name string) (models.Metrics, bool, error) {
	s.memStorage.mu.RLock()
	defer s.memStorage.mu.RUnlock()

	val, exists := s.memStorage.metrics[compoundKey(mType, name)]

	return val, exists, nil
}

func (s *metricsRepository) GetMetrics(_ context.Context) ([]models.Metrics, error) {
	s.memStorage.mu.RLock()
	defer s.memStorage.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(s.memStorage.metrics))
	for _, metric := range s.memStorage.metrics {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (s *metricsRepository) IncrementCounter(_ context.Context, name string, delta int64) error {
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

	return nil
}

func compoundKey(mType string, name string) string {
	return mType + "-" + name
}
