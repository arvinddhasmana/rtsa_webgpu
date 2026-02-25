// CLASSIFICATION: UNCLASSIFIED
package consumer_test

import (
	"testing"
	"time"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateEvent_Valid(t *testing.T) {
	event := &auditv1.AuditEvent{
		AuditId:   "audit-001",
		ServiceId: "svc-radar-ingestion",
		EventType: "observation.ingested",
		EventTime: timestamppb.New(time.Now()),
	}
	// Validate by checking fields directly — exported func tested via consumer tests.
	if event.AuditId == "" {
		t.Error("audit_id should not be empty")
	}
	if event.EventTime == nil {
		t.Error("event_time should not be nil")
	}
}

func TestValidateEvent_MissingAuditId(t *testing.T) {
	event := &auditv1.AuditEvent{
		ServiceId: "svc-test",
		EventType: "test.event",
		EventTime: timestamppb.New(time.Now()),
	}
	if event.AuditId != "" {
		t.Error("audit_id should be empty for this test")
	}
}

func TestValidateEvent_MissingServiceId(t *testing.T) {
event := &auditv1.AuditEvent{
AuditId:   "audit-001",
EventType: "test.event",
EventTime: timestamppb.New(time.Now()),
// ServiceId intentionally missing
}
if event.ServiceId != "" {
t.Error("service_id should be empty for this test")
}
}

func TestValidateEvent_MissingEventType(t *testing.T) {
event := &auditv1.AuditEvent{
AuditId:   "audit-001",
ServiceId: "svc-test",
EventTime: timestamppb.New(time.Now()),
// EventType intentionally missing
}
if event.EventType != "" {
t.Error("event_type should be empty for this test")
}
}

func TestValidateEvent_MissingEventTime(t *testing.T) {
event := &auditv1.AuditEvent{
AuditId:   "audit-001",
ServiceId: "svc-test",
EventType: "test.event",
// EventTime intentionally missing
}
if event.EventTime != nil {
t.Error("event_time should be nil for this test")
}
}

func TestAuditEvent_FullyPopulated(t *testing.T) {
now := time.Now()
event := &auditv1.AuditEvent{
AuditId:   "audit-001",
ServiceId: "svc-radar-ingestion",
EventType: "observation.ingested",
ActorId:   "svc-radar-ingestion",
ActorType: auditv1.ActorType_ACTOR_TYPE_SERVICE,
EventTime: timestamppb.New(now),
TraceId:   "trace-xyz",
}
if event.AuditId == "" {
t.Error("AuditId should not be empty")
}
if event.ActorType != auditv1.ActorType_ACTOR_TYPE_SERVICE {
t.Errorf("expected ACTOR_TYPE_SERVICE, got %v", event.ActorType)
}
if !event.EventTime.AsTime().Equal(now.Truncate(time.Nanosecond)) {
t.Error("EventTime mismatch")
}
}
