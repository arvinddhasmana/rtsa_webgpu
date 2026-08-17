#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Sync environment-scoped GitHub variables from live Terraform outputs.
#
# Reads outputs from infra/terraform/environments/<env> and updates:
# - AKS_RESOURCE_GROUP
# - AKS_CLUSTER_NAME
# - KEY_VAULT_NAME
# - ACR_NAME
# - WEBTRANSPORT_IDENTITY_CLIENT_ID
# - ISTIO_REVISION (when discoverable)
#
# Optional flags can also set frontend URLs.

set -euo pipefail

REPO_SLUG=""
ENV_NAME="dev"
TF_ENV_DIR=""
BOOTSTRAP_DIR="infra/terraform/bootstrap"
VITE_API_GATEWAY_URL=""
VITE_WEBTRANSPORT_URL=""

usage() {
  cat <<'USAGE'
Usage:
  sync-dev-github-vars-from-tf.sh [options]

Options:
  --repo <owner/repo>            GitHub repository slug (required).
  --environment <name>           Environment name (default: dev).
  --tf-env-dir <path>            Terraform environment directory.
                                 Default: infra/terraform/environments/<environment>
  --bootstrap-dir <path>         Bootstrap terraform directory (fallback source).
                                 Default: infra/terraform/bootstrap
  --vite-api-url <url>           Set VITE_API_GATEWAY_URL for this environment.
  --vite-webtransport-url <url>  Set VITE_WEBTRANSPORT_URL for this environment.
  -h, --help                     Show this help.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo)
      REPO_SLUG="${2:-}"
      shift 2
      ;;
    --environment)
      ENV_NAME="${2:-}"
      shift 2
      ;;
    --tf-env-dir)
      TF_ENV_DIR="${2:-}"
      shift 2
      ;;
    --bootstrap-dir)
      BOOTSTRAP_DIR="${2:-}"
      shift 2
      ;;
    --vite-api-url)
      VITE_API_GATEWAY_URL="${2:-}"
      shift 2
      ;;
    --vite-webtransport-url)
      VITE_WEBTRANSPORT_URL="${2:-}"
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

[[ -n "$REPO_SLUG" ]] || { echo "ERROR: --repo owner/repo is required." >&2; exit 1; }

if [[ -z "$TF_ENV_DIR" ]]; then
  TF_ENV_DIR="infra/terraform/environments/${ENV_NAME}"
fi

command -v gh >/dev/null 2>&1 || { echo "ERROR: gh CLI not found." >&2; exit 1; }
command -v terraform >/dev/null 2>&1 || { echo "ERROR: terraform not found." >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "ERROR: python3 not found." >&2; exit 1; }

gh auth status >/dev/null 2>&1 || {
  echo "ERROR: gh is not authenticated. Run: gh auth login" >&2
  exit 1
}

[[ -d "$TF_ENV_DIR" ]] || { echo "ERROR: terraform env dir not found: $TF_ENV_DIR" >&2; exit 1; }

has_env_outputs=true
if ! terraform -chdir="$TF_ENV_DIR" output -json >/dev/null 2>&1; then
  has_env_outputs=false
fi

has_bootstrap_outputs=true
if ! terraform -chdir="$BOOTSTRAP_DIR" output -json >/dev/null 2>&1; then
  has_bootstrap_outputs=false
fi

if [[ "$has_env_outputs" == "false" && "$has_bootstrap_outputs" == "false" ]]; then
  cat <<EOF >&2
ERROR: Terraform outputs are unavailable in both:
  - ${TF_ENV_DIR}
  - ${BOOTSTRAP_DIR}

Run bootstrap first:
  scripts/azure/bootstrap.sh

Then run Infra Up for ${ENV_NAME} (or local env apply) and re-run this sync script.
EOF
  exit 1
fi

set_env_var() {
  local key="$1" value="$2"
  gh variable set "$key" --repo "${REPO_SLUG}" --env "$ENV_NAME" --body "$value" >/dev/null
}

create_env_if_missing() {
  # GET only needs "Environments: read"; the PUT (create/update protection
  # rules) needs "Administration: write" on the PAT, so skip it when the
  # environment already exists to avoid requiring that broader grant.
  if gh api "repos/${REPO_SLUG}/environments/${ENV_NAME}" >/dev/null 2>&1; then
    return 0
  fi
  gh api --silent -X PUT "repos/${REPO_SLUG}/environments/${ENV_NAME}" >/dev/null
}

create_env_if_missing

resource_group=""
cluster_name=""
key_vault_name=""
acr_name=""
webtransport_identity=""

if [[ "$has_env_outputs" == "true" ]]; then
  resource_group="$(terraform -chdir="$TF_ENV_DIR" output -raw resource_group)"
  cluster_name="$(terraform -chdir="$TF_ENV_DIR" output -raw cluster_name)"
  key_vault_uri="$(terraform -chdir="$TF_ENV_DIR" output -raw key_vault_uri)"
  acr_login_server="$(terraform -chdir="$TF_ENV_DIR" output -raw acr_login_server)"
  workload_ids_json="$(terraform -chdir="$TF_ENV_DIR" output -json workload_identity_client_ids)"

  key_vault_name="$(python3 - <<'PY' "$key_vault_uri"
import re, sys
uri = sys.argv[1].strip()
m = re.match(r'^https://([^./]+)\.', uri)
if not m:
    raise SystemExit("invalid key_vault_uri: %s" % uri)
print(m.group(1))
PY
  )"

  acr_name="${acr_login_server%%.*}"

  webtransport_identity="$(python3 - <<'PY' "$workload_ids_json"
import json, re, sys
m = json.loads(sys.argv[1])
for k in ("webtransport", "svc-webtransport", "svc_webtransport"):
    if k in m and m[k]:
        print(m[k]); raise SystemExit(0)
for k, v in m.items():
    if re.search(r'webtransport', k, re.I) and v:
        print(v); raise SystemExit(0)
raise SystemExit("could not find webtransport identity key in workload_identity_client_ids")
PY
  )"
else
  echo "WARN: No outputs in ${TF_ENV_DIR}; applying bootstrap fallback only."
  if [[ "$has_bootstrap_outputs" == "true" ]]; then
    acr_name="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw acr_name)"
  fi
fi

if [[ -n "$resource_group" ]]; then
  set_env_var "AKS_RESOURCE_GROUP" "$resource_group"
fi
if [[ -n "$cluster_name" ]]; then
  set_env_var "AKS_CLUSTER_NAME" "$cluster_name"
fi
if [[ -n "$key_vault_name" ]]; then
  set_env_var "KEY_VAULT_NAME" "$key_vault_name"
fi
if [[ -n "$acr_name" ]]; then
  set_env_var "ACR_NAME" "$acr_name"
fi
if [[ -n "$webtransport_identity" ]]; then
  set_env_var "WEBTRANSPORT_IDENTITY_CLIENT_ID" "$webtransport_identity"
fi

istio_revision=""
if [[ -n "$resource_group" && -n "$cluster_name" ]] && command -v az >/dev/null 2>&1 && az account show >/dev/null 2>&1; then
  istio_revision="$(az aks show -g "$resource_group" -n "$cluster_name" --query 'serviceMeshProfile.istio.revisions[0]' -o tsv 2>/dev/null || true)"
fi
if [[ -n "$istio_revision" ]]; then
  set_env_var "ISTIO_REVISION" "$istio_revision"
fi

if [[ -n "$VITE_API_GATEWAY_URL" ]]; then
  set_env_var "VITE_API_GATEWAY_URL" "$VITE_API_GATEWAY_URL"
fi
if [[ -n "$VITE_WEBTRANSPORT_URL" ]]; then
  set_env_var "VITE_WEBTRANSPORT_URL" "$VITE_WEBTRANSPORT_URL"
fi

echo "Updated environment variables for ${REPO_SLUG} / ${ENV_NAME}:"
if [[ -n "$resource_group" ]]; then
  echo "  AKS_RESOURCE_GROUP=${resource_group}"
else
  echo "  AKS_RESOURCE_GROUP unchanged (env outputs not available yet)"
fi
if [[ -n "$cluster_name" ]]; then
  echo "  AKS_CLUSTER_NAME=${cluster_name}"
else
  echo "  AKS_CLUSTER_NAME unchanged (env outputs not available yet)"
fi
if [[ -n "$key_vault_name" ]]; then
  echo "  KEY_VAULT_NAME=${key_vault_name}"
else
  echo "  KEY_VAULT_NAME unchanged (env outputs not available yet)"
fi
if [[ -n "$acr_name" ]]; then
  echo "  ACR_NAME=${acr_name}"
else
  echo "  ACR_NAME unchanged"
fi
if [[ -n "$webtransport_identity" ]]; then
  echo "  WEBTRANSPORT_IDENTITY_CLIENT_ID=<set>"
else
  echo "  WEBTRANSPORT_IDENTITY_CLIENT_ID unchanged (env outputs not available yet)"
fi
if [[ -n "$istio_revision" ]]; then
  echo "  ISTIO_REVISION=${istio_revision}"
else
  echo "  ISTIO_REVISION unchanged (not discovered)"
fi
if [[ -n "$VITE_API_GATEWAY_URL" ]]; then
  echo "  VITE_API_GATEWAY_URL=${VITE_API_GATEWAY_URL}"
fi
if [[ -n "$VITE_WEBTRANSPORT_URL" ]]; then
  echo "  VITE_WEBTRANSPORT_URL=${VITE_WEBTRANSPORT_URL}"
fi
