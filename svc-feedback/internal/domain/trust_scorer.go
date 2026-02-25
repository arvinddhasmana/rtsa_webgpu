// CLASSIFICATION: UNCLASSIFIED

// Package domain contains core business logic for the feedback service:
// trust scoring, anti-poisoning guards, and rate limiting.
package domain

import (
"math"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/state"
)

const (
weightClearance = 0.2
weightAccuracy  = 0.3
weightTemporal  = 0.2
weightDeviation = 0.3

// Minimum feedback count before accuracy is computed (default 0.5 below threshold).
minFeedbackForAccuracy = 5
)

// TrustParams carries the inputs required to compute a trust score.
type TrustParams struct {
OperatorID        string
OperatorClearance commonv1.ClassificationLevel
TrackID           string
FeedbackType      commonv1.FeedbackType
EventTime         time.Time // when the tracked entity/alert event occurred
FeedbackTime      time.Time // when the operator submitted feedback
}

// TrustResult holds the computed trust score and its component breakdown.
type TrustResult struct {
TotalScore     float64 // 0.0 to 1.0
ClearanceScore float64
AccuracyScore  float64
TemporalScore  float64
// DeviationScore is the consensus-adjusted component (1-D), range 0.0–1.0.
DeviationScore float64
Validated      bool // true if TotalScore >= 0.5
}

// TrustScorer computes the 4-component trust score for operator feedback.
type TrustScorer struct {
history *state.OperatorHistory
}

// NewTrustScorer constructs a TrustScorer backed by the given history store.
func NewTrustScorer(history *state.OperatorHistory) *TrustScorer {
return &TrustScorer{history: history}
}

// Score computes Trust = 0.2*C + 0.3*A + 0.2*T + 0.3*(1-D).
func (ts *TrustScorer) Score(params TrustParams) *TrustResult {
c := CalculateClearanceScore(params.OperatorClearance)
a := ts.CalculateAccuracyScore(params.OperatorID)
t := CalculateTemporalScore(params.EventTime, params.FeedbackTime)
oneMinusD := ts.CalculateDeviationScore(params.TrackID, params.FeedbackType)

total := weightClearance*c + weightAccuracy*a + weightTemporal*t + weightDeviation*oneMinusD
total = math.Round(total*1e10) / 1e10 // avoid floating-point drift

return &TrustResult{
TotalScore:     total,
ClearanceScore: c,
AccuracyScore:  a,
TemporalScore:  t,
DeviationScore: oneMinusD,
Validated:      total >= 0.5,
}
}

// CalculateClearanceScore maps a ClassificationLevel to a clearance weight.
//
//SECRET       → 1.00
//PROTECTED_C  → 0.85
//PROTECTED_B  → 0.70
//PROTECTED_A  → 0.50
//UNCLASSIFIED → 0.30
//others       → 0.30
func CalculateClearanceScore(level commonv1.ClassificationLevel) float64 {
switch level {
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET:
return 1.00
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C:
return 0.85
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B:
return 0.70
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A:
return 0.50
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED:
return 0.30
default:
return 0.30
}
}

// CalculateAccuracyScore returns the operator's historical accuracy ratio.
// Returns the default 0.5 for operators with fewer than minFeedbackForAccuracy
// submissions, avoiding cold-start penalisation.
func (ts *TrustScorer) CalculateAccuracyScore(operatorID string) float64 {
stats := ts.history.GetStats(operatorID)
if stats == nil || stats.TotalFeedback < minFeedbackForAccuracy {
return 0.5
}
return float64(stats.ConfirmedCorrect) / float64(stats.TotalFeedback)
}

// CalculateTemporalScore decays based on elapsed time between event and feedback.
//
//≤ 5 min          → 1.0
//5 min – 30 min   → linear decay 1.0 → 0.5
//30 min – 2 hr    → linear decay 0.5 → 0.1
//> 2 hr           → 0.1
func CalculateTemporalScore(eventTime, feedbackTime time.Time) float64 {
elapsed := feedbackTime.Sub(eventTime)
if elapsed < 0 {
elapsed = 0
}

const (
t5m  = 5 * time.Minute
t30m = 30 * time.Minute
t2h  = 2 * time.Hour
)

switch {
case elapsed <= t5m:
return 1.0
case elapsed <= t30m:
// Linear from 1.0 at 5 min to 0.5 at 30 min.
progress := float64(elapsed-t5m) / float64(t30m-t5m)
return 1.0 - progress*(1.0-0.5)
case elapsed <= t2h:
// Linear from 0.5 at 30 min to 0.1 at 2 hr.
progress := float64(elapsed-t30m) / float64(t2h-t30m)
return 0.5 - progress*(0.5-0.1)
default:
return 0.1
}
}

// CalculateDeviationScore returns (1-D) — the consensus-alignment component.
//
//Matches ALL other feedback on this track  → 1.0
//Contradicts ALL other feedback             → 0.0
//Mixed / no other feedback exists           → 0.5
func (ts *TrustScorer) CalculateDeviationScore(trackID string, feedbackType commonv1.FeedbackType) float64 {
entries := ts.history.GetFeedbackByTrack(trackID)
if len(entries) == 0 {
// No prior feedback for this track — default to neutral.
return 0.5
}

matches := 0
total := len(entries)
for _, e := range entries {
if e.FeedbackType == feedbackType {
matches++
}
}

switch {
case matches == total:
return 1.0
case matches == 0:
return 0.0
default:
// Partial consensus — normalise to [0, 1].
return float64(matches) / float64(total)
}
}
