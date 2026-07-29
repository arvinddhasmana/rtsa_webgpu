// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"context"
"errors"
"log/slog"
"os"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/domain"
"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/testutil"
)

func newTestLogger() *slog.Logger {
return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestAcknowledgerMetrics() *domain.AcknowledgerMetrics {
return &domain.AcknowledgerMetrics{
TimeToAcknowledge: prometheus.NewHistogramVec(prometheus.HistogramOpts{
Name:    "test_time_to_acknowledge_seconds",
Help:    "Test histogram.",
Buckets: []float64{1, 10, 60},
}, []string{"severity"}),
}
}

// TestAcknowledger_Success verifies a valid acknowledgment succeeds.
func TestAcknowledger_Success(t *testing.T) {
q := domain.NewAlertQueue(100)
metrics := newTestAcknowledgerMetrics()
ack := domain.NewAcknowledger(q, metrics, newTestLogger())

alert := makeAlert("alert-ack-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, time.Now().Add(-5*time.Minute))
q.Enqueue(alert)

req := &inferencev1.AcknowledgeAlertRequest{
AlertId:    "alert-ack-1",
OperatorId: "op-007",
Comment:    "Investigated and resolved",
}

ackedAt, err := ack.Acknowledge(context.Background(), req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if ackedAt == nil {
t.Fatal("expected non-nil ackedAt")
}

// Verify the metric was observed. CollectAndCount returns the number of
// distinct metric samples; for a histogram with one observed label it is > 0.
collected := testutil.CollectAndCount(metrics.TimeToAcknowledge)
if collected == 0 {
t.Error("expected time-to-acknowledge metric to have observations")
}
}

// TestAcknowledger_MissingAlertID validates that empty alert_id is rejected.
func TestAcknowledger_MissingAlertID(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())

_, err := ack.Acknowledge(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "",
OperatorId: "op-001",
})
if err == nil {
t.Error("expected error for empty alert_id")
}
}

// TestAcknowledger_MissingOperatorID validates that empty operator_id is rejected.
func TestAcknowledger_MissingOperatorID(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())

_, err := ack.Acknowledge(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "some-alert",
OperatorId: "",
})
if err == nil {
t.Error("expected error for empty operator_id")
}
}

// TestAcknowledger_NotFound verifies ErrAlertNotFound is returned for unknown IDs.
func TestAcknowledger_NotFound(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())

_, err := ack.Acknowledge(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "ghost-alert",
OperatorId: "op-001",
})
if err == nil {
t.Fatal("expected error for non-existent alert")
}
if !errors.Is(err, domain.ErrAlertNotFound) {
t.Errorf("expected ErrAlertNotFound, got %v", err)
}
}

// TestAcknowledger_NilMetrics ensures nil metrics struct doesn't panic.
func TestAcknowledger_NilMetrics(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())

q.Enqueue(makeAlert("alert-nil-m", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, time.Now()))

_, err := ack.Acknowledge(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "alert-nil-m",
OperatorId: "op-001",
})
if err != nil {
t.Errorf("unexpected error with nil metrics: %v", err)
}
}
