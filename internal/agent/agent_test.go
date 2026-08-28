package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	models "go-metrics-collector/internal/model"

	"github.com/go-resty/resty/v2"
)

var runtimeMetricNames = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
	"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
	"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
	"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
	"StackSys", "Sys", "TotalAlloc", "RandomValue",
}

func TestHTTPAgentCollectMetrics(t *testing.T) {
	agent := &HTTPAgent{
		MValues: MetricsValues{
			Counters: map[string]int64{
				"PollCount":     41,
				"CustomCounter": 7,
			},
			Gauges: map[string]float64{
				"CustomGauge": 1.5,
			},
		},
	}

	agent.CollectMetrics()

	if got := agent.MValues.Counters["PollCount"]; got != 42 {
		t.Fatalf("PollCount = %d, want 42", got)
	}
	if got := agent.MValues.Counters["CustomCounter"]; got != 7 {
		t.Errorf("CustomCounter = %d, want 7", got)
	}
	if got := agent.MValues.Gauges["CustomGauge"]; got != 1.5 {
		t.Errorf("CustomGauge = %v, want 1.5", got)
	}

	for _, name := range runtimeMetricNames {
		if _, ok := agent.MValues.Gauges[name]; !ok {
			t.Errorf("metric %q was not collected", name)
		}
	}

	randomValue := agent.MValues.Gauges["RandomValue"]
	if randomValue < 0 || randomValue >= 1 {
		t.Errorf("RandomValue = %v, want value in [0, 1)", randomValue)
	}
}

// newTestAgent поднимает тестовый сервер и агента с тем же клиентом,
// что собирает app.NewAgent — то есть со сжатием тела в gzip.
func newTestAgent(t *testing.T, handler http.HandlerFunc) *HTTPAgent {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &HTTPAgent{
		BaseURL:    server.URL,
		HTTPClient: resty.New().OnBeforeRequest(GzipRequest()),
		MValues: MetricsValues{
			Counters: map[string]int64{"PollCount": 3},
			Gauges:   map[string]float64{"Alloc": 12.5},
		},
	}
}

// readBody распаковывает тело запроса, если агент сжал его в gzip.
func readBody(t *testing.T, r *http.Request) []byte {
	t.Helper()

	var reader io.Reader = r.Body

	if r.Header.Get("Content-Encoding") == "gzip" {
		zr, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("не удалось распаковать тело: %v", err)
		}
		defer zr.Close()

		reader = zr
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("не удалось прочитать тело: %v", err)
	}

	return body
}

func TestHTTPAgentSendMetricsSendsBatch(t *testing.T) {
	var (
		mu       sync.Mutex
		paths    []string
		encoding string
		batch    []models.Metrics
	)

	agent := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)

		mu.Lock()
		defer mu.Unlock()

		paths = append(paths, r.URL.Path)
		encoding = r.Header.Get("Content-Encoding")

		if err := json.Unmarshal(body, &batch); err != nil {
			t.Errorf("тело запроса не разобралось как []Metrics: %v", err)
		}
	})

	agent.SendMetrics()

	if len(paths) != 1 {
		t.Fatalf("запросов = %d, want 1 — метрики должны уходить одним батчем", len(paths))
	}
	if paths[0] != "/updates/" {
		t.Errorf("путь = %q, want %q", paths[0], "/updates/")
	}
	if encoding != "gzip" {
		t.Errorf("Content-Encoding = %q, want %q", encoding, "gzip")
	}
	if len(batch) != 2 {
		t.Fatalf("метрик в батче = %d, want 2", len(batch))
	}

	for _, metric := range batch {
		switch metric.ID {
		case "PollCount":
			if metric.MType != models.Counter || metric.Delta == nil || *metric.Delta != 3 {
				t.Errorf("PollCount = %+v, want counter delta 3", metric)
			}
		case "Alloc":
			if metric.MType != models.Gauge || metric.Value == nil || *metric.Value != 12.5 {
				t.Errorf("Alloc = %+v, want gauge value 12.5", metric)
			}
		default:
			t.Errorf("в батче лишняя метрика %q", metric.ID)
		}
	}
}

func TestHTTPAgentSendMetricsSkipsEmptyBatch(t *testing.T) {
	var requests int

	agent := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
	})

	agent.MValues.Counters = map[string]int64{}
	agent.MValues.Gauges = map[string]float64{}

	agent.SendMetrics()

	if requests != 0 {
		t.Errorf("запросов = %d, want 0 — пустой батч отправлять не нужно", requests)
	}
}

func TestHTTPAgentSendMetricsFallsBackToSingleUpdates(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)

	agent := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		paths = append(paths, r.URL.Path)

		// старый сервер про батчи не знает
		if r.URL.Path == "/updates/" {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	agent.SendMetrics()

	if len(paths) != 3 {
		t.Fatalf("запросов = %d, want 3 (батч + две поштучные отправки): %v", len(paths), paths)
	}
	if paths[0] != "/updates/" {
		t.Errorf("первый запрос = %q, want %q", paths[0], "/updates/")
	}
	for _, path := range paths[1:] {
		if path != "/update/" {
			t.Errorf("запрос отката = %q, want %q", path, "/update/")
		}
	}
}

func TestHTTPAgentSendMetricsIsRaceFree(t *testing.T) {
	agent := newTestAgent(t, func(w http.ResponseWriter, r *http.Request) {
		readBody(t, r)
	})

	var wg sync.WaitGroup

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.CollectMetrics()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.SendMetrics()
		}()
	}

	wg.Wait()
}
