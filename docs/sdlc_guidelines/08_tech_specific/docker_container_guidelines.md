# Docker & Container Guidelines

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Technology-Specific Standard
> **Parent**: `00_master_policy.md`
> **Last Updated**: 2026-03-02

---

## 1. Purpose

This document defines best practices for building, securing, and operating Docker containers in production and pre-production environments. It covers container image construction, security hardening, orchestration readiness, resource management, and operational patterns applicable to containerized microservice architectures.

## 2. Container Image Construction

### 2.1 Multi-Stage Build Pattern

Every production Dockerfile must use multi-stage builds. The final stage should contain only the compiled binary and its minimal runtime dependencies.

```dockerfile
# CLASSIFICATION: UNCLASSIFIED

# --- Stage 1: Build ---
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /bin/service ./cmd/server/

# --- Stage 2: Runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/service /service
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER nonroot:nonroot
EXPOSE 50051 9090
ENTRYPOINT ["/service"]
```

### 2.2 Base Image Selection

| Use Case | Recommended Base | Image Size | Shell Available |
|---|---|---|---|
| Go services (production) | `gcr.io/distroless/static` | ~2 MB | No |
| Go services (with net/cgo) | `gcr.io/distroless/base` | ~20 MB | No |
| Go services (debug) | `gcr.io/distroless/static:debug` | ~5 MB | Busybox |
| Node.js services | `gcr.io/distroless/nodejs` | ~60 MB | No |
| Alpine-based (fallback) | `alpine:3.x` | ~7 MB | Yes |

**Rules:**
- **Never** use full OS base images (Ubuntu, Debian, Fedora) for production containers
- **Never** use `latest` tags — always pin to a specific version or digest
- Periodically update base images to incorporate security patches

### 2.3 Layer Optimization

- Order `COPY` and `RUN` instructions from least to most frequently changed
- Combine related `RUN` commands to minimize layers
- Remove build artifacts and package manager caches in the same `RUN` layer
- Use `.dockerignore` to exclude `.git`, documentation, test fixtures, IDE configs

```dockerfile
# GOOD — combined, cleanup in same layer
RUN apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

# BAD — separate layers, cache not cleaned
RUN apk add ca-certificates
RUN apk add tzdata
```

### 2.4 Reproducible Builds

- Pin all base image versions to specific tags or SHA digests
- Pin all tool versions in the build stage
- Use `go mod download` or `npm ci` (not `npm install`) for deterministic dependency resolution
- Set `CGO_ENABLED=0` for statically-linked Go binaries
- Use `-trimpath` to remove local filesystem paths from the binary

## 3. Container Security

### 3.1 Non-Root Execution

Every container must run as a non-root user. Never run processes as UID 0.

```dockerfile
# Create and use non-root user
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup
USER appuser:appgroup

# Or use distroless 'nonroot' variant
FROM gcr.io/distroless/static:nonroot
USER nonroot:nonroot
```

### 3.2 Read-Only Root Filesystem

Production containers should use a read-only root filesystem. Write to ephemeral tmpfs mounts if temporary storage is needed:

```yaml
# Kubernetes SecurityContext
securityContext:
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  allowPrivilegeEscalation: false
  capabilities:
    drop: ["ALL"]
  seccompProfile:
    type: RuntimeDefault
```

```yaml
# Docker Compose
services:
  my-service:
    read_only: true
    tmpfs:
      - /tmp
```

### 3.3 Minimal Capabilities

Drop all Linux capabilities and add back only what is strictly required:

```yaml
securityContext:
  capabilities:
    drop: ["ALL"]
    add: []  # Add specific capabilities only if absolutely necessary
```

**Common capabilities to avoid granting:**
- `NET_RAW` — unless the container needs raw socket access
- `SYS_ADMIN` — grants near-root privileges; never use
- `SYS_PTRACE` — only for debugging containers

### 3.4 Secrets Management in Containers

- **Never** embed secrets in Docker images — no `ENV SECRET=...` in Dockerfiles
- **Never** pass secrets via `docker build --build-arg` (they appear in image history)
- Mount secrets at runtime via Kubernetes Secrets, Docker Secrets, or environment variables
- For build-time secrets (e.g., private registry access), use BuildKit secret mounts:

```dockerfile
RUN --mount=type=secret,id=registry_token \
    cat /run/secrets/registry_token | docker login ...
```

### 3.5 Vulnerability Scanning

Scan container images at every stage of the lifecycle:

| Stage | Tool | Action on Finding |
|---|---|---|
| Local development | `trivy image <image>` | Fix before commit |
| CI pipeline | `trivy`, `grype`, `snyk` | Block on Critical/High |
| Container registry | Registry-integrated scanning | Alert on new CVEs |
| Runtime | `falco`, `sysdig` | Detect anomalous behavior |

## 4. Health Check Patterns

### 4.1 Health Check Types

Every container must implement health checks for orchestrator integration:

| Check | Purpose | Example |
|---|---|---|
| **Liveness** | Is the process alive and not deadlocked? | HTTP `GET /healthz` returns 200 |
| **Readiness** | Is the service ready to accept traffic? | HTTP `GET /readyz` returns 200 (checks DB, broker connectivity) |
| **Startup** | Has initial startup completed? | HTTP `GET /startupz` returns 200 (model loading, cache warming) |

### 4.2 Health Check Implementation

```dockerfile
HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=3 \
  CMD ["/service", "--health-check"]
```

**Best practices:**
- Health check endpoints must be lightweight — no heavy computation or database queries
- Liveness checks should verify the process is responsive, not verify external dependencies
- Readiness checks should verify connectivity to critical dependencies (database, message broker)
- Use a separate HTTP port for health/metrics (e.g., `:8080`) distinct from the application port

### 4.3 Graceful Shutdown

Containers must handle `SIGTERM` gracefully:

1. Stop accepting new connections / requests
2. Drain in-flight requests (with a timeout)
3. Close database and broker connections cleanly
4. Flush telemetry and metrics
5. Exit with code 0

Set `terminationGracePeriodSeconds` (Kubernetes) or `stop_grace_period` (Compose) appropriately:

```yaml
# Kubernetes
terminationGracePeriodSeconds: 30

# Docker Compose
services:
  my-service:
    stop_grace_period: 30s
```

## 5. Resource Management

### 5.1 CPU and Memory Limits

Always set CPU and memory limits to prevent a single container from consuming all host resources:

```yaml
# Kubernetes
resources:
  requests:
    cpu: "250m"
    memory: "128Mi"
  limits:
    cpu: "1000m"
    memory: "512Mi"
```

**Guidelines:**
- Set requests based on normal operating usage (profiling data)
- Set limits based on peak usage + safety margin
- Memory limits should match the application's memory budget (e.g., `GOMEMLIMIT` in Go)
- CPU limits prevent noisy neighbor issues but may cause throttling — set judiciously

### 5.2 OOM Prevention

- Set memory limits slightly above the application's configured memory budget
- For Go: set `GOMEMLIMIT` to ~85% of the container memory limit (allows headroom for GC)
- For Node.js: set `--max-old-space-size` appropriately
- Monitor container OOM kill events and adjust limits based on data

## 6. Container Logging

### 6.1 Logging Best Practices

- Log to `stdout` (informational) and `stderr` (errors) — container orchestrators capture these streams
- Use structured JSON logging — do not use unstructured text output
- Include container/service identity in log entries (service name, instance ID)
- Never log sensitive data (credentials, PII, classified information)
- Avoid excessive debug logging in production — use log level configuration

### 6.2 Log Rotation

In Docker environments, configure log driver options to prevent disk exhaustion:

```yaml
logging:
  driver: json-file
  options:
    max-size: "10m"
    max-file: "5"
    tag: "{{.Name}}"
```

## 7. Container Networking

### 7.1 Network Best Practices

- Each service should listen on `0.0.0.0` (all interfaces) inside the container, not `127.0.0.1`
- Use DNS-based service discovery (container names or Kubernetes DNS) — never hardcode IP addresses
- Minimize exposed ports — only expose ports that need to be accessed externally
- Use network policies (Kubernetes) to restrict pod-to-pod communication

### 7.2 Port Conventions

| Port Use | Convention |
|---|---|
| gRPC application port | 50051 (or per service configuration) |
| HTTP health/readiness | 8080 |
| Prometheus metrics | 9090 |
| Debug/pprof (dev only) | 6060 |

## 8. Image Tagging and Versioning

### 8.1 Tagging Strategy

| Tag Format | When Used | Mutable? |
|---|---|---|
| `<git-sha-short>` (e.g., `a1b2c3d`) | Every CI build | No — immutable |
| `<semver>` (e.g., `1.2.3`) | Release tags | No — immutable |
| `<branch>-<sha>` (e.g., `main-a1b2c3d`) | Branch builds | No — immutable |
| `latest` | Local development only | Yes — mutable |

**Rules:**
- Production deployments must reference immutable tags (SHA or semver) — **never** `latest`
- Tag images at build time with Git SHA, version, and build timestamp
- Implement image retention policies in the registry (e.g., keep last 50 tags per repository)

### 8.2 OCI Labels

Apply standard OCI labels to all images for metadata tracking:

```dockerfile
LABEL org.opencontainers.image.title="service-name" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.source="https://github.com/org/repo" \
      org.opencontainers.image.vendor="Organization Name"
```

## 9. Container Registry Management

### 9.1 Registry Best Practices

- Use a private container registry — never push project images to public registries
- Enable image vulnerability scanning on the registry
- Enable image signing and verification (cosign, Docker Content Trust)
- Attach SBOM (CycloneDX or SPDX) to every image
- Configure garbage collection to reclaim storage from untagged/expired images

### 9.2 Air-Gap Registry

For disconnected environments:
- Maintain a local registry mirror with all required base images and project images
- Use `docker save` / `docker load` or `cosign` bundles for air-gap transfers
- Verify image signatures after transfer to ensure integrity

## 10. Orchestration Readiness

### 10.1 Kubernetes-Ready Containers

Ensure every container is "orchestration-ready" by implementing:

| Requirement | How |
|---|---|
| Health probes | `/healthz`, `/readyz`, `/startupz` endpoints |
| Graceful shutdown | Handle `SIGTERM`; drain in-flight work |
| Configuration | Environment variables; never bake config into the image |
| Statelessness | No local state that must survive restarts; use external stores |
| Horizontal scaling | No assumption of singleton; handle concurrent instances |
| Resource bounds | Set memory/CPU limits and requests |
| Security context | Non-root, read-only FS, dropped capabilities |

### 10.2 Init Containers Pattern

Use init containers for one-time setup tasks that must complete before the main container starts:

- Database schema migration
- Configuration fetching from a vault
- Certificate or secret provisioning
- Dependency readiness verification

## 11. AI Agent Instructions

When generating Docker or container-related code:

1. Always use multi-stage builds — separate build and runtime stages
2. Always use distroless or Alpine-based images for the final stage — never full OS images
3. Always run as non-root in the final image stage
4. Always include health check endpoints (liveness, readiness, startup)
5. Always set resource limits (CPU, memory) on containers
6. Never embed secrets in Docker images or Dockerfiles
7. Never use `latest` tags for base images — pin to specific versions
8. Log to stdout/stderr — never write logs to files inside the container
9. Handle `SIGTERM` for graceful shutdown with configurable drain timeout
10. Apply security context: read-only FS, drop all capabilities, seccomp profile
