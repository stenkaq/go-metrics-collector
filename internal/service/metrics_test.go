package service

import (
	"context"
	"sort"
	"testing"

	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/repository"
)

type metricsRepositoryStub struct {
	metrics map[string]models.Metrics
}

func newMetricsRepositoryStub() repository.MetricsRepository {
	return &metricsRepositoryStub{metrics: make(map[string]models.Metrics)}
}

// key намеренно отличается от формата настоящего репозитория:
// сервис о раскладке хранилища знать не должен.
func (r *metricsRepositoryStub) key(mType string, name string) string {
	return mType + "|" + name
}

func (r *metricsRepositoryStub) Save(_ context.Context, metric models.Metrics) error {
	r.metrics[r.key(metric.MType, metric.ID)] = metric
	return nil
}

func (r *metricsRepositoryStub) SaveBatch(_ context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		key := r.key(metric.MType, metric.ID)

		if metric.MType == models.Counter {
			if stored, exists := r.metrics[key]; exists && stored.Delta != nil && metric.Delta != nil {
				delta := *stored.Delta + *metric.Delta
				metric.Delta = &delta
			}
		}

		r.metrics[key] = metric
	}

	return nil
}

func (r *metricsRepositoryStub) Get(_ context.Context, mType string, name string) (models.Metrics, bool, error) {
	metric, exists := r.metrics[r.key(mType, name)]
	return metric, exists, nil
}

func (r *metricsRepositoryStub) GetMetrics(_ context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, len(r.metrics))
	for _, metric := range r.metrics {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (r *metricsRepositoryStub) UpdateGauge(_ context.Context, name string, value float64) error {
	r.metrics[r.key(models.Gauge, name)] = models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	}

	return nil
}

func (r *metricsRepositoryStub) IncrementCounter(_ context.Context, name string, delta int64) error {
	key := r.key(models.Counter, name)
	metric, exists := r.metrics[key]

	if exists && metric.Delta != nil {
		delta += *metric.Delta
	}

	metric.ID = name
	metric.MType = models.Counter
	metric.Delta = &delta
	metric.Value = nil

	r.metrics[key] = metric

	return nil
}

func TestMetricsServiceUpdateCounterMetricValue(t *testing.T) {
	ctx := context.Background()
	repository := newMetricsRepositoryStub()
	metricsService := NewMetricsService(repository)

	metricsService.UpdateCounterMetricValue(ctx, UpdateCounterMetricValueParams{
		Type:  models.Counter,
		Name:  "PollCount",
		Value: 3,
	})
	metricsService.UpdateCounterMetricValue(ctx, UpdateCounterMetricValueParams{
		Type:  models.Counter,
		Name:  "PollCount",
		Value: 2,
	})

	metric, exists := mustGet(t, repository, models.Counter, "PollCount")
	if !exists {
		t.Fatal("counter metric не была сохранена")
	}
	if metric.ID != "PollCount" {
		t.Errorf("metric id = %q, want %q", metric.ID, "PollCount")
	}
	if metric.MType != models.Counter {
		t.Errorf("metric type = %q, want %q", metric.MType, models.Counter)
	}
	if metric.Delta == nil {
		t.Fatal("counter value is nil")
	}
	if *metric.Delta != 5 {
		t.Errorf("counter value = %d, want 5", *metric.Delta)
	}
}

func TestMetricsServiceUpdateGaugeMetricValue(t *testing.T) {
	ctx := context.Background()
	repository := newMetricsRepositoryStub()
	metricsService := NewMetricsService(repository)

	metricsService.UpdateGaugeMetricValue(ctx, UpdateGaugeMetricValueParams{
		Type:  models.Gauge,
		Name:  "Alloc",
		Value: 12.5,
	})
	metricsService.UpdateGaugeMetricValue(ctx, UpdateGaugeMetricValueParams{
		Type:  models.Gauge,
		Name:  "Alloc",
		Value: 4.25,
	})

	metric, exists := mustGet(t, repository, models.Gauge, "Alloc")
	if !exists {
		t.Fatal("gauge metric was not saved")
	}
	if metric.ID != "Alloc" {
		t.Errorf("metric id = %q, want %q", metric.ID, "Alloc")
	}
	if metric.MType != models.Gauge {
		t.Errorf("metric type = %q, want %q", metric.MType, models.Gauge)
	}
	if metric.Delta != nil {
		t.Errorf("gauge delta = %v, want nil", *metric.Delta)
	}
	if metric.Value == nil {
		t.Fatal("gauge value is nil")
	}
	if *metric.Value != 4.25 {
		t.Errorf("gauge value = %v, want 4.25", *metric.Value)
	}
}

func TestMetricsServiceGetMetric(t *testing.T) {
	ctx := context.Background()
	repository := newMetricsRepositoryStub()
	metricsService := NewMetricsService(repository)

	metricsService.UpdateCounterMetricValue(ctx, UpdateCounterMetricValueParams{Name: "Same", Value: 7})
	metricsService.UpdateGaugeMetricValue(ctx, UpdateGaugeMetricValueParams{Name: "Same", Value: 1.5})

	counter, exists := mustGetFromService(t, metricsService, models.Counter, "Same")
	if !exists || counter.Delta == nil || *counter.Delta != 7 {
		t.Errorf("counter = %+v, want delta 7", counter)
	}

	gauge, exists := mustGetFromService(t, metricsService, models.Gauge, "Same")
	if !exists || gauge.Value == nil || *gauge.Value != 1.5 {
		t.Errorf("gauge = %+v, want value 1.5", gauge)
	}

	if _, exists := mustGetFromService(t, metricsService, models.Gauge, "Nope"); exists {
		t.Error("GetMetric вернул несуществующую метрику")
	}
}

func TestMetricsServiceRestore(t *testing.T) {
	repository := newMetricsRepositoryStub()
	metricsService := NewMetricsService(repository)

	delta := int64(42)
	value := 12.5
	mustRestore(t, metricsService, []models.Metrics{
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
		{ID: "Alloc", MType: models.Gauge, Value: &value},
	})

	metrics := mustGetMetrics(t, metricsService)
	sort.Slice(metrics, func(i, j int) bool { return metrics[i].ID < metrics[j].ID })

	if len(metrics) != 2 {
		t.Fatalf("метрик = %d, want 2", len(metrics))
	}
	if metrics[0].ID != "Alloc" || metrics[0].Value == nil || *metrics[0].Value != value {
		t.Errorf("metrics[0] = %+v, want Alloc = %v", metrics[0], value)
	}
	if metrics[1].ID != "PollCount" || metrics[1].Delta == nil || *metrics[1].Delta != delta {
		t.Errorf("metrics[1] = %+v, want PollCount = %v", metrics[1], delta)
	}
}

func mustGet(
	t *testing.T,
	r repository.MetricsRepository,
	mType string,
	name string,
) (models.Metrics, bool) {
	t.Helper()

	metric, exists, err := r.Get(context.Background(), mType, name)
	if err != nil {
		t.Fatalf("repository.Get: %v", err)
	}

	return metric, exists
}

func mustGetFromService(
	t *testing.T,
	s MetricsService,
	mType string,
	name string,
) (models.Metrics, bool) {
	t.Helper()

	metric, exists, err := s.GetMetric(context.Background(), mType, name)
	if err != nil {
		t.Fatalf("GetMetric: %v", err)
	}

	return metric, exists
}

func mustGetMetrics(t *testing.T, s MetricsService) []models.Metrics {
	t.Helper()

	metrics, err := s.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}

	return metrics
}

func mustRestore(t *testing.T, s MetricsService, metrics []models.Metrics) {
	t.Helper()

	if err := s.Restore(context.Background(), metrics); err != nil {
		t.Fatalf("Restore: %v", err)
	}
}
