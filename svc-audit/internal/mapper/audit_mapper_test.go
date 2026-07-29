// CLASSIFICATION: UNCLASSIFIED
package mapper_test

import (
"testing"
"time"

auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/mapper"
)

func TestRowToAuditEvent_FullFields(t *testing.T) {
eventTime := time.Now().Truncate(time.Millisecond)

event := mapper.RowToAuditEvent(
"audit-001",
"svc-radar-ingestion",
"observation.ingested",
"svc-radar-ingestion",
"SERVICE",
"track",
"track-001",
"AUDIT_ACTION_INGEST",
`{"source":"radar"}`,
1, // UNCLASSIFIED
eventTime,
)

if event.AuditId != "audit-001" {
t.Errorf("AuditId: got %q, want %q", event.AuditId, "audit-001")
}
if event.ServiceId != "svc-radar-ingestion" {
t.Errorf("ServiceId mismatch")
}
if event.ActorType != auditv1.ActorType_ACTOR_TYPE_SERVICE {
t.Errorf("ActorType: got %v, want ACTOR_TYPE_SERVICE", event.ActorType)
}
if event.Action != auditv1.AuditAction_AUDIT_ACTION_INGEST {
t.Errorf("Action: got %v, want AUDIT_ACTION_INGEST", event.Action)
}
if event.ClassificationLevel != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("ClassificationLevel: got %v", event.ClassificationLevel)
}
if event.EventTime == nil {
t.Error("EventTime should not be nil")
}
}

func TestRowToAuditEvent_UnknownActorType(t *testing.T) {
event := mapper.RowToAuditEvent(
"a1", "svc", "type", "actor", "UNKNOWN_TYPE",
"res", "res-id", "AUDIT_ACTION_READ", "{}", 1, time.Now(),
)
if event.ActorType != auditv1.ActorType_ACTOR_TYPE_UNSPECIFIED {
t.Errorf("expected ACTOR_TYPE_UNSPECIFIED for unknown type, got %v", event.ActorType)
}
}

func TestRowToAuditEvent_AllActorTypes(t *testing.T) {
tests := []struct {
input    string
expected auditv1.ActorType
}{
{"SERVICE", auditv1.ActorType_ACTOR_TYPE_SERVICE},
{"OPERATOR", auditv1.ActorType_ACTOR_TYPE_OPERATOR},
{"SYSTEM", auditv1.ActorType_ACTOR_TYPE_SYSTEM},
{"INVALID", auditv1.ActorType_ACTOR_TYPE_UNSPECIFIED},
}
for _, tt := range tests {
t.Run(tt.input, func(t *testing.T) {
event := mapper.RowToAuditEvent(
"id", "svc", "type", "actor", tt.input,
"res", "res-id", "AUDIT_ACTION_READ", "{}", 1, time.Now(),
)
if event.ActorType != tt.expected {
t.Errorf("got %v, want %v", event.ActorType, tt.expected)
}
})
}
}

func TestRowToAuditEvent_AllActions(t *testing.T) {
tests := []struct {
input    string
expected auditv1.AuditAction
}{
{"AUDIT_ACTION_CREATE", auditv1.AuditAction_AUDIT_ACTION_CREATE},
{"AUDIT_ACTION_READ", auditv1.AuditAction_AUDIT_ACTION_READ},
{"AUDIT_ACTION_UPDATE", auditv1.AuditAction_AUDIT_ACTION_UPDATE},
{"AUDIT_ACTION_DELETE", auditv1.AuditAction_AUDIT_ACTION_DELETE},
{"AUDIT_ACTION_QUERY", auditv1.AuditAction_AUDIT_ACTION_QUERY},
{"AUDIT_ACTION_INGEST", auditv1.AuditAction_AUDIT_ACTION_INGEST},
{"AUDIT_ACTION_EXPORT", auditv1.AuditAction_AUDIT_ACTION_EXPORT},
{"AUDIT_ACTION_AUTHENTICATE", auditv1.AuditAction_AUDIT_ACTION_AUTHENTICATE},
{"AUDIT_ACTION_AUTHORIZE", auditv1.AuditAction_AUDIT_ACTION_AUTHORIZE},
{"AUDIT_ACTION_CLASSIFY", auditv1.AuditAction_AUDIT_ACTION_CLASSIFY},
{"AUDIT_ACTION_FEEDBACK", auditv1.AuditAction_AUDIT_ACTION_FEEDBACK},
{"UNKNOWN", auditv1.AuditAction_AUDIT_ACTION_UNSPECIFIED},
}
for _, tt := range tests {
t.Run(tt.input, func(t *testing.T) {
event := mapper.RowToAuditEvent(
"id", "svc", "type", "actor", "SERVICE",
"res", "res-id", tt.input, "{}", 1, time.Now(),
)
if event.Action != tt.expected {
t.Errorf("got %v, want %v", event.Action, tt.expected)
}
})
}
}
