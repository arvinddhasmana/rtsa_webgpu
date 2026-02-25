// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"strings"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
)

func TestGenerateExplanation_T21_Speed(t *testing.T) {
// T21: Speed anomaly explanation contains speed values and sigma.
fv := &FeatureVector{
TrackID:           "T-123",
EntityType:        commonv1.EntityType_ENTITY_TYPE_SURFACE,
CurrentSpeedKnots: 45.2,
AvgSpeed30Min:     12.1,
SpeedStdDev:       7.9,
SpeedDeltaSigma:   4.2,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_SPEED, fv, 0.85)

if !strings.Contains(explanation, "45.2") {
t.Errorf("Speed explanation missing speed value: %s", explanation)
}
if !strings.Contains(explanation, "σ") {
t.Errorf("Speed explanation missing sigma symbol: %s", explanation)
}
if !strings.Contains(explanation, "12.1") {
t.Errorf("Speed explanation missing average speed: %s", explanation)
}
}

func TestGenerateExplanation_T22_AIS(t *testing.T) {
// T22: AIS manipulation explanation contains distance and "spoofing".
fv := &FeatureVector{
TrackID:            "T-789",
EntityType:         commonv1.EntityType_ENTITY_TYPE_SURFACE,
AISPositionDeltaNM: 2.3,
HasAISSource:       true,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION, fv, 0.92)

if !strings.Contains(explanation, "2.3") {
t.Errorf("AIS explanation missing distance: %s", explanation)
}
if !strings.Contains(strings.ToLower(explanation), "spoofing") {
t.Errorf("AIS explanation missing 'spoofing': %s", explanation)
}
}

func TestGenerateExplanation_Route(t *testing.T) {
fv := &FeatureVector{
TrackID:          "T-456",
EntityType:       commonv1.EntityType_ENTITY_TYPE_AIR,
HeadingDeviation: 47.0,
ExpectedHeading:  90.0,
CurrentHeading:   137.0,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION, fv, 0.75)

if !strings.Contains(explanation, "47") {
t.Errorf("Route explanation missing deviation: %s", explanation)
}
if !strings.Contains(explanation, "90") {
t.Errorf("Route explanation missing expected heading: %s", explanation)
}
}

func TestGenerateExplanation_Behavioral_Loitering(t *testing.T) {
fv := &FeatureVector{
TrackID:           "T-012",
EntityType:        commonv1.EntityType_ENTITY_TYPE_SURFACE,
PatternConfidence: 0.85,
TrackAgeMin:       45.0,
ActivityPattern:   make([]float64, 20), // All zeros = loitering
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL, fv, 0.85)

if !strings.Contains(strings.ToLower(explanation), "loitering") {
t.Errorf("Behavioral explanation missing 'loitering': %s", explanation)
}
}

func TestGenerateExplanation_Temporal(t *testing.T) {
fv := &FeatureVector{
TrackID:        "T-345",
EntityType:     commonv1.EntityType_ENTITY_TYPE_SURFACE,
HourOfDay:      3,
TemporalPValue: 0.02,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL, fv, 0.98)

if !strings.Contains(explanation, "03") {
t.Errorf("Temporal explanation missing hour: %s", explanation)
}
if !strings.Contains(explanation, "0.02") {
t.Errorf("Temporal explanation missing p-value: %s", explanation)
}
}

func TestGenerateExplanation_Proximity_InZone(t *testing.T) {
fv := &FeatureVector{
TrackID:                    "T-678",
EntityType:                 commonv1.EntityType_ENTITY_TYPE_SUBSURFACE,
InExclusionZone:            true,
NearestZoneName:            "Halifax Naval Base",
NearestZoneRadiusNM:        2.0,
NearestExclusionZoneDistNM: 0.0,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY, fv, 1.0)

if !strings.Contains(explanation, "Halifax Naval Base") {
t.Errorf("Proximity explanation missing zone name: %s", explanation)
}
if !strings.Contains(strings.ToLower(explanation), "entered") {
t.Errorf("Proximity explanation missing 'entered': %s", explanation)
}
}

func TestGenerateExplanation_Proximity_Approaching(t *testing.T) {
fv := &FeatureVector{
TrackID:                    "T-679",
EntityType:                 commonv1.EntityType_ENTITY_TYPE_SURFACE,
InExclusionZone:            false,
NearestZoneName:            "Zone-Alpha",
NearestZoneRadiusNM:        1.5,
NearestExclusionZoneDistNM: 1.0,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY, fv, 0.75)

if !strings.Contains(strings.ToLower(explanation), "approaching") {
t.Errorf("Proximity explanation for approach missing 'approaching': %s", explanation)
}
}

func TestGenerateExplanation_Unknown(t *testing.T) {
fv := &FeatureVector{
TrackID:    "T-999",
EntityType: commonv1.EntityType_ENTITY_TYPE_LAND,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED, fv, 0.75)
if explanation == "" {
t.Error("Explanation should not be empty for unknown type")
}
}

func TestGenerateExplanation_Behavioral_SpeedPulsing(t *testing.T) {
// Create a speed pulsing pattern (alternating high/low > 4 times)
pattern := make([]float64, 20)
for i := range pattern {
if i%2 == 0 {
pattern[i] = 20.0 // fast
} else {
pattern[i] = 5.0 // slow
}
}
fv := &FeatureVector{
TrackID:           "T-SP",
EntityType:        commonv1.EntityType_ENTITY_TYPE_SURFACE,
PatternConfidence: 0.85,
TrackAgeMin:       30.0,
ActivityPattern:   pattern,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL, fv, 0.85)
if explanation == "" {
t.Error("Expected non-empty explanation for speed pulsing")
}
}

func TestGenerateExplanation_Behavioral_Zigzag(t *testing.T) {
// Zigzag: speeds > 2 but less pulsing → zigzag
pattern := make([]float64, 20)
for i := range pattern {
pattern[i] = 15.0 // constant speed (no pulsing) → should be zigzag
}
fv := &FeatureVector{
TrackID:           "T-ZZ",
EntityType:        commonv1.EntityType_ENTITY_TYPE_AIR,
PatternConfidence: 0.85,
TrackAgeMin:       25.0,
ActivityPattern:   pattern,
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL, fv, 0.85)
if explanation == "" {
t.Error("Expected non-empty explanation for zigzag")
}
}

func TestGenerateExplanation_Behavioral_LowConfidence(t *testing.T) {
// Low confidence → default explanation
fv := &FeatureVector{
TrackID:           "T-LC",
EntityType:        commonv1.EntityType_ENTITY_TYPE_SUBSURFACE,
PatternConfidence: 0.60,
TrackAgeMin:       10.0,
ActivityPattern:   nil, // < 3 entries → unknown
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL, fv, 0.60)
if explanation == "" {
t.Error("Expected non-empty explanation for low confidence behavioral")
}
}

func TestAbs64(t *testing.T) {
if abs64(-5.5) != 5.5 {
t.Error("abs64(-5.5) != 5.5")
}
if abs64(3.3) != 3.3 {
t.Error("abs64(3.3) != 3.3")
}
}

func TestEntityTypeString_AllTypes(t *testing.T) {
tests := []struct {
et   commonv1.EntityType
want string
}{
{commonv1.EntityType_ENTITY_TYPE_SURFACE, "SURFACE"},
{commonv1.EntityType_ENTITY_TYPE_AIR, "AIR"},
{commonv1.EntityType_ENTITY_TYPE_SUBSURFACE, "SUBSURFACE"},
{commonv1.EntityType_ENTITY_TYPE_LAND, "LAND"},
{commonv1.EntityType_ENTITY_TYPE_CYBER, "CYBER"},
{commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED, "UNKNOWN"},
}
for _, tt := range tests {
got := entityTypeString(tt.et)
if got != tt.want {
t.Errorf("entityTypeString(%v) = %q, want %q", tt.et, got, tt.want)
}
}
}

func TestGenerateSpeedExplanation_ZeroDelta(t *testing.T) {
// SpeedDeltaSigma=0 → compute from scratch
fv := &FeatureVector{
TrackID:           "T-ZD",
EntityType:        commonv1.EntityType_ENTITY_TYPE_SURFACE,
CurrentSpeedKnots: 20.0,
AvgSpeed30Min:     10.0,
SpeedStdDev:       5.0,
SpeedDeltaSigma:   0, // triggers recompute
}
explanation := GenerateExplanation(commonv1.AnomalyType_ANOMALY_TYPE_SPEED, fv, 0.75)
if !strings.Contains(explanation, "20.0") {
t.Errorf("Speed explanation missing current speed: %s", explanation)
}
}
