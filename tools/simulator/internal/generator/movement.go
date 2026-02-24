// CLASSIFICATION: UNCLASSIFIED
package generator

import (
"math"
"math/rand"
"time"
)

const (
// knotsToNMPerHour converts knots to nautical miles per hour (they're the same).
knotsToNMPerHour = 1.0
// nmToDegreesLat converts nautical miles to degrees of latitude (1 NM = 1/60 degree).
nmToDegreesLat = 1.0 / 60.0
// degreesToRadians converts degrees to radians.
degreesToRadians = math.Pi / 180.0
)

// AdvancePosition moves a position based on its current speed and heading.
// Uses simple spherical geometry:
//
//lat_new = lat + (speed_nm_per_hour * dt_hours) * cos(heading) / 60
//lon_new = lon + (speed_nm_per_hour * dt_hours) * sin(heading) / (60 * cos(lat))
//
// Distances are in nautical miles, heading in degrees (0 = N, 90 = E).
func AdvancePosition(pos *Position, dt time.Duration) {
dtHours := dt.Hours()
distNM := pos.SpeedKn * knotsToNMPerHour * dtHours
headingRad := pos.Heading * degreesToRadians
latRad := pos.Lat * degreesToRadians

deltaLat := distNM * math.Cos(headingRad) * nmToDegreesLat
cosLat := math.Cos(latRad)
deltaLon := 0.0
if cosLat != 0 {
deltaLon = distNM * math.Sin(headingRad) * nmToDegreesLat / cosLat
}

pos.Lat += deltaLat
pos.Lon += deltaLon
}

// normalizeHeading clamps heading to [0, 360).
func normalizeHeading(h float64) float64 {
for h < 0 {
h += 360
}
for h >= 360 {
h -= 360
}
return h
}

// MoveStraightLine advances an entity at constant heading/speed with tiny perturbations.
// Perturbation: heading ±2°, speed ±0.5 knots per tick.
func MoveStraightLine(e *SimEntity, dt time.Duration, rng *rand.Rand) {
// Small random perturbation.
e.Position.Heading = normalizeHeading(e.Position.Heading + (rng.Float64()*4 - 2))
e.Position.SpeedKn += rng.Float64()*1.0 - 0.5
if e.Position.SpeedKn < 0.5 {
e.Position.SpeedKn = 0.5
}
AdvancePosition(&e.Position, dt)
}

// MovePatrol follows a list of waypoints, reversing direction at ends.
func MovePatrol(e *SimEntity, dt time.Duration, _ *rand.Rand, waypoints []Position) {
if len(waypoints) == 0 {
AdvancePosition(&e.Position, dt)
return
}

target := waypoints[e.WaypointIdx]
// Compute bearing to next waypoint.
e.Position.Heading = bearingTo(e.Position, target)

// Check if we've reached the waypoint (within ~0.1 NM tolerance).
dist := distanceNM(e.Position, target)
if dist < 0.1 {
// Advance to next waypoint.
if e.WaypointFwd {
e.WaypointIdx++
if e.WaypointIdx >= len(waypoints) {
e.WaypointIdx = len(waypoints) - 2
if e.WaypointIdx < 0 {
e.WaypointIdx = 0
}
e.WaypointFwd = false
}
} else {
e.WaypointIdx--
if e.WaypointIdx < 0 {
e.WaypointIdx = 1
if e.WaypointIdx >= len(waypoints) {
e.WaypointIdx = 0
}
e.WaypointFwd = true
}
}
}

AdvancePosition(&e.Position, dt)
}

// MoveLoitering moves an entity in a circle around center.
// The angular velocity is computed from the speed and radius.
func MoveLoitering(e *SimEntity, dt time.Duration, _ *rand.Rand, center Position) {
if e.LoiterRadiusNM <= 0 {
e.LoiterRadiusNM = 1.0
}
dtHours := dt.Hours()
distNM := e.Position.SpeedKn * dtHours
// Angular change = distance / radius (in radians).
deltaAngle := distNM / e.LoiterRadiusNM
e.LoiterAngle += deltaAngle

// Position on circle: lat/lon offsets in degrees.
// 1 NM = 1/60 degree lat. For lon: 1 NM = 1/(60 * cos(lat)) degrees.
latOffset := e.LoiterRadiusNM * math.Sin(e.LoiterAngle) * nmToDegreesLat
cosLat := math.Cos(center.Lat * degreesToRadians)
lonOffset := 0.0
if cosLat != 0 {
lonOffset = e.LoiterRadiusNM * math.Cos(e.LoiterAngle) * nmToDegreesLat / cosLat
}

e.Position.Lat = center.Lat + latOffset
e.Position.Lon = center.Lon + lonOffset
// Heading is tangent to the circle.
e.Position.Heading = normalizeHeading(e.LoiterAngle*180/math.Pi + 90)
}

// MoveZigzag alternates heading changes at regular intervals.
// Heading changes by ±30-45° every ~30 ticks.
func MoveZigzag(e *SimEntity, dt time.Duration, rng *rand.Rand) {
e.ZigzagTick++
period := 30 // ticks between direction changes
if e.ZigzagTick%period == 0 {
change := 30.0 + rng.Float64()*15.0 // 30-45°
if e.ZigzagPhase == 0 {
e.Position.Heading = normalizeHeading(e.Position.Heading + change)
e.ZigzagPhase = 1
} else {
e.Position.Heading = normalizeHeading(e.Position.Heading - change)
e.ZigzagPhase = 0
}
}
AdvancePosition(&e.Position, dt)
}

// MoveRandom applies a random walk: random perturbation to heading each tick.
func MoveRandom(e *SimEntity, dt time.Duration, rng *rand.Rand) {
e.Position.Heading = normalizeHeading(e.Position.Heading + (rng.Float64()*20 - 10))
e.Position.SpeedKn += rng.Float64()*2 - 1
if e.Position.SpeedKn < 0.5 {
e.Position.SpeedKn = 0.5
}
AdvancePosition(&e.Position, dt)
}

// bearingTo calculates the true initial bearing from 'from' to 'to' in degrees.
// 0° = North, 90° = East, 180° = South, 270° = West.
// Uses the standard spherical forward azimuth formula:
//
//y = sin(Δlon) × cos(lat2)
//x = cos(lat1) × sin(lat2) − sin(lat1) × cos(lat2) × cos(Δlon)
//θ = atan2(y, x)
func bearingTo(from, to Position) float64 {
dLon := (to.Lon - from.Lon) * degreesToRadians
fromLatRad := from.Lat * degreesToRadians
toLatRad := to.Lat * degreesToRadians

y := math.Sin(dLon) * math.Cos(toLatRad)
x := math.Cos(fromLatRad)*math.Sin(toLatRad) - math.Sin(fromLatRad)*math.Cos(toLatRad)*math.Cos(dLon)
brng := math.Atan2(y, x) * 180 / math.Pi
return normalizeHeading(brng)
}

// distanceNM calculates the approximate distance in nautical miles between two positions.
func distanceNM(a, b Position) float64 {
dLat := (b.Lat - a.Lat) * 60 // NM (1 degree lat = 60 NM)
cosLat := math.Cos(a.Lat * degreesToRadians)
dLon := (b.Lon - a.Lon) * 60 * cosLat
return math.Sqrt(dLat*dLat + dLon*dLon)
}
