// CLASSIFICATION: UNCLASSIFIED
package generator

import (
"math"
"math/rand"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
)

// Exclusion zone constants (Halifax Harbour vicinity).
const (
ExclusionLat    = 45.5
ExclusionLon    = -60.0
ExclusionRadiusNM = 5.0
)

// AnomalyInjector modifies entity behaviour to simulate anomalies.
type AnomalyInjector struct {
rng *rand.Rand
}

// NewAnomalyInjector creates a new AnomalyInjector with the given RNG.
func NewAnomalyInjector(rng *rand.Rand) *AnomalyInjector {
return &AnomalyInjector{rng: rng}
}

// InjectSpeedAnomaly suddenly increases the entity speed to >3σ from its normal range.
//
//Surface normal: 12±5 knots → anomaly ≥35 knots
//Air normal:     300±100 knots → anomaly ≥700 knots
func (inj *AnomalyInjector) InjectSpeedAnomaly(e *SimEntity) {
switch e.EntityType {
case commonv1.EntityType_ENTITY_TYPE_SURFACE:
// 3σ from mean 12, σ=5 → 12+3*5=27, we use 35+ for clear anomaly
e.Position.SpeedKn = 35.0 + inj.rng.Float64()*15.0
case commonv1.EntityType_ENTITY_TYPE_AIR:
// 3σ from mean 300, σ=100 → 600+; use 700+
e.Position.SpeedKn = 700.0 + inj.rng.Float64()*100.0
case commonv1.EntityType_ENTITY_TYPE_SUBSURFACE:
// Normal 10±3 knots → anomaly 25+
e.Position.SpeedKn = 25.0 + inj.rng.Float64()*10.0
default:
e.Position.SpeedKn = 35.0 + inj.rng.Float64()*15.0
}
}

// InjectRouteDeviation applies a sudden heading change >30° sustained.
// Changes heading by 45-90° (randomly left or right).
func (inj *AnomalyInjector) InjectRouteDeviation(e *SimEntity) {
change := 45.0 + inj.rng.Float64()*45.0 // 45-90°
if inj.rng.Intn(2) == 0 {
change = -change
}
e.Position.Heading = normalizeHeading(e.Position.Heading + change)
}

// InjectAISManipulation returns an offset Position that should be reported by AIS.
// The true radar position is e.Position; the AIS position is offset by 0.5-2.0 NM.
// Both observations must be generated for the same entity to create the anomaly.
func (inj *AnomalyInjector) InjectAISManipulation(e *SimEntity) Position {
offsetNM := 0.5 + inj.rng.Float64()*1.5 // 0.5-2.0 NM
directionDeg := inj.rng.Float64() * 360
directionRad := directionDeg * degreesToRadians

cosLat := math.Cos(e.Position.Lat * degreesToRadians)
deltaLat := offsetNM * math.Cos(directionRad) * nmToDegreesLat
deltaLon := 0.0
if cosLat != 0 {
deltaLon = offsetNM * math.Sin(directionRad) * nmToDegreesLat / cosLat
}

manipulated := Position{
Lat:     e.Position.Lat + deltaLat,
Lon:     e.Position.Lon + deltaLon,
SpeedKn: e.Position.SpeedKn,
Heading: e.Position.Heading,
AltM:    0,
}
clampToDefaultArea(&manipulated)
return manipulated
}

// InjectBehavioral changes the movement pattern to simulate unexpected behaviour.
// A straight-line entity starts loitering; a loitering entity starts zigzagging.
func (inj *AnomalyInjector) InjectBehavioral(e *SimEntity) {
switch e.MovementPattern {
case PatternStraightLine:
e.MovementPattern = PatternLoitering
e.LoiterCenter = e.Position
e.LoiterRadiusNM = 0.5 + inj.rng.Float64()*0.5
e.LoiterAngle = inj.rng.Float64() * 2 * math.Pi
case PatternLoitering:
e.MovementPattern = PatternZigzag
case PatternZigzag:
e.MovementPattern = PatternPatrol
default:
e.MovementPattern = PatternZigzag
}
}

// InjectProximity moves the entity toward the exclusion zone centre.
// Zone: 45.5°N, -60.0°W, radius 5 NM.
func (inj *AnomalyInjector) InjectProximity(e *SimEntity) {
zone := Position{Lat: ExclusionLat, Lon: ExclusionLon}
// Move entity 1-3 NM closer to the zone centre.
bearing := bearingTo(e.Position, zone)
e.Position.Heading = bearing
offsetNM := 1.0 + inj.rng.Float64()*2.0
directionRad := bearing * degreesToRadians
cosLat := math.Cos(e.Position.Lat * degreesToRadians)
e.Position.Lat += offsetNM * math.Cos(directionRad) * nmToDegreesLat
if cosLat != 0 {
e.Position.Lon += offsetNM * math.Sin(directionRad) * nmToDegreesLat / cosLat
}
clampToDefaultArea(&e.Position)
}

// OffsetNM calculates the approximate offset in NM between two positions.
func OffsetNM(a, b Position) float64 {
return distanceNM(a, b)
}
