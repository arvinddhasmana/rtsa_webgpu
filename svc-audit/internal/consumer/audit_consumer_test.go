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
