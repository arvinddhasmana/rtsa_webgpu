// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"math"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/state"
"go.uber.org/zap"
)

const (
// Thresholds and minimums for anti-poisoning checks.
chiSquaredPValueThreshold  = 0.05 // p > 0.05 → pass (uniform distribution)
labelFlipRateThreshold     = 0.20 // ≤ 20% flips → pass
minSourceDiversity         = 3    // ≥ 3 unique sources → pass
burstWindowDuration        = 5 * time.Minute
burstRatioThreshold        = 0.40 // > 40% in any 5-min window → fail
validatedRatioThreshold    = 0.60 // ≥ 60% validated → pass

// Minimum submission counts before a check is applied.
minForDistribution = 10
minForLabelFlip    = 5
minForSource       = 5
minForHighTrust    = 10
)

// PoisonCheckResult holds the outcome of all anti-poisoning checks.
type PoisonCheckResult struct {
Passed  bool
Checks  map[string]bool   // check name → passed
Details map[string]string // check name → human-readable detail
}

// AntiPoisonGuard implements 5 statistical checks to detect bulk anomalous
// feedback patterns consistent with ML data-poisoning attacks.
type AntiPoisonGuard struct {
history *state.OperatorHistory
logger  *zap.Logger
}

// NewAntiPoisonGuard constructs an AntiPoisonGuard.
func NewAntiPoisonGuard(history *state.OperatorHistory, logger *zap.Logger) *AntiPoisonGuard {
return &AntiPoisonGuard{history: history, logger: logger}
}

// Check runs all 5 anti-poisoning checks against the operator's history.
// Returns Passed=true only if every applicable check passes.
// Anti-poisoning failure NEVER blocks submission — it is the caller's
// responsibility to act on the result (e.g. force trust_score below 0.5).
func (g *AntiPoisonGuard) Check(operatorID string) *PoisonCheckResult {
result := &PoisonCheckResult{
Checks:  make(map[string]bool, 5),
Details: make(map[string]string, 5),
}

stats := g.history.GetStats(operatorID)

// All checks pass trivially if the operator has no history yet.
if stats == nil {
for _, name := range []string{
"distribution", "label_flip", "source_diversity",
"temporal_clustering", "high_trust_ratio",
} {
result.Checks[name] = true
result.Details[name] = "no history — auto-pass"
}
result.Passed = true
return result
}

result.Checks["distribution"] = g.checkDistribution(stats, result)
result.Checks["label_flip"] = g.checkLabelFlip(stats, result)
result.Checks["source_diversity"] = g.checkSourceDiversity(stats, result)
result.Checks["temporal_clustering"] = g.checkTemporalClustering(stats, result)
result.Checks["high_trust_ratio"] = g.checkHighTrustRatio(stats, result)

allPassed := true
for _, passed := range result.Checks {
if !passed {
allPassed = false
break
}
}
result.Passed = allPassed

if !allPassed {
g.logger.Warn(
"anti-poisoning check failed",
zap.String("operator_id", operatorID),
zap.Any("checks", result.Checks),
)
}

return result
}

// checkDistribution applies a chi-squared goodness-of-fit test comparing
// the operator's feedback type distribution to a uniform distribution.
// Requires ≥ minForDistribution submissions; below threshold auto-passes.
func (g *AntiPoisonGuard) checkDistribution(stats *state.OperatorStats, result *PoisonCheckResult) bool {
const name = "distribution"
if stats.TotalFeedback < minForDistribution {
result.Details[name] = "auto-pass: insufficient submissions"
return true
}

// We test against a uniform distribution over the 5 known feedback types.
// Chi-squared statistic: Σ (O-E)² / E where E = total/numCategories.
numCategories := 5 // FEEDBACK_TYPE_CONFIRM_HOSTILE through CONFIRM_ANOMALY
expected := float64(stats.TotalFeedback) / float64(numCategories)

chiSq := 0.0
for _, count := range stats.FeedbackByType {
diff := float64(count) - expected
chiSq += (diff * diff) / expected
}
// Add zero-count categories to the statistic.
nonZeroCategories := len(stats.FeedbackByType)
for i := 0; i < numCategories-nonZeroCategories; i++ {
chiSq += expected // (0 - expected)² / expected = expected
}

// Degrees of freedom = numCategories - 1 = 4
// p-value approximation: p ≈ 1 - CDF(chi², df=4)
pValue := chiSquaredPValue(chiSq, 4)
passed := pValue > chiSquaredPValueThreshold

if passed {
result.Details[name] = "distribution uniform enough"
} else {
result.Details[name] = "distribution biased — possible type-stuffing"
}
return passed
}

// checkLabelFlip verifies that the operator's label flip rate is ≤ 20%.
func (g *AntiPoisonGuard) checkLabelFlip(stats *state.OperatorStats, result *PoisonCheckResult) bool {
const name = "label_flip"
if stats.TotalFeedback < minForLabelFlip {
result.Details[name] = "auto-pass: insufficient submissions"
return true
}
rate := stats.LabelFlipRate()
passed := rate <= labelFlipRateThreshold
if passed {
result.Details[name] = "label flip rate acceptable"
} else {
result.Details[name] = "excessive label flipping detected"
}
return passed
}

// checkSourceDiversity verifies the operator has submitted feedback on tracks
// from ≥ 3 distinct sensor sources.
func (g *AntiPoisonGuard) checkSourceDiversity(stats *state.OperatorStats, result *PoisonCheckResult) bool {
const name = "source_diversity"
if stats.TotalFeedback < minForSource {
result.Details[name] = "auto-pass: insufficient submissions"
return true
}
unique := stats.UniqueSensorSources()
passed := unique >= minSourceDiversity
if passed {
result.Details[name] = "sufficient source diversity"
} else {
result.Details[name] = "insufficient source diversity — possible targeted poisoning"
}
return passed
}

// checkTemporalClustering verifies no 5-minute sliding window contains
// more than 40% of the operator's total feedback (burst-submission attack).
func (g *AntiPoisonGuard) checkTemporalClustering(stats *state.OperatorStats, result *PoisonCheckResult) bool {
const name = "temporal_clustering"
total := stats.TotalFeedback
if total == 0 {
result.Details[name] = "auto-pass: no submissions"
return true
}

timestamps := stats.RecentFeedback
threshold := int(math.Ceil(float64(total) * burstRatioThreshold))

// Sliding window: for each timestamp t, count how many fall in [t, t+5min].
for i, t := range timestamps {
windowEnd := t.Add(burstWindowDuration)
count := 0
for _, other := range timestamps[i:] {
if !other.After(windowEnd) {
count++
}
}
if count > threshold {
result.Details[name] = "burst submission pattern detected"
return false
}
}

result.Details[name] = "no burst pattern detected"
return true
}

// checkHighTrustRatio verifies ≥ 60% of the operator's feedback was validated.
// Requires ≥ minForHighTrust submissions; below threshold auto-passes.
func (g *AntiPoisonGuard) checkHighTrustRatio(stats *state.OperatorStats, result *PoisonCheckResult) bool {
const name = "high_trust_ratio"
if stats.TotalFeedback < minForHighTrust {
result.Details[name] = "auto-pass: insufficient submissions"
return true
}
ratio := stats.ValidatedRatio()
passed := ratio >= validatedRatioThreshold
if passed {
result.Details[name] = "validated ratio acceptable"
} else {
result.Details[name] = "insufficient validated feedback ratio"
}
return passed
}

// chiSquaredPValue computes an approximation of the chi-squared CDF survival
// function (1 - CDF) for the given statistic and integer degrees of freedom.
// This is a regularised incomplete gamma function approximation sufficient
// for the threshold check used here (p > 0.05 with df=4).
//
// Uses the series approximation of the regularised lower incomplete gamma
// function P(a, x) where a = df/2, x = chiSq/2.
func chiSquaredPValue(chiSq float64, df int) float64 {
if chiSq <= 0 {
return 1.0
}
// p-value = 1 - P(df/2, chiSq/2) = Q(df/2, chiSq/2)
a := float64(df) / 2.0
x := chiSq / 2.0
return 1.0 - regularisedGammaP(a, x)
}

// regularisedGammaP computes P(a, x) via a series expansion.
// Accurate for a > 0, x >= 0.
func regularisedGammaP(a, x float64) float64 {
if x <= 0 {
return 0
}
if x >= a+1 {
// Use continued fraction for large x.
return 1.0 - regularisedGammaQ_cf(a, x)
}
// Series expansion.
sum := 1.0 / a
term := sum
for n := 1; n < 200; n++ {
term *= x / (a + float64(n))
sum += term
if math.Abs(term) < 1e-12*math.Abs(sum) {
break
}
}
return sum * math.Exp(-x+a*math.Log(x)-lgamma(a))
}

// regularisedGammaQ_cf computes Q(a, x) via Lentz's continued fraction.
func regularisedGammaQ_cf(a, x float64) float64 {
const fpMin = 1e-300
b := x + 1.0 - a
c := 1.0 / fpMin
d := 1.0 / b
h := d
for i := 1; i <= 200; i++ {
fi := float64(i)
an := -fi * (fi - a)
b += 2.0
d = an*d + b
if math.Abs(d) < fpMin {
d = fpMin
}
c = b + an/c
if math.Abs(c) < fpMin {
c = fpMin
}
d = 1.0 / d
del := d * c
h *= del
if math.Abs(del-1.0) < 1e-12 {
break
}
}
return math.Exp(-x+a*math.Log(x)-lgamma(a)) * h
}

// lgamma returns ln(Γ(x)).
func lgamma(x float64) float64 {
lg, _ := math.Lgamma(x)
return lg
}
