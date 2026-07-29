// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"os"
	"testing"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-nato-adapter/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-nato-adapter")
	os.Unsetenv("RTSA_GRPC_PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GRPCAddr != ":50051" {
		t.Errorf("expected :50051, got %s", cfg.GRPCAddr)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-nato-adapter")
	t.Setenv("RTSA_GRPC_PORT", "60000")
	t.Setenv("RTSA_LOG_LEVEL", "debug")

	cfg, _ := config.Load()
	if cfg.GRPCAddr != ":60000" {
		t.Errorf("expected :60000, got %s", cfg.GRPCAddr)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-nato-adapter")
	t.Setenv("RTSA_LOG_LEVEL", "invalid")
	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid log level")
	}
}
