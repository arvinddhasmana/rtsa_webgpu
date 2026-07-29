// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"strconv"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-feedback/internal/state"
"go.uber.org/zap"
)

func newTestGuard() (*AntiPoisonGuard, *state.OperatorHistory) {
h := state.NewOperatorHistory()
logger, _ := zap.NewDevelopment()
guard := NewAntiPoisonGuard(h, logger)
return guard, h
}

// recordN records n feedback entries for operatorID, all on distinct tracks,
// with the specified validated flag.
func recordN(h *state.OperatorHistory, operatorID string, n int, ft commonv1.FeedbackType, validated bool, source string) {
now := time.Now()
for i := 0; i < n; i++ {
h.RecordFeedback(operatorID, state.FeedbackEntry{
FeedbackID:   "fb-" + operatorID + "-" + strconv.Itoa(i),
TrackID:      "trk-" + strconv.Itoa(i),
FeedbackType: ft,
Timestamp:    now,
Validated:    validated,
SensorSource: source,
})
}
}

// T13: Anti-poison: all checks pass
func TestAntiPoison_T13_AllPass(t *testing.T) {
guard, h := newTestGuard()
now := time.Now()

// Record 15 entries with diverse types, sources, spread over time.
types := []commonv1.FeedbackType{
commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY,
commonv1.FeedbackType_FEEDBACK_TYPE_RECLASSIFY,
commonv1.FeedbackType_FEEDBACK_TYPE_REJECT_ANOMALY,
commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_ANOMALY,
}
sources := []string{"radar-A", "ew-B", "isr-C", "ais-D"}
for i := 0; i < 15; i++ {
h.RecordFeedback("op-clean", state.FeedbackEntry{
FeedbackID:   "fb-" + string(rune('a'+i)),
TrackID:      "trk-" + string(rune('a'+i)),
FeedbackType: types[i%len(types)],
// Spread across 15 minutes so no burst.
Timestamp:    now.Add(time.Duration(i) * time.Minute),
Validated:    true,
SensorSource: sources[i%len(sources)],
})
}

result := guard.Check("op-clean")
if !result.Passed {
t.Errorf("T13: expected all checks to pass, got: %v | details: %v", result.Checks, result.Details)
}
}

// T14: Anti-poison: label flip > 20%
func TestAntiPoison_T14_LabelFlipFails(t *testing.T) {
guard, h := newTestGuard()
now := time.Now()

// First: establish a consistent track for flipping.
for i := 0; i < 5; i++ {
ft := commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE
if i%2 != 0 {
ft = commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY
}
h.RecordFeedback("op-flipper", state.FeedbackEntry{
FeedbackID:   "fb-" + string(rune('a'+i)),
TrackID:      "trk-flip",
FeedbackType: ft,
Timestamp:    now.Add(time.Duration(i) * time.Minute),
Validated:    true,
SensorSource: "radar-A",
})
}

result := guard.Check("op-flipper")
if result.Checks["label_flip"] {
t.Error("T14: expected label_flip check to fail")
}
}

// T15: Anti-poison: < 3 sensor sources
func TestAntiPoison_T15_SourceDiversityFails(t *testing.T) {
guard, h := newTestGuard()
now := time.Now()

// 6 submissions, all from single sensor source.
for i := 0; i < 6; i++ {
h.RecordFeedback("op-narrow", state.FeedbackEntry{
FeedbackID:   "fb-" + string(rune('a'+i)),
TrackID:      "trk-" + string(rune('a'+i)),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now.Add(time.Duration(i) * time.Minute),
Validated:    true,
SensorSource: "only-radar", // single source
})
}

result := guard.Check("op-narrow")
if result.Checks["source_diversity"] {
t.Errorf("T15: expected source_diversity to fail, got checks: %v", result.Checks)
}
}

// T16: Anti-poison: burst >40% in 5-min window
func TestAntiPoison_T16_TemporalClusteringFails(t *testing.T) {
guard, h := newTestGuard()
now := time.Now()

// Total: 10 entries. 5 in the same 1-minute window → 50% > 40%.
for i := 0; i < 5; i++ {
h.RecordFeedback("op-burst", state.FeedbackEntry{
FeedbackID:   "fb-burst-" + string(rune('a'+i)),
TrackID:      "trk-" + string(rune('a'+i)),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now, // all at same time
Validated:    true,
SensorSource: "radar-A",
})
}
for i := 5; i < 10; i++ {
h.RecordFeedback("op-burst", state.FeedbackEntry{
FeedbackID:   "fb-spread-" + string(rune('a'+i)),
TrackID:      "trk-" + string(rune('p'+i-5)),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now.Add(time.Duration(i*20) * time.Minute), // spread
Validated:    true,
SensorSource: "radar-B",
})
}

result := guard.Check("op-burst")
if result.Checks["temporal_clustering"] {
t.Errorf("T16: expected temporal_clustering to fail, got checks: %v | details: %v", result.Checks, result.Details)
}
}

// T17: Anti-poison: < 60% validated
func TestAntiPoison_T17_HighTrustRatioFails(t *testing.T) {
guard, h := newTestGuard()
now := time.Now()

// 12 submissions, only 5 validated (41.7% < 60%).
for i := 0; i < 12; i++ {
h.RecordFeedback("op-lowvalid", state.FeedbackEntry{
FeedbackID:   "fb-" + string(rune('a'+i)),
TrackID:      "trk-" + string(rune('a'+i)),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now.Add(time.Duration(i) * time.Minute),
Validated:    i < 5,
SensorSource: "radar-A",
})
}

result := guard.Check("op-lowvalid")
if result.Checks["high_trust_ratio"] {
t.Errorf("T17: expected high_trust_ratio to fail, got checks: %v", result.Checks)
}
}

// New operator → all auto-pass
func TestAntiPoison_NewOperator_AutoPass(t *testing.T) {
guard, _ := newTestGuard()
result := guard.Check("op-brand-new")
if !result.Passed {
t.Error("expected auto-pass for new operator with no history")
}
for name, passed := range result.Checks {
if !passed {
t.Errorf("check %s should auto-pass for new operator", name)
}
}
}

// Distribution check auto-passes below 10 submissions
func TestAntiPoison_DistributionAutoPassBelow10(t *testing.T) {
guard, h := newTestGuard()
now := time.Now()
// Only 9 submissions.
for i := 0; i < 9; i++ {
h.RecordFeedback("op-dist", state.FeedbackEntry{
FeedbackID:   "fb-" + string(rune('a'+i)),
TrackID:      "trk-" + string(rune('a'+i)),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, // biased
Timestamp:    now.Add(time.Duration(i) * time.Minute),
Validated:    true,
SensorSource: "radar-A",
})
}
result := guard.Check("op-dist")
if !result.Checks["distribution"] {
t.Error("distribution should auto-pass below 10 submissions")
}
}

// Source diversity auto-passes below 5 submissions
func TestAntiPoison_SourceDiversityAutoPassBelow5(t *testing.T) {
guard, h := newTestGuard()
now := time.Now()
for i := 0; i < 4; i++ {
h.RecordFeedback("op-src", state.FeedbackEntry{
FeedbackID:   "fb-" + string(rune('a'+i)),
TrackID:      "trk-" + string(rune('a'+i)),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now,
Validated:    true,
SensorSource: "only-one", // would fail if check applied
})
}
result := guard.Check("op-src")
if !result.Checks["source_diversity"] {
t.Error("source_diversity should auto-pass below 5 submissions")
}
}

func TestChiSquaredPValue_HighStat(t *testing.T) {
// Very high chi-squared should give very low p-value → non-uniform distribution.
p := chiSquaredPValue(20.0, 4)
if p >= 0.05 {
t.Errorf("expected p < 0.05 for chiSq=20, df=4, got %f", p)
}
}

func TestChiSquaredPValue_LowStat(t *testing.T) {
// Very low chi-squared should give high p-value → uniform distribution.
p := chiSquaredPValue(0.5, 4)
if p <= 0.05 {
t.Errorf("expected p > 0.05 for chiSq=0.5, df=4, got %f", p)
}
}

func TestChiSquaredPValue_Zero(t *testing.T) {
p := chiSquaredPValue(0, 4)
if p != 1.0 {
t.Errorf("expected p=1.0 for chiSq=0, got %f", p)
}
}
