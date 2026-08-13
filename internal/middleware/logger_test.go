package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-collector/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func newObservedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.InfoLevel)
	return zap.New(core), logs
}

func TestLogResponse(t *testing.T) {
	tests := []struct {
		name       string
		handler    gin.HandlerFunc
		wantStatus int64
		wantSize   int64
	}{
		{
			name: "logs status and size of a written response",
			handler: func(ctx *gin.Context) {
				ctx.String(http.StatusCreated, "hello")
			},
			wantStatus: http.StatusCreated,
			wantSize:   5,
		},
		{
			name:       "logs an empty response",
			handler:    func(ctx *gin.Context) {},
			wantStatus: http.StatusOK,
			wantSize:   -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger, logs := newObservedLogger()

			router := gin.New()
			router.Use(middleware.LogResponse(logger))
			router.GET("/", test.handler)

			router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			entries := logs.FilterMessage("[RESPONSE]").All()
			if len(entries) != 1 {
				t.Fatalf("записей в логе = %d, want 1", len(entries))
			}

			fields := entries[0].ContextMap()
			if got := fields["status code"]; got != test.wantStatus {
				t.Errorf("status code = %v, want %v", got, test.wantStatus)
			}
			if got := fields["size"]; got != test.wantSize {
				t.Errorf("size = %v, want %v", got, test.wantSize)
			}
		})
	}
}

func TestLogRequest(t *testing.T) {
	logger, logs := newObservedLogger()

	router := gin.New()
	router.Use(middleware.LogRequest(logger))
	router.POST("/update/gauge/Alloc/1.5", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	router.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/1.5", nil),
	)

	entries := logs.FilterMessage("[REQUEST]").All()
	if len(entries) != 1 {
		t.Fatalf("записей в логе = %d, want 1", len(entries))
	}

	fields := entries[0].ContextMap()
	if got := fields["method"]; got != http.MethodPost {
		t.Errorf("method = %v, want %v", got, http.MethodPost)
	}
	if got := fields["URI"]; got != "/update/gauge/Alloc/1.5" {
		t.Errorf("URI = %v, want /update/gauge/Alloc/1.5", got)
	}
	if _, exists := fields["time taken"]; !exists {
		t.Error("в логе нет поля «time taken»")
	}
}
