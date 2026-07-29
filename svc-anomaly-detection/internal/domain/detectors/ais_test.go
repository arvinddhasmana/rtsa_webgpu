// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
)

func TestAISManipulationDetector_T07_LargeDelta(t *testing.T) {
// T07: 2.0 NM delta → Detected, high confidence
d := NewAISManipulationDetector(0.5)
fv := &domain.FeatureVector{
HasAISSource:       true,
AISPositionDeltaNM: 2.0,
AISReportedPosition: &domain.Position{Latitude: 44.65, Longitude: -63.57},
FusedPosition:       &domain.Position{Latitude: 44.66, Longitude: -63.59},
}
result := d.Detect(fv)
if !result.Detected {
t.Error("T07: expected Detected=true for 2.0 NM AIS delta")
}
if result.Confidence < 0.5 {
t.Errorf("T07: expected high confidence, got %v", result.Confidence)
}
}

func TestAISManipulationDetector_T08_SmallDelta(t *testing.T) {
// T08: 0.3 NM delta → Not detected (below 0.5 NM)
d := NewAISManipulationDetector(0.5)
fv := &domain.FeatureVector{
HasAISSource:       true,
AISPositionDeltaNM: 0.3,
AISReportedPosition: &domain.Position{Latitude: 44.65, Longitude: -63.57},
FusedPosition:       &domain.Position{Latitude: 44.651, Longitude: -63.571},
}
result := d.Detect(fv)
if result.Detected {
t.Errorf("T08: expected Detected=false for 0.3 NM delta, got confidence=%v", result.Confidence)
}
}

func TestAISManipulationDetector_T09_NoAISSource(t *testing.T) {
// T09: No AIS source → Not detected (skipped)
d := NewAISManipulationDetector(0.5)
fv := &domain.FeatureVector{
HasAISSource:       false,
AISPositionDeltaNM: 5.0, // Large delta, but no AIS source.
}
result := d.Detect(fv)
if result.Detected {
t.Error("T09: expected Detected=false when no AIS source")
}
}

func TestAISManipulationDetector_ConfidenceClamped(t *testing.T) {
d := NewAISManipulationDetector(0.5)
fv := &domain.FeatureVector{
HasAISSource:       true,
AISPositionDeltaNM: 100.0, // Extremely large.
}
result := d.Detect(fv)
if result.Confidence > 1.0 {
t.Errorf("Confidence exceeds 1.0: %v", result.Confidence)
}
}

func TestAISManipulationDetector_FeaturesPopulated(t *testing.T) {
d := NewAISManipulationDetector(0.5)
fv := &domain.FeatureVector{
HasAISSource:       true,
AISPositionDeltaNM: 2.0,
AISReportedPosition: &domain.Position{Latitude: 44.65, Longitude: -63.57},
FusedPosition:       &domain.Position{Latitude: 44.66, Longitude: -63.59},
}
result := d.Detect(fv)
// Should have at least 3 features (delta + ais_lat/lon + fused_lat/lon).
if len(result.Features) < 3 {
t.Errorf("Expected >= 3 features, got %d", len(result.Features))
}
}
