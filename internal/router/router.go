package router

import (
	"encoding/json"
	"io"
	"net/http"
	"path"

	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handlers struct {
	Metrics handler.MetricsHandler
}

type MetricsUpdateParams struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
type MetricsValueParams struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}

func New(handlers Handlers, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use(gin.Recovery(), middleware.LogRequest(logger), middleware.LogResponse(logger), rejectUncleanPath())

	router.POST("/update", func(ctx *gin.Context) {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			http.Error(ctx.Writer, "Ошибка при чтении тела запроса", http.StatusBadRequest)
			return
		}

		var body MetricsUpdateParams
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			http.Error(ctx.Writer, "Ошибка при чтении тела запроса", http.StatusBadRequest)
			return
		}

		if body.ID == "" || body.MType == "" || (body.Delta == nil && body.Value == nil) || (body.Delta != nil && body.Value != nil) {
			http.Error(ctx.Writer, "Неверные значения полей в теле запроса", http.StatusBadRequest)
			return
		}

		params := handler.MetricsUpdateParams{
			ID:    body.ID,
			MType: body.MType,
		}

		if body.Value != nil {
			params.Value = *body.Value
		}

		if body.Delta != nil {
			params.Delta = *body.Delta
		}

		handlers.Metrics.UpdateMetricV2(ctx.Writer, params)
	})
	router.POST("/update/:type/:name/:value", func(ctx *gin.Context) {
		handlers.Metrics.UpdateMetric(ctx.Writer, ctx.Param("type"), ctx.Param("name"), ctx.Param("value"))
	})
	router.POST("/value", func(ctx *gin.Context) {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			http.Error(ctx.Writer, "Ошибка при чтении тела запроса", http.StatusBadRequest)
			return
		}

		var body MetricsValueParams
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			http.Error(ctx.Writer, "Ошибка при чтении тела запроса", http.StatusBadRequest)
			return
		}

		handlers.Metrics.GetMetricV2(ctx.Writer, handler.MetricsGetParams{ID: body.ID, MType: body.MType})

	})
	router.GET("/", func(ctx *gin.Context) {
		handlers.Metrics.GetMetrics(ctx.Writer)
	})
	router.GET("/value/:type/:name", func(ctx *gin.Context) {
		handlers.Metrics.GetMetric(ctx.Writer, ctx.Param("type"), ctx.Param("name"))
	})

	return router
}

func rejectUncleanPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPath := c.Request.URL.Path
		cleanPath := path.Clean(rawPath)

		if cleanPath != rawPath && cleanPath+"/" != rawPath {
			c.Abort()
			c.String(http.StatusNotFound, "Ресурс не найден")
			return
		}

		c.Next()
	}
}
