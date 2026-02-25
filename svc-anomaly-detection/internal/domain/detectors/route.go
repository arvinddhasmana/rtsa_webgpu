// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"math"

inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
)

// RouteDeviationDetector detects sustained heading changes > deviationThreshold
// degrees from the expected course.
type RouteDeviationDetector struct {
deviationThreshold float64 // default: 30.0 degrees
sustainedUpdates   int     // consecutive updates required (default: 3)
// consecutiveCount tracks how many consecutive updates have shown deviation.
consecutiveCount int
}

// NewRouteDeviationDetector creates a RouteDeviationDetector.
func NewRouteDeviationDetector(deviationDeg float64, sustainedN int) *RouteDeviationDetector {
if deviationDeg <= 0 {
deviationDeg = 30.0
}
if sustainedN <= 0 {
sustainedN = 3
}
return &RouteDeviationDetector{
deviationThreshold: deviationDeg,
sustainedUpdates:   sustainedN,
}
}

// Detect checks for sustained route deviation.
// Algorithm:
//  1. Calculate HeadingDeviation = |CurrentHeading - ExpectedHeading|
//  2. Handle wraparound: if deviation > 180, deviation = 360 - deviation
//  3. If deviation > deviationThreshold increment sustained counter; else reset
//  4. If sustained counter >= sustainedUpdates → Detected
//  5. Confidence = min(1.0, deviation / 90.0)
//  6. Features: heading_deviation, expected_heading, current_heading, heading_change_rate
func (d *RouteDeviationDetector) Detect(fv *domain.FeatureVector) *DetectionResult {
deviation := math.Abs(fv.HeadingDeviation)
// Ensure deviation is in [0, 180].
if deviation > 180 {
deviation = 360 - deviation
}

// Update sustained counter.
if deviation > d.deviationThreshold {
d.consecutiveCount++
} else {
d.consecutiveCount = 0
}

detected := d.consecutiveCount >= d.sustainedUpdates
confidence := 0.0
if deviation > d.deviationThreshold {
confidence = clamp(deviation/90.0, 0, 1.0)
}

return &DetectionResult{
Detected:   detected,
Confidence: confidence,
Features: []*inferencev1.FeatureContribution{
featureContrib("heading_deviation", deviation, 0.4,
"Absolute heading deviation from expected course (degrees)"),
featureContrib("expected_heading", fv.ExpectedHeading, 0.2,
"Expected heading based on recent track history (degrees)"),
featureContrib("current_heading", fv.CurrentHeading, 0.2,
"Current track heading (degrees)"),
featureContrib("heading_change_rate", fv.HeadingChangeRate, 0.2,
"Rate of heading change (degrees/minute)"),
},
}
}

// Reset resets the sustained counter (e.g., for testing).
func (d *RouteDeviationDetector) Reset() {
d.consecutiveCount = 0
}
