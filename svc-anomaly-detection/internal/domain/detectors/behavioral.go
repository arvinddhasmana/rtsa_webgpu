// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
)

// BehavioralDetector detects anomalous behavioral patterns based on
// activity sequence analysis. Uses statistical deviation scoring for MVP.
type BehavioralDetector struct {
confidenceThreshold float64 // default: 0.75
}

// NewBehavioralDetector creates a BehavioralDetector.
func NewBehavioralDetector(confidenceThreshold float64) *BehavioralDetector {
if confidenceThreshold <= 0 {
confidenceThreshold = 0.75
}
return &BehavioralDetector{confidenceThreshold: confidenceThreshold}
}

// Detect checks for behavioral anomalies.
// Algorithm (rule-based MVP):
//  1. Analyse ActivityPattern sequence
//  2. Look for suspicious patterns: loitering, zigzag, speed pulsing
//  3. If PatternConfidence > confidenceThreshold → Detected
//  4. Confidence = PatternConfidence
//  5. Features: pattern_type, pattern_confidence, pattern_duration_min
func (d *BehavioralDetector) Detect(fv *domain.FeatureVector) *DetectionResult {
confidence := fv.PatternConfidence
detected := confidence > d.confidenceThreshold

patternType := "none"
if detected {
patternType = classifyPattern(fv.ActivityPattern)
}

return &DetectionResult{
Detected:   detected,
Confidence: confidence,
Features: []*inferencev1.FeatureContribution{
featureContrib("pattern_type", encodePatternType(patternType), 0.4,
"Detected behavioral pattern type: "+patternType),
featureContrib("pattern_confidence", confidence, 0.5,
"Confidence score for the detected behavioral pattern (0–1)"),
featureContrib("pattern_duration_min", fv.TrackAgeMin, 0.1,
"Duration of tracked behavior (minutes)"),
},
}
}

// classifyPattern returns a human-readable pattern type from activity sequence.
func classifyPattern(activity []float64) string {
if len(activity) < 3 {
return "unknown"
}
// Detect speed pulsing: many alternations.
var mean float64
for _, v := range activity {
mean += v
}
mean /= float64(len(activity))

pulsing := 0
prevFast := activity[0] > mean
for _, v := range activity[1:] {
currFast := v > mean
if currFast != prevFast {
pulsing++
}
prevFast = currFast
}

// Detect near-stationary (loitering).
maxSpeed := 0.0
for _, v := range activity {
if v > maxSpeed {
maxSpeed = v
}
}
if maxSpeed < 2.0 {
return "loitering"
}
if pulsing > 4 {
return "speed_pulsing"
}
return "zigzag"
}

// encodePatternType converts pattern name to a numeric code for feature vector.
func encodePatternType(pt string) float64 {
switch pt {
case "loitering":
return 1.0
case "zigzag":
return 2.0
case "speed_pulsing":
return 3.0
default:
return 0.0
}
}
