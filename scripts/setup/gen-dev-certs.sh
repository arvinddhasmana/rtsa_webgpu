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
  if ! command -v mkcert &>/dev/null; then
    echo "ERROR: mkcert is not installed."
    echo "  Install via: go install filippo.io/mkcert@latest"
    echo "  Or on macOS: brew install mkcert"
    echo "  Run 'mkcert -install' after installation to create the local CA."
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
# ─────────────────────────────────────────────────────────────
generate_certs() {
  mkdir -p "${CERT_DIR}"
  log_info "Generating development TLS certificates in ${CERT_DIR}/"

  # Ensure the local CA is installed in the system trust store
  mkcert -install

  # CA certificate location (set by mkcert)
  local ca_root
  ca_root="$(mkcert -CAROOT)"

  # Copy the CA cert into our cert dir for service configuration
  cp "${ca_root}/rootCA.pem" "${CERT_DIR}/ca.crt"
  cp "${ca_root}/rootCA-key.pem" "${CERT_DIR}/ca.key"
  log_pass "Local CA copied to ${CERT_DIR}/ca.crt"

  # Generate server certificate
  # SANs cover localhost and *.rtsa.local for all services
  mkcert \
    -cert-file "${CERT_DIR}/server.crt" \
    -key-file  "${CERT_DIR}/server.key" \
    "localhost" \
    "127.0.0.1" \
    "*.rtsa.local" \
    "rtsa.local"
  log_pass "Server certificate: ${CERT_DIR}/server.crt"

  # Generate client certificate for mTLS client authentication
  mkcert \
    -client \
    -cert-file "${CERT_DIR}/client.crt" \
    -key-file  "${CERT_DIR}/client.key" \
    "rtsa-client" \
    "localhost"
  log_pass "Client certificate: ${CERT_DIR}/client.crt"

  # Set restrictive permissions on private keys
  chmod 600 "${CERT_DIR}"/*.key
  chmod 644 "${CERT_DIR}"/*.crt
  log_pass "Certificate permissions set (keys: 600, certs: 644)"
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
