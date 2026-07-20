package app

import (
	"go-metrics-collector/internal/agent"
	"net/http"
)

func NewAgent() agent.Agent {
	return &agent.HTTPAgent{
		MValues: agent.MetricsValues{
			Counters: make(map[string]int64),
			Gauges:   make(map[string]float64),
		},
		BaseUrl:    "http://localhost:8080",
		HttpClient: http.Client{},
	}
}
