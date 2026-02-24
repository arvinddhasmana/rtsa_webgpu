// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"os"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

func TestLoadBase_Defaults(t *testing.T) {
	cfg := config.LoadBase()
	if cfg.ServiceName != "unknown" {
		t.Errorf("expected default service name 'unknown', got %q", cfg.ServiceName)
	}
	if cfg.GRPCPort != 50051 {
		t.Errorf("expected default gRPC port 50051, got %d", cfg.GRPCPort)
	}
	if len(cfg.RedpandaBrokers) == 0 {
		t.Error("expected non-empty default Redpanda brokers")
	}
}

func TestLoadBase_EnvOverride(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "test-svc")
	t.Setenv("RTSA_GRPC_PORT", "9090")
	t.Setenv("RTSA_REDPANDA_BROKERS", "broker1:9092,broker2:9092")
	t.Setenv("RTSA_REDPANDA_TLS_ENABLED", "false")

	cfg := config.LoadBase()
	if cfg.ServiceName != "test-svc" {
		t.Errorf("expected 'test-svc', got %q", cfg.ServiceName)
	}
	if cfg.GRPCPort != 9090 {
		t.Errorf("expected 9090, got %d", cfg.GRPCPort)
	}
	if len(cfg.RedpandaBrokers) != 2 {
		t.Errorf("expected 2 brokers, got %d", len(cfg.RedpandaBrokers))
	}
	if cfg.RedpandaTLSEnabled {
		t.Error("expected TLS disabled")
	}
}

func TestGetEnvFloat(t *testing.T) {
	os.Setenv("RTSA_TEST_FLOAT", "3.14")
	defer os.Unsetenv("RTSA_TEST_FLOAT")

	v := config.GetEnvFloat("RTSA_TEST_FLOAT", 0.0)
	if v != 3.14 {
		t.Errorf("expected 3.14, got %f", v)
	}
	v2 := config.GetEnvFloat("RTSA_NOT_SET", 2.71)
	if v2 != 2.71 {
		t.Errorf("expected default 2.71, got %f", v2)
	}
}

func TestGetEnvInt_InvalidFallsBack(t *testing.T) {
	os.Setenv("RTSA_TEST_INT", "notanumber")
	defer os.Unsetenv("RTSA_TEST_INT")
	v := config.GetEnvInt("RTSA_TEST_INT", 42)
	if v != 42 {
		t.Errorf("expected 42 fallback, got %d", v)
	}
}
