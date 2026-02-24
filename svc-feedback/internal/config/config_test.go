// CLASSIFICATION: UNCLASSIFIED

package config

import (
"os"
"testing"
)

func TestLoad_Defaults(t *testing.T) {
// Clear any env vars that might be set.
unset := []string{
"RTSA_GRPC_PORT", "RTSA_HEALTH_PORT", "RTSA_REDPANDA_BROKERS",
"RTSA_RATE_LIMIT_PER_MINUTE", "RTSA_LOG_LEVEL", "RTSA_SERVICE_NAME",
}
for _, k := range unset {
os.Unsetenv(k)
}

cfg, err := Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.GRPCPort != "50051" {
t.Errorf("expected GRPCPort=50051, got %s", cfg.GRPCPort)
}
if cfg.HealthPort != "8081" {
t.Errorf("expected HealthPort=8081, got %s", cfg.HealthPort)
}
if cfg.RateLimitPerMin != 10 {
t.Errorf("expected RateLimitPerMin=10, got %d", cfg.RateLimitPerMin)
}
if cfg.LogLevel != "info" {
t.Errorf("expected LogLevel=info, got %s", cfg.LogLevel)
}
if cfg.ServiceName != "svc-feedback" {
t.Errorf("expected ServiceName=svc-feedback, got %s", cfg.ServiceName)
}
if len(cfg.RedpandaBrokers) != 1 || cfg.RedpandaBrokers[0] != "localhost:19092" {
t.Errorf("expected default broker, got %v", cfg.RedpandaBrokers)
}
}

func TestLoad_CustomValues(t *testing.T) {
os.Setenv("RTSA_GRPC_PORT", "50052")
os.Setenv("RTSA_HEALTH_PORT", "9090")
os.Setenv("RTSA_REDPANDA_BROKERS", "broker1:9092,broker2:9092")
os.Setenv("RTSA_RATE_LIMIT_PER_MINUTE", "20")
os.Setenv("RTSA_LOG_LEVEL", "debug")
os.Setenv("RTSA_SERVICE_NAME", "svc-feedback-custom")
defer func() {
os.Unsetenv("RTSA_GRPC_PORT")
os.Unsetenv("RTSA_HEALTH_PORT")
os.Unsetenv("RTSA_REDPANDA_BROKERS")
os.Unsetenv("RTSA_RATE_LIMIT_PER_MINUTE")
os.Unsetenv("RTSA_LOG_LEVEL")
os.Unsetenv("RTSA_SERVICE_NAME")
}()

cfg, err := Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.GRPCPort != "50052" {
t.Errorf("expected 50052, got %s", cfg.GRPCPort)
}
if cfg.RateLimitPerMin != 20 {
t.Errorf("expected 20, got %d", cfg.RateLimitPerMin)
}
if len(cfg.RedpandaBrokers) != 2 {
t.Errorf("expected 2 brokers, got %v", cfg.RedpandaBrokers)
}
if cfg.ServiceName != "svc-feedback-custom" {
t.Errorf("expected custom service name, got %s", cfg.ServiceName)
}
}

func TestLoad_InvalidRateLimit(t *testing.T) {
os.Setenv("RTSA_RATE_LIMIT_PER_MINUTE", "not-a-number")
defer os.Unsetenv("RTSA_RATE_LIMIT_PER_MINUTE")

_, err := Load()
if err == nil {
t.Error("expected error for invalid rate limit")
}
}

func TestLoad_ZeroRateLimit(t *testing.T) {
os.Setenv("RTSA_RATE_LIMIT_PER_MINUTE", "0")
defer os.Unsetenv("RTSA_RATE_LIMIT_PER_MINUTE")

_, err := Load()
if err == nil {
t.Error("expected error for zero rate limit")
}
}

func TestLoad_NegativeRateLimit(t *testing.T) {
os.Setenv("RTSA_RATE_LIMIT_PER_MINUTE", "-5")
defer os.Unsetenv("RTSA_RATE_LIMIT_PER_MINUTE")

_, err := Load()
if err == nil {
t.Error("expected error for negative rate limit")
}
}

func TestSplitBrokers_MultipleBrokers(t *testing.T) {
brokers := splitBrokers("broker1:9092, broker2:9092, broker3:9092")
if len(brokers) != 3 {
t.Errorf("expected 3 brokers, got %d: %v", len(brokers), brokers)
}
}

func TestSplitBrokers_EmptyString(t *testing.T) {
brokers := splitBrokers("")
if len(brokers) != 1 || brokers[0] != "localhost:19092" {
t.Errorf("expected default broker for empty string, got %v", brokers)
}
}

func TestSplitBrokers_SingleBroker(t *testing.T) {
brokers := splitBrokers("redpanda:19092")
if len(brokers) != 1 || brokers[0] != "redpanda:19092" {
t.Errorf("expected single broker, got %v", brokers)
}
}

func TestGetEnv_UsesDefault(t *testing.T) {
os.Unsetenv("RTSA_NON_EXISTENT_KEY")
val := getEnv("RTSA_NON_EXISTENT_KEY", "default-val")
if val != "default-val" {
t.Errorf("expected default-val, got %s", val)
}
}

func TestGetEnv_UsesEnvVar(t *testing.T) {
os.Setenv("RTSA_TEST_KEY", "custom-val")
defer os.Unsetenv("RTSA_TEST_KEY")
val := getEnv("RTSA_TEST_KEY", "default-val")
if val != "custom-val" {
t.Errorf("expected custom-val, got %s", val)
}
}
