// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
)

func TestSpeedDetector_T01_HighSigma(t *testing.T) {
// T01: Speed = avg + 4σ → Detected, confidence > 0.5
d := NewSpeedDetector(3.0)
fv := &domain.FeatureVector{
CurrentSpeedKnots: 40.0,
AvgSpeed30Min:     10.0,   // delta = 30
SpeedStdDev:       7.5,    // sigma = 4.0
SpeedDeltaSigma:   4.0,
}
result := d.Detect(fv)
if !result.Detected {
t.Error("T01: expected Detected=true for 4σ speed anomaly")
}
if result.Confidence <= 0.5 {
t.Errorf("T01: expected confidence > 0.5, got %v", result.Confidence)
}
}

func TestSpeedDetector_T02_LowSigma(t *testing.T) {
// T02: Speed = avg + 1σ → Not detected
d := NewSpeedDetector(3.0)
fv := &domain.FeatureVector{
CurrentSpeedKnots: 17.5,
AvgSpeed30Min:     10.0, // delta = 7.5
SpeedStdDev:       7.5,  // sigma = 1.0
SpeedDeltaSigma:   1.0,
}
result := d.Detect(fv)
if result.Detected {
t.Errorf("T02: expected Detected=false for 1σ, got confidence=%v", result.Confidence)
}
}

func TestSpeedDetector_T03_InsufficientVariance(t *testing.T) {
// T03: StdDev < 0.1 → Not detected (insufficient history)
d := NewSpeedDetector(3.0)
fv := &domain.FeatureVector{
CurrentSpeedKnots: 100.0,
AvgSpeed30Min:     10.0,
SpeedStdDev:       0.05, // < 0.1
SpeedDeltaSigma:   0,
}
result := d.Detect(fv)
if result.Detected {
t.Error("T03: expected Detected=false for insufficient variance (stddev < 0.1)")
}
}

func TestSpeedDetector_ConfidenceClamped(t *testing.T) {
// Confidence should never exceed 1.0.
d := NewSpeedDetector(3.0)
fv := &domain.FeatureVector{
CurrentSpeedKnots: 1000.0,
AvgSpeed30Min:     10.0,
SpeedStdDev:       5.0, // sigma = 198
SpeedDeltaSigma:   198,
}
result := d.Detect(fv)
if result.Confidence > 1.0 {
t.Errorf("Confidence exceeds 1.0: %v", result.Confidence)
}
}

func TestSpeedDetector_FeaturesPopulated(t *testing.T) {
d := NewSpeedDetector(3.0)
fv := &domain.FeatureVector{
CurrentSpeedKnots: 40.0,
AvgSpeed30Min:     10.0,
SpeedStdDev:       7.5,
}
result := d.Detect(fv)
if len(result.Features) == 0 {
t.Error("Expected features to be populated")
}
}

func TestSpeedDetector_DefaultSigma(t *testing.T) {
// Zero sigma threshold should default to 3.0.
d := NewSpeedDetector(0)
if d.sigmaThreshold != 3.0 {
t.Errorf("Default sigma threshold = %v, want 3.0", d.sigmaThreshold)
}
}
