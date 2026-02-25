// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"os"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServiceName != "svc-audit" {
		t.Errorf("expected default ServiceName 'svc-audit', got %q", cfg.ServiceName)
	}
	if cfg.BatchSize != 500 {
		t.Errorf("expected default BatchSize 500, got %d", cfg.BatchSize)
	}
	if cfg.ConsumerGroup != "svc-audit" {
		t.Errorf("expected default ConsumerGroup 'svc-audit', got %q", cfg.ConsumerGroup)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("RTSA_AUDIT_BATCH_SIZE", "100")
	os.Setenv("RTSA_AUDIT_CONSUMER_GROUP", "svc-audit-custom")
	defer func() {
		os.Unsetenv("RTSA_AUDIT_BATCH_SIZE")
		os.Unsetenv("RTSA_AUDIT_CONSUMER_GROUP")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("expected BatchSize 100, got %d", cfg.BatchSize)
	}
	if cfg.ConsumerGroup != "svc-audit-custom" {
		t.Errorf("expected ConsumerGroup 'svc-audit-custom', got %q", cfg.ConsumerGroup)
	}
}

func TestLoad_InvalidBatchSize(t *testing.T) {
	os.Setenv("RTSA_AUDIT_BATCH_SIZE", "notanint")
	defer os.Unsetenv("RTSA_AUDIT_BATCH_SIZE")

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid batch size")
	}
}
