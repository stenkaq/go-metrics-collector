package agent

import (
	"testing"
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
