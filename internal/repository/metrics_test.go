package repository

import (
	"testing"

	models "go-metrics-collector/internal/model"
)

func TestMetricsRepositoryGetMetricsReturnsSnapshot(t *testing.T) {
	repository := NewMetricsRepository()
	value := 42.0
	repository.Save("gauge-test", models.Metrics{MType: models.Gauge, Value: &value})

	metrics := repository.GetMetrics()
	delete(metrics, "gauge-test")

	stored, exists := repository.GetByKey("gauge-test")
	if !exists {
		t.Fatal("GetMetrics returned the repository storage map")
	}
	if stored.Value == nil || *stored.Value != value {
		t.Fatalf("stored metric = %+v, want gauge value %v", stored, value)
	}
}
