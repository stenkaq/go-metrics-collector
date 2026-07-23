package router

import (
	"net/http"
	"path"

	"go-metrics-collector/internal/handler"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Metrics handler.MetricsHandler
}

func New(handlers Handlers) *gin.Engine {
	router := gin.Default()

	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use(rejectUncleanPath())

	router.POST("/update/:type/:name/:value", func(ctx *gin.Context) {
		handlers.Metrics.UpdateMetric(ctx.Writer, ctx.Param("type"), ctx.Param("name"), ctx.Param("value"))
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
