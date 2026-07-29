// CLASSIFICATION: UNCLASSIFIED

package state

import (
"strconv"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
feedbackv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1"
)

func TestNewOperatorHistory(t *testing.T) {
oh := NewOperatorHistory()
if oh == nil {
t.Fatal("expected non-nil OperatorHistory")
}
if oh.operators == nil {
t.Fatal("expected initialised operators map")
}
}

func TestRecordFeedback_FirstEntry(t *testing.T) {
oh := NewOperatorHistory()
entry := FeedbackEntry{
FeedbackID:   "fb-001",
TrackID:      "trk-01",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    time.Now(),
TrustScore:   0.8,
SensorSource: "radar-A",
Validated:    true,
}
oh.RecordFeedback("op-001", entry)

stats := oh.GetStats("op-001")
if stats == nil {
t.Fatal("expected stats for op-001")
}
if stats.TotalFeedback != 1 {
t.Errorf("expected TotalFeedback=1, got %d", stats.TotalFeedback)
}
if stats.ConfirmedCorrect != 1 {
t.Errorf("expected ConfirmedCorrect=1, got %d", stats.ConfirmedCorrect)
}
if stats.LabelFlipCount != 0 {
t.Errorf("expected LabelFlipCount=0, got %d", stats.LabelFlipCount)
}
if !stats.TrackSensorSources["radar-A"] {
t.Error("expected sensor source radar-A recorded")
}
}

func TestRecordFeedback_LabelFlipDetected(t *testing.T) {
oh := NewOperatorHistory()
now := time.Now()

first := FeedbackEntry{
FeedbackID:   "fb-001",
TrackID:      "trk-01",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now,
}
second := FeedbackEntry{
FeedbackID:   "fb-002",
TrackID:      "trk-01",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY, // Different type → flip
Timestamp:    now.Add(time.Minute),
}

oh.RecordFeedback("op-001", first)
oh.RecordFeedback("op-001", second)

stats := oh.GetStats("op-001")
if stats.LabelFlipCount != 1 {
t.Errorf("expected LabelFlipCount=1, got %d", stats.LabelFlipCount)
}
}

func TestRecordFeedback_NoFlipOnSameType(t *testing.T) {
oh := NewOperatorHistory()
now := time.Now()
for i := 0; i < 3; i++ {
oh.RecordFeedback("op-001", FeedbackEntry{
FeedbackID:   "fb-00" + strconv.Itoa(i+1),
TrackID:      "trk-01",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now.Add(time.Duration(i) * time.Minute),
})
}
stats := oh.GetStats("op-001")
if stats.LabelFlipCount != 0 {
t.Errorf("expected no label flips, got %d", stats.LabelFlipCount)
}
}

func TestGetStats_UnknownOperator(t *testing.T) {
oh := NewOperatorHistory()
stats := oh.GetStats("unknown-op")
if stats != nil {
t.Errorf("expected nil stats for unknown operator, got %+v", stats)
}
}

func TestOperatorStats_TypeDistribution(t *testing.T) {
oh := NewOperatorHistory()
now := time.Now()
types := []commonv1.FeedbackType{
commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY,
}
for i, ft := range types {
oh.RecordFeedback("op-001", FeedbackEntry{
FeedbackID:   "fb-00" + strconv.Itoa(i+1),
TrackID:      "trk-0" + strconv.Itoa(i+1),
FeedbackType: ft,
Timestamp:    now,
})
}

stats := oh.GetStats("op-001")
dist := stats.TypeDistribution()

hostileKey := commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE.String()
friendlyKey := commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY.String()

if dist[hostileKey] < 0.66 || dist[hostileKey] > 0.68 {
t.Errorf("expected hostile ratio ~0.667, got %f", dist[hostileKey])
}
if dist[friendlyKey] < 0.32 || dist[friendlyKey] > 0.34 {
t.Errorf("expected friendly ratio ~0.333, got %f", dist[friendlyKey])
}
}

func TestOperatorStats_LabelFlipRate(t *testing.T) {
stats := &OperatorStats{
TotalFeedback:  10,
LabelFlipCount: 2,
}
rate := stats.LabelFlipRate()
if rate != 0.2 {
t.Errorf("expected rate 0.2, got %f", rate)
}
}

func TestOperatorStats_LabelFlipRate_ZeroTotal(t *testing.T) {
stats := &OperatorStats{}
if stats.LabelFlipRate() != 0.0 {
t.Error("expected 0.0 for empty stats")
}
}

func TestOperatorStats_UniqueSensorSources(t *testing.T) {
stats := &OperatorStats{
TrackSensorSources: map[string]bool{"radar-A": true, "ew-B": true},
}
if stats.UniqueSensorSources() != 2 {
t.Errorf("expected 2 unique sources, got %d", stats.UniqueSensorSources())
}
}

func TestOperatorStats_ValidatedRatio(t *testing.T) {
stats := &OperatorStats{
TotalFeedback:    10,
ConfirmedCorrect: 7,
}
if stats.ValidatedRatio() != 0.7 {
t.Errorf("expected 0.7, got %f", stats.ValidatedRatio())
}
}

func TestGetFeedbackByTrack(t *testing.T) {
oh := NewOperatorHistory()
now := time.Now()

oh.RecordFeedback("op-001", FeedbackEntry{
FeedbackID:   "fb-001",
TrackID:      "trk-A",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now,
})
oh.RecordFeedback("op-002", FeedbackEntry{
FeedbackID:   "fb-002",
TrackID:      "trk-A",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY,
Timestamp:    now,
})

entries := oh.GetFeedbackByTrack("trk-A")
if len(entries) != 2 {
t.Errorf("expected 2 entries for trk-A, got %d", len(entries))
}
}

func TestQueryHistory_FilterByOperator(t *testing.T) {
oh := NewOperatorHistory()
now := time.Now()

fb1 := &feedbackv1.OperatorFeedback{FeedbackId: "fb-001", OperatorId: "op-001", TrackId: "trk-A"}
fb2 := &feedbackv1.OperatorFeedback{FeedbackId: "fb-002", OperatorId: "op-002", TrackId: "trk-B"}
oh.RecordProto(fb1)
oh.RecordProto(fb2)
_ = now

results := oh.QueryHistory("op-001", "")
if len(results) != 1 || results[0].GetFeedbackId() != "fb-001" {
t.Errorf("expected 1 result for op-001, got %d", len(results))
}
}

func TestQueryHistory_FilterByTrack(t *testing.T) {
oh := NewOperatorHistory()
oh.RecordProto(&feedbackv1.OperatorFeedback{FeedbackId: "fb-001", OperatorId: "op-001", TrackId: "trk-A"})
oh.RecordProto(&feedbackv1.OperatorFeedback{FeedbackId: "fb-002", OperatorId: "op-002", TrackId: "trk-B"})

results := oh.QueryHistory("", "trk-B")
if len(results) != 1 || results[0].GetFeedbackId() != "fb-002" {
t.Errorf("expected 1 result for trk-B, got %d", len(results))
}
}

func TestQueryHistory_NoFilter(t *testing.T) {
oh := NewOperatorHistory()
oh.RecordProto(&feedbackv1.OperatorFeedback{FeedbackId: "fb-001"})
oh.RecordProto(&feedbackv1.OperatorFeedback{FeedbackId: "fb-002"})

results := oh.QueryHistory("", "")
if len(results) != 2 {
t.Errorf("expected 2 results with no filter, got %d", len(results))
}
}

func TestConcurrentRecordFeedback(t *testing.T) {
oh := NewOperatorHistory()
done := make(chan struct{})

for i := 0; i < 100; i++ {
go func(n int) {
oh.RecordFeedback("op-concurrent", FeedbackEntry{
FeedbackID:   "fb-" + strconv.Itoa(n),
TrackID:      "trk-concurrent",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    time.Now(),
})
done <- struct{}{}
}(i)
}

for i := 0; i < 100; i++ {
<-done
}

stats := oh.GetStats("op-concurrent")
if stats == nil {
t.Fatal("expected stats after concurrent writes")
}
if stats.TotalFeedback != 100 {
t.Errorf("expected 100 feedback entries, got %d", stats.TotalFeedback)
}
}
