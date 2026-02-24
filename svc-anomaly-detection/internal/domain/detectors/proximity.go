// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
"math"

inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
)

// ProximityDetector detects entities within or approaching exclusion zones.
type ProximityDetector struct{}

// NewProximityDetector creates a ProximityDetector.
func NewProximityDetector() *ProximityDetector {
return &ProximityDetector{}
}

// Detect checks for proximity to exclusion zones.
// Algorithm:
//  1. If InExclusionZone → Detected, Confidence = 1.0
//  2. If NearestExclusionZoneDistNM < zone radius * 1.5 → Detected (approaching)
//  3. Confidence = 1.0 - (distance / (zone_radius * 2))
//  4. Features: nearest_zone_name, nearest_zone_distance_nm, in_exclusion_zone
func (d *ProximityDetector) Detect(fv *domain.FeatureVector) *DetectionResult {
// No exclusion zones configured.
if fv.NearestExclusionZoneDistNM == math.MaxFloat64 {
return &DetectionResult{Detected: false, Confidence: 0}
}

inZone := fv.InExclusionZone
distNM := fv.NearestExclusionZoneDistNM
radiusNM := fv.NearestZoneRadiusNM

if radiusNM <= 0 {
radiusNM = 1.0
}

approaching := distNM < radiusNM*1.5

detected := inZone || approaching
confidence := 0.0
if inZone {
confidence = 1.0
} else if approaching {
confidence = clamp(1.0-(distNM/(radiusNM*2)), 0, 1.0)
}

inZoneVal := 0.0
if inZone {
inZoneVal = 1.0
}

return &DetectionResult{
Detected:   detected,
Confidence: confidence,
Features: []*inferencev1.FeatureContribution{
featureContrib("nearest_zone_distance_nm", distNM, 0.6,
"Distance to nearest exclusion zone boundary (nautical miles)"),
featureContrib("in_exclusion_zone", inZoneVal, 0.3,
"Whether the entity is currently inside the exclusion zone (1=yes)"),
featureContrib("nearest_zone_radius_nm", radiusNM, 0.1,
"Radius of the nearest exclusion zone (nautical miles)"),
},
}
}
