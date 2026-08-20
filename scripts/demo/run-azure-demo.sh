#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Run a finite RTSA simulator Job against services already deployed in AKS.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

NAMESPACE="rtsa"
SCENARIO="maritime-demo.yaml"
DURATION_MINUTES="20"
IMAGE=""
ACR_NAME="${ACR_NAME:-}"
IMAGE_TAG="${IMAGE_TAG:-}"
BUILD_IMAGE="true"
CLEANUP="false"
WAIT_TIMEOUT_MINUTES="30"
JOB_NAME="rtsa-demo-$(date -u +%Y%m%d%H%M%S)"

usage() {
  cat <<'USAGE'
Run a finite RTSA simulator Job in an AKS dev namespace.

Usage:
  bash scripts/demo/run-azure-demo.sh [options]

Options:
  --scenario FILE       Scenario under tools/simulator/scenarios (default: maritime-demo.yaml)
  --duration MINUTES    Simulation duration (default: 20)
  --namespace NAME      AKS namespace (default: rtsa)
  --acr NAME            Azure Container Registry name (or ACR_NAME)
  --tag TAG             Immutable image tag (default: current git commit)
  --image IMAGE         Existing simulator image; skips ACR build
  --no-build             Do not build/push the simulator image
  --cleanup              Delete the Job after it finishes (TTL still applies otherwise)
  --timeout MINUTES     Job wait timeout (default: 30)
  --job-name NAME       Explicit Job name
  -h, --help            Show this help

Examples:
  bash scripts/demo/run-azure-demo.sh --acr acrrtsashared --cleanup
  bash scripts/demo/run-azure-demo.sh --image acrrtsashared.azurecr.io/rtsa-simulator:demo-20260820 --duration 10

The AKS context must already be authenticated. Run cleanup separately with
scripts/demo/cleanup-azure-demo.sh if a terminated shell leaves a Job behind.
USAGE
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --scenario) SCENARIO="$2"; shift 2 ;;
    --duration) DURATION_MINUTES="$2"; shift 2 ;;
    --namespace) NAMESPACE="$2"; shift 2 ;;
    --acr) ACR_NAME="$2"; shift 2 ;;
    --tag) IMAGE_TAG="$2"; shift 2 ;;
    --image) IMAGE="$2"; BUILD_IMAGE="false"; shift 2 ;;
    --no-build) BUILD_IMAGE="false"; shift ;;
    --cleanup) CLEANUP="true"; shift ;;
    --timeout) WAIT_TIMEOUT_MINUTES="$2"; shift 2 ;;
    --job-name) JOB_NAME="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

[[ "$DURATION_MINUTES" =~ ^[1-9][0-9]*$ ]] || { echo "duration must be a positive integer" >&2; exit 1; }
[[ "$WAIT_TIMEOUT_MINUTES" =~ ^[1-9][0-9]*$ ]] || { echo "timeout must be a positive integer" >&2; exit 1; }
[[ "$SCENARIO" != */* ]] || { echo "scenario must be a file name, not a path" >&2; exit 1; }
[[ "$SCENARIO" == *.yaml ]] || { echo "scenario must be a YAML file" >&2; exit 1; }
[[ "$JOB_NAME" =~ ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$ ]] || { echo "job name is not a valid Kubernetes name" >&2; exit 1; }
[[ -f "$PROJECT_ROOT/tools/simulator/scenarios/$SCENARIO" ]] || { echo "scenario file not found: $SCENARIO" >&2; exit 1; }

command -v kubectl >/dev/null || { echo "kubectl is required" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }
kubectl get namespace "$NAMESPACE" >/dev/null

if [[ -z "$IMAGE_TAG" ]]; then
  IMAGE_TAG="$(git -C "$PROJECT_ROOT" rev-parse HEAD)"
fi

if [[ -z "$IMAGE" ]]; then
  [[ -n "$ACR_NAME" ]] || { echo "--acr or ACR_NAME is required when building the image" >&2; exit 1; }
  IMAGE="${ACR_NAME}.azurecr.io/rtsa-simulator:${IMAGE_TAG}"
fi

if [[ "$BUILD_IMAGE" == "true" ]]; then
  command -v az >/dev/null || { echo "Azure CLI is required to build the simulator image" >&2; exit 1; }
  [[ -n "$ACR_NAME" ]] || { echo "--acr or ACR_NAME is required to build the image" >&2; exit 1; }
  echo "Building immutable simulator image ${IMAGE} in ACR..."
  az acr build --registry "$ACR_NAME" --platform linux/arm64 \
    --image "rtsa-simulator:${IMAGE_TAG}" \
    --file tools/simulator/Dockerfile "$PROJECT_ROOT"
fi

cleanup_job() {
  if [[ "$CLEANUP" == "true" && -n "${JOB_NAME:-}" ]]; then
    kubectl delete job "$JOB_NAME" --namespace "$NAMESPACE" --ignore-not-found
  fi
}

delete_job_on_signal() {
  kubectl delete job "$JOB_NAME" --namespace "$NAMESPACE" --ignore-not-found || true
  exit 143
}
trap delete_job_on_signal INT TERM

echo "Creating finite simulator Job ${JOB_NAME} in namespace ${NAMESPACE}..."
deadline_seconds=$((WAIT_TIMEOUT_MINUTES * 60))
kubectl create job "$JOB_NAME" \
  --namespace "$NAMESPACE" \
  --image "$IMAGE" \
  --restart Never \
  --dry-run=client -o json \
  -- /app/simulator \
    --scenario "/app/scenarios/${SCENARIO}" \
    --duration "${DURATION_MINUTES}m" \
  | jq \
      --arg scenario "$SCENARIO" \
      --argjson deadline "$deadline_seconds" \
      '.metadata.labels = {
         "app.kubernetes.io/name": "rtsa-demo-simulator",
         "app.kubernetes.io/component": "demo",
         "rtsa.dev/scenario": $scenario
       }
       | .spec.backoffLimit = 0
       | .spec.activeDeadlineSeconds = $deadline
       | .spec.ttlSecondsAfterFinished = 3600
       | .spec.template.spec.containers[0].env = [
           {name: "SIM_RADAR_ENDPOINT", value: "svc-radar-ingestion:50051"},
           {name: "SIM_EW_ENDPOINT", value: "svc-ew-ingestion:50051"},
           {name: "SIM_ELINT_ENDPOINT", value: "svc-elint-ingestion:50051"},
           {name: "SIM_ISR_ENDPOINT", value: "svc-isr-ingestion:50051"},
           {name: "SIM_AIS_ENDPOINT", value: "svc-ais-ingestion:50051"},
           {name: "SIM_CYBER_ENDPOINT", value: "svc-cyber-ingestion:50051"}
         ]' \
  | kubectl apply -f -

set +e
kubectl wait --for=condition=complete --timeout="${WAIT_TIMEOUT_MINUTES}m" \
  "job/${JOB_NAME}" --namespace "$NAMESPACE"
wait_status=$?
set -e

kubectl logs "job/${JOB_NAME}" --namespace "$NAMESPACE" --tail=-1 || true
if [[ "$wait_status" -ne 0 ]]; then
  echo "Simulator Job ${JOB_NAME} did not complete successfully." >&2
  kubectl describe job "$JOB_NAME" --namespace "$NAMESPACE" >&2 || true
  cleanup_job
  exit "$wait_status"
fi

echo "Simulator Job ${JOB_NAME} completed successfully."
cleanup_job
