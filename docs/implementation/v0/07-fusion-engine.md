<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 07 — Fusion Engine

> **Module**: 07-fusion-engine
> **Phase**: P2 (Core Processing)
> **Dependencies**: Module 02 (protos), Module 03 (shared libraries), Module 04/05 (ingestion services produce to sensors.\* topics)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 6 days

---

## 1. Objective

Implement the Multi-Source Fusion Engine (`svc-fusion-engine`) that consumes sensor observations from all `sensors.*` topics, correlates them using configurable gating and scoring algorithms, maintains fused track state with Kalman filter estimation, and produces `FusedTrack` events to `tracks.fused.{entity_type}` topics.

**Acceptance Criteria**:

- Consumes from all 6 sensor topics simultaneously (consumer group: `fusion-engine`)
- Gating filter applies configurable spatial/temporal thresholds per entity type
- Correlation scorer uses weighted scoring (position 0.35, velocity 0.25, type 0.20, temporal 0.20)
- Auto-correlate at ≥0.85, tentative 0.60–0.84, new track <0.60
- Kalman filter updates fused position/velocity estimates
- Track merging when two tracks correlate ≥0.85
- Stale track detection (60s no update → STALE, 5 min → DROPPED)
- Produces FusedTrack to `tracks.fused.{surface|air|subsurface|land|cyber}`
- Classification propagation: MAX of all contributing sources
- ≥80% line coverage on domain logic

---

## 2. Service Structure

```
svc-fusion-engine/
├── cmd/
│   └── fusion-engine/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── gating.go              # Spatial/temporal gating filter
│   │   ├── gating_test.go
│   │   ├── scorer.go              # Correlation scoring
│   │   ├── scorer_test.go
│   │   ├── kalman.go              # Kalman filter state estimator
│   │   ├── kalman_test.go
│   │   ├── track_manager.go       # In-memory track state management
│   │   ├── track_manager_test.go
│   │   ├── merger.go              # Track merging logic
│   │   ├── merger_test.go
│   │   └── stale_monitor.go       # Background stale track detector
│   │   └── stale_monitor_test.go
│   ├── consumer/
│   │   ├── sensor_consumer.go     # Multi-topic consumer
│   │   └── sensor_consumer_test.go
│   ├── producer/
│   │   ├── track_producer.go      # Produces to tracks.fused.{type}
│   │   └── track_producer_test.go
│   └── handler/
│       ├── fusion_pipeline.go     # Orchestrates the fusion flow
│       └── fusion_pipeline_test.go
├── go.mod
├── Dockerfile
└── README.md
```

---

## 3. Configuration

```go
// CLASSIFICATION: UNCLASSIFIED
package config

type Config struct {
    config.BaseConfig

    // ── Consumer ──
    InputTopics    []string // RTSA_FUSION_INPUT_TOPICS (default: all sensors.* topics)
    ConsumerGroup  string   // RTSA_FUSION_CONSUMER_GROUP (default: "fusion-engine")

    // ── Gating Thresholds ──
    GateSurfaceDistanceNM float64 // RTSA_FUSION_GATE_SURFACE_DISTANCE (default: 5.0)
    GateSurfaceTimeSec    int     // RTSA_FUSION_GATE_SURFACE_TIME (default: 30)
    GateAirDistanceNM     float64 // RTSA_FUSION_GATE_AIR_DISTANCE (default: 20.0)
    GateAirTimeSec        int     // RTSA_FUSION_GATE_AIR_TIME (default: 15)
    GateSubDistanceNM     float64 // RTSA_FUSION_GATE_SUB_DISTANCE (default: 2.0)
    GateSubTimeSec        int     // RTSA_FUSION_GATE_SUB_TIME (default: 60)

    // ── Correlation Weights ──
    WeightPosition float64 // RTSA_FUSION_WEIGHT_POSITION (default: 0.35)
    WeightVelocity float64 // RTSA_FUSION_WEIGHT_VELOCITY (default: 0.25)
    WeightType     float64 // RTSA_FUSION_WEIGHT_TYPE (default: 0.20)
    WeightTemporal float64 // RTSA_FUSION_WEIGHT_TEMPORAL (default: 0.20)

    // ── Correlation Thresholds ──
    AutoCorrelateThreshold    float64 // RTSA_FUSION_AUTO_THRESHOLD (default: 0.85)
    TentativeCorrelateThreshold float64 // RTSA_FUSION_TENTATIVE_THRESHOLD (default: 0.60)

    // ── Track Lifecycle ──
    StaleTimeoutSec   int // RTSA_FUSION_STALE_TIMEOUT (default: 60)
    DropTimeoutSec    int // RTSA_FUSION_DROP_TIMEOUT (default: 300)
    StaleCheckInterval int // RTSA_FUSION_STALE_CHECK_INTERVAL (default: 10)

    // ── Output ──
    OutputTopicPrefix string // RTSA_FUSION_OUTPUT_PREFIX (default: "tracks.fused")
}
```

---

## 4. Domain Logic Specifications

### 4.1 Gating Filter (`gating.go`)

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// GatingConfig holds per-entity-type gating parameters.
type GatingConfig struct {
    MaxDistanceNM float64
    MaxTimeDelta  time.Duration
}

// GatingFilter determines which existing tracks are candidates
// for correlation with a new observation.
type GatingFilter struct {
    configs map[commonv1.EntityType]GatingConfig
}

// NewGatingFilter creates a gating filter with configured thresholds.
func NewGatingFilter(surface, air, sub GatingConfig) *GatingFilter { /* implementation */ }

// FindCandidates returns all tracks within the gate for the given observation.
// Algorithm:
//   1. Determine entity type from observation (radar → Surface/Air, AIS → Surface, etc.)
//   2. Look up GatingConfig for that entity type
//   3. For each active track of compatible entity type:
//      a. Calculate Haversine distance between observation position and track position
//      b. Calculate time delta between observation time and track last update
//      c. If distance ≤ MaxDistanceNM AND timeDelta ≤ MaxTimeDelta → candidate
//   4. Return sorted by distance (nearest first)
func (g *GatingFilter) FindCandidates(obs *ingestionv1.SensorObservation,
    tracks []*TrackState) []*TrackState { /* implementation */ }

// HaversineDistanceNM calculates the great-circle distance between two
// positions in nautical miles using the Haversine formula.
func HaversineDistanceNM(lat1, lon1, lat2, lon2 float64) float64 { /* implementation */ }
```

### 4.2 Correlation Scorer (`scorer.go`)

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// CorrelationScorer computes a weighted correlation score between
// a new observation and an existing track.
type CorrelationScorer struct {
    weightPosition float64 // 0.35
    weightVelocity float64 // 0.25
    weightType     float64 // 0.20
    weightTemporal float64 // 0.20
}

// CorrelationResult holds the scoring outcome.
type CorrelationResult struct {
    Score           float64 // 0.0 to 1.0
    PositionScore   float64 // Component score
    VelocityScore   float64
    TypeScore       float64
    TemporalScore   float64
    Action          CorrelationAction
}

type CorrelationAction int
const (
    ActionAutoCorrelate CorrelationAction = iota // ≥ 0.85
    ActionTentative                               // 0.60 – 0.84
    ActionNewTrack                                // < 0.60
)

// Score calculates the weighted correlation score.
// Components:
//
// PositionScore (weight: 0.35):
//   1 - (distance_nm / gate_distance_nm)
//   Clamped to [0, 1]
//
// VelocityScore (weight: 0.25):
//   If both have velocity: 1 - |speed_diff| / max_speed
//   If either missing: 0.5 (neutral)
//
// TypeScore (weight: 0.20):
//   1.0 if same entity type
//   0.5 if either is UNSPECIFIED
//   0.0 if different types
//
// TemporalScore (weight: 0.20):
//   1 - (time_delta_sec / gate_time_sec)
//   Clamped to [0, 1]
//
func (s *CorrelationScorer) Score(obs *ingestionv1.SensorObservation,
    track *TrackState, gateConfig GatingConfig) *CorrelationResult { /* implementation */ }
```

### 4.3 Kalman Filter (`kalman.go`)

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// KalmanState holds the kinematic state vector [lat, lon, vN, vE].
type KalmanState struct {
    Latitude  float64 // degrees
    Longitude float64 // degrees
    VelocityN float64 // m/s north
    VelocityE float64 // m/s east

    // Covariance matrix (4x4, stored as flat array)
    P [16]float64

    // Last update time
    LastUpdate time.Time
}

// KalmanFilter implements a constant-velocity Kalman filter
// for position/velocity state estimation.
type KalmanFilter struct {
    // Process noise spectral density (m/s²/√Hz)
    processNoise float64
    // Measurement noise: position (deg²), velocity (m²/s²)
    measurementNoisePos float64
    measurementNoiseVel float64
}

// NewKalmanFilter creates a Kalman filter with default noise parameters.
func NewKalmanFilter() *KalmanFilter { /* implementation */ }

// Predict extrapolates the state forward by dt seconds.
// Uses constant-velocity model: x(t+dt) = F * x(t)
// F = [1, 0, dt, 0]
//     [0, 1, 0, dt]
//     [0, 0, 1,  0]
//     [0, 0, 0,  1]
// Converts velocity from m/s to degrees using:
//   dLat = vN * dt / (111320)
//   dLon = vE * dt / (111320 * cos(lat))
func (kf *KalmanFilter) Predict(state *KalmanState, dt float64) { /* implementation */ }

// Update incorporates a new measurement to refine the state estimate.
// Uses the standard Kalman update equations:
//   K = P * H' * inv(H * P * H' + R)
//   x = x + K * (z - H * x)
//   P = (I - K * H) * P
func (kf *KalmanFilter) Update(state *KalmanState, measurement *Measurement) { /* implementation */ }

// Measurement represents a sensor measurement for Kalman update.
type Measurement struct {
    Latitude  float64
    Longitude float64
    VelocityN *float64 // nil if not available
    VelocityE *float64
    Time      time.Time
}
```

### 4.4 Track Manager (`track_manager.go`)

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

import (
    "sync"
    entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
)

// TrackState holds the mutable state of a fused track in memory.
type TrackState struct {
    TrackID       string
    EntityType    commonv1.EntityType
    HostileClass  commonv1.HostileClassification
    KalmanState   *KalmanState
    Sources       map[string]*SourceInfo // key: sensor_id
    Classification commonv1.ClassificationLevel
    Status        commonv1.TrackStatus
    CreatedAt     time.Time
    UpdatedAt     time.Time
    Label         string
}

// SourceInfo tracks per-sensor contribution metadata.
type SourceInfo struct {
    SensorID         string
    SensorType       commonv1.SensorType
    Confidence       float64
    LastContribution time.Time
    ObservationCount uint32
}

// TrackManager manages the in-memory state of all active tracks.
// Thread-safe via read-write mutex.
type TrackManager struct {
    mu     sync.RWMutex
    tracks map[string]*TrackState // key: track_id
    kf     *KalmanFilter
}

// NewTrackManager creates a track manager with Kalman filter.
func NewTrackManager(kf *KalmanFilter) *TrackManager { /* implementation */ }

// CreateTrack creates a new fused track from an initial observation.
// Generates UUID v7 for track_id.
// Sets initial Kalman state from observation position/velocity.
func (tm *TrackManager) CreateTrack(obs *ingestionv1.SensorObservation) (*TrackState, error) { /* implementation */ }

// UpdateTrack incorporates a new observation into an existing track.
// Steps:
//   1. Predict Kalman state to observation time
//   2. Update Kalman state with observation measurement
//   3. Update source attribution (add or update sensor info)
//   4. Recompute confidence from source count and Kalman uncertainty
//   5. Propagate classification (MAX of all sources)
//   6. Update timestamp
func (tm *TrackManager) UpdateTrack(trackID string, obs *ingestionv1.SensorObservation) (*TrackState, error) { /* implementation */ }

// MergeTracks merges trackB into trackA.
// trackB status set to MERGED, all sources transferred to trackA.
// Returns the updated trackA.
func (tm *TrackManager) MergeTracks(trackAID, trackBID string) (*TrackState, error) { /* implementation */ }

// GetActiveTracks returns all tracks with status ACTIVE or STALE.
func (tm *TrackManager) GetActiveTracks() []*TrackState { /* implementation */ }

// GetTrack returns a specific track by ID.
func (tm *TrackManager) GetTrack(trackID string) (*TrackState, bool) { /* implementation */ }

// MarkStale marks a track as STALE.
func (tm *TrackManager) MarkStale(trackID string) { /* implementation */ }

// MarkDropped marks a track as DROPPED and removes from active set.
func (tm *TrackManager) MarkDropped(trackID string) { /* implementation */ }

// TrackCount returns the total number of active tracks.
func (tm *TrackManager) TrackCount() int { /* implementation */ }

// ToFusedTrack converts internal TrackState to a FusedTrack proto.
func (ts *TrackState) ToFusedTrack() *entityv1.FusedTrack { /* implementation */ }
```

### 4.5 Stale Track Monitor (`stale_monitor.go`)

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// StaleMonitor runs a background goroutine that periodically scans
// active tracks for staleness.
type StaleMonitor struct {
    manager       *TrackManager
    staleTimeout  time.Duration // default: 60s
    dropTimeout   time.Duration // default: 5 min
    checkInterval time.Duration // default: 10s
    onStatusChange func(track *TrackState, oldStatus, newStatus commonv1.TrackStatus)
}

// Start begins the stale monitoring loop. Blocks until ctx is cancelled.
// For each check cycle:
//   1. Get all active tracks
//   2. For each track:
//      - If ACTIVE and (now - lastUpdate) > staleTimeout → mark STALE
//      - If STALE and (now - lastUpdate) > dropTimeout → mark DROPPED
//   3. Call onStatusChange callback for each transition
func (sm *StaleMonitor) Start(ctx context.Context) { /* implementation */ }
```

### 4.6 Fusion Pipeline (`handler/fusion_pipeline.go`)

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

// FusionPipeline orchestrates the end-to-end fusion flow.
// It is the message handler passed to the Redpanda consumer.
type FusionPipeline struct {
    gating   *domain.GatingFilter
    scorer   *domain.CorrelationScorer
    manager  *domain.TrackManager
    producer *producer.TrackProducer
    audit    *audit.Emitter
    logger   *zap.Logger
    metrics  *FusionMetrics
}

// HandleObservation processes a single sensor observation through the fusion pipeline.
// Flow:
//   1. Deserialize SensorObservation from record value
//   2. Find gated candidates from track manager
//   3. If no candidates → CreateTrack → produce NEW event
//   4. If candidates exist → score each:
//      a. Best score ≥ autoCorrelateThreshold → UpdateTrack → produce UPDATED event
//      b. Best score ≥ tentativeThreshold → UpdateTrack with tentative flag
//      c. Best score < tentativeThreshold → CreateTrack → produce NEW event
//   5. Check if any pair of tracks should merge (bidirectional scoring ≥ 0.85)
//   6. Emit audit event for track creation/update/merge
//   7. Update metrics
func (fp *FusionPipeline) HandleObservation(ctx context.Context, record *kgo.Record) error { /* implementation */ }
```

---

## 5. Topic Routing

FusedTrack output topics are determined by entity type:

| Entity Type | Output Topic              |
| ----------- | ------------------------- |
| SURFACE     | `tracks.fused.surface`    |
| AIR         | `tracks.fused.air`        |
| SUBSURFACE  | `tracks.fused.subsurface` |
| LAND        | `tracks.fused.land`       |
| CYBER       | `tracks.fused.cyber`      |

Message key: `track_id` (ensures all updates for a track go to same partition)

---

## 6. Metrics

| Metric                                       | Type      | Labels                  |
| -------------------------------------------- | --------- | ----------------------- |
| `rtsa_fusion_observations_processed_total`   | Counter   | `sensor_type`           |
| `rtsa_fusion_tracks_active`                  | Gauge     | `entity_type`, `status` |
| `rtsa_fusion_correlation_score`              | Histogram | `entity_type`, `action` |
| `rtsa_fusion_correlation_duration_seconds`   | Histogram | `entity_type`           |
| `rtsa_fusion_tracks_created_total`           | Counter   | `entity_type`           |
| `rtsa_fusion_tracks_merged_total`            | Counter   | `entity_type`           |
| `rtsa_fusion_tracks_dropped_total`           | Counter   | `entity_type`           |
| `rtsa_fusion_kalman_update_duration_seconds` | Histogram | -                       |

---

## 7. Test Scenarios

### 7.1 Unit Tests

| #   | Test                                          | Expected                     |
| --- | --------------------------------------------- | ---------------------------- |
| T01 | Haversine: known distance NYC→London          | ≈2998 NM (±1 NM)             |
| T02 | Haversine: same point                         | 0 NM                         |
| T03 | Gating: observation within gate               | In candidates                |
| T04 | Gating: observation outside distance gate     | Not in candidates            |
| T05 | Gating: observation outside time gate         | Not in candidates            |
| T06 | Gating: wrong entity type                     | Not in candidates            |
| T07 | Scorer: identical position+velocity+type+time | Score ≈ 1.0                  |
| T08 | Scorer: max distance, max time, wrong type    | Score < 0.60                 |
| T09 | Scorer: position match, velocity miss         | Score between 0.60-0.85      |
| T10 | Kalman predict: constant velocity 10s         | Position advanced correctly  |
| T11 | Kalman update: reduces uncertainty            | P diagonal decreases         |
| T12 | Track create: sets UUID v7                    | Valid UUID                   |
| T13 | Track create: initial Kalman state            | Position matches observation |
| T14 | Track update: source attribution              | Source list updated          |
| T15 | Track update: classification propagation      | MAX applied                  |
| T16 | Track merge: sources combined                 | All sources in merged track  |
| T17 | Track merge: source track marked MERGED       | Status = MERGED              |
| T18 | Stale monitor: marks stale after 60s          | Status transitions           |
| T19 | Stale monitor: drops after 5min               | Status = DROPPED             |
| T20 | Pipeline: new observation, no tracks          | New track created            |
| T21 | Pipeline: observation matches existing        | Track updated                |
| T22 | Pipeline: two tracks correlate ≥0.85          | Tracks merged                |

### 7.2 Integration Tests

| #    | Test                                    | Expected                                |
| ---- | --------------------------------------- | --------------------------------------- |
| IT01 | Produce radar obs → consume fused track | FusedTrack in tracks.fused.surface      |
| IT02 | Produce 2 correlated radar obs          | Single track with 2 source attributions |
| IT03 | Produce uncorrelated obs                | Two separate tracks                     |
| IT04 | Classification propagation              | MAX classification on fused track       |

---

## 8. Agent Invocation

```
@greatest-ever-developer Implement Module 07 from docs/implementation/07-fusion-engine.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for proto types (SensorObservation, FusedTrack)
- Read docs/implementation/03-shared-go-libraries.md for pkg/ interfaces
- Read docs/architecture/component_design.md §4 for the fusion component diagram
- This is the most algorithmically complex service
- Kalman filter: implement constant-velocity model (4-state: lat, lon, vN, vE)
- Use Haversine formula for distance calculations
- All thresholds must be configurable via environment variables
- Track state is in-memory only (no persistence — tracks are ephemeral)
- Consumer group: "fusion-engine"

Deliverables:
1. Complete svc-fusion-engine/ with all files
2. Gating filter with configurable per-entity-type thresholds
3. Correlation scorer with 4-component weighted scoring
4. Kalman filter with predict and update
5. Track manager with thread-safe operations
6. Stale monitor background goroutine
7. Unit tests for ALL domain logic (≥80% coverage)
8. Integration tests with testcontainers
```
