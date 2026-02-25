// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
)

// GenerateExplanation creates a human-readable explanation for an anomaly alert.
// Template-based with variable substitution.
// Parameters:
//   - anomalyType: the type of anomaly detected
//   - fv:          the feature vector that triggered detection
//   - confidence:  the detection confidence score (0.0–1.0)
func GenerateExplanation(
anomalyType commonv1.AnomalyType,
fv *FeatureVector,
confidence float64,
) string {
entityTypeStr := entityTypeString(fv.EntityType)

switch anomalyType {
case commonv1.AnomalyType_ANOMALY_TYPE_SPEED:
return generateSpeedExplanation(fv, entityTypeStr)

case commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION:
return generateRouteExplanation(fv, entityTypeStr)

case commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION:
return generateAISExplanation(fv, entityTypeStr)

case commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL:
return generateBehavioralExplanation(fv, entityTypeStr)

case commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL:
return generateTemporalExplanation(fv, entityTypeStr)

case commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY:
return generateProximityExplanation(fv, entityTypeStr)

default:
return fmt.Sprintf("Track %s (%s) triggered anomaly detection with confidence %.2f.",
fv.TrackID, entityTypeStr, confidence)
}
}

func generateSpeedExplanation(fv *FeatureVector, entityType string) string {
sigma := fv.SpeedDeltaSigma
if sigma == 0 && fv.SpeedStdDev > 0.1 {
sigma = abs64(fv.CurrentSpeedKnots-fv.AvgSpeed30Min) / fv.SpeedStdDev
}
return fmt.Sprintf(
"Track %s (%s) is moving at %.1f knots, which is %.1fσ above the 30-minute average of %.1f knots.",
fv.TrackID, entityType,
fv.CurrentSpeedKnots,
sigma,
fv.AvgSpeed30Min,
)
}

func generateRouteExplanation(fv *FeatureVector, entityType string) string {
deviation := abs64(fv.HeadingDeviation)
if deviation > 180 {
deviation = 360 - deviation
}
return fmt.Sprintf(
"Track %s (%s) has deviated %.0f° from its expected heading of %.0f° for the last %d updates.",
fv.TrackID, entityType,
deviation,
fv.ExpectedHeading,
3, // sustained updates — use default for explanation
)
}

func generateAISExplanation(fv *FeatureVector, entityType string) string {
return fmt.Sprintf(
"Track %s (%s) AIS position differs from fused position by %.1f NM, indicating possible AIS spoofing.",
fv.TrackID, entityType,
fv.AISPositionDeltaNM,
)
}

func generateBehavioralExplanation(fv *FeatureVector, entityType string) string {
patternType := classifyActivityPattern(fv.ActivityPattern)
switch patternType {
case "loitering":
return fmt.Sprintf(
"Track %s (%s) is exhibiting a loitering pattern, remaining within 0.1 NM for %.0f minutes.",
fv.TrackID, entityType, fv.TrackAgeMin,
)
case "zigzag":
return fmt.Sprintf(
"Track %s (%s) is exhibiting a zigzag pattern with repeated heading reversals over %.0f minutes.",
fv.TrackID, entityType, fv.TrackAgeMin,
)
case "speed_pulsing":
return fmt.Sprintf(
"Track %s (%s) is exhibiting speed pulsing behaviour (alternating fast/slow) over %.0f minutes.",
fv.TrackID, entityType, fv.TrackAgeMin,
)
default:
return fmt.Sprintf(
"Track %s (%s) is exhibiting anomalous behavioural patterns (confidence: %.2f).",
fv.TrackID, entityType, fv.PatternConfidence,
)
}
}

func generateTemporalExplanation(fv *FeatureVector, entityType string) string {
return fmt.Sprintf(
"Track %s (%s) showing activity at %02d:00 UTC, unusual for this area (p-value: %.2f).",
fv.TrackID, entityType,
fv.HourOfDay,
fv.TemporalPValue,
)
}

func generateProximityExplanation(fv *FeatureVector, entityType string) string {
if fv.InExclusionZone {
depth := fv.NearestZoneRadiusNM - fv.NearestExclusionZoneDistNM
return fmt.Sprintf(
"Track %s (%s) has entered exclusion zone '%s' (%.1f NM inside perimeter).",
fv.TrackID, entityType,
fv.NearestZoneName,
depth,
)
}
return fmt.Sprintf(
"Track %s (%s) is approaching exclusion zone '%s' (%.1f NM from perimeter).",
fv.TrackID, entityType,
fv.NearestZoneName,
fv.NearestExclusionZoneDistNM,
)
}

// classifyActivityPattern returns a human-readable pattern type from activity sequence.
func classifyActivityPattern(activity []float64) string {
if len(activity) < 3 {
return "unknown"
}
var mean float64
for _, v := range activity {
mean += v
}
mean /= float64(len(activity))

maxSpeed := 0.0
for _, v := range activity {
if v > maxSpeed {
maxSpeed = v
}
}
if maxSpeed < 2.0 {
return "loitering"
}

pulsing := 0
prevFast := activity[0] > mean
for _, v := range activity[1:] {
currFast := v > mean
if currFast != prevFast {
pulsing++
}
prevFast = currFast
}
if pulsing > 4 {
return "speed_pulsing"
}
return "zigzag"
}

// entityTypeString converts EntityType enum to a human-readable string.
func entityTypeString(et commonv1.EntityType) string {
switch et {
case commonv1.EntityType_ENTITY_TYPE_SURFACE:
return "SURFACE"
case commonv1.EntityType_ENTITY_TYPE_AIR:
return "AIR"
case commonv1.EntityType_ENTITY_TYPE_SUBSURFACE:
return "SUBSURFACE"
case commonv1.EntityType_ENTITY_TYPE_LAND:
return "LAND"
case commonv1.EntityType_ENTITY_TYPE_CYBER:
return "CYBER"
default:
return "UNKNOWN"
}
}

func abs64(v float64) float64 {
if v < 0 {
return -v
}
return v
}
