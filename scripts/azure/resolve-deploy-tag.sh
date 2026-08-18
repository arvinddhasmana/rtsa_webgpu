#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Universal per-service deploy decision:
#   a. Not deployed yet  -> deploy the latest image pushed to ACR.
#   b. Already deployed  -> only redeploy if ACR has a newer tag than what's running.
#
# On stdout: the tag to deploy (only when a deploy is needed).
# Exit codes: 0 = deploy needed (tag printed), 3 = already up to date (skip), 1 = error.

set -uo pipefail

ACR_NAME=""
SERVICE=""
NAMESPACE=""

usage() {
  echo "Usage: resolve-deploy-tag.sh --acr <name> --service <repo/deployment> --namespace <ns>" >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --acr)
      ACR_NAME="${2:-}"
      shift 2
      ;;
    --service)
      SERVICE="${2:-}"
      shift 2
      ;;
    --namespace)
      NAMESPACE="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
done

[[ -n "$ACR_NAME" && -n "$SERVICE" && -n "$NAMESPACE" ]] || { usage; exit 2; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

latest_tag="$("$SCRIPT_DIR/resolve-acr-tag.sh" --acr "$ACR_NAME" --repository "$SERVICE")" || exit 1

current_image="$(kubectl get deployment "$SERVICE" -n "$NAMESPACE" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || true)"

if [[ -z "$current_image" ]]; then
  echo "NOTICE: $SERVICE is not deployed yet — deploying latest ACR tag ($latest_tag)." >&2
  echo "$latest_tag"
  exit 0
fi

# A Helm --wait timeout can leave the release itself in failed state even after
# its Deployment becomes healthy. The image tag is then unchanged, so comparing
# images alone would skip the upgrade forever and the post-deploy verifier would
# keep failing. Redeploy the same tag once to create a successful Helm revision.
helm_status="$(helm status "$SERVICE" --namespace "$NAMESPACE" -o json 2>/dev/null |
  jq -r '.info.status // empty' 2>/dev/null || true)"
if [[ "$helm_status" == "failed" ]]; then
  echo "NOTICE: $SERVICE has failed Helm release status - redeploying latest ACR tag ($latest_tag) to recover it." >&2
  echo "$latest_tag"
  exit 0
fi

current_tag="${current_image##*:}"
if [[ "$current_tag" == "$latest_tag" ]]; then
  echo "NOTICE: $SERVICE is already running the latest ACR tag ($latest_tag) — skipping." >&2
  exit 3
fi

echo "NOTICE: $SERVICE is running $current_tag; newer tag $latest_tag is available in ACR — redeploying." >&2
echo "$latest_tag"
