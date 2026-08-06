package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type AgentEnvConfig struct {
	Address        *string `env:"ADDRESS"`
	ReportInterval *int    `env:"REPORT_INTERVAL"`
	PollInterval   *int    `env:"POLL_INTERVAL"`
}

type ServerEnvConfig struct {
	Address *string `env:"ADDRESS"`
}
type ServerConfig struct {
	Address string
}

func ParseAgentConfig() AgentConfig {
	var address *string
	var report *int
	var poll *int

	var envCfg AgentEnvConfig

	err := env.Parse(&envCfg)
	if err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	address = flag.String("a", "localhost:8080", "адрес сервера")
	report = flag.Int("r", 10, "интервал отправки метрик (сек.)")
	poll = flag.Int("p", 2, "интервал опроса метрик (сек.)")
	flag.Parse()

	if envCfg.Address != nil {
		address = envCfg.Address
	}

	if envCfg.PollInterval != nil {
		poll = envCfg.PollInterval
	}

	if envCfg.ReportInterval != nil {
		report = envCfg.ReportInterval
	}

	return AgentConfig{
		Address:        *address,
		ReportInterval: time.Duration(*report) * time.Second,
		PollInterval:   time.Duration(*poll) * time.Second,
	}
}

func ParseServerConfig() ServerConfig {
	var envCfg ServerEnvConfig

	address := flag.String("a", "localhost:8080", "адрес сервера")

	flag.Parse()

	err := env.Parse(&envCfg)
	if err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	if envCfg.Address != nil {
		address = envCfg.Address
	}

	return ServerConfig{
		Address: *address,
	}
}
