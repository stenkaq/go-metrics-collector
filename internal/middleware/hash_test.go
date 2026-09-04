package middleware_test

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-collector/internal/middleware"

	"github.com/gin-gonic/gin"
)

func sign(key, body string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(body))

	return hex.EncodeToString(h.Sum(nil))
}

func TestHashRequest(t *testing.T) {
	const (
		key  = "secret"
		body = `{"id":"Alloc","type":"gauge","value":1}`
	)

	tests := []struct {
		name       string
		key        string
		hash       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "пропускает запрос с верным хешем",
			key:        key,
			hash:       sign(key, body),
			wantStatus: http.StatusOK,
			wantBody:   body,
		},
		{
			name:       "отклоняет запрос с чужим хешем",
			key:        key,
			hash:       sign("другой ключ", body),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "отклоняет хеш, который не является hex",
			key:        key,
			hash:       "не-hex",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "не проверяет ничего без ключа у сервера",
			key:        "",
			hash:       sign("любой", body),
			wantStatus: http.StatusOK,
			wantBody:   body,
		},
		{
			name:       "пропускает запрос без заголовка",
			key:        key,
			hash:       "",
			wantStatus: http.StatusOK,
			wantBody:   body,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(middleware.HashRequest(tt.key))
			// хендлер дочитывает тело — проверяем, что middleware его вернул
			router.POST("/update/", func(ctx *gin.Context) {
				read, err := io.ReadAll(ctx.Request.Body)
				if err != nil {
					t.Errorf("не удалось прочитать тело в хендлере: %v", err)
				}

				ctx.String(http.StatusOK, string(read))
			})

			request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
			if tt.hash != "" {
				request.Header.Set(middleware.HashHeader, tt.hash)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("статус = %d, ожидали %d", recorder.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && recorder.Body.String() != tt.wantBody {
				t.Errorf("тело = %q, ожидали %q", recorder.Body.String(), tt.wantBody)
			}
		})
	}
}

// Агент считает хеш от несжатого тела, а потом жмёт его,
// поэтому HashRequest обязан стоять после GzipRequest.
func TestHashRequestAfterGzip(t *testing.T) {
	const (
		key  = "secret"
		body = `{"id":"Alloc","type":"gauge","value":1}`
	)

	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte(body)); err != nil {
		t.Fatalf("не удалось сжать тело: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("не удалось закрыть gzip.Writer: %v", err)
	}

	router := gin.New()
	router.Use(middleware.GzipRequest(), middleware.HashRequest(key))
	router.POST("/update/", func(ctx *gin.Context) {
		read, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			t.Errorf("не удалось прочитать тело в хендлере: %v", err)
		}

		ctx.String(http.StatusOK, string(read))
	})

	request := httptest.NewRequest(http.MethodPost, "/update/", &buf)
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set(middleware.HashHeader, sign(key, body))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидали %d, тело: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	if recorder.Body.String() != body {
		t.Errorf("тело = %q, ожидали %q", recorder.Body.String(), body)
	}
}
