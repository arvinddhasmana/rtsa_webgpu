// CLASSIFICATION: UNCLASSIFIED
package main

import (
"crypto/ecdsa"
"crypto/elliptic"
"crypto/rand"
"crypto/x509"
"crypto/x509/pkix"
"encoding/pem"
"math/big"
"os"
"path/filepath"
"testing"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/config"
)

// generateTestCerts creates a self-signed CA, server cert, and key in a temp dir.
// Returns paths to (caFile, certFile, keyFile) and a cleanup function.
func generateTestCerts(t *testing.T) (caFile, certFile, keyFile string, cleanup func()) {
t.Helper()
dir := t.TempDir()

// Generate CA key
caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
if err != nil {
t.Fatalf("generate CA key: %v", err)
}

caTemplate := &x509.Certificate{
SerialNumber:          big.NewInt(1),
Subject:               pkix.Name{CommonName: "test-ca"},
NotBefore:             time.Now().Add(-time.Hour),
NotAfter:              time.Now().Add(time.Hour),
IsCA:                  true,
KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
BasicConstraintsValid: true,
}

caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
if err != nil {
t.Fatalf("create CA cert: %v", err)
}

caFile = filepath.Join(dir, "ca.crt")
if err := writePEM(caFile, "CERTIFICATE", caDER); err != nil {
t.Fatalf("write CA cert: %v", err)
}

// Parse CA cert for signing
caCert, err := x509.ParseCertificate(caDER)
if err != nil {
t.Fatalf("parse CA cert: %v", err)
}

// Generate server key
serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
if err != nil {
t.Fatalf("generate server key: %v", err)
}

serverTemplate := &x509.Certificate{
SerialNumber: big.NewInt(2),
Subject:      pkix.Name{CommonName: "localhost"},
NotBefore:    time.Now().Add(-time.Hour),
NotAfter:     time.Now().Add(time.Hour),
KeyUsage:     x509.KeyUsageDigitalSignature,
ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
}

serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
if err != nil {
t.Fatalf("create server cert: %v", err)
}

certFile = filepath.Join(dir, "server.crt")
if err := writePEM(certFile, "CERTIFICATE", serverDER); err != nil {
t.Fatalf("write server cert: %v", err)
}

keyFile = filepath.Join(dir, "server.key")
keyDER, err := x509.MarshalECPrivateKey(serverKey)
if err != nil {
t.Fatalf("marshal server key: %v", err)
}
if err := writePEM(keyFile, "EC PRIVATE KEY", keyDER); err != nil {
t.Fatalf("write server key: %v", err)
}

return caFile, certFile, keyFile, func() {}
}

func writePEM(path, blockType string, data []byte) error {
f, err := os.Create(path)
if err != nil {
return err
}
defer f.Close()
return pem.Encode(f, &pem.Block{Type: blockType, Bytes: data})
}

// TestLoadTLSCredentials_ValidCerts verifies that loadTLSCredentials succeeds
// with well-formed CA, server cert, and server key (T01).
func TestLoadTLSCredentials_ValidCerts(t *testing.T) {
caFile, certFile, keyFile, cleanup := generateTestCerts(t)
defer cleanup()

cfg := &config.Config{
TLSEnabled:    true,
TLSCACert:     caFile,
TLSServerCert: certFile,
TLSServerKey:  keyFile,
}

opt, err := loadTLSCredentials(cfg)
if err != nil {
t.Fatalf("loadTLSCredentials: unexpected error: %v", err)
}
if opt == nil {
t.Fatal("loadTLSCredentials: returned nil ServerOption")
}
}

// TestLoadTLSCredentials_MissingCACert verifies that loadTLSCredentials returns
// an error when the CA cert file does not exist (T02).
func TestLoadTLSCredentials_MissingCACert(t *testing.T) {
_, certFile, keyFile, cleanup := generateTestCerts(t)
defer cleanup()

cfg := &config.Config{
TLSEnabled:    true,
TLSCACert:     "/nonexistent/ca.crt",
TLSServerCert: certFile,
TLSServerKey:  keyFile,
}

_, err := loadTLSCredentials(cfg)
if err == nil {
t.Fatal("expected error for missing CA cert, got nil")
}
}

// TestLoadTLSCredentials_MissingServerCert verifies that loadTLSCredentials
// returns an error when the server cert or key file does not exist (T02).
func TestLoadTLSCredentials_MissingServerCert(t *testing.T) {
caFile, _, _, cleanup := generateTestCerts(t)
defer cleanup()

cfg := &config.Config{
TLSEnabled:    true,
TLSCACert:     caFile,
TLSServerCert: "/nonexistent/server.crt",
TLSServerKey:  "/nonexistent/server.key",
}

_, err := loadTLSCredentials(cfg)
if err == nil {
t.Fatal("expected error for missing server cert, got nil")
}
}

// TestLoadTLSCredentials_InvalidCAPEM verifies that loadTLSCredentials returns
// an error when the CA cert file contains invalid PEM data (T02).
func TestLoadTLSCredentials_InvalidCAPEM(t *testing.T) {
dir := t.TempDir()
_, certFile, keyFile, cleanup := generateTestCerts(t)
defer cleanup()

badCA := filepath.Join(dir, "bad-ca.crt")
if err := os.WriteFile(badCA, []byte("not valid PEM"), 0600); err != nil {
t.Fatalf("write bad CA: %v", err)
}

cfg := &config.Config{
TLSEnabled:    true,
TLSCACert:     badCA,
TLSServerCert: certFile,
TLSServerKey:  keyFile,
}

_, err := loadTLSCredentials(cfg)
if err == nil {
t.Fatal("expected error for invalid CA PEM, got nil")
}
}
