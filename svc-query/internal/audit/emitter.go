// CLASSIFICATION: UNCLASSIFIED

// Package audit provides the audit event emitter for the query service.
// Every query execution emits an immutable audit record to the audit.events
// Redpanda topic, satisfying ITSG-33 accountability requirements.
package audit

import (
"context"
"encoding/json"
"fmt"
"log/slog"
"time"

auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/google/uuid"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

// topicAuditEvents is the Redpanda topic to which audit events are produced.
const topicAuditEvents = "audit.events"

// AuditParams carries the parameters for a single audit event emission.
type AuditParams struct {
// EventType categorises the audited operation (e.g. "query_executed").
EventType string
// ResourceType is the type of the resource acted upon (e.g. "tracks").
ResourceType string
// ResourceID is the identifier of the resource (empty for query operations).
ResourceID string
// Action is the audit action performed.
Action auditv1.AuditAction
// ActorID is the operator or service identifier.
ActorID string
// ActorType classifies the actor.
ActorType auditv1.ActorType
// ClassificationLevel is the classification ceiling of the operation.
ClassificationLevel commonv1.ClassificationLevel
// Details are key-value pairs serialised into detail_json.
// MUST NOT contain classified data, PII, or raw sensor payloads.
Details map[string]interface{}
// TraceID for distributed tracing correlation.
TraceID string
}

// Emitter is the interface for emitting audit events.
// Mock implementations satisfy this interface in tests.
type Emitter interface {
Emit(ctx context.Context, serviceID string, params AuditParams) error
Close() error
}

// RedpandaEmitter produces audit events to the audit.events Redpanda topic.
type RedpandaEmitter struct {
client    *kgo.Client
serviceID string
}

// NewRedpandaEmitter creates an emitter that publishes to the given brokers.
func NewRedpandaEmitter(brokers []string, serviceID string) (*RedpandaEmitter, error) {
client, err := kgo.NewClient(
kgo.SeedBrokers(brokers...),
kgo.ProducerBatchCompression(kgo.SnappyCompression()),
kgo.RequiredAcks(kgo.AllISRAcks()),
kgo.ProducerBatchMaxBytes(1<<20),
)
if err != nil {
return nil, fmt.Errorf("audit.NewRedpandaEmitter: %w", err)
}
return &RedpandaEmitter{client: client, serviceID: serviceID}, nil
}

// Emit serialises and produces a single audit event synchronously.
// Errors are surfaced to the caller; audit failure is logged but never panics.
func (e *RedpandaEmitter) Emit(ctx context.Context, serviceID string, params AuditParams) error {
detailJSON, err := marshalDetails(params.Details)
if err != nil {
slog.ErrorContext(ctx, "audit.Emit: marshal details failed, using empty details",
"error", err,
"event_type", params.EventType)
detailJSON = "{}"
}

event := &auditv1.AuditEvent{
AuditId:             uuid.New().String(),
ServiceId:           serviceID,
EventType:           params.EventType,
ActorId:             params.ActorID,
ActorType:           params.ActorType,
ResourceType:        params.ResourceType,
ResourceId:          params.ResourceID,
Action:              params.Action,
DetailJson:          detailJSON,
ClassificationLevel: params.ClassificationLevel,
EventTime:           timestamppb.New(time.Now().UTC()),
TraceId:             params.TraceID,
}

payload, err := proto.Marshal(event)
if err != nil {
return fmt.Errorf("audit.Emit: proto marshal: %w", err)
}

record := &kgo.Record{
Topic: topicAuditEvents,
Key:   []byte(serviceID),
Value: payload,
}

results := e.client.ProduceSync(ctx, record)
if err := results.FirstErr(); err != nil {
return fmt.Errorf("audit.Emit: produce to Redpanda: %w", err)
}

return nil
}

// Close flushes pending records and shuts down the Redpanda client.
func (e *RedpandaEmitter) Close() error {
e.client.Close()
return nil
}

// marshalDetails converts the details map to a JSON string.
// Returns "{}" on error or nil input.
func marshalDetails(details map[string]interface{}) (string, error) {
if len(details) == 0 {
return "{}", nil
}
b, err := json.Marshal(details)
if err != nil {
return "{}", fmt.Errorf("audit.marshalDetails: %w", err)
}
return string(b), nil
}

// NoopEmitter is a no-op audit emitter that records events in memory for testing.
type NoopEmitter struct {
Events []AuditParams
}

// Emit records the event in-memory and returns nil.
func (n *NoopEmitter) Emit(_ context.Context, _ string, params AuditParams) error {
n.Events = append(n.Events, params)
return nil
}

// Close is a no-op.
func (n *NoopEmitter) Close() error { return nil }
