#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Remove RTSA simulator Jobs created by the Azure demo runner.

set -euo pipefail

NAMESPACE="rtsa"
JOB_NAME=""

usage() {
  cat <<'USAGE'
Clean up completed or abandoned RTSA Azure demo Jobs.

Usage:
  bash scripts/demo/cleanup-azure-demo.sh [--namespace NAME] [--job-name NAME]

Without --job-name, every Job labelled app.kubernetes.io/name=rtsa-demo-simulator
in the selected namespace is deleted. This does not delete RTSA services,
Redpanda, ClickHouse, or persistent volumes.
USAGE
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --namespace) NAMESPACE="$2"; shift 2 ;;
    --job-name) JOB_NAME="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

command -v kubectl >/dev/null || { echo "kubectl is required" >&2; exit 1; }
if [[ -n "$JOB_NAME" ]]; then
  kubectl delete job "$JOB_NAME" --namespace "$NAMESPACE" --ignore-not-found
else
  kubectl delete jobs \
    --namespace "$NAMESPACE" \
    --selector app.kubernetes.io/name=rtsa-demo-simulator \
    --ignore-not-found
fi
