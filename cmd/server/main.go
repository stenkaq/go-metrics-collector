package main

import (
	"log"

	"go-metrics-collector/internal/app"
	"go-metrics-collector/internal/config"
)

func main() {
	cfg := config.ParseServerConfig()

	router := app.NewMetricsRouter()

	if err := router.Run(cfg.Address); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
