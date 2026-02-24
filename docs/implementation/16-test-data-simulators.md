<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 16 — Test Data Simulators

> **Module**: 16-test-data-simulators
> **Phase**: P5 (Testing Infrastructure)
> **Dependencies**: Module 02 (protos), Module 03 (shared libs)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 3 days

---

## 1. Objective

Implement a configurable test data simulator that generates synthetic sensor observations for all 6 sensor types. These simulators feed the ingestion services during development, integration testing, and demo scenarios.

**All generated data is UNCLASSIFIED — no real operational data.**

**Acceptance Criteria**:

- Generates synthetic data for all 6 sensor types (Radar, EW, ELINT, ISR, AIS, Cyber)
- Mid-Atlantic coordinates (43–47°N, 55–65°W) for geographic data
- Configurable: entity count, update rate, anomaly injection rate
- Produces realistic movement patterns (straight line, patrol, loitering, zigzag)
- Injects anomalies at configurable rate for anomaly detection validation
- Sends via gRPC to ingestion services
- Can be run as CLI tool or Docker container
- Deterministic mode (seed-based) for reproducible test scenarios
- ≥80% line coverage

---

## 2. Service Structure

```
tools/
└── simulator/
    ├── cmd/
    │   └── simulator/
    │       └── main.go              # CLI entry point
    ├── internal/
    │   ├── config/
    │   │   └── config.go            # Simulator configuration
    │   ├── generator/
    │   │   ├── entity.go            # Entity lifecycle manager
    │   │   ├── entity_test.go
    │   │   ├── movement.go          # Movement pattern generators
    │   │   ├── movement_test.go
    │   │   ├── anomaly_injector.go  # Anomaly pattern injector
    │   │   └── anomaly_injector_test.go
    │   ├── sensor/
    │   │   ├── radar.go             # Radar observation generator
    │   │   ├── radar_test.go
    │   │   ├── ew.go                # EW/SIGINT observation generator
    │   │   ├── ew_test.go
    │   │   ├── elint.go             # ELINT/COMINT observation generator
    │   │   ├── elint_test.go
    │   │   ├── isr.go               # ISR observation generator
    │   │   ├── isr_test.go
    │   │   ├── ais.go               # AIS/BFT observation generator
    │   │   ├── ais_test.go
    │   │   ├── cyber.go             # Cyber IOC generator
    │   │   └── cyber_test.go
    │   ├── scenario/
    │   │   ├── scenario.go          # Scenario definition & loader
    │   │   ├── default.go           # Default demo scenario
    │   │   └── stress.go            # High-volume stress scenario
    │   └── client/
    │       ├── grpc_sender.go       # gRPC client to ingestion services
    │       └── grpc_sender_test.go
    ├── scenarios/
    │   ├── default.yaml             # Default scenario config
    │   ├── stress.yaml              # Stress test scenario
    │   └── anomaly-demo.yaml        # Anomaly detection demo
    ├── go.mod
    ├── Dockerfile
    └── README.md
```

---

## 3. Configuration

```go
// CLASSIFICATION: UNCLASSIFIED
package config

type SimulatorConfig struct {
    // Target ingestion services
    RadarEndpoint   string `env:"SIM_RADAR_ENDPOINT" default:"localhost:50051"`
    EWEndpoint      string `env:"SIM_EW_ENDPOINT" default:"localhost:50052"`
    ELINTEndpoint   string `env:"SIM_ELINT_ENDPOINT" default:"localhost:50053"`
    ISREndpoint     string `env:"SIM_ISR_ENDPOINT" default:"localhost:50054"`
    AISEndpoint     string `env:"SIM_AIS_ENDPOINT" default:"localhost:50055"`
    CyberEndpoint   string `env:"SIM_CYBER_ENDPOINT" default:"localhost:50056"`

    // Entity configuration
    SurfaceEntityCount int `env:"SIM_SURFACE_ENTITIES" default:"20"`
    AirEntityCount     int `env:"SIM_AIR_ENTITIES" default:"10"`
    SubEntityCount     int `env:"SIM_SUB_ENTITIES" default:"5"`

    // Timing
    UpdateIntervalMs   int `env:"SIM_UPDATE_INTERVAL_MS" default:"1000"`
    DurationMinutes    int `env:"SIM_DURATION_MINUTES" default:"0"` // 0 = infinite

    // Anomaly injection
    AnomalyRate        float64 `env:"SIM_ANOMALY_RATE" default:"0.05"` // 5% of entities anomalous
    AnomalyTypes       string  `env:"SIM_ANOMALY_TYPES" default:"speed,route_deviation,ais_manipulation,behavioral"`

    // Reproducibility
    RandomSeed         int64   `env:"SIM_RANDOM_SEED" default:"0"` // 0 = random
    ScenarioFile       string  `env:"SIM_SCENARIO_FILE" default:""`

    // TLS
    TLSEnabled         bool   `env:"SIM_TLS_ENABLED" default:"false"`
    TLSCertFile        string `env:"SIM_TLS_CERT_FILE" default:""`
    TLSKeyFile         string `env:"SIM_TLS_KEY_FILE" default:""`
    TLSCAFile          string `env:"SIM_TLS_CA_FILE" default:""`
}
```

---

## 4. Entity Lifecycle

```go
// CLASSIFICATION: UNCLASSIFIED
package generator

// SimEntity represents a simulated entity moving through the operational area.
type SimEntity struct {
    ID              string
    EntityType      commonv1.EntityType
    HostileClass    commonv1.HostileClassification
    Position        Position
    Velocity        Velocity
    MovementPattern MovementPattern
    IsAnomalous     bool
    AnomalyType     commonv1.AnomalyType
    CreatedAt       time.Time
    TTL             time.Duration // Entity lifespan (0 = infinite)
}

// Position in WGS-84
type Position struct {
    Lat     float64 // 43.0 to 47.0 (Mid-Atlantic)
    Lon     float64 // -65.0 to -55.0 (Mid-Atlantic)
    AltM    float64 // Altitude in meters
    SpeedKn float64 // Speed in knots
    Heading float64 // Heading in degrees
}

// MovementPattern defines how an entity moves.
type MovementPattern int

const (
    PatternStraightLine MovementPattern = iota // Constant heading and speed
    PatternPatrol                               // Back-and-forth between waypoints
    PatternLoitering                            // Circular pattern around a point
    PatternZigzag                               // Alternating heading changes
    PatternRandom                               // Random walk
)

// EntityManager maintains all active simulated entities.
type EntityManager struct {
    entities map[string]*SimEntity
    rng      *rand.Rand
    config   *config.SimulatorConfig
}

// NewEntityManager creates the manager and initializes entities.
// Surface entities: speed 8-25 knots, heading 0-360
// Air entities: speed 150-550 knots, altitude 1000-15000m
// Subsurface: speed 5-15 knots, depth -50 to -400m (negative altitude)
func NewEntityManager(cfg *config.SimulatorConfig, rng *rand.Rand) *EntityManager { /* */ }

// Tick advances all entities by one time step.
// For each entity:
//   1. Apply movement pattern
//   2. If anomalous: inject anomaly behavior
//   3. Clamp to operational area boundaries
//   4. Update speed/heading per pattern
func (m *EntityManager) Tick(dt time.Duration) { /* */ }
```

---

## 5. Movement Patterns

```go
// CLASSIFICATION: UNCLASSIFIED
package generator

// MoveStraightLine: constant heading and speed with small random perturbation.
// Perturbation: heading ±2°, speed ±0.5 knots per tick
func MoveStraightLine(e *SimEntity, dt time.Duration, rng *rand.Rand) { /* */ }

// MovePatrol: follows a list of waypoints, reversing direction at ends.
// Speed maintained constant. Heading recalculated per tick toward next waypoint.
func MovePatrol(e *SimEntity, dt time.Duration, rng *rand.Rand, waypoints []Position) { /* */ }

// MoveLoitering: circular movement around a center point.
// Radius: 0.5-2.0 NM. Speed: reduced (5-8 knots for surface).
func MoveLoitering(e *SimEntity, dt time.Duration, rng *rand.Rand, center Position) { /* */ }

// MoveZigzag: alternating heading changes at regular intervals.
// Heading change: ±30-45° every 30-60 seconds.
func MoveZigzag(e *SimEntity, dt time.Duration, rng *rand.Rand) { /* */ }

// AdvancePosition moves the entity based on current speed and heading.
// Uses spherical geometry (great circle) for position update.
// lat_new = lat + (speed_nm / 60) * cos(heading) * dt_hours
// lon_new = lon + (speed_nm / 60) * sin(heading) / cos(lat) * dt_hours
func AdvancePosition(pos *Position, dt time.Duration) { /* */ }
```

---

## 6. Anomaly Injector

```go
// CLASSIFICATION: UNCLASSIFIED
package generator

// AnomalyInjector modifies entity behavior to simulate anomalies
// that the Anomaly Detection service (Module 08) should detect.
type AnomalyInjector struct {
    rng *rand.Rand
}

// InjectSpeedAnomaly: suddenly change speed to >3σ from normal.
// Surface: normal 12±5 knots → anomaly 35+ knots
// Air: normal 300±100 knots → anomaly 700+ knots
func (i *AnomalyInjector) InjectSpeedAnomaly(e *SimEntity) { /* */ }

// InjectRouteDeviation: sudden heading change >30° sustained for 3+ updates.
// Changes heading by 45-90° for 3-5 ticks, then resumes.
func (i *AnomalyInjector) InjectRouteDeviation(e *SimEntity) { /* */ }

// InjectAISManipulation: offsets AIS-reported position by >0.5 NM from true position.
// The radar observation reports the true position.
// The AIS observation reports a position offset by 0.5-2.0 NM.
// This requires generating BOTH radar and AIS for the same entity.
func (i *AnomalyInjector) InjectAISManipulation(e *SimEntity) Position { /* */ }

// InjectBehavioral: simulates loitering in non-loitering entity or
// zigzag pattern in straight-line entity.
// Confidence threshold: 0.75+
func (i *AnomalyInjector) InjectBehavioral(e *SimEntity) { /* */ }

// InjectProximity: moves entity into an exclusion zone.
// Exclusion zone center: 45.5°N, -60.0°W, radius 5 NM
func (i *AnomalyInjector) InjectProximity(e *SimEntity) { /* */ }
```

---

## 7. Sensor-Specific Generators

### 7.1 Radar Generator

```go
// CLASSIFICATION: UNCLASSIFIED
// Generates SensorObservation with RadarTrack payload.
// Fields populated:
//   - sensor_id: "RADAR-SIM-001" through "RADAR-SIM-003"
//   - sensor_type: SENSOR_TYPE_RADAR
//   - observation_time: current time
//   - classification: CLASSIFICATION_UNCLASSIFIED
//   - position: from entity
//   - radar_track.track_number: sequential per sensor
//   - radar_track.rcs_dbsm: -10 to 40 (surface), 0 to 50 (air)
//   - radar_track.iff_mode: random from {1, 2, 3A, 3C, 4, 5}
func GenerateRadarObservation(entity *SimEntity, sensorID string) *ingestionv1.SensorObservation { /* */ }
```

### 7.2 AIS Generator

```go
// CLASSIFICATION: UNCLASSIFIED
// Generates SensorObservation with AISPosition payload.
// Fields populated:
//   - sensor_id: "AIS-SIM-001"
//   - sensor_type: SENSOR_TYPE_AIS_BFT
//   - mmsi: 9-digit MMSI (generated per entity, stable)
//   - vessel_name: randomly selected from list
//   - vessel_type: 30 (fishing), 70 (cargo), 80 (tanker)
//   - ais_message_type: 1, 2, 3 (position reports)
//   - classification: CLASSIFICATION_UNCLASSIFIED
//
// If entity has AIS manipulation anomaly:
//   position is offset by 0.5-2.0 NM from true position
func GenerateAISObservation(entity *SimEntity, manipulatedPos *Position) *ingestionv1.SensorObservation { /* */ }
```

### 7.3 EW Generator

```go
// CLASSIFICATION: UNCLASSIFIED
// Generates EW/SIGINT intercepts.
// Only generated for entities with radar emissions.
//   - frequency_mhz: 2000-18000 (S-band to Ku-band)
//   - bearing_degrees: bearing from EW sensor to entity
//   - signal_strength_dbm: -80 to -20
//   - modulation_type: "PULSE", "CW", "FMCW"
//   - classification: CLASSIFICATION_UNCLASSIFIED
func GenerateEWObservation(entity *SimEntity, sensorID string, sensorPos Position) *ingestionv1.SensorObservation { /* */ }
```

### 7.4 Cyber Generator

```go
// CLASSIFICATION: UNCLASSIFIED
// Generates Cyber IOC observations.
// Not tied to a specific entity — generates random IOCs.
//   - ioc_type: "ipv4-addr", "domain-name", "url", "file-hash"
//   - ioc_value: randomly generated
//   - stix_id: "indicator--<uuid>"
//   - mitre_attack_ids: randomly selected ATT&CK technique IDs
//   - classification: CLASSIFICATION_UNCLASSIFIED
func GenerateCyberObservation(rng *rand.Rand) *ingestionv1.SensorObservation { /* */ }
```

---

## 8. Scenario Definitions

### 8.1 Default Scenario

```yaml
# CLASSIFICATION: UNCLASSIFIED
# tools/simulator/scenarios/default.yaml
name: "Default Demo Scenario"
description: "Standard operational scenario for development and demos"
seed: 42
duration_minutes: 0 # Run indefinitely

entities:
  surface:
    count: 20
    hostile_ratio: 0.15 # 15% hostile
    friendly_ratio: 0.40 # 40% friendly
    neutral_ratio: 0.30 # 30% neutral
    unknown_ratio: 0.15 # 15% unknown
    speed_range: [8, 25] # knots
    patterns:
      straight_line: 0.50
      patrol: 0.20
      loitering: 0.15
      random: 0.15

  air:
    count: 10
    hostile_ratio: 0.20
    friendly_ratio: 0.50
    neutral_ratio: 0.20
    unknown_ratio: 0.10
    speed_range: [150, 550]
    altitude_range: [1000, 15000]
    patterns:
      straight_line: 0.60
      patrol: 0.30
      random: 0.10

  subsurface:
    count: 5
    hostile_ratio: 0.20
    friendly_ratio: 0.40
    neutral_ratio: 0.20
    unknown_ratio: 0.20
    speed_range: [5, 15]
    depth_range: [50, 400]
    patterns:
      straight_line: 0.60
      patrol: 0.40

sensors:
  radar:
    count: 3
    sensor_ids: ["RADAR-SIM-001", "RADAR-SIM-002", "RADAR-SIM-003"]
    update_interval_ms: 2000
    coverage_nm: 100
  ais:
    count: 1
    sensor_ids: ["AIS-SIM-001"]
    update_interval_ms: 5000
  ew:
    count: 2
    sensor_ids: ["EW-SIM-001", "EW-SIM-002"]
    update_interval_ms: 3000
  cyber:
    count: 1
    sensor_ids: ["CYBER-SIM-001"]
    update_interval_ms: 10000
    ioc_rate: 2 # IOCs per interval

anomalies:
  injection_rate: 0.05 # 5% of entities exhibit anomalous behavior
  types:
    speed: 0.30
    route_deviation: 0.25
    ais_manipulation: 0.20
    behavioral: 0.15
    proximity: 0.10

operational_area:
  center: { lat: 45.0, lon: -60.0 }
  radius_nm: 150
  exclusion_zones:
    - name: "Halifax Harbor"
      center: { lat: 44.65, lon: -63.57 }
      radius_nm: 5
```

### 8.2 Stress Scenario

```yaml
# CLASSIFICATION: UNCLASSIFIED
# tools/simulator/scenarios/stress.yaml
name: "Stress Test Scenario"
description: "High-volume scenario for performance testing"
seed: 12345
duration_minutes: 30

entities:
  surface: { count: 200 }
  air: { count: 100 }
  subsurface: { count: 50 }

sensors:
  radar: { count: 10, update_interval_ms: 500 }
  ais: { count: 5, update_interval_ms: 1000 }
  ew: { count: 5, update_interval_ms: 1000 }
  elint: { count: 3, update_interval_ms: 2000 }
  isr: { count: 3, update_interval_ms: 5000 }
  cyber: { count: 2, update_interval_ms: 2000, ioc_rate: 10 }

anomalies:
  injection_rate: 0.10 # 10% anomalous
```

---

## 9. CLI Usage

```bash
# Run with default scenario
go run ./cmd/simulator/

# Run with specific scenario file
go run ./cmd/simulator/ --scenario scenarios/default.yaml

# Run with overrides
go run ./cmd/simulator/ \
  --surface-entities 50 \
  --air-entities 20 \
  --update-interval 500ms \
  --anomaly-rate 0.10 \
  --seed 42 \
  --duration 30m

# Run as Docker container
docker run --rm \
  --network rtsa-net \
  -e SIM_RADAR_ENDPOINT=svc-radar-ingestion:50051 \
  -e SIM_AIS_ENDPOINT=svc-ais-ingestion:50055 \
  rtsa/simulator:latest
```

---

## 10. Test Scenarios

| #   | Test                              | Expected                                |
| --- | --------------------------------- | --------------------------------------- |
| T01 | EntityManager with seed=42        | Deterministic entity positions          |
| T02 | MoveStraightLine 10 ticks         | Position advances correctly             |
| T03 | MovePatrol reaches waypoint       | Reverses direction                      |
| T04 | MoveLoitering 360°                | Returns near start position             |
| T05 | MoveZigzag heading changes        | Alternating ± heading                   |
| T06 | AdvancePosition math              | Correct great-circle advance            |
| T07 | Speed anomaly injection           | Speed >3σ from mean                     |
| T08 | AIS manipulation injection        | Offset >0.5 NM                          |
| T09 | Route deviation injection         | Heading change >30°                     |
| T10 | Radar observation valid           | All fields populated, passes validation |
| T11 | AIS observation valid             | MMSI 9 digits, valid fields             |
| T12 | Entities stay in operational area | Position clamped to bounds              |
| T13 | Scenario file parsing             | Config loaded correctly                 |
| T14 | gRPC sender connects              | Observations sent successfully          |

---

## 11. Agent Invocation

```
@greatest-ever-developer Implement Module 16 from docs/implementation/16-test-data-simulators.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for SensorObservation proto
- Read docs/implementation/04-sensor-ingestion-radar.md §4 for validation rules
- Read docs/implementation/05-sensor-ingestion-batch.md for all sensor validation

Deliverables:
1. Complete tools/simulator/ project
2. Entity manager with 5 movement patterns
3. 6 sensor-specific generators (radar, ew, elint, isr, ais, cyber)
4. Anomaly injector (speed, route, AIS manipulation, behavioral, proximity)
5. Scenario files (default.yaml, stress.yaml, anomaly-demo.yaml)
6. CLI with flag overrides
7. Dockerfile for container-based execution
8. Unit tests (≥80% coverage)

CRITICAL:
- ALL generated data MUST be CLASSIFICATION_UNCLASSIFIED
- Coordinates MUST be in Mid-Atlantic (43-47°N, 55-65°W)
- Generated observations MUST pass ingestion validation rules
- Deterministic mode (seed-based) for reproducible tests
- AIS manipulation anomaly requires both radar AND AIS observations for same entity
```
