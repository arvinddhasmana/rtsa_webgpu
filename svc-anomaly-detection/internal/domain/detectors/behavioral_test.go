// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
)

func TestBehavioralDetector_T10_Loitering(t *testing.T) {
// T10: Loitering for 45 min → Detected
d := NewBehavioralDetector(0.75)
fv := &domain.FeatureVector{
PatternConfidence: 0.85,
ActivityPattern:   makeSlowPattern(20),
TrackAgeMin:       45.0,
}
result := d.Detect(fv)
if !result.Detected {
t.Errorf("T10: expected Detected=true for loitering, confidence=%v", fv.PatternConfidence)
}
}

func TestBehavioralDetector_T11_NormalTransit(t *testing.T) {
// T11: Normal transit → Not detected
d := NewBehavioralDetector(0.75)
fv := &domain.FeatureVector{
PatternConfidence: 0.20,
ActivityPattern:   makeNormalPattern(20),
TrackAgeMin:       60.0,
}
result := d.Detect(fv)
if result.Detected {
t.Errorf("T11: expected Detected=false for normal transit, got confidence=%v", result.Confidence)
}
}

func TestBehavioralDetector_AtThreshold(t *testing.T) {
// Exactly at threshold should NOT be detected (must be strictly greater).
d := NewBehavioralDetector(0.75)
fv := &domain.FeatureVector{
PatternConfidence: 0.75,
ActivityPattern:   makeSlowPattern(20),
}
result := d.Detect(fv)
if result.Detected {
t.Error("Confidence exactly at threshold should not be detected")
}
}

func TestBehavioralDetector_FeaturesPopulated(t *testing.T) {
d := NewBehavioralDetector(0.75)
fv := &domain.FeatureVector{
PatternConfidence: 0.85,
ActivityPattern:   makeSlowPattern(20),
TrackAgeMin:       45.0,
}
result := d.Detect(fv)
if len(result.Features) == 0 {
t.Error("Expected features to be populated")
}
}

// makeSlowPattern creates an activity pattern of near-zero speeds (loitering).
func makeSlowPattern(n int) []float64 {
p := make([]float64, n)
for i := range p {
p[i] = 0.3
}
return p
}

// makeNormalPattern creates an activity pattern of normal transit speeds.
func makeNormalPattern(n int) []float64 {
p := make([]float64, n)
for i := range p {
p[i] = 12.0 + float64(i%3)*0.5
}
return p
}
