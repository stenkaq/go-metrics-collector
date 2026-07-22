package main

import (
	"context"
	"go-metrics-collector/internal/app"
	"go-metrics-collector/internal/config"
	"log"
	"sync"
	"time"
)

func scheduleFunc(ctx context.Context, interval time.Duration, fn func(), wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(interval)

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	agent := app.NewAgent(agentConfig.Address)

	log.Println("Агент успешно запущен")

	wg.Add(1)
	go scheduleFunc(ctx, agentConfig.PollInterval, func() {
		agent.CollectMetrics()
	}, &wg)

	wg.Add(1)
	go scheduleFunc(ctx, agentConfig.ReportInterval, func() {
		agent.SendMetrics()
	}, &wg)

	wg.Wait()
}
