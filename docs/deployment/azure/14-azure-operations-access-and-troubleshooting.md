<!-- CLASSIFICATION: UNCLASSIFIED -->

# 14 — Azure Operations, Manual Verification, and Internet Access

> **Parent**: [Azure deployment index](README.md)
> **Classification**: UNCLASSIFIED
> **Audience**: new operators, developers, and deployment maintainers
> **Scope**: dev walking-skeleton deployment on AKS

This guide is the manual operating procedure for checking the Azure platform,
AKS workloads, public endpoints, and deployment failures. It describes the
current implementation, not a future production topology.

## 1. Current Public Access Model

The dev deployment has two different network paths:

| Component           | Kubernetes Service             | Current purpose                          | Protocol and port                   |
| ------------------- | ------------------------------ | ---------------------------------------- | ----------------------------------- |
| Static browser UI   | Istio gateway -> `web-cop-gpu` | Browser downloads the WebGPU application | HTTPS TCP 443                       |
| Live track stream   | `svc-webtransport`             | Browser WebTransport hot path            | UDP 4443, public Azure LoadBalancer |
| WebTransport health | `svc-webtransport`             | Operator health check only               | TCP 8081, public Azure LoadBalancer |
| Internal services   | `svc-*`                        | Service-to-service traffic               | ClusterIP only                      |

`web-cop-gpu` is intentionally `ClusterIP`. This is why this works:

```bash
kubectl port-forward svc/web-cop-gpu -n rtsa 8080:80
```

Port forwarding is an operator tunnel from the local machine. It does not
create an internet endpoint. The intended internet endpoint is the managed
Istio gateway at `4.239.211.124`; it must have an Istio `Gateway`/
`VirtualService` (or Gateway API route) and a TLS certificate configured for
the chosen DNS host before a browser can use it. TLS clients may omit SNI for
a raw IP, so the raw IP is not a browser endpoint. The dev workflow uses:

```text
https://4-239-211-124.sslip.io/
```

Production must set `COP_HOSTNAME` to an approved DNS name and use a trusted
PKI or Key Vault certificate. The dev certificate is self-signed and requires
an explicit browser trust decision.

Obtain the current gateway address with:

```bash
kubectl get svc aks-istio-ingressgateway-external \
  -n aks-istio-ingress \
  -o jsonpath='https://{.status.loadBalancer.ingress[0].ip}{"\n"}'
```

Open the DNS URL only after the gateway route and certificate have been
installed. Prefer the DNS name on the certificate, not the raw IP.

### 1.1 Why `telnet <ip> 443` fails

The current manifests do not create an HTTPS listener on TCP 443. `telnet` is
also TCP-only. A failed connection to `4.239.211.124:443` therefore does not
prove that the pods are unhealthy:

- `web-cop-gpu` has no TCP 443 port in its Service.
- `svc-webtransport` exposes UDP 4443, not TCP 443.
- There is no deployed Ingress, Gateway, certificate issuer, or TLS
  termination resource in the current dev charts.
- `4.239.211.124` may be a stale or unrelated public IP. Always read the
  current address from Kubernetes.

For the current cluster, test the managed gateway on TCP 443 after its route is
configured. For production, use a DNS name and an HTTPS ingress or Azure Front Door /
Application Gateway with a managed certificate. Do not present a bare public
HTTP frontend for classified or production data.

### 1.2 Verify the two public services

```bash
kubectl get svc web-cop-gpu svc-webtransport -n rtsa -o wide

frontend_ip="$(kubectl get svc web-cop-gpu -n rtsa \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"
webtransport_ip="$(kubectl get svc svc-webtransport -n rtsa \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"

echo "frontend:      http://${frontend_ip}/"
echo "webtransport:  ${webtransport_ip}:4443/udp"

curl --fail --silent --show-error --connect-timeout 10 \
  "http://${frontend_ip}/" -o /dev/null

# TCP health endpoint; this is not the WebTransport data path.
curl --fail --silent --show-error --connect-timeout 10 \
  "http://${webtransport_ip}:8081/healthz"

# UDP reachability check, where netcat supports UDP probing.
nc -vzu -w 3 "$webtransport_ip" 4443
```

A browser loading the HTML only proves the static UI path. The browser
Developer Tools Network/Console views must also show successful WebTransport
and API connections. The frontend URLs are compiled into the image through
`VITE_API_GATEWAY_URL` and `VITE_WEBTRANSPORT_URL`; changing Kubernetes
Services does not change those values in an already-built static bundle.

## 2. Required Operator Context

Set the environment once per terminal. Never put tokens, passwords, private
keys, or certificate contents in this guide or in shell history.

```bash
export AZ_SUBSCRIPTION_ID="<subscription-id>"
export AZ_RESOURCE_GROUP="rg-rtsa-dev-cc"
export AKS_CLUSTER="aks-rtsa-dev"
export NAMESPACE="rtsa"

az account set --subscription "$AZ_SUBSCRIPTION_ID"
az account show --query '{subscription:id,tenant:tenantId,user:user.name}' -o yaml
az aks get-credentials \
  --resource-group "$AZ_RESOURCE_GROUP" \
  --name "$AKS_CLUSTER" \
  --overwrite-existing
kubelogin convert-kubeconfig -l azurecli
kubectl config current-context
kubectl get namespace "$NAMESPACE"
```

Confirm that the signed-in identity is authorized for the target cluster. Use
an approved named operator identity; do not use shared credentials.

## 3. Azure Infrastructure Verification

Run the repository verification script first:

```bash
export ARM_SUBSCRIPTION_ID="$AZ_SUBSCRIPTION_ID"
scripts/azure/verify-infrastructure-deployment.sh --env dev
```

Then inspect the Azure resources and recent control-plane activity:

```bash
az aks show \
  --resource-group "$AZ_RESOURCE_GROUP" \
  --name "$AKS_CLUSTER" \
  --query '{state:provisioningState,power:powerState.code,kubernetes: kubernetesVersion, fqdn:fqdn, oidc:oidcIssuerProfile.issuerUrl, workloadIdentity:securityProfile.workloadIdentity.enabled}' \
  -o yaml

az aks nodepool list \
  --resource-group "$AZ_RESOURCE_GROUP" \
  --cluster-name "$AKS_CLUSTER" \
  --query '[].{name:name,state:provisioningState,count:count,vmSize:vmSize,mode:mode}' \
  -o table

az monitor activity-log list \
  --resource-group "$AZ_RESOURCE_GROUP" \
  --max-events 30 \
  --offset 24h \
  --query '[].{time:eventTimestamp,level:level,status:status.value,operation:operationName.value,caller:caller}' \
  -o table
```

Check the associated managed resource group when diagnosing Azure Load
Balancer or public IP issues:

```bash
node_rg="$(az aks show -g "$AZ_RESOURCE_GROUP" -n "$AKS_CLUSTER" \
  --query nodeResourceGroup -o tsv)"
az network public-ip list -g "$node_rg" \
  --query '[].{name:name,ip:ipAddress,state:provisioningState,sku:sku.name}' -o table
```

## 4. Kubernetes Deployment Verification

### 4.1 Helm releases

```bash
helm list -n "$NAMESPACE" -a

for release in rtsa-mesh redpanda clickhouse \
  svc-radar-ingestion svc-fusion-engine svc-track svc-query \
  svc-webtransport web-cop-gpu; do
  helm status "$release" -n "$NAMESPACE" -o json \
    | jq -r '[.name, .info.status, .info.last_deployed] | @tsv'
done
```

Every release should normally report `deployed`. A `failed` release requires
inspection before another upgrade. The deployment workflow retries a failed
stateless release with its current or resolved image tag; it does not silently
ignore a failed release.

### 4.2 Nodes, pods, deployments, and stateful components

```bash
kubectl get nodes -o wide
kubectl get pods -n "$NAMESPACE" -o wide
kubectl get deploy,statefulset,hpa,pdb -n "$NAMESPACE"

kubectl wait --for=condition=Available \
  deployment --all -n "$NAMESPACE" --timeout=300s
kubectl rollout status statefulset/redpanda -n "$NAMESPACE" --timeout=300s
kubectl rollout status statefulset/clickhouse -n "$NAMESPACE" --timeout=300s

scripts/azure/verify-workload-deployment.sh --namespace "$NAMESPACE"
```

For a selective deployment, verify only the changed stateless workloads while
leaving healthy unchanged releases untouched:

```bash
scripts/azure/verify-workload-deployment.sh \
  --namespace "$NAMESPACE" \
  --changed-deployments "web-cop-gpu svc-webtransport"
```

The script checks namespace mesh injection, Helm status, rollouts, pod
readiness, Istio sidecars, workload identity checks, and image tags when an
expected tag is supplied.

### 4.3 Services and endpoints

```bash
kubectl get svc -n "$NAMESPACE" -o wide
kubectl get endpoints -n "$NAMESPACE"
kubectl get endpointslices -n "$NAMESPACE" -o wide

kubectl describe svc web-cop-gpu -n "$NAMESPACE"
kubectl describe svc svc-webtransport -n "$NAMESPACE"
```

For every Service, confirm:

1. `spec.selector` matches pod labels.
2. EndpointSlices contain ready pod IPs.
3. `targetPort` matches a named container port or a numeric port.
4. A public Service has `status.loadBalancer.ingress` populated.
5. Only intended ports are public. Health and metrics should not be used as
   browser application endpoints.

### 4.4 Port-forward services for operator access

Port forwarding temporarily binds a local port to a Kubernetes Service. It
requires an active `kubectl` session and stops when the command exits. Run
each `kubectl port-forward` in its own terminal; keep the terminal open while
using the local endpoint.

```bash
# Static frontend, served locally over HTTP.
kubectl port-forward -n "$NAMESPACE" svc/web-cop-gpu 8080:80
# Open http://127.0.0.1:8080/ in a browser or use:
curl --fail --silent --show-error http://127.0.0.1:8080/ -o /dev/null
```

Other useful internal endpoints can be forwarded in the same way:

```bash
# WebTransport health and metrics; these are TCP endpoints, not WebTransport.
kubectl port-forward -n "$NAMESPACE" svc/svc-webtransport 8081:8081
curl --fail --silent --show-error http://127.0.0.1:8081/healthz

# ClickHouse HTTP and native protocol ports.
kubectl port-forward -n "$NAMESPACE" svc/clickhouse 8123:8123 9000:9000

# Redpanda admin API, schema registry, and HTTP proxy.
kubectl port-forward -n "$NAMESPACE" svc/redpanda 9644:9644 8081:8081 8082:8082
```

Do not expose a port-forward through a public listener or bind it to
`0.0.0.0`. The default `127.0.0.1` binding limits access to the operator's
machine. Port-forwarding Redpanda's Kafka port (`9092`) is not a general
external Kafka access method because broker advertised addresses may remain
cluster-local; use `rpk` inside the broker pod for administrative checks.

### 4.5 Verify and access Redpanda

Redpanda is a single-broker dev StatefulSet with a headless Service. Verify
the pod, Service endpoints, and topic bootstrap Job before diagnosing a
consumer:

```bash
kubectl get statefulset/redpanda pod/redpanda-0 svc/redpanda \
  -n "$NAMESPACE" -o wide
kubectl get endpoints redpanda -n "$NAMESPACE" -o wide
kubectl get job/redpanda-create-topics -n "$NAMESPACE" 2>/dev/null || true
topic_job_pod="$(kubectl get pods -n "$NAMESPACE" \
  -l app.kubernetes.io/name=redpanda,app.kubernetes.io/component=topic-bootstrap \
  -o jsonpath='{.items[0].metadata.name}')"
if [[ -n "$topic_job_pod" ]]; then
  kubectl logs -n "$NAMESPACE" "$topic_job_pod" -c create-topics --tail=200
else
  echo "Topic hook pod is absent; verify the live topic list below."
fi
```

Use the Redpanda image's `rpk` client from inside the broker pod. This avoids
exposing Kafka outside the cluster and uses the same broker address as the
topic bootstrap Job:

```bash
broker="redpanda.${NAMESPACE}.svc.cluster.local:9092"

kubectl exec -n "$NAMESPACE" redpanda-0 -- \
  rpk cluster info --brokers "$broker"
kubectl exec -n "$NAMESPACE" redpanda-0 -- \
  rpk topic list --brokers "$broker"

# Inspect partitions, replicas, leaders, and retention for one topic.
kubectl exec -n "$NAMESPACE" redpanda-0 -- \
  rpk topic describe sensors.radar.tracks --brokers "$broker"
kubectl exec -n "$NAMESPACE" redpanda-0 -- \
  rpk topic describe audit.events --brokers "$broker"
```

The expected topic set is declared in
`deploy/charts/redpanda-dev/topics.json`. The manifest includes sensor topics,
fused-track topics, anomaly alerts, operator feedback, audit events, NATO
exchange topics, and model topics. Compare the live list with that manifest
after a Helm upgrade; a topic missing from the list is a bootstrap failure,
not a reason to create an ad-hoc production topic.

Check consumer groups and lag when a downstream service is Ready but is not
processing events:

```bash
kubectl exec -n "$NAMESPACE" redpanda-0 -- \
  rpk group list --brokers "$broker"
kubectl exec -n "$NAMESPACE" redpanda-0 -- \
  rpk group describe <consumer-group> --brokers "$broker"
```

Use topic reads only for controlled synthetic troubleshooting. Do not print
raw sensor, track, audit, or operator data into a terminal transcript or
incident ticket. Prefer metadata commands such as `topic describe`, group
lag, and message counts.

The Redpanda HTTP endpoints are available locally when the port-forward in
Section 4.4 is running:

```bash
curl --fail --silent http://127.0.0.1:9644/v1/status/ready
curl --fail --silent http://127.0.0.1:9644/v1/brokers | jq
curl --fail --silent http://127.0.0.1:8081/subjects | jq
```

The admin API and schema registry are operator endpoints. Do not make them
public or use them as application data paths.

### 4.6 Verify and access ClickHouse

ClickHouse is a single-node dev StatefulSet. Confirm the pod, Service
endpoints, and schema migration Job before querying application tables:

```bash
kubectl get statefulset/clickhouse pod/clickhouse-0 svc/clickhouse \
  -n "$NAMESPACE" -o wide
kubectl get endpoints clickhouse -n "$NAMESPACE" -o wide
kubectl get job -n "$NAMESPACE" \
  -l app.kubernetes.io/name=clickhouse,app.kubernetes.io/component=schema-migration \
  2>/dev/null || true
migration_pod="$(kubectl get pods -n "$NAMESPACE" \
  -l app.kubernetes.io/name=clickhouse,app.kubernetes.io/component=schema-migration \
  -o jsonpath='{.items[0].metadata.name}')"
if [[ -n "$migration_pod" ]]; then
  kubectl logs -n "$NAMESPACE" "$migration_pod" -c migrate --tail=200
else
  echo "Schema migration hook pod is absent; verify the rtsa schema below."
fi
```

The HTTP interface is the simplest operator path. After forwarding port 8123,
use `--data-binary` for SQL and restrict results to metadata or aggregate
values:

```bash
kubectl port-forward -n "$NAMESPACE" svc/clickhouse 8123:8123

curl --fail --silent --show-error http://127.0.0.1:8123/
curl --fail --silent --show-error 'http://127.0.0.1:8123/?query=SELECT%201'

curl --fail --silent --show-error http://127.0.0.1:8123/ \
  --data-binary "SELECT name, engine FROM system.tables WHERE database = 'rtsa' ORDER BY name FORMAT PrettyCompact"
curl --fail --silent --show-error http://127.0.0.1:8123/ \
  --data-binary "SELECT name, total_rows, total_bytes FROM system.tables WHERE database = 'rtsa' ORDER BY name FORMAT PrettyCompact"
```

The initial schema creates the `rtsa` database and these core tables:
`sensor_observations`, `tracks_fused`, `anomaly_detections`, `audit_log`, and
`operator_feedback`. The migration also creates materialized views for active
tracks, sensor throughput, and alert acknowledgement latency. Verify their
definitions without dumping event data:

```bash
curl --fail --silent --show-error http://127.0.0.1:8123/ \
  --data-binary "SELECT name, engine FROM system.tables WHERE database = 'rtsa' AND (name LIKE 'mv_%' OR name NOT LIKE 'mv_%') ORDER BY name FORMAT PrettyCompact"

curl --fail --silent --show-error http://127.0.0.1:8123/ \
  --data-binary "DESCRIBE TABLE rtsa.tracks_fused FORMAT PrettyCompact"
curl --fail --silent --show-error http://127.0.0.1:8123/ \
  --data-binary "SELECT table, sum(rows) AS rows, formatReadableSize(sum(bytes_on_disk)) AS disk FROM system.parts WHERE database = 'rtsa' AND active GROUP BY table ORDER BY table FORMAT PrettyCompact"
```

For the native ClickHouse protocol, use the client from the ClickHouse pod;
this avoids requiring a local `clickhouse-client` installation:

```bash
kubectl exec -n "$NAMESPACE" clickhouse-0 -- \
  clickhouse-client --query "SELECT version()"
kubectl exec -n "$NAMESPACE" clickhouse-0 -- \
  clickhouse-client --query "SHOW DATABASES"
kubectl exec -n "$NAMESPACE" clickhouse-0 -- \
  clickhouse-client --query "SHOW TABLES FROM rtsa"
kubectl exec -n "$NAMESPACE" clickhouse-0 -- \
  clickhouse-client --query "SELECT name, engine FROM system.tables WHERE database = 'rtsa' ORDER BY name FORMAT PrettyCompact"
```

If a table is missing, inspect the migration Job and ClickHouse logs before
manually applying SQL. Schema changes must remain versioned in
`deploy/charts/clickhouse-dev/migrations/` and be reviewed as deployment
changes. Avoid unrestricted `SELECT *` in operational sessions because event
and audit records may be classified at runtime.

## 5. Manual Deployment and Rollback

### 5.1 Deploy the current image set

Use the normal GitHub Actions workflow for approved deployments. For a
controlled manual Helm operation, supply a real immutable ACR tag and the
required environment-specific values:

```bash
export ACR_NAME="<acr-name>"
export IMAGE_TAG="<full-commit-sha>"
export KEY_VAULT_NAME="<key-vault-name>"
export TENANT_ID="<tenant-id>"
export WEBTRANSPORT_IDENTITY_CLIENT_ID="<client-id>"

helm upgrade --install web-cop-gpu deploy/charts/rtsa-service \
  -f deploy/charts/values/web-cop-gpu.yaml \
  --namespace "$NAMESPACE" \
  --set image.repository="${ACR_NAME}.azurecr.io/web-cop-gpu" \
  --set image.tag="$IMAGE_TAG" \
  --wait --timeout 10m
```

Do not use `latest` for an operational deployment. Confirm that the image
exists before changing the release:

```bash
az acr repository show-manifests \
  --name "$ACR_NAME" \
  --repository web-cop-gpu \
  --query "[?tags[?@ == '$IMAGE_TAG']].{digest:digest,tags:tags}" -o json
```

### 5.2 Roll back a failed release

```bash
helm history web-cop-gpu -n "$NAMESPACE"
helm rollback web-cop-gpu <revision> -n "$NAMESPACE" --wait --timeout 10m
kubectl rollout status deployment/web-cop-gpu -n "$NAMESPACE" --timeout=300s
scripts/azure/verify-workload-deployment.sh \
  --namespace "$NAMESPACE" \
  --changed-deployments web-cop-gpu
```

Record the failed revision, error, rollback revision, and verification result
in the deployment incident or change record.

## 6. Logs, Events, and Failure Diagnosis

Start with events, then the workload description, then logs. Capture only
synthetic or approved operational information in tickets.

```bash
kubectl get events -n "$NAMESPACE" --sort-by='.lastTimestamp' | tail -80

kubectl describe deployment web-cop-gpu -n "$NAMESPACE"
kubectl describe pods -n "$NAMESPACE" \
  -l app.kubernetes.io/instance=web-cop-gpu

kubectl logs -n "$NAMESPACE" \
  -l app.kubernetes.io/instance=web-cop-gpu \
  -c web-cop-gpu --tail=200 --prefix

kubectl logs -n "$NAMESPACE" \
  -l app.kubernetes.io/instance=web-cop-gpu \
  -c istio-proxy --tail=100 --prefix
```

Common symptoms:

| Symptom                                 | First checks                                                                                                            |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `EXTERNAL-IP` is `<pending>`            | `kubectl describe svc`; Azure activity log; managed resource group public IP quota and provisioning errors              |
| Browser gets connection refused         | Current Service IP, Service type, Service ports, Network Security Group, and whether the URL uses the declared TCP port |
| Browser gets HTTP 404                   | Nginx pod logs, endpoint readiness, and SPA fallback in the frontend image                                              |
| Frontend loads but live data is offline | Compiled `VITE_WEBTRANSPORT_URL`, UDP 4443 reachability, certificate/JWT/origin configuration, and browser console      |
| `ImagePullBackOff`                      | Image tag exists in ACR, AKS identity can pull ACR, pod events, and architecture (`arm64`)                              |
| `CrashLoopBackOff`                      | Container logs, previous logs with `--previous`, SecretProviderClass, and required environment variables                |
| Helm release is `failed`                | `helm status`, `helm history`, pod events, and the first failed resource in the Helm output                             |
| Pod is `Pending`                        | Node selectors, taints/tolerations, node-pool capacity, and cluster autoscaler events                                   |
| Pod is `Running` but not Ready          | Readiness path/port, sidecar status, application logs, and `kubectl describe pod`                                       |

For a pod that restarted, always inspect the previous container instance:

```bash
kubectl logs -n "$NAMESPACE" <pod-name> -c <container-name> --previous --tail=200
kubectl get pod <pod-name> -n "$NAMESPACE" -o json \
  | jq '.status.containerStatuses'
```

## 7. Network and Mesh Checks

```bash
kubectl get namespace "$NAMESPACE" --show-labels
kubectl get peerauthentication,authorizationpolicy,sidecar -n "$NAMESPACE"

# Test the frontend from inside the cluster.
kubectl run netcheck -n "$NAMESPACE" --rm -i --restart=Never \
  --image=curlimages/curl:8.10.1 -- \
  curl -fsS http://web-cop-gpu/

# Test the WebTransport health endpoint from inside the cluster.
kubectl run netcheck-health -n "$NAMESPACE" --rm -i --restart=Never \
  --image=curlimages/curl:8.10.1 -- \
  curl -fsS http://svc-webtransport:8081/healthz
```

The static `web-cop-gpu` edge frontend is intentionally excluded from sidecar
injection. It is public HTTP in dev and must be protected by an HTTPS gateway
and application authentication before production use. Internal Go services
remain inside the STRICT-mTLS mesh.

If in-cluster requests fail while pods are Ready, inspect the namespace
`PeerAuthentication`, `AuthorizationPolicy`, and `Sidecar` resources. The
current baseline uses STRICT mTLS and allows traffic from selected namespaces;
adding a public LoadBalancer to a mesh-injected service does not bypass
application authentication.

## 8. HTTPS and Production Access

The dev chart intentionally provides HTTP only so the static bundle can be
validated quickly. A production browser endpoint must use:

1. A DNS name, for example `cop.<approved-domain>`.
2. An HTTPS-capable ingress or Azure Front Door/Application Gateway.
3. A certificate managed by the approved enterprise PKI or Azure-managed
   certificate service.
4. Authentication and authorization before classified data is returned.
5. A separate WebTransport endpoint with a certificate whose name matches the
   browser URL and UDP/443 or the approved UDP port.
6. Network controls that do not expose admin, metrics, Kubernetes API, or
   internal service ports.

Do not solve production HTTPS by opening TCP 443 on `svc-webtransport`: that
Service carries UDP WebTransport on 4443 and is not an HTTP reverse proxy.
Likewise, do not expose Envoy's admin port 9901 publicly.

## 9. Evidence Collection for an Incident

Create a timestamped, unclassified evidence directory outside the repository
when possible:

```bash
evidence="/tmp/rtsa-incident-$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$evidence"

kubectl get nodes -o wide > "$evidence/nodes.txt"
kubectl get all -n "$NAMESPACE" -o wide > "$evidence/workloads.txt"
kubectl get svc,endpoints,endpointslices -n "$NAMESPACE" -o yaml > "$evidence/network.yaml"
kubectl get events -n "$NAMESPACE" --sort-by='.lastTimestamp' > "$evidence/events.txt"
helm list -n "$NAMESPACE" -a > "$evidence/helm.txt"

az aks show -g "$AZ_RESOURCE_GROUP" -n "$AKS_CLUSTER" -o json \
  > "$evidence/aks.json"
```

Review evidence for secrets, tokens, certificate material, personal
identifiers, and classified operational data before sharing it. Redact or
delete anything outside the repository's UNCLASSIFIED handling boundary.

## 10. Escalation Order

1. Confirm the symptom and exact URL, port, protocol, and current public IP.
2. Check Azure resource health and activity log.
3. Check Service type, ports, selectors, and EndpointSlices.
4. Check pod readiness and rollout status.
5. Check Helm status/history and deployment events.
6. Check application and sidecar logs.
7. Check mesh policies, certificates, identity, and DNS.
8. Roll back the last known-bad Helm revision when user impact continues.
9. Preserve sanitized evidence and record the remediation.

The existing repository scripts remain the preferred repeatable gates:

```bash
scripts/azure/verify-infrastructure-deployment.sh --env dev
scripts/azure/verify-workload-deployment.sh --namespace rtsa
```
