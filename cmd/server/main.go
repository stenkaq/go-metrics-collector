package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"go-metrics-collector/internal/app"
	"go-metrics-collector/internal/config"
)

func main() {
	cfg := config.ParseServerConfig()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	metricsApp, err := app.NewMetricsApp(ctx, cfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации сервера: %v", err)
	}

	if err := metricsApp.Run(ctx); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
