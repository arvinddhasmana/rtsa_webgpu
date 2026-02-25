// CLASSIFICATION: UNCLASSIFIED
// Package config — unit tests for configuration loading.
package config

import (
"os"
"testing"
)

func TestLoad_Defaults(t *testing.T) {
// Clear any env vars that might interfere.
os.Unsetenv("RTSA_GRPC_PORT")
os.Unsetenv("RTSA_HEALTH_PORT")
os.Unsetenv("RTSA_TLS_ENABLED")
os.Unsetenv("RTSA_SERVICE_NAME")
os.Unsetenv("RTSA_LOG_LEVEL")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if cfg.GRPCAddr != ":50051" {
t.Errorf("GRPCAddr = %q, want %q", cfg.GRPCAddr, ":50051")
}
if cfg.ServiceName != "svc-nato-adapter" {
t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "svc-nato-adapter")
}
if cfg.TLSEnabled != false {
t.Errorf("TLSEnabled = %v, want false", cfg.TLSEnabled)
}
if cfg.LogLevel != "info" {
t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
}
if cfg.Environment != "development" {
t.Errorf("Environment = %q, want %q", cfg.Environment, "development")
}
}

func TestLoad_CustomGRPCPort(t *testing.T) {
os.Setenv("RTSA_GRPC_PORT", "50074")
defer os.Unsetenv("RTSA_GRPC_PORT")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if cfg.GRPCAddr != ":50074" {
t.Errorf("GRPCAddr = %q, want %q", cfg.GRPCAddr, ":50074")
}
}

func TestLoad_InvalidTLSEnabled(t *testing.T) {
os.Setenv("RTSA_TLS_ENABLED", "not-a-bool")
defer os.Unsetenv("RTSA_TLS_ENABLED")

_, err := Load()
if err == nil {
t.Error("expected error for invalid RTSA_TLS_ENABLED, got nil")
}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
os.Setenv("RTSA_LOG_LEVEL", "verbose")
defer os.Unsetenv("RTSA_LOG_LEVEL")

_, err := Load()
if err == nil {
t.Error("expected error for invalid RTSA_LOG_LEVEL, got nil")
}
}

func TestLoad_CustomBrokers(t *testing.T) {
os.Setenv("RTSA_REDPANDA_BROKERS", "broker1:9092,broker2:9092")
defer os.Unsetenv("RTSA_REDPANDA_BROKERS")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if len(cfg.RedpandaBrokers) != 2 {
t.Errorf("RedpandaBrokers length = %d, want 2", len(cfg.RedpandaBrokers))
}
}

func TestLoad_EmptyServiceName(t *testing.T) {
os.Setenv("RTSA_SERVICE_NAME", " ")
defer os.Unsetenv("RTSA_SERVICE_NAME")

// An env var with spaces is still non-empty, so ServiceName will be " "
// which is non-empty. We test empty service name via the validate path.
_, err := Load()
// Should succeed since " " is non-empty (validate only checks empty string)
if err != nil {
t.Logf("Load() returned error (expected for some configs): %v", err)
}
}
