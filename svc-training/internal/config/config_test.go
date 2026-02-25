// CLASSIFICATION: UNCLASSIFIED
// Package config — unit tests for configuration loading.
package config

import (
"os"
"testing"
)

func TestLoad_Defaults(t *testing.T) {
os.Unsetenv("RTSA_HEALTH_PORT")
os.Unsetenv("INPUT_TOPIC")
os.Unsetenv("OUTPUT_TOPIC")
os.Unsetenv("RTSA_REDPANDA_BROKERS")
os.Unsetenv("CONSUMER_GROUP")
os.Unsetenv("RTSA_LOG_LEVEL")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if cfg.InputTopic != "feedback.operator.validated" {
t.Errorf("InputTopic = %q, want %q", cfg.InputTopic, "feedback.operator.validated")
}
if cfg.OutputTopic != "models.anomaly.candidates" {
t.Errorf("OutputTopic = %q, want %q", cfg.OutputTopic, "models.anomaly.candidates")
}
if cfg.ConsumerGroup != "svc-training" {
t.Errorf("ConsumerGroup = %q, want %q", cfg.ConsumerGroup, "svc-training")
}
if cfg.ServiceName != "svc-training" {
t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "svc-training")
}
if cfg.LogLevel != "info" {
t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
}
}

func TestLoad_CustomTopics(t *testing.T) {
os.Setenv("INPUT_TOPIC", "custom.input.topic")
os.Setenv("OUTPUT_TOPIC", "custom.output.topic")
defer os.Unsetenv("INPUT_TOPIC")
defer os.Unsetenv("OUTPUT_TOPIC")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if cfg.InputTopic != "custom.input.topic" {
t.Errorf("InputTopic = %q, want %q", cfg.InputTopic, "custom.input.topic")
}
if cfg.OutputTopic != "custom.output.topic" {
t.Errorf("OutputTopic = %q, want %q", cfg.OutputTopic, "custom.output.topic")
}
}

func TestLoad_CustomBrokers(t *testing.T) {
os.Setenv("RTSA_REDPANDA_BROKERS", "broker1:9092,broker2:9092")
defer os.Unsetenv("RTSA_REDPANDA_BROKERS")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if len(cfg.Brokers) != 2 {
t.Errorf("Brokers length = %d, want 2", len(cfg.Brokers))
}
if cfg.Brokers[0] != "broker1:9092" {
t.Errorf("Brokers[0] = %q, want %q", cfg.Brokers[0], "broker1:9092")
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

func TestLoad_CustomConsumerGroup(t *testing.T) {
os.Setenv("CONSUMER_GROUP", "my-training-group")
defer os.Unsetenv("CONSUMER_GROUP")

cfg, err := Load()
if err != nil {
t.Fatalf("Load() error: %v", err)
}
if cfg.ConsumerGroup != "my-training-group" {
t.Errorf("ConsumerGroup = %q, want %q", cfg.ConsumerGroup, "my-training-group")
}
}
