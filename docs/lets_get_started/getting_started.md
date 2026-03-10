<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA — Getting Started Guide

> **CLASSIFICATION: UNCLASSIFIED**
> **Project**: Real-Time Situational Awareness & Risk Assessment (RTSA)
> **Audience**: Developers setting up the local development environment
> **Last Updated**: 2026-02-23

---

## Table of Contents

1. [Prerequisites Overview](#1-prerequisites-overview)
2. [Step 1 — Install System Dependencies](#2-step-1--install-system-dependencies)
3. [Step 2 — Install Go Toolchain](#3-step-2--install-go-toolchain)
4. [Step 3 — Install Protobuf & gRPC Tools](#4-step-3--install-protobuf--grpc-tools)
5. [Step 4 — Install Node.js & Frontend Tools](#5-step-4--install-nodejs--frontend-tools)
6. [Step 5 — Install Docker & Container Tools](#6-step-5--install-docker--container-tools)
7. [Step 6 — Install Security & Linting Tools](#7-step-6--install-security--linting-tools)
8. [Step 7 — Clone the Repository](#8-step-7--clone-the-repository)
9. [Step 8 — Run Automated Setup Script](#9-step-8--run-automated-setup-script)
10. [Step 9 — Configure Environment Variables](#10-step-9--configure-environment-variables)
11. [Step 10 — Generate TLS Certificates (Dev)](#11-step-10--generate-tls-certificates-dev)
12. [Step 11 — Start the Development Stack](#12-step-11--start-the-development-stack)
13. [Step 12 — Verify the Installation](#13-step-12--verify-the-installation)
14. [Step 13 — Run Tests](#14-step-13--run-tests)
15. [Development Workflow Reference](#15-development-workflow-reference)
16. [Troubleshooting](#16-troubleshooting)

---

## 1. Prerequisites Overview

The following tools are required to develop and run RTSA locally. The automated setup script (`scripts/setup/setup-dev.sh`) installs most tools, but some require manual steps detailed below.

| Tool                   | Version     | Install Method              | Purpose                         |
| ---------------------- | ----------- | --------------------------- | ------------------------------- |
| **Git**                | 2.40+       | Manual                      | Source control                  |
| **Go**                 | 1.22+       | Manual                      | Service development             |
| **buf**                | 1.32+       | Automated                   | Protobuf toolchain              |
| **protoc-gen-go**      | latest      | Automated                   | Go gRPC code generation         |
| **protoc-gen-go-grpc** | latest      | Automated                   | Go gRPC code generation         |
| **Node.js**            | 20 LTS      | Manual                      | Frontend development            |
| **pnpm**               | 9+          | Automated                   | Frontend package manager        |
| **Rust**               | 1.77+       | Manual                      | Wasm decoder compilation        |
| **wasm-pack**          | 0.12+       | Manual (`cargo install`)    | Rust → Wasm build tool          |
| **Docker Desktop**     | 26+         | Manual                      | Container runtime               |
| **Docker Compose**     | v2 (plugin) | Bundled with Docker Desktop | Local dev stack                 |
| **kubectl**            | 1.29+       | Automated                   | Kubernetes CLI (staging/prod)   |
| **helm**               | 3.14+       | Automated                   | Kubernetes chart deployment     |
| **k3d**                | 5.6+        | Automated                   | Local K3s cluster (optional)    |
| **gitleaks**           | 8.18+       | Automated                   | Secret scanning                 |
| **gosec**              | 2.19+       | Automated                   | Go SAST                         |
| **govulncheck**        | latest      | Automated                   | Go vulnerability scanning       |
| **semgrep**            | latest      | Automated                   | Multi-language SAST             |
| **golangci-lint**      | 1.57+       | Automated                   | Go linting                      |
| **trivy**              | 0.50+       | Automated                   | Container image scanning        |
| **mkcert**             | 1.4+        | Automated                   | Local TLS certificate authority |
| **make**               | 4.3+        | Manual (Linux/macOS)        | Build automation                |

**Estimated Setup Time**: 30–60 minutes depending on download speeds.

---

## 2. Step 1 — Install System Dependencies

### Linux (Ubuntu / Debian)

```bash
sudo apt-get update
sudo apt-get install -y \
  git \
  curl \
  wget \
  unzip \
  make \
  build-essential \
  ca-certificates \
  gnupg \
  lsb-release \
  jq \
  openssl
```

### macOS

Install [Homebrew](https://brew.sh/) if not already installed:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Then install base tools:

```bash
brew install git curl wget make jq openssl
```

### Windows (WSL2 — Required)

RTSA development on Windows requires **WSL2 (Windows Subsystem for Linux)** with Ubuntu 22.04 or 24.04.

1. Open PowerShell as Administrator and run:

   ```powershell
   wsl --install -d Ubuntu-24.04
   ```

2. Restart your machine when prompted.

3. After restart, open the Ubuntu terminal, create your user, then install base dependencies:

   ```bash
   sudo apt-get update
   sudo apt-get install -y git curl wget unzip make build-essential ca-certificates gnupg lsb-release jq openssl
   ```

4. For VS Code users, install the [WSL extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl) and open VS Code from within WSL:
   ```bash
   code .
   ```

> All subsequent steps are run **inside the WSL2 terminal**, not PowerShell.

---

## 3. Step 2 — Install Go Toolchain

> **Required Version**: Go 1.22 or later.

### Linux / WSL2

```bash
# Download Go 1.22
GO_VERSION="1.22.4"
curl -LO "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"

# Remove any prior Go installation and extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"

# Add Go to PATH (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

### macOS

```bash
brew install go@1.22
# or download from https://go.dev/dl/
```

### Verify

```bash
go version
# Expected: go version go1.22.x linux/amd64
```

---

## 4. Step 3 — Install Protobuf & gRPC Tools

> The automated setup script handles these, but you can install manually if needed.

```bash
# Install buf CLI (Protobuf toolchain)
BUF_VERSION="1.32.2"
curl -sSL "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-x86_64" \
  -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# Install Go Protobuf code generators
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify
buf --version
protoc-gen-go --version
```

---

## 5. Step 4 — Install Node.js & Frontend Tools

> **Required Version**: Node.js 20 LTS

### Using nvm (Recommended)

```bash
# Install nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc

# Install Node.js 20 LTS
nvm install 20
nvm use 20
nvm alias default 20

# Verify
node --version   # Expected: v20.x.x
npm --version
```

### macOS

```bash
brew install node@20
```

### Install pnpm

```bash
npm install -g pnpm@9

# Verify
pnpm --version
```

---

## 6. Step 5 — Install Docker & Container Tools

### Docker Desktop (Windows / macOS)

1. Download **Docker Desktop** from [https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/)
2. Run the installer and follow the on-screen prompts.
3. **Windows**: Enable WSL2 integration in Docker Desktop settings → Resources → WSL Integration.
4. Verify installation:
   ```bash
   docker version
   docker compose version
   ```

### Docker Engine (Linux)

```bash
# Add Docker's official GPG key
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Add the Docker repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Add your user to the docker group (no sudo required)
sudo usermod -aG docker "$USER"
newgrp docker

# Verify
docker version
docker compose version
```

---

## 7. Step 6 — Install Security & Linting Tools

> These are also installed by the automated script. Listed here for reference.

```bash
# gitleaks — secret detection
GITLEAKS_VERSION="8.18.4"
curl -sSfL "https://github.com/zricethezav/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_linux_x64.tar.gz" \
  | tar -xz -C /usr/local/bin gitleaks

# gosec — Go SAST
go install github.com/securego/gosec/v2/cmd/gosec@latest

# govulncheck — Go vulnerability scanner
go install golang.org/x/vuln/cmd/govulncheck@latest

# golangci-lint — Go linter
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" v1.57.2

# trivy — container image scanner
TRIVY_VERSION="0.50.4"
curl -sSfL "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.tar.gz" \
  | tar -xz -C /usr/local/bin trivy

# semgrep — multi-language SAST
pip3 install semgrep

# mkcert — local CA for dev TLS
go install filippo.io/mkcert@latest
# Linux: also install nss-tools for Firefox/Chrome trust
sudo apt-get install -y libnss3-tools
mkcert -install
```

---

## 8. Step 7 — Clone the Repository

```bash
# Clone into your preferred workspace directory
mkdir -p ~/workspace && cd ~/workspace
git clone https://github.com/<org>/rtsa.git
cd rtsa
```

> Replace `<org>` with the actual GitHub organization name provided by your team lead.

---

## 9. Step 8 — Run Automated Setup Script

The automated script installs any remaining tools, validates versions, configures `git`, and prepares your workspace.

```bash
# Make scripts executable
chmod +x scripts/setup/*.sh scripts/dev/*.sh

# Run full developer setup
./scripts/setup/setup-dev.sh
```

The script will:

- Validate that all required tools are installed (and install missing ones where possible)
- Configure Git commit message template and hooks
- Install Go workspace dependencies (`go mod download`)
- Install Node.js frontend dependencies (`pnpm install`)
- Generate Protobuf code from `.proto` files
- Create a local `.env` file from `.env.example`
- Generate dev TLS certificates
- Pull Docker images for the dev stack

---

## 10. Step 9 — Configure Environment Variables

The automated script creates `.env` from `.env.example`. Review and adjust values as needed:

```bash
# Open .env for review
nano .env        # or: code .env
```

Key variables:

```dotenv
# CLASSIFICATION: UNCLASSIFIED
# Development environment configuration — DO NOT commit real secrets

# Redpanda
REDPANDA_BROKERS=localhost:19092
REDPANDA_SCHEMA_REGISTRY_URL=http://localhost:18081
REDPANDA_ADMIN_API_URL=http://localhost:19644

# ClickHouse
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=rtsa
CLICKHOUSE_USER=rtsa_dev
CLICKHOUSE_PASSWORD=dev_password_change_me

# Service TLS (generated by setup script)
TLS_CERT_DIR=./certs/dev
TLS_CA_CERT=./certs/dev/ca.crt
TLS_SERVER_CERT=./certs/dev/server.crt
TLS_SERVER_KEY=./certs/dev/server.key

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
PROMETHEUS_PORT=9090
GRAFANA_PORT=3000

# Development flags
LOG_LEVEL=debug
LOG_FORMAT=text   # Use 'json' for production-like output
DEV_MODE=true
```

> **Never commit `.env`** — it is listed in `.gitignore`. Use real secrets management (Vault / AWS Secrets Manager) in staging and production.

---

## 11. Step 10 — Generate TLS Certificates (Dev)

mTLS is required for all gRPC inter-service communication, even in local development.

```bash
# Run the certificate generation script
./scripts/setup/gen-dev-certs.sh
```

This script uses `mkcert` to:

1. Create a local Certificate Authority (CA) trusted by your system.
2. Issue a wildcard server certificate for `*.rtsa.local` and `localhost`.
3. Output certificates to `./certs/dev/`.

```
certs/dev/
├── ca.crt              # Local dev CA certificate
├── ca.key              # Local dev CA private key (never commit)
├── server.crt          # Server certificate (SAN: *.rtsa.local, localhost)
├── server.key          # Server private key (never commit)
├── client.crt          # Client certificate for mTLS
└── client.key          # Client private key (never commit)
```

> The `certs/` directory is in `.gitignore`. Never commit private keys.

---

## 12. Step 11 — Start the Development Stack

```bash
# Start all infrastructure services (Redpanda, ClickHouse, Grafana, etc.)
docker compose -f deploy/docker-compose.yml up -d

# Check that all containers are running
docker compose -f deploy/docker-compose.yml ps
```

The dev stack includes:

| Service                  | URL / Port                                     | Purpose                     |
| ------------------------ | ---------------------------------------------- | --------------------------- |
| Redpanda broker          | `localhost:19092`                              | Kafka-compatible API        |
| Redpanda Console         | [http://localhost:8080](http://localhost:8080) | Broker UI                   |
| Redpanda Schema Registry | `localhost:18081`                              | Protobuf schema registry    |
| Redpanda Admin API       | `localhost:19644`                              | Cluster management          |
| ClickHouse HTTP          | [http://localhost:8123](http://localhost:8123) | HTTP query interface        |
| ClickHouse native        | `localhost:9000`                               | Native client port          |
| Prometheus               | [http://localhost:9090](http://localhost:9090) | Metrics scraping            |
| Grafana                  | [http://localhost:3000](http://localhost:3000) | Dashboards (admin/admin)    |
| Loki                     | `localhost:3100`                               | Log aggregation             |
| Tempo                    | `localhost:4317` (OTLP)                        | Distributed tracing         |
| Envoy Gateway            | `localhost:8443`                               | gRPC-Web API gateway (mTLS) |

### Initialize Redpanda Topics

```bash
# Create all required Redpanda topics for development
./scripts/dev/init-topics.sh
```

### Initialize ClickHouse Schema

```bash
# Apply ClickHouse schema migrations
./scripts/dev/init-clickhouse.sh
```

---

## 13. Step 12 — Verify the Installation

```bash
# Run the full health check script
./scripts/dev/health-check.sh
```

The health check verifies:

- All Docker containers are running
- Redpanda cluster is healthy and topics exist
- ClickHouse is reachable and the `rtsa` database exists
- Protobuf code generation is up to date
- Go services compile without errors
- Frontend dependencies are installed
- TLS certificates are valid

Expected output:

```
[✓] Docker daemon running
[✓] Redpanda broker reachable (localhost:19092)
[✓] Redpanda topics initialized (12/12)
[✓] ClickHouse reachable (localhost:9000)
[✓] ClickHouse schema applied (rtsa database)
[✓] Protobuf generation up to date
[✓] Go compilation OK (./services/...)
[✓] Frontend dependencies installed
[✓] Dev TLS certificates valid (expires: ...)

All checks passed. Development environment is ready.
```

---

## 14. Step 13 — Run Tests

### Go Unit Tests

```bash
# Run all unit tests with race detector and coverage
./scripts/dev/test-go.sh

# Or manually per service
cd services/radar-ingestion
go test -race -cover ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Frontend Unit Tests

```bash
cd ui
pnpm test --coverage

# Watch mode during development
pnpm test --watch
```

### Integration Tests

Integration tests require the full dev stack to be running:

```bash
# Ensure docker compose stack is up
docker compose -f deploy/docker-compose.yml up -d

# Run integration tests
./scripts/dev/test-integration.sh
```

### All Tests (Pre-PR Check)

Run the full local gate before opening a Pull Request:

```bash
./scripts/dev/pre-pr-check.sh
```

This mirrors all 5 CI security gates locally:

1. Secret scan (gitleaks)
2. Classification header check
3. Go formatting check
4. Build and proto generation
5. Unit tests ≥ 80% coverage
6. SAST (gosec, semgrep)
7. Dependency vulnerability scan

---

## 15. Development Workflow Reference

### Starting a Feature

```bash
# Create a feature branch from main
git checkout main
git pull --rebase
git checkout -b feature/RTSA-XXX-brief-description

# Start the dev stack if not already running
docker compose -f deploy/docker-compose.yml up -d
```

### Regenerating Protobuf Code

After modifying any `.proto` file:

```bash
buf generate proto/
```

Or from a service directory:

```bash
make proto
```

### Running a Single Service

Each service has a `Makefile` with standard targets:

```bash
cd services/radar-ingestion

make run        # Run the service locally
make test       # Run unit tests
make lint       # Run golangci-lint
make proto      # Regenerate protobuf code
make build      # Build the service binary
make docker     # Build the Docker image
```

### Stopping the Dev Stack

```bash
docker compose -f deploy/docker-compose.yml down

# To also remove volumes (full reset)
docker compose -f deploy/docker-compose.yml down -v
```

### Checking Logs

```bash
# View all container logs
docker compose -f deploy/docker-compose.yml logs -f

# View a specific service's logs
docker compose -f deploy/docker-compose.yml logs -f redpanda
docker compose -f deploy/docker-compose.yml logs -f clickhouse
```

---

## 16. Troubleshooting

### Docker Compose fails to start

```bash
# Check available disk space (ClickHouse and Redpanda can be large)
df -h

# Remove old volumes and retry
docker compose -f deploy/docker-compose.yml down -v
docker system prune -f
docker compose -f deploy/docker-compose.yml up -d
```

### Redpanda broker unreachable

```bash
# Check Redpanda container logs
docker compose -f deploy/docker-compose.yml logs redpanda

# Verify the port is not already in use
lsof -i :19092
```

### gRPC certificate errors

```bash
# Regenerate dev certificates
./scripts/setup/gen-dev-certs.sh

# Verify certificate SANs
openssl x509 -in certs/dev/server.crt -noout -text | grep -A5 "Subject Alternative Name"
```

### Protobuf generation errors

```bash
# Ensure buf plugins are up to date
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Verify buf configuration
buf lint proto/
buf generate proto/
```

### Go module dependency issues

```bash
# Tidy all service modules
find services -name "go.mod" -execdir go mod tidy \;

# Verify all modules download correctly
find services -name "go.mod" -execdir go mod download \;
```

### Port conflicts

Default ports used by the dev stack:

| Port  | Service                  |
| ----- | ------------------------ |
| 8080  | Redpanda Console         |
| 8123  | ClickHouse HTTP          |
| 9000  | ClickHouse native        |
| 9090  | Prometheus               |
| 3000  | Grafana                  |
| 3100  | Loki                     |
| 4317  | OTLP (Tempo)             |
| 8443  | Envoy gRPC-Web gateway   |
| 19092 | Redpanda Kafka API       |
| 18081 | Redpanda Schema Registry |
| 19644 | Redpanda Admin API       |

Adjust port mappings in `deploy/docker-compose.yml` if any conflict with existing services on your machine.

---

## Additional Resources

| Resource                | Link                                                                                                                             |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| README                  | [README.md](README.md)                                                                                                           |
| Master Policy           | [docs/sdlc_guidelines/00_master_policy.md](docs/sdlc_guidelines/00_master_policy.md)                                             |
| Go Standards            | [docs/sdlc_guidelines/04_coding_standards/go_standards.md](docs/sdlc_guidelines/04_coding_standards/go_standards.md)             |
| High-Level Architecture | [docs/architecture/high_level_architecture.md](docs/architecture/high_level_architecture.md)                                     |
| Redpanda Guidelines     | [docs/sdlc_guidelines/08_tech_specific/redpanda_guidelines.md](docs/sdlc_guidelines/08_tech_specific/redpanda_guidelines.md)     |
| ClickHouse Guidelines   | [docs/sdlc_guidelines/08_tech_specific/clickhouse_guidelines.md](docs/sdlc_guidelines/08_tech_specific/clickhouse_guidelines.md) |
| CI/CD Pipeline          | [docs/sdlc_guidelines/06_integration_cicd/ci_cd_pipeline.md](docs/sdlc_guidelines/06_integration_cicd/ci_cd_pipeline.md)         |

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
