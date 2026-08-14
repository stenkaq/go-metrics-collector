package main

import (
	"log"

	"go-metrics-collector/internal/app"
	"go-metrics-collector/internal/config"
)

func main() {
	cfg := config.ParseServerConfig()

	metricsApp, err := app.NewMetricsApp(cfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации сервера: %v", err)
	}

	if err := metricsApp.Run(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
