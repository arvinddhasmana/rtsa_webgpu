#!/bin/bash
# CLASSIFICATION: UNCLASSIFIED
# deploy-transforms.sh — Builds and deploys all Wasm transforms to Redpanda.
#
# Usage:
#   RTSA_REDPANDA_BROKERS=localhost:9092 ./scripts/deploy-transforms.sh
#
# Environment variables:
#   RTSA_REDPANDA_BROKERS  Comma-separated list of broker addresses.
#                          Default: localhost:9092

set -euo pipefail

BROKERS="${RTSA_REDPANDA_BROKERS:-localhost:9092}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TRANSFORMS_DIR="${REPO_ROOT}/wasm-transforms"

echo "=== Building Wasm transforms ==="
for dir in sensor-validator track-validator alert-validator feedback-validator; do
    echo "  Building ${dir}..."
    (cd "${TRANSFORMS_DIR}/${dir}" && make build)
done

echo ""
echo "=== Deploying transforms to Redpanda (${BROKERS}) ==="

rpk transform deploy "${TRANSFORMS_DIR}/sensor-validator/sensor-validator.wasm" \
    --name sensor-schema-validator \
    --input-topic 'sensors.radar.tracks,sensors.ew.intercepts,sensors.elint.detections,sensors.isr.observations,sensors.ais.positions,sensors.cyber.iocs' \
    --brokers "${BROKERS}"

rpk transform deploy "${TRANSFORMS_DIR}/track-validator/track-validator.wasm" \
    --name track-schema-validator \
    --input-topic 'tracks.fused.surface,tracks.fused.air,tracks.fused.subsurface,tracks.fused.land,tracks.fused.cyber' \
    --brokers "${BROKERS}"

rpk transform deploy "${TRANSFORMS_DIR}/alert-validator/alert-validator.wasm" \
    --name alert-schema-validator \
    --input-topic 'alerts.anomaly.critical,alerts.anomaly.elevated,alerts.anomaly.watch' \
    --brokers "${BROKERS}"

rpk transform deploy "${TRANSFORMS_DIR}/feedback-validator/feedback-validator.wasm" \
    --name feedback-schema-validator \
    --input-topic 'feedback.operator.submissions,feedback.operator.validated' \
    --brokers "${BROKERS}"

echo ""
echo "=== All transforms deployed ==="
rpk transform list --brokers "${BROKERS}"
