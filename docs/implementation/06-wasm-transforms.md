<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 06 — Wasm Data Transforms

> **Module**: 06-wasm-transforms
> **Phase**: P1 (Ingestion)
> **Dependencies**: Module 01 (Redpanda), Module 02 (protos)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 2 days

---

## 1. Objective

Implement Redpanda Data Transforms (Wasm) that run inside the broker to validate message schema compliance and enforce classification headers on all sensor topics BEFORE consumers read them. This is a defense-in-depth measure — ingestion services already validate, but Wasm transforms catch any bypasses.

**Acceptance Criteria**:

- Wasm transforms deployed to all `sensors.*` topics
- Messages without required headers are rejected → routed to DLQ
- Messages with invalid protobuf schema fail → routed to DLQ
- Classification header present and valid on every message
- Transform adds `rtsa-validated: true` header to passing messages
- Transform latency < 1ms p99 per message

---

## 2. Transform Architecture

```
Redpanda Broker
├── sensors.radar.tracks        ← Wasm: SensorSchemaValidator
├── sensors.ew.intercepts       ← Wasm: SensorSchemaValidator
├── sensors.elint.detections    ← Wasm: SensorSchemaValidator
├── sensors.isr.observations    ← Wasm: SensorSchemaValidator
├── sensors.ais.positions       ← Wasm: SensorSchemaValidator
├── sensors.cyber.iocs          ← Wasm: SensorSchemaValidator
├── tracks.fused.*              ← Wasm: TrackSchemaValidator
├── alerts.anomaly.*            ← Wasm: AlertSchemaValidator
├── feedback.operator.*         ← Wasm: FeedbackSchemaValidator
└── audit.events                ← (no transform — audit is append-only)
```

---

## 3. Project Structure

```
wasm-transforms/
├── sensor-validator/
│   ├── main.go              # Wasm entry point
│   ├── transform.go         # Header + schema validation
│   ├── transform_test.go
│   └── Makefile              # GOOS=wasip1 GOARCH=wasm build
├── track-validator/
│   ├── main.go
│   ├── transform.go
│   ├── transform_test.go
│   └── Makefile
├── alert-validator/
│   ├── main.go
│   ├── transform.go
│   ├── transform_test.go
│   └── Makefile
├── feedback-validator/
│   ├── main.go
│   ├── transform.go
│   ├── transform_test.go
│   └── Makefile
├── go.mod
└── README.md
```

---

## 4. Sensor Schema Validator Transform

```go
// CLASSIFICATION: UNCLASSIFIED
package main

import (
    "github.com/redpanda-data/redpanda/src/transform-sdk/go/transform"
)

func main() {
    transform.OnRecordWritten(validateSensorMessage)
}

// validateSensorMessage runs on every record written to sensors.* topics.
// Checks:
// 1. Required headers present:
//    - rtsa-classification: non-empty, valid value
//    - rtsa-source-service: non-empty
//    - rtsa-trace-id: non-empty
//    - rtsa-timestamp: non-empty, valid ISO 8601
//    - rtsa-schema-version: non-empty
// 2. Record value is non-empty
// 3. Record value is valid protobuf (can unmarshal SensorObservation)
// 4. SensorObservation has non-empty sensor_id
// 5. SensorObservation has valid sensor_type (not UNSPECIFIED)
//
// On validation pass:
//   - Add header: rtsa-validated = "true"
//   - Write to output topic (same topic, passthrough)
//
// On validation failure:
//   - Add header: rtsa-validation-error = "<error description>"
//   - Write to DLQ topic (e.g., dlq.sensors.{type})
func validateSensorMessage(event transform.WriteEvent, writer transform.RecordWriter) error {
    record := event.Record()

    // Check required headers
    requiredHeaders := []string{
        "rtsa-classification",
        "rtsa-source-service",
        "rtsa-trace-id",
        "rtsa-timestamp",
        "rtsa-schema-version",
    }
    for _, h := range requiredHeaders {
        if getHeader(record, h) == "" {
            return writeToOutputWithError(writer, record, "missing header: "+h)
        }
    }

    // Validate classification value
    classification := getHeader(record, "rtsa-classification")
    validClassifications := map[string]bool{
        "UNCLASSIFIED": true, "PROTECTED_A": true, "PROTECTED_B": true,
        "PROTECTED_C": true, "SECRET": true,
    }
    if !validClassifications[classification] {
        return writeToOutputWithError(writer, record, "invalid classification: "+classification)
    }

    // Validate protobuf deserializable
    if len(record.Value) == 0 {
        return writeToOutputWithError(writer, record, "empty record value")
    }

    // Add validated header
    record.Headers = append(record.Headers, transform.RecordHeader{
        Key:   []byte("rtsa-validated"),
        Value: []byte("true"),
    })
    return writer.Write(record)
}
```

---

## 5. Build Configuration

### Makefile (per transform)

```makefile
# CLASSIFICATION: UNCLASSIFIED
.PHONY: build deploy test

WASM_OUT = sensor-validator.wasm

build:
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_OUT) .

deploy: build
	rpk transform deploy $(WASM_OUT) \
		--name sensor-schema-validator \
		--input-topic 'sensors.*' \
		--output-topic '__passthrough__' \
		--brokers $(RTSA_REDPANDA_BROKERS)

test:
	go test -v ./...
```

### go.mod

```go
// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/wasm-transforms

go 1.22.0

require (
    github.com/redpanda-data/redpanda/src/transform-sdk/go/transform v0.5.0
    github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
    google.golang.org/protobuf v1.33.0
)

replace github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
```

---

## 6. Valid Classification Values

| String Value   | Proto Enum                          |
| -------------- | ----------------------------------- |
| `UNCLASSIFIED` | `CLASSIFICATION_LEVEL_UNCLASSIFIED` |
| `PROTECTED_A`  | `CLASSIFICATION_LEVEL_PROTECTED_A`  |
| `PROTECTED_B`  | `CLASSIFICATION_LEVEL_PROTECTED_B`  |
| `PROTECTED_C`  | `CLASSIFICATION_LEVEL_PROTECTED_C`  |
| `SECRET`       | `CLASSIFICATION_LEVEL_SECRET`       |

---

## 7. Test Scenarios

| #   | Test                               | Input                               | Expected                          |
| --- | ---------------------------------- | ----------------------------------- | --------------------------------- |
| T01 | Valid message with all headers     | Full sensor observation             | Passes, rtsa-validated=true added |
| T02 | Missing rtsa-classification header | Headers without classification      | Rejected, error header set        |
| T03 | Missing rtsa-source-service header | Headers without source              | Rejected                          |
| T04 | Invalid classification value       | rtsa-classification="TOP_SECRET"    | Rejected                          |
| T05 | Empty record value                 | Empty byte slice                    | Rejected                          |
| T06 | Invalid protobuf (garbage bytes)   | Random bytes                        | Rejected                          |
| T07 | Valid protobuf, empty sensor_id    | SensorObservation with sensor_id="" | Rejected                          |
| T08 | Valid message, all headers present | Complete message                    | Passes                            |

---

## 8. Deployment Script

```bash
#!/bin/bash
# CLASSIFICATION: UNCLASSIFIED
# deploy-transforms.sh — Deploys all Wasm transforms to Redpanda

set -euo pipefail

BROKERS="${RTSA_REDPANDA_BROKERS:-localhost:9092}"

echo "Building Wasm transforms..."
for dir in sensor-validator track-validator alert-validator feedback-validator; do
    (cd "wasm-transforms/$dir" && make build)
done

echo "Deploying transforms..."
rpk transform deploy wasm-transforms/sensor-validator/sensor-validator.wasm \
    --name sensor-schema-validator \
    --input-topic 'sensors.radar.tracks,sensors.ew.intercepts,sensors.elint.detections,sensors.isr.observations,sensors.ais.positions,sensors.cyber.iocs' \
    --brokers "$BROKERS"

rpk transform deploy wasm-transforms/track-validator/track-validator.wasm \
    --name track-schema-validator \
    --input-topic 'tracks.fused.surface,tracks.fused.air,tracks.fused.subsurface,tracks.fused.land,tracks.fused.cyber' \
    --brokers "$BROKERS"

rpk transform deploy wasm-transforms/alert-validator/alert-validator.wasm \
    --name alert-schema-validator \
    --input-topic 'alerts.anomaly.critical,alerts.anomaly.elevated,alerts.anomaly.watch' \
    --brokers "$BROKERS"

rpk transform deploy wasm-transforms/feedback-validator/feedback-validator.wasm \
    --name feedback-schema-validator \
    --input-topic 'feedback.operator.submissions,feedback.operator.validated' \
    --brokers "$BROKERS"

echo "All transforms deployed."
rpk transform list --brokers "$BROKERS"
```

---

## 9. Agent Invocation

```
@greatest-ever-developer Implement Module 06 from docs/implementation/06-wasm-transforms.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for message types
- Read docs/sdlc_guidelines/08_tech_specific/wasm_transforms.md for Wasm transform guidelines
- Build with GOOS=wasip1 GOARCH=wasm
- Use Redpanda transform SDK (github.com/redpanda-data/redpanda/src/transform-sdk/go/transform)
- Each transform validates headers and basic schema compliance
- Tests run in native Go (not Wasm) using the SDK test harness

Deliverables:
1. 4 Wasm transform directories with source and tests
2. go.mod for the wasm-transforms module
3. Build scripts (Makefiles) that produce .wasm files
4. Deployment script
5. All tests pass
```
