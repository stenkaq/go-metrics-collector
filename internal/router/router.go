package router

import (
	"go-metrics-collector/internal/handler"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Metrics handler.MetricsHandler
}

func New(handlers Handlers) *gin.Engine {
	router := gin.Default()

	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.Use()
	router.POST("/update/:type/:name/:value", func(c *gin.Context) {
		handlers.Metrics.UpdateMetric(c.Writer, c.Request)
	})

	return router
}

func rejectUncleanPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPath := c.Request.URL.Path
		cleanPath := path.Clean(rawPath)

		if cleanPath != rawPath && cleanPath+"/" != rawPath {
			c.Abort()
			c.String(http.StatusNotFound, "Неверный тип")
			return
		}

		c.Next()
	}
}
