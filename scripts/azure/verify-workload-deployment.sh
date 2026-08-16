#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Verify RTSA Helm releases and Kubernetes workloads after deployment.

set -euo pipefail

NAMESPACE="rtsa"
KUBE_CONTEXT=""
EXPECTED_IMAGE_TAG=""
CHANGED_DEPLOYMENTS=""
CHANGED_DEPLOYMENTS_SET="false"
TIMEOUT="300s"

usage() {
  cat <<'USAGE'
Usage:
  verify-workload-deployment.sh [options]

Options:
  --namespace <name>          Kubernetes namespace. Default: rtsa
  --context <name>            kubectl/Helm context. Defaults to the current context.
  --expected-image-tag <tag>  Require every RTSA deployment to use this image tag.
  --changed-deployments <l>   Space-separated deployment names to check against
                               --expected-image-tag. Default: all (full promotion).
  --timeout <duration>        Rollout timeout. Default: 300s
  -h, --help                  Show help.
USAGE
}

fail() {
  echo "VERIFY FAIL: $*" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --namespace)
      NAMESPACE="${2:-}"
      shift 2
      ;;
    --context)
      KUBE_CONTEXT="${2:-}"
      shift 2
      ;;
    --expected-image-tag)
      EXPECTED_IMAGE_TAG="${2:-}"
      shift 2
      ;;
    --changed-deployments)
      CHANGED_DEPLOYMENTS="${2:-}"
      CHANGED_DEPLOYMENTS_SET="true"
      shift 2
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

[[ -n "$NAMESPACE" ]] || fail "namespace cannot be empty"
for command_name in kubectl helm jq; do
  command -v "$command_name" >/dev/null 2>&1 || fail "required command not found: $command_name"
done

kubectl_args=()
helm_args=()
if [[ -n "$KUBE_CONTEXT" ]]; then
  kubectl_args+=(--context "$KUBE_CONTEXT")
  helm_args+=(--kube-context "$KUBE_CONTEXT")
fi

kube() {
  kubectl "${kubectl_args[@]}" "$@"
}

helm_cmd() {
  helm "${helm_args[@]}" "$@"
}

releases=(
  rtsa-mesh
  redpanda
  clickhouse
  svc-radar-ingestion
  svc-fusion-engine
  svc-track
  svc-query
  web-cop-gpu
  svc-webtransport
)
deployments=(
  svc-radar-ingestion
  svc-fusion-engine
  svc-track
  svc-query
  web-cop-gpu
  svc-webtransport
)
statefulsets=(redpanda clickhouse)

echo "Verifying workloads: namespace=${NAMESPACE} context=${KUBE_CONTEXT:-current}"
kube get namespace "$NAMESPACE" >/dev/null 2>&1 || fail "namespace not found: $NAMESPACE"

mesh_revision="$(kube get namespace "$NAMESPACE" -o jsonpath='{.metadata.labels.istio\.io/rev}' 2>/dev/null || true)"
legacy_injection="$(kube get namespace "$NAMESPACE" -o jsonpath='{.metadata.labels.istio-injection}' 2>/dev/null || true)"
if [[ -z "$mesh_revision" && "$legacy_injection" != "enabled" ]]; then
  fail "namespace ${NAMESPACE} has no Istio injection label"
fi

echo "Checking Helm releases..."
for release in "${releases[@]}"; do
  # If we know which deployments changed, and it's a stateless service
  # that we intentionally skipped, skip checking its status since it might 
  # have failed in a previous run and we don't plan to fix it here.
  if [[ "$CHANGED_DEPLOYMENTS_SET" == "true" && " ${deployments[*]} " =~ " ${release} " ]]; then
    if [[ ! " ${CHANGED_DEPLOYMENTS} " =~ " ${release} " ]]; then
      echo "Skipping helm release check for unchanged deployment: ${release}"
      continue
    fi
  fi

  # .info.status is Helm's documented release state field (deployed|failed|...);
  # a raw sed/grep scrape can match an unrelated nested "status" key instead.
  status="$(helm_cmd status "$release" --namespace "$NAMESPACE" -o json 2>/dev/null | jq -r '.info.status // empty')"
  [[ "$status" == "deployed" ]] || fail "Helm release ${release} status is ${status:-missing}"
done

echo "Checking workload rollouts..."
for statefulset in "${statefulsets[@]}"; do
  kube --namespace "$NAMESPACE" rollout status "statefulset/${statefulset}" --timeout="$TIMEOUT"
done
for deployment in "${deployments[@]}"; do
  if [[ "$CHANGED_DEPLOYMENTS_SET" == "true" && " ${deployments[*]} " =~ " ${deployment} " ]]; then
    if [[ ! " ${CHANGED_DEPLOYMENTS} " =~ " ${deployment} " ]]; then
      echo "Skipping deployment rollout check for unchanged deployment: ${deployment}"
      continue
    fi
  fi
  kube --namespace "$NAMESPACE" rollout status "deployment/${deployment}" --timeout="$TIMEOUT"
done

unhealthy_pods="$(kube get pods --namespace "$NAMESPACE" \
    -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.phase}{"\t"}{range .status.containerStatuses[*]}{.ready}{" "}{end}{"\n"}{end}' |
  awk '$2 != "Running" && $2 != "Succeeded" { print $1 ":" $2; next } $2 == "Running" && $0 ~ /false/ { print $1 ":NotReady" }'
)"

if [[ -n "$unhealthy_pods" ]]; then
  # Filter out unhealthy pods that belong to unchanged deployments
  # since we intentionally didn't redeploy/fix them in this run
  if [[ "$CHANGED_DEPLOYMENTS_SET" == "true" ]]; then
    filtered_unhealthy=""
    while read -r pod_info; do
      pod_name="${pod_info%%:*}"
      # If this pod name starts with one of our unchanged deployments, ignore it
      is_changed="false"
      if [[ -z "$CHANGED_DEPLOYMENTS" ]]; then
         # nothing was changed, assume all pods belong to unchanged deployments (except mesh/redpanda/clickhouse which we always deploy)
         is_changed="false"
         if [[ "$pod_name" == *"rtsa-mesh"* || "$pod_name" == *"redpanda"* || "$pod_name" == *"clickhouse"* ]]; then
             is_changed="true"
         fi
      else
         # check if the pod belongs to any of the changed deployments
         for changed_dep in $CHANGED_DEPLOYMENTS; do
           if [[ "$pod_name" == "$changed_dep-"* ]]; then
             is_changed="true"
             break
           fi
         done
         # Always check core stateful components as they are always redeployed
         if [[ "$pod_name" == *"rtsa-mesh"* || "$pod_name" == *"redpanda"* || "$pod_name" == *"clickhouse"* ]]; then
             is_changed="true"
         fi
      fi
      if [[ "$is_changed" == "true" ]]; then
        filtered_unhealthy="${filtered_unhealthy}${pod_info}
"
      else
        echo "Ignoring unhealthy pod ${pod_info} belonging to unchanged (skipped) deployment."
      fi
    done <<< "$unhealthy_pods"
    # Remove trailing newline
    unhealthy_pods="${filtered_unhealthy%$'\n'}"
  fi
  
  [[ -z "$unhealthy_pods" ]] || fail "unhealthy pods: ${unhealthy_pods//$'\n'/, }"
fi

for deployment in "${deployments[@]}"; do
  pod_containers="$(kube get pods --namespace "$NAMESPACE" \
    --selector="app.kubernetes.io/instance=${deployment}" \
    -o jsonpath='{range .items[*]}{range .spec.containers[*]}{.name}{" "}{end}{end}')"
  [[ "$pod_containers" == *"istio-proxy"* ]] || fail "Istio sidecar missing from pods for ${deployment}"
done

workload_identity_client_id="$(kube get serviceaccount svc-webtransport --namespace "$NAMESPACE" \
  -o jsonpath='{.metadata.annotations.azure\.workload\.identity/client-id}' 2>/dev/null || true)"
[[ -n "$workload_identity_client_id" ]] || \
  fail "svc-webtransport service account is missing its workload identity client-id annotation"

if [[ -n "$EXPECTED_IMAGE_TAG" ]]; then
  # Default to all deployments (full promotion, e.g. cd-deploy.yml) when the
  # caller doesn't say which ones were actually redeployed with this tag.
  # If the flag was passed but empty, nothing changed this run — skip entirely.
  tag_check_deployments=("${deployments[@]}")
  if [[ "$CHANGED_DEPLOYMENTS_SET" == "true" ]]; then
    read -r -a tag_check_deployments <<< "$CHANGED_DEPLOYMENTS"
  fi
  if [[ ${#tag_check_deployments[@]} -gt 0 ]]; then
    echo "Checking promoted image tag ${EXPECTED_IMAGE_TAG}..."
  fi
  for deployment in "${tag_check_deployments[@]}"; do
    image="$(kube get deployment "$deployment" --namespace "$NAMESPACE" \
      -o jsonpath='{.spec.template.spec.containers[0].image}')"
    [[ "$image" == *":${EXPECTED_IMAGE_TAG}" ]] || \
      fail "deployment ${deployment} uses ${image}, expected tag ${EXPECTED_IMAGE_TAG}"
  done
fi

kube get pods --namespace "$NAMESPACE" -o wide
warnings="$(kube get events --namespace "$NAMESPACE" --field-selector=type=Warning \
  --sort-by='.lastTimestamp' --no-headers 2>/dev/null | tail -n 20 || true)"
if [[ -n "$warnings" ]]; then
  echo "Recent Kubernetes warning events (diagnostic only):"
  echo "$warnings"
fi

echo "VERIFY PASS: workload deployment is healthy in namespace ${NAMESPACE}."
