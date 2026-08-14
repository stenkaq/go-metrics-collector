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
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
}
type ServerConfig struct {
	Address         string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
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
	storeInterval := flag.Int("i", 300, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	fileStoragePath := flag.String("f", "/tmp/log.json", "путь до файла, куда сохраняются текущие значения")
	restore := flag.Bool("r", true, "определяет, загружать ли ранее сохранённые значения из указанного файла при старте сервера.")

	flag.Parse()

	err := env.Parse(&envCfg)
	if err != nil {
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	if envCfg.Address != nil {
		address = envCfg.Address
	}
	if envCfg.FileStoragePath != nil {
		fileStoragePath = envCfg.FileStoragePath
	}
	if envCfg.StoreInterval != nil {
		storeInterval = envCfg.StoreInterval
	}
	if envCfg.Restore != nil {
		restore = envCfg.Restore
	}

	return ServerConfig{
		Address:         *address,
		FileStoragePath: *fileStoragePath,
		StoreInterval:   time.Duration(*storeInterval) * time.Second,
		Restore:         *restore,
	}
}
