package health

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockChecker struct {
	name   string
	status HealthStatus
}

func (m *mockChecker) Check(ctx context.Context) HealthStatus { return m.status }
func (m *mockChecker) Name() string                           { return m.name }

func TestHTTPServer_ServeHTTP_AllUp(t *testing.T) {
	logger := slog.Default()
	s := NewHTTPServer(logger)

	s.Register(&mockChecker{"c1", HealthStatus{Status: StatusUp}})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHTTPServer_ServeHTTP_Degraded(t *testing.T) {
	logger := slog.Default()
	s := NewHTTPServer(logger)

	s.Register(&mockChecker{"c1", HealthStatus{Status: StatusUp}})
	s.Register(&mockChecker{"c2", HealthStatus{Status: StatusDegraded}})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHTTPServer_ServeHTTP_Down(t *testing.T) {
	logger := slog.Default()
	s := NewHTTPServer(logger)

	s.Register(&mockChecker{"c1", HealthStatus{Status: StatusUp}})
	s.Register(&mockChecker{"c2", HealthStatus{Status: StatusDown}})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestHTTPServer_ListenAndServe(t *testing.T) {
	logger := slog.Default()
	s := NewHTTPServer(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := s.ListenAndServe(ctx, "localhost:0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
