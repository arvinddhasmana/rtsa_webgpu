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
#   --set-github  After apply, push outputs to GitHub repo/environments via `gh`.
#
# Prereqs: az (logged in), terraform >= 1.9. Optional: gh (for --set-github).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
BOOTSTRAP_DIR="${REPO_ROOT}/infra/terraform/bootstrap"

PLAN_ONLY=false
SET_GITHUB=false
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

echo "== terraform init =="
terraform -chdir="${BOOTSTRAP_DIR}" init -input=false

echo "== terraform plan =="
terraform -chdir="${BOOTSTRAP_DIR}" plan -input=false -out=tfplan.bootstrap

if [[ "${PLAN_ONLY}" == "true" ]]; then
  echo "Plan-only mode — not applying."
  exit 0
fi

echo "== terraform apply =="
terraform -chdir="${BOOTSTRAP_DIR}" apply -input=false tfplan.bootstrap
rm -f "${BOOTSTRAP_DIR}/tfplan.bootstrap"

echo
echo "== Bootstrap outputs =="
terraform -chdir="${BOOTSTRAP_DIR}" output

if [[ "${SET_GITHUB}" == "true" ]]; then
  command -v gh >/dev/null 2>&1 || { echo "ERROR: gh CLI not found (needed for --set-github)." >&2; exit 1; }

  owner="$(terraform -chdir="${BOOTSTRAP_DIR}" output -json github_actions_variables >/dev/null 2>&1; grep -E '^github_owner' "${TFVARS}" | sed -E 's/.*=\s*"?([^" ]+)"?.*/\1/')"
  repo="$(grep -E '^github_repo' "${TFVARS}" | sed -E 's/.*=\s*"?([^" ]+)"?.*/\1/')"
  slug="${owner}/${repo}"

  tenant="$(terraform -chdir="${BOOTSTRAP_DIR}" output -raw tenant_id)"
  sub="$(az account show --query id -o tsv)"
  acr="$(terraform -chdir="${BOOTSTRAP_DIR}" output -raw acr_login_server)"
  tfstate_rg="$(terraform -chdir="${BOOTSTRAP_DIR}" output -raw shared_resource_group)"
  tfstate_sa="$(terraform -chdir="${BOOTSTRAP_DIR}" output -raw state_storage_account_name)"
  tfstate_container="$(terraform -chdir="${BOOTSTRAP_DIR}" output -raw state_container_name)"
  ci_client="$(terraform -chdir="${BOOTSTRAP_DIR}" output -raw ci_plan_identity_client_id)"

  echo "== Setting repo-level variables on ${slug} =="
  gh variable set AZURE_TENANT_ID       --repo "${slug}" --body "${tenant}"
  gh variable set AZURE_SUBSCRIPTION_ID --repo "${slug}" --body "${sub}"
  gh variable set ACR_LOGIN_SERVER      --repo "${slug}" --body "${acr}"
  gh variable set TFSTATE_RG            --repo "${slug}" --body "${tfstate_rg}"
  gh variable set TFSTATE_SA            --repo "${slug}" --body "${tfstate_sa}"
  gh variable set TFSTATE_CONTAINER     --repo "${slug}" --body "${tfstate_container}"
  gh variable set AZURE_CLIENT_ID_CI    --repo "${slug}" --body "${ci_client}"

  echo "== Setting per-environment AZURE_CLIENT_ID (create the GitHub Environments first) =="
  for env in $(terraform -chdir="${BOOTSTRAP_DIR}" output -json deployer_identity_client_ids | python3 -c 'import json,sys; [print(k) for k in json.load(sys.stdin)]'); do
    client_id="$(terraform -chdir="${BOOTSTRAP_DIR}" output -json deployer_identity_client_ids | python3 -c "import json,sys; print(json.load(sys.stdin)['${env}'])")"
    echo "  - ${env}: ${client_id}"
    gh variable set AZURE_CLIENT_ID --repo "${slug}" --env "${env}" --body "${client_id}" || \
      echo "    (create GitHub Environment '${env}' in repo settings, then re-run --set-github)"
  done
fi

echo
echo "Done. Next: 'make env-up ENV=dev' from infra/terraform/ to build the landing zone."
