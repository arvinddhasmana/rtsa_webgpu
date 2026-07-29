// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"testing"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-training/internal/config"
)

func TestLoad_Exhaustive(t *testing.T) {
	t.Setenv("RTSA_REDPANDA_BROKERS", "broker1:9092, broker2:9092")
	t.Setenv("RTSA_HEALTH_PORT", "9090")
	t.Setenv("INPUT_TOPIC", "in")
	t.Setenv("OUTPUT_TOPIC", "out")
	t.Setenv("CONSUMER_GROUP", "group")
	t.Setenv("RTSA_ENVIRONMENT", "prod")
	t.Setenv("RTSA_OTEL_ENDPOINT", "otel:4317")
	t.Setenv("RTSA_SERVICE_NAME", "training-svc")
	t.Setenv("RTSA_LOG_LEVEL", "debug")
	t.Setenv("RTSA_LOG_FORMAT", "text")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Brokers) != 2 || cfg.Brokers[0] != "broker1:9092" {
		t.Error("brokers mismatch")
	}
	if cfg.HealthPort != ":9090" {
		t.Error("health port mismatch")
	}
	if cfg.InputTopic != "in" || cfg.OutputTopic != "out" || cfg.ConsumerGroup != "group" {
		t.Error("topic/group mismatch")
	}
	if cfg.LogLevel != "debug" || cfg.LogFormat != "text" {
		t.Error("log mismatch")
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("RTSA_LOG_LEVEL", "invalid")
	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}
