package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MetricsDump func(ctx context.Context) error

func DumpMetrics(dump MetricsDump, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if err := dump(c.Request.Context()); err != nil {
			log.Error("не удалось сохранить метрики на диск", zap.Error(err))
		}
	}
}
