package middleware

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LogResponse(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
		c.Writer = blw
		c.Next()

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
