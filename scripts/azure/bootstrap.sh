#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# One-time RTSA bootstrap: provisions the shared foundation (state storage, ACR,
# Log Analytics) and the GitHub OIDC deployer identities. Run by an admin with
# Owner (or Contributor + User Access Administrator) on the target subscription.
#
# Usage:
#   scripts/azure/bootstrap.sh [--plan-only] [--set-github]
#
#   --plan-only   Run terraform plan and stop (no changes applied).
#   --set-github  After apply, seed GitHub repo/environment variables via the
#                 canonical setup script.
#
# Prereqs: az (logged in), terraform >= 1.9. Optional: gh (for --set-github).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
BOOTSTRAP_DIR="${REPO_ROOT}/infra/terraform/bootstrap"

PLAN_ONLY=false
SET_GITHUB=false
LOCK_TIMEOUT="${TF_LOCK_TIMEOUT:-30s}"
for arg in "$@"; do
  case "$arg" in
    --plan-only) PLAN_ONLY=true ;;
    --set-github) SET_GITHUB=true ;;
    *) echo "Unknown argument: $arg" >&2; exit 2 ;;
  esac
done

command -v az >/dev/null 2>&1 || { echo "ERROR: az CLI not found." >&2; exit 1; }
command -v terraform >/dev/null 2>&1 || { echo "ERROR: terraform not found." >&2; exit 1; }

if ! az account show >/dev/null 2>&1; then
  echo "ERROR: not logged in. Run 'az login' first." >&2
  exit 1
fi

TFVARS="${BOOTSTRAP_DIR}/terraform.tfvars"
if [[ ! -f "${TFVARS}" ]]; then
  echo "ERROR: ${TFVARS} not found." >&2
  echo "       cp ${BOOTSTRAP_DIR}/terraform.tfvars.example ${TFVARS} and edit it." >&2
  exit 1
fi

# The bootstrap stack intentionally uses local state for first run. If an earlier
# terraform process was interrupted, the local lock can remain and make future
# runs appear stuck at plan/apply.
LOCK_FILE="${BOOTSTRAP_DIR}/.terraform.tfstate.lock.info"
if [[ -f "${LOCK_FILE}" ]]; then
  if pgrep -af terraform | grep -F "${BOOTSTRAP_DIR}" >/dev/null 2>&1; then
    echo "ERROR: Terraform appears to be running in ${BOOTSTRAP_DIR}." >&2
    echo "       Wait for that run to finish or stop it before re-running bootstrap." >&2
    exit 1
  fi
  echo "WARN: Found stale local Terraform lock file; removing ${LOCK_FILE}."
  rm -f "${LOCK_FILE}"
fi

# Remove stale plan output from interrupted runs to avoid confusion.
rm -f "${BOOTSTRAP_DIR}/tfplan.bootstrap"

echo "== terraform init =="
terraform -chdir="${BOOTSTRAP_DIR}" init -input=false

echo "== terraform plan =="
echo "Using state lock timeout: ${LOCK_TIMEOUT}"
echo "Note: first run can take several minutes while Azure provider registration settles."
if ! terraform -chdir="${BOOTSTRAP_DIR}" plan -input=false -lock-timeout="${LOCK_TIMEOUT}" -out=tfplan.bootstrap; then
  echo "ERROR: terraform plan failed." >&2
  echo "       If you see a state-lock error, ensure no other Terraform run is active" >&2
  echo "       and remove stale local lock: rm -f ${LOCK_FILE}" >&2
  exit 1
fi

if [[ "${PLAN_ONLY}" == "true" ]]; then
  echo "Plan-only mode — not applying."
  exit 0
fi

echo "== terraform apply =="
if ! terraform -chdir="${BOOTSTRAP_DIR}" apply -input=false -lock-timeout="${LOCK_TIMEOUT}" tfplan.bootstrap; then
  echo "ERROR: terraform apply failed." >&2
  echo "       If you see a state-lock error, ensure no other Terraform run is active" >&2
  echo "       and remove stale local lock: rm -f ${LOCK_FILE}" >&2
  exit 1
fi
rm -f "${BOOTSTRAP_DIR}/tfplan.bootstrap"

echo
echo "== Bootstrap outputs =="
terraform -chdir="${BOOTSTRAP_DIR}" output

if [[ "${SET_GITHUB}" == "true" ]]; then
  command -v gh >/dev/null 2>&1 || { echo "ERROR: gh CLI not found (needed for --set-github)." >&2; exit 1; }
  owner="$(grep -E '^github_owner' "${TFVARS}" | sed -E 's/.*=\s*"?([^" ]+)"?.*/\1/')"
  repo="$(grep -E '^github_repo' "${TFVARS}" | sed -E 's/.*=\s*"?([^" ]+)"?.*/\1/')"
  if [[ -z "${owner}" || -z "${repo}" ]]; then
    echo "ERROR: could not read github_owner/github_repo from ${TFVARS}." >&2
    exit 1
  fi

  echo "== Seeding GitHub environments and variables =="
  "${REPO_ROOT}/scripts/azure/setup-github-environments.sh" \
    --repo "${owner}/${repo}"
fi

echo
echo "Done. Next: 'make env-up ENV=dev' from infra/terraform/ to build the landing zone."
