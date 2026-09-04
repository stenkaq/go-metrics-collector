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
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

type ServerConfig struct {
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

func ParseAgentConfig() AgentConfig {
	var cfg AgentEnvConfig

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "адрес сервера")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "интервал отправки метрик (сек.)")
	flag.IntVar(&cfg.PollInterval, "p", 2, "интервал опроса метрик (сек.)")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	return AgentConfig{
		Address:        cfg.Address,
		ReportInterval: time.Duration(cfg.ReportInterval) * time.Second,
		PollInterval:   time.Duration(cfg.PollInterval) * time.Second,
	}
}

func ParseServerConfig() ServerConfig {
	var cfg ServerEnvConfig

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "адрес сервера")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/log.json", "путь до файла, куда сохраняются текущие значения")
	flag.BoolVar(&cfg.Restore, "r", true, "определяет, загружать ли ранее сохранённые значения из указанного файла при старте сервера.")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "строка подключения к БД")

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	return ServerConfig{
		Address:         cfg.Address,
		FileStoragePath: cfg.FileStoragePath,
		StoreInterval:   time.Duration(cfg.StoreInterval) * time.Second,
		Restore:         cfg.Restore,
		DatabaseDSN:     cfg.DatabaseDSN,
	}
}
