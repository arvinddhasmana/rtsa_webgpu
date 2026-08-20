#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Create development-only WebTransport JWT and TLS material in Azure Key Vault.

set -euo pipefail

ENV_NAME="dev"
TERRAFORM_DIR="infra/terraform/environments/dev"
VALID_DAYS="30"
FORCE=false
DRY_RUN=false

usage() {
  cat <<'USAGE'
Usage:
  bootstrap-dev-key-vault-secrets.sh [options]

Options:
  --env <name>             Environment name. Only dev is accepted. Default: dev
  --terraform-dir <path>   Initialized Terraform environment root.
  --valid-days <days>      Self-signed TLS certificate lifetime. Default: 30
  --force                  Replace existing secret versions.
  --dry-run                Validate prerequisites without creating secrets.
  --help                   Show help.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env)
      ENV_NAME="${2:-}"
      shift 2
      ;;
    --terraform-dir)
      TERRAFORM_DIR="${2:-}"
      shift 2
      ;;
    --valid-days)
      VALID_DAYS="${2:-}"
      shift 2
      ;;
    --force)
      FORCE=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
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

if [[ "$ENV_NAME" != "dev" ]]; then
  echo "ERROR: this script only creates development secret material." >&2
  exit 1
fi

if [[ ! "$VALID_DAYS" =~ ^[1-9][0-9]*$ ]]; then
  echo "ERROR: --valid-days must be a positive integer." >&2
  exit 2
fi

for command_name in az terraform openssl; do
  command -v "$command_name" >/dev/null 2>&1 || {
    echo "ERROR: required command not found: $command_name" >&2
    exit 1
  }
done

if [[ ! -d "$TERRAFORM_DIR" ]]; then
  echo "ERROR: Terraform directory not found: $TERRAFORM_DIR" >&2
  exit 1
fi

tfvars_path="${TERRAFORM_DIR}/${ENV_NAME}.tfvars"
if [[ ! -f "$tfvars_path" ]]; then
  echo "ERROR: Terraform variables file not found: $tfvars_path" >&2
  exit 1
fi

expected_subscription="$(
  grep -E '^[[:space:]]*expected_subscription_id[[:space:]]*=' "$tfvars_path" |
    sed -E 's/^[^=]*=[[:space:]]*"([^"]*)".*$/\1/'
)"
current_subscription="$(az account show --query id -o tsv)"

if [[ -z "$expected_subscription" || "$current_subscription" != "$expected_subscription" ]]; then
  echo "ERROR: Azure subscription does not match the dev Terraform root." >&2
  echo "  expected: ${expected_subscription:-<empty>}" >&2
  echo "  current:  ${current_subscription:-<empty>}" >&2
  exit 1
fi

key_vault_uri="$(terraform -chdir="$TERRAFORM_DIR" output -raw key_vault_uri)"
resource_group="$(terraform -chdir="$TERRAFORM_DIR" output -raw resource_group)"
key_vault_name="${key_vault_uri#https://}"
key_vault_name="${key_vault_name%%.*}"

if [[ -z "$key_vault_name" || "$key_vault_name" == "$key_vault_uri" ]]; then
  echo "ERROR: invalid key_vault_uri Terraform output." >&2
  exit 1
fi

az keyvault show --resource-group "$resource_group" --name "$key_vault_name" \
  --query id -o tsv >/dev/null

echo "Development Key Vault secret bootstrap"
echo "  environment: $ENV_NAME"
echo "  subscription: $current_subscription"
echo "  resource group: $resource_group"
echo "  key vault: $key_vault_name"

if [[ "$DRY_RUN" == true ]]; then
  echo "DRY RUN PASS: prerequisites validated; no secrets created."
  exit 0
fi

secret_names=(wt-jwt-secret wt-tls-crt wt-tls-key)
existing=()
for secret_name in "${secret_names[@]}"; do
  if az keyvault secret show --vault-name "$key_vault_name" --name "$secret_name" \
    --query id -o tsv >/dev/null 2>&1; then
    existing+=("$secret_name")
  fi
done

if [[ ${#existing[@]} -gt 0 && "$FORCE" != true ]]; then
  echo "SKIP: existing secret material retained (${existing[*]})."
  echo "Use --force only when intentional development credential rotation is required."
  exit 0
fi

secret_dir="$(mktemp -d)"
trap 'rm -rf "$secret_dir"' EXIT
chmod 700 "$secret_dir"
umask 077

openssl rand -base64 48 >"${secret_dir}/jwt"
openssl req -x509 -newkey rsa:3072 -sha256 -nodes \
  -keyout "${secret_dir}/tls.key" \
  -out "${secret_dir}/tls.crt" \
  -days "$VALID_DAYS" \
  -subj "/CN=svc-webtransport.rtsa.svc.cluster.local" >/dev/null 2>&1

az keyvault secret set --vault-name "$key_vault_name" --name wt-jwt-secret \
  --file "${secret_dir}/jwt" --content-type text/plain --output none
az keyvault secret set --vault-name "$key_vault_name" --name wt-tls-crt \
  --file "${secret_dir}/tls.crt" --content-type application/x-pem-file --output none
az keyvault secret set --vault-name "$key_vault_name" --name wt-tls-key \
  --file "${secret_dir}/tls.key" --content-type application/x-pem-file --output none

echo "PASS: development WebTransport secrets are present in $key_vault_name."
