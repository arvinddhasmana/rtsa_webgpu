// CLASSIFICATION: UNCLASSIFIED
package generator_test

import (
"math"
"math/rand"
"testing"
"time"

"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/generator"
)

// newTestEntity creates a minimal SimEntity for movement tests.
func newTestEntity() *generator.SimEntity {
return &generator.SimEntity{
ID: "TEST-001",
Position: generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 10.0,
Heading: 90.0, // East
AltM:    0,
},
LoiterRadiusNM: 1.0,
LoiterAngle:    0,
LoiterCenter:   generator.Position{Lat: 45.0, Lon: -60.0},
WaypointFwd:    true,
}
}

func TestAdvancePosition_East(t *testing.T) {
pos := &generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 60.0, // 60 knots for easy math
Heading: 90.0, // Due east
}

// At 60 knots for 1 hour → 60 NM east → 60/60/cos(45°) ≈ 1.414° lon change
generator.AdvancePosition(pos, time.Hour)

if pos.Lat != 45.0 {
t.Errorf("lat should not change for due-east travel, got %f", pos.Lat)
}
// Expected longitude change: 60NM * (1/60) / cos(45°) ≈ 1.414°
expectedDLon := 60.0 * (1.0 / 60.0) / math.Cos(45.0*math.Pi/180.0)
gotDLon := pos.Lon - (-60.0)
if math.Abs(gotDLon-expectedDLon) > 0.001 {
t.Errorf("expected dLon≈%f, got %f", expectedDLon, gotDLon)
}
}

func TestAdvancePosition_North(t *testing.T) {
pos := &generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 60.0, // 60 NM per hour
Heading: 0.0,  // Due north
}

generator.AdvancePosition(pos, time.Hour)

// 60 NM north = 1° latitude
if math.Abs(pos.Lat-46.0) > 0.001 {
t.Errorf("expected lat≈46.0, got %f", pos.Lat)
}
if math.Abs(pos.Lon-(-60.0)) > 0.001 {
t.Errorf("lon should be unchanged for due-north travel, got %f", pos.Lon)
}
}

func TestAdvancePosition_SmallStep(t *testing.T) {
pos := &generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 10.0,
Heading: 45.0,
}
origLat := pos.Lat
origLon := pos.Lon

generator.AdvancePosition(pos, time.Second)

// Position should change but only very slightly for 1-second step at 10 knots.
if pos.Lat == origLat && pos.Lon == origLon {
t.Error("position should have changed after advance")
}
// Change should be tiny.
dLat := math.Abs(pos.Lat - origLat)
if dLat > 0.001 {
t.Errorf("lat change too large for 1-second step: %f", dLat)
}
}

func TestMoveStraightLine_PositionAdvances(t *testing.T) {
rng := rand.New(rand.NewSource(42))
e := newTestEntity()
origLat := e.Position.Lat
origLon := e.Position.Lon

for i := 0; i < 10; i++ {
generator.MoveStraightLine(e, time.Second, rng)
}

if e.Position.Lat == origLat && e.Position.Lon == origLon {
t.Error("straight-line entity should have moved after 10 ticks")
}
}

func TestMoveStraightLine_HeadingPerturbation(t *testing.T) {
rng := rand.New(rand.NewSource(42))
e := newTestEntity()
origHeading := e.Position.Heading

// Run many ticks; heading should vary slightly.
headings := make([]float64, 0, 100)
for i := 0; i < 100; i++ {
generator.MoveStraightLine(e, time.Second, rng)
headings = append(headings, e.Position.Heading)
}

// Check that heading changes occurred (not all the same as original).
allSame := true
for _, h := range headings {
if math.Abs(h-origHeading) > 0.01 {
allSame = false
break
}
}
if allSame {
t.Error("heading should perturb over 100 ticks of straight-line movement")
}
}

func TestMovePatrol_ReachesWaypoint(t *testing.T) {
rng := rand.New(rand.NewSource(42))
e := newTestEntity()
e.Position.SpeedKn = 600 // Very fast to reach waypoints quickly

waypoints := []generator.Position{
{Lat: 45.1, Lon: -60.0},
{Lat: 44.9, Lon: -60.0},
}
e.Waypoints = waypoints
e.WaypointFwd = true

// Run until waypoint index changes (patrol reverses).
startIdx := e.WaypointIdx
for i := 0; i < 1000; i++ {
generator.MovePatrol(e, time.Second, rng, waypoints)
}

// After many ticks, the waypoint index should have changed at least once.
if e.WaypointIdx == startIdx && e.WaypointFwd {
t.Error("patrol should have reached a waypoint and changed direction")
}
}

func TestMovePatrol_EmptyWaypoints(t *testing.T) {
rng := rand.New(rand.NewSource(42))
e := newTestEntity()
origLat := e.Position.Lat

// Should not panic with empty waypoints.
generator.MovePatrol(e, time.Second, rng, nil)

// Position should still advance (falls back to AdvancePosition).
_ = origLat
}

func TestMoveLoitering_CircularPath(t *testing.T) {
rng := rand.New(rand.NewSource(42))
e := newTestEntity()
e.LoiterCenter = generator.Position{Lat: 45.0, Lon: -60.0}
e.LoiterRadiusNM = 1.0
e.LoiterAngle = 0
e.Position.SpeedKn = 5.0

// Record initial position relative to center.
center := e.LoiterCenter

// Run enough ticks to make a near-complete circle.
// Circumference ≈ 2π * 1 NM ≈ 6.28 NM; at 5 knots → 1.256 hours → 4524 seconds.
totalSec := 0
for totalSec < 4600 {
generator.MoveLoitering(e, time.Second, rng, center)
totalSec++
}

// After a full circle, should be close to initial offset from center.
distFromCenter := distNM(e.Position.Lat, e.Position.Lon, center.Lat, center.Lon)
if math.Abs(distFromCenter-1.0) > 0.1 {
t.Errorf("loitering entity should be ~1 NM from center, got %f NM", distFromCenter)
}
}

func TestMoveZigzag_HeadingAlternates(t *testing.T) {
rng := rand.New(rand.NewSource(42))
e := newTestEntity()
e.Position.Heading = 90.0

// Run 70 ticks (2+ zigzag periods of 30 ticks).
headings := make([]float64, 0, 70)
for i := 0; i < 70; i++ {
generator.MoveZigzag(e, time.Second, rng)
headings = append(headings, e.Position.Heading)
}

// Heading should have changed from initial.
finalHeading := headings[len(headings)-1]
if math.Abs(finalHeading-90.0) < 1.0 {
t.Error("zigzag entity heading should have changed from initial 90°")
}
}

func TestMoveRandom_MovesPosition(t *testing.T) {
rng := rand.New(rand.NewSource(42))
e := newTestEntity()
origLat := e.Position.Lat
origLon := e.Position.Lon

for i := 0; i < 10; i++ {
generator.MoveRandom(e, time.Second, rng)
}

if e.Position.Lat == origLat && e.Position.Lon == origLon {
t.Error("random-walk entity should have moved after 10 ticks")
}
}

// distNM is a helper for test distance calculation.
func distNM(lat1, lon1, lat2, lon2 float64) float64 {
dLat := (lat2 - lat1) * 60
cosLat := math.Cos(lat1 * math.Pi / 180)
dLon := (lon2 - lon1) * 60 * cosLat
return math.Sqrt(dLat*dLat + dLon*dLon)
}
