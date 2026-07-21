package main

import (
	"go-metrics-collector/internal/app"
)

func main() {
	router := app.NewMetricsRouter()

	router.Run("localhost:8080")
}
