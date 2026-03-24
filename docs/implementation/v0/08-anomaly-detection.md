<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 08 — Anomaly Detection Service

> **Module**: 08-anomaly-detection
> **Phase**: P2 (Core Processing)
> **Dependencies**: Module 02 (protos), Module 03 (shared libraries), Module 07 (fusion engine produces to tracks.fused.\*)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 5 days

---

## 1. Objective

Implement the Anomaly Detection Service (`svc-anomaly-detection`) that consumes fused tracks from `tracks.fused.*` topics, extracts behavioral features, runs anomaly detection inference, generates severity-rated alerts, and produces them to `alerts.anomaly.{severity}` topics.

For MVP, the "inference engine" uses a rule-based/statistical approach (no ML model). A model-based approach is future work (Module 11 — Training Pipeline, which is stubbed).

**Acceptance Criteria**:

- Consumes from all 5 `tracks.fused.*` topics
- Detects all 6 anomaly types: Speed, Route Deviation, AIS Manipulation, Behavioral, Temporal, Proximity
- Maps anomaly confidence to severity: NORMAL (< 0.50), WATCH (0.50–0.69), ELEVATED (0.70–0.89), CRITICAL (≥ 0.90)
- Generates human-readable explanations for each alert
- Produces alerts to severity-specific topics
- Feature contributions included in every alert
- Model version tracking (rule-based = "rules-v1.0.0")
- ≥80% coverage on detection logic

---

## 2. Service Structure

```
svc-anomaly-detection/
├── cmd/
│   └── anomaly-detection/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── feature_extractor.go       # Extracts features from track state
│   │   ├── feature_extractor_test.go
│   │   ├── detectors/
│   │   │   ├── speed.go               # Speed anomaly detector
│   │   │   ├── speed_test.go
│   │   │   ├── route.go               # Route deviation detector
│   │   │   ├── route_test.go
│   │   │   ├── ais.go                 # AIS manipulation detector
│   │   │   ├── ais_test.go
│   │   │   ├── behavioral.go          # Behavioral pattern detector
│   │   │   ├── behavioral_test.go
│   │   │   ├── temporal.go            # Temporal anomaly detector
│   │   │   ├── temporal_test.go
│   │   │   ├── proximity.go           # Proximity alert detector
│   │   │   └── proximity_test.go
│   │   ├── severity.go                # Score → severity mapping
│   │   ├── severity_test.go
│   │   ├── explainer.go               # Human-readable explanation gen
│   │   └── explainer_test.go
│   ├── consumer/
│   │   ├── track_consumer.go
│   │   └── track_consumer_test.go
│   ├── producer/
│   │   ├── alert_producer.go
│   │   └── alert_producer_test.go
│   ├── handler/
│   │   ├── detection_pipeline.go
│   │   └── detection_pipeline_test.go
│   └── state/
│       ├── track_history.go           # Maintains recent track history for anomaly detection
│       └── track_history_test.go
├── go.mod
├── Dockerfile
└── README.md
```

---

## 3. Feature Extraction

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// FeatureVector contains all extracted features for anomaly detection.
type FeatureVector struct {
    // ── Speed Features ──
    CurrentSpeedKnots   float64
    AvgSpeed30Min       float64 // 30-minute rolling average
    SpeedStdDev         float64 // Standard deviation of speed over history
    SpeedDeltaSigma     float64 // (current - avg) / stddev

    // ── Route Features ──
    CurrentHeading      float64
    HeadingChangeRate   float64 // degrees/minute over last 5 updates
    ExpectedHeading     float64 // From linear regression of last 10 positions
    HeadingDeviation    float64 // |current - expected|

    // ── AIS Features (surface only) ──
    AISReportedPosition *Position // From AIS sensor
    FusedPosition       *Position // From fusion engine
    AISPositionDeltaNM  float64   // Haversine distance between AIS and fused
    HasAISSource        bool

    // ── Behavioral Features ──
    ActivityPattern     []float64 // Encoded activity sequence (last 20 states)
    PatternConfidence   float64   // How anomalous the pattern is (0-1)

    // ── Temporal Features ──
    HourOfDay           int
    DayOfWeek           int
    IsNighttime         bool
    TemporalPValue      float64 // p-value for time-of-day activity

    // ── Proximity Features ──
    NearestExclusionZoneDistNM float64 // Distance to nearest exclusion zone
    InExclusionZone            bool
    NearestCriticalAssetDistNM float64

    // ── Track Metadata ──
    TrackID     string
    EntityType  commonv1.EntityType
    SourceCount uint32
    TrackAgeMin float64
}

// FeatureExtractor builds FeatureVectors from track state and history.
type FeatureExtractor struct {
    history        *state.TrackHistory
    exclusionZones []ExclusionZone
}

// ExclusionZone defines a geographic area to monitor for proximity alerts.
type ExclusionZone struct {
    Name      string
    CenterLat float64
    CenterLon float64
    RadiusNM  float64
}

// Extract builds a FeatureVector for the given track update.
func (fe *FeatureExtractor) Extract(track *entityv1.FusedTrack) (*FeatureVector, error) { /* implementation */ }
```

---

## 4. Anomaly Detectors

### 4.1 Speed Anomaly Detector

```go
// CLASSIFICATION: UNCLASSIFIED
package detectors

// SpeedDetector detects speed anomalies where current speed exceeds
// 3 standard deviations from the track's historical mean speed.
type SpeedDetector struct {
    sigmaThreshold float64 // default: 3.0
}

// DetectionResult holds the output of a detector.
type DetectionResult struct {
    Detected   bool
    Confidence float64 // 0.0 to 1.0
    Features   []*inferencev1.FeatureContribution
}

// Detect checks for speed anomaly.
// Algorithm:
//   1. If SpeedStdDev < 0.1 (insufficient variance), skip detection → not detected
//   2. delta = |CurrentSpeed - AvgSpeed30Min|
//   3. sigma = delta / SpeedStdDev
//   4. If sigma > sigmaThreshold → Detected = true
//   5. Confidence = min(1.0, sigma / (sigmaThreshold * 2))
//   6. Features: speed_delta, speed_stddev, sigma_value
func (d *SpeedDetector) Detect(fv *FeatureVector) *DetectionResult { /* implementation */ }
```

### 4.2 Route Deviation Detector

```go
// CLASSIFICATION: UNCLASSIFIED
package detectors

// RouteDeviationDetector detects sustained heading changes > 30° from expected course.
type RouteDeviationDetector struct {
    deviationThreshold float64 // default: 30.0 degrees
    sustainedUpdates   int     // number of consecutive updates required (default: 3)
}

// Detect checks for route deviation.
// Algorithm:
//   1. Calculate HeadingDeviation = |CurrentHeading - ExpectedHeading|
//   2. Handle wraparound: if deviation > 180, deviation = 360 - deviation
//   3. If deviation > deviationThreshold for sustainedUpdates consecutive checks → Detected
//   4. Confidence = min(1.0, deviation / 90.0)
//   5. Features: heading_deviation, expected_heading, current_heading, heading_change_rate
func (d *RouteDeviationDetector) Detect(fv *FeatureVector) *DetectionResult { /* implementation */ }
```

### 4.3 AIS Manipulation Detector

```go
// CLASSIFICATION: UNCLASSIFIED
package detectors

// AISManipulationDetector detects discrepancies between AIS-reported
// position and fused (multi-sensor) position exceeding 0.5 NM.
type AISManipulationDetector struct {
    discrepancyThresholdNM float64 // default: 0.5
}

// Detect checks for AIS manipulation.
// Algorithm:
//   1. If !HasAISSource → skip (not detected)
//   2. If AISPositionDeltaNM > discrepancyThresholdNM → Detected
//   3. Confidence = min(1.0, AISPositionDeltaNM / (discrepancyThresholdNM * 3))
//   4. Features: ais_position_delta_nm, ais_lat, ais_lon, fused_lat, fused_lon
// Only applicable to SURFACE entity type.
func (d *AISManipulationDetector) Detect(fv *FeatureVector) *DetectionResult { /* implementation */ }
```

### 4.4 Behavioral Pattern Detector

```go
// CLASSIFICATION: UNCLASSIFIED
package detectors

// BehavioralDetector detects anomalous behavioral patterns
// based on activity sequence analysis.
// Uses statistical deviation scoring for MVP (no ML model).
type BehavioralDetector struct {
    confidenceThreshold float64 // default: 0.75
}

// Detect checks for behavioral anomalies.
// Algorithm (rule-based MVP):
//   1. Analyze ActivityPattern sequence
//   2. Look for suspicious patterns:
//      - Loitering (repeated positions within 0.1 NM over > 30 min)
//      - Zigzag pattern (> 5 heading reversals in 10 min)
//      - Speed pulsing (alternating fast/slow > 4 times in 20 min)
//   3. If PatternConfidence > confidenceThreshold → Detected
//   4. Confidence = PatternConfidence
//   5. Features: pattern_type, pattern_confidence, pattern_duration_min
func (d *BehavioralDetector) Detect(fv *FeatureVector) *DetectionResult { /* implementation */ }
```

### 4.5 Temporal Anomaly Detector

```go
// CLASSIFICATION: UNCLASSIFIED
package detectors

// TemporalDetector detects activity at unusual times for the entity's
// historical pattern (e.g., vessel movement at 3 AM in a normally idle port).
type TemporalDetector struct {
    pValueThreshold float64 // default: 0.05
}

// Detect checks for temporal anomalies.
// Algorithm:
//   1. If track age < 24 hours → skip (insufficient history)
//   2. TemporalPValue from feature extractor (computed via chi-squared test)
//   3. If TemporalPValue < pValueThreshold → Detected
//   4. Confidence = 1.0 - TemporalPValue
//   5. Features: hour_of_day, day_of_week, temporal_p_value
func (d *TemporalDetector) Detect(fv *FeatureVector) *DetectionResult { /* implementation */ }
```

### 4.6 Proximity Alert Detector

```go
// CLASSIFICATION: UNCLASSIFIED
package detectors

// ProximityDetector detects entities within exclusion zones.
type ProximityDetector struct{}

// Detect checks for proximity to exclusion zones.
// Algorithm:
//   1. If InExclusionZone → Detected, Confidence = 1.0
//   2. If NearestExclusionZoneDistNM < zone radius * 1.5 → Detected (approaching)
//   3. Confidence = 1.0 - (distance / (zone_radius * 2))
//   4. Features: nearest_zone_name, nearest_zone_distance_nm, in_exclusion_zone
func (d *ProximityDetector) Detect(fv *FeatureVector) *DetectionResult { /* implementation */ }
```

---

## 5. Severity Mapping

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// MapSeverity converts an anomaly confidence score to an AlertSeverity.
//   < 0.50 → NORMAL (no alert produced)
//   0.50 – 0.69 → WATCH
//   0.70 – 0.89 → ELEVATED
//   ≥ 0.90 → CRITICAL
func MapSeverity(confidence float64) commonv1.AlertSeverity { /* implementation */ }

// SeverityTopic returns the output topic for a given severity.
func SeverityTopic(severity commonv1.AlertSeverity) string {
    switch severity {
    case commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL: return "alerts.anomaly.critical"
    case commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED: return "alerts.anomaly.elevated"
    case commonv1.AlertSeverity_ALERT_SEVERITY_WATCH:    return "alerts.anomaly.watch"
    default: return ""
    }
}
```

---

## 6. Explanation Generator

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// GenerateExplanation creates a human-readable explanation for an anomaly alert.
// Template-based with variable substitution.
// Examples:
//   Speed: "Track T-123 (SURFACE) is moving at 45.2 knots, which is 4.2σ above the 30-minute average of 12.1 knots."
//   Route: "Track T-456 (AIR) has deviated 47° from its expected heading of 090° for the last 3 updates."
//   AIS: "Track T-789 (SURFACE) AIS position differs from fused position by 2.3 NM, indicating possible AIS spoofing."
//   Behavioral: "Track T-012 (SURFACE) is exhibiting a loitering pattern, remaining within 0.1 NM for 45 minutes."
//   Temporal: "Track T-345 (SURFACE) showing activity at 03:15 UTC, unusual for this area (p-value: 0.02)."
//   Proximity: "Track T-678 (SUBSURFACE) has entered exclusion zone 'Halifax Naval Base' (0.3 NM inside perimeter)."
func GenerateExplanation(anomalyType commonv1.AnomalyType, fv *FeatureVector, result *DetectionResult) string { /* implementation */ }
```

---

## 7. Track History State

```go
// CLASSIFICATION: UNCLASSIFIED
package state

// TrackHistory maintains a sliding window of recent track states
// needed for feature extraction (speed averages, heading trends, etc.).
type TrackHistory struct {
    mu       sync.RWMutex
    // history maps track_id → circular buffer of HistoryEntry
    history  map[string]*CircularBuffer
    maxEntries int // default: 100 entries per track
    maxAge     time.Duration // default: 2 hours
}

// HistoryEntry is a snapshot of a track at a point in time.
type HistoryEntry struct {
    Timestamp time.Time
    Latitude  float64
    Longitude float64
    SpeedKnots float64
    Heading   float64
    EntityType commonv1.EntityType
}

// Append adds a new entry to a track's history.
func (th *TrackHistory) Append(trackID string, entry *HistoryEntry) { /* implementation */ }

// GetHistory returns the history entries for a track within the time window.
func (th *TrackHistory) GetHistory(trackID string, window time.Duration) []*HistoryEntry { /* implementation */ }

// AvgSpeed calculates the average speed over the given window.
func (th *TrackHistory) AvgSpeed(trackID string, window time.Duration) float64 { /* implementation */ }

// SpeedStdDev calculates the standard deviation of speed.
func (th *TrackHistory) SpeedStdDev(trackID string, window time.Duration) float64 { /* implementation */ }

// RecentHeadings returns the last N headings for heading change analysis.
func (th *TrackHistory) RecentHeadings(trackID string, n int) []float64 { /* implementation */ }

// Cleanup removes entries older than maxAge across all tracks.
func (th *TrackHistory) Cleanup() { /* implementation */ }
```

---

## 8. Test Scenarios

### 8.1 Unit Tests

| #   | Test                                       | Expected                            |
| --- | ------------------------------------------ | ----------------------------------- |
| T01 | Speed: current = avg + 4σ                  | Detected, confidence > 0.5          |
| T02 | Speed: current = avg + 1σ                  | Not detected                        |
| T03 | Speed: insufficient history (stddev < 0.1) | Not detected                        |
| T04 | Route: 45° deviation sustained 3 updates   | Detected                            |
| T05 | Route: 15° deviation                       | Not detected (below 30°)            |
| T06 | Route: 45° deviation only 1 update         | Not detected (not sustained)        |
| T07 | AIS: 2.0 NM delta                          | Detected, high confidence           |
| T08 | AIS: 0.3 NM delta                          | Not detected (below 0.5 NM)         |
| T09 | AIS: no AIS source                         | Not detected (skipped)              |
| T10 | Behavioral: loitering 45 min               | Detected                            |
| T11 | Behavioral: normal transit                 | Not detected                        |
| T12 | Temporal: activity at p=0.01               | Detected                            |
| T13 | Temporal: activity at p=0.10               | Not detected                        |
| T14 | Temporal: track < 24h old                  | Not detected (insufficient history) |
| T15 | Proximity: inside exclusion zone           | Detected, confidence=1.0            |
| T16 | Proximity: 10 NM from zone                 | Not detected                        |
| T17 | Severity: 0.95 confidence                  | CRITICAL                            |
| T18 | Severity: 0.75 confidence                  | ELEVATED                            |
| T19 | Severity: 0.55 confidence                  | WATCH                               |
| T20 | Severity: 0.30 confidence                  | NORMAL (no alert)                   |
| T21 | Explanation: speed anomaly                 | Contains speed values and sigma     |
| T22 | Explanation: AIS manipulation              | Contains distance and "spoofing"    |
| T23 | Track history: avg speed                   | Correct average                     |
| T24 | Track history: cleanup old entries         | Old entries removed                 |

### 8.2 Integration Tests

| #    | Test                                    | Expected                           |
| ---- | --------------------------------------- | ---------------------------------- |
| IT01 | Publish fused track with speed anomaly  | Alert in alerts.anomaly.{severity} |
| IT02 | Publish normal fused track              | No alert produced                  |
| IT03 | Alert has correct feature contributions | Features populated                 |
| IT04 | Classification propagated from track    | Alert classification matches track |

---

## 9. Agent Invocation

```
@greatest-ever-developer Implement Module 08 from docs/implementation/08-anomaly-detection.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for AnomalyAlert proto
- Read docs/architecture/component_design.md §5 for anomaly detection component diagram
- Read docs/architecture/data_architecture.md for anomaly types and thresholds
- For MVP: use rule-based/statistical detectors, NOT ML models
- Model version for rule-based: "rules-v1.0.0"
- Each detector is independent and can be enabled/disabled via config
- Feature extraction maintains sliding window of track history
- Exclusion zones are configurable (env var or config file)

Deliverables:
1. Complete svc-anomaly-detection/ with all files
2. All 6 anomaly detectors implemented
3. Feature extractor with track history
4. Severity mapping and topic routing
5. Explanation generator with templates
6. Unit tests for each detector (≥80% coverage)
7. Integration tests with testcontainers
```
