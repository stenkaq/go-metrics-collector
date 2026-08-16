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
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

type ServerEnvConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

type ServerConfig struct {
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

func ParseAgentConfig() AgentConfig {
	var raw AgentEnvConfig

	flag.StringVar(&raw.Address, "a", "localhost:8080", "адрес сервера")
	flag.IntVar(&raw.ReportInterval, "r", 10, "интервал отправки метрик (сек.)")
	flag.IntVar(&raw.PollInterval, "p", 2, "интервал опроса метрик (сек.)")
	flag.Parse()

	if err := env.Parse(&raw); err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	return AgentConfig{
		Address:        raw.Address,
		ReportInterval: time.Duration(raw.ReportInterval) * time.Second,
		PollInterval:   time.Duration(raw.PollInterval) * time.Second,
	}
}

func ParseServerConfig() ServerConfig {
	var raw ServerEnvConfig

	flag.StringVar(&raw.Address, "a", "localhost:8080", "адрес сервера")
	flag.IntVar(&raw.StoreInterval, "i", 300, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&raw.FileStoragePath, "f", "/tmp/log.json", "путь до файла, куда сохраняются текущие значения")
	flag.BoolVar(&raw.Restore, "r", true, "определяет, загружать ли ранее сохранённые значения из указанного файла при старте сервера.")
	flag.Parse()

	if err := env.Parse(&raw); err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	return ServerConfig{
		Address:         raw.Address,
		FileStoragePath: raw.FileStoragePath,
		StoreInterval:   time.Duration(raw.StoreInterval) * time.Second,
		Restore:         raw.Restore,
	}
}
