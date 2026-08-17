package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MetricsDump func() error

func DumpMetrics(dump MetricsDump, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if err := dump(); err != nil {
			log.Error("не удалось сохранить метрики на диск", zap.Error(err))
		}
	}
}
