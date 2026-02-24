<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Test Data Simulator

> **Module**: 16-test-data-simulators  
> **Classification**: UNCLASSIFIED  
> **All generated data is synthetic — no real operational data.**

Generates synthetic sensor observations for all 6 RTSA sensor types and sends them to the ingestion services via gRPC.

## Features

- **6 sensor generators**: Radar, EW/SIGINT, ELINT/COMINT, ISR, AIS/BFT, Cyber
- **5 movement patterns**: Straight-line, Patrol, Loitering, Zigzag, Random
- **5 anomaly types**: Speed, Route Deviation, AIS Manipulation, Behavioral, Proximity
- **Mid-Atlantic coordinates**: 43–47°N, 55–65°W
- **Deterministic mode**: Seed-based RNG for reproducible test scenarios
- **Scenario files**: YAML-based configuration for different test scenarios
- **CLI + Docker**: Flexible deployment options

## Quick Start

```bash
# Dry-run (no gRPC required)
go run ./cmd/simulator/ --dry-run --seed 42

# Run with default scenario
go run ./cmd/simulator/ --scenario scenarios/default.yaml --dry-run

# Run with overrides
go run ./cmd/simulator/ \
  --surface-entities 10 \
  --air-entities 5 \
  --update-interval 1s \
  --anomaly-rate 0.10 \
  --seed 42 \
  --dry-run
```

## Docker

```bash
# Build (from repo root)
docker build -f tools/simulator/Dockerfile -t rtsa/simulator:latest .

# Run with default scenario
docker run --rm \
  --network rtsa-net \
  -e SIM_RADAR_ENDPOINT=svc-radar-ingestion:50051 \
  -e SIM_AIS_ENDPOINT=svc-ais-ingestion:50055 \
  rtsa/simulator:latest

# Run stress scenario
docker run --rm \
  --network rtsa-net \
  rtsa/simulator:latest --scenario /app/scenarios/stress.yaml
```

## Scenario Files

| File | Description |
|------|-------------|
| `scenarios/default.yaml` | Standard demo: 20 surface, 10 air, 5 sub entities, 5% anomalies |
| `scenarios/stress.yaml` | High-volume: 200/100/50 entities, 10% anomalies, 30-minute run |
| `scenarios/anomaly-demo.yaml` | Anomaly showcase: 30% anomaly rate, all 5 anomaly types |

## Configuration

All settings can be overridden via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SIM_RADAR_ENDPOINT` | `localhost:50051` | Radar ingestion gRPC endpoint |
| `SIM_EW_ENDPOINT` | `localhost:50052` | EW ingestion endpoint |
| `SIM_ELINT_ENDPOINT` | `localhost:50053` | ELINT ingestion endpoint |
| `SIM_ISR_ENDPOINT` | `localhost:50054` | ISR ingestion endpoint |
| `SIM_AIS_ENDPOINT` | `localhost:50055` | AIS ingestion endpoint |
| `SIM_CYBER_ENDPOINT` | `localhost:50056` | Cyber ingestion endpoint |
| `SIM_SURFACE_ENTITIES` | `20` | Number of surface entities |
| `SIM_AIR_ENTITIES` | `10` | Number of air entities |
| `SIM_SUB_ENTITIES` | `5` | Number of subsurface entities |
| `SIM_UPDATE_INTERVAL_MS` | `1000` | Tick interval in milliseconds |
| `SIM_DURATION_MINUTES` | `0` | Run duration (0 = infinite) |
| `SIM_ANOMALY_RATE` | `0.05` | Fraction of anomalous entities |
| `SIM_RANDOM_SEED` | `0` | RNG seed (0 = random) |
| `SIM_SCENARIO_FILE` | `` | Path to YAML scenario file |
| `SIM_TLS_ENABLED` | `false` | Enable mTLS for gRPC |
| `SIM_TLS_CERT_FILE` | `` | Client certificate path |
| `SIM_TLS_KEY_FILE` | `` | Client key path |
| `SIM_TLS_CA_FILE` | `` | CA certificate path |

## AIS Manipulation Anomaly

When an entity has the `AIS_MANIPULATION` anomaly type, the simulator generates **both** a radar observation (true position) and an AIS observation (offset position, 0.5–2.0 NM away). This creates the discrepancy that the Anomaly Detection service (Module 08) must detect.

## Coordinate Bounds

All geographic coordinates are constrained to the Mid-Atlantic operational area:
- **Latitude**: 43.0° to 47.0°N
- **Longitude**: −65.0° to −55.0°W

## Tests

```bash
go test ./...                    # Run all tests
go test -cover ./...             # With coverage
go test -run TestEntityManager   # Specific test
```
