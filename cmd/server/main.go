package main

import (
	"go-metrics-collector/internal/app"
	"net/http"
)

func main() {
	httpHandler := app.New()

	http.ListenAndServe("localhost:8080", httpHandler)
}
