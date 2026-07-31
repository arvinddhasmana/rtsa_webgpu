#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Break a stale Terraform remote state lock on Azure Blob Storage.
#
# Uses `az storage account keys` (unconditional access) to break the blob lease,
# then runs `terraform force-unlock` to clean up the .terraform.lock.info metadata.
#
# Usage:
#   scripts/azure/tf-break-lock.sh --environment dev
#   scripts/azure/tf-break-lock.sh --environment dev --bootstrap-dir infra/terraform/bootstrap

set -euo pipefail

ENV_NAME="dev"
BOOTSTRAP_DIR="infra/terraform/bootstrap"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --environment) ENV_NAME="${2:-}"; shift 2 ;;
    --bootstrap-dir) BOOTSTRAP_DIR="${2:-}"; shift 2 ;;
    *) echo "ERROR: unknown argument: $1" >&2; exit 2 ;;
  esac
done

command -v az >/dev/null 2>&1        || { echo "ERROR: az CLI not found." >&2; exit 1; }
command -v terraform >/dev/null 2>&1 || { echo "ERROR: terraform not found." >&2; exit 1; }

if ! terraform -chdir="$BOOTSTRAP_DIR" output -json >/dev/null 2>&1; then
  echo "ERROR: Bootstrap Terraform outputs not available. Run scripts/azure/bootstrap.sh first." >&2
  exit 1
fi

TFSTATE_RG="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw shared_resource_group)"
TFSTATE_SA="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw state_storage_account_name)"
TFSTATE_CONTAINER="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw state_container_name)"
STATE_BLOB="rtsa-${ENV_NAME}.tfstate"

echo "== Breaking Azure blob lease: ${TFSTATE_SA}/${TFSTATE_CONTAINER}/${STATE_BLOB} =="
KEY="$(az storage account keys list \
  --account-name "$TFSTATE_SA" \
  --resource-group "$TFSTATE_RG" \
  --query '[0].value' -o tsv)"

LEASE_RESULT="$(az storage blob lease break \
  --account-name "$TFSTATE_SA" \
  --container-name "$TFSTATE_CONTAINER" \
  --blob-name "$STATE_BLOB" \
  --account-key "$KEY" 2>&1)" && {
    echo "Blob lease broken (remaining lease time: ${LEASE_RESULT}s)"
} || {
    echo "No active blob lease on ${STATE_BLOB} (already clear)"
}

LEASE_STATUS="$(az storage blob show \
  --account-name "$TFSTATE_SA" \
  --container-name "$TFSTATE_CONTAINER" \
  --name "$STATE_BLOB" \
  --account-key "$KEY" \
  --query 'properties.leaseStatus' -o tsv 2>/dev/null || echo unknown)"
echo "Blob lease status after break: ${LEASE_STATUS}"

# Terraform force-unlock for any orphaned .terraform.lock.info entries
# Read lock ID from state blob metadata if present
TF_ENV_DIR="infra/terraform/environments/${ENV_NAME}"
if [[ -d "$TF_ENV_DIR" ]]; then
  LOCK_ID="$(az storage blob metadata show \
    --account-name "$TFSTATE_SA" \
    --container-name "$TFSTATE_CONTAINER" \
    --name "${STATE_BLOB}.tflock" \
    --account-key "$KEY" \
    --query 'ID' -o tsv 2>/dev/null || true)"
  if [[ -n "$LOCK_ID" ]]; then
    echo "== Running terraform force-unlock for lock ID: ${LOCK_ID} =="
    terraform -chdir="$TF_ENV_DIR" force-unlock -force "$LOCK_ID" || true
  fi
fi

echo "Done. You can now run terraform plan/apply for environment: ${ENV_NAME}"
