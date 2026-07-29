// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
)

func TestRouteDeviationDetector_T04_SustainedDeviation(t *testing.T) {
// T04: 45° deviation sustained for 3 updates → Detected
d := NewRouteDeviationDetector(30.0, 3)
fv := &domain.FeatureVector{
CurrentHeading:   135.0,
ExpectedHeading:  90.0,
HeadingDeviation: 45.0,
}
// First two calls: not yet sustained.
r1 := d.Detect(fv)
if r1.Detected {
t.Error("T04: should not detect on first update")
}
r2 := d.Detect(fv)
if r2.Detected {
t.Error("T04: should not detect on second update")
}
// Third call: sustained threshold reached.
r3 := d.Detect(fv)
if !r3.Detected {
t.Error("T04: expected Detected=true after 3 sustained updates")
}
}

func TestRouteDeviationDetector_T05_SmallDeviation(t *testing.T) {
// T05: 15° deviation → not detected (below 30°)
d := NewRouteDeviationDetector(30.0, 3)
fv := &domain.FeatureVector{
CurrentHeading:   105.0,
ExpectedHeading:  90.0,
HeadingDeviation: 15.0,
}
for i := 0; i < 5; i++ {
r := d.Detect(fv)
if r.Detected {
t.Errorf("T05: should not detect 15° deviation on update %d", i+1)
}
}
}

func TestRouteDeviationDetector_T06_NotSustained(t *testing.T) {
// T06: 45° deviation for 1 update only → Not detected.
d := NewRouteDeviationDetector(30.0, 3)
devFV := &domain.FeatureVector{
CurrentHeading:   135.0,
ExpectedHeading:  90.0,
HeadingDeviation: 45.0,
}
normalFV := &domain.FeatureVector{
CurrentHeading:   92.0,
ExpectedHeading:  90.0,
HeadingDeviation: 2.0,
}
// One deviation update, then reset.
d.Detect(devFV)
d.Detect(normalFV) // Resets counter.
r := d.Detect(devFV)
if r.Detected {
t.Error("T06: should not detect non-sustained deviation")
}
}

func TestRouteDeviationDetector_Wraparound(t *testing.T) {
d := NewRouteDeviationDetector(30.0, 1)
// Heading 350°, expected 10° — only 20° difference after wraparound.
fv := &domain.FeatureVector{
CurrentHeading:   350.0,
ExpectedHeading:  10.0,
HeadingDeviation: -20.0, // angular diff = -20, abs = 20°
}
r := d.Detect(fv)
if r.Detected {
t.Error("Wraparound: 20° deviation should not trigger 30° threshold")
}
}

func TestRouteDeviationDetector_Confidence(t *testing.T) {
d := NewRouteDeviationDetector(30.0, 1)
fv := &domain.FeatureVector{
CurrentHeading:   180.0,
ExpectedHeading:  90.0,
HeadingDeviation: 90.0,
}
r := d.Detect(fv)
if !r.Detected {
t.Error("90° deviation should be detected")
}
if r.Confidence != 1.0 {
t.Errorf("90° deviation confidence = %v, want 1.0", r.Confidence)
}
}
