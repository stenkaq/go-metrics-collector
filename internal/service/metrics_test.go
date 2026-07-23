package service

import (
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

func (r *metricsRepositoryStub) Save(key string, metric models.Metrics) {
	r.metrics[key] = metric
}

func (r *metricsRepositoryStub) GetByKey(key string) (models.Metrics, bool) {
	metric, exists := r.metrics[key]
	return metric, exists
}

func (r *metricsRepositoryStub) GetMetrics() map[string]models.Metrics {
	return r.metrics
}

func (r *metricsRepositoryStub) IncrementCounter(params repository.IncrementCounterParams) {
	key := models.Counter + "-" + params.Name
	metric, exists := r.metrics[key]

	delta := params.Value
	if exists && metric.Delta != nil {
		delta += *metric.Delta
	}

	metric.ID = params.Name
	metric.MType = models.Counter
	metric.Delta = &delta
	metric.Value = nil

	r.metrics[key] = metric
}

func TestMetricsServiceUpdateCounterMetricValue(t *testing.T) {
	repository := newMetricsRepositoryStub()
	metricsService := NewMetricsService(repository)

	metricsService.UpdateCounterMetricValue(UpdateCounterMetricValueParams{
		Type:  models.Counter,
		Name:  "PollCount",
		Value: 3,
	})
	metricsService.UpdateCounterMetricValue(UpdateCounterMetricValueParams{
		Type:  models.Counter,
		Name:  "PollCount",
		Value: 2,
	})

	metric, exists := repository.GetByKey("counter-PollCount")
	if !exists {
		t.Fatal("counter metric не была сохранена")
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
	repository := newMetricsRepositoryStub()
	metricsService := NewMetricsService(repository)

	metricsService.UpdateGaugeMetricValue(UpdateGaugeMetricValueParams{
		Type:  models.Gauge,
		Name:  "Alloc",
		Value: 12.5,
	})
	metricsService.UpdateGaugeMetricValue(UpdateGaugeMetricValueParams{
		Type:  models.Gauge,
		Name:  "Alloc",
		Value: 4.25,
	})

	metric, exists := repository.GetByKey("gauge-Alloc")
	if !exists {
		t.Fatal("gauge metric was not saved")
	}
	if metric.MType != models.Gauge {
		t.Errorf("metric type = %q, want %q", metric.MType, models.Gauge)
	}
	if metric.Value == nil {
		t.Fatal("gauge value is nil")
	}
	if *metric.Value != 4.25 {
		t.Errorf("gauge value = %v, want 4.25", *metric.Value)
	}
}

func TestMetricsServiceGetCompoundKey(t *testing.T) {
	service := &metricsService{}

	if got := service.getCompoundKey(models.Gauge, "Alloc"); got != "gauge-Alloc" {
		t.Errorf("compound key = %q, want %q", got, "gauge-Alloc")
	}
}
