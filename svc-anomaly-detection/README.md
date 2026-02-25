<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-anomaly-detection

> **Module**: 08 — Anomaly Detection Service
> **Classification**: UNCLASSIFIED

## Overview

The Anomaly Detection Service consumes fused tracks from `tracks.fused.*` Redpanda topics, extracts behavioural features, runs rule-based anomaly detectors, and produces severity-rated `AnomalyAlert` messages to `alerts.anomaly.{severity}` topics.

For MVP this service uses **rule-based / statistical detectors** (model version `rules-v1.0.0`). A model-based ML approach is reserved for Module 11 (Training Pipeline).

---

## Architecture

```
tracks.fused.*
      │
      ▼
TrackConsumer (Redpanda consumer group: anomaly-detection)
      │
      ▼
FeatureExtractor
  ├── Speed features     (30-min rolling avg, σ)
  ├── Route features     (expected heading, deviation)
  ├── AIS features       (AIS vs fused position delta)
  ├── Behavioral         (loitering, zigzag, speed pulsing)
  ├── Temporal           (time-of-day p-value)
  └── Proximity          (exclusion zone distance)
      │
      ▼
Detectors (each independent, enable/disable per config)
  ├── SpeedDetector       → if sigma > 3.0
  ├── RouteDeviationDetector → if > 30° sustained for 3 updates
  ├── AISManipulationDetector → if AIS-fused delta > 0.5 NM
  ├── BehavioralDetector  → if pattern confidence > 0.75
  ├── TemporalDetector    → if p-value < 0.05 (track age > 24h)
  └── ProximityDetector   → if inside / approaching exclusion zone
      │
      ▼
Severity Mapping
  < 0.50 → NORMAL  (no alert)
  0.50–0.69 → WATCH
  0.70–0.89 → ELEVATED
  ≥ 0.90 → CRITICAL
      │
      ▼
ExplanationGenerator (template-based, human-readable)
      │
      ▼
AlertProducer → alerts.anomaly.{critical,elevated,watch}
```

---

## Anomaly Detectors

| Detector | Threshold | Confidence Formula |
|---|---|---|
| Speed | sigma > 3.0 | sigma / (3.0 × 2) |
| Route Deviation | > 30° for 3 consecutive updates | deviation / 90 |
| AIS Manipulation | > 0.5 NM delta | delta / (0.5 × 3) |
| Behavioral | PatternConfidence > 0.75 | PatternConfidence |
| Temporal | p-value < 0.05 (track ≥ 24h) | 1.0 - p-value |
| Proximity | Inside / approaching zone | 1.0 / (1 - dist/radius×2) |

---

## Configuration

All settings are read from environment variables with `RTSA_` prefix.

| Variable | Default | Description |
|---|---|---|
| `RTSA_REDPANDA_BROKERS` | `localhost:19092` | Comma-separated broker list |
| `RTSA_ANOMALY_CONSUMER_GROUP` | `anomaly-detection` | Redpanda consumer group |
| `RTSA_MODEL_VERSION` | `rules-v1.0.0` | Model version tag on alerts |
| `RTSA_HEALTH_ADDR` | `:8081` | Health check HTTP listen address |
| `RTSA_LOG_LEVEL` | `info` | debug / info / warn / error |
| `RTSA_LOG_FORMAT` | `json` | json / text |
| `RTSA_TRACK_HISTORY_MAX_ENTRIES` | `100` | Circular buffer size per track |
| `RTSA_TRACK_HISTORY_MAX_AGE` | `2h` | Maximum track history retention |
| `RTSA_EXCLUSION_ZONES_JSON` | _(empty)_ | JSON array of ExclusionZone objects |
| `RTSA_EXCLUSION_ZONES_FILE` | _(empty)_ | Path to JSON file of exclusion zones |
| `RTSA_DETECTOR_SPEED_ENABLED` | `true` | Enable speed anomaly detector |
| `RTSA_DETECTOR_SPEED_SIGMA_THRESHOLD` | `3.0` | Sigma threshold for speed detection |
| `RTSA_DETECTOR_ROUTE_ENABLED` | `true` | Enable route deviation detector |
| `RTSA_DETECTOR_ROUTE_DEVIATION_DEG` | `30.0` | Deviation threshold (degrees) |
| `RTSA_DETECTOR_ROUTE_SUSTAINED_N` | `3` | Required consecutive deviation updates |
| `RTSA_DETECTOR_AIS_ENABLED` | `true` | Enable AIS manipulation detector |
| `RTSA_DETECTOR_AIS_DISCREPANCY_NM` | `0.5` | AIS-fused position discrepancy threshold (NM) |
| `RTSA_DETECTOR_BEHAVIORAL_ENABLED` | `true` | Enable behavioral pattern detector |
| `RTSA_DETECTOR_BEHAVIORAL_CONFIDENCE_THRESHOLD` | `0.75` | Confidence threshold for behavioral detection |
| `RTSA_DETECTOR_TEMPORAL_ENABLED` | `true` | Enable temporal anomaly detector |
| `RTSA_DETECTOR_TEMPORAL_P_VALUE` | `0.05` | P-value threshold for temporal detection |
| `RTSA_DETECTOR_PROXIMITY_ENABLED` | `true` | Enable proximity alert detector |

### Exclusion Zones JSON Format

```json
[
  {
    "name": "Halifax Naval Base",
    "center_lat": 44.6476,
    "center_lon": -63.5728,
    "radius_nm": 2.0
  }
]
```

---

## Input Topics

| Topic | Message Type | Description |
|---|---|---|
| `tracks.fused.surface` | `entityv1.FusedTrack` | Surface entity fused tracks |
| `tracks.fused.air` | `entityv1.FusedTrack` | Air entity fused tracks |
| `tracks.fused.subsurface` | `entityv1.FusedTrack` | Subsurface entity fused tracks |
| `tracks.fused.land` | `entityv1.FusedTrack` | Land entity fused tracks |
| `tracks.fused.cyber` | `entityv1.FusedTrack` | Cyber entity fused tracks |

## Output Topics

| Topic | Severity | Message Type |
|---|---|---|
| `alerts.anomaly.critical` | ≥ 0.90 | `inferencev1.AnomalyAlert` |
| `alerts.anomaly.elevated` | 0.70–0.89 | `inferencev1.AnomalyAlert` |
| `alerts.anomaly.watch` | 0.50–0.69 | `inferencev1.AnomalyAlert` |

---

## Building

```bash
# From repo root:
cd svc-anomaly-detection
go build ./...
go test ./... -race -count=1

# Docker build (from repo root):
docker build -f svc-anomaly-detection/Dockerfile -t svc-anomaly-detection:dev .
```

---

## Health Check

The service exposes a health endpoint at `:8081/healthz` (JSON).

```bash
curl http://localhost:8081/healthz
```

---

## Security Notes

- No hardcoded credentials, connection strings, or classified data
- All external inputs (Redpanda messages) are validated before processing
- Classification level is propagated from track to alert without elevation
- Structured JSON logging — raw track payloads are never logged
- Runs as `nonroot` user in distroless container
