package router_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-metrics-collector/internal/handler"
	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

type handlerCall struct {
	method       string
	updateV2     handler.MetricsUpdateParams
	getV2        handler.MetricsGetParams
	mType        string
	name         string
	value        string
	responseBody string
}

// metricsHandlerStub записывает вызовы и отвечает 200 OK, что бы проверять
// только маршрутизацию и разбор запроса.
type metricsHandlerStub struct {
	calls []handlerCall
}

func (h *metricsHandlerStub) UpdateMetric(w http.ResponseWriter, mType, name, value string) {
	h.calls = append(h.calls, handlerCall{method: "UpdateMetric", mType: mType, name: name, value: value})
	io.WriteString(w, "UpdateMetric")
}

func (h *metricsHandlerStub) UpdateMetricV2(w http.ResponseWriter, params handler.MetricsUpdateParams) {
	h.calls = append(h.calls, handlerCall{method: "UpdateMetricV2", updateV2: params})
	io.WriteString(w, "UpdateMetricV2")
}

func (h *metricsHandlerStub) GetMetric(w http.ResponseWriter, mType, name string) {
	h.calls = append(h.calls, handlerCall{method: "GetMetric", mType: mType, name: name})
	io.WriteString(w, "GetMetric")
}

func (h *metricsHandlerStub) GetMetricV2(w http.ResponseWriter, params handler.MetricsGetParams) {
	h.calls = append(h.calls, handlerCall{method: "GetMetricV2", getV2: params})
	io.WriteString(w, "GetMetricV2")
}

func (h *metricsHandlerStub) GetMetrics(w http.ResponseWriter) {
	h.calls = append(h.calls, handlerCall{method: "GetMetrics"})
	io.WriteString(w, "GetMetrics")
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("тело недоступно")
}

func newTestRouter() (*gin.Engine, *metricsHandlerStub) {
	metrics := &metricsHandlerStub{}
	engine := router.New(router.Handlers{Metrics: metrics}, zap.NewNop())

	return engine, metrics
}

func doRequest(t *testing.T, engine *gin.Engine, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(method, target, body))

	return recorder
}

func TestRouterUpdateJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       io.Reader
		wantStatus int
		wantParams handler.MetricsUpdateParams
		wantCall   bool
	}{
		{
			name:       "routes a counter update",
			body:       strings.NewReader(`{"id":"PollCount","type":"counter","delta":42}`),
			wantStatus: http.StatusOK,
			wantParams: handler.MetricsUpdateParams{ID: "PollCount", MType: models.Counter, Delta: 42},
			wantCall:   true,
		},
		{
			name:       "routes a gauge update",
			body:       strings.NewReader(`{"id":"Alloc","type":"gauge","value":12.5}`),
			wantStatus: http.StatusOK,
			wantParams: handler.MetricsUpdateParams{ID: "Alloc", MType: models.Gauge, Value: 12.5},
			wantCall:   true,
		},
		{
			name:       "rejects a broken body",
			body:       strings.NewReader(`{"id":`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects an unreadable body",
			body:       errReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects a missing id",
			body:       strings.NewReader(`{"type":"counter","delta":42}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects a missing type",
			body:       strings.NewReader(`{"id":"PollCount","delta":42}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects a missing value",
			body:       strings.NewReader(`{"id":"PollCount","type":"counter"}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects both delta and value",
			body:       strings.NewReader(`{"id":"PollCount","type":"counter","delta":42,"value":12.5}`),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine, metrics := newTestRouter()

			recorder := doRequest(t, engine, http.MethodPost, "/update", test.body)

			if recorder.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, test.wantStatus)
			}

			if !test.wantCall {
				if len(metrics.calls) != 0 {
					t.Fatalf("вызовов хендлера = %d, want 0", len(metrics.calls))
				}
				return
			}

			if len(metrics.calls) != 1 {
				t.Fatalf("вызовов хендлера = %d, want 1", len(metrics.calls))
			}
			call := metrics.calls[0]
			if call.method != "UpdateMetricV2" {
				t.Fatalf("вызван %s, want UpdateMetricV2", call.method)
			}
			if call.updateV2 != test.wantParams {
				t.Errorf("params = %+v, want %+v", call.updateV2, test.wantParams)
			}
		})
	}
}

func TestRouterUpdatePath(t *testing.T) {
	engine, metrics := newTestRouter()

	recorder := doRequest(t, engine, http.MethodPost, "/update/gauge/Alloc/12.5", nil)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if len(metrics.calls) != 1 {
		t.Fatalf("вызовов хендлера = %d, want 1", len(metrics.calls))
	}

	call := metrics.calls[0]
	if call.method != "UpdateMetric" {
		t.Fatalf("вызван %s, want UpdateMetric", call.method)
	}
	if call.mType != "gauge" || call.name != "Alloc" || call.value != "12.5" {
		t.Errorf("параметры = %+v, want gauge/Alloc/12.5", call)
	}
}

func TestRouterValueJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       io.Reader
		wantStatus int
		wantParams handler.MetricsGetParams
		wantCall   bool
	}{
		{
			name:       "routes a value request",
			body:       strings.NewReader(`{"id":"Alloc","type":"gauge"}`),
			wantStatus: http.StatusOK,
			wantParams: handler.MetricsGetParams{ID: "Alloc", MType: models.Gauge},
			wantCall:   true,
		},
		{
			name:       "rejects a broken body",
			body:       strings.NewReader(`{"id":`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects an unreadable body",
			body:       errReader{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine, metrics := newTestRouter()

			recorder := doRequest(t, engine, http.MethodPost, "/value", test.body)

			if recorder.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, test.wantStatus)
			}

			if !test.wantCall {
				if len(metrics.calls) != 0 {
					t.Fatalf("вызовов хендлера = %d, want 0", len(metrics.calls))
				}
				return
			}

			if len(metrics.calls) != 1 {
				t.Fatalf("вызовов хендлера = %d, want 1", len(metrics.calls))
			}
			call := metrics.calls[0]
			if call.method != "GetMetricV2" {
				t.Fatalf("вызван %s, want GetMetricV2", call.method)
			}
			if call.getV2 != test.wantParams {
				t.Errorf("params = %+v, want %+v", call.getV2, test.wantParams)
			}
		})
	}
}

func TestRouterValuePath(t *testing.T) {
	engine, metrics := newTestRouter()

	recorder := doRequest(t, engine, http.MethodGet, "/value/counter/PollCount", nil)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if len(metrics.calls) != 1 {
		t.Fatalf("вызовов хендлера = %d, want 1", len(metrics.calls))
	}

	call := metrics.calls[0]
	if call.method != "GetMetric" {
		t.Fatalf("вызван %s, want GetMetric", call.method)
	}
	if call.mType != "counter" || call.name != "PollCount" {
		t.Errorf("параметры = %+v, want counter/PollCount", call)
	}
}

func TestRouterIndex(t *testing.T) {
	engine, metrics := newTestRouter()

	recorder := doRequest(t, engine, http.MethodGet, "/", nil)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "GetMetrics" {
		t.Errorf("body = %q, want %q", recorder.Body.String(), "GetMetrics")
	}
	if len(metrics.calls) != 1 || metrics.calls[0].method != "GetMetrics" {
		t.Fatalf("вызовы хендлера = %+v, want один GetMetrics", metrics.calls)
	}
}

func TestRouterRejectsUncleanPath(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantCalls  int
	}{
		{
			name:       "rejects a doubled slash",
			target:     "/value//counter/PollCount",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "rejects a dot segment",
			target:     "/value/counter/../counter/PollCount",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "allows a trailing slash through the middleware",
			target:     "/value/counter/PollCount/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "allows a clean path",
			target:     "/value/counter/PollCount",
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine, metrics := newTestRouter()

			recorder := doRequest(t, engine, http.MethodGet, test.target, nil)

			if recorder.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if len(metrics.calls) != test.wantCalls {
				t.Errorf("вызовов хендлера = %d, want %d", len(metrics.calls), test.wantCalls)
			}
		})
	}
}

func TestRouterDoesNotRedirect(t *testing.T) {
	engine, _ := newTestRouter()

	recorder := doRequest(t, engine, http.MethodGet, "/value/counter/PollCount/", nil)

	if recorder.Code == http.StatusMovedPermanently || recorder.Code == http.StatusTemporaryRedirect {
		t.Errorf("status = %d, редиректы должны быть выключены", recorder.Code)
	}
}
