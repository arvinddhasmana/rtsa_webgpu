// CLASSIFICATION: UNCLASSIFIED
package detectors

import (
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
)

// AISManipulationDetector detects discrepancies between AIS-reported
// position and fused (multi-sensor) position exceeding discrepancyThresholdNM.
type AISManipulationDetector struct {
discrepancyThresholdNM float64 // default: 0.5
}

// NewAISManipulationDetector creates an AISManipulationDetector.
func NewAISManipulationDetector(thresholdNM float64) *AISManipulationDetector {
if thresholdNM <= 0 {
thresholdNM = 0.5
}
return &AISManipulationDetector{discrepancyThresholdNM: thresholdNM}
}

// Detect checks for AIS position manipulation.
// Algorithm:
//  1. If !HasAISSource → skip (not detected)
//  2. If AISPositionDeltaNM > discrepancyThresholdNM → Detected
//  3. Confidence = min(1.0, AISPositionDeltaNM / (discrepancyThresholdNM * 3))
//  4. Features: ais_position_delta_nm, ais_lat, ais_lon, fused_lat, fused_lon
//
// Only applicable to SURFACE entity type.
func (d *AISManipulationDetector) Detect(fv *domain.FeatureVector) *DetectionResult {
// Skip if no AIS source.
if !fv.HasAISSource {
return &DetectionResult{Detected: false, Confidence: 0}
}

delta := fv.AISPositionDeltaNM
detected := delta > d.discrepancyThresholdNM
confidence := clamp(delta/(d.discrepancyThresholdNM*3), 0, 1.0)

features := []*inferencev1.FeatureContribution{
featureContrib("ais_position_delta_nm", delta, 0.6,
"Distance between AIS-reported and fused position (nautical miles)"),
}

if fv.AISReportedPosition != nil {
features = append(features,
featureContrib("ais_lat", fv.AISReportedPosition.Latitude, 0.1,
"AIS-reported latitude (degrees)"),
featureContrib("ais_lon", fv.AISReportedPosition.Longitude, 0.1,
"AIS-reported longitude (degrees)"),
)
}
if fv.FusedPosition != nil {
features = append(features,
featureContrib("fused_lat", fv.FusedPosition.Latitude, 0.1,
"Fused position latitude (degrees)"),
featureContrib("fused_lon", fv.FusedPosition.Longitude, 0.1,
"Fused position longitude (degrees)"),
)
}

return &DetectionResult{
Detected:   detected,
Confidence: confidence,
Features:   features,
}
}
