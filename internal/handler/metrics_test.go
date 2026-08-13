package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-collector/internal/handler"
	models "go-metrics-collector/internal/model"
	"go-metrics-collector/internal/service"
)

type serviceCall struct {
	metricType string
	name       string
	counter    int64
	gauge      float64
}

type metricsServiceStub struct {
	calls *serviceCalls

	metric  models.Metrics
	exists  bool
	metrics map[string]models.Metrics
}

type serviceCalls struct {
	values []serviceCall
	gets   []serviceCall
}

func newMetricsServiceStub() (*metricsServiceStub, *serviceCalls) {
	calls := &serviceCalls{}
	return &metricsServiceStub{calls: calls}, calls
}

func (s *metricsServiceStub) UpdateCounterMetricValue(params service.UpdateCounterMetricValueParams) {
	s.calls.values = append(s.calls.values, serviceCall{
		metricType: params.Type,
		name:       params.Name,
		counter:    params.Value,
	})
}

func (s *metricsServiceStub) UpdateGaugeMetricValue(params service.UpdateGaugeMetricValueParams) {
	s.calls.values = append(s.calls.values, serviceCall{
		metricType: params.Type,
		name:       params.Name,
		gauge:      params.Value,
	})
}

func (s *metricsServiceStub) GetMetric(mType, name string) (models.Metrics, bool) {
	s.calls.gets = append(s.calls.gets, serviceCall{metricType: mType, name: name})
	return s.metric, s.exists
}

func (s *metricsServiceStub) GetMetrics() map[string]models.Metrics {
	return s.metrics
}

func counterMetric(name string, delta int64) models.Metrics {
	return models.Metrics{ID: name, MType: models.Counter, Delta: &delta}
}

func gaugeMetric(name string, value float64) models.Metrics {
	return models.Metrics{ID: name, MType: models.Gauge, Value: &value}
}

func TestMetricsHandlerUpdateV2(t *testing.T) {
	tests := []struct {
		name            string
		params          handler.MetricsUpdateParams
		wantStatus      int
		wantBody        string
		wantContentType string
		wantCalls       []serviceCall
	}{
		{
			name: "updates counter",
			params: handler.MetricsUpdateParams{
				ID:    "PollCount",
				MType: "counter",
				Delta: 42,
			},
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain",
			wantCalls: []serviceCall{{
				metricType: "counter",
				name:       "PollCount",
				counter:    42,
			}},
		},
		{
			name: "updates gauge",
			params: handler.MetricsUpdateParams{
				ID:    "Alloc",
				MType: "gauge",
				Value: 12.5,
			},
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain",
			wantCalls: []serviceCall{{
				metricType: "gauge",
				name:       "Alloc",
				gauge:      12.5,
			}},
		},
		{
			name: "ignores gauge value for counter",
			params: handler.MetricsUpdateParams{
				ID:    "PollCount",
				MType: "counter",
				Delta: 7,
				Value: 12.5,
			},
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain",
			wantCalls: []serviceCall{{
				metricType: "counter",
				name:       "PollCount",
				counter:    7,
			}},
		},
		{
			name: "rejects unsupported metric type",
			params: handler.MetricsUpdateParams{
				ID:    "RequestTime",
				MType: "timer",
				Delta: 1,
			},
			wantStatus:      http.StatusBadRequest,
			wantBody:        "Неизвестный тип метрики\n",
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			name: "rejects empty metric type",
			params: handler.MetricsUpdateParams{
				ID: "RequestTime",
			},
			wantStatus:      http.StatusBadRequest,
			wantBody:        "Неизвестный тип метрики\n",
			wantContentType: "text/plain; charset=utf-8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsService, calls := newMetricsServiceStub()
			h := handler.NewMetricsHandler(metricsService)

			recorder := httptest.NewRecorder()
			h.UpdateMetricV2(recorder, test.params)

			body := checkResponse(t, recorder, test.wantStatus, test.wantBody, test.wantContentType)
			_ = body

			checkCalls(t, calls.values, test.wantCalls)
		})
	}
}

func TestMetricsHandlerUpdate(t *testing.T) {
	tests := []struct {
		name            string
		mType           string
		metricName      string
		value           string
		wantStatus      int
		wantBody        string
		wantContentType string
		wantCalls       []serviceCall
	}{
		{
			name:            "updates counter",
			mType:           "counter",
			metricName:      "PollCount",
			value:           "42",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain",
			wantCalls: []serviceCall{{
				metricType: "counter",
				name:       "PollCount",
				counter:    42,
			}},
		},
		{
			name:            "updates gauge",
			mType:           "gauge",
			metricName:      "Alloc",
			value:           "12.5",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain",
			wantCalls: []serviceCall{{
				metricType: "gauge",
				name:       "Alloc",
				gauge:      12.5,
			}},
		},
		{
			name:            "rejects empty name",
			mType:           "gauge",
			metricName:      "   ",
			value:           "12.5",
			wantStatus:      http.StatusNotFound,
			wantBody:        "Пустое имя метрики\n",
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			name:            "rejects non integer counter",
			mType:           "counter",
			metricName:      "PollCount",
			value:           "12.5",
			wantStatus:      http.StatusBadRequest,
			wantBody:        "Неверное значение метрики\n",
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			name:            "rejects non numeric gauge",
			mType:           "gauge",
			metricName:      "Alloc",
			value:           "big",
			wantStatus:      http.StatusBadRequest,
			wantBody:        "Неверное значение метрики\n",
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			name:            "rejects unsupported metric type",
			mType:           "timer",
			metricName:      "RequestTime",
			value:           "1",
			wantStatus:      http.StatusBadRequest,
			wantBody:        "Неизвестный тип метрики\n",
			wantContentType: "text/plain; charset=utf-8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsService, calls := newMetricsServiceStub()
			h := handler.NewMetricsHandler(metricsService)

			recorder := httptest.NewRecorder()
			h.UpdateMetric(recorder, test.mType, test.metricName, test.value)

			checkResponse(t, recorder, test.wantStatus, test.wantBody, test.wantContentType)
			checkCalls(t, calls.values, test.wantCalls)
		})
	}
}

func TestMetricsHandlerGetMetric(t *testing.T) {
	tests := []struct {
		name            string
		metric          models.Metrics
		exists          bool
		wantStatus      int
		wantBody        string
		wantContentType string
	}{
		{
			name:            "returns counter value",
			metric:          counterMetric("PollCount", 42),
			exists:          true,
			wantStatus:      http.StatusOK,
			wantBody:        "42",
			wantContentType: "text/html; charset=utf-8",
		},
		{
			name:            "returns gauge value",
			metric:          gaugeMetric("Alloc", 12.5),
			exists:          true,
			wantStatus:      http.StatusOK,
			wantBody:        "12.5",
			wantContentType: "text/html; charset=utf-8",
		},
		{
			name:            "returns 404 for unknown metric",
			wantStatus:      http.StatusNotFound,
			wantBody:        "Неизвестная метрика\n",
			wantContentType: "text/plain; charset=utf-8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsService, calls := newMetricsServiceStub()
			metricsService.metric = test.metric
			metricsService.exists = test.exists

			h := handler.NewMetricsHandler(metricsService)

			recorder := httptest.NewRecorder()
			h.GetMetric(recorder, "gauge", "Alloc")

			checkResponse(t, recorder, test.wantStatus, test.wantBody, test.wantContentType)

			if len(calls.gets) != 1 {
				t.Fatalf("service lookups = %d, want 1", len(calls.gets))
			}
			if got := calls.gets[0]; got.metricType != "gauge" || got.name != "Alloc" {
				t.Errorf("service lookup = %#v, want type gauge name Alloc", got)
			}
		})
	}
}

func TestMetricsHandlerGetMetricEscapesValue(t *testing.T) {
	metricsService, _ := newMetricsServiceStub()
	metricsService.metric = models.Metrics{ID: "Weird", MType: "<script>"}
	value := 1.0
	metricsService.metric.Value = &value
	metricsService.exists = true

	h := handler.NewMetricsHandler(metricsService)

	recorder := httptest.NewRecorder()
	h.GetMetric(recorder, "gauge", "Weird")

	if body := recorder.Body.String(); body != "1" {
		t.Errorf("body = %q, want %q", body, "1")
	}
}

func TestMetricsHandlerGetMetricV2(t *testing.T) {
	tests := []struct {
		name            string
		params          handler.MetricsGetParams
		metric          models.Metrics
		exists          bool
		wantStatus      int
		wantContentType string
		wantResponse    handler.MetricsGetResponse
		wantBody        string
	}{
		{
			name:            "returns counter",
			params:          handler.MetricsGetParams{ID: "PollCount", MType: models.Counter},
			metric:          counterMetric("PollCount", 42),
			exists:          true,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json; charset=utf-8",
			wantResponse: handler.MetricsGetResponse{
				ID:    "PollCount",
				MType: models.Counter,
				Delta: ptr[int64](42),
			},
		},
		{
			name:            "returns gauge",
			params:          handler.MetricsGetParams{ID: "Alloc", MType: models.Gauge},
			metric:          gaugeMetric("Alloc", 12.5),
			exists:          true,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json; charset=utf-8",
			wantResponse: handler.MetricsGetResponse{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: ptr(12.5),
			},
		},
		{
			name:            "returns 404 for unknown metric",
			params:          handler.MetricsGetParams{ID: "Nope", MType: models.Gauge},
			wantStatus:      http.StatusNotFound,
			wantContentType: "text/plain; charset=utf-8",
			wantBody:        "Неизвестная метрика\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsService, _ := newMetricsServiceStub()
			metricsService.metric = test.metric
			metricsService.exists = test.exists

			h := handler.NewMetricsHandler(metricsService)

			recorder := httptest.NewRecorder()
			h.GetMetricV2(recorder, test.params)

			response := recorder.Result()
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("ошибка при чтении тела: %v", err)
			}

			if response.StatusCode != test.wantStatus {
				t.Errorf("status = %d, want %d", response.StatusCode, test.wantStatus)
			}
			if contentType := response.Header.Get("Content-Type"); contentType != test.wantContentType {
				t.Errorf("Content-Type = %q, want %q", contentType, test.wantContentType)
			}

			if !test.exists {
				if string(body) != test.wantBody {
					t.Errorf("body = %q, want %q", body, test.wantBody)
				}
				return
			}

			var got handler.MetricsGetResponse
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("ошибка при разборе тела %q: %v", body, err)
			}

			if got.ID != test.wantResponse.ID || got.MType != test.wantResponse.MType {
				t.Errorf("response = %+v, want %+v", got, test.wantResponse)
			}
			if !equalPtr(got.Delta, test.wantResponse.Delta) {
				t.Errorf("delta = %v, want %v", deref(got.Delta), deref(test.wantResponse.Delta))
			}
			if !equalPtr(got.Value, test.wantResponse.Value) {
				t.Errorf("value = %v, want %v", deref(got.Value), deref(test.wantResponse.Value))
			}
		})
	}
}

func TestMetricsHandlerGetMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  map[string]models.Metrics
		wantBody string
	}{
		{
			name:     "renders empty list",
			metrics:  map[string]models.Metrics{},
			wantBody: "<!doctype html><html><body><ul></ul></body></html>",
		},
		{
			name: "renders metrics sorted by key",
			metrics: map[string]models.Metrics{
				"gauge-Alloc":         gaugeMetric("Alloc", 12.5),
				"counter-PollCount":   counterMetric("PollCount", 42),
				"gauge-<script>alert": gaugeMetric("<script>alert", 1),
			},
			wantBody: "<!doctype html><html><body><ul>" +
				"<li>counter-PollCount: 42</li>" +
				"<li>gauge-&lt;script&gt;alert: 1</li>" +
				"<li>gauge-Alloc: 12.5</li>" +
				"</ul></body></html>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsService, _ := newMetricsServiceStub()
			metricsService.metrics = test.metrics

			h := handler.NewMetricsHandler(metricsService)

			recorder := httptest.NewRecorder()
			h.GetMetrics(recorder)

			checkResponse(t, recorder, http.StatusOK, test.wantBody, "text/html; charset=utf-8")
		})
	}
}

func checkResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantBody, wantContentType string) []byte {
	t.Helper()

	response := recorder.Result()
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ошибка при чтении тела: %v", err)
	}

	if response.StatusCode != wantStatus {
		t.Errorf("status = %d, want %d", response.StatusCode, wantStatus)
	}
	if string(body) != wantBody {
		t.Errorf("body = %q, want %q", body, wantBody)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != wantContentType {
		t.Errorf("Content-Type = %q, want %q", contentType, wantContentType)
	}

	return body
}

func checkCalls(t *testing.T, got, want []serviceCall) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("service calls = %d, want %d", len(got), len(want))
	}
	for i, wantCall := range want {
		if gotCall := got[i]; gotCall != wantCall {
			t.Errorf("service call %d = %#v, want %#v", i, gotCall, wantCall)
		}
	}
}

func ptr[T any](v T) *T {
	return &v
}

func deref[T any](v *T) any {
	if v == nil {
		return nil
	}
	return *v
}

func equalPtr[T comparable](got, want *T) bool {
	if got == nil || want == nil {
		return got == want
	}
	return *got == *want
}
