package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-metrics-collector/internal/app"
	"go-metrics-collector/internal/config"
)

func scheduleFunc(ctx context.Context, interval time.Duration, fn func(), wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fn()
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	agentConfig := config.ParseAgentConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	agent := app.NewAgent(agentConfig.Address)

	log.Println("Агент успешно запущен")

	wg.Add(2)
	go scheduleFunc(ctx, agentConfig.PollInterval, func() {
		agent.CollectMetrics()
	}, &wg)

	go scheduleFunc(ctx, agentConfig.ReportInterval, func() {
		agent.SendMetrics()
	}, &wg)

	wg.Wait()
	log.Println("Агент остановлен")
}
