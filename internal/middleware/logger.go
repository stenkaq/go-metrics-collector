package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LogResponse(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		size := c.Writer.Size()
		if size < 0 {
			size = 0
		}

		log.Info("[RESPONSE]", zap.Int("status code", c.Writer.Status()), zap.Int("size", c.Writer.Size()))
	}
}

func LogRequest(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log.Info("[REQUEST]",
			zap.String("method", c.Request.Method),
			zap.String("URI", c.Request.RequestURI),
			zap.Duration("time taken", time.Since(start)))
	}
}
