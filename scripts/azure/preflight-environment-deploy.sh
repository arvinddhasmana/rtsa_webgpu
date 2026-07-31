#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Validate local prerequisites for deploying one Terraform environment root.
#
# Checks:
# - environment -> subscription mapping from bootstrap outputs
# - ARM_SUBSCRIPTION_ID and az account context alignment
# - required Azure provider registration in target subscription
# - shared backend container data-plane access
# - local Terraform root existence
#
# Usage:
#   scripts/azure/preflight-environment-deploy.sh --env dev
#   scripts/azure/preflight-environment-deploy.sh --env staging --bootstrap-dir infra/terraform/bootstrap

set -euo pipefail

ENV_NAME="dev"
BOOTSTRAP_DIR="infra/terraform/bootstrap"
ENV_ROOT_BASE="infra/terraform/environments"

usage() {
  cat <<'USAGE'
Usage:
  preflight-environment-deploy.sh [options]

Options:
  --env <name>            Environment name (dev|test|staging|prod). Default: dev
  --bootstrap-dir <path>  Bootstrap Terraform directory. Default: infra/terraform/bootstrap
  --help                  Show help.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env)
      ENV_NAME="${2:-}"
      shift 2
      ;;
    --bootstrap-dir)
      BOOTSTRAP_DIR="${2:-}"
      shift 2
      ;;
    --help|-h)
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

command -v az >/dev/null 2>&1 || { echo "ERROR: az CLI not found." >&2; exit 1; }
command -v terraform >/dev/null 2>&1 || { echo "ERROR: terraform not found." >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "ERROR: python3 not found." >&2; exit 1; }

if ! az account show >/dev/null 2>&1; then
  echo "ERROR: Azure login missing. Run: az login" >&2
  exit 1
fi

if [[ ! -d "$BOOTSTRAP_DIR" ]]; then
  echo "ERROR: bootstrap dir not found: $BOOTSTRAP_DIR" >&2
  exit 1
fi

if ! terraform -chdir="$BOOTSTRAP_DIR" output -json >/dev/null 2>&1; then
  echo "ERROR: bootstrap outputs unavailable. Run bootstrap first." >&2
  exit 1
fi

ENV_ROOT_DIR="${ENV_ROOT_BASE}/${ENV_NAME}"
if [[ ! -d "$ENV_ROOT_DIR" ]]; then
  echo "ERROR: environment root not found: $ENV_ROOT_DIR" >&2
  exit 1
fi

shared_sub="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw subscription_id)"
tenant_id="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw tenant_id)"
state_rg="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw shared_resource_group)"
state_sa="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw state_storage_account_name)"
state_container="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw state_container_name)"

env_sub="$(terraform -chdir="$BOOTSTRAP_DIR" output -json environment_subscription_ids | \
python3 -c 'import json,sys; m=json.load(sys.stdin); print(m.get(sys.argv[1], ""))' "$ENV_NAME")"

if [[ -z "$env_sub" ]]; then
  echo "ERROR: no subscription mapping found for environment: $ENV_NAME" >&2
  exit 1
fi

arm_sub="${ARM_SUBSCRIPTION_ID:-}"
if [[ -z "$arm_sub" ]]; then
  arm_sub="$(az account show --query id -o tsv)"
fi

current_sub="$(az account show --query id -o tsv)"

echo "Preflight: environment=$ENV_NAME"
echo "  mapped env subscription: $env_sub"
echo "  shared backend subscription: $shared_sub"
echo "  ARM_SUBSCRIPTION_ID effective: $arm_sub"
echo "  az account current subscription: $current_sub"
echo "  tenant: $tenant_id"

if [[ "$arm_sub" != "$env_sub" ]]; then
  echo "BLOCKER: deployment subscription mismatch." >&2
  echo "Set ARM_SUBSCRIPTION_ID=$env_sub before terraform plan/apply." >&2
  exit 1
fi

providers=(
  Microsoft.Compute
  Microsoft.ContainerService
  Microsoft.Network
  Microsoft.Storage
  Microsoft.KeyVault
  Microsoft.ManagedIdentity
  Microsoft.OperationalInsights
)

echo "Checking required providers in target subscription..."
az account set --subscription "$env_sub" >/dev/null
not_registered=0
for p in "${providers[@]}"; do
  state="$(az provider show --namespace "$p" --query registrationState -o tsv 2>/dev/null || echo Unknown)"
  printf "  - %s: %s\n" "$p" "$state"
  if [[ "$state" != "Registered" ]]; then
    not_registered=1
  fi
done

if [[ "$not_registered" -ne 0 ]]; then
  echo "BLOCKER: one or more required providers are not Registered in subscription $env_sub." >&2
  exit 1
fi

echo "Checking tfstate backend access in shared subscription..."
az account set --subscription "$shared_sub" >/dev/null
exists_json="$(az storage container exists --name "$state_container" --account-name "$state_sa" --auth-mode login -o json 2>/dev/null || true)"
if [[ -z "$exists_json" ]]; then
  echo "BLOCKER: unable to query tfstate container existence with Azure AD auth." >&2
  exit 1
fi

container_exists="$(python3 -c 'import json,sys; print(str(json.loads(sys.argv[1]).get("exists", False)).lower())' "$exists_json")"
if [[ "$container_exists" != "true" ]]; then
  echo "BLOCKER: tfstate container does not exist: $state_container" >&2
  exit 1
fi

if ! az storage blob list --container-name "$state_container" --account-name "$state_sa" --auth-mode login --num-results 1 -o none >/dev/null 2>&1; then
  echo "BLOCKER: tfstate backend data-plane access denied (Storage Blob Data Contributor likely missing or not propagated)." >&2
  exit 1
fi

echo "Preflight PASS: local deployment path is ready."
