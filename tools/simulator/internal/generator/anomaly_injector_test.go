// CLASSIFICATION: UNCLASSIFIED
package generator_test

import (
"math"
"math/rand"
"testing"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/generator"
)

func testSurfaceEntity() *generator.SimEntity {
return &generator.SimEntity{
ID:         "SURF-TEST-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Position: generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 12.0,
Heading: 90.0,
},
MovementPattern: generator.PatternStraightLine,
IsAnomalous:     true,
}
}

func testAirEntity() *generator.SimEntity {
return &generator.SimEntity{
ID:         "AIR-TEST-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
Position: generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 300.0,
Heading: 180.0,
AltM:    5000,
},
MovementPattern: generator.PatternStraightLine,
IsAnomalous:     true,
}
}

func TestInjectSpeedAnomaly_Surface(t *testing.T) {
rng := rand.New(rand.NewSource(42))
inj := generator.NewAnomalyInjector(rng)
e := testSurfaceEntity()

inj.InjectSpeedAnomaly(e)

// Surface anomaly: speed should be ≥35 knots (>3σ from mean 12, σ=5).
if e.Position.SpeedKn < 35.0 {
t.Errorf("surface speed anomaly should be ≥35 knots, got %f", e.Position.SpeedKn)
}
}

func TestInjectSpeedAnomaly_Air(t *testing.T) {
rng := rand.New(rand.NewSource(42))
inj := generator.NewAnomalyInjector(rng)
e := testAirEntity()

inj.InjectSpeedAnomaly(e)

// Air anomaly: speed should be ≥700 knots.
if e.Position.SpeedKn < 700.0 {
t.Errorf("air speed anomaly should be ≥700 knots, got %f", e.Position.SpeedKn)
}
}

func TestInjectRouteDeviation_HeadingChange(t *testing.T) {
rng := rand.New(rand.NewSource(42))
inj := generator.NewAnomalyInjector(rng)
e := testSurfaceEntity()
origHeading := e.Position.Heading

inj.InjectRouteDeviation(e)

// Heading change must be >30°.
delta := math.Abs(e.Position.Heading - origHeading)
if delta > 180 {
delta = 360 - delta
}
if delta < 30.0 {
t.Errorf("route deviation should cause heading change >30°, got %f°", delta)
}
}

func TestInjectAISManipulation_OffsetSufficient(t *testing.T) {
rng := rand.New(rand.NewSource(42))
inj := generator.NewAnomalyInjector(rng)
e := testSurfaceEntity()

manipulated := inj.InjectAISManipulation(e)

// Offset must be ≥0.5 NM.
offsetNM := generator.OffsetNM(e.Position, manipulated)
if offsetNM < 0.5 {
t.Errorf("AIS manipulation offset must be ≥0.5 NM, got %f NM", offsetNM)
}
// And ≤2.0 NM (with some tolerance).
if offsetNM > 2.5 {
t.Errorf("AIS manipulation offset should be ≤2.0 NM, got %f NM", offsetNM)
}
}

func TestInjectAISManipulation_InBounds(t *testing.T) {
rng := rand.New(rand.NewSource(42))
inj := generator.NewAnomalyInjector(rng)
e := testSurfaceEntity()
e.Position = generator.Position{Lat: 45.0, Lon: -60.0, SpeedKn: 10.0, Heading: 90.0}

manipulated := inj.InjectAISManipulation(e)

if manipulated.Lat < generator.MinLat || manipulated.Lat > generator.MaxLat {
t.Errorf("manipulated lat %f out of bounds", manipulated.Lat)
}
if manipulated.Lon < generator.MinLon || manipulated.Lon > generator.MaxLon {
t.Errorf("manipulated lon %f out of bounds", manipulated.Lon)
}
}

func TestInjectBehavioral_ChangesPattern(t *testing.T) {
rng := rand.New(rand.NewSource(42))
inj := generator.NewAnomalyInjector(rng)
e := testSurfaceEntity()
e.MovementPattern = generator.PatternStraightLine

inj.InjectBehavioral(e)

if e.MovementPattern == generator.PatternStraightLine {
t.Error("behavioral injection should change movement pattern")
}
}

func TestInjectProximity_MovesTowardZone(t *testing.T) {
rng := rand.New(rand.NewSource(42))
inj := generator.NewAnomalyInjector(rng)
e := testSurfaceEntity()
// Place entity far from exclusion zone.
e.Position = generator.Position{Lat: 43.5, Lon: -64.0, SpeedKn: 10.0, Heading: 0.0}

zone := generator.Position{Lat: generator.ExclusionLat, Lon: generator.ExclusionLon}
distBefore := generator.OffsetNM(e.Position, zone)

inj.InjectProximity(e)

distAfter := generator.OffsetNM(e.Position, zone)
if distAfter >= distBefore {
t.Errorf("proximity injection should move entity closer to zone: before %f NM, after %f NM",
distBefore, distAfter)
}
}

func TestAnomalyInjector_Deterministic(t *testing.T) {
rng1 := rand.New(rand.NewSource(42))
rng2 := rand.New(rand.NewSource(42))

inj1 := generator.NewAnomalyInjector(rng1)
inj2 := generator.NewAnomalyInjector(rng2)

e1 := testSurfaceEntity()
e2 := testSurfaceEntity()

inj1.InjectSpeedAnomaly(e1)
inj2.InjectSpeedAnomaly(e2)

if e1.Position.SpeedKn != e2.Position.SpeedKn {
t.Errorf("speed anomaly should be deterministic: %f vs %f",
e1.Position.SpeedKn, e2.Position.SpeedKn)
}
}
