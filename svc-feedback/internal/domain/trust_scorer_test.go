// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"math"
	"strconv"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-feedback/internal/state"
)

// tolerance for floating-point comparisons.
const tol = 1e-9

func approxEqual(a, b, tolerance float64) bool {
return math.Abs(a-b) <= tolerance
}

// newTestHistory returns a ready-to-use OperatorHistory.
func newTestHistory() *state.OperatorHistory {
return state.NewOperatorHistory()
}

// ─── CalculateClearanceScore ───────────────────────────────────────────────

// T01: Clearance SECRET → score 1.0
func TestCalculateClearanceScore_Secret(t *testing.T) {
score := CalculateClearanceScore(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
if score != 1.0 {
t.Errorf("T01: expected 1.0, got %f", score)
}
}

// T02: Clearance UNCLASSIFIED → score 0.3
func TestCalculateClearanceScore_Unclassified(t *testing.T) {
score := CalculateClearanceScore(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if score != 0.3 {
t.Errorf("T02: expected 0.3, got %f", score)
}
}

func TestCalculateClearanceScore_ProtectedC(t *testing.T) {
score := CalculateClearanceScore(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C)
if score != 0.85 {
t.Errorf("expected 0.85, got %f", score)
}
}

func TestCalculateClearanceScore_ProtectedB(t *testing.T) {
score := CalculateClearanceScore(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B)
if score != 0.70 {
t.Errorf("expected 0.70, got %f", score)
}
}

func TestCalculateClearanceScore_ProtectedA(t *testing.T) {
score := CalculateClearanceScore(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A)
if score != 0.50 {
t.Errorf("expected 0.50, got %f", score)
}
}

func TestCalculateClearanceScore_Unspecified(t *testing.T) {
score := CalculateClearanceScore(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED)
if score != 0.30 {
t.Errorf("expected 0.30 for unspecified, got %f", score)
}
}

// ─── CalculateAccuracyScore ────────────────────────────────────────────────

// T03: Accuracy 8/10 correct → 0.8
func TestCalculateAccuracyScore_EightOfTen(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
now := time.Now()

for i := 0; i < 10; i++ {
validated := i < 8 // 8 validated, 2 not
h.RecordFeedback("op-001", state.FeedbackEntry{
FeedbackID:   "fb-" + strconv.Itoa(i),
TrackID:      "trk-" + strconv.Itoa(i),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now,
Validated:    validated,
})
}

score := ts.CalculateAccuracyScore("op-001")
if !approxEqual(score, 0.8, 1e-6) {
t.Errorf("T03: expected 0.8, got %f", score)
}
}

// T04: Accuracy new operator < 5 → 0.5
func TestCalculateAccuracyScore_NewOperator(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
// Operator with 0 feedback.
score := ts.CalculateAccuracyScore("brand-new-op")
if score != 0.5 {
t.Errorf("T04: expected 0.5 default, got %f", score)
}
}

func TestCalculateAccuracyScore_FewFeedback(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
now := time.Now()
// Only 4 submissions → below minFeedbackForAccuracy=5.
for i := 0; i < 4; i++ {
h.RecordFeedback("op-few", state.FeedbackEntry{
FeedbackID:   "fb-" + strconv.Itoa(i),
TrackID:      "trk-" + strconv.Itoa(i),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now,
Validated:    true,
})
}
score := ts.CalculateAccuracyScore("op-few")
if score != 0.5 {
t.Errorf("expected 0.5 default for < 5 submissions, got %f", score)
}
}

// ─── CalculateTemporalScore ────────────────────────────────────────────────

// T05: Temporal within 5 min → 1.0
func TestCalculateTemporalScore_Within5Min(t *testing.T) {
now := time.Now()
event := now.Add(-3 * time.Minute)
score := CalculateTemporalScore(event, now)
if score != 1.0 {
t.Errorf("T05: expected 1.0, got %f", score)
}
}

// T05b: Exactly at 0 minutes
func TestCalculateTemporalScore_Immediate(t *testing.T) {
now := time.Now()
score := CalculateTemporalScore(now, now)
if score != 1.0 {
t.Errorf("expected 1.0 for immediate feedback, got %f", score)
}
}

// T06: Temporal at 15 min → 0.8 (linear from 1.0 at 5min to 0.5 at 30min)
// progress = (15-5)/(30-5) = 10/25 = 0.4; score = 1.0 - 0.4*(1.0-0.5) = 0.8
func TestCalculateTemporalScore_At15Min(t *testing.T) {
now := time.Now()
event := now.Add(-15 * time.Minute)
score := CalculateTemporalScore(event, now)
expected := 0.8
if !approxEqual(score, expected, 1e-4) {
t.Errorf("T06: expected ~%f, got %f", expected, score)
}
}

// T07: Temporal >2hr → 0.1
func TestCalculateTemporalScore_Over2Hours(t *testing.T) {
now := time.Now()
event := now.Add(-3 * time.Hour)
score := CalculateTemporalScore(event, now)
if score != 0.1 {
t.Errorf("T07: expected 0.1, got %f", score)
}
}

func TestCalculateTemporalScore_At30Min(t *testing.T) {
now := time.Now()
event := now.Add(-30 * time.Minute)
score := CalculateTemporalScore(event, now)
if !approxEqual(score, 0.5, 1e-6) {
t.Errorf("expected 0.5 at 30 min, got %f", score)
}
}

func TestCalculateTemporalScore_At2Hours(t *testing.T) {
now := time.Now()
event := now.Add(-2 * time.Hour)
score := CalculateTemporalScore(event, now)
if !approxEqual(score, 0.1, 1e-6) {
t.Errorf("expected 0.1 at exactly 2hr, got %f", score)
}
}

func TestCalculateTemporalScore_NegativeElapsed(t *testing.T) {
// feedback submitted before event time — clamp to 0 elapsed → 1.0
now := time.Now()
event := now.Add(5 * time.Minute) // event in the future
score := CalculateTemporalScore(event, now)
if score != 1.0 {
t.Errorf("expected 1.0 for negative elapsed, got %f", score)
}
}

// ─── CalculateDeviationScore ───────────────────────────────────────────────

// T08: Deviation matches all consensus → (1-D)=1.0
func TestCalculateDeviationScore_MatchesAll(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
now := time.Now()

// Two prior entries both HOSTILE for trk-X.
h.RecordFeedback("op-A", state.FeedbackEntry{
FeedbackID: "fb-001", TrackID: "trk-X",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now,
})
h.RecordFeedback("op-B", state.FeedbackEntry{
FeedbackID: "fb-002", TrackID: "trk-X",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now,
})

// New submission also HOSTILE → matches all.
score := ts.CalculateDeviationScore("trk-X", commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE)
if score != 1.0 {
t.Errorf("T08: expected 1.0, got %f", score)
}
}

// T09: Deviation contradicts all → (1-D)=0.0
func TestCalculateDeviationScore_ContradictAll(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
now := time.Now()

h.RecordFeedback("op-A", state.FeedbackEntry{
FeedbackID: "fb-001", TrackID: "trk-Y",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now,
})
h.RecordFeedback("op-B", state.FeedbackEntry{
FeedbackID: "fb-002", TrackID: "trk-Y",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now,
})

// New submission is FRIENDLY → contradicts all.
score := ts.CalculateDeviationScore("trk-Y", commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY)
if score != 0.0 {
t.Errorf("T09: expected 0.0, got %f", score)
}
}

// T10: Deviation no other feedback → (1-D)=0.5
func TestCalculateDeviationScore_NoOtherFeedback(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)

score := ts.CalculateDeviationScore("trk-brand-new", commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE)
if score != 0.5 {
t.Errorf("T10: expected 0.5, got %f", score)
}
}

func TestCalculateDeviationScore_Partial(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
now := time.Now()

// 2 HOSTILE, 1 FRIENDLY.
h.RecordFeedback("op-A", state.FeedbackEntry{FeedbackID: "fb-001", TrackID: "trk-Z", FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now})
h.RecordFeedback("op-B", state.FeedbackEntry{FeedbackID: "fb-002", TrackID: "trk-Z", FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now})
h.RecordFeedback("op-C", state.FeedbackEntry{FeedbackID: "fb-003", TrackID: "trk-Z", FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY, Timestamp: now})

// New submission is HOSTILE → matches 2/3.
score := ts.CalculateDeviationScore("trk-Z", commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE)
expected := 2.0 / 3.0
if !approxEqual(score, expected, 1e-9) {
t.Errorf("expected %f partial match, got %f", expected, score)
}
}

// ─── Score (composite) ─────────────────────────────────────────────────────

// T11: SECRET, accuracy=0.8, 3min, matches all → 0.2*1.0+0.3*0.8+0.2*1.0+0.3*1.0 = 0.94
func TestScore_T11_HighTrustComposite(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
now := time.Now()

// Seed 10 validated submissions for op-secret.
for i := 0; i < 10; i++ {
h.RecordFeedback("op-secret", state.FeedbackEntry{
FeedbackID:   "fb-s" + strconv.Itoa(i),
TrackID:      "trk-s" + strconv.Itoa(i),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now,
Validated:    i < 8, // 8/10 = 0.8
})
}

// Pre-populate track consensus: 2 matching entries.
h.RecordFeedback("op-other1", state.FeedbackEntry{FeedbackID: "fb-c1", TrackID: "trk-target", FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now})
h.RecordFeedback("op-other2", state.FeedbackEntry{FeedbackID: "fb-c2", TrackID: "trk-target", FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now})

params := TrustParams{
OperatorID:        "op-secret",
OperatorClearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
TrackID:           "trk-target",
FeedbackType:      commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
EventTime:         now.Add(-3 * time.Minute),
FeedbackTime:      now,
}

result := ts.Score(params)
expected := 0.2*1.0 + 0.3*0.8 + 0.2*1.0 + 0.3*1.0 // = 0.94
if !approxEqual(result.TotalScore, expected, 1e-6) {
t.Errorf("T11: expected %f, got %f", expected, result.TotalScore)
}
if !result.Validated {
t.Error("T11: expected validated=true")
}
}

// T12: UNCLASSIFIED, accuracy=0.3, 3hr, contradicts → 0.2*0.3+0.3*0.3+0.2*0.1+0.3*0.0 = 0.17
func TestScore_T12_LowTrustComposite(t *testing.T) {
h := newTestHistory()
ts := NewTrustScorer(h)
now := time.Now()

// Seed 10 submissions: 3 validated for accuracy=0.3.
for i := 0; i < 10; i++ {
h.RecordFeedback("op-unc", state.FeedbackEntry{
FeedbackID:   "fb-u" + strconv.Itoa(i),
TrackID:      "trk-u" + strconv.Itoa(i),
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    now,
Validated:    i < 3, // 3/10 = 0.3
})
}

// Populate consensus: all HOSTILE for trk-W.
h.RecordFeedback("op-A", state.FeedbackEntry{FeedbackID: "fb-w1", TrackID: "trk-W", FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now})
h.RecordFeedback("op-B", state.FeedbackEntry{FeedbackID: "fb-w2", TrackID: "trk-W", FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE, Timestamp: now})

params := TrustParams{
OperatorID:        "op-unc",
OperatorClearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
TrackID:           "trk-W",
FeedbackType:      commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY, // contradicts
EventTime:         now.Add(-3 * time.Hour),
FeedbackTime:      now,
}

result := ts.Score(params)
// C=0.3, A=0.3, T=0.1, (1-D)=0.0
expected := 0.2*0.3 + 0.3*0.3 + 0.2*0.1 + 0.3*0.0 // = 0.17
if !approxEqual(result.TotalScore, expected, 1e-6) {
t.Errorf("T12: expected %f, got %f", expected, result.TotalScore)
}
if result.Validated {
t.Error("T12: expected validated=false")
}
}

// ─── Table-driven clearance score ─────────────────────────────────────────

func TestCalculateClearanceScore_Table(t *testing.T) {
cases := []struct {
name     string
level    commonv1.ClassificationLevel
expected float64
}{
{"SECRET", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 1.00},
{"PROTECTED_C", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C, 0.85},
{"PROTECTED_B", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B, 0.70},
{"PROTECTED_A", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A, 0.50},
{"UNCLASSIFIED", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.30},
{"UNSPECIFIED", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED, 0.30},
}
for _, tc := range cases {
tc := tc
t.Run(tc.name, func(t *testing.T) {
got := CalculateClearanceScore(tc.level)
if got != tc.expected {
t.Errorf("expected %f, got %f", tc.expected, got)
}
})
}
}

// ─── Table-driven temporal score ──────────────────────────────────────────

func TestCalculateTemporalScore_Table(t *testing.T) {
now := time.Now()
cases := []struct {
name     string
elapsed  time.Duration
expected float64
tol      float64
}{
{"0min", 0, 1.0, 1e-6},
{"4min59s", 4*time.Minute + 59*time.Second, 1.0, 1e-6},
{"5min", 5 * time.Minute, 1.0, 1e-6},
{"15min", 15 * time.Minute, 0.8, 1e-4},
{"30min", 30 * time.Minute, 0.5, 1e-6},
{"60min", 60 * time.Minute, 11.0 / 30.0, 1e-4},
{"120min", 120 * time.Minute, 0.1, 1e-6},
{"180min", 180 * time.Minute, 0.1, 1e-6},
}
for _, tc := range cases {
tc := tc
t.Run(tc.name, func(t *testing.T) {
score := CalculateTemporalScore(now.Add(-tc.elapsed), now)
if !approxEqual(score, tc.expected, tc.tol) {
t.Errorf("elapsed=%v: expected %f, got %f", tc.elapsed, tc.expected, score)
}
})
}
}
