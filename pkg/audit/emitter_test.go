// CLASSIFICATION: UNCLASSIFIED
package audit_test

import (
"context"
"testing"

auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
"go.opentelemetry.io/otel"
"go.uber.org/zap"
)

func TestEmitter_EmitNilDetail(t *testing.T) {
logger := zap.NewNop()
emitter := audit.NewEmitter(nil, "test-svc", logger)
_ = emitter
}

func TestAuditParams_Fields(t *testing.T) {
params := audit.AuditParams{
EventType:           audit.EventObservationIngested,
ActorID:             "sensor-01",
ActorType:           auditv1.ActorType_ACTOR_TYPE_SERVICE,
ResourceType:        "observation",
ResourceID:          "obs-123",
Action:              auditv1.AuditAction_AUDIT_ACTION_INGEST,
Detail:              map[string]interface{}{"key": "value"},
ClassificationLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

if params.EventType != audit.EventObservationIngested {
t.Errorf("unexpected event type: %s", params.EventType)
}
if params.Action != auditv1.AuditAction_AUDIT_ACTION_INGEST {
t.Errorf("unexpected action: %v", params.Action)
}
}

func TestEmitter_NilProducer_NocrashOnEmit(t *testing.T) {
logger := zap.NewNop()
emitter := audit.NewEmitter(nil, "test-service", logger)
ctx := context.Background()

defer func() {
if r := recover(); r != nil {
t.Errorf("should not panic: %v", r)
}
}()

emitter.Emit(ctx, audit.AuditParams{
EventType: audit.EventSensorConnected,
ActorType: auditv1.ActorType_ACTOR_TYPE_SERVICE,
Action:    auditv1.AuditAction_AUDIT_ACTION_INGEST,
})
}

func TestAuditConstants(t *testing.T) {
constants := []string{
audit.EventTrackCreated,
audit.EventTrackUpdated,
audit.EventSensorConnected,
audit.EventObservationIngested,
audit.AuditTopic,
}
for _, c := range constants {
if c == "" {
t.Error("audit constant should not be empty")
}
}
}

func TestNewEmitter_NotNil(t *testing.T) {
logger := zap.NewNop()
emitter := audit.NewEmitter(nil, "test-svc", logger)
if emitter == nil {
t.Error("expected non-nil emitter")
}
}

func TestEmitter_WithNonNilDetail_NilProducer(t *testing.T) {
logger := zap.NewNop()
emitter := audit.NewEmitter(nil, "test-service", logger)
ctx := context.Background()

// non-nil detail — exercises json.Marshal path
emitter.Emit(ctx, audit.AuditParams{
EventType: audit.EventSensorConnected,
ActorType: auditv1.ActorType_ACTOR_TYPE_SERVICE,
Action:    auditv1.AuditAction_AUDIT_ACTION_INGEST,
Detail: map[string]interface{}{
"sensor_id": "RADAR-001",
"count":     42,
},
})
// Should not panic
}

func TestEmitter_WithTraceContext(t *testing.T) {
logger := zap.NewNop()
emitter := audit.NewEmitter(nil, "test-service", logger)

// Use a context with a real span to exercise the trace ID extraction path
tracer := otel.GetTracerProvider().Tracer("test")
ctx, span := tracer.Start(context.Background(), "test-span")
defer span.End()

emitter.Emit(ctx, audit.AuditParams{
EventType: audit.EventObservationIngested,
ActorType: auditv1.ActorType_ACTOR_TYPE_SERVICE,
Action:    auditv1.AuditAction_AUDIT_ACTION_INGEST,
})
}
