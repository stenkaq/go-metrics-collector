package repository

import (
	"testing"

	models "go-metrics-collector/internal/model"
)

func TestMetricsRepositoryGetMetricsReturnsSnapshot(t *testing.T) {
	repository := NewMetricsRepository()
	value := 42.0
	repository.Save(models.Metrics{ID: "test", MType: models.Gauge, Value: &value})

	metrics := repository.GetMetrics()
	metrics[0] = models.Metrics{}

	stored, exists := repository.Get(models.Gauge, "test")
	if !exists {
		t.Fatal("метрика пропала из репозитория")
	}
	if stored.Value == nil || *stored.Value != value {
		t.Fatalf("stored metric = %+v, want gauge value %v", stored, value)
	}
}

func TestMetricsRepositoryGetMetricsReturnsNamesNotKeys(t *testing.T) {
	repository := NewMetricsRepository()
	value := 12.5
	repository.Save(models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &value})

	metrics := repository.GetMetrics()
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
	repository := NewMetricsRepository()
	value := 1.5
	repository.Save(models.Metrics{ID: "Same", MType: models.Gauge, Value: &value})
	repository.IncrementCounter("Same", 7)

	gauge, exists := repository.Get(models.Gauge, "Same")
	if !exists || gauge.Value == nil || *gauge.Value != value {
		t.Errorf("gauge = %+v, want value %v", gauge, value)
	}

	counter, exists := repository.Get(models.Counter, "Same")
	if !exists || counter.Delta == nil || *counter.Delta != 7 {
		t.Errorf("counter = %+v, want delta 7", counter)
	}
}

func TestMetricsRepositoryIncrementCounterAccumulates(t *testing.T) {
	repository := NewMetricsRepository()

	repository.IncrementCounter("PollCount", 3)
	repository.IncrementCounter("PollCount", 2)

	metric, exists := repository.Get(models.Counter, "PollCount")
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
	repository := NewMetricsRepository()

	if _, exists := repository.Get(models.Gauge, "Nope"); exists {
		t.Error("Get вернул несуществующую метрику")
	}
}
