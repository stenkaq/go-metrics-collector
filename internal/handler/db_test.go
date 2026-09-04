package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-collector/internal/handler"
)

// dbServiceStub подменяет доступ к БД: считает пинги и отдаёт заданную ошибку.
type dbServiceStub struct {
	err   error
	pings int
}

func (s *dbServiceStub) Ping(ctx context.Context) error {
	s.pings++
	return s.err
}

func TestDBHandlerPing(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "reports a healthy database",
			wantStatus: http.StatusOK,
		},
		{
			name:       "reports a broken database",
			pingErr:    errors.New("БД недоступна"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Ошибка при пинге БД\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dbService := &dbServiceStub{err: test.pingErr}

			h := handler.NewDBHandler(dbService)

			recorder := httptest.NewRecorder()
			h.Ping(newParamsContext(recorder, nil))

			if recorder.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if body := recorder.Body.String(); body != test.wantBody {
				t.Errorf("body = %q, want %q", body, test.wantBody)
			}
			if dbService.pings != 1 {
				t.Errorf("пингов БД = %d, want 1", dbService.pings)
			}
		})
	}
}
