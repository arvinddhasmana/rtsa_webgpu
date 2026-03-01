// CLASSIFICATION: UNCLASSIFIED
package audit

import (
	"context"
	"encoding/json"
	"time"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuditParams contains the parameters for an audit event.
type AuditParams struct {
EventType           string
ActorID             string
ActorType           auditv1.ActorType
ResourceType        string
ResourceID          string
Action              auditv1.AuditAction
Detail              map[string]interface{}
ClassificationLevel commonv1.ClassificationLevel
}

// Emitter produces audit events to Redpanda.
type Emitter struct {
producer  *redpanda.Producer
serviceID string
logger    *zap.Logger
}

// NewEmitter creates an audit event emitter for the given service.
func NewEmitter(producer *redpanda.Producer, serviceID string, logger *zap.Logger) *Emitter {
return &Emitter{
producer:  producer,
serviceID: serviceID,
logger:    logger,
}
}

// NewLogEmitter creates an audit emitter that logs events without producing to Redpanda.
// Use for development, testing, or services that only need local audit logging.
func NewLogEmitter(logger *zap.Logger) *Emitter {
return &Emitter{logger: logger}
}

// Emit produces an audit event. Never returns an error to the caller.
func (e *Emitter) Emit(ctx context.Context, params AuditParams) {
// Extract trace ID from context
traceID := ""
if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
traceID = span.SpanContext().TraceID().String()
}

// Serialize detail to JSON
detailJSON := "{}"
if params.Detail != nil {
b, err := json.Marshal(params.Detail)
if err != nil {
e.logger.Error("audit: failed to marshal detail", zap.Error(err))
} else {
detailJSON = string(b)
}
}

event := &auditv1.AuditEvent{
AuditId:             uuid.New().String(),
ServiceId:           e.serviceID,
EventType:           params.EventType,
ActorId:             params.ActorID,
ActorType:           params.ActorType,
ResourceType:        params.ResourceType,
ResourceId:          params.ResourceID,
Action:              params.Action,
DetailJson:          detailJSON,
ClassificationLevel: params.ClassificationLevel,
EventTime:           timestamppb.New(time.Now().UTC()),
TraceId:             traceID,
}

// protojson so Redpanda Connect Bloblang and svc-audit can parse field names.
b, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(event)
if err != nil {
e.logger.Error("audit: failed to marshal event", zap.Error(err))
return
}

classification := "UNCLASSIFIED"
	if e.producer == nil {
		e.logger.Info("audit event (log-only mode)",
			zap.String("event_type", params.EventType),
			zap.String("resource_type", params.ResourceType),
			zap.String("resource_id", params.ResourceID),
		)
		return
	}
	if err := e.producer.Produce(ctx, AuditTopic,
		[]byte(e.serviceID), b, classification, traceID); err != nil {
e.logger.Error("audit: failed to produce event",
zap.String("event_type", params.EventType),
zap.Error(err))
}
}
