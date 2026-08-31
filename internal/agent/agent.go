package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"

	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/retry"

	"github.com/go-resty/resty/v2"
)

type Agent interface {
	CollectMetrics()
	SendMetrics(ctx context.Context)
}

type HTTPAgent struct {
	HTTPClient       *resty.Client
	BaseURL          string
	MValues          MetricsValues
	batchUnsupported atomic.Bool
}

type retriableError struct {
	StatusCode int
	Err        error
}

func (e *retriableError) Error() string {
	return e.Err.Error()
}

func (e *retriableError) Unwrap() error {
	return e.Err
}

type MetricsValues struct {
	mu       sync.RWMutex
	Counters map[string]int64
	Gauges   map[string]float64
}

func (h *HTTPAgent) CollectMetrics() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	h.MValues.mu.Lock()
	defer h.MValues.mu.Unlock()

	h.MValues.Gauges["Alloc"] = float64(stats.Alloc)
	h.MValues.Gauges["BuckHashSys"] = float64(stats.BuckHashSys)
	h.MValues.Gauges["Frees"] = float64(stats.Frees)
	h.MValues.Gauges["GCCPUFraction"] = stats.GCCPUFraction
	h.MValues.Gauges["GCSys"] = float64(stats.GCSys)
	h.MValues.Gauges["HeapAlloc"] = float64(stats.HeapAlloc)
	h.MValues.Gauges["HeapIdle"] = float64(stats.HeapIdle)
	h.MValues.Gauges["HeapInuse"] = float64(stats.HeapInuse)
	h.MValues.Gauges["HeapObjects"] = float64(stats.HeapObjects)
	h.MValues.Gauges["HeapReleased"] = float64(stats.HeapReleased)
	h.MValues.Gauges["HeapSys"] = float64(stats.HeapSys)
	h.MValues.Gauges["LastGC"] = float64(stats.LastGC)
	h.MValues.Gauges["Lookups"] = float64(stats.Lookups)
	h.MValues.Gauges["MCacheInuse"] = float64(stats.MCacheInuse)
	h.MValues.Gauges["MCacheSys"] = float64(stats.MCacheSys)
	h.MValues.Gauges["MSpanInuse"] = float64(stats.MSpanInuse)
	h.MValues.Gauges["MSpanSys"] = float64(stats.MSpanSys)
	h.MValues.Gauges["Mallocs"] = float64(stats.Mallocs)
	h.MValues.Gauges["NextGC"] = float64(stats.NextGC)
	h.MValues.Gauges["NumForcedGC"] = float64(stats.NumForcedGC)
	h.MValues.Gauges["NumGC"] = float64(stats.NumGC)
	h.MValues.Gauges["OtherSys"] = float64(stats.OtherSys)
	h.MValues.Gauges["PauseTotalNs"] = float64(stats.PauseTotalNs)
	h.MValues.Gauges["StackInuse"] = float64(stats.StackInuse)
	h.MValues.Gauges["StackSys"] = float64(stats.StackSys)
	h.MValues.Gauges["Sys"] = float64(stats.Sys)
	h.MValues.Gauges["TotalAlloc"] = float64(stats.TotalAlloc)

	h.MValues.Gauges["RandomValue"] = rand.Float64()

	h.MValues.Counters["PollCount"]++
}

var errBatchUnsupported = errors.New("сервер не поддерживает батчи")

func (h *HTTPAgent) GetBatch() []models.Metrics {
	h.MValues.mu.RLock()
	defer h.MValues.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(h.MValues.Counters)+len(h.MValues.Gauges))

	for name, value := range h.MValues.Counters {
		delta := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &delta,
		})
	}

	for name, value := range h.MValues.Gauges {
		gauge := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &gauge,
		})
	}

	return metrics
}

func (h *HTTPAgent) post(ctx context.Context, path string, payload any) (int, error) {
	if h.HTTPClient == nil {
		return 0, fmt.Errorf("HTTP-клиент не настроен")
	}

	url := h.BaseURL + path

	parsedBody, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("запрос %s: %w", url, err)
	}

	response, err := h.HTTPClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(parsedBody).
		Post(url)
	if err != nil {
		return 0, &retriableError{Err: fmt.Errorf("запрос %s: %w", url, err)}
	}

	status := response.StatusCode()

	if status == http.StatusOK {
		return status, nil
	}

	respErr := fmt.Errorf("ответ %s: status=%s body=%q", url, response.Status(), response.String())

	if isRetriableStatus(status) {
		return status, &retriableError{StatusCode: status, Err: respErr}
	}

	return status, respErr
}

func (h *HTTPAgent) sendBatch(ctx context.Context, metrics []models.Metrics) error {
	status, err := h.post(ctx, "/updates/", metrics)

	if err != nil && status == http.StatusNotFound {
		return fmt.Errorf("%w: %w", err, errBatchUnsupported)
	}

	return err
}

func (h *HTTPAgent) sendMetric(ctx context.Context, metric models.Metrics) error {
	_, err := h.post(ctx, "/update/", metric)

	return err
}

func (h *HTTPAgent) SendMetrics(ctx context.Context) {
	metrics := h.GetBatch()

	if len(metrics) == 0 {
		return
	}

	if !h.batchUnsupported.Load() {
		err := retry.Do(ctx, isRetriableError, func() error {
			return h.sendBatch(ctx, metrics)
		})
		if err == nil {
			return
		}

		if !errors.Is(err, errBatchUnsupported) {
			log.Printf("Ошибка отправки батча метрик: %v", err)
			return
		}

		h.batchUnsupported.Store(true)
	}

	for _, metric := range metrics {
		if err := retry.Do(ctx, isRetriableError, func() error {
			return h.sendMetric(ctx, metric)
		}); err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", metric.ID, err)
		}
	}
}
