// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"math"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
)

func TestProximityDetector_T15_InsideZone(t *testing.T) {
// T15: Inside exclusion zone → Detected, confidence=1.0
d := NewProximityDetector()
fv := &domain.FeatureVector{
InExclusionZone:            true,
NearestExclusionZoneDistNM: 0.0,
NearestZoneName:            "Halifax Naval Base",
NearestZoneRadiusNM:        2.0,
}
result := d.Detect(fv)
if !result.Detected {
t.Error("T15: expected Detected=true when inside exclusion zone")
}
if result.Confidence != 1.0 {
t.Errorf("T15: expected confidence=1.0 when inside zone, got %v", result.Confidence)
}
}

func TestProximityDetector_T16_FarFromZone(t *testing.T) {
// T16: 10 NM from zone (radius 2 NM) → Not detected
d := NewProximityDetector()
fv := &domain.FeatureVector{
InExclusionZone:            false,
NearestExclusionZoneDistNM: 10.0,
NearestZoneName:            "Halifax Naval Base",
NearestZoneRadiusNM:        2.0,
}
result := d.Detect(fv)
if result.Detected {
t.Errorf("T16: expected Detected=false when 10 NM from zone, got confidence=%v", result.Confidence)
}
}

func TestProximityDetector_Approaching(t *testing.T) {
// Within 1.5 * radius → Detected (approaching)
d := NewProximityDetector()
fv := &domain.FeatureVector{
InExclusionZone:            false,
NearestExclusionZoneDistNM: 2.5, // < 3.0 (2.0 * 1.5)
NearestZoneName:            "Zone-A",
NearestZoneRadiusNM:        2.0,
}
result := d.Detect(fv)
if !result.Detected {
t.Error("Approaching zone within 1.5x radius should be detected")
}
if result.Confidence <= 0 || result.Confidence >= 1.0 {
t.Errorf("Approaching confidence = %v, want (0, 1)", result.Confidence)
}
}

func TestProximityDetector_NoZones(t *testing.T) {
// MaxFloat64 distance = no zones configured.
d := NewProximityDetector()
fv := &domain.FeatureVector{
InExclusionZone:            false,
NearestExclusionZoneDistNM: math.MaxFloat64,
}
result := d.Detect(fv)
if result.Detected {
t.Error("No zones configured should not trigger detection")
}
}

func TestProximityDetector_FeaturesPopulated(t *testing.T) {
d := NewProximityDetector()
fv := &domain.FeatureVector{
InExclusionZone:            true,
NearestExclusionZoneDistNM: 0.0,
NearestZoneName:            "Zone-B",
NearestZoneRadiusNM:        1.5,
}
result := d.Detect(fv)
if len(result.Features) == 0 {
t.Error("Expected features to be populated")
}
}
