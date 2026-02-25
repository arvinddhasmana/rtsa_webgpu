// CLASSIFICATION: UNCLASSIFIED
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// Status represents the health state of a component.
type HTTPStatus string

const (
	StatusUp       HTTPStatus = "UP"
	StatusDown     HTTPStatus = "DOWN"
	StatusDegraded HTTPStatus = "DEGRADED"
)

// HealthStatus holds the status and details for a component.
type HealthStatus struct {
	Status  HTTPStatus        `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

// HealthChecker defines the interface for health-checkable components.
type HealthChecker interface {
	Check(ctx context.Context) HealthStatus
	Name() string
}

// Server is an HTTP health check server.
type HTTPServer struct {
	mu       sync.RWMutex
	checkers []HealthChecker
	logger   *slog.Logger
}

// NewServer creates a new health check HTTP server.
func NewHTTPServer(logger *slog.Logger) *HTTPServer {
	return &HTTPServer{logger: logger}
}

// Register adds a HealthChecker to the server.
func (s *HTTPServer) Register(checker HealthChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkers = append(s.checkers, checker)
}

// ServeHTTP handles /healthz requests.
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	checkers := make([]HealthChecker, len(s.checkers))
	copy(checkers, s.checkers)
	s.mu.RUnlock()

	overall := StatusUp
	results := make(map[string]HealthStatus, len(checkers))
	for _, c := range checkers {
		status := c.Check(r.Context())
		results[c.Name()] = status
		if status.Status == StatusDown {
			overall = StatusDown
		} else if status.Status == StatusDegraded && overall != StatusDown {
			overall = StatusDegraded
		}
	}

	code := http.StatusOK
	if overall == StatusDown {
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"status":     overall,
		"components": results,
	}); err != nil {
		s.logger.Error("health encode error", "error", err)
	}
}

// ListenAndServe starts the health server on the given address.
func (s *HTTPServer) ListenAndServe(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/healthz", s)
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-ctx.Done()
		if err := srv.Close(); err != nil {
			s.logger.Error("health server close error", "error", err)
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("[health.Server.ListenAndServe]: %w", err)
	}
	return nil
}
