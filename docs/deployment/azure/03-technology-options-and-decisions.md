<!-- CLASSIFICATION: UNCLASSIFIED -->

# 03 — Technology Options & Decisions

> **Parent**: [README](README.md) · **Prev**: [02 Target Architecture](02-target-architecture.md) · **Next**: [04 Resiliency & WAF](04-resiliency-and-well-architected.md)
> **Classification**: UNCLASSIFIED · **Status**: DRAFT FOR REVIEW — **these are the choices to finalize**

---

## How to use this document

Each section presents options with trade-offs, **licensing** (must be **no upfront cost**),
**PAYG cost signal**, and a **recommendation**. All recommendations are **confirmed ✅**; per user direction DL-08 is set to **Redpanda OSS in all environments**.
Confirmed items (CI/CD = GitHub Actions, IaC = Terraform) are not re-litigated here.

Scoring legend: ●●● strong · ●●○ moderate · ●○○ weak.

---

## 1. Orchestrator (AKS SKU)

| Criterion                                         | AKS **Standard** ✅            | AKS Automatic      |
| ------------------------------------------------- | ------------------------------ | ------------------ |
| Control over node pools (spot, taints, bulkheads) | ●●●                            | ●○○ (NAP-managed)  |
| Scale-to-zero user pools for teardown/cost        | ●●●                            | ●●○                |
| Best-practice defaults out of the box             | ●●○                            | ●●●                |
| Ops overhead                                      | ●●○                            | ●●●                |
| Enterprise reuse / portability                    | ●●●                            | ●●○                |
| Licensing                                         | Free control plane (Free tier) | Free control plane |

**Recommendation ✅ (DL-07): AKS Standard.** Your requirements — spot pools, explicit
bulkhead node pools, and aggressive scale-to-zero for ephemeral cost control — need the
fine-grained control Standard provides. Adopt **AKS Automatic** later for enterprise
turnkey clusters once the Terraform baseline is proven. (Use the **Free** control-plane tier
for non-prod; **Standard** tier with uptime SLA only for prod.)

---

## 2. Event Backbone

> Constraint: **no upfront licensing**. Redpanda Community is source-available (BSL, free to
> run); ClickHouse is Apache-2. Azure Event Hubs is pure consumption.

| Criterion                             | **Redpanda OSS on AKS**   | **Azure Event Hubs** (Kafka API)               | Confluent/Redpanda Cloud |
| ------------------------------------- | ------------------------- | ---------------------------------------------- | ------------------------ |
| Fidelity to current code              | ●●● (drop-in)             | ●●○ (Kafka API compat; no Redpanda transforms) | ●●●                      |
| Wasm data transforms (anti-poisoning) | ●●● native                | ●○○ (must move to a service)                   | ●●○                      |
| Ops burden                            | ●○○ (operator, disks, RF) | ●●● (managed)                                  | ●●●                      |
| Ephemeral spin-up/down                | ●●○                       | ●●● (create/delete namespace)                  | ●●○                      |
| PAYG cost (idle)                      | node + disk cost          | ●●● low (Basic/Standard TU)                    | subscription             |
| Licensing (upfront)                   | **None** (BSL, free)      | **None** (consumption)                         | Paid tiers               |
| Enterprise portability                | ●●● (cloud-agnostic)      | ●○○ (Azure-locked)                             | ●●○                      |

**Decision ✅ (DL-08): Redpanda OSS on AKS in _all_ environments.**

- **dev / test**: **single broker**, ephemeral disk, spot node — cheap and torn down with the environment.
- **staging / prod**: **3 brokers, RF=3**, on-demand nodes + PodDisruptionBudget — full fidelity, portable to enterprise/on-prem.
- One technology everywhere: services speak the Kafka protocol via `pkg/redpanda` unchanged, and the **Wasm anti-poisoning transforms** run natively in every environment (full trust-validation parity). No managed-service lock-in.

> Rationale: user selected a single event backbone for operational simplicity and complete
> Wasm-transform parity across dev → prod. Dev cost stays low via a single-broker spot deployment.

---

## 3. OLAP Store

| Criterion                      | **ClickHouse OSS on AKS**      | **Azure Data Explorer (ADX)** | ClickHouse Cloud |
| ------------------------------ | ------------------------------ | ----------------------------- | ---------------- |
| Fidelity to current SQL/schema | ●●● (identical)                | ●○○ (KQL rewrite of queries)  | ●●●              |
| Ops burden                     | ●○○ (Altinity operator, disks) | ●●● managed                   | ●●●              |
| Ephemeral                      | ●●○                            | ●●○ (cluster create is slow)  | ●●○              |
| PAYG cost (idle)               | node + disk                    | ●●○ (can stop cluster)        | subscription     |
| Licensing (upfront)            | **None** (Apache-2)            | **None** (consumption)        | Paid             |
| Enterprise portability         | ●●●                            | ●○○ (Azure-locked, KQL)       | ●●○              |

**Recommendation ✅ (DL-09): ClickHouse OSS on AKS** for all environments (single node in
dev, sharded+replicated in staging/prod). It preserves the existing SQL, `svc-query`, and
Redpanda Connect ETL **unchanged**, and is fully portable. ADX would force a KQL rewrite of
`svc-query` and the ETL — avoid unless you specifically want a managed OLAP.

---

## 4. Service Mesh / Resiliency Plane

> This layer delivers **rate limiting, circuit breaking, bulkheads, retries, timeouts, and
> mTLS** — your explicit "resiliency out of code" requirement.

| Criterion                            | **Istio (AKS managed add-on)** ✅ | Linkerd | Cilium Service Mesh   | No mesh (NGINX + app) |
| ------------------------------------ | --------------------------------- | ------- | --------------------- | --------------------- |
| mTLS everywhere (policy match)       | ●●●                               | ●●●     | ●●●                   | ●○○                   |
| Circuit breaking (outlier detection) | ●●●                               | ●●○     | ●●○                   | ●○○                   |
| Bulkheads (connection pools)         | ●●●                               | ●●○     | ●●○                   | ●○○                   |
| Rate limiting (local + global)       | ●●●                               | ●○○     | ●●○                   | ●●○ (NGINX)           |
| Managed by AKS (less ops)            | ●●● (add-on)                      | ●○○     | ●●○ (with Cilium CNI) | ●●●                   |
| Maturity / ecosystem                 | ●●●                               | ●●●     | ●●○                   | ●●●                   |
| Licensing                            | **None** (OSS add-on)             | None    | None                  | None                  |

**Recommendation ✅ (DL-10): AKS managed Istio add-on.** It is the single most direct way to
satisfy the resiliency-out-of-code mandate: `DestinationRule.outlierDetection` (circuit
breaker), `DestinationRule.connectionPool` (bulkhead), `EnvoyFilter`/ratelimit (rate limit),
`VirtualService` retries/timeouts, and STRICT mTLS — all as **configuration**. Managed by
AKS means no control-plane ops. Details in [04 Resiliency](04-resiliency-and-well-architected.md).

---

## 5. Ingress / Edge

| Layer                               | Options                                                                  | Recommendation ✅                                                                                |
| ----------------------------------- | ------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| In-cluster ingress (cold path)      | Istio ingress gateway · App Routing (NGINX) · App Gateway for Containers | **Istio ingress gateway** (DL-11) — one data plane for edge + mesh                               |
| Internet edge (optional, hardening) | Azure **Front Door** (+ WAF) · Application Gateway (+ WAF)               | **Front Door** for global TLS/WAF/caching of cold path + static assets                           |
| Hot path (WebTransport)             | Standard LB UDP/443 (direct)                                             | **Standard LB + dedicated public IP** (DL-12) — Front Door/App Gateway cannot proxy WebTransport |

Rationale: a single Istio gateway keeps CSP/origin management simple and applies mesh
policy at the edge. Front Door is **additive** for WAF and caching in the hardening phase and
only ever fronts the **cold path + static** content, never the QUIC hot path.

---

## 6. Secrets & Identity

| Concern             | Options                                                     | Recommendation ✅ (DL-13)                                |
| ------------------- | ----------------------------------------------------------- | -------------------------------------------------------- |
| Cloud secrets store | **Azure Key Vault** · HashiCorp Vault (self-host)           | **Key Vault** — managed, consumption-priced, no license  |
| Pod → secret access | Secrets Store CSI · env injection                           | **Secrets Store CSI driver** (projects KV secrets/certs) |
| Pod identity        | **Entra Workload Identity** · AAD Pod Identity (deprecated) | **Entra Workload Identity** (federated, no secrets)      |
| CI/CD → Azure       | OIDC federation · service principal secret                  | **OIDC federated credentials** (no stored secrets)       |
| mTLS certs / PKI    | cert-manager + Key Vault issuer · mesh-issued               | **Istio-issued mesh mTLS** + cert-manager for public TLS |

This fully replaces the on-prem `sealed-secrets` approach and satisfies "no secrets in
code" with zero long-lived credentials in GitHub or images.

---

## 7. Observability

| Criterion                      | **Azure Managed Prometheus + Grafana + Container Insights** | Self-host kube-prometheus-stack + Loki + Tempo |
| ------------------------------ | ----------------------------------------------------------- | ---------------------------------------------- |
| Ops burden                     | ●●●                                                         | ●○○                                            |
| Fidelity to current dashboards | ●●○ (re-import)                                             | ●●● (identical)                                |
| PAYG cost (idle)               | ●●○ (ingestion/query priced)                                | ●●● (just node cost; scales to zero)           |
| Ephemeral friendliness         | ●●● (nothing to persist)                                    | ●●○                                            |
| Enterprise portability         | ●●○                                                         | ●●●                                            |
| Licensing                      | **None** (consumption)                                      | **None** (OSS)                                 |

**Recommendation ✅ (DL-14): hybrid.**

- Keep the **OTel Collector** (unchanged) as the in-cluster telemetry funnel.
- **dev/test**: **self-host** the existing Prometheus/Grafana/Loki/Tempo (zero managed cost, dashboards identical, dies with the environment).
- **staging/prod**: **Azure Managed Prometheus + Managed Grafana + Container Insights** for durability and low ops.
- This keeps developer inner-loop free and gives production a managed backend.

---

## 8. Frontend Hosting

| Criterion                                | **Static Web Apps + CDN** | **Nginx Deployment on AKS** ✅ | Azure Blob + Front Door |
| ---------------------------------------- | ------------------------- | ------------------------------ | ----------------------- |
| Cost                                     | ●●● (free/near-free)      | ●●○ (in-cluster)               | ●●●                     |
| Single origin / CSP simplicity with APIs | ●○○ (cross-origin)        | ●●● (same ingress)             | ●○○                     |
| WebTransport/gRPC-Web co-location        | ●○○                       | ●●●                            | ●○○                     |
| Ephemeral parity with backend            | ●●○                       | ●●● (same Helm release)        | ●●○                     |

**Recommendation ✅: Nginx Deployment on AKS behind the Istio ingress** for the baseline —
it keeps the COP, cold-path gateway, and CSP under one origin and one Helm release, which
matters given the strict Content-Security-Policy in the existing `web-cop-gpu` Dockerfile.
Move static assets to **Static Web Apps / Front Door CDN** in the hardening phase if you want
global edge caching.

---

## 9. GitOps / Delivery Engine

| Criterion                   | **Flux** ✅ | ArgoCD              | Pipeline-driven Helm  |
| --------------------------- | ----------- | ------------------- | --------------------- |
| Footprint / cost            | ●●● (light) | ●●○ (UI, more pods) | ●●● (none in-cluster) |
| Pull-based drift correction | ●●●         | ●●●                 | ●○○                   |
| UI / visualization          | ●○○         | ●●●                 | ●○○                   |
| Ephemeral fit               | ●●●         | ●●○                 | ●●●                   |
| Enterprise scale            | ●●●         | ●●●                 | ●●○                   |

**Recommendation ✅ (DL-15): Flux** for GitOps (lightweight, CNCF, great for ephemeral
clusters and multi-tenant enterprise). Choose **ArgoCD** if you want a rich UI for operators.
For the very first walking skeleton, **pipeline-driven `helm upgrade`** from GitHub Actions is
acceptable and is the fastest path; adopt Flux as environments multiply. See
[05 DevOps](05-devops-cicd-and-environments.md#4-gitops--promotion).

---

## 10. Autoscaling

| Layer             | Technology             | Purpose                                                        |
| ----------------- | ---------------------- | -------------------------------------------------------------- |
| Event-driven pods | **KEDA**               | Scale ingestion/processing on Redpanda **lag**, 0→N            |
| CPU/mem pods      | **HPA**                | Classic resource-based scaling for presentation/query          |
| Nodes             | **Cluster Autoscaler** | Grow/shrink spot pools; scale user pools to **zero** when idle |

**Recommendation ✅ (DL-16): KEDA + HPA + Cluster Autoscaler.** KEDA is the linchpin of
event-driven cost efficiency and is a free add-on.

---

## 11. Consolidated Decision Summary

| ID    | Area             | Recommendation                                                                     | Status |
| ----- | ---------------- | ---------------------------------------------------------------------------------- | :----: |
| DL-01 | CI/CD            | GitHub Actions                                                                     |   ✅   |
| DL-02 | IaC              | Terraform                                                                          |   ✅   |
| DL-07 | Orchestrator     | AKS Standard                                                                       |   ✅   |
| DL-08 | Event backbone   | **Redpanda OSS on AKS — all environments** (1-broker dev · 3-broker RF=3 stg/prod) |   ✅   |
| DL-09 | OLAP             | ClickHouse OSS on AKS                                                              |   ✅   |
| DL-10 | Mesh             | AKS managed Istio add-on                                                           |   ✅   |
| DL-11 | Ingress          | Istio gateway (+ Front Door later)                                                 |   ✅   |
| DL-12 | Hot path         | Standard LB UDP/443                                                                |   ✅   |
| DL-13 | Secrets/identity | Key Vault + Workload Identity + CSI                                                |   ✅   |
| DL-14 | Observability    | Hybrid (self-host dev · managed prod)                                              |   ✅   |
| DL-15 | GitOps           | Flux (Helm-from-CI to start)                                                       |   ✅   |
| DL-16 | Autoscaling      | KEDA + HPA + Cluster Autoscaler                                                    |   ✅   |

> **All decisions confirmed ✅.** They drive
> [06 IaC](06-iac-terraform-and-lifecycle.md) and [07 Roadmap](07-implementation-roadmap.md).

> Continue to **[04 — Resiliency & Well-Architected »](04-resiliency-and-well-architected.md)**
