<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 05 — Sensor Ingestion: Batch (EW, ELINT, ISR, AIS, Cyber)

> **Module**: 05-sensor-ingestion-batch
> **Phase**: P1 (Ingestion)
> **Dependencies**: Module 02 (protos), Module 03 (shared libraries), Module 04 (reference implementation)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 4 days

---

## 1. Objective

Implement the remaining 5 sensor ingestion services by cloning the Module 04 reference pattern (Radar). Each service follows IDENTICAL structure to `svc-radar-ingestion` with sensor-specific validators and normalizers.

**Services to implement**:

| #   | Service                | Go Module             | Output Topic               | DLQ Topic           | gRPC Port |
| --- | ---------------------- | --------------------- | -------------------------- | ------------------- | --------- |
| 1   | EW/SIGINT Ingestion    | `svc-ew-ingestion`    | `sensors.ew.intercepts`    | `dlq.sensors.ew`    | 50052     |
| 2   | ELINT/COMINT Ingestion | `svc-elint-ingestion` | `sensors.elint.detections` | `dlq.sensors.elint` | 50053     |
| 3   | ISR Metadata Ingestion | `svc-isr-ingestion`   | `sensors.isr.observations` | `dlq.sensors.isr`   | 50054     |
| 4   | AIS/BFT Ingestion      | `svc-ais-ingestion`   | `sensors.ais.positions`    | `dlq.sensors.ais`   | 50055     |
| 5   | Cyber Threat Ingestion | `svc-cyber-ingestion` | `sensors.cyber.iocs`       | `dlq.sensors.cyber` | 50056     |

---

## 2. Shared Pattern (From Module 04)

Each service has IDENTICAL structure:

```
svc-<sensor>-ingestion/
├── cmd/<sensor>-ingestion/main.go
├── internal/
│   ├── config/config.go      # Sensor-specific env vars
│   ├── domain/
│   │   ├── validator.go      # SENSOR-SPECIFIC validation
│   │   ├── validator_test.go
│   │   ├── normalizer.go     # SENSOR-SPECIFIC normalization
│   │   └── normalizer_test.go
│   ├── handler/
│   │   ├── ingestion.go      # IDENTICAL to radar (uses interfaces)
│   │   └── ingestion_test.go
│   ├── producer/
│   │   ├── observation.go    # IDENTICAL to radar
│   │   └── observation_test.go
│   └── mapper/
│       ├── enricher.go       # IDENTICAL to radar
│       └── enricher_test.go
├── go.mod
├── Dockerfile
└── README.md
```

**IMPORTANT**: The handler, producer, and enricher are IDENTICAL across all 6 ingestion services. Only `validator.go` and `normalizer.go` differ. Consider extracting the handler/producer/enricher into `pkg/ingestion` if code duplication exceeds 50 lines.

---

## 3. Sensor-Specific Validation Rules

### 3.1 EW/SIGINT Validator

```go
// CLASSIFICATION: UNCLASSIFIED
// Validate checks EW/SIGINT-specific rules:
//
// REJECT TO DLQ:
//   - sensor_id: non-empty, max 128 chars
//   - sensor_type: must be SENSOR_TYPE_EW_SIGINT
//   - observation_time: not > 5min future, not > 24h past
//   - classification: valid enum (not UNSPECIFIED)
//   - ew_sigint.emitter_id: non-empty
//   - ew_sigint.frequency_mhz: 0.5 to 40000
//   - ew_sigint.bearing_degrees: 0.0 to 360.0
//   - ew_sigint.confidence: 0.0 to 1.0
//   - ew_sigint.modulation_type: non-empty
//   - position (if provided): lat -90 to +90, lon -180 to +180
//
// FLAG AS SUSPECT:
//   - ew_sigint.power_dbm: -200 to +100
//   - ew_sigint.pri_microseconds: 0.1 to 100000
```

### 3.2 ELINT/COMINT Validator

```go
// CLASSIFICATION: UNCLASSIFIED
// Validate checks ELINT/COMINT-specific rules:
//
// REJECT TO DLQ:
//   - sensor_id: non-empty, max 128 chars
//   - sensor_type: must be SENSOR_TYPE_ELINT_COMINT
//   - observation_time: not > 5min future, not > 24h past
//   - classification: valid enum (not UNSPECIFIED)
//   - elint_comint.emitter_id: non-empty
//   - elint_comint.radar_type: non-empty
//   - elint_comint.frequency_mhz: 0.5 to 40000
//   - elint_comint.cep_meters: > 0
//   - elint_comint.confidence: 0.0 to 1.0
//   - elint_comint.content_classification: valid enum
//   - position (if provided): lat -90 to +90, lon -180 to +180
//
// ADDITIONAL:
//   - content_classification must not exceed data classification
//   - scan_type must be one of: "circular", "sector", "track-while-scan"
```

### 3.3 ISR Metadata Validator

```go
// CLASSIFICATION: UNCLASSIFIED
// Validate checks ISR-specific rules:
//
// REJECT TO DLQ:
//   - sensor_id: non-empty, max 128 chars
//   - sensor_type: must be SENSOR_TYPE_ISR
//   - observation_time: not > 5min future, not > 24h past
//   - classification: valid enum (not UNSPECIFIED)
//   - isr.platform_id: non-empty
//   - isr.sensor_name: one of "EO", "IR", "SAR", "MTI"
//   - isr.image_id: non-empty
//   - isr.coverage_polygon: at least 3 vertices
//   - Each polygon vertex: lat -90 to +90, lon -180 to +180
//   - Each detection confidence: 0.0 to 1.0
//
// FLAG AS SUSPECT:
//   - isr.gsd_meters: 0.01 to 50.0 (typical range)
```

### 3.4 AIS/BFT Validator

```go
// CLASSIFICATION: UNCLASSIFIED
// Validate checks AIS/BFT-specific rules:
//
// REJECT TO DLQ:
//   - sensor_id: non-empty, max 128 chars
//   - sensor_type: must be SENSOR_TYPE_AIS_BFT
//   - observation_time: not > 5min future, not > 24h past
//   - classification: valid enum (not UNSPECIFIED)
//   - ais_bft.mmsi: exactly 9 digits (regex: ^[0-9]{9}$)
//   - ais_bft.vessel_name: non-empty
//   - ais_bft.vessel_type_code: 1 to 99
//   - ais_bft.ais_message_type: one of 1, 2, 3, 5, 18, 24
//   - position: required for AIS (lat -90 to +90, lon -180 to +180)
//   - position.speed_knots: 0 to 999
//   - position.heading_degrees: 0 to 360
//
// AIS SPOOFING CHECKS (flag suspect):
//   - Speed jump > 50 knots between consecutive reports from same MMSI
//   - Position jump > 100 NM between consecutive reports from same MMSI
//   - MMSI format anomalies (MID country code mismatch with declared flag)
//
// CLASSIFICATION:
//   - If is_bft=true, classification must be ≥ PROTECTED_B
```

### 3.5 Cyber Threat Validator

```go
// CLASSIFICATION: UNCLASSIFIED
// Validate checks Cyber-specific rules:
//
// REJECT TO DLQ:
//   - sensor_id: non-empty, max 128 chars
//   - sensor_type: must be SENSOR_TYPE_CYBER
//   - observation_time: not > 5min future, not > 24h past
//   - classification: valid enum (not UNSPECIFIED)
//   - cyber.stix_id: non-empty, starts with "indicator--"
//   - cyber.ioc_type: one of "ipv4-addr", "domain-name", "file:hashes", "url"
//   - cyber.ioc_value: non-empty
//   - cyber.confidence: 0.0 to 1.0
//   - cyber.valid_from: not in future
//   - cyber.source_feed: non-empty
//   - cyber.dedup_hash: non-empty, 64 hex chars (SHA-256)
//
// DEDUPLICATION:
//   - Check dedup_hash against recent hash cache (last 1000 entries)
//   - Reject duplicates to DLQ with reason "duplicate"
//
// IOC TYPE VALIDATION:
//   - "ipv4-addr": valid IPv4 format
//   - "domain-name": valid domain format
//   - "url": valid URL format
//   - "file:hashes": non-empty hash value
```

---

## 4. Test Scenarios Per Service

### 4.1 EW/SIGINT Tests

| #   | Test                      | Expected |
| --- | ------------------------- | -------- |
| T01 | Valid EW intercept        | Accepted |
| T02 | Frequency below 0.5 MHz   | Rejected |
| T03 | Frequency above 40000 MHz | Rejected |
| T04 | Missing emitter_id        | Rejected |
| T05 | Bearing out of range      | Rejected |
| T06 | Confidence out of range   | Rejected |

### 4.2 ELINT/COMINT Tests

| #   | Test                                    | Expected |
| --- | --------------------------------------- | -------- |
| T01 | Valid ELINT detection                   | Accepted |
| T02 | CEP ≤ 0                                 | Rejected |
| T03 | Invalid scan_type                       | Rejected |
| T04 | Content classification exceeds metadata | Rejected |
| T05 | Missing radar_type                      | Rejected |

### 4.3 ISR Tests

| #   | Test                                    | Expected |
| --- | --------------------------------------- | -------- |
| T01 | Valid ISR observation with 2 detections | Accepted |
| T02 | Polygon < 3 vertices                    | Rejected |
| T03 | Invalid sensor_name                     | Rejected |
| T04 | Detection confidence > 1.0              | Rejected |
| T05 | Missing platform_id                     | Rejected |

### 4.4 AIS/BFT Tests

| #   | Test                      | Expected                         |
| --- | ------------------------- | -------------------------------- |
| T01 | Valid AIS position report | Accepted                         |
| T02 | MMSI not 9 digits         | Rejected                         |
| T03 | Vessel type code > 99     | Rejected                         |
| T04 | Invalid AIS message type  | Rejected                         |
| T05 | BFT with UNCLASSIFIED     | Rejected (must be ≥ PROTECTED_B) |
| T06 | Missing position          | Rejected (required for AIS)      |
| T07 | Speed jump > 50 knots     | Warning flag                     |

### 4.5 Cyber Tests

| #   | Test                              | Expected             |
| --- | --------------------------------- | -------------------- |
| T01 | Valid cyber IOC                   | Accepted             |
| T02 | Invalid STIX ID format            | Rejected             |
| T03 | Invalid IOC type                  | Rejected             |
| T04 | Invalid IPv4 format for ipv4-addr | Rejected             |
| T05 | Duplicate dedup_hash              | Rejected (duplicate) |
| T06 | Invalid SHA-256 hash              | Rejected             |
| T07 | Future valid_from                 | Rejected             |

---

## 5. Agent Invocation

```
@greatest-ever-developer Implement Module 05 from docs/implementation/05-sensor-ingestion-batch.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/04-sensor-ingestion-radar.md — this is the reference implementation you must clone
- READ the actual source code of svc-radar-ingestion/ — replicate its structure exactly
- Create 5 new service directories: svc-ew-ingestion, svc-elint-ingestion, svc-isr-ingestion, svc-ais-ingestion, svc-cyber-ingestion
- Each follows IDENTICAL structure to svc-radar-ingestion
- Only validator.go and normalizer.go differ per service
- The handler, producer, enricher, and main.go differ only in config values (topic names, ports, sensor types)
- Consider extracting shared handler/producer/enricher into pkg/ingestion to reduce duplication
- Each service must have its own go.mod, Dockerfile, and README.md

Deliverables:
1. 5 service directories with complete source code
2. Sensor-specific validators with all rules from §3
3. Unit tests for each validator (≥80% coverage)
4. Integration tests for one service (AIS — most complex validator)
5. go vet ./... passes for each service
6. Add all 5 services to go.work
```
