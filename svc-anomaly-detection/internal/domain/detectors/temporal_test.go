// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
)

func TestTemporalDetector_T12_LowPValue(t *testing.T) {
// T12: p=0.01 → Detected
d := NewTemporalDetector(0.05)
fv := &domain.FeatureVector{
TemporalPValue: 0.01,
TrackAgeMin:    minTrackAgeForTemporalMin + 60, // Over 24h
HourOfDay:      2,
DayOfWeek:      1,
}
result := d.Detect(fv)
if !result.Detected {
t.Error("T12: expected Detected=true for p=0.01")
}
if result.Confidence <= 0 {
t.Errorf("T12: expected confidence > 0, got %v", result.Confidence)
}
}

func TestTemporalDetector_T13_HighPValue(t *testing.T) {
// T13: p=0.10 → Not detected
d := NewTemporalDetector(0.05)
fv := &domain.FeatureVector{
TemporalPValue: 0.10,
TrackAgeMin:    minTrackAgeForTemporalMin + 60,
HourOfDay:      14,
DayOfWeek:      3,
}
result := d.Detect(fv)
if result.Detected {
t.Errorf("T13: expected Detected=false for p=0.10, got confidence=%v", result.Confidence)
}
}

func TestTemporalDetector_T14_YoungTrack(t *testing.T) {
// T14: Track < 24h → Not detected (insufficient history)
d := NewTemporalDetector(0.05)
fv := &domain.FeatureVector{
TemporalPValue: 0.001, // Very low p-value, but track too young.
TrackAgeMin:    60.0,  // Only 1 hour old.
HourOfDay:      3,
DayOfWeek:      1,
}
result := d.Detect(fv)
if result.Detected {
t.Error("T14: expected Detected=false for track < 24h old")
}
}

func TestTemporalDetector_ConfidenceCalculation(t *testing.T) {
d := NewTemporalDetector(0.05)
fv := &domain.FeatureVector{
TemporalPValue: 0.01,
TrackAgeMin:    minTrackAgeForTemporalMin + 60,
HourOfDay:      2,
}
result := d.Detect(fv)
expectedConf := 1.0 - 0.01
if result.Confidence < expectedConf-0.001 || result.Confidence > expectedConf+0.001 {
t.Errorf("Confidence = %v, want ~%v", result.Confidence, expectedConf)
}
}

func TestTemporalDetector_FeaturesPopulated(t *testing.T) {
d := NewTemporalDetector(0.05)
fv := &domain.FeatureVector{
TemporalPValue: 0.01,
TrackAgeMin:    minTrackAgeForTemporalMin + 60,
HourOfDay:      2,
DayOfWeek:      1,
}
result := d.Detect(fv)
if len(result.Features) == 0 {
t.Error("Expected features to be populated")
}
}
