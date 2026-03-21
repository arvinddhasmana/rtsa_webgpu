// CLASSIFICATION: UNCLASSIFIED
package generator

import (
"fmt"
"math"
"math/rand"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/config"
)

// defaultAreaBounds is the North Atlantic fallback used when no scenario
// OperationalArea is configured.
const (
MinLat = 43.0
MaxLat = 47.0
MinLon = -65.0
MaxLon = -55.0
)

// AreaBounds defines the geographic bounding box for entity generation.
// All values are in decimal degrees (WGS-84).
type AreaBounds struct {
MinLat float64
MaxLat float64
MinLon float64
MaxLon float64
}

// defaultBounds is the North Atlantic Mid-Atlantic operational area.
var defaultBounds = AreaBounds{
MinLat: MinLat,
MaxLat: MaxLat,
MinLon: MinLon,
MaxLon: MaxLon,
}

// NauticalMilesDegrees is the number of degrees latitude per nautical mile.
const NauticalMilesDegrees = 1.0 / 60.0

// AreaBoundsFromCenter derives a bounding box from a center point and radius.
// radiusNM is the radius in nautical miles. The longitude extent is widened
// by 1/cos(lat) to approximate equal coverage at the given latitude.
func AreaBoundsFromCenter(centerLat, centerLon, radiusNM float64) AreaBounds {
latDelta := radiusNM * NauticalMilesDegrees
cosLat := math.Cos(centerLat * math.Pi / 180.0)
if math.Abs(cosLat) < 0.01 {
cosLat = 0.01
}
lonDelta := latDelta / cosLat
return AreaBounds{
MinLat: centerLat - latDelta,
MaxLat: centerLat + latDelta,
MinLon: centerLon - lonDelta,
MaxLon: centerLon + lonDelta,
}
}

// MovementPattern defines how an entity moves over time.
type MovementPattern int

const (
PatternStraightLine MovementPattern = iota // Constant heading and speed
PatternPatrol                               // Back-and-forth between waypoints
PatternLoitering                            // Circular pattern around a point
PatternZigzag                               // Alternating heading changes
PatternRandom                               // Random walk
)

// Position represents a WGS-84 geographic position with speed/heading.
type Position struct {
Lat     float64 // Latitude in degrees  (43-47°N)
Lon     float64 // Longitude in degrees (-65 to -55°W)
AltM    float64 // Altitude in meters (negative = below sea level)
SpeedKn float64 // Speed over ground in knots
Heading float64 // True heading in degrees (0-360)
}

// Velocity represents a 3-axis velocity vector in m/s.
type Velocity struct {
NorthMPS float64
EastMPS  float64
DownMPS  float64
}

// SimEntity represents a single simulated entity.
type SimEntity struct {
ID              string
EntityType      commonv1.EntityType
HostileClass    commonv1.HostileClassification
Position        Position
Velocity        Velocity
MovementPattern MovementPattern
IsAnomalous     bool
AnomalyType     commonv1.AnomalyType

// For patrol pattern: waypoints and current waypoint index
Waypoints      []Position
WaypointIdx    int
WaypointFwd    bool // travelling forward through waypoints
LoiterCenter   Position
LoiterRadiusNM float64
LoiterAngle    float64 // current angle in loiter circle (radians)

// Zigzag state
ZigzagTick  int
ZigzagPhase int // 0 = left, 1 = right

// AIS manipulation: offset from true position
AISOffset Position

CreatedAt time.Time
TTL       time.Duration // 0 = infinite
}

// EntityManager maintains all active simulated entities.
type EntityManager struct {
entities map[string]*SimEntity
rng      *rand.Rand
cfg      *config.SimulatorConfig
injector *AnomalyInjector
tick     int64
area     AreaBounds // geographic bounding box for entity generation
}

// NewEntityManager creates the manager using the default North Atlantic bounds.
func NewEntityManager(cfg *config.SimulatorConfig, rng *rand.Rand) *EntityManager {
return NewEntityManagerWithBounds(cfg, rng, defaultBounds)
}

// NewEntityManagerWithBounds creates the manager using the supplied AreaBounds.
// Use AreaBoundsFromCenter to derive bounds from a scenario OperationalArea.
func NewEntityManagerWithBounds(cfg *config.SimulatorConfig, rng *rand.Rand, area AreaBounds) *EntityManager {
m := &EntityManager{
entities: make(map[string]*SimEntity),
rng:      rng,
cfg:      cfg,
injector: NewAnomalyInjector(rng),
area:     area,
}

// Determine anomalous entity count per domain.
totalSurface := cfg.SurfaceEntityCount
totalAir := cfg.AirEntityCount
totalSub := cfg.SubEntityCount

anomalousSurface := int(float64(totalSurface) * cfg.AnomalyRate)
anomalousAir := int(float64(totalAir) * cfg.AnomalyRate)
anomalousSub := int(float64(totalSub) * cfg.AnomalyRate)

for i := 0; i < totalSurface; i++ {
e := m.newSurfaceEntity(fmt.Sprintf("SURF-%04d", i+1), i < anomalousSurface)
m.entities[e.ID] = e
}
for i := 0; i < totalAir; i++ {
e := m.newAirEntity(fmt.Sprintf("AIR-%04d", i+1), i < anomalousAir)
m.entities[e.ID] = e
}
for i := 0; i < totalSub; i++ {
e := m.newSubEntity(fmt.Sprintf("SUB-%04d", i+1), i < anomalousSub)
m.entities[e.ID] = e
}

return m
}

// Entities returns a snapshot of all active entities.
func (m *EntityManager) Entities() map[string]*SimEntity {
return m.entities
}

// Tick advances all entities by dt.
func (m *EntityManager) Tick(dt time.Duration) {
m.tick++
injector := m.injector
for _, e := range m.entities {
// Apply movement pattern
switch e.MovementPattern {
case PatternStraightLine:
MoveStraightLine(e, dt, m.rng)
case PatternPatrol:
MovePatrol(e, dt, m.rng, e.Waypoints)
case PatternLoitering:
MoveLoitering(e, dt, m.rng, e.LoiterCenter)
case PatternZigzag:
MoveZigzag(e, dt, m.rng)
case PatternRandom:
MoveRandom(e, dt, m.rng)
}

// Inject anomaly behaviour for anomalous entities periodically.
if e.IsAnomalous && m.tick%5 == 0 {
switch e.AnomalyType {
case commonv1.AnomalyType_ANOMALY_TYPE_SPEED:
injector.InjectSpeedAnomaly(e)
case commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION:
injector.InjectRouteDeviation(e)
case commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION:
e.AISOffset = injector.InjectAISManipulation(e)
case commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL:
injector.InjectBehavioral(e)
case commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY:
injector.InjectProximity(e)
}
}

// Clamp to operational area.
		m.clampToArea(&e.Position)
}
}

// ── Entity constructors ────────────────────────────────────────────────────

func (m *EntityManager) newSurfaceEntity(id string, anomalous bool) *SimEntity {
e := &SimEntity{
ID:         id,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
HostileClass: randomHostile(m.rng,
[]commonv1.HostileClassification{
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
},
[]float64{0.40, 0.15, 0.30, 0.15},
),
Position: Position{
Lat:     m.area.MinLat + m.rng.Float64()*(m.area.MaxLat-m.area.MinLat),
Lon:     m.area.MinLon + m.rng.Float64()*(m.area.MaxLon-m.area.MinLon),
AltM:    0,
SpeedKn: 8 + m.rng.Float64()*17, // 8-25 knots
Heading: m.rng.Float64() * 360,
},
MovementPattern: randomPattern(m.rng, []MovementPattern{
PatternStraightLine, PatternPatrol, PatternLoitering, PatternRandom,
}, []float64{0.50, 0.20, 0.15, 0.15}),
IsAnomalous: anomalous,
CreatedAt:   time.Now(),
}
m.finaliseEntity(e, anomalous)
return e
}

func (m *EntityManager) newAirEntity(id string, anomalous bool) *SimEntity {
e := &SimEntity{
ID:         id,
EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
HostileClass: randomHostile(m.rng,
[]commonv1.HostileClassification{
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
},
[]float64{0.50, 0.20, 0.20, 0.10},
),
Position: Position{
Lat:     m.area.MinLat + m.rng.Float64()*(m.area.MaxLat-m.area.MinLat),
Lon:     m.area.MinLon + m.rng.Float64()*(m.area.MaxLon-m.area.MinLon),
AltM:    1000 + m.rng.Float64()*14000, // 1000-15000 m
SpeedKn: 150 + m.rng.Float64()*400,    // 150-550 knots
Heading: m.rng.Float64() * 360,
},
MovementPattern: randomPattern(m.rng, []MovementPattern{
PatternStraightLine, PatternPatrol, PatternRandom,
}, []float64{0.60, 0.30, 0.10}),
IsAnomalous: anomalous,
CreatedAt:   time.Now(),
}
m.finaliseEntity(e, anomalous)
return e
}

func (m *EntityManager) newSubEntity(id string, anomalous bool) *SimEntity {
e := &SimEntity{
ID:         id,
EntityType: commonv1.EntityType_ENTITY_TYPE_SUBSURFACE,
HostileClass: randomHostile(m.rng,
[]commonv1.HostileClassification{
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
},
[]float64{0.40, 0.20, 0.20, 0.20},
),
Position: Position{
Lat:     m.area.MinLat + m.rng.Float64()*(m.area.MaxLat-m.area.MinLat),
Lon:     m.area.MinLon + m.rng.Float64()*(m.area.MaxLon-m.area.MinLon),
AltM:    -(50 + m.rng.Float64()*350), // -50 to -400 m
SpeedKn: 5 + m.rng.Float64()*10,      // 5-15 knots
Heading: m.rng.Float64() * 360,
},
MovementPattern: randomPattern(m.rng, []MovementPattern{
PatternStraightLine, PatternPatrol,
}, []float64{0.60, 0.40}),
IsAnomalous: anomalous,
CreatedAt:   time.Now(),
}
m.finaliseEntity(e, anomalous)
return e
}

// finaliseEntity sets anomaly type and patrol/loiter parameters.
func (m *EntityManager) finaliseEntity(e *SimEntity, anomalous bool) {
if anomalous {
e.AnomalyType = randomAnomalyType(m.rng)
}

// Generate patrol waypoints if needed.
if e.MovementPattern == PatternPatrol {
wp1 := Position{
Lat: m.area.MinLat + m.rng.Float64()*(m.area.MaxLat-m.area.MinLat),
Lon: m.area.MinLon + m.rng.Float64()*(m.area.MaxLon-m.area.MinLon),
}
wp2 := Position{
Lat: m.area.MinLat + m.rng.Float64()*(m.area.MaxLat-m.area.MinLat),
Lon: m.area.MinLon + m.rng.Float64()*(m.area.MaxLon-m.area.MinLon),
}
e.Waypoints = []Position{wp1, wp2}
e.WaypointFwd = true
}

// Set loiter parameters.
if e.MovementPattern == PatternLoitering {
e.LoiterCenter = e.Position
e.LoiterRadiusNM = 0.5 + m.rng.Float64()*1.5 // 0.5-2.0 NM
e.LoiterAngle = m.rng.Float64() * 2 * 3.14159265358979
}
}

// ── Helpers ────────────────────────────────────────────────────────────────

// clampToDefaultArea clamps a position to the default North Atlantic bounds.
// Used by AnomalyInjector which operates without access to the EntityManager.
// Entity generation and per-tick clamping use the manager method instead.
func clampToDefaultArea(pos *Position) {
if pos.Lat < defaultBounds.MinLat {
pos.Lat = defaultBounds.MinLat
}
if pos.Lat > defaultBounds.MaxLat {
pos.Lat = defaultBounds.MaxLat
}
if pos.Lon < defaultBounds.MinLon {
pos.Lon = defaultBounds.MinLon
}
if pos.Lon > defaultBounds.MaxLon {
pos.Lon = defaultBounds.MaxLon
}
}

// clampToArea clamps a position to the manager's configured AreaBounds.
func (m *EntityManager) clampToArea(pos *Position) {
if pos.Lat < m.area.MinLat {
pos.Lat = m.area.MinLat
}
if pos.Lat > m.area.MaxLat {
pos.Lat = m.area.MaxLat
}
if pos.Lon < m.area.MinLon {
pos.Lon = m.area.MinLon
}
if pos.Lon > m.area.MaxLon {
pos.Lon = m.area.MaxLon
}
}

// randomHostile picks a hostile classification using probability weights.
func randomHostile(rng *rand.Rand, classes []commonv1.HostileClassification, weights []float64) commonv1.HostileClassification {
r := rng.Float64()
cum := 0.0
for i, w := range weights {
cum += w
if r < cum {
return classes[i]
}
}
return classes[len(classes)-1]
}

// randomPattern picks a movement pattern using probability weights.
func randomPattern(rng *rand.Rand, patterns []MovementPattern, weights []float64) MovementPattern {
r := rng.Float64()
cum := 0.0
for i, w := range weights {
cum += w
if r < cum {
return patterns[i]
}
}
return patterns[len(patterns)-1]
}

// randomAnomalyType picks a random anomaly type.
func randomAnomalyType(rng *rand.Rand) commonv1.AnomalyType {
types := []commonv1.AnomalyType{
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION,
commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION,
commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL,
commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY,
}
return types[rng.Intn(len(types))]
}
