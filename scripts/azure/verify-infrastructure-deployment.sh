#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Verify a deployed Terraform environment and its AKS platform.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
ENV_NAME="dev"
SUBSCRIPTION_ID="${ARM_SUBSCRIPTION_ID:-}"
TERRAFORM_DIR=""
SKIP_DRIFT=false
TIMEOUT="300s"

usage() {
  cat <<'USAGE'
Usage:
  verify-infrastructure-deployment.sh [options]

Options:
  --env <name>              Environment (dev|test|staging|prod). Default: dev
  --subscription-id <id>    Expected deployment subscription. Defaults to ARM_SUBSCRIPTION_ID.
  --terraform-dir <path>    Environment Terraform root. Defaults to infra/terraform/environments/<env>.
  --skip-drift              Skip the post-apply Terraform drift check.
  --timeout <duration>      kubectl wait timeout. Default: 300s
  -h, --help                Show help.
USAGE
}

fail() {
  echo "VERIFY FAIL: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env)
      ENV_NAME="${2:-}"
      shift 2
      ;;
    --subscription-id)
      SUBSCRIPTION_ID="${2:-}"
      shift 2
      ;;
    --terraform-dir)
      TERRAFORM_DIR="${2:-}"
      shift 2
      ;;
    --skip-drift)
      SKIP_DRIFT=true
      shift
      ;;
    --timeout)
      TIMEOUT="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

case "$ENV_NAME" in
  dev|test|staging|prod) ;;
  *) fail "unsupported environment: $ENV_NAME" ;;
esac

if [[ -z "$TERRAFORM_DIR" ]]; then
  TERRAFORM_DIR="${REPO_ROOT}/infra/terraform/environments/${ENV_NAME}"
elif [[ "$TERRAFORM_DIR" != /* ]]; then
  TERRAFORM_DIR="${PWD}/${TERRAFORM_DIR}"
fi

[[ -d "$TERRAFORM_DIR" ]] || fail "Terraform directory not found: $TERRAFORM_DIR"
[[ -f "${TERRAFORM_DIR}/${ENV_NAME}.tfvars" ]] || fail "tfvars not found: ${TERRAFORM_DIR}/${ENV_NAME}.tfvars"

for command_name in az terraform kubectl kubelogin; do
  require_command "$command_name"
done

kubelogin_version="$(kubelogin --version 2>&1 || true)"
[[ "$kubelogin_version" == *"kubelogin"* ]] || \
  fail "Azure kubelogin is invalid; reinstall it with: az aks install-cli"

az account show >/dev/null 2>&1 || fail "Azure login missing; run az login"

expected_subscription="$(
  terraform -chdir="$TERRAFORM_DIR" console -var-file="${ENV_NAME}.tfvars" \
    <<< 'var.expected_subscription_id' 2>/dev/null | tr -d '"[:space:]'
)"
[[ -n "$expected_subscription" ]] || fail "could not read expected_subscription_id from ${ENV_NAME}.tfvars"

if [[ -z "$SUBSCRIPTION_ID" ]]; then
  SUBSCRIPTION_ID="$(az account show --query id -o tsv)"
fi
[[ "$SUBSCRIPTION_ID" == "$expected_subscription" ]] || \
  fail "subscription mismatch: effective=${SUBSCRIPTION_ID}, expected=${expected_subscription}"

export ARM_SUBSCRIPTION_ID="$SUBSCRIPTION_ID"
az account set --subscription "$SUBSCRIPTION_ID" >/dev/null

echo "Verifying infrastructure: environment=${ENV_NAME} subscription=${SUBSCRIPTION_ID}"

if [[ "$SKIP_DRIFT" == "false" ]]; then
  echo "Checking Terraform drift..."
  set +e
  terraform -chdir="$TERRAFORM_DIR" plan -input=false -lock-timeout=30s \
    -detailed-exitcode -no-color -var-file="${ENV_NAME}.tfvars"
  plan_status=$?
  set -e
  case "$plan_status" in
    0) echo "  Terraform state matches configuration." ;;
    2) fail "Terraform detected changes after deployment" ;;
    *) fail "Terraform drift check failed with exit code ${plan_status}" ;;
  esac
fi

resource_group="$(terraform -chdir="$TERRAFORM_DIR" output -raw resource_group)"
cluster_name="$(terraform -chdir="$TERRAFORM_DIR" output -raw cluster_name)"
[[ -n "$resource_group" && -n "$cluster_name" ]] || fail "required Terraform outputs are empty"

echo "Checking Azure resources..."
rg_state="$(az group show --subscription "$SUBSCRIPTION_ID" --name "$resource_group" \
  --query properties.provisioningState -o tsv)"
[[ "$rg_state" == "Succeeded" ]] || fail "resource group ${resource_group} state is ${rg_state}"

aks_state="$(az aks show --subscription "$SUBSCRIPTION_ID" --resource-group "$resource_group" \
  --name "$cluster_name" --query provisioningState -o tsv)"
aks_power="$(az aks show --subscription "$SUBSCRIPTION_ID" --resource-group "$resource_group" \
  --name "$cluster_name" --query powerState.code -o tsv)"
oidc_enabled="$(az aks show --subscription "$SUBSCRIPTION_ID" --resource-group "$resource_group" \
  --name "$cluster_name" --query oidcIssuerProfile.enabled -o tsv)"
workload_identity_enabled="$(az aks show --subscription "$SUBSCRIPTION_ID" --resource-group "$resource_group" \
  --name "$cluster_name" --query securityProfile.workloadIdentity.enabled -o tsv)"

[[ "$aks_state" == "Succeeded" ]] || fail "AKS ${cluster_name} provisioning state is ${aks_state}"
[[ "$aks_power" == "Running" ]] || fail "AKS ${cluster_name} power state is ${aks_power}"
[[ "$oidc_enabled" == "true" ]] || fail "AKS OIDC issuer is not enabled"
[[ "$workload_identity_enabled" == "true" ]] || fail "AKS workload identity is not enabled"

unhealthy_pools="$(az aks nodepool list --subscription "$SUBSCRIPTION_ID" \
  --resource-group "$resource_group" --cluster-name "$cluster_name" \
  --query "[?provisioningState!='Succeeded' || powerState.code!='Running'].name" -o tsv)"
[[ -z "$unhealthy_pools" ]] || fail "unhealthy AKS node pools: ${unhealthy_pools//$'\n'/, }"

kubeconfig="$(mktemp)"
trap 'rm -f "$kubeconfig"' EXIT
az aks get-credentials --subscription "$SUBSCRIPTION_ID" --resource-group "$resource_group" \
  --name "$cluster_name" --file "$kubeconfig" --overwrite-existing >/dev/null
KUBECONFIG="$kubeconfig" kubelogin convert-kubeconfig -l azurecli

echo "Checking Kubernetes platform health..."
KUBECONFIG="$kubeconfig" kubectl wait --for=condition=Ready nodes --all --timeout="$TIMEOUT"

unhealthy_system_pods="$(
  KUBECONFIG="$kubeconfig" kubectl get pods -n kube-system \
    -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\t"}{range .status.containerStatuses[*]}{.ready}{" "}{end}{"\n"}{end}' |
  awk '$2 != "Running" && $2 != "Succeeded" { print $1 ":" $2; next } $2 == "Running" && $0 ~ /false/ { print $1 ":NotReady" }'
)"
[[ -z "$unhealthy_system_pods" ]] || fail "unhealthy kube-system pods: ${unhealthy_system_pods//$'\n'/, }"

KUBECONFIG="$kubeconfig" kubectl get nodes -o wide
warnings="$(KUBECONFIG="$kubeconfig" kubectl get events -A --field-selector=type=Warning \
  --sort-by='.lastTimestamp' --no-headers 2>/dev/null | tail -n 20 || true)"
if [[ -n "$warnings" ]]; then
  echo "Recent Kubernetes warning events (diagnostic only):"
  echo "$warnings"
fi

echo "VERIFY PASS: infrastructure deployment is healthy for ${ENV_NAME}."
