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
  *) exit 0 ;;
esac
EOF

cat >"${FAKE_BIN}/az" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
args=" $* "
case "$args" in
  *" account show "*" --query id "*) echo "00000000-0000-0000-0000-000000000001" ;;
  *" account show "*|*" account set "*) exit 0 ;;
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
  echo '{"info": {"status": "deployed"}}'
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
  *" get deployment "*) echo "example.azurecr.io/service:sha-test" ;;
  *" get pods "*"--selector="*) echo "service istio-proxy" ;;
  *" get pods "*"jsonpath"*) echo $'pod-1\tRunning\ttrue true' ;;
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
TF_DIR="${REPO_ROOT}/infra/terraform/environments/dev"
TEST_SUBSCRIPTION="00000000-0000-0000-0000-000000000001"

bash "$INFRA_SCRIPT" --env dev --subscription-id "$TEST_SUBSCRIPTION" \
  --terraform-dir "$TF_DIR" --timeout 1s >/dev/null || fail "infrastructure happy path"

if bash "$INFRA_SCRIPT" --env dev \
  --subscription-id "00000000-0000-0000-0000-000000000002" \
  --terraform-dir "$TF_DIR" >/dev/null 2>&1; then
  fail "infrastructure subscription mismatch was accepted"
fi

bash "$WORKLOAD_SCRIPT" --namespace rtsa --expected-image-tag sha-test \
  --timeout 1s >/dev/null || fail "workload happy path"

if bash "$WORKLOAD_SCRIPT" --namespace rtsa --expected-image-tag sha-wrong \
  --timeout 1s >/dev/null 2>&1; then
  fail "incorrect workload image tag was accepted"
fi

echo "TEST PASS: deployment verification scripts"
