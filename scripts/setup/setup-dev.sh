#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Developer Environment Setup Script
# Installs and validates all required tools for RTSA development.
# Run from the repository root: ./scripts/setup/setup-dev.sh

set -euo pipefail

# ─────────────────────────────────────────────────────────────
# Constants
# ─────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

BUF_VERSION="1.32.2"
GOLANGCI_LINT_VERSION="1.57.2"
GITLEAKS_VERSION="8.18.4"
TRIVY_VERSION="0.50.4"
HELM_VERSION="3.14.4"
K3D_VERSION="5.6.3"
KUBECTL_VERSION="1.29.4"
GO_MIN_VERSION="1.22"
NODE_MIN_VERSION="20"
WASM_PACK_VERSION="0.13.1"

# ─────────────────────────────────────────────────────────────
# WSL detection & sudo-user HOME fix
# ─────────────────────────────────────────────────────────────
IS_WSL=false
if grep -qi microsoft /proc/version 2>/dev/null; then
  IS_WSL=true
fi

# When invoked via "sudo bash ...", HOME becomes /root which breaks
# paths like $HOME/go/bin for the actual developer.  Restore it.
if [ -n "${SUDO_USER:-}" ]; then
  export HOME="$(getent passwd "$SUDO_USER" | cut -d: -f6)"
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Colour

PASS="${GREEN}[✓]${NC}"
FAIL="${RED}[✗]${NC}"
INFO="${CYAN}[ℹ]${NC}"
WARN="${YELLOW}[!]${NC}"

ERRORS=0

# ─────────────────────────────────────────────────────────────
# Helper functions
# ─────────────────────────────────────────────────────────────

log_info()  { echo -e "${INFO} $*"; }
log_pass()  { echo -e "${PASS} $*"; }
log_fail()  { echo -e "${FAIL} $*"; ERRORS=$(( ERRORS + 1 )); }
log_warn()  { echo -e "${WARN} $*"; }
log_step()  { echo -e "\n${CYAN}══ $* ══${NC}"; }

# Checks if a command is available
has_cmd() { command -v "$1" &>/dev/null; }

# Compares semantic versions: returns 0 if $1 >= $2
version_gte() {
  printf '%s\n%s' "$2" "$1" | sort -C -V
}

# Downloads a file to a temp location then installs it (using sudo if the
# destination directory is not writable by the current user).
install_binary() {
  local url="$1"
  local dest="$2"
  local mode="${3:-755}"
  log_info "Downloading $(basename "$dest") from ${url} ..."
  local tmp_file
  tmp_file="$(mktemp)"
  curl -sSfL "$url" -o "$tmp_file"
  chmod "$mode" "$tmp_file"
  if [ -w "$(dirname "$dest")" ]; then
    mv "$tmp_file" "$dest"
  else
    sudo mv "$tmp_file" "$dest"
  fi
}

# ─────────────────────────────────────────────────────────────
# Detect OS / architecture
# ─────────────────────────────────────────────────────────────
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux)  PLATFORM="linux" ;;
    Darwin) PLATFORM="darwin" ;;
    *)      log_fail "Unsupported OS: $OS. Use Linux, macOS, or WSL2."; exit 1 ;;
  esac

  case "$ARCH" in
    x86_64)  GOARCH="amd64" ;;
    aarch64|arm64) GOARCH="arm64" ;;
    *)       log_fail "Unsupported architecture: $ARCH"; exit 1 ;;
  esac

  log_info "Detected platform: ${PLATFORM}/${GOARCH}"
}

# ─────────────────────────────────────────────────────────────
# Step 1: Go toolchain — auto-install on Linux/WSL2 if missing
# ─────────────────────────────────────────────────────────────
GO_INSTALL_VERSION="1.22.4"

check_go() {
  log_step "Checking / Installing Go toolchain"

  if ! has_cmd go && [ ! -x /usr/local/go/bin/go ]; then
    log_warn "Go not found — auto-installing Go ${GO_INSTALL_VERSION} ..."

    local go_archive="go${GO_INSTALL_VERSION}.${PLATFORM}-${GOARCH}.tar.gz"
    local tmp_dir
    tmp_dir="$(mktemp -d)"

    curl -sSfL "https://go.dev/dl/${go_archive}" -o "${tmp_dir}/${go_archive}"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "${tmp_dir}/${go_archive}"
    rm -rf "$tmp_dir"

    # Expose Go for the rest of this script session
    export PATH="$PATH:/usr/local/go/bin:${HOME}/go/bin"

    # Persist to the invoking user's shell config
    local target_rc="${HOME}/.bashrc"
    if [ -n "${ZSH_VERSION:-}" ]; then
      target_rc="${HOME}/.zshrc"
    fi
    if ! grep -q '/usr/local/go/bin' "$target_rc" 2>/dev/null; then
      echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> "$target_rc"
      log_info "Go PATH entry added to ${target_rc}. Run: source ${target_rc}"
    fi
    log_pass "Go ${GO_INSTALL_VERSION} installed to /usr/local/go"

  elif [ -x /usr/local/go/bin/go ] && ! has_cmd go; then
    # Binary present but PATH not yet updated for this session
    export PATH="$PATH:/usr/local/go/bin:${HOME}/go/bin"
  fi

  if has_cmd go; then
    local installed_go
    installed_go="$(go version | awk '{print $3}' | sed 's/go//')"
    if version_gte "$installed_go" "$GO_MIN_VERSION"; then
      log_pass "Go ${installed_go} (>= ${GO_MIN_VERSION} required)"
    else
      log_fail "Go ${installed_go} is too old. Minimum required: ${GO_MIN_VERSION}"
    fi
  else
    log_fail "Go installation failed. Install manually: https://go.dev/dl/"
  fi
}

# ─────────────────────────────────────────────────────────────
# Step 2: buf CLI (Protobuf toolchain)
# ─────────────────────────────────────────────────────────────
install_buf() {
  log_step "Installing buf CLI (Protobuf toolchain)"

  if has_cmd buf; then
    INSTALLED_BUF="$(buf --version 2>/dev/null | head -1 | sed 's/[^0-9.]//g')"
    if version_gte "$INSTALLED_BUF" "$BUF_VERSION"; then
      log_pass "buf ${INSTALLED_BUF} already installed"
      return
    fi
    log_warn "buf ${INSTALLED_BUF} found but ${BUF_VERSION}+ required. Upgrading..."
  fi

  local buf_os
  case "$PLATFORM" in
    linux)  buf_os="Linux" ;;
    darwin) buf_os="Darwin" ;;
  esac

  local buf_arch
  case "$GOARCH" in
    amd64) buf_arch="x86_64" ;;
    arm64) buf_arch="arm64" ;;
  esac

  install_binary \
    "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-${buf_os}-${buf_arch}" \
    "/usr/local/bin/buf"

  log_pass "buf ${BUF_VERSION} installed"
}

install_proto_plugins() {
  log_step "Installing Go Protobuf plugins"

  if ! has_cmd go; then
    log_warn "Go not found — skipping Protobuf plugin installation"
    return
  fi

  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

  log_pass "protoc-gen-go and protoc-gen-go-grpc installed to \$(go env GOPATH)/bin"
}

# ─────────────────────────────────────────────────────────────
# Step 3: Node.js / pnpm
# ─────────────────────────────────────────────────────────────
check_node() {
  log_step "Checking Node.js"

  if ! has_cmd node; then
    log_fail "Node.js is not installed. Install Node.js ${NODE_MIN_VERSION} LTS via nvm:"
    echo "  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash"
    echo "  nvm install 20 && nvm use 20"
    return
  fi

  NODE_VERSION="$(node --version | sed 's/v//' | cut -d. -f1)"
  if [ "$NODE_VERSION" -ge "$NODE_MIN_VERSION" ]; then
    log_pass "Node.js v$(node --version | sed 's/v//') (>= ${NODE_MIN_VERSION} required)"
  else
    log_fail "Node.js ${NODE_VERSION} is too old. Minimum required: ${NODE_MIN_VERSION} LTS"
  fi
}

install_pnpm() {
  log_step "Installing pnpm"

  if has_cmd pnpm; then
    log_pass "pnpm $(pnpm --version) already installed"
    return
  fi

  npm install -g pnpm@9
  log_pass "pnpm $(pnpm --version) installed"
}

install_wasm_pack_binary() {
  local wasm_platform wasm_arch tarball url tmp_dir extracted_dir bin_dest

  case "$PLATFORM" in
    linux) wasm_platform="unknown-linux-musl" ;;
    darwin) wasm_platform="apple-darwin" ;;
    *)
      log_warn "Automatic wasm-pack install not supported on ${PLATFORM}/${GOARCH}."
      return 1
      ;;
  esac

  case "$GOARCH" in
    amd64) wasm_arch="x86_64" ;;
    arm64) wasm_arch="aarch64" ;;
    *)
      log_warn "Automatic wasm-pack install not supported on ${PLATFORM}/${GOARCH}."
      return 1
      ;;
  esac

  tarball="wasm-pack-v${WASM_PACK_VERSION}-${wasm_arch}-${wasm_platform}.tar.gz"
  url="https://github.com/rustwasm/wasm-pack/releases/download/v${WASM_PACK_VERSION}/${tarball}"
  tmp_dir="$(mktemp -d)"

  log_info "Downloading wasm-pack v${WASM_PACK_VERSION}..."
  if ! curl -sSfL "$url" -o "${tmp_dir}/${tarball}"; then
    log_warn "Failed to download wasm-pack from ${url}"
    rm -rf "$tmp_dir"
    return 1
  fi

  if ! tar -xzf "${tmp_dir}/${tarball}" -C "$tmp_dir"; then
    log_warn "Failed to extract wasm-pack archive"
    rm -rf "$tmp_dir"
    return 1
  fi

  extracted_dir="${tmp_dir}/wasm-pack-v${WASM_PACK_VERSION}-${wasm_arch}-${wasm_platform}"
  bin_dest="${extracted_dir}/wasm-pack"
  if [ ! -x "$bin_dest" ]; then
    log_warn "Unexpected wasm-pack layout in ${tarball}"
    rm -rf "$tmp_dir"
    return 1
  fi

  if [ -w /usr/local/bin ]; then
    mv "$bin_dest" /usr/local/bin/wasm-pack
  else
    sudo mv "$bin_dest" /usr/local/bin/wasm-pack
  fi
  chmod 755 /usr/local/bin/wasm-pack
  rm -rf "$tmp_dir"
  log_pass "wasm-pack v${WASM_PACK_VERSION} installed"
}

# ─────────────────────────────────────────────────────────────
# Step 3b: Rust toolchain + wasm-pack (for web-cop-gpu wasm-decoder)
# ─────────────────────────────────────────────────────────────

# Install rustup and the stable Rust toolchain non-interactively.
# After installation, sources the cargo environment so subsequent
# commands (rustc, cargo, rustup) resolve from ~/.cargo/bin.
install_rustup() {
  if has_cmd rustup; then
    log_pass "rustup already installed"
    return 0
  fi

  log_info "Installing rustup (official Rust toolchain manager)..."
  if ! curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path; then
    log_warn "rustup installation failed"
    return 1
  fi

  # Make rustup/cargo available for the remainder of this script session.
  # shellcheck source=/dev/null
  if [ -f "${HOME}/.cargo/env" ]; then
    . "${HOME}/.cargo/env"
  fi

  log_pass "rustup $(rustup --version 2>/dev/null | awk '{print $2}') installed"
}

# Add wasm32-unknown-unknown to the active Rust toolchain when missing.
# wasm-pack requires this target even on non-Rustup setups; using rustup
# is the supported path.
ensure_wasm_target() {
  if has_cmd rustup; then
    if rustup target list --installed 2>/dev/null | grep -q "wasm32-unknown-unknown"; then
      log_pass "wasm32-unknown-unknown target already installed"
      return 0
    fi
    log_info "Adding wasm32-unknown-unknown target..."
    if rustup target add wasm32-unknown-unknown; then
      log_pass "wasm32-unknown-unknown target added"
      return 0
    fi
    log_warn "Failed to add wasm32-unknown-unknown via rustup"
    return 1
  fi

  # rustup not available — check if the target sysroot exists for a
  # manually-installed Rust (e.g. installed via apt).
  local sysroot
  sysroot="$(rustc --print sysroot 2>/dev/null || true)"
  if [ -n "$sysroot" ] && [ -d "${sysroot}/lib/rustlib/wasm32-unknown-unknown" ]; then
    log_pass "wasm32-unknown-unknown target present in sysroot"
    return 0
  fi

  log_warn "rustup is not installed — cannot add wasm32-unknown-unknown automatically."
  log_info "  Install rustup first: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
  log_info "  Then run: rustup target add wasm32-unknown-unknown"
  return 1
}

check_rust() {
  log_step "Checking Rust toolchain (required for wasm-decoder)"

  if ! has_cmd rustc; then
    log_warn "Rust is not installed. Installing via rustup..."
    install_rustup || {
      log_info "  Install manually: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
      return
    }
  fi

  log_pass "Rust $(rustc --version | awk '{print $2}')"

  if has_cmd wasm-pack; then
    log_pass "wasm-pack $(wasm-pack --version | awk '{print $2}') installed"
  else
    log_warn "wasm-pack not found. Attempting to install prebuilt v${WASM_PACK_VERSION}..."
    if install_wasm_pack_binary; then
      log_pass "wasm-pack $(wasm-pack --version | awk '{print $2}') installed"
    else
      log_warn "Automatic wasm-pack install failed — install manually via cargo install wasm-pack or download a release from https://github.com/rustwasm/wasm-pack/releases"
    fi
  fi

  ensure_wasm_target || log_warn "wasm32-unknown-unknown target missing; install via rustup target add wasm32-unknown-unknown"
}

# ─────────────────────────────────────────────────────────────
# Step 4: Docker
# ─────────────────────────────────────────────────────────────
check_docker() {
  log_step "Checking Docker"

  if ! has_cmd docker; then
    log_fail "Docker is not installed."
    log_info "Install Docker Desktop from: https://www.docker.com/products/docker-desktop/"
    log_info "For Linux: https://docs.docker.com/engine/install/"
    return
  fi

  log_pass "Docker $(docker --version | awk '{print $3}' | tr -d ',')"

  if ! docker info &>/dev/null; then
    log_fail "Docker daemon is not running. Start Docker Desktop or 'sudo systemctl start docker'"
    return
  fi

  if ! docker compose version &>/dev/null 2>&1; then
    log_fail "Docker Compose v2 plugin not found. Update Docker Desktop or install the plugin."
  else
    log_pass "Docker Compose $(docker compose version 2>/dev/null | awk '{print $4}')"
  fi
}

# ─────────────────────────────────────────────────────────────
# Step 5: Security tools
# ─────────────────────────────────────────────────────────────
install_gitleaks() {
  log_step "Installing gitleaks"

  if has_cmd gitleaks; then
    log_pass "gitleaks $(gitleaks version 2>/dev/null | head -1) already installed"
    return
  fi

  local gl_os gl_arch
  case "$PLATFORM" in
    linux)  gl_os="linux" ;;
    darwin) gl_os="darwin" ;;
  esac
  case "$GOARCH" in
    amd64) gl_arch="x64" ;;
    arm64) gl_arch="arm64" ;;
  esac

  local tmp_dir
  tmp_dir="$(mktemp -d)"
  curl -sSfL \
    "https://github.com/zricethezav/gitleaks/releases/download/v${GITLEAKS_VERSION}/gitleaks_${GITLEAKS_VERSION}_${gl_os}_${gl_arch}.tar.gz" \
    | tar -xz -C "$tmp_dir" gitleaks
  sudo mv "$tmp_dir/gitleaks" /usr/local/bin/gitleaks
  rm -rf "$tmp_dir"

  log_pass "gitleaks ${GITLEAKS_VERSION} installed"
}

install_gosec() {
  log_step "Installing gosec"

  if has_cmd gosec; then
    log_pass "gosec $(gosec --version 2>/dev/null | head -1) already installed"
    return
  fi

  if ! has_cmd go; then
    log_warn "Go not found — skipping gosec installation"
    return
  fi

  go install github.com/securego/gosec/v2/cmd/gosec@latest
  log_pass "gosec installed"
}

install_govulncheck() {
  log_step "Installing govulncheck"

  if has_cmd govulncheck; then
    log_pass "govulncheck already installed"
    return
  fi

  if ! has_cmd go; then
    log_warn "Go not found — skipping govulncheck installation"
    return
  fi

  go install golang.org/x/vuln/cmd/govulncheck@latest
  log_pass "govulncheck installed"
}

install_golangci_lint() {
  log_step "Installing golangci-lint"

  if has_cmd golangci-lint; then
    log_pass "golangci-lint $(golangci-lint --version | awk '{print $4}') already installed"
    return
  fi

  if ! has_cmd go; then
    log_warn "Go not found — skipping golangci-lint installation"
    return
  fi

  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s -- -b "$(go env GOPATH)/bin" "v${GOLANGCI_LINT_VERSION}"

  log_pass "golangci-lint ${GOLANGCI_LINT_VERSION} installed"
}

install_trivy() {
  log_step "Installing trivy"

  if has_cmd trivy; then
    log_pass "trivy $(trivy --version | head -1) already installed"
    return
  fi

  local trivy_os trivy_arch
  case "$PLATFORM" in
    linux)  trivy_os="Linux" ;;
    darwin) trivy_os="macOS" ;;
  esac
  case "$GOARCH" in
    amd64) trivy_arch="64bit" ;;
    arm64) trivy_arch="ARM64" ;;
  esac

  local tmp_dir
  tmp_dir="$(mktemp -d)"
  curl -sSfL \
    "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_${trivy_os}-${trivy_arch}.tar.gz" \
    | tar -xz -C "$tmp_dir" trivy
  sudo mv "$tmp_dir/trivy" /usr/local/bin/trivy
  rm -rf "$tmp_dir"

  log_pass "trivy ${TRIVY_VERSION} installed"
}

install_semgrep() {
  log_step "Installing semgrep"

  if has_cmd semgrep; then
    log_pass "semgrep $(semgrep --version 2>/dev/null) already installed"
    return
  fi

  if has_cmd pip3; then
    pip3 install --quiet semgrep
    log_pass "semgrep installed via pip3"
  elif has_cmd pip; then
    pip install --quiet semgrep
    log_pass "semgrep installed via pip"
  else
    log_warn "pip/pip3 not found — install semgrep manually: pip3 install semgrep"
  fi
}

# ─────────────────────────────────────────────────────────────
# Step 6: mkcert (Local CA for dev TLS)
# ─────────────────────────────────────────────────────────────
install_mkcert() {
  log_step "Installing mkcert"

  if has_cmd mkcert; then
    log_pass "mkcert $(mkcert --version 2>/dev/null) already installed"
    return
  fi

  case "$PLATFORM" in
    linux)
      local mkcert_arch
      case "$GOARCH" in
        amd64) mkcert_arch="amd64" ;;
        arm64) mkcert_arch="arm64" ;;
      esac
      sudo curl -sSfL \
        "https://dl.filippo.io/mkcert/latest?for=linux/${mkcert_arch}" \
        -o /usr/local/bin/mkcert
      sudo chmod +x /usr/local/bin/mkcert
      ;;
    darwin)
      brew install mkcert nss
      ;;
  esac

  # Install local CA into system trust store.
  # On WSL, libnss3-tools must be present and the cert won't propagate to
  # the Windows trust store automatically — that step must be done on the host.
  if [ "$IS_WSL" = true ]; then
    if ! dpkg -l libnss3-tools &>/dev/null 2>&1; then
      log_info "Installing libnss3-tools required by mkcert on WSL ..."
      sudo apt-get install -y libnss3-tools &>/dev/null \
        || log_warn "Could not install libnss3-tools — 'mkcert -install' may fail"
    fi
  fi
  mkcert -install 2>/dev/null \
    || log_warn "Could not install mkcert CA — run 'mkcert -install' manually with appropriate permissions"

  log_pass "mkcert installed and local CA configured"
}

# ─────────────────────────────────────────────────────────────
# Step 7: kubectl & helm
# ─────────────────────────────────────────────────────────────
install_kubectl() {
  log_step "Installing kubectl"

  if has_cmd kubectl; then
    log_pass "kubectl $(kubectl version --client 2>/dev/null | head -1) already installed"
    return
  fi

  local kube_os
  case "$PLATFORM" in
    linux)  kube_os="linux" ;;
    darwin) kube_os="darwin" ;;
  esac

  # install_binary handles sudo internally via the temp-file pattern
  install_binary \
    "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/${kube_os}/${GOARCH}/kubectl" \
    "/usr/local/bin/kubectl"

  log_pass "kubectl ${KUBECTL_VERSION} installed"
}

install_helm() {
  log_step "Installing helm"

  if has_cmd helm; then
    log_pass "helm $(helm version --short 2>/dev/null) already installed"
    return
  fi

  curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  log_pass "helm installed"
}

# ─────────────────────────────────────────────────────────────
# Step 8: Git configuration
# ─────────────────────────────────────────────────────────────
configure_git() {
  log_step "Configuring Git"

  if ! git config --get user.name &>/dev/null; then
    log_warn "Git user.name not set. Run: git config --global user.name \"Your Name\""
  else
    log_pass "Git user.name: $(git config --get user.name)"
  fi

  if ! git config --get user.email &>/dev/null; then
    log_warn "Git user.email not set. Run: git config --global user.email \"you@example.com\""
  else
    log_pass "Git user.email: $(git config --get user.email)"
  fi

  # Install git hooks
  if [ -d "${REPO_ROOT}/.git" ]; then
    local hooks_dir="${REPO_ROOT}/.githooks"
    if [ -d "$hooks_dir" ]; then
      git -C "$REPO_ROOT" config core.hooksPath .githooks
      chmod +x "$hooks_dir"/* 2>/dev/null || true
      log_pass "Git hooks configured from .githooks/"
    fi
  fi

  # Set pull rebase by default
  git config --global pull.rebase true 2>/dev/null || true
  log_pass "Git pull.rebase = true"
}

# ─────────────────────────────────────────────────────────────
# Step 9: Go workspace & modules
# ─────────────────────────────────────────────────────────────
setup_go_modules() {
  log_step "Setting up Go modules"

  if ! has_cmd go; then
    log_warn "Go not found — skipping module setup"
    return
  fi

  # Download dependencies for each service
  local services_dir="${REPO_ROOT}/services"
  if [ -d "$services_dir" ]; then
    while IFS= read -r -d '' mod_file; do
      local service_dir
      service_dir="$(dirname "$mod_file")"
      log_info "go mod download: ${service_dir}"
      (cd "$service_dir" && go mod download && go mod tidy) \
        && log_pass "  $(basename "$service_dir") dependencies ready" \
        || log_warn "  $(basename "$service_dir") had module issues — check go.mod"
    done < <(find "$services_dir" -name "go.mod" -print0 2>/dev/null)
  else
    log_info "No services/ directory found yet — skipping Go module setup"
  fi
}

# ─────────────────────────────────────────────────────────────
# Step 10: Frontend dependencies
# ─────────────────────────────────────────────────────────────
setup_frontend() {
  log_step "Setting up frontend dependencies"

  local ui_dir="${REPO_ROOT}/web-cop-gpu"
  if [ ! -d "$ui_dir" ]; then
    log_info "No web-cop-gpu/ directory found yet — skipping frontend setup"
    return
  fi

  if ! has_cmd pnpm; then
    log_warn "pnpm not found — skipping frontend dependency installation"
    return
  fi

  (cd "$ui_dir" && pnpm install --frozen-lockfile 2>/dev/null || pnpm install) \
    && log_pass "Frontend dependencies installed"
}

# ─────────────────────────────────────────────────────────────
# Step 11: Protobuf code generation
# ─────────────────────────────────────────────────────────────
generate_proto() {
  log_step "Generating Protobuf code"

  local proto_dir="${REPO_ROOT}/proto"
  if [ ! -d "$proto_dir" ]; then
    log_info "No proto/ directory found yet — skipping code generation"
    return
  fi

  if ! has_cmd buf; then
    log_warn "buf not found — skipping proto generation"
    return
  fi

  (cd "$REPO_ROOT" && buf generate proto/) \
    && log_pass "Protobuf code generated" \
    || log_fail "Protobuf generation failed — check proto/ for errors"
}

# ─────────────────────────────────────────────────────────────
# Step 12: Create .env file
# ─────────────────────────────────────────────────────────────
setup_env() {
  log_step "Setting up environment variables"

  local env_file="${REPO_ROOT}/.env"
  local example_file="${REPO_ROOT}/.env.example"

  if [ -f "$env_file" ]; then
    log_pass ".env already exists — not overwriting"
    return
  fi

  if [ -f "$example_file" ]; then
    cp "$example_file" "$env_file"
    log_pass ".env created from .env.example — review and update as needed"
  else
    log_warn ".env.example not found — creating minimal .env"
    cat > "$env_file" << 'EOF'
# CLASSIFICATION: UNCLASSIFIED
# RTSA Development Environment Variables
# DO NOT commit this file — it is in .gitignore

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

# TLS
TLS_CERT_DIR=./certs/dev
TLS_CA_CERT=./certs/dev/ca.crt
TLS_SERVER_CERT=./certs/dev/server.crt
TLS_SERVER_KEY=./certs/dev/server.key
TLS_CLIENT_CERT=./certs/dev/client.crt
TLS_CLIENT_KEY=./certs/dev/client.key

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
OTEL_SERVICE_NAME=rtsa-dev

# Development
LOG_LEVEL=debug
LOG_FORMAT=text
DEV_MODE=true
EOF
    log_pass ".env created with defaults"
  fi
}

# ─────────────────────────────────────────────────────────────
# Step 13: Generate dev TLS certificates
# ─────────────────────────────────────────────────────────────
generate_certs() {
  log_step "Generating development TLS certificates"

  local cert_script="${SCRIPT_DIR}/gen-dev-certs.sh"
  if [ -f "$cert_script" ]; then
    bash "$cert_script" && log_pass "Dev TLS certificates generated" \
      || log_warn "Certificate generation failed — run ./scripts/setup/gen-dev-certs.sh manually"
  else
    log_warn "gen-dev-certs.sh not found — skipping certificate generation"
  fi
}

# ─────────────────────────────────────────────────────────────
# Step 14: Pull Docker images
# ─────────────────────────────────────────────────────────────
pull_docker_images() {
  log_step "Pulling Docker images for dev stack"

  if ! has_cmd docker || ! docker info &>/dev/null; then
    log_warn "Docker not running — skipping image pull"
    return
  fi

  local compose_file="${REPO_ROOT}/deploy/docker-compose.yml"
  if [ ! -f "$compose_file" ]; then
    log_info "deploy/docker-compose.yml not found — skipping image pull"
    return
  fi

  docker compose -f "$compose_file" pull --quiet \
    && log_pass "Docker images pulled" \
    || log_warn "Some Docker images failed to pull — retry with: docker compose pull"
}

# ─────────────────────────────────────────────────────────────
# Final summary
# ─────────────────────────────────────────────────────────────
print_summary() {
  echo ""
  echo "════════════════════════════════════════════════════"
  if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}  Setup complete! No errors detected.${NC}"
    echo ""
    echo "  Next steps:"
    echo "    1. Review and update .env"
    echo "    2. docker compose -f deploy/docker-compose.yml up -d"
    echo "    3. ./scripts/dev/init-topics.sh"
    echo "    4. ./scripts/dev/init-clickhouse.sh"
    echo "    5. ./scripts/dev/health-check.sh"
  else
    echo -e "${RED}  Setup completed with ${ERRORS} error(s).${NC}"
    echo "  Review the errors above and fix before proceeding."
    echo "  See GETTING_STARTED.md for detailed instructions."
  fi
  echo "════════════════════════════════════════════════════"
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${CYAN}════════════════════════════════════════════════════${NC}"
  echo -e "${CYAN}  RTSA Developer Environment Setup                  ${NC}"
  echo -e "${CYAN}  CLASSIFICATION: UNCLASSIFIED                      ${NC}"
  echo -e "${CYAN}════════════════════════════════════════════════════${NC}"

  detect_platform

  check_go
  install_buf
  install_proto_plugins
  check_node
  install_pnpm
  check_rust
  check_docker
  install_gitleaks
  install_gosec
  install_govulncheck
  install_golangci_lint
  install_trivy
  install_semgrep
  install_mkcert
  install_kubectl
  install_helm
  configure_git
  setup_go_modules
  setup_frontend
  generate_proto
  setup_env
  generate_certs
  pull_docker_images

  print_summary

  exit "$ERRORS"
}

main "$@"
  echo -e "${CYAN}════════════════════════════════════════════════════${NC}"

  detect_platform

  check_go
  install_buf
  install_proto_plugins
  check_node
  install_pnpm
  check_rust
  check_docker
  install_gitleaks
  install_gosec
  install_govulncheck
  install_golangci_lint
  install_trivy
  install_semgrep
  install_mkcert
  install_kubectl
  install_helm
  configure_git
  setup_go_modules
  setup_frontend
  generate_proto
  setup_env
  generate_certs
  pull_docker_images

  print_summary

  exit "$ERRORS"
}

main "$@"
