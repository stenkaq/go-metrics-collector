package config

import (
	"flag"
	"time"
)

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

type ServerConfig struct {
	Address string
}

func ParseAgentConfig() AgentConfig {
	address := flag.String("a", "localhost:8080", "адрес сервера")
	report := flag.Int("r", 10, "интервал отправки метрик (сек.)")
	poll := flag.Int("p", 2, "интервал опроса метрик (сек.)")

	flag.Parse()

	return AgentConfig{
		Address: *address,
		ReportInterval: time.Duration(*report) * time.Second,
		PollInterval: time.Duration(*poll) * time.Second,
	}
}

func ParseServerConfig() ServerConfig {
	address := flag.String("a", "localhost:8080", "адрес сервера")

	flag.Parse()

	return ServerConfig{
		Address: *address,
	}
}
