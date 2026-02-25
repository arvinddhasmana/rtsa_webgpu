// CLASSIFICATION: UNCLASSIFIED
package mapper

import (
commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

// PassesClassificationFilter returns true when the alert's classification level
// does not exceed the operator's clearance level.
//
// UNSPECIFIED classification is treated as UNCLASSIFIED.
// UNSPECIFIED clearance is treated as UNCLASSIFIED.
//
// This filter is MANDATORY and must be applied before sending any alert to a client.
func PassesClassificationFilter(
alertClassification commonv1.ClassificationLevel,
clearanceLevel commonv1.ClassificationLevel,
) bool {
effective := func(level commonv1.ClassificationLevel) int32 {
if level == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
return int32(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
}
return int32(level)
}

return effective(alertClassification) <= effective(clearanceLevel)
}

// PassesStreamFilter returns true when the alert satisfies all filters in the
// StreamAlertsRequest:
//   - Classification <= clearance level (MANDATORY)
//   - Severity >= min_severity
//   - AnomalyType in allowed types (empty slice = allow all)
//   - EntityType in allowed entity types (empty slice = allow all)
func PassesStreamFilter(
alert *inferencev1.AnomalyAlert,
req *inferencev1.StreamAlertsRequest,
) bool {
if alert == nil || req == nil {
return false
}

// 1. Classification filter — MANDATORY
if !PassesClassificationFilter(alert.GetClassification(), req.GetClearanceLevel()) {
return false
}

// 2. Severity filter
minSev := req.GetMinSeverity()
if minSev == commonv1.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED {
minSev = commonv1.AlertSeverity_ALERT_SEVERITY_WATCH
}
if !meetsMinSeverity(alert.GetSeverity(), minSev) {
return false
}

// 3. AnomalyType filter (empty = all allowed)
if len(req.GetAnomalyTypes()) > 0 && !containsAnomalyType(req.GetAnomalyTypes(), alert.GetAnomalyType()) {
return false
}

// 4. EntityType filter (empty = all allowed)
if len(req.GetEntityTypes()) > 0 && !containsEntityType(req.GetEntityTypes(), alert.GetEntityType()) {
return false
}

return true
}

// meetsMinSeverity returns true if actual severity is >= min severity (by rank).
func meetsMinSeverity(actual, min commonv1.AlertSeverity) bool {
return severityRank(actual) >= severityRank(min)
}

// severityRank returns the numeric priority rank used for comparison.
func severityRank(s commonv1.AlertSeverity) int {
switch s {
case commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL:
return 3
case commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED:
return 2
case commonv1.AlertSeverity_ALERT_SEVERITY_WATCH:
return 1
case commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL:
return 0
default:
return -1
}
}

func containsAnomalyType(list []commonv1.AnomalyType, target commonv1.AnomalyType) bool {
for _, t := range list {
if t == target {
return true
}
}
return false
}

func containsEntityType(list []commonv1.EntityType, target commonv1.EntityType) bool {
for _, t := range list {
if t == target {
return true
}
}
return false
}

// QueuedAlertToProto converts a domain.QueuedAlert back to the wire proto,
// ensuring the Acknowledged field reflects domain state.
// This is largely a pass-through since we store the proto directly in the queue.
func QueuedAlertToProto(qa *domain.QueuedAlert) *inferencev1.AnomalyAlert {
if qa == nil {
return nil
}
// The proto is stored directly; return it (the Acknowledge method already mutates it).
return qa.Alert
}

// QueuedAlertAckedAtProto returns the acknowledgment timestamp as a Timestamp proto.
func QueuedAlertAckedAtProto(qa *domain.QueuedAlert) *timestamppb.Timestamp {
if qa == nil || qa.AckedAt == nil {
return nil
}
return timestamppb.New(*qa.AckedAt)
}
