package router

import (
	"net/http"
	"path"

	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handlers struct {
	Metrics handler.MetricsHandler
	DB      handler.DBHandler
}

func New(handlers Handlers, logger *zap.Logger, updateMiddlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()

	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use(
		middleware.LogRequest(logger),
		middleware.LogResponse(logger),
		middleware.GzipRequest(),
		middleware.GzipResponse(),
		gin.Recovery(),
		rejectUncleanPath(),
	)

	update := router.Group("/update", updateMiddlewares...)
	update.POST("/", handlers.Metrics.UpdateMetricV2)
	update.POST("/:type/:name/:value", handlers.Metrics.UpdateMetric)

	updates := router.Group("/updates", updateMiddlewares...)
	updates.POST("/", handlers.Metrics.UpdateMetrics)

	router.POST("/value/", handlers.Metrics.GetMetricV2)
	router.GET("/value/:type/:name", handlers.Metrics.GetMetric)
	router.GET("/ping", handlers.DB.Ping)
	router.GET("/", handlers.Metrics.GetMetrics)

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
