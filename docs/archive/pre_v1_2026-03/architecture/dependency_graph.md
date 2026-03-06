<!-- CLASSIFICATION: UNCLASSIFIED -->
# Dependency Graph

> **Document**: RTSA System Dependency Graph
> **Version**: 2.0
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-02-28

---

## 1. System Dependency Overview

```mermaid
flowchart TD
    subgraph External["External Dependencies"]
        SENSORS[Sensor Systems]
        NATO[NATO Systems]
        PKI[Government PKI]
        NTP[NTP Stratum 1]
        SIEM[Enterprise SIEM]
    end

    subgraph Infrastructure["Infrastructure Layer"]
        K8S[Kubernetes / K3s]
        OS[Linux OS<br/>RHEL / Ubuntu]
        HW[Hardware / VM]
        NET[Network Fabric]
    end

    subgraph Platform["Platform Layer"]
        RP[Redpanda<br/>Event Streaming]
        CH[ClickHouse<br/>OLAP Storage]
        RPC[Redpanda Connect<br/>ETL Pipeline]
        SR[Schema Registry<br/>Protobuf Schemas]
        CM[cert-manager<br/>Certificate Lifecycle]
    end

    subgraph Services["Application Services"]
        ING[Ingestion Services<br/>6 sensor types]
        FUS[Fusion Engine]
        ANO[Anomaly Detection]
        FB[Feedback Service]
        TRAIN[Training Pipeline]
        NATO_SVC[NATO Adapter]
        TRACK[Track Service]
        ALERT[Alert Service]
        QUERY[Query Service]
        AUDIT[Audit Service]
        GW[API Gateway]
        SYNC[Edge Sync Agent]
    end

    subgraph Presentation["Presentation Layer"]
        UI[COP Web Application<br/>React + TypeScript]
    end

    subgraph Observability["Observability Layer"]
        PROM[Prometheus]
        GRAF[Grafana]
        LOKI[Loki]
        TEMPO[Tempo]
        OTEL[OpenTelemetry Collector]
    end

    %% External → Infrastructure
    PKI --> CM
    NTP --> K8S

    %% Infrastructure → Platform
    K8S --> RP
    K8S --> CH
    K8S --> RPC

    %% Platform → Services
    RP --> ING
    RP --> FUS
    RP --> ANO
    RP --> FB
    RP --> TRAIN
    RP --> NATO_SVC
    RP --> TRACK
    RP --> ALERT
    RP --> AUDIT
    RP --> SYNC
    SR --> ING
    SR --> FUS
    CH --> QUERY
    RPC --> CH
    RPC --> RP
    CM --> ING
    CM --> FUS
    CM --> ANO

    %% External → Services
    SENSORS --> ING
    NATO --> NATO_SVC
    NATO_SVC --> NATO
    AUDIT --> SIEM

    %% Services → Services
    ING -->|sensors.*| RP
    FUS -->|tracks.fused.*| RP
    ANO -->|alerts.*| RP
    FB -->|feedback.*| RP

    %% Services → Presentation
    TRACK --> GW
    ALERT --> GW
    QUERY --> GW
    GW --> UI

    %% Observability
    OTEL --> PROM
    OTEL --> LOKI
    OTEL --> TEMPO
    PROM --> GRAF
    LOKI --> GRAF
    TEMPO --> GRAF
```

---

## 2. Service Dependency Matrix

| Service | Depends On | Depended On By |
|---|---|---|
| **Radar Ingestion** | Redpanda, Schema Registry, cert-manager | Fusion Engine (via Redpanda) |
| **EW/SIGINT Ingestion** | Redpanda, Schema Registry, cert-manager | Fusion Engine (via Redpanda) |
| **ELINT/COMINT Ingestion** | Redpanda, Schema Registry, cert-manager | Fusion Engine (via Redpanda) |
| **ISR Ingestion** | Redpanda, Schema Registry, cert-manager | Fusion Engine (via Redpanda) |
| **AIS/BFT Ingestion** | Redpanda, Schema Registry, cert-manager | Fusion Engine (via Redpanda) |
| **Cyber Ingestion** | Redpanda, Schema Registry, cert-manager | Fusion Engine (via Redpanda) |
| **Fusion Engine** | Redpanda, Schema Registry | Anomaly Detection, Track Service (via Redpanda) |
| **Anomaly Detection** | Redpanda, Schema Registry | Alert Service (via Redpanda), Feedback Service |
| **Feedback Service** | Redpanda, Schema Registry | Training Pipeline (via Redpanda) |
| **Training Pipeline** | Redpanda, Model Registry | Anomaly Detection (model updates via Redpanda) |
| **NATO Adapter** | Redpanda, Link 16 Terminal, NFFI Gateway | Fusion Engine (inbound), NATO Systems (outbound) |
| **Track Service** | Redpanda | API Gateway |
| **Alert Service** | Redpanda | API Gateway |
| **Query Service** | ClickHouse | API Gateway |
| **Audit Service** | Redpanda, ClickHouse | SIEM (outbound) |
| **API Gateway** | Track/Alert/Query Services, cert-manager | COP Web Application |
| **Edge Sync Agent** | Redpanda (edge), Redpanda (DC) | None |
| **Redpanda Connect** | Redpanda, ClickHouse | None (ETL pipeline) |

---

## 3. Data Flow Dependency Chain

```mermaid
flowchart LR
    S[Sensor Data] --> ING[Ingestion]
    ING --> RP1[Redpanda<br/>sensors.*]
    RP1 --> FUS[Fusion]
    FUS --> RP2[Redpanda<br/>tracks.fused.*]
    RP2 --> ANO[Anomaly<br/>Detection]
    ANO --> RP3[Redpanda<br/>alerts.*]

    RP2 --> TS[Track<br/>Service]
    RP3 --> AS[Alert<br/>Service]

    TS --> GW[API<br/>Gateway]
    AS --> GW
    GW --> UI[COP UI]

    UI --> FB[Feedback<br/>Service]
    FB --> RP4[Redpanda<br/>feedback.*]
    RP4 --> TRAIN[Training<br/>Pipeline]
    TRAIN --> RP5[Redpanda<br/>models.*]
    RP5 --> ANO

    RP1 --> RPC[Redpanda<br/>Connect]
    RP2 --> RPC
    RP3 --> RPC
    RPC --> CH[(ClickHouse)]
    CH --> QS[Query<br/>Service]
    QS --> GW
```

**Critical Path** (sensor-to-display latency budget):

| Stage | Target Latency | Cumulative |
|---|---|---|
| Sensor → Ingestion | < 10 ms | 10 ms |
| Ingestion → Redpanda | < 5 ms | 15 ms |
| Redpanda → Fusion | < 20 ms | 35 ms |
| Fusion processing | < 50 ms | 85 ms |
| Fusion → Redpanda | < 5 ms | 90 ms |
| Redpanda → Track Service | < 10 ms | 100 ms |
| Track Service → Gateway | < 5 ms | 105 ms |
| Gateway → UI render | < 45 ms | **150 ms** |

---

## 4. Startup Order

```mermaid
flowchart TD
    L0[Level 0: Infrastructure] --> L1[Level 1: Platform]
    L1 --> L2[Level 2: Core Services]
    L2 --> L3[Level 3: Processing Services]
    L3 --> L4[Level 4: Presentation]
    L4 --> L5[Level 5: Integration]

    subgraph L0D["Level 0"]
        K8S[Kubernetes]
        NET[Network]
        CERT[cert-manager]
        NTP_C[NTP sync]
    end

    subgraph L1D["Level 1"]
        RP[Redpanda]
        CH[ClickHouse]
        SR[Schema Registry]
    end

    subgraph L2D["Level 2"]
        AUDIT[Audit Service]
        RPC[Redpanda Connect]
        OTEL[OTel Collector]
    end

    subgraph L3D["Level 3"]
        ING[Ingestion Services]
        FUS[Fusion Engine]
        ANO[Anomaly Detection]
        FB[Feedback Service]
    end

    subgraph L4D["Level 4"]
        TRACK[Track Service]
        ALERT[Alert Service]
        QUERY[Query Service]
        GW[API Gateway]
        UI[COP Web App]
    end

    subgraph L5D["Level 5"]
        NATO[NATO Adapter]
        SYNC[Edge Sync Agent]
        TRAIN[Training Pipeline]
    end
```

| Level | Services | Ready When |
|---|---|---|
| 0 | Kubernetes, Network, cert-manager, NTP | Cluster healthy, certificates issued |
| 1 | Redpanda, ClickHouse, Schema Registry | Brokers accepting connections, topics created |
| 2 | Audit Service, Redpanda Connect, OTel Collector | Consuming audit events, ETL running |
| 3 | Ingestion Services, Fusion, Anomaly, Feedback | gRPC servers ready, consuming/producing |
| 4 | Track, Alert, Query Services, Gateway, UI | Streaming to clients, queries returning |
| 5 | NATO Adapter, Sync Agent, Training Pipeline | NATO link established, sync running |

---

## 5. Technology Dependency Tree

```mermaid
flowchart TD
    subgraph Go["Go Ecosystem"]
        GO[Go 1.22+]
        GRPC_GO[google.golang.org/grpc]
        PROTO_GO[google.golang.org/protobuf]
        KGO[github.com/twmb/franz-go]
        CH_GO[github.com/ClickHouse/clickhouse-go]
        OTEL_GO[go.opentelemetry.io/otel]
        SLOG[log/slog stdlib]
        CRYPTO[crypto/tls stdlib]
    end

    subgraph React["React Ecosystem"]
        REACT[React 18+]
        TS[TypeScript 5+]
        GRPC_WEB[grpc-web]
        ZUSTAND[Zustand]
        MAPLIB[Map Library]
        VITE[Vite]
    end

    subgraph Platform_Tech["Platform Technologies"]
        REDPANDA[Redpanda 24.x]
        CLICKHOUSE[ClickHouse 24.x]
        RP_CONNECT[Redpanda Connect 4.x]
        BUF[buf CLI]
    end

    subgraph Security_Tools["Security Toolchain"]
        COSIGN[cosign]
        TRIVY[trivy]
        SYFT[syft]
        SEMGREP[semgrep]
        GOSEC[gosec]
        GOLINT[golangci-lint]
    end

    subgraph Observability_Stack["Observability"]
        PROMETHEUS[Prometheus]
        GRAFANA[Grafana]
        LOKI[Loki]
        TEMPO[Tempo]
        OTEL_COL[OTel Collector]
    end

    subgraph Infra_Tools["Infrastructure"]
        K8S_T[Kubernetes 1.29+]
        K3S[K3s]
        HELM[Helm 3]
        GH_ACTIONS[GitHub Actions]
    end

    GO --> GRPC_GO
    GO --> PROTO_GO
    GO --> KGO
    GO --> CH_GO
    GO --> OTEL_GO
    PROTO_GO --> BUF

    REACT --> TS
    REACT --> GRPC_WEB
    REACT --> ZUSTAND

    COSIGN --> GH_ACTIONS
    TRIVY --> GH_ACTIONS
    SYFT --> GH_ACTIONS
    SEMGREP --> GH_ACTIONS
    GOSEC --> GH_ACTIONS
    GOLINT --> GH_ACTIONS

    HELM --> K8S_T
    HELM --> K3S
```

---

## 6. Failure Impact Analysis

| Component Failure | Impact | Blast Radius | Mitigation |
|---|---|---|---|
| **Redpanda** (total) | All data flow stops | **Critical** — entire system | 5-node cluster, RF=3, automated recovery |
| **Redpanda** (1 broker) | Temporary rebalance | Low — auto-recovery | Leader election < 2s |
| **ClickHouse** (total) | Historical queries unavailable | Medium — analytics only | 2-replica per shard, separate read path |
| **ClickHouse** (1 replica) | No impact (failover) | None | Automatic replica promotion |
| **Fusion Engine** | No fused tracks produced | High — COP shows stale data | 3 replicas, consumer group rebalance |
| **Anomaly Detection** | No new alerts generated | Medium — tracks still visible | 2 replicas, last-known alerts persist |
| **Single Ingestion Service** | One sensor type not ingested | Low — other sensors continue | 2 replicas per sensor type |
| **API Gateway** | UI cannot connect | High — no operator access | 2 replicas, health-check failover |
| **Track Service** | No real-time track updates | High — stale COP | 2 replicas, cached last-known state |
| **Query Service** | No historical queries | Low — real-time unaffected | 2 replicas |
| **Audit Service** | Audit events queued in Redpanda | Low — no data loss | Events buffered in Redpanda until recovery |
| **NATO Adapter** | No NATO data exchange | Medium — local ops continue | 2 replicas, queued messages in Redpanda |
| **Edge WAN Link** | Edge operates independently | Low (edge) — edge continues | Store-and-forward, automatic resync |
| **cert-manager** | No new certificates | Low — existing certs valid | 90-day cert lifetime provides buffer |

---

## 7. Document Dependency Map

```mermaid
flowchart TD
    subgraph Guidelines["SDLC Guidelines"]
        MP[00_master_policy.md]
        SC[01_security_compliance/]
        REQ[02_requirements/]
        ARCH_G[03_architecture_design/]
        CODE[04_coding_standards/]
        TEST[05_testing/]
        CICD[06_integration_cicd/]
        DEPLOY_G[07_deployment_operations/]
        TECH[08_tech_specific/]
        GOV[09_governance/]
    end

    subgraph Business["Business Documentation"]
        BRD[requirements.md]
        FEAT[feature_list.md]
        UC[usecases/UC001-UC015]
    end

    subgraph Architecture["Architecture Documentation"]
        HLA[high_level_architecture.md]
        CD[component_design.md]
        DA[data_architecture.md]
        SA[security_architecture.md]
        DEPL[deployment_architecture.md]
        IA[integration_architecture.md]
        DG[dependency_graph.md]
    end

    MP --> SC
    MP --> REQ
    MP --> ARCH_G
    MP --> CODE
    MP --> TEST
    MP --> CICD
    MP --> DEPLOY_G
    MP --> TECH
    MP --> GOV

    REQ --> BRD
    BRD --> FEAT
    FEAT --> UC

    ARCH_G --> HLA
    HLA --> CD
    HLA --> DA
    HLA --> SA
    HLA --> DEPL
    HLA --> IA
    CD --> DG
    DA --> DG
    IA --> DG

    SC --> SA
    CODE --> CD
    TECH --> DA
    DEPLOY_G --> DEPL
    UC --> CD
    UC --> IA
```

---

## 8. Requirement → Feature → Component Traceability

| Requirement Group | Features | Components | Use Cases |
|---|---|---|---|
| CR-ING (Ingestion) | FEAT-01..06 | 6 Ingestion Services, Wasm Transforms | UC001..UC007 |
| CR-ING-011..012 *(v2.0)* | FEAT-16, FEAT-19 | Track Service (`StreamSensorObservations`), Ingestion Service (`ListSensorStatuses` + `SensorCoverage`) | UC016, UC017 |
| CR-FUS (Fusion) | FEAT-07 | Fusion Engine | UC008 |
| CR-INF (Inference) | FEAT-08 | Anomaly Detection Service | UC009 |
| CR-FB (Feedback) | FEAT-12, FEAT-13 | Feedback Service, Training Pipeline | UC010, UC011 |
| CR-FB-008 *(v2.0)* | FEAT-18 | Alert Service (`AssignAlert`) | UC010 |
| CR-UI (User Interface) | FEAT-13, FEAT-16..19 | Track/Alert/Query Services, Gateway, COP UI (Two-Level RBAC Shell) | UC012, UC016, UC017 |
| CR-UI-009..020 *(v2.0)* | FEAT-13, FEAT-16..19 | `uiStore`, `RoleSelector`, `DashboardSelector`, `FusionDashboard`, `MultiDomainDashboard`, `OperatorDashboard`, `SensorHealthDashboard` | UC012, UC016, UC017 |
| CR-HIS (Historical) | FEAT-14, FEAT-20 | Query Service, ClickHouse, Redpanda Connect | UC013 |
| CR-HIS-008..009 *(v2.0)* | FEAT-20 | Query Service (`GetEventTimeline`), ClickHouse materialized views | UC013 |
| CR-NATO (Interop) | FEAT-15 | NATO Adapter, Cross-Domain Guard | UC014, UC015 |
| CR-SEC (Security) | All features | All services (cross-cutting) | All use cases |
