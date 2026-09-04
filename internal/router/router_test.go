package router_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-collector/internal/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

type handlerCall struct {
	method string
	params gin.Params
}

// metricsHandlerStub записывает вызовы и отвечает 200 OK, что бы проверять
// только маршрутизацию.
type metricsHandlerStub struct {
	calls []handlerCall
}

func (h *metricsHandlerStub) record(c *gin.Context, method string) {
	h.calls = append(h.calls, handlerCall{method: method, params: c.Params})
	io.WriteString(c.Writer, method)
}

func (h *metricsHandlerStub) UpdateMetrics(c *gin.Context)  { h.record(c, "UpdateMetrics") }
func (h *metricsHandlerStub) UpdateMetric(c *gin.Context)   { h.record(c, "UpdateMetric") }
func (h *metricsHandlerStub) UpdateMetricV2(c *gin.Context) { h.record(c, "UpdateMetricV2") }
func (h *metricsHandlerStub) GetMetric(c *gin.Context)      { h.record(c, "GetMetric") }
func (h *metricsHandlerStub) GetMetricV2(c *gin.Context)    { h.record(c, "GetMetricV2") }
func (h *metricsHandlerStub) GetMetrics(c *gin.Context)     { h.record(c, "GetMetrics") }

// dbHandlerStub отвечает на /ping тем же способом, что и остальные заглушки.
type dbHandlerStub struct {
	calls []handlerCall
}

func (h *dbHandlerStub) Ping(c *gin.Context) {
	h.calls = append(h.calls, handlerCall{method: "Ping", params: c.Params})
	io.WriteString(c.Writer, "Ping")
}

func newTestRouter(updateMiddlewares ...gin.HandlerFunc) (*gin.Engine, *metricsHandlerStub, *dbHandlerStub) {
	metrics := &metricsHandlerStub{}
	db := &dbHandlerStub{}
	engine := router.New(router.Handlers{Metrics: metrics, DB: db}, zap.NewNop(), "", updateMiddlewares...)

	return engine, metrics, db
}

func doRequest(t *testing.T, engine *gin.Engine, method, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(method, target, body))

	return recorder
}

func TestRouterRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		wantMethod string
		wantParams map[string]string
	}{
		{
			name:       "routes a JSON update",
			method:     http.MethodPost,
			target:     "/update/",
			wantMethod: "UpdateMetricV2",
		},
		{
			name:       "routes a batch update",
			method:     http.MethodPost,
			target:     "/updates/",
			wantMethod: "UpdateMetrics",
		},
		{
			name:       "routes a path update",
			method:     http.MethodPost,
			target:     "/update/gauge/Alloc/12.5",
			wantMethod: "UpdateMetric",
			wantParams: map[string]string{"type": "gauge", "name": "Alloc", "value": "12.5"},
		},
		{
			name:       "routes a JSON value request",
			method:     http.MethodPost,
			target:     "/value/",
			wantMethod: "GetMetricV2",
		},
		{
			name:       "routes a path value request",
			method:     http.MethodGet,
			target:     "/value/counter/PollCount",
			wantMethod: "GetMetric",
			wantParams: map[string]string{"type": "counter", "name": "PollCount"},
		},
		{
			name:       "routes the index",
			method:     http.MethodGet,
			target:     "/",
			wantMethod: "GetMetrics",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			engine, metrics, _ := newTestRouter()

			recorder := doRequest(t, engine, test.method, test.target, nil)

			if recorder.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if recorder.Body.String() != test.wantMethod {
				t.Errorf("body = %q, want %q", recorder.Body.String(), test.wantMethod)
			}

			if len(metrics.calls) != 1 {
				t.Fatalf("вызовов хендлера = %d, want 1", len(metrics.calls))
			}

			call := metrics.calls[0]
			if call.method != test.wantMethod {
				t.Fatalf("вызван %s, want %s", call.method, test.wantMethod)
			}

			for key, want := range test.wantParams {
				if got := call.params.ByName(key); got != want {
					t.Errorf("параметр %s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestRouterRoutesPingToDBHandler(t *testing.T) {
	engine, metrics, db := newTestRouter()

	recorder := doRequest(t, engine, http.MethodGet, "/ping", nil)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if len(db.calls) != 1 {
		t.Fatalf("вызовов DB хендлера = %d, want 1", len(db.calls))
	}
	if len(metrics.calls) != 0 {
		t.Errorf("вызовов хендлера метрик = %d, want 0", len(metrics.calls))
	}
}

func TestRouterUpdateMiddlewareRunsOnlyOnUpdates(t *testing.T) {
	var calls int
	countMiddleware := func(c *gin.Context) {
		calls++
		c.Next()
	}

	engine, _, _ := newTestRouter(countMiddleware)

	doRequest(t, engine, http.MethodPost, "/update/", nil)
	doRequest(t, engine, http.MethodPost, "/update/gauge/Alloc/12.5", nil)
	if calls != 2 {
		t.Fatalf("вызовов middleware на /update = %d, want 2", calls)
	}

	doRequest(t, engine, http.MethodGet, "/", nil)
	doRequest(t, engine, http.MethodGet, "/value/counter/PollCount", nil)
	if calls != 2 {
		t.Errorf("вызовов middleware = %d, middleware не должно срабатывать вне /update", calls)
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
			engine, metrics, _ := newTestRouter()

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
	engine, _, _ := newTestRouter()

	recorder := doRequest(t, engine, http.MethodGet, "/value/counter/PollCount/", nil)

	if recorder.Code == http.StatusMovedPermanently || recorder.Code == http.StatusTemporaryRedirect {
		t.Errorf("status = %d, редиректы должны быть выключены", recorder.Code)
	}
}
