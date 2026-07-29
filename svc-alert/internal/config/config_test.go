// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/config"
)

// TestLoad_Defaults verifies that all defaults are correctly applied.
func TestLoad_Defaults(t *testing.T) {
// Clear any relevant env vars
for _, key := range []string{
"RTSA_GRPC_PORT", "RTSA_HEALTH_PORT", "RTSA_METRICS_PORT",
"RTSA_REDPANDA_BROKERS", "RTSA_ALERT_CONSUMER_GROUP",
"RTSA_ALERT_MAX_QUEUE_SIZE", "RTSA_LOG_LEVEL",
} {
os.Unsetenv(key)
}

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if cfg.GRPCPort != 50051 {
t.Errorf("expected GRPCPort=50051, got %d", cfg.GRPCPort)
}
if cfg.HealthPort != 8081 {
t.Errorf("expected HealthPort=8081, got %d", cfg.HealthPort)
}
if cfg.MetricsPort != 9090 {
t.Errorf("expected MetricsPort=9090, got %d", cfg.MetricsPort)
}
if cfg.MaxQueueSize != 10000 {
t.Errorf("expected MaxQueueSize=10000, got %d", cfg.MaxQueueSize)
}
if cfg.LogLevel != "info" {
t.Errorf("expected LogLevel=info, got %s", cfg.LogLevel)
}
if len(cfg.Topics) == 0 {
t.Error("expected default topics to be set")
}
}

// TestLoad_EnvOverride verifies environment variables override defaults.
func TestLoad_EnvOverride(t *testing.T) {
os.Setenv("RTSA_GRPC_PORT", "50052")
os.Setenv("RTSA_HEALTH_PORT", "8082")
os.Setenv("RTSA_METRICS_PORT", "9091")
os.Setenv("RTSA_REDPANDA_BROKERS", "broker1:9092,broker2:9092")
os.Setenv("RTSA_ALERT_CONSUMER_GROUP", "test-group")
os.Setenv("RTSA_ALERT_TOPICS", "alerts.anomaly.critical")
os.Setenv("RTSA_ALERT_MAX_QUEUE_SIZE", "500")
os.Setenv("RTSA_LOG_LEVEL", "debug")
os.Setenv("RTSA_LOG_FORMAT", "text")
os.Setenv("RTSA_SERVICE_NAME", "test-service")
os.Setenv("RTSA_OTEL_ENDPOINT", "otel:4317")
os.Setenv("RTSA_TLS_CA_CERT", "/tmp/ca.crt")
os.Setenv("RTSA_TLS_SERVER_CERT", "/tmp/server.crt")
os.Setenv("RTSA_TLS_SERVER_KEY", "/tmp/server.key")
defer func() {
for _, k := range []string{
"RTSA_GRPC_PORT", "RTSA_HEALTH_PORT", "RTSA_METRICS_PORT",
"RTSA_REDPANDA_BROKERS", "RTSA_ALERT_CONSUMER_GROUP",
"RTSA_ALERT_TOPICS", "RTSA_ALERT_MAX_QUEUE_SIZE",
"RTSA_LOG_LEVEL", "RTSA_LOG_FORMAT", "RTSA_SERVICE_NAME",
"RTSA_OTEL_ENDPOINT", "RTSA_TLS_CA_CERT", "RTSA_TLS_SERVER_CERT", "RTSA_TLS_SERVER_KEY",
} {
os.Unsetenv(k)
}
}()

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}

if cfg.GRPCPort != 50052 {
t.Errorf("expected 50052, got %d", cfg.GRPCPort)
}
if cfg.HealthPort != 8082 {
t.Errorf("expected 8082, got %d", cfg.HealthPort)
}
if cfg.MetricsPort != 9091 {
t.Errorf("expected 9091, got %d", cfg.MetricsPort)
}
if len(cfg.RedpandaBrokers) != 2 {
t.Errorf("expected 2 brokers, got %d", len(cfg.RedpandaBrokers))
}
if cfg.ConsumerGroup != "test-group" {
t.Errorf("expected test-group, got %s", cfg.ConsumerGroup)
}
if len(cfg.Topics) != 1 || cfg.Topics[0] != "alerts.anomaly.critical" {
t.Errorf("expected one topic, got %v", cfg.Topics)
}
if cfg.MaxQueueSize != 500 {
t.Errorf("expected 500, got %d", cfg.MaxQueueSize)
}
if cfg.LogLevel != "debug" {
t.Errorf("expected debug, got %s", cfg.LogLevel)
}
if cfg.TLSCACert != "/tmp/ca.crt" {
t.Errorf("expected /tmp/ca.crt, got %s", cfg.TLSCACert)
}
}

// TestLoad_InvalidGRPCPort verifies error on bad port.
func TestLoad_InvalidGRPCPort(t *testing.T) {
os.Setenv("RTSA_GRPC_PORT", "not-a-number")
defer os.Unsetenv("RTSA_GRPC_PORT")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid RTSA_GRPC_PORT")
}
}

// TestLoad_InvalidHealthPort verifies error on bad port.
func TestLoad_InvalidHealthPort(t *testing.T) {
os.Setenv("RTSA_HEALTH_PORT", "bad")
defer os.Unsetenv("RTSA_HEALTH_PORT")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid RTSA_HEALTH_PORT")
}
}

// TestLoad_InvalidMetricsPort verifies error on bad port.
func TestLoad_InvalidMetricsPort(t *testing.T) {
os.Setenv("RTSA_METRICS_PORT", "bad")
defer os.Unsetenv("RTSA_METRICS_PORT")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid RTSA_METRICS_PORT")
}
}

// TestLoad_InvalidMaxQueueSize verifies error on bad queue size.
func TestLoad_InvalidMaxQueueSize(t *testing.T) {
os.Setenv("RTSA_ALERT_MAX_QUEUE_SIZE", "not-an-int")
defer os.Unsetenv("RTSA_ALERT_MAX_QUEUE_SIZE")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid RTSA_ALERT_MAX_QUEUE_SIZE")
}
}

// TestLoad_Validate_ZeroPort verifies that port 0 fails validation.
func TestLoad_Validate_ZeroPort(t *testing.T) {
os.Setenv("RTSA_GRPC_PORT", "0")
defer os.Unsetenv("RTSA_GRPC_PORT")

_, err := config.Load()
if err == nil {
t.Error("expected validation error for port 0")
}
}
