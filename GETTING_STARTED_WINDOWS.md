<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA — Getting Started Guide (Windows)

> **CLASSIFICATION: UNCLASSIFIED**
> **Project**: Real-Time Situational Awareness & Risk Assessment (RTSA)
> **Audience**: Windows developers setting up a local development environment
> **Platform**: Windows 10 (21H2+) or Windows 11 with WSL2
> **Last Updated**: 2026-02-23

---

## Overview — What You Will Install

RTSA development on Windows runs entirely inside **WSL2 (Windows Subsystem for Linux)** with Ubuntu 24.04. All Go services, scripts, and Docker tooling execute in WSL2. Your Windows host only needs Docker Desktop and VS Code.

```
Windows Host
 ├── Docker Desktop  (with WSL2 integration)
 ├── VS Code         (with Remote - WSL extension)
 └── WSL2
      └── Ubuntu 24.04
           ├── Go 1.22+, buf, protobuf plugins
           ├── Node.js 20 LTS + pnpm
           ├── Security tools (gitleaks, gosec, trivy, ...)
           ├── mkcert (dev TLS certificate authority)
           └── RTSA repository + dev stack
```

---

## Step-By-Step Order

> **Important**: Two manual steps must complete (and the system may need a restart) **before** running any automated scripts. Follow the steps in order.

```
Phase 1 — Manual prerequisites (Windows host, ~15 min)
  Step 1  Check Windows requirements
  Step 2  Enable WSL2
  Step 3  Install Docker Desktop
  Step 4  Install Git for Windows
  Step 5  Install VS Code  (optional but strongly recommended)

Phase 2 — Automated setup (WSL2, ~20–40 min)
  Step 6  Clone the repository inside WSL2
  Step 7  Run the automated bootstrap script

Phase 3 — Manual finish (WSL2, ~5 min)
  Step 8  Review environment variables (.env)
  Step 9  Start the development stack
  Step 10 Initialize Redpanda topics
  Step 11 Initialize ClickHouse schema
  Step 12 Verify the installation
  Step 13 Run tests
```

---

## Phase 1 — Manual Prerequisites

### Step 1 — Check Windows Requirements

Verify your system meets the minimum requirements before proceeding.

**Required**:

| Requirement        | Minimum                                     | Check command                                     |
| ------------------ | ------------------------------------------- | ------------------------------------------------- |
| Windows version    | Windows 10 21H2 (build 19044) or Windows 11 | `winver` in Run dialog                            |
| RAM                | 16 GB                                       | Task Manager → Performance                        |
| Free disk space    | 40 GB                                       | File Explorer → This PC                           |
| CPU virtualisation | Enabled in BIOS/UEFI                        | Task Manager → Performance → CPU → Virtualization |

If virtualisation is **Disabled**, reboot into your BIOS/UEFI firmware settings and enable **Intel VT-x** or **AMD-V** before continuing.

---

### Step 2 — Enable WSL2

> Skip this step if you already have WSL2 running with Ubuntu 24.04.

**Option A — Automated (recommended)**

Open **PowerShell as Administrator** and run:

```powershell
wsl --install -d Ubuntu-24.04
```

This single command enables the WSL2 feature, installs the WSL2 kernel, and installs Ubuntu 24.04. **Restart your computer when prompted.**

**Option B — Manual (if Option A fails)**

Open **PowerShell as Administrator** and run each command separately:

```powershell
# 1. Enable WSL feature
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart

# 2. Enable Virtual Machine Platform
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# 3. Restart now
Restart-Computer
```

After restart, open **PowerShell as Administrator** again:

```powershell
# 4. Set WSL2 as the default version
wsl --set-default-version 2

# 5. Update the WSL2 kernel
wsl --update

# 6. Install Ubuntu 24.04
wsl --install -d Ubuntu-24.04
```

**First-time Ubuntu setup**:

After Ubuntu installs, a terminal window opens asking you to create a UNIX username and password. Choose a simple username (e.g., your first name) and a password you will remember. This is your WSL2 Linux account — it is separate from your Windows account.

```
Enter new UNIX username: yourname
New password: ********
Retype new password: ********
```

**Verify WSL2 is working**:

```powershell
wsl --list --verbose
```

Expected output:

```
  NAME            STATE           VERSION
* Ubuntu-24.04    Running         2
```

---

### Step 3 — Install Docker Desktop

Docker Desktop provides the Docker engine with WSL2 integration, which powers the full RTSA dev stack (Redpanda, ClickHouse, Grafana, etc.).

1. Download **Docker Desktop for Windows** from:
   [https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/)

2. Run the installer (`Docker Desktop Installer.exe`).
   - When prompted, ensure **"Use WSL 2 instead of Hyper-V"** is checked.
   - When prompted, ensure **"Add shortcut to desktop"** is checked (optional).

3. After installation completes, **restart your computer** if prompted.

4. Open Docker Desktop. Wait for the whale icon in the system tray to show **"Docker Desktop is running"**.

5. Configure WSL2 integration:
   - Open Docker Desktop → click the gear icon (Settings)
   - Go to **Resources** → **WSL Integration**
   - Enable **"Enable integration with my default WSL distro"**
   - Also toggle on **Ubuntu-24.04**
   - Click **Apply & Restart**

6. Verify Docker is accessible from WSL2:

   Open your Ubuntu terminal (search "Ubuntu" in Start menu) and run:

   ```bash
   docker version
   docker compose version
   ```

   Both commands should return version information without errors.

---

### Step 4 — Install Git for Windows

Git for Windows is needed to clone the repository from your Windows host. The WSL2 environment has its own Git installation (handled by the automated script).

1. Download from: [https://git-scm.com/download/win](https://git-scm.com/download/win)
2. Run the installer with default options.
   - On the **"Adjusting your PATH"** screen, select **"Git from the command line and also from 3rd-party software"**.
   - On the **"Configuring line ending conversions"** screen, select **"Checkout as-is, commit Unix-style line endings"** (important for WSL2 compatibility).

3. Verify:

   ```powershell
   git --version
   ```

4. Configure your identity (required for commits):

   ```powershell
   git config --global user.name "Your Name"
   git config --global user.email "you@example.com"
   ```

---

### Step 5 — Install VS Code (Recommended)

VS Code with the WSL extension gives you a full IDE experience running inside WSL2.

1. Download from: [https://code.visualstudio.com/](https://code.visualstudio.com/)
2. Run the installer with default options. Ensure **"Add to PATH"** is checked.
3. After installation, install the WSL extension:

   ```powershell
   code --install-extension ms-vscode-remote.remote-wsl
   code --install-extension ms-vscode-remote.remote-containers
   ```

4. Recommended additional extensions (install from within VS Code WSL session later):
   - `golang.go` — Go language support
   - `zxh404.vscode-proto3` — Protobuf syntax
   - `dbaeumer.vscode-eslint` — ESLint for React/TypeScript
   - `esbenp.prettier-vscode` — Prettier formatter
   - `bradlc.vscode-tailwindcss` — Tailwind CSS IntelliSense

---

## Phase 2 — Automated Setup

> All steps from here onwards run **inside the WSL2 Ubuntu terminal** unless otherwise stated.
>
> Open the Ubuntu terminal: search **"Ubuntu"** in the Windows Start menu, or run `wsl` in PowerShell.

---

### Step 6 — Clone the Repository

Inside the **Ubuntu WSL2 terminal**:

```bash
# Install git inside WSL2 (if not already present)
sudo apt-get update -q
sudo apt-get install -y git

# Configure git identity (if not already set)
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
git config --global pull.rebase true

# Clone into your Linux home directory (NOT /mnt/c — see note below)
mkdir -p ~/workspace
cd ~/workspace
git clone https://github.com/<org>/rtsa.git
cd rtsa
```

> **Important — Where to clone**: Always clone inside the WSL2 filesystem (`~/workspace/rtsa`), **not** on the Windows filesystem (`/mnt/c/...`). Accessing Windows drives from WSL2 is significantly slower and causes file permission issues with scripts and Go tooling.

> Replace `<org>` with the GitHub organization name provided by your team lead.

---

### Step 7 — Run the Automated Bootstrap Script

This single script installs all remaining tools inside WSL2 and prepares the workspace. It handles everything listed in the prerequisites that has **"Automated"** in the Install Method column.

```bash
# Make all scripts executable
find scripts/ -name "*.sh" -exec chmod +x {} \;

# Run the full developer setup
./scripts/setup/setup-dev.sh
```

**What the script installs and configures:**

| Tool / Action                      | Details                                           |
| ---------------------------------- | ------------------------------------------------- |
| buf CLI                            | Protobuf toolchain (v1.32+)                       |
| protoc-gen-go & protoc-gen-go-grpc | Go Protobuf + gRPC code generators                |
| pnpm                               | Frontend package manager (v9+)                    |
| gitleaks                           | Secret scanning                                   |
| gosec                              | Go SAST                                           |
| govulncheck                        | Go vulnerability scanner                          |
| golangci-lint                      | Go linters                                        |
| trivy                              | Container image scanner                           |
| semgrep                            | Multi-language SAST                               |
| mkcert                             | Local TLS certificate authority                   |
| kubectl                            | Kubernetes CLI                                    |
| helm                               | Kubernetes chart deployment                       |
| Git hooks                          | Pre-commit and commit-msg hooks from `.githooks/` |
| Go modules                         | `go mod download` for all services                |
| Frontend deps                      | `pnpm install` in `ui/`                           |
| Protobuf codegen                   | `buf generate proto/`                             |
| `.env` file                        | Created from `.env.example`                       |
| Dev TLS certs                      | Generated via `gen-dev-certs.sh`                  |
| Docker images                      | Pre-pull dev stack images                         |

**Expected output** (truncated):

```
════════════════════════════════════════════════════
  RTSA Developer Environment Setup
  CLASSIFICATION: UNCLASSIFIED
════════════════════════════════════════════════════
[ℹ] Detected platform: linux/amd64
[✓] Go 1.22.4 (>= 1.22 required)
[✓] buf 1.32.2 installed
[✓] protoc-gen-go and protoc-gen-go-grpc installed
[✓] Node.js v20.x.x (>= 20 required)
[✓] pnpm 9.x.x installed
[✓] Docker daemon running
[✓] Docker Compose v2.x.x
[✓] gitleaks 8.18.4 installed
[✓] gosec installed
[✓] govulncheck installed
[✓] golangci-lint 1.57.2 installed
[✓] trivy 0.50.4 installed
[✓] semgrep installed via pip3
[✓] mkcert installed and local CA configured
[✓] kubectl 1.29.4 installed
[✓] helm installed
[✓] Git hooks configured from .githooks/
... (module downloads) ...
[✓] Protobuf code generated
[✓] .env created from .env.example
[✓] Dev TLS certificates generated
[✓] Docker images pulled

════════════════════════════════════════════════════
  Setup complete! No errors detected.
════════════════════════════════════════════════════
```

If the script reports errors, see the [Troubleshooting](#troubleshooting) section.

**Go not installed yet?**

If Go was not installed before running the setup script, install it now and re-run:

```bash
GO_VERSION="1.22.4"
curl -LO "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc
go version
# Then re-run:
./scripts/setup/setup-dev.sh
```

**Node.js not installed yet?**

```bash
# Install nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc

# Install Node.js 20 LTS
nvm install 20
nvm use 20
nvm alias default 20
node --version
# Then re-run:
./scripts/setup/setup-dev.sh
```

---

## Phase 3 — Manual Finish

### Step 8 — Review Environment Variables

The setup script created a `.env` file from `.env.example`. Review it and update any values specific to your environment:

```bash
# Open in VS Code (from WSL2)
code .env

# Or use nano
nano .env
```

Key variables to check:

```dotenv
# ClickHouse password — change from the default for anything beyond throwaway dev
CLICKHOUSE_PASSWORD=dev_password_change_me

# Log format — use 'json' for production-like output or 'text' for readability
LOG_FORMAT=text

# TLS cert paths — these should point to certs generated by the setup script
TLS_CA_CERT=./certs/dev/ca.crt
TLS_SERVER_CERT=./certs/dev/server.crt
TLS_SERVER_KEY=./certs/dev/server.key
```

> **Never commit `.env`** — it is in `.gitignore`. Use real secrets management (Vault / AWS Secrets Manager) in staging and production.

---

### Step 9 — Start the Development Stack

```bash
docker compose -f deploy/docker-compose.yml up -d
```

Wait 15–30 seconds for all containers to initialise, then check their status:

```bash
docker compose -f deploy/docker-compose.yml ps
```

All containers should show **Status: running** or **healthy**. The full dev stack includes:

| Service            | URL                                            | Purpose                              |
| ------------------ | ---------------------------------------------- | ------------------------------------ |
| Redpanda broker    | `localhost:19092`                              | Kafka-compatible event streaming API |
| Redpanda Console   | [http://localhost:8080](http://localhost:8080) | Broker and topic management UI       |
| Schema Registry    | `localhost:18081`                              | Protobuf schema registry             |
| Redpanda Admin API | `localhost:19644`                              | Cluster management API               |
| ClickHouse HTTP    | [http://localhost:8123](http://localhost:8123) | HTTP query interface                 |
| ClickHouse native  | `localhost:9000`                               | Native Go client port                |
| Prometheus         | [http://localhost:9090](http://localhost:9090) | Metrics scraping and querying        |
| Grafana            | [http://localhost:3000](http://localhost:3000) | Dashboards (login: admin / admin)    |
| Loki               | `localhost:3100`                               | Log aggregation backend              |
| Tempo (OTLP)       | `localhost:4317`                               | Distributed tracing                  |
| Envoy API Gateway  | `localhost:8443`                               | gRPC-Web mTLS endpoint               |

> These URLs are accessible from both your **Windows browser** and from **within WSL2**. Docker Desktop bridges the network automatically.

---

### Step 10 — Initialize Redpanda Topics

```bash
./scripts/dev/init-topics.sh
```

This creates all 20 required Redpanda topics:

- `sensor.raw.*` (radar, ew_sigint, elint_comint, isr, ais_bft, cyber) + dead-letter queues
- `entity.tracks.fused`
- `inference.anomaly.scores`
- `feedback.operator.submissions` / `feedback.operator.validated`
- `audit.events`
- `nato.exchange.inbound` / `nato.exchange.outbound`

Expected output:

```
[✓] Redpanda is ready
[✓]   sensor.raw.radar — created (3 partitions, retention 86400000ms)
[✓]   sensor.raw.ew_sigint — created
...
[✓] All topics created/verified.
```

---

### Step 11 — Initialize ClickHouse Schema

```bash
./scripts/dev/init-clickhouse.sh
```

This creates the `rtsa` database and all core tables:

- `sensor_events` — raw sensor data from all 6 sensor categories
- `entity_tracks` — fused multi-source entity tracks
- `anomaly_scores` — AI/ML inference results
- `audit_events` — immutable audit trail (ITSG-33 AU-3)
- `feedback_events` — operator feedback with trust scores

Expected output:

```
[✓] ClickHouse is ready
[✓] Database 'rtsa' created/verified
[✓]   sensor_events table ready
[✓]   entity_tracks table ready
[✓]   anomaly_scores table ready
[✓]   audit_events table ready
[✓]   feedback_events table ready
Done. ClickHouse schema is ready for development.
```

---

### Step 12 — Verify the Installation

Run the health check to confirm every component is working:

```bash
./scripts/dev/health-check.sh
```

Expected output:

```
RTSA Development Environment Health Check
CLASSIFICATION: UNCLASSIFIED

── Required Tools ──
[✓]   go 1.22.4
[✓]   buf 1.32.2
[✓]   node 20.x.x
[✓]   pnpm 9.x.x
[✓]   docker 26.x.x
[✓]   gitleaks 8.18.x
[✓]   golangci-lint 1.57.x
[✓]   gosec 2.x.x
[✓]   trivy 0.50.x
[✓]   mkcert 1.4.x

── Docker ──
[✓] Docker daemon running
[✓] All 10/10 dev stack containers running

── Redpanda ──
[✓] Redpanda broker reachable (http://localhost:19644)
[✓] Redpanda cluster is healthy
[✓] Redpanda topics active: 20

── ClickHouse ──
[✓] ClickHouse reachable (http://localhost:8123)
[✓] Database 'rtsa' exists
[✓] ClickHouse tables in 'rtsa': 5

── Observability Stack ──
[✓] Prometheus healthy (localhost:9090)
[✓] Grafana healthy (localhost:3000)
[✓] Loki ready (localhost:3100)

── TLS Certificates ──
[✓]   ca.crt — valid (expires: ...)
[✓]   server.crt — valid (expires: ...)
[✓]   client.crt — valid (expires: ...)

══════════════════════════════════════════════════════
  All checks passed. Development environment is ready.
══════════════════════════════════════════════════════
```

---

### Step 13 — Run Tests

#### Go Unit Tests

```bash
./scripts/dev/test-go.sh
```

Runs all Go unit tests with the race detector and verifies ≥ 80% coverage per service.

#### Frontend Unit Tests

```bash
cd ui
pnpm test --coverage
```

#### Pre-PR Gate (Run before every Pull Request)

```bash
./scripts/dev/pre-pr-check.sh
```

Mirrors all 5 CI security gates locally (secret scan → build → unit tests → SAST → integration). All gates must pass before opening a PR.

---

## Opening VS Code in WSL2

Open VS Code with full WSL2 integration from anywhere inside the Ubuntu terminal:

```bash
cd ~/workspace/rtsa
code .
```

VS Code opens on Windows but executes entirely inside WSL2 — files, extensions, terminals, and the debugger all run in the Linux environment. Install the Go, Protobuf, and ESLint extensions when prompted.

---

## Development Workflow Reference

```bash
# Start a feature branch
git checkout main && git pull --rebase
git checkout -b feature/RTSA-XXX-brief-description

# Start the dev stack (if not already running)
docker compose -f deploy/docker-compose.yml up -d

# Make code changes...
# After modifying .proto files:
buf generate proto/

# Run tests for a single service
cd services/radar-ingestion
go test -race -cover ./...

# Run the full local gate before PR
./scripts/dev/pre-pr-check.sh

# Push and open PR
git push origin feature/RTSA-XXX-brief-description

# Stop the dev stack when done
docker compose -f deploy/docker-compose.yml down
```

### Makefile Targets (Per Service)

```bash
cd services/<service-name>

make run      # Run the service process locally
make test     # Unit tests
make lint     # golangci-lint
make proto    # Regenerate protobuf code
make build    # Build binary
make docker   # Build Docker image
```

---

## Troubleshooting

### WSL2 — "A required feature is not installed"

The Virtual Machine Platform feature is not enabled. Re-run Step 2 as Administrator, or enable **Hyper-V** / **Virtual Machine Platform** in **Windows Features** (Control Panel → Programs → Turn Windows features on or off).

### WSL2 — Clock drift after Windows sleep/resume

Services may fail with expired TLS certificates or time-skewed Kafka offsets after the machine sleeps. Re-sync the WSL2 clock:

```bash
sudo hwclock -s
# or
sudo ntpdate pool.ntp.org
```

If the issue persists, restart WSL2:

```powershell
wsl --shutdown
wsl -d Ubuntu-24.04
```

### Docker Desktop — "Docker daemon not running" in WSL2

1. Ensure Docker Desktop is running (whale icon in system tray).
2. Verify WSL2 integration is enabled: Docker Desktop → Settings → Resources → WSL Integration → Ubuntu-24.04 enabled.
3. Restart Docker Desktop.
4. If still failing: `wsl --shutdown`, then restart Docker Desktop.

### Docker Compose — Port already in use

Find the conflicting process and stop it, or change the port mapping in `deploy/docker-compose.yml`:

```bash
# Find what is using port 9000 (ClickHouse native)
sudo lsof -i :9000
# or on Windows PowerShell:
# netstat -ano | findstr :9000
```

Default ports: Redpanda 19092 / 18081 / 19644 / 8080, ClickHouse 9000 / 8123, Grafana 3000, Prometheus 9090, Envoy 8443.

### `/mnt/c/...` path is very slow

This is a known WSL2 limitation — accessing the Windows filesystem from Linux is slow. Always work inside the WSL2 filesystem (`~/workspace/rtsa`), not under `/mnt/c/`. Re-clone the repo if needed:

```bash
mkdir -p ~/workspace && cd ~/workspace
git clone https://github.com/<org>/rtsa.git
```

### mkcert — "rootCA.pem not found"

Run `mkcert -install` to initialise the local CA, then regenerate certificates:

```bash
mkcert -install
./scripts/setup/gen-dev-certs.sh
```

### go: command not found after install

The Go `bin` directory is not on `PATH`. Add it:

```bash
echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc
go version
```

### buf generate fails — "plugin not found"

The Go plugin binaries are not on `PATH`. Ensure `$(go env GOPATH)/bin` is on `PATH`:

```bash
echo 'export PATH="$PATH:'"$(go env GOPATH)/bin"'"' >> ~/.bashrc
source ~/.bashrc

# Reinstall plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
buf generate proto/
```

---

## Additional Resources

| Resource                           | Link                                                                                                                     |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Main Getting Started (Linux/macOS) | [GETTING_STARTED.md](GETTING_STARTED.md)                                                                                 |
| README                             | [README.md](README.md)                                                                                                   |
| Master Policy                      | [docs/sdlc_guidelines/00_master_policy.md](docs/sdlc_guidelines/00_master_policy.md)                                     |
| Go Standards                       | [docs/sdlc_guidelines/04_coding_standards/go_standards.md](docs/sdlc_guidelines/04_coding_standards/go_standards.md)     |
| High-Level Architecture            | [docs/architecture/high_level_architecture.md](docs/architecture/high_level_architecture.md)                             |
| CI/CD Pipeline                     | [docs/sdlc_guidelines/06_integration_cicd/ci_cd_pipeline.md](docs/sdlc_guidelines/06_integration_cicd/ci_cd_pipeline.md) |

---

> **CLASSIFICATION: UNCLASSIFIED** — This document contains no classified information.
