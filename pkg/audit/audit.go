// CLASSIFICATION: UNCLASSIFIED
package audit

import (
"context"
"time"
)

// EventType classifies an audit event.
type EventType string

const (
EventTypeAlertProduced  EventType = "ALERT_PRODUCED"
EventTypeTrackConsumed  EventType = "TRACK_CONSUMED"
EventTypeDetectionRun   EventType = "DETECTION_RUN"
EventTypeServiceStart   EventType = "SERVICE_START"
EventTypeServiceStop    EventType = "SERVICE_STOP"
)

// AuditEvent represents a single auditable action.
type AuditEvent struct {
EventID    string            `json:"event_id"`
EventType  EventType         `json:"event_type"`
ServiceID  string            `json:"service_id"`
TraceID    string            `json:"trace_id,omitempty"`
Actor      string            `json:"actor,omitempty"`
ResourceID string            `json:"resource_id,omitempty"`
Outcome    string            `json:"outcome"`
Timestamp  time.Time         `json:"timestamp"`
Details    map[string]string `json:"details,omitempty"`
}

// AuditEmitter defines the interface for emitting audit events.
type AuditEmitter interface {
Emit(ctx context.Context, event AuditEvent) error
}

// NoopEmitter is a no-op AuditEmitter for testing or when audit is disabled.
type NoopEmitter struct{}

// Emit is a no-op that satisfies the AuditEmitter interface.
func (n *NoopEmitter) Emit(_ context.Context, _ AuditEvent) error {
return nil
}
