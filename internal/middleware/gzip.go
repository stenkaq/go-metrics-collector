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
	gz         *gzip.Writer
	compressed bool
}

func (w *gzipWriter) Write(data []byte) (int, error) {
	ct := w.Header().Get("Content-Type")
	if strings.Contains(ct, "application/json") ||
		strings.Contains(ct, "text/html") {

		w.compressed = true
		w.Header().Set("Content-Encoding", "gzip")

		return w.gz.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *gzipWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
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

		gw := &gzipWriter{
			ResponseWriter: ctx.Writer,
			gz:             gz,
			compressed:     false,
		}
		ctx.Writer = gw

		ctx.Next()

		if gw.compressed {
			gz.Close()
		}
	}
}
