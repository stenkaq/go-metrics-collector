package main

import (
	"go-metrics-collector/internal/app"
	"go-metrics-collector/internal/config"
)

func main() {
	cfg := config.ParseServerConfig()

	router := app.NewMetricsRouter()

	router.Run(cfg.Address)
}
