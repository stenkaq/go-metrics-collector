package handler_test

import (
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
}

type serviceCalls struct {
	values []serviceCall
}

func newMetricsServiceStub() (service.MetricsService, *serviceCalls) {
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

func (s *metricsServiceStub) GetMetric(string, string) (models.Metrics, bool) {
	return models.Metrics{}, false
}

func (s *metricsServiceStub) GetMetrics() map[string]models.Metrics {
	return nil
}

func TestMetricsHandlerUpdate(t *testing.T) {
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metricsService, calls := newMetricsServiceStub()
			h := handler.NewMetricsHandler(metricsService)

			recorder := httptest.NewRecorder()
			h.UpdateMetricV2(recorder, test.params)

			response := recorder.Result()
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("ошибка при чтении тела: %v", err)
			}

			if response.StatusCode != test.wantStatus {
				t.Errorf("status = %d, want %d", response.StatusCode, test.wantStatus)
			}
			if string(body) != test.wantBody {
				t.Errorf("body = %q, want %q", body, test.wantBody)
			}
			if contentType := response.Header.Get("Content-Type"); contentType != test.wantContentType {
				t.Errorf("Content-Type = %q, want %q", contentType, test.wantContentType)
			}
			if len(calls.values) != len(test.wantCalls) {
				t.Fatalf("service calls = %d, want %d", len(calls.values), len(test.wantCalls))
			}
			for i, wantCall := range test.wantCalls {
				if gotCall := calls.values[i]; gotCall != wantCall {
					t.Errorf("service call %d = %#v, want %#v", i, gotCall, wantCall)
				}
			}
		})
	}
}
