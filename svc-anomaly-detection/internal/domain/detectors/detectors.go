// CLASSIFICATION: UNCLASSIFIED
// Package detectors provides anomaly detector implementations for the
// RTSA anomaly detection service. Each detector is independent and
// operates on a FeatureVector extracted from a fused track.
package detectors

import (
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
)

// DetectionResult holds the output of an anomaly detector.
type DetectionResult struct {
Detected   bool
Confidence float64 // 0.0 to 1.0
Features   []*inferencev1.FeatureContribution
}

// Detector is the common interface for all anomaly detectors.
type Detector interface {
Detect(fv *domain.FeatureVector) *DetectionResult
}

// featureContrib is a helper to build a FeatureContribution proto.
func featureContrib(name string, value, weight float64, description string) *inferencev1.FeatureContribution {
return &inferencev1.FeatureContribution{
FeatureName:        name,
Value:              value,
ContributionWeight: weight,
Description:        description,
}
}

// featureContribWithBaseline is a helper that also sets a baseline value.
func featureContribWithBaseline(name string, value, weight, baseline float64, description string) *inferencev1.FeatureContribution {
b := baseline
return &inferencev1.FeatureContribution{
FeatureName:        name,
Value:              value,
ContributionWeight: weight,
BaselineValue:      &b,
Description:        description,
}
}

func clamp(v, min, max float64) float64 {
if v < min {
return min
}
if v > max {
return max
}
return v
}
