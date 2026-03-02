# Best Practices for Development in Docker Environment

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Development Guideline
> **Parent**: `00_master_policy.md`
> **Last Updated**: 2026-03-02

---

## 1. Purpose

This document defines best practices for developing, building, and running services in Docker and Docker Compose environments. These guidelines ensure consistent local development workflows, reproducible builds, and smooth transitions from development to staging and production.

## 2. Docker Compose for Local Development

### 2.1 Compose File Organization

Organize Docker Compose files by purpose to keep configurations manageable:

| File | Purpose |
|---|---|
| `docker-compose.yml` | Core infrastructure services (databases, message brokers, caches) |
| `docker-compose.services.yml` | Application microservices |
| `docker-compose.override.yml` | Developer-specific overrides (auto-loaded by Docker Compose) |
| `docker-compose.test.yml` | Test-specific configurations (test databases, mock services) |

Use the `extends` or `include` directives to share common configuration across files. Avoid duplicating service definitions.

### 2.2 Service Profiles

Use Docker Compose profiles to selectively start groups of services:

```yaml
services:
  database:
    image: clickhouse/clickhouse-server:latest
    profiles: ["infra", "full"]

  broker:
    image: redpandadata/redpanda:latest
    profiles: ["infra", "full"]

  my-service:
    build: ./svc-my-service
    profiles: ["app", "full"]
    depends_on:
      database:
        condition: service_healthy
```

```bash
# Start only infrastructure
docker compose --profile infra up -d

# Start everything
docker compose --profile full up -d

# Start specific services
docker compose up -d database broker
```

### 2.3 Service Dependencies and Startup Order

Always use `depends_on` with health check conditions to manage startup order. Never rely on `sleep` or manual delays:

```yaml
services:
  my-service:
    depends_on:
      database:
        condition: service_healthy
      broker:
        condition: service_healthy
```

Define health checks on every infrastructure service:

```yaml
services:
  database:
    healthcheck:
      test: ["CMD", "clickhouse-client", "--query", "SELECT 1"]
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 10s
```

### 2.4 Environment Variables

- Use `.env` files for default development configuration
- Provide a `.env.example` with documented placeholder values committed to the repository
- Never commit `.env` files containing real secrets — add `.env` to `.gitignore`
- Use `environment` or `env_file` directives in Compose — never hardcode secrets in Compose files

```yaml
services:
  my-service:
    env_file:
      - .env
    environment:
      - SERVICE_LOG_LEVEL=debug  # Dev-specific override
```

### 2.5 Docker Compose Networking

- Use explicit named networks instead of relying on the default bridge network
- Use service names as hostnames for inter-service communication (Docker DNS resolves service names automatically)
- Expose only necessary ports to the host — other services communicate over the Docker network

```yaml
networks:
  app-network:
    driver: bridge

services:
  database:
    networks:
      - app-network
    # No 'ports' — only accessible within the Docker network

  api-gateway:
    networks:
      - app-network
    ports:
      - "8080:8080"  # Only exposed service
```

## 3. Dockerfile Best Practices

### 3.1 Multi-Stage Builds

Always use multi-stage builds to minimize final image size and reduce the attack surface:

```dockerfile
# CLASSIFICATION: UNCLASSIFIED

# --- Stage 1: Build ---
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server/

# --- Stage 2: Runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

### 3.2 Layer Caching

Order Dockerfile instructions from least-frequently changed to most-frequently changed to maximize layer caching:

1. Base image selection
2. System dependencies installation
3. Dependency manifest copy (`go.mod`, `package.json`)
4. Dependency download (`go mod download`, `npm ci`)
5. Source code copy
6. Build command

**Never** copy the entire source tree before downloading dependencies — this invalidates the dependency cache on every code change.

### 3.3 Minimal Base Images

| Use Case | Recommended Base | Avoid |
|---|---|---|
| Go services (production) | `gcr.io/distroless/static` | `ubuntu`, `debian` |
| Go services (build stage) | `golang:X.Y-alpine` | `golang:X.Y` (full Debian) |
| Node.js (production) | `node:X-alpine` or distroless | `node:X` (full Debian) |
| Debug / troubleshooting | `alpine:3.x` | `ubuntu` |

### 3.4 Security Hardening in Dockerfiles

- **Always** run as a non-root user in the final stage
- **Never** include secrets, credentials, or private keys in Docker images
- **Never** use `latest` tags for base images in production — pin to specific versions
- Use `.dockerignore` to exclude unnecessary files (`.git`, `node_modules`, test fixtures, documentation)
- Set `HEALTHCHECK` instructions in the Dockerfile or Compose file

```dockerfile
# Run as non-root
USER 65534:65534

# Health check
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD ["/server", "--health-check"]
```

### 3.5 Build Arguments and Labels

Use build arguments for version injection and labels for metadata:

```dockerfile
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.source="https://github.com/org/repo"
```

## 4. Volume Mounts for Development

### 4.1 Source Code Mounts

Mount source code as volumes for live reload during development. This avoids rebuilding the container on every code change:

```yaml
services:
  my-service:
    build:
      context: ./svc-my-service
      target: builder  # Use the build stage, not the final stage
    volumes:
      - ./svc-my-service:/app
      - go-mod-cache:/go/pkg/mod  # Cache Go modules across restarts
    command: ["go", "run", "./cmd/server/"]
```

### 4.2 Named Volumes for Persistent Data

Use named volumes for data that must survive container restarts (databases, message broker state):

```yaml
volumes:
  db-data:
  broker-data:

services:
  database:
    volumes:
      - db-data:/var/lib/clickhouse
  broker:
    volumes:
      - broker-data:/var/lib/redpanda/data
```

### 4.3 Volume Permissions

- Ensure the container user has read/write permissions on mounted volumes
- Use `user:` in Compose or `chown` in the Dockerfile to set appropriate ownership
- On Linux, host UID/GID must match the container user to avoid permission issues

## 5. Container Resource Limits

### 5.1 Always Set Limits in Development

Set CPU and memory limits even in development to surface resource issues early:

```yaml
services:
  my-service:
    deploy:
      resources:
        limits:
          cpus: "1.0"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 128M
```

### 5.2 Resource Limit Guidelines

| Component Type | CPU Limit | Memory Limit | Rationale |
|---|---|---|---|
| Application services | 0.5–2.0 cores | 256M–1G | Typical Go/Node microservice |
| Databases (OLAP) | 2.0–4.0 cores | 1G–4G | Query and write workloads |
| Message brokers | 1.0–2.0 cores | 512M–2G | Throughput-dependent |
| UI dev servers | 0.5–1.0 cores | 512M–1G | Build/HMR tooling |

## 6. Logging and Debugging

### 6.1 Container Logging Best Practices

- Log to `stdout` and `stderr` — never write logs to files inside the container
- Use JSON structured logging for consistency with the observability stack
- Use Docker Compose log drivers for centralized log collection in development
- Limit log output with `logging.options.max-size` to prevent disk exhaustion

```yaml
services:
  my-service:
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
```

### 6.2 Viewing Logs

```bash
# Follow logs for a specific service
docker compose logs -f my-service

# Follow logs for multiple services
docker compose logs -f my-service database

# Show last N lines
docker compose logs --tail=100 my-service

# Filter by time
docker compose logs --since="5m" my-service
```

### 6.3 Debugging Inside Containers

- Use `docker compose exec` to open a shell inside a running container
- For distroless images, use ephemeral debug containers or a sidecar with debugging tools
- Use Delve (`dlv`) for Go remote debugging inside containers:

```yaml
services:
  my-service-debug:
    build:
      context: ./svc-my-service
      target: builder
    command: ["dlv", "debug", "--headless", "--listen=:2345", "--api-version=2", "./cmd/server/"]
    ports:
      - "2345:2345"  # Delve debugger port
```

### 6.4 Health Check Debugging

When a container fails health checks:

1. Check `docker inspect <container>` for health check status and output
2. Run the health check command manually inside the container
3. Verify the health check endpoint is bound to `0.0.0.0` (not `localhost`) inside the container
4. Ensure the service is fully initialized before the health check starts (`start_period`)

## 7. Container Build Optimization

### 7.1 Build Cache Management

- Use BuildKit (`DOCKER_BUILDKIT=1`) for improved caching and parallel builds
- Use `--mount=type=cache` for build tool caches (Go module cache, NPM cache):

```dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/server ./cmd/server/
```

### 7.2 Reducing Image Size

| Technique | Savings | How |
|---|---|---|
| Multi-stage builds | 80–95% | Copy only the binary to the final stage |
| Strip debug symbols | 20–40% | `go build -ldflags="-s -w"` |
| Distroless base | 90%+ vs Debian | No shell, no package manager |
| `.dockerignore` | Variable | Exclude `.git`, docs, test fixtures |
| Compress binaries | 50–70% | Use UPX (trade-off: slower startup) |

### 7.3 Image Tagging Strategy

| Tag | Use |
|---|---|
| `<service>:<git-sha>` | Every CI build — immutable reference |
| `<service>:<semver>` | Release versions |
| `<service>:latest` | Local development only — never in production |
| `<service>:dev` | Development builds with debug tooling |

## 8. Docker Compose Development Workflow

### 8.1 Recommended Workflow

```bash
# 1. Start infrastructure
docker compose --profile infra up -d

# 2. Wait for health checks to pass
docker compose --profile infra ps

# 3. Run the service you're developing locally (outside Docker)
#    This allows IDE debugging, fast recompile, etc.
cd svc-my-service && go run ./cmd/server/

# 4. Or run all services in Docker
docker compose --profile full up -d

# 5. Run tests against the Docker infrastructure
go test -tags=integration ./...

# 6. Clean up
docker compose down           # Stop and remove containers
docker compose down -v        # Also remove volumes (fresh state)
```

### 8.2 Selective Rebuild

Rebuild only the changed service, not the entire stack:

```bash
# Rebuild and restart one service
docker compose up -d --build my-service

# Force a clean rebuild (no cache)
docker compose build --no-cache my-service
```

### 8.3 Shell Scripts for Common Operations

Wrap frequently used Docker Compose commands in shell scripts for consistency:

```bash
#!/bin/bash
# scripts/dev/start.sh — Start the development environment
set -euo pipefail

echo "Starting infrastructure services..."
docker compose --profile infra up -d --wait

echo "Infrastructure ready. Run your service locally or use:"
echo "  docker compose --profile full up -d"
```

## 9. Registry and Image Management

### 9.1 Private Registry Best Practices

- Use a private container registry for all project images
- Enable vulnerability scanning on the registry
- Set retention policies to clean up old images (keep last N tags per repository)
- Sign images with `cosign` or Docker Content Trust before pushing
- Attach SBOM (Software Bill of Materials) to every production image

### 9.2 Pull-Through Cache

For air-gapped or bandwidth-constrained environments, configure a pull-through registry cache to avoid re-downloading base images from the public internet.

## 10. AI Agent Instructions

When generating Docker-related code:

1. Always use multi-stage builds in Dockerfiles — separate build and runtime stages
2. Always run containers as non-root users in the final stage
3. Always use health checks on infrastructure services in Docker Compose
4. Always use `depends_on` with `condition: service_healthy` for service ordering
5. Never hardcode secrets in Dockerfiles or Compose files — use environment variables
6. Never use `latest` tags for base images — pin to specific versions
7. Use named volumes for persistent data; bind mounts for source code in development
8. Set resource limits (CPU, memory) on all containers
9. Log to stdout/stderr — never write log files inside containers
10. Use `.dockerignore` to exclude unnecessary files from the build context
