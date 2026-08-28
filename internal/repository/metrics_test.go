package repository

import (
	"context"
	"testing"

	models "go-metrics-collector/internal/model"
)

func TestMetricsRepositoryGetMetricsReturnsSnapshot(t *testing.T) {
	ctx := context.Background()
	repository := NewMetricsRepository()
	value := 42.0
	mustUpdateGauge(t, repository, ctx, "test", value)

	metrics := mustGetMetrics(t, repository, ctx)
	metrics[0] = models.Metrics{}

	stored, exists := mustGet(t, repository, ctx, models.Gauge, "test")
	if !exists {
		t.Fatal("метрика пропала из репозитория")
	}
	if stored.Value == nil || *stored.Value != value {
		t.Fatalf("stored metric = %+v, want gauge value %v", stored, value)
	}
}

func TestMetricsRepositoryGetMetricsReturnsNamesNotKeys(t *testing.T) {
	ctx := context.Background()
	repository := NewMetricsRepository()
	value := 12.5
	mustUpdateGauge(t, repository, ctx, "Alloc", value)

	metrics := mustGetMetrics(t, repository, ctx)
	if len(metrics) != 1 {
		t.Fatalf("метрик = %d, want 1", len(metrics))
	}
	if metrics[0].ID != "Alloc" {
		t.Errorf("ID = %q, want %q — наружу не должен утекать ключ хранилища", metrics[0].ID, "Alloc")
	}
	if metrics[0].MType != models.Gauge {
		t.Errorf("MType = %q, want %q", metrics[0].MType, models.Gauge)
	}
}

func TestMetricsRepositorySeparatesTypesWithTheSameName(t *testing.T) {
	ctx := context.Background()
	repository := NewMetricsRepository()
	value := 1.5
	mustUpdateGauge(t, repository, ctx, "Same", value)
	mustIncrement(t, repository, ctx, "Same", 7)

	gauge, exists := mustGet(t, repository, ctx, models.Gauge, "Same")
	if !exists || gauge.Value == nil || *gauge.Value != value {
		t.Errorf("gauge = %+v, want value %v", gauge, value)
	}

	counter, exists := mustGet(t, repository, ctx, models.Counter, "Same")
	if !exists || counter.Delta == nil || *counter.Delta != 7 {
		t.Errorf("counter = %+v, want delta 7", counter)
	}
}

func TestMetricsRepositoryIncrementCounterAccumulates(t *testing.T) {
	ctx := context.Background()
	repository := NewMetricsRepository()

	mustIncrement(t, repository, ctx, "PollCount", 3)
	mustIncrement(t, repository, ctx, "PollCount", 2)

	metric, exists := mustGet(t, repository, ctx, models.Counter, "PollCount")
	if !exists {
		t.Fatal("counter metric не была сохранена")
	}
	if metric.ID != "PollCount" || metric.MType != models.Counter {
		t.Errorf("metric = %+v, want PollCount/counter", metric)
	}
	if metric.Delta == nil || *metric.Delta != 5 {
		t.Fatalf("delta = %v, want 5", metric.Delta)
	}
}

func TestMetricsRepositoryGetMissingMetric(t *testing.T) {
	ctx := context.Background()
	repository := NewMetricsRepository()

	if _, exists := mustGet(t, repository, ctx, models.Gauge, "Nope"); exists {
		t.Error("Get вернул несуществующую метрику")
	}
}

func mustGet(
	t *testing.T,
	r MetricsRepository,
	ctx context.Context,
	mType string,
	name string,
) (models.Metrics, bool) {
	t.Helper()

	metric, exists, err := r.Get(ctx, mType, name)
	if err != nil {
		t.Fatalf("Get(%q, %q): %v", mType, name, err)
	}

	return metric, exists
}

func mustGetMetrics(t *testing.T, r MetricsRepository, ctx context.Context) []models.Metrics {
	t.Helper()

	metrics, err := r.GetMetrics(ctx)
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}

	return metrics
}

func mustUpdateGauge(t *testing.T, r MetricsRepository, ctx context.Context, name string, value float64) {
	t.Helper()

	if err := r.UpdateGauge(ctx, name, value); err != nil {
		t.Fatalf("UpdateGauge(%q, %v): %v", name, value, err)
	}
}

func mustIncrement(t *testing.T, r MetricsRepository, ctx context.Context, name string, delta int64) {
	t.Helper()

	if err := r.IncrementCounter(ctx, name, delta); err != nil {
		t.Fatalf("IncrementCounter(%q, %d): %v", name, delta, err)
	}
}
