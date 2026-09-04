package service

import (
	"context"
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
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
	UpdateGaugeMetricValue(ctx context.Context, params UpdateGaugeMetricValueParams) error
	UpdateCounterMetricValue(ctx context.Context, params UpdateCounterMetricValueParams) error
	GetMetric(ctx context.Context, mType string, name string) (models.Metrics, bool, error)
	GetMetrics(ctx context.Context) ([]models.Metrics, error)
	Restore(ctx context.Context, metrics []models.Metrics) error
}

var _ MetricsService = (*metricsService)(nil)

type metricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(r repository.MetricsRepository) MetricsService {
	return &metricsService{repository: r}
}

func (s *metricsService) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	return s.repository.SaveBatch(ctx, metrics)
}

func (s *metricsService) UpdateGaugeMetricValue(ctx context.Context, params UpdateGaugeMetricValueParams) error {
	return s.repository.UpdateGauge(ctx, params.Name, params.Value)
}

func (s *metricsService) UpdateCounterMetricValue(ctx context.Context, params UpdateCounterMetricValueParams) error {
	return s.repository.IncrementCounter(ctx, params.Name, params.Value)
}

func (s *metricsService) GetMetrics(ctx context.Context) ([]models.Metrics, error) {
	return s.repository.GetMetrics(ctx)
}

func (s *metricsService) GetMetric(ctx context.Context, mType string, name string) (models.Metrics, bool, error) {
	return s.repository.Get(ctx, mType, name)
}

func (s *metricsService) Restore(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := s.repository.Save(ctx, metric); err != nil {
			return err
		}
	}

	return nil
}
