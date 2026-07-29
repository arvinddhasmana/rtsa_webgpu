<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 02 — Protobuf Schemas & Code Generation

> **Module**: 02-protobuf-schemas
> **Phase**: P0 (Foundation)
> **Dependencies**: None
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 3 days

---

## 1. Objective

Define all Protobuf schemas for the RTSA system and configure buf for code generation. This module produces the source-of-truth `.proto` files, `buf.yaml`, `buf.gen.yaml`, and generates Go and TypeScript code under `gen/`.

**Acceptance Criteria**:

- `buf lint` passes with zero errors
- `buf generate` produces Go code in `gen/go/` and TypeScript code in `gen/ts/`
- All enum types have `_UNSPECIFIED = 0` as the first value
- All messages use field numbers 1–15 for high-frequency fields
- Every `.proto` file has a classification header comment
- Generated Go code compiles: `cd gen/go && go build ./...`

---

## 2. Project Configuration Files

### 2.1 `buf.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
version: v2
modules:
  - path: proto
    name: buf.build/arvinddhasmana/rtsa
lint:
  use:
    - STANDARD
  except:
    - PACKAGE_VERSION_SUFFIX
breaking:
  use:
    - FILE
```

### 2.2 `buf.gen.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
version: v2
plugins:
  # Go code generation
  - remote: buf.build/protocolbuffers/go
    out: gen/go
    opt:
      - paths=source_relative
  # Go gRPC code generation
  - remote: buf.build/grpc/go
    out: gen/go
    opt:
      - paths=source_relative
      - require_unimplemented_servers=false
  # TypeScript code generation (for COP Web App)
  - remote: buf.build/connectrpc/es
    out: gen/ts
    opt:
      - target=ts
```

### 2.3 `gen/go/go.mod`

```go
// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/rtsa_webgpu/gen/go

go 1.22.0

require (
    google.golang.org/grpc v1.62.0
    google.golang.org/protobuf v1.33.0
)
```

---

## 3. Proto File Structure

```
proto/
└── rtsa/
    ├── common/
    │   └── v1/
    │       ├── types.proto          # Shared types (Position, Velocity, enums)
    │       └── health.proto         # Health check service
    ├── ingestion/
    │   └── v1/
    │       ├── sensor_observation.proto  # All sensor data types
    │       └── ingestion_service.proto   # Ingestion gRPC service
    ├── entity/
    │   └── v1/
    │       ├── fused_track.proto    # FusedTrack, SourceAttribution
    │       └── track_service.proto  # Track streaming/query service
    ├── inference/
    │   └── v1/
    │       ├── anomaly_alert.proto  # AnomalyAlert, FeatureContribution
    │       └── alert_service.proto  # Alert streaming service
    ├── feedback/
    │   └── v1/
    │       ├── operator_feedback.proto  # OperatorFeedback, TrustBreakdown
    │       └── feedback_service.proto   # Feedback submission service
    ├── query/
    │   └── v1/
    │       └── query_service.proto  # Historical query service
    └── audit/
        └── v1/
            └── audit_event.proto    # AuditEvent and types
```

---

## 4. Proto File Specifications

### 4.1 `proto/rtsa/common/v1/types.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.common.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1;commonv1";

import "google/protobuf/timestamp.proto";

// ──────────────────────────────────────────
// Classification Levels — Government of Canada
// ──────────────────────────────────────────
enum ClassificationLevel {
  CLASSIFICATION_LEVEL_UNSPECIFIED = 0;
  CLASSIFICATION_LEVEL_UNCLASSIFIED = 1;
  CLASSIFICATION_LEVEL_PROTECTED_A = 2;
  CLASSIFICATION_LEVEL_PROTECTED_B = 3;
  CLASSIFICATION_LEVEL_PROTECTED_C = 4;
  CLASSIFICATION_LEVEL_SECRET = 5;
}

// ──────────────────────────────────────────
// Sensor Types
// ──────────────────────────────────────────
enum SensorType {
  SENSOR_TYPE_UNSPECIFIED = 0;
  SENSOR_TYPE_RADAR = 1;
  SENSOR_TYPE_EW_SIGINT = 2;
  SENSOR_TYPE_ELINT_COMINT = 3;
  SENSOR_TYPE_ISR = 4;
  SENSOR_TYPE_AIS_BFT = 5;
  SENSOR_TYPE_CYBER = 6;
  SENSOR_TYPE_NATO = 7;
}

// ──────────────────────────────────────────
// Entity Types
// ──────────────────────────────────────────
enum EntityType {
  ENTITY_TYPE_UNSPECIFIED = 0;
  ENTITY_TYPE_SURFACE = 1;
  ENTITY_TYPE_AIR = 2;
  ENTITY_TYPE_SUBSURFACE = 3;
  ENTITY_TYPE_LAND = 4;
  ENTITY_TYPE_CYBER = 5;
}

// ──────────────────────────────────────────
// Hostile Classification (IFF)
// ──────────────────────────────────────────
enum HostileClassification {
  HOSTILE_CLASSIFICATION_UNSPECIFIED = 0;
  HOSTILE_CLASSIFICATION_HOSTILE = 1;
  HOSTILE_CLASSIFICATION_FRIENDLY = 2;
  HOSTILE_CLASSIFICATION_NEUTRAL = 3;
  HOSTILE_CLASSIFICATION_UNKNOWN = 4;
  HOSTILE_CLASSIFICATION_SUSPECT = 5;
  HOSTILE_CLASSIFICATION_PENDING = 6;
}

// ──────────────────────────────────────────
// Track Status
// ──────────────────────────────────────────
enum TrackStatus {
  TRACK_STATUS_UNSPECIFIED = 0;
  TRACK_STATUS_NEW = 1;
  TRACK_STATUS_ACTIVE = 2;
  TRACK_STATUS_STALE = 3;
  TRACK_STATUS_DROPPED = 4;
  TRACK_STATUS_MERGED = 5;
}

// ──────────────────────────────────────────
// Anomaly Types
// ──────────────────────────────────────────
enum AnomalyType {
  ANOMALY_TYPE_UNSPECIFIED = 0;
  ANOMALY_TYPE_SPEED = 1;
  ANOMALY_TYPE_ROUTE_DEVIATION = 2;
  ANOMALY_TYPE_AIS_MANIPULATION = 3;
  ANOMALY_TYPE_BEHAVIORAL = 4;
  ANOMALY_TYPE_TEMPORAL = 5;
  ANOMALY_TYPE_PROXIMITY = 6;
}

// ──────────────────────────────────────────
// Alert Severity Levels
// ──────────────────────────────────────────
enum AlertSeverity {
  ALERT_SEVERITY_UNSPECIFIED = 0;
  ALERT_SEVERITY_NORMAL = 1;
  ALERT_SEVERITY_WATCH = 2;
  ALERT_SEVERITY_ELEVATED = 3;
  ALERT_SEVERITY_CRITICAL = 4;
}

// ──────────────────────────────────────────
// Feedback Types
// ──────────────────────────────────────────
enum FeedbackType {
  FEEDBACK_TYPE_UNSPECIFIED = 0;
  FEEDBACK_TYPE_CONFIRM_HOSTILE = 1;
  FEEDBACK_TYPE_CONFIRM_FRIENDLY = 2;
  FEEDBACK_TYPE_RECLASSIFY = 3;
  FEEDBACK_TYPE_REJECT_ANOMALY = 4;
  FEEDBACK_TYPE_CONFIRM_ANOMALY = 5;
}

// ──────────────────────────────────────────
// Geographic Position (WGS-84)
// ──────────────────────────────────────────
message Position {
  // Latitude in decimal degrees, WGS-84. Range: -90.0 to +90.0
  double latitude = 1;
  // Longitude in decimal degrees, WGS-84. Range: -180.0 to +180.0
  double longitude = 2;
  // Altitude above mean sea level in meters. Optional.
  optional double altitude_meters = 3;
  // Speed over ground in knots. Range: 0 to 2500 (air), 0 to 999 (surface)
  optional double speed_knots = 4;
  // Heading/course over ground in degrees true. Range: 0.0 to 360.0
  optional double heading_degrees = 5;
}

// ──────────────────────────────────────────
// Velocity Vector
// ──────────────────────────────────────────
message Velocity {
  // North component in m/s
  double north_mps = 1;
  // East component in m/s
  double east_mps = 2;
  // Down component in m/s (positive = descending)
  optional double down_mps = 3;
}

// ──────────────────────────────────────────
// Bounding Box for spatial queries
// ──────────────────────────────────────────
message BoundingBox {
  double min_latitude = 1;
  double max_latitude = 2;
  double min_longitude = 3;
  double max_longitude = 4;
}

// ──────────────────────────────────────────
// Time Range for queries
// ──────────────────────────────────────────
message TimeRange {
  google.protobuf.Timestamp start_time = 1;
  google.protobuf.Timestamp end_time = 2;
}

// ──────────────────────────────────────────
// Pagination
// ──────────────────────────────────────────
message PaginationRequest {
  int32 page_size = 1;     // Max items per page (default 100, max 1000)
  string page_token = 2;   // Opaque token from previous response
}

message PaginationResponse {
  string next_page_token = 1;  // Empty if no more pages
  int32 total_count = 2;       // Total matching records (if known)
}
```

### 4.2 `proto/rtsa/common/v1/health.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.common.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1;commonv1";

// Standard gRPC health check service
service HealthService {
  // Unary health check
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
  // Server-streaming health watch
  rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse);
}

message HealthCheckRequest {
  string service = 1;  // Service name to check (empty = overall)
}

message HealthCheckResponse {
  enum ServingStatus {
    SERVING_STATUS_UNSPECIFIED = 0;
    SERVING_STATUS_SERVING = 1;
    SERVING_STATUS_NOT_SERVING = 2;
    SERVING_STATUS_SERVICE_UNKNOWN = 3;
  }
  ServingStatus status = 1;
}
```

### 4.3 `proto/rtsa/ingestion/v1/sensor_observation.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.ingestion.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1;ingestionv1";

import "google/protobuf/timestamp.proto";
import "rtsa/common/v1/types.proto";

// ──────────────────────────────────────────
// Base Sensor Observation
// ──────────────────────────────────────────
message SensorObservation {
  // UUID v7 (time-ordered). Set by ingestion service.
  string observation_id = 1;
  // Source sensor identifier (e.g., "RADAR-HALIFAX-01")
  string sensor_id = 2;
  // Sensor type enum
  rtsa.common.v1.SensorType sensor_type = 3;
  // Time of observation at the sensor
  google.protobuf.Timestamp observation_time = 4;
  // Data classification level
  rtsa.common.v1.ClassificationLevel classification = 5;
  // Geographic position (optional — not all sensors provide position)
  rtsa.common.v1.Position position = 6;
  // Sensor-specific metadata as key-value pairs
  map<string, string> metadata = 7;
  // Sensor-specific payload — exactly one must be set
  oneof sensor_data {
    RadarTrack radar = 10;
    EWIntercept ew_sigint = 11;
    ELINTDetection elint_comint = 12;
    ISRObservation isr = 13;
    AISPosition ais_bft = 14;
    CyberIOC cyber = 15;
  }
}

// ──────────────────────────────────────────
// Radar Track Report
// ──────────────────────────────────────────
message RadarTrack {
  // Radar-assigned track number
  string track_number = 1;
  // Range from radar in nautical miles
  double range_nm = 2;
  // Bearing from radar in degrees true (0-360)
  double bearing_degrees = 3;
  // Radar cross-section in dBsm
  optional double radar_cross_section_dbsm = 4;
  // Target classification from radar (if available)
  optional rtsa.common.v1.EntityType entity_type = 5;
  // IFF mode code (if available)
  optional string iff_code = 6;
  // Signal-to-noise ratio in dB
  optional double snr_db = 7;
  // Track quality (0.0 to 1.0)
  double track_quality = 8;
}

// ──────────────────────────────────────────
// EW/SIGINT Intercept
// ──────────────────────────────────────────
message EWIntercept {
  // Emitter identifier from emitter library
  string emitter_id = 1;
  // Center frequency in MHz. Range: 0.5 to 40000
  double frequency_mhz = 2;
  // Bandwidth in MHz
  optional double bandwidth_mhz = 3;
  // Signal power in dBm
  optional double power_dbm = 4;
  // Bearing from receiver in degrees true (0-360)
  double bearing_degrees = 5;
  // Pulse repetition interval in microseconds
  optional double pri_microseconds = 6;
  // Modulation type
  string modulation_type = 7;
  // Threat assessment level
  optional string threat_level = 8;
  // Intercept confidence (0.0 to 1.0)
  double confidence = 9;
}

// ──────────────────────────────────────────
// ELINT/COMINT Detection
// ──────────────────────────────────────────
message ELINTDetection {
  // Emitter identifier
  string emitter_id = 1;
  // Radar type from emitter library (e.g., "SA-11 FIRE CONTROL")
  string radar_type = 2;
  // Center frequency in MHz
  double frequency_mhz = 3;
  // Pulse width in microseconds
  optional double pulse_width_us = 4;
  // Scan type (circular, sector, track-while-scan)
  string scan_type = 5;
  // Circular error probable in meters (geolocation accuracy)
  double cep_meters = 6;
  // Detection confidence (0.0 to 1.0)
  double confidence = 7;
  // Content classification may differ from metadata classification
  rtsa.common.v1.ClassificationLevel content_classification = 8;
}

// ──────────────────────────────────────────
// ISR Observation
// ──────────────────────────────────────────
message ISRObservation {
  // ISR platform identifier
  string platform_id = 1;
  // Sensor type on platform (EO, IR, SAR, MTI)
  string sensor_name = 2;
  // Reference to imagery product (no raw imagery in stream)
  string image_id = 3;
  // Coverage area (polygon vertices as lat,lon pairs)
  repeated rtsa.common.v1.Position coverage_polygon = 4;
  // Ground sample distance in meters (resolution)
  optional double gsd_meters = 5;
  // Detection reports within this observation
  repeated ISRDetection detections = 6;
}

message ISRDetection {
  // Detection position
  rtsa.common.v1.Position position = 1;
  // Detected entity type
  rtsa.common.v1.EntityType entity_type = 2;
  // Detection confidence (0.0 to 1.0)
  double confidence = 3;
  // Description of detection
  string description = 4;
}

// ──────────────────────────────────────────
// AIS Position Report
// ──────────────────────────────────────────
message AISPosition {
  // Maritime Mobile Service Identity (9 digits)
  string mmsi = 1;
  // IMO number (optional)
  optional string imo_number = 2;
  // Vessel name
  string vessel_name = 3;
  // AIS vessel type code (1-99)
  int32 vessel_type_code = 4;
  // Call sign
  optional string call_sign = 5;
  // Navigation status (per AIS spec)
  string nav_status = 6;
  // Rate of turn in degrees/min
  optional double rate_of_turn = 7;
  // Draught in meters
  optional double draught_meters = 8;
  // Destination port
  optional string destination = 9;
  // Estimated time of arrival
  optional google.protobuf.Timestamp eta = 10;
  // Whether this is a BFT position (classified PROTECTED)
  bool is_bft = 11;
  // AIS message type (1, 2, 3, 5, 18, 24)
  int32 ais_message_type = 12;
}

// ──────────────────────────────────────────
// Cyber Threat Indicator (IOC)
// ──────────────────────────────────────────
message CyberIOC {
  // STIX indicator ID
  string stix_id = 1;
  // IOC type (ipv4-addr, domain-name, file:hashes, url)
  string ioc_type = 2;
  // IOC value
  string ioc_value = 3;
  // MITRE ATT&CK technique IDs
  repeated string mitre_attack_ids = 4;
  // Threat actor attribution (if known)
  optional string threat_actor = 5;
  // Malware family (if known)
  optional string malware_family = 6;
  // IOC confidence score (0.0 to 1.0)
  double confidence = 7;
  // STIX valid-from timestamp
  google.protobuf.Timestamp valid_from = 8;
  // STIX valid-until timestamp (optional)
  optional google.protobuf.Timestamp valid_until = 9;
  // Source feed name
  string source_feed = 10;
  // SHA-256 hash for deduplication
  string dedup_hash = 11;
}
```

### 4.4 `proto/rtsa/ingestion/v1/ingestion_service.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.ingestion.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1;ingestionv1";

import "rtsa/ingestion/v1/sensor_observation.proto";
import "rtsa/common/v1/types.proto";
import "google/protobuf/timestamp.proto";

// ──────────────────────────────────────────
// Ingestion Service — shared by all sensor types
// ──────────────────────────────────────────
service IngestionService {
  // Client-streaming: bulk sensor data ingestion
  // Deadline: 60s per stream session
  rpc IngestSensorData(stream SensorObservation) returns (IngestSummary);

  // Unary: single observation ingestion
  // Deadline: 5s
  rpc IngestSingleObservation(SensorObservation) returns (IngestionAck);

  // Unary: query sensor status
  // Deadline: 5s
  rpc GetSensorStatus(GetSensorStatusRequest) returns (SensorStatusResponse);
}

message IngestionAck {
  string observation_id = 1;
  bool accepted = 2;
  string rejection_reason = 3;
}

message IngestSummary {
  int64 total_received = 1;
  int64 accepted = 2;
  int64 rejected = 3;
  repeated RejectionDetail rejections = 4;
}

message RejectionDetail {
  string observation_id = 1;
  string reason = 2;
}

message GetSensorStatusRequest {
  string sensor_id = 1;
}

message SensorStatusResponse {
  string sensor_id = 1;
  rtsa.common.v1.SensorType sensor_type = 2;
  bool connected = 3;
  int64 total_received = 4;
  int64 total_accepted = 5;
  int64 total_rejected = 6;
  google.protobuf.Timestamp last_observation_time = 7;
  double events_per_second = 8;
}
```

### 4.5 `proto/rtsa/entity/v1/fused_track.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.entity.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1;entityv1";

import "google/protobuf/timestamp.proto";
import "rtsa/common/v1/types.proto";

// ──────────────────────────────────────────
// Fused Track — multi-source correlated entity
// ──────────────────────────────────────────
message FusedTrack {
  // UUID v7 (time-ordered)
  string track_id = 1;
  // Entity type (surface, air, subsurface, land, cyber)
  rtsa.common.v1.EntityType entity_type = 2;
  // IFF classification
  rtsa.common.v1.HostileClassification hostile_class = 3;
  // Kalman-filtered position estimate
  rtsa.common.v1.Position estimated_position = 4;
  // Overall confidence in the track (0.0 to 1.0)
  double confidence_score = 5;
  // Number of contributing sensors
  uint32 source_count = 6;
  // Attribution per contributing sensor
  repeated SourceAttribution sources = 7;
  // Track lifecycle status
  rtsa.common.v1.TrackStatus status = 8;
  // Data classification (MAX of all contributing sources)
  rtsa.common.v1.ClassificationLevel classification = 9;
  // Track creation time
  google.protobuf.Timestamp created_at = 10;
  // Last update time
  google.protobuf.Timestamp updated_at = 11;
  // Velocity estimate
  rtsa.common.v1.Velocity velocity = 12;
  // Kinematic state uncertainty (covariance diagonal)
  optional KinematicUncertainty uncertainty = 13;
  // Human-assigned label (if any)
  optional string label = 14;
  // Track age in seconds since creation
  double age_seconds = 15;
}

message SourceAttribution {
  // Sensor identifier
  string sensor_id = 1;
  // Sensor type
  rtsa.common.v1.SensorType sensor_type = 2;
  // Contribution confidence
  double confidence = 3;
  // Time of last contribution from this sensor
  google.protobuf.Timestamp last_contribution = 4;
  // Number of observations from this sensor
  uint32 observation_count = 5;
}

message KinematicUncertainty {
  // Position uncertainty semi-major axis in meters
  double position_error_meters = 1;
  // Velocity uncertainty in m/s
  double velocity_error_mps = 2;
  // Heading uncertainty in degrees
  double heading_error_degrees = 3;
}
```

### 4.6 `proto/rtsa/entity/v1/track_service.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.entity.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1;entityv1";

import "rtsa/entity/v1/fused_track.proto";
import "rtsa/common/v1/types.proto";
import "google/protobuf/timestamp.proto";

// ──────────────────────────────────────────
// Track Service — real-time track streaming
// ──────────────────────────────────────────
service TrackService {
  // Server-streaming: receive real-time track updates
  // Sends initial snapshot then incremental updates
  // Deadline: none (long-lived stream)
  rpc StreamTracks(StreamTracksRequest) returns (stream TrackUpdate);

  // Unary: get details of a specific track
  // Deadline: 5s
  rpc GetTrackDetails(GetTrackDetailsRequest) returns (FusedTrack);

  // Unary: get recent position history for a track
  // Deadline: 10s
  rpc GetTrackHistory(GetTrackHistoryRequest) returns (TrackHistoryResponse);
}

message StreamTracksRequest {
  // Filter by entity types (empty = all)
  repeated rtsa.common.v1.EntityType entity_types = 1;
  // Filter by hostile classification (empty = all)
  repeated rtsa.common.v1.HostileClassification hostile_classes = 2;
  // Filter by geographic bounding box (empty = global)
  optional rtsa.common.v1.BoundingBox bounding_box = 3;
  // Filter by minimum confidence score (default: 0.0)
  double min_confidence = 4;
  // Caller's classification clearance level
  rtsa.common.v1.ClassificationLevel clearance_level = 5;
}

message TrackUpdate {
  enum UpdateType {
    UPDATE_TYPE_UNSPECIFIED = 0;
    UPDATE_TYPE_SNAPSHOT = 1;    // Initial full state
    UPDATE_TYPE_CREATED = 2;    // New track
    UPDATE_TYPE_UPDATED = 3;    // Position/state change
    UPDATE_TYPE_DROPPED = 4;    // Track dropped
    UPDATE_TYPE_MERGED = 5;     // Track merged into another
  }
  UpdateType update_type = 1;
  FusedTrack track = 2;
  // For MERGED: the target track ID that this track merged into
  optional string merged_into_track_id = 3;
}

message GetTrackDetailsRequest {
  string track_id = 1;
  rtsa.common.v1.ClassificationLevel clearance_level = 2;
}

message GetTrackHistoryRequest {
  string track_id = 1;
  rtsa.common.v1.TimeRange time_range = 2;
  int32 max_points = 3;  // Maximum history points to return (default: 100)
  rtsa.common.v1.ClassificationLevel clearance_level = 4;
}

message TrackHistoryResponse {
  string track_id = 1;
  repeated TrackHistoryPoint points = 2;
}

message TrackHistoryPoint {
  rtsa.common.v1.Position position = 1;
  google.protobuf.Timestamp timestamp = 2;
  double confidence = 3;
  rtsa.common.v1.TrackStatus status = 4;
}
```

### 4.7 `proto/rtsa/inference/v1/anomaly_alert.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.inference.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1;inferencev1";

import "google/protobuf/timestamp.proto";
import "rtsa/common/v1/types.proto";

// ──────────────────────────────────────────
// Anomaly Alert — detection result
// ──────────────────────────────────────────
message AnomalyAlert {
  // UUID v7 (time-ordered)
  string alert_id = 1;
  // Reference to the fused track that triggered this alert
  string track_id = 2;
  // Type of anomaly detected
  rtsa.common.v1.AnomalyType anomaly_type = 3;
  // Severity level based on confidence thresholds
  rtsa.common.v1.AlertSeverity severity = 4;
  // Anomaly confidence score (0.0 to 1.0)
  double confidence_score = 5;
  // Human-readable explanation of why this alert was raised
  string explanation = 6;
  // Contributing feature values
  repeated FeatureContribution features = 7;
  // Data classification (inherited from track)
  rtsa.common.v1.ClassificationLevel classification = 8;
  // Time of detection
  google.protobuf.Timestamp detected_at = 9;
  // Model version that produced this alert
  string model_version = 10;
  // Track position at time of detection
  rtsa.common.v1.Position track_position = 11;
  // Whether this alert has been acknowledged by an operator
  bool acknowledged = 12;
  // Track entity type (denormalized for filtering)
  rtsa.common.v1.EntityType entity_type = 13;
}

message FeatureContribution {
  // Feature name (e.g., "speed_delta", "heading_change_rate")
  string feature_name = 1;
  // Raw feature value
  double value = 2;
  // Weight of this feature in the anomaly score
  double contribution_weight = 3;
  // Baseline/expected value for comparison
  optional double baseline_value = 4;
  // Feature description for explanation
  string description = 5;
}
```

### 4.8 `proto/rtsa/inference/v1/alert_service.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.inference.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1;inferencev1";

import "rtsa/inference/v1/anomaly_alert.proto";
import "rtsa/common/v1/types.proto";
import "google/protobuf/timestamp.proto";

// ──────────────────────────────────────────
// Alert Service — real-time alert streaming
// ──────────────────────────────────────────
service AlertService {
  // Server-streaming: receive real-time alerts
  // Delivers in priority order (CRITICAL first)
  // Deadline: none (long-lived stream)
  rpc StreamAlerts(StreamAlertsRequest) returns (stream AnomalyAlert);

  // Unary: acknowledge an alert
  // Deadline: 5s
  rpc AcknowledgeAlert(AcknowledgeAlertRequest) returns (AcknowledgeAlertResponse);

  // Unary: get details of a specific alert
  // Deadline: 5s
  rpc GetAlertDetails(GetAlertDetailsRequest) returns (AnomalyAlert);
}

message StreamAlertsRequest {
  // Minimum severity to receive (default: WATCH)
  rtsa.common.v1.AlertSeverity min_severity = 1;
  // Filter by anomaly types (empty = all)
  repeated rtsa.common.v1.AnomalyType anomaly_types = 2;
  // Filter by entity types (empty = all)
  repeated rtsa.common.v1.EntityType entity_types = 3;
  // Caller's classification clearance level
  rtsa.common.v1.ClassificationLevel clearance_level = 4;
}

message AcknowledgeAlertRequest {
  string alert_id = 1;
  string operator_id = 2;
  string comment = 3;
}

message AcknowledgeAlertResponse {
  bool success = 1;
  google.protobuf.Timestamp acknowledged_at = 2;
}

message GetAlertDetailsRequest {
  string alert_id = 1;
  rtsa.common.v1.ClassificationLevel clearance_level = 2;
}
```

### 4.9 `proto/rtsa/feedback/v1/operator_feedback.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.feedback.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1;feedbackv1";

import "google/protobuf/timestamp.proto";
import "rtsa/common/v1/types.proto";

// ──────────────────────────────────────────
// Operator Feedback — human-in-the-loop
// ──────────────────────────────────────────
message OperatorFeedback {
  // UUID v7 (time-ordered)
  string feedback_id = 1;
  // Track being evaluated
  string track_id = 2;
  // Operator submitting feedback (from mTLS cert in production)
  string operator_id = 3;
  // Type of feedback
  rtsa.common.v1.FeedbackType feedback_type = 4;
  // Operator's justification
  string justification = 5;
  // Computed trust score (0.0 to 1.0)
  double trust_score = 6;
  // Component breakdown of trust score
  TrustBreakdown trust_breakdown = 7;
  // Data classification
  rtsa.common.v1.ClassificationLevel classification = 8;
  // When feedback was submitted
  google.protobuf.Timestamp submitted_at = 9;
  // Associated alert ID (if feedback is about an anomaly)
  optional string alert_id = 10;
  // New hostile classification (for RECLASSIFY type)
  optional rtsa.common.v1.HostileClassification new_hostile_class = 11;
  // Operator's clearance level
  rtsa.common.v1.ClassificationLevel operator_clearance = 12;
}

message TrustBreakdown {
  // Clearance component: mapped from operator clearance level (0.0 to 1.0)
  // SECRET=1.0, PROTECTED_C=0.85, PROTECTED_B=0.7, PROTECTED_A=0.5, UNCLASSIFIED=0.3
  double clearance_score = 1;
  // Accuracy component: historical accuracy ratio (0.0 to 1.0)
  // Default 0.5 for new operators
  double accuracy_score = 2;
  // Temporal component: decay function based on time since event
  // 1.0 within 5min, 0.5 at 30min, 0.1 at 2hr
  double temporal_score = 3;
  // Deviation component: 1 - divergence from consensus (0.0 to 1.0)
  // 1.0 if matches majority, 0.0 if contradicts all
  double deviation_score = 4;
}
```

### 4.10 `proto/rtsa/feedback/v1/feedback_service.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.feedback.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1;feedbackv1";

import "rtsa/feedback/v1/operator_feedback.proto";
import "rtsa/common/v1/types.proto";
import "google/protobuf/timestamp.proto";

// ──────────────────────────────────────────
// Feedback Service — operator feedback collection
// ──────────────────────────────────────────
service FeedbackService {
  // Unary: submit operator feedback
  // Deadline: 10s
  rpc SubmitFeedback(SubmitFeedbackRequest) returns (SubmitFeedbackResponse);

  // Unary: query feedback history
  // Deadline: 15s
  rpc GetFeedbackHistory(GetFeedbackHistoryRequest) returns (GetFeedbackHistoryResponse);
}

message SubmitFeedbackRequest {
  string track_id = 1;
  string operator_id = 2;
  rtsa.common.v1.FeedbackType feedback_type = 3;
  string justification = 4;
  optional string alert_id = 5;
  optional rtsa.common.v1.HostileClassification new_hostile_class = 6;
  rtsa.common.v1.ClassificationLevel operator_clearance = 7;
}

message SubmitFeedbackResponse {
  string feedback_id = 1;
  double trust_score = 2;
  TrustBreakdown trust_breakdown = 3;
  bool validated = 4;  // true if trust_score >= 0.5
  google.protobuf.Timestamp submitted_at = 5;
}

message GetFeedbackHistoryRequest {
  // Filter by track ID (optional)
  optional string track_id = 1;
  // Filter by operator ID (optional)
  optional string operator_id = 2;
  // Time range
  rtsa.common.v1.TimeRange time_range = 3;
  // Pagination
  rtsa.common.v1.PaginationRequest pagination = 4;
}

message GetFeedbackHistoryResponse {
  repeated OperatorFeedback feedback = 1;
  rtsa.common.v1.PaginationResponse pagination = 2;
}
```

### 4.11 `proto/rtsa/query/v1/query_service.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.query.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/query/v1;queryv1";

import "rtsa/common/v1/types.proto";
import "rtsa/entity/v1/fused_track.proto";
import "rtsa/inference/v1/anomaly_alert.proto";
import "rtsa/feedback/v1/operator_feedback.proto";

// ──────────────────────────────────────────
// Query Service — historical ClickHouse queries
// ──────────────────────────────────────────
service QueryService {
  // Unary: query historical track data
  // Deadline: 30s
  // Guardrails: max 30-day range, max 100K rows, parameterized only
  rpc QueryTracks(QueryTracksRequest) returns (QueryTracksResponse);

  // Unary: query anomaly detection history
  // Deadline: 30s
  rpc QueryAnomalies(QueryAnomaliesRequest) returns (QueryAnomaliesResponse);

  // Unary: query audit log
  // Deadline: 30s
  rpc QueryAuditLog(QueryAuditLogRequest) returns (QueryAuditLogResponse);
}

// ──── Track Queries ────────────────────────

message QueryTracksRequest {
  rtsa.common.v1.TimeRange time_range = 1;
  // Filter by entity types (empty = all)
  repeated rtsa.common.v1.EntityType entity_types = 2;
  // Filter by hostile classification (empty = all)
  repeated rtsa.common.v1.HostileClassification hostile_classes = 3;
  // Spatial filter (empty = global)
  optional rtsa.common.v1.BoundingBox bounding_box = 4;
  // Minimum confidence
  double min_confidence = 5;
  // Track ID filter (empty = all)
  optional string track_id = 6;
  // Pagination
  rtsa.common.v1.PaginationRequest pagination = 7;
  // Caller's classification clearance
  rtsa.common.v1.ClassificationLevel clearance_level = 8;
}

message QueryTracksResponse {
  repeated rtsa.entity.v1.FusedTrack tracks = 1;
  rtsa.common.v1.PaginationResponse pagination = 2;
}

// ──── Anomaly Queries ──────────────────────

message QueryAnomaliesRequest {
  rtsa.common.v1.TimeRange time_range = 1;
  // Filter by anomaly types (empty = all)
  repeated rtsa.common.v1.AnomalyType anomaly_types = 2;
  // Filter by severity (empty = all)
  repeated rtsa.common.v1.AlertSeverity severities = 3;
  // Filter by track ID (optional)
  optional string track_id = 4;
  // Minimum confidence
  double min_confidence = 5;
  // Pagination
  rtsa.common.v1.PaginationRequest pagination = 6;
  // Caller's classification clearance
  rtsa.common.v1.ClassificationLevel clearance_level = 7;
}

message QueryAnomaliesResponse {
  repeated rtsa.inference.v1.AnomalyAlert alerts = 1;
  rtsa.common.v1.PaginationResponse pagination = 2;
}

// ──── Audit Queries ────────────────────────

message QueryAuditLogRequest {
  rtsa.common.v1.TimeRange time_range = 1;
  // Filter by service name (empty = all)
  optional string service_id = 2;
  // Filter by event type (empty = all)
  optional string event_type = 3;
  // Filter by actor (empty = all)
  optional string actor_id = 4;
  // Filter by resource type (empty = all)
  optional string resource_type = 5;
  // Pagination
  rtsa.common.v1.PaginationRequest pagination = 6;
  // Caller's classification clearance
  rtsa.common.v1.ClassificationLevel clearance_level = 7;
}

message QueryAuditLogResponse {
  repeated AuditLogEntry entries = 1;
  rtsa.common.v1.PaginationResponse pagination = 2;
}

message AuditLogEntry {
  string audit_id = 1;
  string service_id = 2;
  string event_type = 3;
  string actor_id = 4;
  string actor_type = 5;
  string resource_type = 6;
  string resource_id = 7;
  string action = 8;
  string detail_json = 9;
  rtsa.common.v1.ClassificationLevel classification_level = 10;
  string event_time = 11;
}
```

### 4.12 `proto/rtsa/audit/v1/audit_event.proto`

```protobuf
// CLASSIFICATION: UNCLASSIFIED
syntax = "proto3";

package rtsa.audit.v1;

option go_package = "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1;auditv1";

import "google/protobuf/timestamp.proto";
import "rtsa/common/v1/types.proto";

// ──────────────────────────────────────────
// Audit Event — immutable audit trail
// ──────────────────────────────────────────
message AuditEvent {
  // UUID v7 (time-ordered)
  string audit_id = 1;
  // Service that generated this event
  string service_id = 2;
  // Event type categorization
  string event_type = 3;
  // Actor: service name or operator ID
  string actor_id = 4;
  // Actor type
  ActorType actor_type = 5;
  // Resource being acted upon
  string resource_type = 6;
  // Resource identifier
  string resource_id = 7;
  // Action performed
  AuditAction action = 8;
  // JSON-encoded additional context
  string detail_json = 9;
  // Classification level of the audited operation
  rtsa.common.v1.ClassificationLevel classification_level = 10;
  // Event timestamp
  google.protobuf.Timestamp event_time = 11;
  // Correlation/trace ID for distributed tracing
  string trace_id = 12;
}

enum ActorType {
  ACTOR_TYPE_UNSPECIFIED = 0;
  ACTOR_TYPE_SERVICE = 1;
  ACTOR_TYPE_OPERATOR = 2;
  ACTOR_TYPE_SYSTEM = 3;
}

enum AuditAction {
  AUDIT_ACTION_UNSPECIFIED = 0;
  AUDIT_ACTION_CREATE = 1;
  AUDIT_ACTION_READ = 2;
  AUDIT_ACTION_UPDATE = 3;
  AUDIT_ACTION_DELETE = 4;
  AUDIT_ACTION_QUERY = 5;
  AUDIT_ACTION_INGEST = 6;
  AUDIT_ACTION_EXPORT = 7;
  AUDIT_ACTION_AUTHENTICATE = 8;
  AUDIT_ACTION_AUTHORIZE = 9;
  AUDIT_ACTION_CLASSIFY = 10;
  AUDIT_ACTION_FEEDBACK = 11;
}
```

---

## 5. Code Generation Output

After `buf generate`, the following directory structure is produced:

```
gen/
├── go/
│   └── rtsa/
│       ├── common/v1/
│       │   ├── types.pb.go
│       │   ├── health.pb.go
│       │   └── health_grpc.pb.go
│       ├── ingestion/v1/
│       │   ├── sensor_observation.pb.go
│       │   ├── ingestion_service.pb.go
│       │   └── ingestion_service_grpc.pb.go
│       ├── entity/v1/
│       │   ├── fused_track.pb.go
│       │   ├── track_service.pb.go
│       │   └── track_service_grpc.pb.go
│       ├── inference/v1/
│       │   ├── anomaly_alert.pb.go
│       │   ├── alert_service.pb.go
│       │   └── alert_service_grpc.pb.go
│       ├── feedback/v1/
│       │   ├── operator_feedback.pb.go
│       │   ├── feedback_service.pb.go
│       │   └── feedback_service_grpc.pb.go
│       ├── query/v1/
│       │   ├── query_service.pb.go
│       │   └── query_service_grpc.pb.go
│       └── audit/v1/
│           └── audit_event.pb.go
└── ts/
    └── rtsa/
        ├── common/v1/
        ├── entity/v1/
        ├── inference/v1/
        ├── feedback/v1/
        └── query/v1/
```

---

## 6. Validation Rules Reference

These field validation rules are embedded in proto comments and must be enforced at ingestion time (Module 04/05):

| Field                      | Type      | Range                                            | On Failure     |
| -------------------------- | --------- | ------------------------------------------------ | -------------- |
| `position.latitude`        | double    | -90.0 to +90.0                                   | Reject to DLQ  |
| `position.longitude`       | double    | -180.0 to +180.0                                 | Reject to DLQ  |
| `position.speed_knots`     | double    | 0 to 999 (surface), 0 to 2500 (air)              | Flag suspect   |
| `position.heading_degrees` | double    | 0.0 to 360.0                                     | Reject to DLQ  |
| `observation_time`         | Timestamp | Not > 5 min future, not > 24h past               | Reject to DLQ  |
| `classification`           | enum      | Valid enum value                                 | Reject to DLQ  |
| `sensor_id`                | string    | Non-empty, max 128 chars                         | Reject to DLQ  |
| `observation_id`           | string    | Valid UUID v7 format                             | Set by service |
| `ew.frequency_mhz`         | double    | 0.5 to 40000                                     | Reject to DLQ  |
| `ew.bearing_degrees`       | double    | 0.0 to 360.0                                     | Reject to DLQ  |
| `ais.mmsi`                 | string    | Exactly 9 digits                                 | Reject to DLQ  |
| `ais.vessel_type_code`     | int32     | 1 to 99                                          | Reject to DLQ  |
| `elint.cep_meters`         | double    | > 0                                              | Reject to DLQ  |
| `cyber.ioc_type`           | string    | One of: ipv4-addr, domain-name, file:hashes, url | Reject to DLQ  |

---

## 7. Test Scenarios

| #   | Scenario                                               | Expected                          |
| --- | ------------------------------------------------------ | --------------------------------- |
| T01 | `buf lint`                                             | Zero errors                       |
| T02 | `buf generate`                                         | All Go and TS files generated     |
| T03 | `cd gen/go && go build ./...`                          | Generated Go code compiles        |
| T04 | All enums have `_UNSPECIFIED = 0`                      | Verified via script               |
| T05 | `buf breaking --against '.git#branch=main'`            | No breaking changes (baseline)    |
| T06 | Field numbers 1-15 used for high-frequency fields      | Verified manually in all messages |
| T07 | All `.proto` files have classification header          | Verified via grep                 |
| T08 | All `go_package` options reference correct module path | Verified                          |

---

## 8. Agent Invocation

```
@greatest-ever-developer Implement Module 02 from docs/implementation/02-protobuf-schemas.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- This module defines ALL protobuf schemas for the RTSA system
- Create proto files EXACTLY as specified in the document
- Create buf.yaml, buf.gen.yaml at project root
- Create gen/go/go.mod for the generated Go module
- Run buf lint and buf generate after creating all files
- Verify generated Go code compiles

Deliverables:
1. buf.yaml and buf.gen.yaml at project root
2. All .proto files under proto/rtsa/
3. gen/go/go.mod
4. buf lint passes
5. buf generate produces code in gen/go/ and gen/ts/
6. Generated Go code compiles
```
