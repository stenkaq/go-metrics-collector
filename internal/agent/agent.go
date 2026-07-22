package agent

import (
	"fmt"
	models "go-metrics-collector/internal/model"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"

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
	h.MValues.mu.Lock()
	defer h.MValues.mu.Unlock()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

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

func (h *HTTPAgent) sendMetric(name string, mType string, value any) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", h.BaseURL, mType, name, value)

	response, err := h.HTTPClient.R().SetHeader("Content-Type", "text/plain").Post(url)

	if err != nil {
		log.Printf("Ошибка запроса %s: %v\n", url, err)
		return
	}

	if response.StatusCode() != http.StatusOK {
		fmt.Printf("Ошибка ответа: %s - status=%s body=%q\n", url, response.Status(), response.String())
		return
	}

	fmt.Printf("Успешно отправлена метрика: %s - %v, status %s\n", name, value, response.Status())
}

func (h *HTTPAgent) SendMetrics() {
	h.MValues.mu.RLock()
	defer h.MValues.mu.RUnlock()

	for name, value := range h.MValues.Counters {
		h.sendMetric(name, models.Counter, value)
	}

	for name, value := range h.MValues.Gauges {
		h.sendMetric(name, models.Gauge, value)
	}
}
