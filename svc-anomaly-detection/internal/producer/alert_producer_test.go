// CLASSIFICATION: UNCLASSIFIED
package producer

import (
"context"
"testing"
"log/slog"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
)

func TestAlertProducer_Produce_Valid(t *testing.T) {
mock := &mockProducer{}
ap := NewAlertProducer(mock, slog.Default())

alert := &inferencev1.AnomalyAlert{
AlertId:         "alert-001",
TrackId:         "track-001",
AnomalyType:     commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
Severity:        commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
ConfidenceScore: 0.75,
ModelVersion:    "rules-v1.0.0",
}

err := ap.Produce(context.Background(), "alerts.anomaly.elevated", alert)
if err != nil {
t.Fatalf("Produce returned error: %v", err)
}
if mock.topic != "alerts.anomaly.elevated" {
t.Errorf("topic = %q, want %q", mock.topic, "alerts.anomaly.elevated")
}
if string(mock.key) != "track-001" {
t.Errorf("key = %q, want %q", string(mock.key), "track-001")
}
}

func TestAlertProducer_Produce_EmptyTopic(t *testing.T) {
mock := &mockProducer{}
ap := NewAlertProducer(mock, slog.Default())
alert := &inferencev1.AnomalyAlert{AlertId: "a1", TrackId: "t1"}
err := ap.Produce(context.Background(), "", alert)
if err == nil {
t.Error("Expected error for empty topic")
}
}

func TestAlertProducer_Produce_NilAlert(t *testing.T) {
mock := &mockProducer{}
ap := NewAlertProducer(mock, slog.Default())
err := ap.Produce(context.Background(), "alerts.anomaly.watch", nil)
if err == nil {
t.Error("Expected error for nil alert")
}
}

func TestAlertProducer_Produce_EmptyAlertID(t *testing.T) {
mock := &mockProducer{}
ap := NewAlertProducer(mock, slog.Default())
alert := &inferencev1.AnomalyAlert{TrackId: "t1"} // No AlertId.
err := ap.Produce(context.Background(), "alerts.anomaly.watch", alert)
if err == nil {
t.Error("Expected error for empty alert_id")
}
}

func TestAlertProducer_Close(t *testing.T) {
mock := &mockProducer{}
ap := NewAlertProducer(mock, slog.Default())
err := ap.Close()
if err != nil {
t.Fatalf("Close returned error: %v", err)
}
if !mock.closed {
t.Error("Expected underlying producer to be closed")
}
}
