package app

import (
	"go-metrics-collector/internal/agent"
	"go-metrics-collector/internal/config"

	"github.com/go-resty/resty/v2"
)

func NewAgent(cfg config.AgentConfig) agent.Agent {
	return &agent.HTTPAgent{
		MValues: agent.MetricsValues{
			Counters: make(map[string]int64),
			Gauges:   make(map[string]float64),
		},
		BaseURL:    "http://" + cfg.Address,
		HTTPClient: resty.New().OnBeforeRequest(agent.HashBody(cfg.Key)).OnBeforeRequest(agent.GzipRequest()),
	}
}
