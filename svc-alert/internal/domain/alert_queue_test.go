// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

// makeAlert is a test helper that creates a minimal AnomalyAlert.
func makeAlert(id string, severity commonv1.AlertSeverity, detectedAt time.Time) *inferencev1.AnomalyAlert {
return &inferencev1.AnomalyAlert{
AlertId:        id,
TrackId:        "track-" + id,
Severity:       severity,
AnomalyType:    commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
DetectedAt:     timestamppb.New(detectedAt),
}
}

// T01: Enqueueing CRITICAL then WATCH → CRITICAL appears first in priority order.
func TestAlertQueue_T01_CriticalBeforeWatch(t *testing.T) {
q := domain.NewAlertQueue(100)

now := time.Now()
watchAlert := makeAlert("watch-1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, now.Add(-1*time.Second))
criticalAlert := makeAlert("critical-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now)

q.Enqueue(watchAlert)
q.Enqueue(criticalAlert)

unacked := q.GetUnacknowledged()
if len(unacked) != 2 {
t.Fatalf("expected 2 unacknowledged alerts, got %d", len(unacked))
}
if unacked[0].Alert.GetAlertId() != "critical-1" {
t.Errorf("expected critical-1 first, got %s", unacked[0].Alert.GetAlertId())
}
if unacked[1].Alert.GetAlertId() != "watch-1" {
t.Errorf("expected watch-1 second, got %s", unacked[1].Alert.GetAlertId())
}
}

// T01b: Within the same severity, newer alerts come first.
func TestAlertQueue_T01b_NewerTimestampFirst(t *testing.T) {
q := domain.NewAlertQueue(100)

now := time.Now()
older := makeAlert("critical-old", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now.Add(-10*time.Second))
newer := makeAlert("critical-new", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now)

q.Enqueue(older)
q.Enqueue(newer)

unacked := q.GetUnacknowledged()
if len(unacked) != 2 {
t.Fatalf("expected 2 alerts, got %d", len(unacked))
}
if unacked[0].Alert.GetAlertId() != "critical-new" {
t.Errorf("expected critical-new first (newer), got %s", unacked[0].Alert.GetAlertId())
}
}

// T02: Enqueueing at maxSize drops the lowest-priority alert.
func TestAlertQueue_T02_MaxSizeDropsLowest(t *testing.T) {
const maxSize = 3
q := domain.NewAlertQueue(maxSize)

now := time.Now()
// Fill queue with WATCH alerts
q.Enqueue(makeAlert("watch-1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, now.Add(-3*time.Second)))
q.Enqueue(makeAlert("watch-2", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, now.Add(-2*time.Second)))
q.Enqueue(makeAlert("watch-3", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, now.Add(-1*time.Second)))

if q.Size() != maxSize {
t.Fatalf("expected queue size %d, got %d", maxSize, q.Size())
}

// Enqueue a CRITICAL alert — should cause a WATCH (lowest) to be dropped
q.Enqueue(makeAlert("critical-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now))

if q.Size() != maxSize {
t.Fatalf("expected queue size still %d after eviction, got %d", maxSize, q.Size())
}

// CRITICAL should still be in the queue
_, found := q.Get("critical-1")
if !found {
t.Error("expected critical-1 to remain in queue after eviction")
}
}

// T03: Acknowledging an existing alert sets Acknowledged=true and records the operator.
func TestAlertQueue_T03_AcknowledgeExisting(t *testing.T) {
q := domain.NewAlertQueue(100)

alert := makeAlert("alert-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, time.Now())
q.Enqueue(alert)

ackedAt, err := q.Acknowledge("alert-1", "op-001", "Confirmed benign")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if ackedAt == nil {
t.Fatal("expected non-nil ackedAt timestamp")
}

qa, found := q.Get("alert-1")
if !found {
t.Fatal("alert not found after acknowledgment")
}
if !qa.Acknowledged {
t.Error("expected Acknowledged=true")
}
if qa.AckedBy != "op-001" {
t.Errorf("expected AckedBy=op-001, got %s", qa.AckedBy)
}
if qa.Comment != "Confirmed benign" {
t.Errorf("expected Comment='Confirmed benign', got %s", qa.Comment)
}
if !qa.Alert.GetAcknowledged() {
t.Error("expected Alert.Acknowledged=true (proto field)")
}
}

// T04: Acknowledging a non-existent alert returns ErrAlertNotFound.
func TestAlertQueue_T04_AcknowledgeNonExistent(t *testing.T) {
q := domain.NewAlertQueue(100)

_, err := q.Acknowledge("does-not-exist", "op-001", "")
if err == nil {
t.Fatal("expected error for non-existent alert, got nil")
}
}

// T05: GetUnacknowledged returns only unacknowledged alerts.
func TestAlertQueue_T05_GetUnacknowledgedFilter(t *testing.T) {
q := domain.NewAlertQueue(100)

now := time.Now()
q.Enqueue(makeAlert("alert-ack", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, now))
q.Enqueue(makeAlert("alert-unack", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now.Add(time.Second)))

_, err := q.Acknowledge("alert-ack", "op-002", "")
if err != nil {
t.Fatalf("acknowledge failed: %v", err)
}

unacked := q.GetUnacknowledged()
if len(unacked) != 1 {
t.Fatalf("expected 1 unacknowledged, got %d", len(unacked))
}
if unacked[0].Alert.GetAlertId() != "alert-unack" {
t.Errorf("expected alert-unack, got %s", unacked[0].Alert.GetAlertId())
}
}

// TestAlertQueue_EnqueueNilIgnored ensures nil and empty-ID alerts are silently ignored.
func TestAlertQueue_EnqueueNilIgnored(t *testing.T) {
q := domain.NewAlertQueue(100)
q.Enqueue(nil)
q.Enqueue(&inferencev1.AnomalyAlert{}) // empty ID
if q.Size() != 0 {
t.Errorf("expected empty queue, got %d", q.Size())
}
}

// TestAlertQueue_UpdateExisting verifies that enqueueing a duplicate ID updates in place.
func TestAlertQueue_UpdateExisting(t *testing.T) {
q := domain.NewAlertQueue(100)

a := makeAlert("alert-1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, time.Now())
q.Enqueue(a)

// Update same ID with higher severity
updated := makeAlert("alert-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, time.Now())
q.Enqueue(updated)

if q.Size() != 1 {
t.Fatalf("expected 1 alert, got %d", q.Size())
}
qa, found := q.Get("alert-1")
if !found {
t.Fatal("alert not found")
}
if qa.Alert.GetSeverity() != commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL {
t.Errorf("expected severity CRITICAL after update, got %v", qa.Alert.GetSeverity())
}
}

// TestAlertQueue_SeverityOrder verifies full ordering: CRITICAL > ELEVATED > WATCH > NORMAL.
func TestAlertQueue_SeverityOrder(t *testing.T) {
q := domain.NewAlertQueue(100)

now := time.Now()
q.Enqueue(makeAlert("normal-1", commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL, now))
q.Enqueue(makeAlert("watch-1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, now))
q.Enqueue(makeAlert("elevated-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, now))
q.Enqueue(makeAlert("critical-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now))

unacked := q.GetUnacknowledged()
if len(unacked) != 4 {
t.Fatalf("expected 4 alerts, got %d", len(unacked))
}

want := []string{"critical-1", "elevated-1", "watch-1", "normal-1"}
for i, qa := range unacked {
if qa.Alert.GetAlertId() != want[i] {
t.Errorf("position %d: expected %s, got %s", i, want[i], qa.Alert.GetAlertId())
}
}
}

// TestAlertQueue_Subscribe_ReceivesNewAlerts verifies that subscribers receive notifications.
func TestAlertQueue_Subscribe_ReceivesNewAlerts(t *testing.T) {
q := domain.NewAlertQueue(100)

ch := q.Subscribe()
defer q.Unsubscribe(ch)

alert := makeAlert("stream-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, time.Now())
q.Enqueue(alert)

select {
case received := <-ch:
if received.GetAlertId() != "stream-1" {
t.Errorf("expected stream-1, got %s", received.GetAlertId())
}
case <-time.After(100 * time.Millisecond):
t.Error("timed out waiting for subscriber notification")
}
}

// TestAlertQueue_UnacknowledgedCount groups counts by severity.
func TestAlertQueue_UnacknowledgedCount(t *testing.T) {
q := domain.NewAlertQueue(100)

now := time.Now()
q.Enqueue(makeAlert("c1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now))
q.Enqueue(makeAlert("c2", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, now))
q.Enqueue(makeAlert("w1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, now))

counts := q.UnacknowledgedCount()
if counts[commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL] != 2 {
t.Errorf("expected 2 critical, got %d", counts[commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL])
}
if counts[commonv1.AlertSeverity_ALERT_SEVERITY_WATCH] != 1 {
t.Errorf("expected 1 watch, got %d", counts[commonv1.AlertSeverity_ALERT_SEVERITY_WATCH])
}
}
