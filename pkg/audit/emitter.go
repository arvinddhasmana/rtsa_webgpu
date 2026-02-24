// CLASSIFICATION: UNCLASSIFIED
package audit

import (
	"context"

	"go.uber.org/zap"
)

// Event describes a single auditable action.
type Event struct {
	EventType    string
	ResourceType string
	ResourceID   string
	Action       string
	DetailJSON   string
}

// Emitter emits structured audit events via the logger.
// In production this would also produce to a Redpanda audit topic.
type Emitter struct {
	logger *zap.Logger
}

// NewEmitter creates an Emitter backed by the given logger.
func NewEmitter(logger *zap.Logger) *Emitter {
	return &Emitter{logger: logger}
}

// Emit logs the audit event. The context is accepted for future Redpanda production.
func (e *Emitter) Emit(_ context.Context, ev Event) {
	e.logger.Info("audit event",
		zap.String("event_type", ev.EventType),
		zap.String("resource_type", ev.ResourceType),
		zap.String("resource_id", ev.ResourceID),
		zap.String("action", ev.Action),
	)
}
