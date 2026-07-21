package app

import (
	"go-metrics-collector/internal/agent"

	"github.com/go-resty/resty/v2"
)

func NewAgent() agent.Agent {
	return &agent.HTTPAgent{
		MValues: agent.MetricsValues{
			Counters: make(map[string]int64),
			Gauges:   make(map[string]float64),
		},
		BaseURL:    "http://localhost:8080",
		HTTPClient: resty.New(),
	}
}
