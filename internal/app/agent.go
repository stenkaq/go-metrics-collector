package app

import (
	"go-metrics-collector/internal/agent"

	"github.com/go-resty/resty/v2"
)

func NewAgent(address string) agent.Agent {
	return &agent.HTTPAgent{
		MValues: agent.MetricsValues{
			Counters: make(map[string]int64),
			Gauges:   make(map[string]float64),
		},
		BaseURL:    "http://" + address,
		HTTPClient: resty.New(),
	}
}
