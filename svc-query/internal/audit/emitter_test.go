// CLASSIFICATION: UNCLASSIFIED

package audit

import (
"context"
"testing"

auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
)

func TestNoopEmitter_Emit(t *testing.T) {
emitter := &NoopEmitter{}

params := AuditParams{
EventType:           "query_executed",
ResourceType:        "tracks",
Action:              auditv1.AuditAction_AUDIT_ACTION_QUERY,
ActorType:           auditv1.ActorType_ACTOR_TYPE_OPERATOR,
ClassificationLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Details: map[string]interface{}{
"result_count": 42,
},
}

if err := emitter.Emit(context.Background(), "svc-query-test", params); err != nil {
t.Fatalf("unexpected error: %v", err)
}

if len(emitter.Events) != 1 {
t.Errorf("expected 1 event, got %d", len(emitter.Events))
}
if emitter.Events[0].EventType != "query_executed" {
t.Errorf("EventType = %q, want %q", emitter.Events[0].EventType, "query_executed")
}
}

func TestNoopEmitter_MultipleEmits(t *testing.T) {
emitter := &NoopEmitter{}

for i := 0; i < 5; i++ {
err := emitter.Emit(context.Background(), "svc-query-test", AuditParams{
EventType: "query_executed",
})
if err != nil {
t.Fatalf("emit %d: unexpected error: %v", i, err)
}
}

if len(emitter.Events) != 5 {
t.Errorf("expected 5 events, got %d", len(emitter.Events))
}
}

func TestNoopEmitter_Close(t *testing.T) {
emitter := &NoopEmitter{}
if err := emitter.Close(); err != nil {
t.Errorf("unexpected error from Close: %v", err)
}
}

func TestMarshalDetails_nil(t *testing.T) {
got, err := marshalDetails(nil)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got != "{}" {
t.Errorf("expected {}, got %q", got)
}
}

func TestMarshalDetails_empty(t *testing.T) {
got, err := marshalDetails(map[string]interface{}{})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got != "{}" {
t.Errorf("expected {}, got %q", got)
}
}

func TestMarshalDetails_withData(t *testing.T) {
details := map[string]interface{}{
"result_count": 42,
"query_type":   "QueryTracks",
}
got, err := marshalDetails(details)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got == "{}" {
t.Error("expected non-empty JSON, got {}")
}
// Should be valid JSON containing both keys
if len(got) < 20 {
t.Errorf("JSON too short: %q", got)
}
}

// Verify NoopEmitter satisfies the Emitter interface at compile time.
var _ Emitter = (*NoopEmitter)(nil)
