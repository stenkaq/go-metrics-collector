package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"

	models "go-metrics-collector/internal/model"

	"github.com/go-resty/resty/v2"
)

type Agent interface {
	CollectMetrics()
	SendMetrics()
}

type HTTPAgent struct {
	HTTPClient *resty.Client
	BaseURL    string
	MValues    MetricsValues
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

func (h *HTTPAgent) sendMetric(name string, mType string, value any) error {
	if h.HTTPClient == nil {
		return fmt.Errorf("HTTP-клиент не настроен")
	}

	url := fmt.Sprintf("%s/update/", h.BaseURL)
	body := models.Metrics{
		MType: mType,
		ID:    name,
	}

	switch mType {
	case models.Counter:
		v := value.(int64)
		body.Delta = &v
	case models.Gauge:
		v := value.(float64)
		body.Value = &v
	}

	parsedBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("запрос %s: %w", url, err)
	}
	response, err := h.HTTPClient.R().SetHeader("Content-Type", "application/json").SetBody(parsedBody).Post(url)

	if err != nil {
		return fmt.Errorf("запрос %s: %w", url, err)
	}

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("ответ %s: status=%s body=%q", url, response.Status(), response.String())
	}

	return nil
}

func (h *HTTPAgent) SendMetrics() {
	h.MValues.mu.RLock()

	gaugeMetrics := maps.Clone(h.MValues.Gauges)
	counterMetrics := maps.Clone(h.MValues.Counters)

	h.MValues.mu.RUnlock()

	for name, value := range counterMetrics {
		err := h.sendMetric(name, models.Counter, value)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", name, err)
		}
	}

	for name, value := range gaugeMetrics {
		err := h.sendMetric(name, models.Gauge, value)
		if err != nil {
			log.Printf("Ошибка отправки метрики %s: %v", name, err)
		}
	}
}
