<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 01 — Development Infrastructure

> **Module**: 01-dev-infrastructure
> **Phase**: P0 (Foundation)
> **Dependencies**: None
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 3 days

---

## 1. Objective

Create the complete development infrastructure: Docker Compose stacks, project root scaffolding (Makefile, go.work, .env), Dockerfiles, and observability configurations. After this module, `docker compose up` brings up the full platform layer (Redpanda, ClickHouse, Prometheus, Grafana, Loki, Tempo, OTel Collector, Envoy, Redpanda Console).

**Acceptance Criteria**:

- `docker compose -f deploy/docker-compose.yml up -d` starts all 10 infrastructure containers
- All containers reach `healthy` status within 60 seconds
- Redpanda Console accessible at `http://localhost:8080`
- ClickHouse accepts queries at `http://localhost:8123`
- Grafana dashboard at `http://localhost:3000` (admin/admin)
- `make help` displays all available targets
- `go work sync` succeeds for multi-module workspace

---

## 2. Files to Create

### 2.1 `deploy/docker-compose.yml` — Infrastructure Stack

```yaml
# CLASSIFICATION: UNCLASSIFIED
# RTSA Development Infrastructure Stack
# Usage: docker compose -f deploy/docker-compose.yml up -d

name: rtsa-dev

services:
  # ──────────────────────────────────────────────
  # Redpanda — Event Streaming Backbone
  # ──────────────────────────────────────────────
  redpanda:
    image: redpandadata/redpanda:v24.1.1
    container_name: rtsa-redpanda
    command:
      - redpanda
      - start
      - --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092
      - --advertise-kafka-addr internal://redpanda:9092,external://localhost:19092
      - --pandaproxy-addr internal://0.0.0.0:8082,external://0.0.0.0:18082
      - --advertise-pandaproxy-addr internal://redpanda:8082,external://localhost:18082
      - --schema-registry-addr internal://0.0.0.0:8081,external://0.0.0.0:18081
      - --advertise-schema-registry-addr internal://redpanda:8081,external://localhost:18081
      - --rpc-addr redpanda:33145
      - --advertise-rpc-addr redpanda:33145
      - --mode dev-container
      - --smp 1
      - --memory 1G
      - --reserve-memory 0M
      - --default-log-level=warn
    ports:
      - "19092:19092" # Kafka API
      - "18081:18081" # Schema Registry
      - "18082:18082" # HTTP Proxy
      - "19644:9644" # Admin API
    volumes:
      - redpanda-data:/var/lib/redpanda/data
    healthcheck:
      test: ["CMD", "rpk", "cluster", "health", "--api-urls", "localhost:9644"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 15s
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 1536M

  # ──────────────────────────────────────────────
  # Redpanda Console — Topic Management UI
  # ──────────────────────────────────────────────
  redpanda-console:
    image: redpandadata/console:v2.4.0
    container_name: rtsa-console
    environment:
      KAFKA_BROKERS: redpanda:9092
      KAFKA_SCHEMAREGISTRY_ENABLED: "true"
      KAFKA_SCHEMAREGISTRY_URLS: http://redpanda:8081
    ports:
      - "8080:8080"
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 256M

  # ──────────────────────────────────────────────
  # ClickHouse — OLAP Analytics Engine
  # ──────────────────────────────────────────────
  clickhouse:
    image: clickhouse/clickhouse-server:24.1
    container_name: rtsa-clickhouse
    environment:
      CLICKHOUSE_DB: rtsa
      CLICKHOUSE_USER: default
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD:-dev_password_change_me}
      CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT: 1
    ports:
      - "8123:8123" # HTTP interface
      - "9000:9000" # Native TCP
    volumes:
      - clickhouse-data:/var/lib/clickhouse
      - clickhouse-logs:/var/log/clickhouse-server
    healthcheck:
      test: ["CMD", "clickhouse-client", "--query", "SELECT 1"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 2G
    ulimits:
      nofile:
        soft: 262144
        hard: 262144

  # ──────────────────────────────────────────────
  # OpenTelemetry Collector
  # ──────────────────────────────────────────────
  otel-collector:
    image: otel/opentelemetry-collector-contrib:0.96.0
    container_name: rtsa-otel
    command: ["--config", "/etc/otel/config.yaml"]
    ports:
      - "4317:4317" # OTLP gRPC
      - "4318:4318" # OTLP HTTP
      - "8888:8888" # Collector metrics
    volumes:
      - ./otel/config.yaml:/etc/otel/config.yaml:ro
    depends_on:
      - prometheus
      - loki
      - tempo
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 256M

  # ──────────────────────────────────────────────
  # Prometheus — Metrics
  # ──────────────────────────────────────────────
  prometheus:
    image: prom/prometheus:v2.51.0
    container_name: rtsa-prometheus
    command:
      - --config.file=/etc/prometheus/prometheus.yml
      - --storage.tsdb.retention.time=7d
      - --web.enable-lifecycle
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--quiet",
          "--tries=1",
          "--spider",
          "http://localhost:9090/-/healthy",
        ]
      interval: 15s
      timeout: 5s
      retries: 3
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 512M

  # ──────────────────────────────────────────────
  # Grafana — Dashboards
  # ──────────────────────────────────────────────
  grafana:
    image: grafana/grafana:10.4.0
    container_name: rtsa-grafana
    environment:
      GF_SECURITY_ADMIN_USER: admin
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_USERS_ALLOW_SIGN_UP: "false"
    ports:
      - "3000:3000"
    volumes:
      - ./grafana/provisioning:/etc/grafana/provisioning:ro
      - grafana-data:/var/lib/grafana
    depends_on:
      - prometheus
      - loki
      - tempo
    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--quiet",
          "--tries=1",
          "--spider",
          "http://localhost:3000/api/health",
        ]
      interval: 15s
      timeout: 5s
      retries: 3
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 256M

  # ──────────────────────────────────────────────
  # Loki — Log Aggregation
  # ──────────────────────────────────────────────
  loki:
    image: grafana/loki:2.9.0
    container_name: rtsa-loki
    command: -config.file=/etc/loki/loki.yaml
    ports:
      - "3100:3100"
    volumes:
      - ./loki/loki.yaml:/etc/loki/loki.yaml:ro
      - loki-data:/loki
    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--quiet",
          "--tries=1",
          "--spider",
          "http://localhost:3100/ready",
        ]
      interval: 15s
      timeout: 5s
      retries: 3
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 256M

  # ──────────────────────────────────────────────
  # Tempo — Distributed Tracing
  # ──────────────────────────────────────────────
  tempo:
    image: grafana/tempo:2.4.0
    container_name: rtsa-tempo
    command: ["-config.file=/etc/tempo/tempo.yaml"]
    ports:
      - "3200:3200" # Tempo HTTP
    volumes:
      - ./tempo/tempo.yaml:/etc/tempo/tempo.yaml:ro
      - tempo-data:/tmp/tempo
    healthcheck:
      test:
        [
          "CMD",
          "wget",
          "--quiet",
          "--tries=1",
          "--spider",
          "http://localhost:3200/ready",
        ]
      interval: 15s
      timeout: 5s
      retries: 3
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 256M

  # ──────────────────────────────────────────────
  # Envoy — API Gateway / gRPC-Web Proxy
  # ──────────────────────────────────────────────
  envoy:
    image: envoyproxy/envoy:v1.29.0
    container_name: rtsa-envoy
    ports:
      - "8443:8443" # gRPC-Web (TLS)
      - "8001:8001" # Envoy admin
    volumes:
      - ./envoy/envoy.yaml:/etc/envoy/envoy.yaml:ro
      - ../certs/dev:/etc/envoy/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net
    deploy:
      resources:
        limits:
          memory: 128M

volumes:
  redpanda-data:
  clickhouse-data:
  clickhouse-logs:
  prometheus-data:
  grafana-data:
  loki-data:
  tempo-data:

networks:
  rtsa-net:
    driver: bridge
    name: rtsa-net
```

### 2.2 `deploy/docker-compose.services.yml` — Application Services Overlay

```yaml
# CLASSIFICATION: UNCLASSIFIED
# RTSA Application Services — overlay for docker-compose.yml
# Usage: docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d

services:
  svc-radar-ingestion:
    build:
      context: ../svc-radar-ingestion
      dockerfile: Dockerfile
    container_name: rtsa-radar-ingestion
    environment:
      RTSA_SERVICE_NAME: svc-radar-ingestion
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
      RTSA_TLS_CA_CERT: /certs/ca.crt
      RTSA_TLS_SERVER_CERT: /certs/server.crt
      RTSA_TLS_SERVER_KEY: /certs/server.key
    ports:
      - "50051:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "/bin/grpc_health_probe", "-addr=:50051"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - rtsa-net

  svc-ew-ingestion:
    build:
      context: ../svc-ew-ingestion
      dockerfile: Dockerfile
    container_name: rtsa-ew-ingestion
    environment:
      RTSA_SERVICE_NAME: svc-ew-ingestion
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50052:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-elint-ingestion:
    build:
      context: ../svc-elint-ingestion
      dockerfile: Dockerfile
    container_name: rtsa-elint-ingestion
    environment:
      RTSA_SERVICE_NAME: svc-elint-ingestion
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50053:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-isr-ingestion:
    build:
      context: ../svc-isr-ingestion
      dockerfile: Dockerfile
    container_name: rtsa-isr-ingestion
    environment:
      RTSA_SERVICE_NAME: svc-isr-ingestion
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50054:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-ais-ingestion:
    build:
      context: ../svc-ais-ingestion
      dockerfile: Dockerfile
    container_name: rtsa-ais-ingestion
    environment:
      RTSA_SERVICE_NAME: svc-ais-ingestion
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50055:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-cyber-ingestion:
    build:
      context: ../svc-cyber-ingestion
      dockerfile: Dockerfile
    container_name: rtsa-cyber-ingestion
    environment:
      RTSA_SERVICE_NAME: svc-cyber-ingestion
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50056:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-fusion-engine:
    build:
      context: ../svc-fusion-engine
      dockerfile: Dockerfile
    container_name: rtsa-fusion-engine
    environment:
      RTSA_SERVICE_NAME: svc-fusion-engine
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50060:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-anomaly-detection:
    build:
      context: ../svc-anomaly-detection
      dockerfile: Dockerfile
    container_name: rtsa-anomaly-detection
    environment:
      RTSA_SERVICE_NAME: svc-anomaly-detection
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50061:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-feedback:
    build:
      context: ../svc-feedback
      dockerfile: Dockerfile
    container_name: rtsa-feedback
    environment:
      RTSA_SERVICE_NAME: svc-feedback
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_CLICKHOUSE_DSN: "clickhouse://default:${CLICKHOUSE_PASSWORD:-dev_password_change_me}@clickhouse:9000/rtsa"
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50062:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-track:
    build:
      context: ../svc-track
      dockerfile: Dockerfile
    container_name: rtsa-track
    environment:
      RTSA_SERVICE_NAME: svc-track
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50070:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-alert:
    build:
      context: ../svc-alert
      dockerfile: Dockerfile
    container_name: rtsa-alert
    environment:
      RTSA_SERVICE_NAME: svc-alert
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50071:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-query:
    build:
      context: ../svc-query
      dockerfile: Dockerfile
    container_name: rtsa-query
    environment:
      RTSA_SERVICE_NAME: svc-query
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_CLICKHOUSE_DSN: "clickhouse://default:${CLICKHOUSE_PASSWORD:-dev_password_change_me}@clickhouse:9000/rtsa"
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50072:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net

  svc-audit:
    build:
      context: ../svc-audit
      dockerfile: Dockerfile
    container_name: rtsa-audit
    environment:
      RTSA_SERVICE_NAME: svc-audit
      RTSA_GRPC_PORT: "50051"
      RTSA_HEALTH_PORT: "8081"
      RTSA_REDPANDA_BROKERS: redpanda:9092
      RTSA_CLICKHOUSE_DSN: "clickhouse://default:${CLICKHOUSE_PASSWORD:-dev_password_change_me}@clickhouse:9000/rtsa"
      RTSA_OTEL_ENDPOINT: otel-collector:4317
      RTSA_LOG_LEVEL: debug
      RTSA_LOG_FORMAT: json
    ports:
      - "50073:50051"
    volumes:
      - ../certs/dev:/certs:ro
    depends_on:
      redpanda:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net

  # ──────────────────────────────────────────
  # Redpanda Connect — ETL Pipelines
  # ──────────────────────────────────────────
  rpconnect-tracks:
    image: redpandadata/connect:4.27.0
    container_name: rtsa-rpconnect-tracks
    volumes:
      - ./redpanda-connect/tracks-to-clickhouse.yaml:/connect.yaml:ro
    environment:
      REDPANDA_BROKERS: redpanda:9092
      CLICKHOUSE_HOST: clickhouse
      CLICKHOUSE_USER: default
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD:-dev_password_change_me}
    depends_on:
      redpanda:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net

  rpconnect-sensors:
    image: redpandadata/connect:4.27.0
    container_name: rtsa-rpconnect-sensors
    volumes:
      - ./redpanda-connect/sensors-to-clickhouse.yaml:/connect.yaml:ro
    environment:
      REDPANDA_BROKERS: redpanda:9092
      CLICKHOUSE_HOST: clickhouse
      CLICKHOUSE_USER: default
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD:-dev_password_change_me}
    depends_on:
      redpanda:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net

  rpconnect-alerts:
    image: redpandadata/connect:4.27.0
    container_name: rtsa-rpconnect-alerts
    volumes:
      - ./redpanda-connect/alerts-to-clickhouse.yaml:/connect.yaml:ro
    environment:
      REDPANDA_BROKERS: redpanda:9092
      CLICKHOUSE_HOST: clickhouse
      CLICKHOUSE_USER: default
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD:-dev_password_change_me}
    depends_on:
      redpanda:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net

  rpconnect-feedback:
    image: redpandadata/connect:4.27.0
    container_name: rtsa-rpconnect-feedback
    volumes:
      - ./redpanda-connect/feedback-to-clickhouse.yaml:/connect.yaml:ro
    environment:
      REDPANDA_BROKERS: redpanda:9092
      CLICKHOUSE_HOST: clickhouse
      CLICKHOUSE_USER: default
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD:-dev_password_change_me}
    depends_on:
      redpanda:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net

  rpconnect-audit:
    image: redpandadata/connect:4.27.0
    container_name: rtsa-rpconnect-audit
    volumes:
      - ./redpanda-connect/audit-to-clickhouse.yaml:/connect.yaml:ro
    environment:
      REDPANDA_BROKERS: redpanda:9092
      CLICKHOUSE_HOST: clickhouse
      CLICKHOUSE_USER: default
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD:-dev_password_change_me}
    depends_on:
      redpanda:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
    networks:
      - rtsa-net
```

### 2.3 Observability Configuration Files

#### `deploy/otel/config.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    send_batch_size: 1024
    timeout: 5s
  memory_limiter:
    check_interval: 1s
    limit_mib: 200
    spike_limit_mib: 50

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
    namespace: rtsa
  loki:
    endpoint: http://loki:3100/loki/api/v1/push
  otlp/tempo:
    endpoint: tempo:4317
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [prometheus]
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [loki]
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [otlp/tempo]
```

#### `deploy/prometheus/prometheus.yml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "otel-collector"
    static_configs:
      - targets: ["otel-collector:8889"]
        labels:
          service: "otel-collector"

  - job_name: "rtsa-services"
    static_configs:
      - targets:
          - "svc-radar-ingestion:9090"
          - "svc-ew-ingestion:9090"
          - "svc-elint-ingestion:9090"
          - "svc-isr-ingestion:9090"
          - "svc-ais-ingestion:9090"
          - "svc-cyber-ingestion:9090"
          - "svc-fusion-engine:9090"
          - "svc-anomaly-detection:9090"
          - "svc-feedback:9090"
          - "svc-track:9090"
          - "svc-alert:9090"
          - "svc-query:9090"
          - "svc-audit:9090"

  - job_name: "redpanda"
    static_configs:
      - targets: ["redpanda:9644"]
```

#### `deploy/grafana/provisioning/datasources/datasources.yml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: false

  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    editable: false
    jsonData:
      tracesToLogs:
        datasourceUid: loki
        filterByTraceID: true
      serviceMap:
        datasourceUid: prometheus
```

#### `deploy/loki/loki.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
auth_enabled: false

server:
  http_listen_port: 3100

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2024-01-01
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

limits_config:
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  ingestion_rate_mb: 16
  ingestion_burst_size_mb: 32
```

#### `deploy/tempo/tempo.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317

storage:
  trace:
    backend: local
    local:
      path: /tmp/tempo/blocks
    wal:
      path: /tmp/tempo/wal
    pool:
      max_workers: 50
      queue_depth: 2000

metrics_generator:
  storage:
    path: /tmp/tempo/generator/wal
```

### 2.4 Envoy Configuration

#### `deploy/envoy/envoy.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 8001

static_resources:
  listeners:
    - name: grpc_web_listener
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 8443
      filter_chains:
        - filters:
            - name: envoy.filters.network.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                codec_type: AUTO
                stat_prefix: ingress_http
                access_log:
                  - name: envoy.access_loggers.stdout
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
                http_filters:
                  - name: envoy.filters.http.grpc_web
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.grpc_web.v3.GrpcWeb
                  - name: envoy.filters.http.cors
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors
                  - name: envoy.filters.http.router
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                route_config:
                  name: local_route
                  virtual_hosts:
                    - name: rtsa_services
                      domains: ["*"]
                      cors:
                        allow_origin_string_match:
                          - prefix: "http://localhost"
                        allow_methods: GET, PUT, DELETE, POST, OPTIONS
                        allow_headers: keep-alive,user-agent,cache-control,content-type,content-transfer-encoding,x-accept-content-transfer-encoding,x-accept-response-streaming,x-user-agent,x-grpc-web,grpc-timeout,authorization
                        expose_headers: grpc-status,grpc-message
                        max_age: "1728000"
                      routes:
                        - match:
                            prefix: /rtsa.entity.v1.TrackService
                          route:
                            cluster: svc-track
                            timeout: 0s
                            max_stream_duration:
                              grpc_timeout_header_max: 0s
                        - match:
                            prefix: /rtsa.inference.v1.AlertService
                          route:
                            cluster: svc-alert
                            timeout: 0s
                            max_stream_duration:
                              grpc_timeout_header_max: 0s
                        - match:
                            prefix: /rtsa.feedback.v1.FeedbackService
                          route:
                            cluster: svc-feedback
                            timeout: 30s
                        - match:
                            prefix: /rtsa.query.v1.QueryService
                          route:
                            cluster: svc-query
                            timeout: 30s
                        - match:
                            prefix: /rtsa.common.v1.HealthService
                          route:
                            cluster: svc-track
                            timeout: 5s
                        - match:
                            prefix: /healthz
                          direct_response:
                            status: 200
                            body:
                              inline_string: "OK"

  clusters:
    - name: svc-track
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: svc-track
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-track
                      port_value: 50051

    - name: svc-alert
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: svc-alert
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-alert
                      port_value: 50051

    - name: svc-feedback
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: svc-feedback
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-feedback
                      port_value: 50051

    - name: svc-query
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: svc-query
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-query
                      port_value: 50051
```

### 2.5 Project Root Files

#### `Makefile`

```makefile
# CLASSIFICATION: UNCLASSIFIED
# RTSA Project Root Makefile

.PHONY: help proto-gen build test lint docker-up docker-down docker-up-all \
        docker-down-all integration-test clean health-check init-topics init-clickhouse

# ──────────────────────────────────────────
# Help
# ──────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

# ──────────────────────────────────────────
# Protobuf
# ──────────────────────────────────────────
proto-gen: ## Generate Go and TypeScript code from .proto files
	buf generate

proto-lint: ## Lint protobuf files
	buf lint

proto-breaking: ## Check for breaking changes in protobuf
	buf breaking --against '.git#branch=main'

# ──────────────────────────────────────────
# Build
# ──────────────────────────────────────────
SERVICES := svc-radar-ingestion svc-ew-ingestion svc-elint-ingestion \
            svc-isr-ingestion svc-ais-ingestion svc-cyber-ingestion \
            svc-fusion-engine svc-anomaly-detection svc-feedback \
            svc-track svc-alert svc-query svc-audit

build: ## Build all services
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd $$svc && go build ./... && cd ..; \
	done

# ──────────────────────────────────────────
# Test
# ──────────────────────────────────────────
test: ## Run all unit tests with race detector
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		cd $$svc && go test -race -count=1 -coverprofile=coverage.out ./... && cd ..; \
	done
	@cd pkg && go test -race -count=1 -coverprofile=coverage.out ./...

test-coverage: ## Run tests and show coverage summary
	@for svc in $(SERVICES); do \
		echo "=== $$svc ==="; \
		cd $$svc && go test -race -count=1 -coverprofile=coverage.out ./... && \
		go tool cover -func=coverage.out | tail -1 && cd ..; \
	done

integration-test: ## Run integration tests (requires docker stack running)
	go test -race -count=1 -tags=integration ./tests/integration/...

# ──────────────────────────────────────────
# Lint
# ──────────────────────────────────────────
lint: ## Run golangci-lint on all services
	@for svc in $(SERVICES); do \
		echo "Linting $$svc..."; \
		cd $$svc && golangci-lint run ./... && cd ..; \
	done
	@cd pkg && golangci-lint run ./...

# ──────────────────────────────────────────
# Docker — Infrastructure Only
# ──────────────────────────────────────────
docker-up: ## Start infrastructure stack (Redpanda, ClickHouse, observability)
	docker compose -f deploy/docker-compose.yml up -d

docker-down: ## Stop infrastructure stack
	docker compose -f deploy/docker-compose.yml down

docker-logs: ## Follow infrastructure logs
	docker compose -f deploy/docker-compose.yml logs -f

# ──────────────────────────────────────────
# Docker — Full Stack (Infrastructure + Services)
# ──────────────────────────────────────────
docker-up-all: ## Start full stack (infrastructure + RTSA services)
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --build

docker-down-all: ## Stop full stack
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down

docker-logs-all: ## Follow full stack logs
	docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml logs -f

# ──────────────────────────────────────────
# Initialization
# ──────────────────────────────────────────
init-topics: ## Create Redpanda topics
	./scripts/dev/init-topics.sh

init-clickhouse: ## Initialize ClickHouse schema
	./scripts/dev/init-clickhouse.sh

health-check: ## Run health check
	./scripts/dev/health-check.sh

# ──────────────────────────────────────────
# Clean
# ──────────────────────────────────────────
clean: ## Remove build artifacts and coverage files
	@for svc in $(SERVICES); do \
		rm -f $$svc/coverage.out; \
	done
	rm -f pkg/coverage.out
	rm -rf gen/
```

#### `go.work`

```go
// CLASSIFICATION: UNCLASSIFIED
go 1.22.0

use (
	./pkg
	./svc-radar-ingestion
	./svc-ew-ingestion
	./svc-elint-ingestion
	./svc-isr-ingestion
	./svc-ais-ingestion
	./svc-cyber-ingestion
	./svc-fusion-engine
	./svc-anomaly-detection
	./svc-feedback
	./svc-track
	./svc-alert
	./svc-query
	./svc-audit
	./tools/simulator
	./tests/integration
)
```

### 2.6 Dockerfile Template

#### `deploy/Dockerfile.service` — Template for All Go Services

```dockerfile
# CLASSIFICATION: UNCLASSIFIED
# Multi-stage build for RTSA Go services
# Usage: Copy to svc-<name>/Dockerfile and adjust binary name

# ── Build Stage ────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Download dependencies first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app ./cmd/*/

# Install grpc_health_probe for health checks
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \
    wget -qO/bin/grpc_health_probe \
    "https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64" && \
    chmod +x /bin/grpc_health_probe

# ── Runtime Stage ──────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="RTSA Service"
LABEL org.opencontainers.image.classification="UNCLASSIFIED"

COPY --from=builder /app /app
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nonroot:nonroot

EXPOSE 50051 8081 9090

ENTRYPOINT ["/app"]
```

### 2.7 Redpanda Connect ETL Pipeline Configs

#### `deploy/redpanda-connect/tracks-to-clickhouse.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - tracks.fused.surface
      - tracks.fused.air
      - tracks.fused.subsurface
      - tracks.fused.land
      - tracks.fused.cyber
    consumer_group: rpconnect-clickhouse-tracks

pipeline:
  processors:
    - mapping: |
        root.track_id = this.track_id
        root.entity_type = this.entity_type
        root.hostile_classification = this.hostile_class
        root.latitude = this.estimated_position.latitude
        root.longitude = this.estimated_position.longitude
        root.altitude_meters = this.estimated_position.altitude_meters
        root.speed_knots = this.estimated_position.speed_knots
        root.heading_degrees = this.estimated_position.heading_degrees
        root.confidence_score = this.confidence_score
        root.source_count = this.source_count
        root.source_sensors = this.sources.map_each(s -> s.sensor_id)
        root.classification_level = this.classification
        root.track_status = this.status
        root.event_time = this.updated_at

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9000/rtsa"
    table: tracks_fused
    columns:
      - track_id
      - entity_type
      - hostile_classification
      - latitude
      - longitude
      - altitude_meters
      - speed_knots
      - heading_degrees
      - confidence_score
      - source_count
      - source_sensors
      - classification_level
      - track_status
      - event_time
    batching:
      count: 1000
      period: 5s
```

#### `deploy/redpanda-connect/sensors-to-clickhouse.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - sensors.radar.tracks
      - sensors.ew.intercepts
      - sensors.elint.detections
      - sensors.isr.observations
      - sensors.ais.positions
      - sensors.cyber.iocs
    consumer_group: rpconnect-clickhouse-sensors

pipeline:
  processors:
    - mapping: |
        root.observation_id = this.observation_id
        root.sensor_id = this.sensor_id
        root.sensor_type = this.sensor_type
        root.latitude = this.position.latitude
        root.longitude = this.position.longitude
        root.altitude_meters = this.position.altitude_meters
        root.speed_knots = this.position.speed_knots
        root.heading_degrees = this.position.heading_degrees
        root.classification_level = this.classification
        root.metadata_json = this.metadata.string()
        root.event_time = this.observation_time

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9000/rtsa"
    table: sensor_observations
    columns:
      - observation_id
      - sensor_id
      - sensor_type
      - latitude
      - longitude
      - altitude_meters
      - speed_knots
      - heading_degrees
      - classification_level
      - metadata_json
      - event_time
    batching:
      count: 1000
      period: 5s
```

#### `deploy/redpanda-connect/alerts-to-clickhouse.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - alerts.anomaly.critical
      - alerts.anomaly.elevated
      - alerts.anomaly.watch
    consumer_group: rpconnect-clickhouse-alerts

pipeline:
  processors:
    - mapping: |
        root.alert_id = this.alert_id
        root.track_id = this.track_id
        root.anomaly_type = this.anomaly_type
        root.severity = this.severity
        root.confidence_score = this.confidence_score
        root.explanation = this.explanation
        root.model_version = this.model_version
        root.classification_level = this.classification
        root.event_time = this.detected_at

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9000/rtsa"
    table: anomaly_detections
    columns:
      - alert_id
      - track_id
      - anomaly_type
      - severity
      - confidence_score
      - explanation
      - model_version
      - classification_level
      - event_time
    batching:
      count: 500
      period: 5s
```

#### `deploy/redpanda-connect/feedback-to-clickhouse.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - feedback.operator.submissions
      - feedback.operator.validated
    consumer_group: rpconnect-clickhouse-feedback

pipeline:
  processors:
    - mapping: |
        root.feedback_id = this.feedback_id
        root.track_id = this.track_id
        root.operator_id = this.operator_id
        root.feedback_type = this.feedback_type
        root.justification = this.justification
        root.trust_score = this.trust_score
        root.clearance_score = this.trust_breakdown.clearance_score
        root.accuracy_score = this.trust_breakdown.accuracy_score
        root.temporal_score = this.trust_breakdown.temporal_score
        root.deviation_score = this.trust_breakdown.deviation_score
        root.validated = if this.trust_score >= 0.5 { 1 } else { 0 }
        root.classification_level = this.classification
        root.event_time = this.submitted_at

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9000/rtsa"
    table: operator_feedback
    columns:
      - feedback_id
      - track_id
      - operator_id
      - feedback_type
      - justification
      - trust_score
      - clearance_score
      - accuracy_score
      - temporal_score
      - deviation_score
      - validated
      - classification_level
      - event_time
    batching:
      count: 100
      period: 5s
```

#### `deploy/redpanda-connect/audit-to-clickhouse.yaml`

```yaml
# CLASSIFICATION: UNCLASSIFIED
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - audit.events
    consumer_group: rpconnect-clickhouse-audit

pipeline:
  processors:
    - mapping: |
        root.audit_id = this.audit_id
        root.service_id = this.service_id
        root.event_type = this.event_type
        root.actor_id = this.actor_id
        root.actor_type = this.actor_type
        root.resource_type = this.resource_type
        root.resource_id = this.resource_id
        root.action = this.action
        root.detail_json = this.detail_json
        root.classification_level = this.classification_level
        root.event_time = this.event_time

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9000/rtsa"
    table: audit_log
    columns:
      - audit_id
      - service_id
      - event_type
      - actor_id
      - actor_type
      - resource_type
      - resource_id
      - action
      - detail_json
      - classification_level
      - event_time
    batching:
      count: 500
      period: 5s
```

---

## 3. Test Scenarios

### 3.1 Infrastructure Smoke Tests

| #   | Scenario                                              | Expected                               |
| --- | ----------------------------------------------------- | -------------------------------------- |
| T01 | `docker compose -f deploy/docker-compose.yml up -d`   | All 10 containers start                |
| T02 | Wait 60s, then `docker compose ps`                    | All show `healthy` or `running`        |
| T03 | `curl http://localhost:8080`                          | Redpanda Console UI loads              |
| T04 | `echo 'SELECT 1' \| curl http://localhost:8123 -d @-` | Returns `1`                            |
| T05 | `curl http://localhost:9090/-/healthy`                | Returns `Prometheus Server is Healthy` |
| T06 | `curl http://localhost:3000/api/health`               | Returns `{"database":"ok"}`            |
| T07 | `curl http://localhost:3100/ready`                    | Returns `ready`                        |
| T08 | `curl http://localhost:3200/ready`                    | Returns OK                             |
| T09 | `rpk cluster health --brokers localhost:19092`        | Shows healthy cluster                  |
| T10 | `curl http://localhost:8443/healthz`                  | Returns `OK` (Envoy)                   |

### 3.2 Makefile Tests

| #   | Command                | Expected                            |
| --- | ---------------------- | ----------------------------------- |
| T11 | `make help`            | Shows all targets with descriptions |
| T12 | `make proto-lint`      | Runs buf lint successfully          |
| T13 | `make docker-up`       | Starts infrastructure               |
| T14 | `make docker-down`     | Stops infrastructure                |
| T15 | `make init-topics`     | Creates all Redpanda topics         |
| T16 | `make init-clickhouse` | Creates ClickHouse tables           |

---

## 4. Agent Invocation

```
@greatest-ever-developer Implement Module 01 from docs/implementation/01-dev-infrastructure.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- This module creates the development infrastructure (Docker Compose, Makefile, go.work)
- No application code — only infrastructure configuration files
- All containers must have health checks
- Use exact image versions specified in docs/implementation/00-implementation-overview.md

Deliverables:
1. deploy/docker-compose.yml (infrastructure stack)
2. deploy/docker-compose.services.yml (application services overlay)
3. deploy/otel/config.yaml
4. deploy/prometheus/prometheus.yml
5. deploy/grafana/provisioning/datasources/datasources.yml
6. deploy/loki/loki.yaml
7. deploy/tempo/tempo.yaml
8. deploy/envoy/envoy.yaml
9. deploy/redpanda-connect/*.yaml (5 ETL pipeline configs)
10. deploy/Dockerfile.service (Go service template)
11. Makefile (project root)
12. go.work (Go workspace)
13. All infrastructure containers healthy after `docker compose up`
```
