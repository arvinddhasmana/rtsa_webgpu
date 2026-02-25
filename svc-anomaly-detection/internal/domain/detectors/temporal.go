// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
)

// minTrackAgeForTemporalMin is the minimum track age (in minutes) required
// before temporal detection is applied (24 hours).
const minTrackAgeForTemporalMin = 24 * 60.0

// TemporalDetector detects activity at unusual times for the entity's
// historical pattern (e.g., vessel movement at 3 AM in a normally idle port).
type TemporalDetector struct {
pValueThreshold float64 // default: 0.05
}

// NewTemporalDetector creates a TemporalDetector.
func NewTemporalDetector(pValueThreshold float64) *TemporalDetector {
if pValueThreshold <= 0 {
pValueThreshold = 0.05
}
return &TemporalDetector{pValueThreshold: pValueThreshold}
}

// Detect checks for temporal anomalies.
// Algorithm:
//  1. If track age < 24 hours → skip (insufficient history)
//  2. TemporalPValue from feature extractor (computed via chi-squared test)
//  3. If TemporalPValue < pValueThreshold → Detected
//  4. Confidence = 1.0 - TemporalPValue
//  5. Features: hour_of_day, day_of_week, temporal_p_value
func (d *TemporalDetector) Detect(fv *domain.FeatureVector) *DetectionResult {
// Step 1: insufficient history — skip detection.
if fv.TrackAgeMin < minTrackAgeForTemporalMin {
return &DetectionResult{
Detected:   false,
Confidence: 0,
Features: []*inferencev1.FeatureContribution{
featureContrib("track_age_min", fv.TrackAgeMin, 1.0,
"Track age in minutes (< 24h — insufficient for temporal detection)"),
},
}
}

detected := fv.TemporalPValue < d.pValueThreshold
confidence := 0.0
if detected {
confidence = clamp(1.0-fv.TemporalPValue, 0, 1.0)
}

return &DetectionResult{
Detected:   detected,
Confidence: confidence,
Features: []*inferencev1.FeatureContribution{
featureContrib("hour_of_day", float64(fv.HourOfDay), 0.4,
"Current hour of day (UTC, 0–23)"),
featureContrib("day_of_week", float64(fv.DayOfWeek), 0.2,
"Current day of week (0=Sunday, 6=Saturday)"),
featureContrib("temporal_p_value", fv.TemporalPValue, 0.4,
"P-value for activity at this time of day (lower = more anomalous)"),
},
}
}
