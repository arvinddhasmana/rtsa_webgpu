// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"math"

inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
)

// SpeedDetector detects speed anomalies where current speed exceeds
// sigmaThreshold standard deviations from the track's historical mean speed.
type SpeedDetector struct {
sigmaThreshold float64 // default: 3.0
}

// NewSpeedDetector creates a SpeedDetector with the given sigma threshold.
func NewSpeedDetector(sigmaThreshold float64) *SpeedDetector {
if sigmaThreshold <= 0 {
sigmaThreshold = 3.0
}
return &SpeedDetector{sigmaThreshold: sigmaThreshold}
}

// Detect checks for a speed anomaly.
// Algorithm:
//  1. If SpeedStdDev < 0.1 (insufficient variance), skip → not detected.
//  2. delta = |CurrentSpeed - AvgSpeed30Min|
//  3. sigma = delta / SpeedStdDev
//  4. If sigma > sigmaThreshold → Detected = true
//  5. Confidence = min(1.0, sigma / (sigmaThreshold * 2))
//  6. Features: speed_delta, speed_stddev, sigma_value
func (d *SpeedDetector) Detect(fv *domain.FeatureVector) *DetectionResult {
// Step 1: insufficient variance — skip detection.
if fv.SpeedStdDev < 0.1 {
return &DetectionResult{
Detected:   false,
Confidence: 0,
Features: []*inferencev1.FeatureContribution{
featureContrib("speed_stddev", fv.SpeedStdDev, 1.0,
"Speed standard deviation too low for detection"),
},
}
}

delta := math.Abs(fv.CurrentSpeedKnots - fv.AvgSpeed30Min)
sigma := delta / fv.SpeedStdDev
detected := sigma > d.sigmaThreshold
confidence := clamp(sigma/(d.sigmaThreshold*2), 0, 1.0)

return &DetectionResult{
Detected:   detected,
Confidence: confidence,
Features: []*inferencev1.FeatureContribution{
featureContribWithBaseline("speed_delta", delta, 0.5, 0,
"Absolute speed delta from 30-minute average (knots)"),
featureContrib("speed_stddev", fv.SpeedStdDev, 0.3,
"Speed standard deviation over 30-minute history"),
featureContrib("sigma_value", sigma, 0.2,
"Speed deviation in standard deviations (σ)"),
},
}
}
