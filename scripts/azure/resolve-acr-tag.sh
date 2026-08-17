#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Resolve the most recently pushed real image tag for a service in ACR.
# cd-build.yml only ever pushes commit-SHA tags — there is no "latest" tag —
# so callers needing a fallback tag for an unmodified/missing service must
# look up what was actually pushed, not assume a fixed tag name.

set -euo pipefail

ACR_NAME=""
REPOSITORY=""

usage() {
  echo "Usage: resolve-acr-tag.sh --acr <name> --repository <service>" >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --acr)
      ACR_NAME="${2:-}"
      shift 2
      ;;
    --repository)
      REPOSITORY="${2:-}"
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

[[ -n "$ACR_NAME" && -n "$REPOSITORY" ]] || { usage; exit 2; }

# Exclude cosign's .sig/.att pseudo-tags — only real, deployable image tags count.
tag="$(
  az acr repository show-tags --name "$ACR_NAME" --repository "$REPOSITORY" \
    --orderby time_desc --top 50 -o tsv 2>/dev/null \
    | grep -Ev '\.(sig|att)$' \
    | head -1 || true
)"

[[ -n "$tag" ]] || {
  echo "ERROR: no pushed image tags found for ${ACR_NAME}.azurecr.io/${REPOSITORY}; it may never have been built." >&2
  exit 1
}

echo "$tag"
