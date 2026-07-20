package main

import (
	"context"
	"go-metrics-collector/internal/app"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	agent := app.NewAgent()
	log.Println("Агент успешно запущен")

	wg.Add(1)
	go scheduleFunc(ctx, 2*time.Second, func() {
		agent.CollectMetrics()
	}, &wg)

	wg.Add(1)
	go scheduleFunc(ctx, 10*time.Second, func() {
		agent.SendMetrics()
	}, &wg)

	wg.Wait()
}
