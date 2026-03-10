#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# RTSA Development TLS Certificate Generator
# Generates a local CA and server/client certificates for mTLS dev environments.
# Run from the repository root: ./scripts/setup/gen-dev-certs.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CERT_DIR="${REPO_ROOT}/certs/dev"

GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[ℹ]${NC} $*"; }
log_pass() { echo -e "${GREEN}[✓]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[!]${NC} $*"; }

# ─────────────────────────────────────────────────────────────
# Validate prerequisites
# ─────────────────────────────────────────────────────────────
check_prereqs() {
  if ! command -v openssl &>/dev/null; then
    echo "ERROR: openssl is not installed. Install it via your package manager."
    exit 1
  fi
}

# ─────────────────────────────────────────────────────────────
# Skip if certs are already valid
# ─────────────────────────────────────────────────────────────
check_existing_certs() {
  local server_cert="${CERT_DIR}/server.crt"

  if [ -f "$server_cert" ]; then
    # Check if cert is still valid for more than 30 days
    if openssl x509 -checkend $((30 * 86400)) -noout -in "$server_cert" 2>/dev/null; then
      log_pass "Existing dev certificates are valid (more than 30 days remaining)"
      log_info "Certificate location: ${CERT_DIR}/"
      log_info "To force regeneration, delete ${CERT_DIR}/ and re-run this script"
      exit 0
    else
      log_warn "Existing certificates are expiring soon — regenerating"
    fi
  fi
}

# ─────────────────────────────────────────────────────────────
# Generate certificates
# Uses plain openssl (no mkcert) to produce PKCS8-format ECDSA P-256 keys.
# PKCS8 format (BEGIN PRIVATE KEY) is required for Envoy/BoringSSL.
# Permissions are set to 644 so Docker containers running as non-root can read
# the key files via bind mounts (these are dev-only, never committed).
# ─────────────────────────────────────────────────────────────
generate_certs() {
  mkdir -p "${CERT_DIR}"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "${tmp}"' EXIT
  log_info "Generating development TLS certificates in ${CERT_DIR}/"

  # ── CA ───────────────────────────────────────────────────
  openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 \
    -out "${CERT_DIR}/ca.key" 2>/dev/null
  openssl req -new -x509 -key "${CERT_DIR}/ca.key" -sha256 -days 3650 \
    -out "${CERT_DIR}/ca.crt" \
    -subj "/O=RTSA Dev CA/CN=rtsa-dev-ca" 2>/dev/null
  log_pass "CA certificate: ${CERT_DIR}/ca.crt"

  # ── Server cert (SANs: localhost, *.rtsa.local, 127.0.0.1) ─
  openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 \
    -out "${CERT_DIR}/server.key" 2>/dev/null
  openssl req -new -key "${CERT_DIR}/server.key" -out "${tmp}/server.csr" \
    -subj "/O=RTSA Dev/CN=localhost" 2>/dev/null
  cat > "${tmp}/server-ext.cnf" <<'EXT'
[SAN]
subjectAltName=DNS:localhost,DNS:*.rtsa.local,DNS:rtsa.local,IP:127.0.0.1
EXT
  openssl x509 -req -in "${tmp}/server.csr" \
    -CA "${CERT_DIR}/ca.crt" -CAkey "${CERT_DIR}/ca.key" -CAcreateserial \
    -out "${CERT_DIR}/server.crt" -days 3650 -sha256 \
    -extfile "${tmp}/server-ext.cnf" -extensions SAN 2>/dev/null
  log_pass "Server certificate: ${CERT_DIR}/server.crt"

  # ── Client cert (mTLS) ──────────────────────────────────
  openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 \
    -out "${CERT_DIR}/client.key" 2>/dev/null
  openssl req -new -key "${CERT_DIR}/client.key" -out "${tmp}/client.csr" \
    -subj "/O=RTSA Dev Client/CN=rtsa-client" 2>/dev/null
  cat > "${tmp}/client-ext.cnf" <<'EXT'
[SAN]
subjectAltName=DNS:rtsa-client,DNS:localhost
extendedKeyUsage=clientAuth
EXT
  openssl x509 -req -in "${tmp}/client.csr" \
    -CA "${CERT_DIR}/ca.crt" -CAkey "${CERT_DIR}/ca.key" -CAcreateserial \
    -out "${CERT_DIR}/client.crt" -days 3650 -sha256 \
    -extfile "${tmp}/client-ext.cnf" -extensions SAN 2>/dev/null
  log_pass "Client certificate: ${CERT_DIR}/client.crt"

  # 644 on keys: Docker containers run as non-root (e.g. envoy UID 101) and
  # need read access via bind mount. These are dev-only certs, never committed.
  chmod 644 "${CERT_DIR}"/*.key "${CERT_DIR}"/*.crt
  log_pass "Certificate permissions set (644 — readable by Docker non-root users)"
}

# ─────────────────────────────────────────────────────────────
# Print certificate info
# ─────────────────────────────────────────────────────────────
print_cert_info() {
  echo ""
  echo "Generated certificates:"
  echo ""
  echo "  ${CERT_DIR}/ca.crt       — Local dev CA certificate"
  echo "  ${CERT_DIR}/ca.key       — Local dev CA private key (DO NOT COMMIT)"
  echo "  ${CERT_DIR}/server.crt   — Server certificate (SAN: localhost, *.rtsa.local)"
  echo "  ${CERT_DIR}/server.key   — Server private key (DO NOT COMMIT)"
  echo "  ${CERT_DIR}/client.crt   — Client certificate for mTLS"
  echo "  ${CERT_DIR}/client.key   — Client private key (DO NOT COMMIT)"
  echo ""

  if command -v openssl &>/dev/null; then
    local expiry
    expiry="$(openssl x509 -enddate -noout -in "${CERT_DIR}/server.crt" | sed 's/notAfter=//')"
    log_info "Server certificate expires: ${expiry}"
  fi

  echo ""
  log_warn "certs/ is in .gitignore — NEVER commit private keys"
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
main() {
  echo ""
  echo -e "${CYAN}══ RTSA Dev TLS Certificate Generator (UNCLASSIFIED) ══${NC}"

  check_prereqs
  check_existing_certs
  generate_certs
  print_cert_info

  echo -e "${GREEN}Done.${NC} Development TLS certificates are ready."
}

main "$@"
