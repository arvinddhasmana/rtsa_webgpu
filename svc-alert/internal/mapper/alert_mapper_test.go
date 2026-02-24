// CLASSIFICATION: UNCLASSIFIED
package mapper_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/mapper"
"google.golang.org/protobuf/types/known/timestamppb"
)

func makeTestAlert(id string, sev commonv1.AlertSeverity, cl commonv1.ClassificationLevel, at commonv1.AnomalyType, et commonv1.EntityType) *inferencev1.AnomalyAlert {
return &inferencev1.AnomalyAlert{
AlertId:        id,
Severity:       sev,
Classification: cl,
AnomalyType:    at,
EntityType:     et,
DetectedAt:     timestamppb.New(time.Now()),
}
}

// TestPassesClassificationFilter covers all level comparisons.
func TestPassesClassificationFilter(t *testing.T) {
tests := []struct {
name        string
alertClass  commonv1.ClassificationLevel
clearance   commonv1.ClassificationLevel
wantPass    bool
}{
{"unclassified passes unclassified clearance", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, true},
{"protected_c passes secret clearance", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, true},
{"secret denied by protected_c clearance", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C, false},
{"unspecified alert passes unspecified clearance", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED, true},
{"unspecified alert passes unclassified clearance", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, true},
{"protected_a denied by unclassified clearance", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
got := mapper.PassesClassificationFilter(tc.alertClass, tc.clearance)
if got != tc.wantPass {
t.Errorf("PassesClassificationFilter(%v, %v) = %v, want %v",
tc.alertClass, tc.clearance, got, tc.wantPass)
}
})
}
}

// TestPassesStreamFilter_Severity tests severity filtering.
func TestPassesStreamFilter_Severity(t *testing.T) {
alert := makeTestAlert("a1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.EntityType_ENTITY_TYPE_SURFACE)

// Watch meets WATCH minimum
req := &inferencev1.StreamAlertsRequest{
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
if !mapper.PassesStreamFilter(alert, req) {
t.Error("expected WATCH alert to pass WATCH min_severity filter")
}

// Watch does NOT meet ELEVATED minimum
req2 := &inferencev1.StreamAlertsRequest{
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
if mapper.PassesStreamFilter(alert, req2) {
t.Error("expected WATCH alert to fail ELEVATED min_severity filter")
}
}

// TestPassesStreamFilter_AnomalyType tests anomaly type filtering.
func TestPassesStreamFilter_AnomalyType(t *testing.T) {
alert := makeTestAlert("a2", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL,
commonv1.EntityType_ENTITY_TYPE_AIR)

// Exact match passes
req := &inferencev1.StreamAlertsRequest{
AnomalyTypes:   []commonv1.AnomalyType{commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL},
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
if !mapper.PassesStreamFilter(alert, req) {
t.Error("expected BEHAVIORAL to match filter")
}

// No match fails
req2 := &inferencev1.StreamAlertsRequest{
AnomalyTypes:   []commonv1.AnomalyType{commonv1.AnomalyType_ANOMALY_TYPE_SPEED},
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
if mapper.PassesStreamFilter(alert, req2) {
t.Error("expected BEHAVIORAL to fail SPEED anomaly type filter")
}

// Empty list = all types allowed
req3 := &inferencev1.StreamAlertsRequest{
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
if !mapper.PassesStreamFilter(alert, req3) {
t.Error("expected alert to pass when anomaly_types is empty (all allowed)")
}
}

// TestPassesStreamFilter_EntityType tests entity type filtering.
func TestPassesStreamFilter_EntityType(t *testing.T) {
alert := makeTestAlert("a3", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION,
commonv1.EntityType_ENTITY_TYPE_SUBSURFACE)

// Match passes
req := &inferencev1.StreamAlertsRequest{
EntityTypes:    []commonv1.EntityType{commonv1.EntityType_ENTITY_TYPE_SUBSURFACE},
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
if !mapper.PassesStreamFilter(alert, req) {
t.Error("expected SUBSURFACE to match filter")
}

// Non-match fails
req2 := &inferencev1.StreamAlertsRequest{
EntityTypes:    []commonv1.EntityType{commonv1.EntityType_ENTITY_TYPE_AIR},
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}
if mapper.PassesStreamFilter(alert, req2) {
t.Error("expected SUBSURFACE to fail AIR entity type filter")
}
}

// TestPassesStreamFilter_NilEdgeCases covers nil inputs.
func TestPassesStreamFilter_NilEdgeCases(t *testing.T) {
req := &inferencev1.StreamAlertsRequest{}
if mapper.PassesStreamFilter(nil, req) {
t.Error("nil alert should fail filter")
}
alert := makeTestAlert("a4", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.EntityType_ENTITY_TYPE_SURFACE)
if mapper.PassesStreamFilter(alert, nil) {
t.Error("nil request should fail filter")
}
}

// TestPassesStreamFilter_DefaultMinSeverity verifies UNSPECIFIED min_severity defaults to WATCH.
func TestPassesStreamFilter_DefaultMinSeverity(t *testing.T) {
watchAlert := makeTestAlert("a5", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.EntityType_ENTITY_TYPE_SURFACE)
normalAlert := makeTestAlert("a6", commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.EntityType_ENTITY_TYPE_SURFACE)

req := &inferencev1.StreamAlertsRequest{
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED, // defaults to WATCH
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}

if !mapper.PassesStreamFilter(watchAlert, req) {
t.Error("WATCH should pass when min_severity is UNSPECIFIED (defaulting to WATCH)")
}
if mapper.PassesStreamFilter(normalAlert, req) {
t.Error("NORMAL should fail when min_severity defaults to WATCH")
}
}

// TestQueuedAlertToProto verifies the pass-through mapping.
func TestQueuedAlertToProto(t *testing.T) {
alert := makeTestAlert("proto-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.EntityType_ENTITY_TYPE_AIR)
qa := &domain.QueuedAlert{Alert: alert}

result := mapper.QueuedAlertToProto(qa)
if result.GetAlertId() != "proto-1" {
t.Errorf("expected proto-1, got %s", result.GetAlertId())
}

// Nil input
if mapper.QueuedAlertToProto(nil) != nil {
t.Error("expected nil for nil QueuedAlert")
}
}

// TestQueuedAlertAckedAtProto verifies timestamp conversion.
func TestQueuedAlertAckedAtProto(t *testing.T) {
now := time.Now()
qa := &domain.QueuedAlert{
Alert:  makeTestAlert("ts-1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, commonv1.AnomalyType_ANOMALY_TYPE_SPEED, commonv1.EntityType_ENTITY_TYPE_SURFACE),
AckedAt: &now,
}

ts := mapper.QueuedAlertAckedAtProto(qa)
if ts == nil {
t.Fatal("expected non-nil timestamp")
}
if ts.AsTime().Unix() != now.Unix() {
t.Errorf("timestamp mismatch")
}

// Nil AckedAt
qa2 := &domain.QueuedAlert{Alert: qa.Alert}
if mapper.QueuedAlertAckedAtProto(qa2) != nil {
t.Error("expected nil timestamp when AckedAt is nil")
}

// Nil QueuedAlert
if mapper.QueuedAlertAckedAtProto(nil) != nil {
t.Error("expected nil for nil input")
}
}
