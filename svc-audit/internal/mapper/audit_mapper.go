// CLASSIFICATION: UNCLASSIFIED
package mapper

import (
"time"

auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"google.golang.org/protobuf/types/known/timestamppb"
)

// RowToAuditEvent converts scanned ClickHouse row values into an AuditEvent proto.
// The audit_log table schema (no trace_id column):
//
//audit_id, service_id, event_type, actor_id, actor_type (Enum8),
//resource_type, resource_id, action, detail_json,
//classification_level (Enum8), event_time
func RowToAuditEvent(
auditID, serviceID, eventType, actorID string,
actorTypeStr string,
resourceType, resourceID, action, detailJSON string,
classLevel int32,
eventTime time.Time,
) *auditv1.AuditEvent {
return &auditv1.AuditEvent{
AuditId:             auditID,
ServiceId:           serviceID,
EventType:           eventType,
ActorId:             actorID,
ActorType:           parseActorType(actorTypeStr),
ResourceType:        resourceType,
ResourceId:          resourceID,
Action:              parseAuditAction(action),
DetailJson:          detailJSON,
ClassificationLevel: commonv1.ClassificationLevel(classLevel),
EventTime:           timestamppb.New(eventTime),
}
}

// parseActorType converts a string actor_type from ClickHouse to the proto enum.
func parseActorType(s string) auditv1.ActorType {
switch s {
case "SERVICE":
return auditv1.ActorType_ACTOR_TYPE_SERVICE
case "OPERATOR":
return auditv1.ActorType_ACTOR_TYPE_OPERATOR
case "SYSTEM":
return auditv1.ActorType_ACTOR_TYPE_SYSTEM
default:
return auditv1.ActorType_ACTOR_TYPE_UNSPECIFIED
}
}

// parseAuditAction converts a string action from ClickHouse to the proto enum.
func parseAuditAction(s string) auditv1.AuditAction {
switch s {
case "AUDIT_ACTION_CREATE":
return auditv1.AuditAction_AUDIT_ACTION_CREATE
case "AUDIT_ACTION_READ":
return auditv1.AuditAction_AUDIT_ACTION_READ
case "AUDIT_ACTION_UPDATE":
return auditv1.AuditAction_AUDIT_ACTION_UPDATE
case "AUDIT_ACTION_DELETE":
return auditv1.AuditAction_AUDIT_ACTION_DELETE
case "AUDIT_ACTION_QUERY":
return auditv1.AuditAction_AUDIT_ACTION_QUERY
case "AUDIT_ACTION_INGEST":
return auditv1.AuditAction_AUDIT_ACTION_INGEST
case "AUDIT_ACTION_EXPORT":
return auditv1.AuditAction_AUDIT_ACTION_EXPORT
case "AUDIT_ACTION_AUTHENTICATE":
return auditv1.AuditAction_AUDIT_ACTION_AUTHENTICATE
case "AUDIT_ACTION_AUTHORIZE":
return auditv1.AuditAction_AUDIT_ACTION_AUTHORIZE
case "AUDIT_ACTION_CLASSIFY":
return auditv1.AuditAction_AUDIT_ACTION_CLASSIFY
case "AUDIT_ACTION_FEEDBACK":
return auditv1.AuditAction_AUDIT_ACTION_FEEDBACK
default:
return auditv1.AuditAction_AUDIT_ACTION_UNSPECIFIED
}
}
