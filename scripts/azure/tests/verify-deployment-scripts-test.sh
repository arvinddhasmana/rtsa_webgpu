#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Command-mocked regression tests for Azure deployment verification scripts.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
FAKE_BIN="$(mktemp -d)"
trap 'rm -rf "$FAKE_BIN"' EXIT

fail() {
  echo "TEST FAIL: $*" >&2
  exit 1
}

cat >"${FAKE_BIN}/terraform" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
args=" $* "
case "$args" in
  *" console "*) echo '"00000000-0000-0000-0000-000000000001"' ;;
  *" plan "*) exit 0 ;;
  *" output -raw resource_group "*) echo "rg-rtsa-dev-cc" ;;
  *" output -raw cluster_name "*) echo "aks-rtsa-dev" ;;
  *" output -raw key_vault_uri "*) echo "https://kv-rtsa-dev.example.vault.azure.net/" ;;
  *) exit 0 ;;
esac
EOF

cat >"${FAKE_BIN}/az" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
args=" $* "
case "$args" in
  *" account show --query id -o tsv "*) echo "00000000-0000-0000-0000-000000000001" ;;
  *" account show "*" --query id "*) echo "00000000-0000-0000-0000-000000000001" ;;
  *" account show "*|*" account set "*) exit 0 ;;
  *" keyvault show "*) echo "/subscriptions/example/resourceGroups/example/providers/Microsoft.KeyVault/vaults/kv-rtsa-dev" ;;
  *" acr repository show-tags "*) echo "sha-test" ;;
  *" group show "*) echo "Succeeded" ;;
  *" aks nodepool list "*) exit 0 ;;
  *" aks get-credentials "*)
    previous=""
    for argument in "$@"; do
      if [[ "$previous" == "--file" ]]; then
        : >"$argument"
        break
      fi
      previous="$argument"
    done
    ;;
  *" aks show "*" --query provisioningState "*) echo "Succeeded" ;;
  *" aks show "*" --query powerState.code "*) echo "Running" ;;
  *" aks show "*" --query oidcIssuerProfile.enabled "*) echo "true" ;;
  *" aks show "*" --query securityProfile.workloadIdentity.enabled "*) echo "true" ;;
  *) exit 0 ;;
esac
EOF

cat >"${FAKE_BIN}/kubelogin" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "--version" ]]; then
  echo "kubelogin version v0.2.8"
fi
exit 0
EOF

cat >"${FAKE_BIN}/helm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ " $* " == *" status "* ]]; then
  if [[ "${TEST_HELM_FAILED:-false}" == "true" ]]; then
    echo '{"info": {"status": "failed"}}'
  else
    echo '{"info": {"status": "deployed"}}'
  fi
fi
EOF

cat >"${FAKE_BIN}/kubectl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
args=" $* "
case "$args" in
  *" wait "*) echo "condition met" ;;
  *" rollout status "*) echo "successfully rolled out" ;;
  *" get namespace "*"istio"*) echo "asm-1-29" ;;
  *" get namespace "*) exit 0 ;;
  *" get serviceaccount svc-webtransport "*) echo "11111111-1111-1111-1111-111111111111" ;;
  *" get deployment web-cop-gpu "*"-o json "*) echo '{"spec":{"template":{"metadata":{"annotations":{"sidecar.istio.io/inject":"false"}}}}}' ;;
  *" get deployment "*"-o json "*) echo '{"spec":{"template":{"metadata":{"annotations":{}}}}}' ;;
  *" get deployment "*) echo "example.azurecr.io/service:sha-test" ;;
  *" get pods "*"--selector="*)
    echo '{"items":[{"spec":{"containers":[{"name":"service"}],"initContainers":[{"name":"istio-proxy"}]}}]}'
    ;;
  *" get pods "*"jsonpath"*) echo $'pod-1\tRunning\ttrue true' ;;
  *" get pods "*"-o json"*)
    echo '{"items":[{"metadata":{"name":"pod-1"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}}]}'
    ;;
  *" get pods "*) echo "pod-1 Running" ;;
  *" get nodes "*) echo "node-1 Ready" ;;
  *" get events "*) exit 0 ;;
  *) exit 0 ;;
esac
EOF

chmod +x "${FAKE_BIN}"/*
export PATH="${FAKE_BIN}:${PATH}"

INFRA_SCRIPT="${REPO_ROOT}/scripts/azure/verify-infrastructure-deployment.sh"
WORKLOAD_SCRIPT="${REPO_ROOT}/scripts/azure/verify-workload-deployment.sh"
SECRET_BOOTSTRAP_SCRIPT="${REPO_ROOT}/scripts/azure/bootstrap-dev-key-vault-secrets.sh"
TF_DIR="${REPO_ROOT}/infra/terraform/environments/dev"
TEST_SUBSCRIPTION="00000000-0000-0000-0000-000000000001"
TEST_TF_DIR="$(mktemp -d)"
trap 'rm -rf "$FAKE_BIN" "$TEST_TF_DIR"' EXIT
cp "${TF_DIR}/dev.tfvars" "${TEST_TF_DIR}/dev.tfvars"
sed -i "s/^expected_subscription_id[[:space:]]*=.*/expected_subscription_id = \"${TEST_SUBSCRIPTION}\"/" \
  "${TEST_TF_DIR}/dev.tfvars"

bash "$INFRA_SCRIPT" --env dev --subscription-id "$TEST_SUBSCRIPTION" \
  --terraform-dir "$TEST_TF_DIR" --timeout 1s >/dev/null || fail "infrastructure happy path"

if bash "$INFRA_SCRIPT" --env dev \
  --subscription-id "00000000-0000-0000-0000-000000000002" \
  --terraform-dir "$TEST_TF_DIR" >/dev/null 2>&1; then
  fail "infrastructure subscription mismatch was accepted"
fi

bash "$WORKLOAD_SCRIPT" --namespace rtsa --expected-image-tag sha-test \
  --timeout 1s >/dev/null || fail "workload happy path"

TEST_HELM_FAILED=true
export TEST_HELM_FAILED
resolved_tag="$(bash "${REPO_ROOT}/scripts/azure/resolve-deploy-tag.sh" \
  --acr acr-test --service svc-webtransport --namespace rtsa 2>/dev/null)"
[[ "$resolved_tag" == "sha-test" ]] || \
  fail "failed Helm release was not scheduled for same-tag recovery"
unset TEST_HELM_FAILED

if bash "$WORKLOAD_SCRIPT" --namespace rtsa --expected-image-tag sha-wrong \
  --timeout 1s >/dev/null 2>&1; then
  fail "incorrect workload image tag was accepted"
fi

bash "$SECRET_BOOTSTRAP_SCRIPT" --env dev --terraform-dir "$TEST_TF_DIR" \
  --dry-run >/dev/null || fail "development secret bootstrap dry run"

if bash "$SECRET_BOOTSTRAP_SCRIPT" --env prod --terraform-dir "$TEST_TF_DIR" \
  --dry-run >/dev/null 2>&1; then
  fail "development secret bootstrap accepted prod"
fi

if find "${REPO_ROOT}/.github/workflows" -mindepth 2 -type f \
  \( -name '*.yml' -o -name '*.yaml' \) | grep -q .; then
  fail "GitHub reusable workflows must be directly under .github/workflows"
fi

while IFS= read -r workflow_path; do
  [[ -f "${REPO_ROOT}/${workflow_path#./}" ]] || \
    fail "local reusable workflow reference does not exist: ${workflow_path}"
done < <(
  grep -Rho 'uses: \./\.github/workflows/[^[:space:]]*' "${REPO_ROOT}/.github/workflows" |
    awk '{print $2}' | sort -u
)

grep -q 'platforms: linux/arm64' \
  "${REPO_ROOT}/.github/workflows/reusable-container.yml" || \
  fail "container workflow must build linux/arm64 images for AKS"

for dockerfile in \
  svc-radar-ingestion/Dockerfile \
  svc-fusion-engine/Dockerfile \
  svc-track/Dockerfile \
  svc-query/Dockerfile \
  svc-webtransport/Dockerfile; do
  grep -q 'ARG TARGETARCH' "${REPO_ROOT}/${dockerfile}" || \
    fail "${dockerfile} must consume BuildKit TARGETARCH"
  if grep -q 'GOARCH=amd64\|grpc_health_probe-linux-amd64' \
    "${REPO_ROOT}/${dockerfile}"; then
    fail "${dockerfile} contains a hard-coded amd64 runtime artifact"
  fi
done

echo "TEST PASS: deployment verification scripts"
