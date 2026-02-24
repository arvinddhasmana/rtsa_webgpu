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
os.Unsetenv("RTSA_REDPANDA_BROKERS")
os.Unsetenv("RTSA_TLS_ENABLED")
os.Unsetenv("RTSA_HISTORY_MAX_POINTS")
os.Unsetenv("RTSA_STREAM_CHANNEL_BUFFER")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if cfg.GRPCAddr != ":50051" {
t.Errorf("expected GRPCAddr=:50051, got %q", cfg.GRPCAddr)
}
if cfg.HistoryMaxPoints != 100 {
t.Errorf("expected HistoryMaxPoints=100, got %d", cfg.HistoryMaxPoints)
}
if cfg.TLSEnabled != false {
t.Errorf("expected TLSEnabled=false, got %v", cfg.TLSEnabled)
}
if cfg.ConsumerGroupID != "track-service" {
t.Errorf("expected ConsumerGroupID=track-service, got %q", cfg.ConsumerGroupID)
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
t.Errorf("expected 2 brokers, got %d", len(cfg.RedpandaBrokers))
}
if cfg.RedpandaBrokers[0] != "broker1:9092" {
t.Errorf("expected broker1:9092, got %q", cfg.RedpandaBrokers[0])
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

func TestLoad_InvalidHistoryMaxPoints(t *testing.T) {
os.Setenv("RTSA_HISTORY_MAX_POINTS", "not-a-number")
defer os.Unsetenv("RTSA_HISTORY_MAX_POINTS")

_, err := Load()
if err == nil {
t.Error("expected error for invalid RTSA_HISTORY_MAX_POINTS, got nil")
}
}

func TestLoad_ZeroHistoryMaxPoints(t *testing.T) {
os.Setenv("RTSA_HISTORY_MAX_POINTS", "0")
defer os.Unsetenv("RTSA_HISTORY_MAX_POINTS")

_, err := Load()
if err == nil {
t.Error("expected error for RTSA_HISTORY_MAX_POINTS=0, got nil")
}
}

func TestLoad_InvalidStreamBuffer(t *testing.T) {
os.Setenv("RTSA_STREAM_CHANNEL_BUFFER", "bad")
defer os.Unsetenv("RTSA_STREAM_CHANNEL_BUFFER")

_, err := Load()
if err == nil {
t.Error("expected error for invalid RTSA_STREAM_CHANNEL_BUFFER, got nil")
}
}

func TestLoad_ZeroStreamBuffer(t *testing.T) {
os.Setenv("RTSA_STREAM_CHANNEL_BUFFER", "0")
defer os.Unsetenv("RTSA_STREAM_CHANNEL_BUFFER")

_, err := Load()
if err == nil {
t.Error("expected error for RTSA_STREAM_CHANNEL_BUFFER=0, got nil")
}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
os.Setenv("RTSA_LOG_LEVEL", "verbose")
defer os.Unsetenv("RTSA_LOG_LEVEL")

_, err := Load()
if err == nil {
t.Error("expected error for invalid log level, got nil")
}
}

func TestLoad_TLSEnabled(t *testing.T) {
os.Setenv("RTSA_TLS_ENABLED", "true")
defer os.Unsetenv("RTSA_TLS_ENABLED")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if !cfg.TLSEnabled {
t.Error("expected TLSEnabled=true")
}
}

func TestGetEnv_Default(t *testing.T) {
os.Unsetenv("RTSA_TEST_KEY_NONEXISTENT")
val := getEnv("RTSA_TEST_KEY_NONEXISTENT", "mydefault")
if val != "mydefault" {
t.Errorf("expected mydefault, got %q", val)
}
}

func TestGetEnv_Set(t *testing.T) {
os.Setenv("RTSA_TEST_KEY_SET", "testvalue")
defer os.Unsetenv("RTSA_TEST_KEY_SET")
val := getEnv("RTSA_TEST_KEY_SET", "default")
if val != "testvalue" {
t.Errorf("expected testvalue, got %q", val)
}
}
