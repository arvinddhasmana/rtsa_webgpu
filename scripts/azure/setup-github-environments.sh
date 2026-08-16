#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Create GitHub Environments and seed required Actions variables for RTSA Azure CI/CD.
#
# This script is idempotent and safe to re-run.
#
# Usage:
#   scripts/azure/setup-github-environments.sh \
#     --repo arvinddhasmana/rtsa_webgpu \
#     --bootstrap-dir infra/terraform/bootstrap \
#     --environments dev,test,staging,prod \
#     --environment-subscriptions dev=<sub>,test=<sub>,staging=<sub>,prod=<sub>
#
# If --repo is omitted, the script tries bootstrap terraform.tfvars
# (github_owner/github_repo) and then git remote origin.

set -euo pipefail

REPO_SLUG=""
BOOTSTRAP_DIR="infra/terraform/bootstrap"
ENVIRONMENTS="dev,test,staging,prod"
ENV_SUBSCRIPTIONS=""
REPO_SUBSCRIPTION_ID=""
REPO_CLIENT_ID=""
SET_LEGACY_REPO_SUBSCRIPTION_ID=false
BOOTSTRAP_ENV_SUBSCRIPTIONS_JSON=""

usage() {
  cat <<'USAGE'
Usage:
  setup-github-environments.sh [options]

Options:
  --repo <owner/repo>            GitHub repository slug.
  --bootstrap-dir <path>         Bootstrap terraform directory (default: infra/terraform/bootstrap).
  --environments <csv>           Environments to create (default: dev,test,staging,prod).
  --environment-subscriptions <csv>
                                 Environment-specific subscription IDs as key=value CSV.
                                 Example: dev=<sub1>,test=<sub2>,staging=<sub3>,prod=<sub4>
  --repo-subscription-id <id>    Shared build subscription ID (default: bootstrap subscription_id).
  --repo-client-id <id>          Shared build client ID (default: deployer identity for dev, fallback first env).
  --set-legacy-repo-subscription-id
                                 Also set repo-level AZURE_SUBSCRIPTION_ID for compatibility.
  -h, --help                     Show this help.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo)
      REPO_SLUG="${2:-}"
      shift 2
      ;;
    --bootstrap-dir)
      BOOTSTRAP_DIR="${2:-}"
      shift 2
      ;;
    --environments)
      ENVIRONMENTS="${2:-}"
      shift 2
      ;;
    --environment-subscriptions)
      ENV_SUBSCRIPTIONS="${2:-}"
      shift 2
      ;;
    --repo-subscription-id)
      REPO_SUBSCRIPTION_ID="${2:-}"
      shift 2
      ;;
    --repo-client-id)
      REPO_CLIENT_ID="${2:-}"
      shift 2
      ;;
    --set-legacy-repo-subscription-id)
      SET_LEGACY_REPO_SUBSCRIPTION_ID=true
      shift
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

command -v gh >/dev/null 2>&1 || { echo "ERROR: gh CLI not found." >&2; exit 1; }
command -v terraform >/dev/null 2>&1 || { echo "ERROR: terraform not found." >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "ERROR: python3 not found." >&2; exit 1; }

gh auth status >/dev/null 2>&1 || {
  echo "ERROR: gh is not authenticated. Run: gh auth login" >&2
  exit 1
}

infer_repo_from_tfvars() {
  local tfvars="${BOOTSTRAP_DIR}/terraform.tfvars"
  [[ -f "$tfvars" ]] || return 1
  local owner repo
  owner="$(grep -E '^github_owner\s*=' "$tfvars" | sed -E 's/.*=\s*"?([^" ]+)"?.*/\1/' | head -n1)"
  repo="$(grep -E '^github_repo\s*=' "$tfvars" | sed -E 's/.*=\s*"?([^" ]+)"?.*/\1/' | head -n1)"
  [[ -n "$owner" && -n "$repo" ]] || return 1
  printf '%s/%s\n' "$owner" "$repo"
}

infer_repo_from_git() {
  local origin
  origin="$(git config --get remote.origin.url 2>/dev/null || true)"
  [[ -n "$origin" ]] || return 1
  python3 - <<'PY' "$origin"
import re, sys
u = sys.argv[1].strip()
m = re.search(r'github\.com[:/]+([^/]+)/([^/.]+)(?:\.git)?$', u)
if not m:
    raise SystemExit(1)
print(f"{m.group(1)}/{m.group(2)}")
PY
}

if [[ -z "$REPO_SLUG" ]]; then
  REPO_SLUG="$(infer_repo_from_tfvars || true)"
fi
if [[ -z "$REPO_SLUG" ]]; then
  REPO_SLUG="$(infer_repo_from_git || true)"
fi
if [[ -z "$REPO_SLUG" ]]; then
  echo "ERROR: could not determine repo slug. Pass --repo owner/repo." >&2
  exit 1
fi

repo_var_exists() {
  local key="$1"
  gh api "repos/${REPO_SLUG}/actions/variables/${key}" >/dev/null 2>&1
}

set_repo_var() {
  local key="$1" value="$2"
  gh variable set "$key" --repo "${REPO_SLUG}" --body "$value" >/dev/null
}

env_var_exists() {
  local env="$1" key="$2"
  gh api "repos/${REPO_SLUG}/environments/${env}/variables/${key}" >/dev/null 2>&1
}

set_env_var() {
  local env="$1" key="$2" value="$3"
  gh variable set "$key" --repo "${REPO_SLUG}" --env "$env" --body "$value" >/dev/null
}

create_env() {
  local env="$1"
  gh api --silent -X PUT "repos/${REPO_SLUG}/environments/${env}" >/dev/null
}

IFS=',' read -r -a ENV_LIST <<<"$ENVIRONMENTS"

echo "== Creating GitHub environments on ${REPO_SLUG} =="
for env in "${ENV_LIST[@]}"; do
  env="$(echo "$env" | xargs)"
  [[ -n "$env" ]] || continue
  create_env "$env"
  echo "  - ensured environment: $env"
done

if [[ ! -d "$BOOTSTRAP_DIR" ]]; then
  echo "ERROR: bootstrap dir not found: $BOOTSTRAP_DIR" >&2
  exit 1
fi

if ! terraform -chdir="$BOOTSTRAP_DIR" output -json >/dev/null 2>&1; then
  cat <<EOF >&2
ERROR: bootstrap Terraform outputs are not available.
Run bootstrap first, for example:
  cp ${BOOTSTRAP_DIR}/terraform.tfvars.example ${BOOTSTRAP_DIR}/terraform.tfvars
  # edit terraform.tfvars
  scripts/azure/bootstrap.sh
Then re-run this script.
EOF
  exit 1
fi

tenant_id="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw tenant_id)"
acr_name="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw acr_name)"
tfstate_rg="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw shared_resource_group)"
tfstate_sa="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw state_storage_account_name)"
tfstate_container="$(terraform -chdir="$BOOTSTRAP_DIR" output -raw state_container_name)"
if terraform -chdir="$BOOTSTRAP_DIR" output -json environment_subscription_ids >/dev/null 2>&1; then
  BOOTSTRAP_ENV_SUBSCRIPTIONS_JSON="$(terraform -chdir="$BOOTSTRAP_DIR" output -json environment_subscription_ids)"
fi
subscription_id="$(python3 - <<'PY' "$BOOTSTRAP_DIR"
import json, pathlib, subprocess, sys
b = pathlib.Path(sys.argv[1])
out = subprocess.check_output(["terraform", f"-chdir={b}", "output", "-json", "github_actions_variables"], text=True)
obj = json.loads(out)
print(obj["AZURE_SUBSCRIPTION_ID"])
PY
)"

deployer_json="$(terraform -chdir="$BOOTSTRAP_DIR" output -json deployer_identity_client_ids)"

# Resolve shared build subscription/client defaults for CD Build.
repo_subscription_id="${REPO_SUBSCRIPTION_ID:-$subscription_id}"
if [[ -z "$REPO_CLIENT_ID" ]]; then
  REPO_CLIENT_ID="$(python3 - <<'PY' "$deployer_json"
import json, sys
m = json.loads(sys.argv[1])
if "dev" in m and m["dev"]:
    print(m["dev"])
elif m:
    first_key = sorted(m.keys())[0]
    print(m[first_key])
else:
    print("")
PY
)"
fi

if [[ -n "$BOOTSTRAP_ENV_SUBSCRIPTIONS_JSON" && -z "$ENV_SUBSCRIPTIONS" ]]; then
  ENV_SUBSCRIPTIONS="$(python3 - <<'PY' "$BOOTSTRAP_ENV_SUBSCRIPTIONS_JSON"
import json, sys
mapping = json.loads(sys.argv[1])
print(','.join(f"{k}={v}" for k, v in sorted(mapping.items())))
PY
  )"
fi

# Repository-level vars used by all environments.
echo "== Ensuring repository variables =="
for kv in \
  "AZURE_TENANT_ID=$tenant_id" \
  "TFSTATE_SUBSCRIPTION_ID=$subscription_id" \
  "REPO_AZURE_SUBSCRIPTION_ID=$repo_subscription_id" \
  "ACR_NAME=$acr_name" \
  "TFSTATE_RESOURCE_GROUP=$tfstate_rg" \
  "TFSTATE_STORAGE_ACCOUNT=$tfstate_sa" \
  "TFSTATE_CONTAINER=$tfstate_container"
do
  key="${kv%%=*}"
  value="${kv#*=}"
  if repo_var_exists "$key"; then
    echo "  - keep existing repo var: $key"
  else
    set_repo_var "$key" "$value"
    echo "  - set repo var: $key"
  fi
done

if [[ -n "$REPO_CLIENT_ID" ]]; then
  if repo_var_exists "REPO_AZURE_CLIENT_ID"; then
    gh variable delete "REPO_AZURE_CLIENT_ID" --repo "$REPO_SLUG" >/dev/null 2>&1 || true
    echo "  - removed outdated repo var: REPO_AZURE_CLIENT_ID"
  fi
fi

if [[ "$SET_LEGACY_REPO_SUBSCRIPTION_ID" == "true" ]]; then
  set_repo_var "AZURE_SUBSCRIPTION_ID" "$repo_subscription_id"
  echo "  - set legacy repo var: AZURE_SUBSCRIPTION_ID"
fi

# Environment-level vars: AZURE_CLIENT_ID from deployer identity map and placeholders for the rest.
placeholders=(
  AKS_CLUSTER_NAME
  AKS_RESOURCE_GROUP
  KEY_VAULT_NAME
  WEBTRANSPORT_IDENTITY_CLIENT_ID
  VITE_API_GATEWAY_URL
  VITE_WEBTRANSPORT_URL
)

echo "== Ensuring environment variables =="
for env in "${ENV_LIST[@]}"; do
  env="$(echo "$env" | xargs)"
  [[ -n "$env" ]] || continue

  env_subscription_id="$(python3 - <<'PY' "$ENV_SUBSCRIPTIONS" "$env" "$subscription_id"
import sys
csv_map = (sys.argv[1] or "").strip()
env = sys.argv[2].strip()
fallback = sys.argv[3].strip()

mapping = {}
if csv_map:
    for part in csv_map.split(','):
        p = part.strip()
        if not p:
            continue
        if '=' not in p:
            continue
        k, v = p.split('=', 1)
        k = k.strip()
        v = v.strip()
        if k and v:
            mapping[k] = v

print(mapping.get(env, fallback))
PY
)"

  set_env_var "$env" "AZURE_SUBSCRIPTION_ID" "$env_subscription_id"
  echo "  - ${env}: set AZURE_SUBSCRIPTION_ID"

  client_id="$(python3 - <<'PY' "$deployer_json" "$env"
import json, sys
m = json.loads(sys.argv[1])
env = sys.argv[2]
print(m.get(env, ""))
PY
)"

  if [[ -n "$client_id" ]]; then
    set_env_var "$env" "AZURE_CLIENT_ID" "$client_id"
    echo "  - ${env}: set AZURE_CLIENT_ID"
  else
    echo "  - ${env}: WARNING no deployer identity found in bootstrap outputs"
  fi

  for key in "${placeholders[@]}"; do
    if env_var_exists "$env" "$key"; then
      echo "  - ${env}: keep existing ${key}"
    else
      set_env_var "$env" "$key" "__SET_ME__"
      echo "  - ${env}: set placeholder ${key}=__SET_ME__"
    fi
  done

  # Optional. Leaving this missing is valid and avoids gh CLI interactive behavior
  # when attempting to set an empty variable body.
  if env_var_exists "$env" "ISTIO_REVISION"; then
    echo "  - ${env}: keep existing ISTIO_REVISION"
  else
    echo "  - ${env}: leave ISTIO_REVISION unset"
  fi

done

echo
echo "Done. Next step: run infra-up for dev, then sync live dev values:"
echo "  scripts/azure/sync-dev-github-vars-from-tf.sh --repo ${REPO_SLUG} --environment dev"
