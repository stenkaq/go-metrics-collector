package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	gz *gzip.Writer
}

func (w *gzipWriter) Write(data []byte) (int, error) {
	return w.gz.Write(data)
}

func (w *gzipWriter) WriteString(s string) (int, error) {
	return w.gz.Write([]byte(s))
}

func GzipRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.Contains(ctx.GetHeader("Content-Encoding"), "gzip") {
			ctx.Next()
			return
		}

		gz, err := gzip.NewReader(ctx.Request.Body)
		if err != nil {
			http.Error(ctx.Writer, fmt.Sprintf("Ошибка при создании reader gz : %v", err), http.StatusInternalServerError)
			ctx.Abort()
			return
		}
		defer gz.Close()

		ctx.Request.Body = gz
		ctx.Next()
	}
}

func GzipResponse() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
			ctx.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(ctx.Writer, gzip.BestSpeed)
		if err != nil {
			http.Error(ctx.Writer, fmt.Sprintf("Ошибка при создании writer gz : %v", err), http.StatusInternalServerError)
			ctx.Abort()
			return
		}
		defer gz.Close()

		ctx.Writer.Header().Set("Content-Encoding", "gzip")
		ctx.Writer = &gzipWriter{ResponseWriter: ctx.Writer, gz: gz}
		ctx.Next()
	}
}
