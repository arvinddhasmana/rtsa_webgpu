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

func TestLoad_GRPCPortOverride(t *testing.T) {
os.Setenv("RTSA_GRPC_PORT", "50099")
defer os.Unsetenv("RTSA_GRPC_PORT")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.GRPCPort != 50099 {
t.Errorf("expected GRPCPort 50099, got %d", cfg.GRPCPort)
}
}

func TestLoad_InvalidGRPCPort(t *testing.T) {
os.Setenv("RTSA_GRPC_PORT", "notanint")
defer os.Unsetenv("RTSA_GRPC_PORT")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid GRPC port")
}
}

func TestLoad_QueryTimeoutOverride(t *testing.T) {
os.Setenv("RTSA_AUDIT_QUERY_TIMEOUT_SEC", "60")
defer os.Unsetenv("RTSA_AUDIT_QUERY_TIMEOUT_SEC")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.QueryTimeoutSec != 60 {
t.Errorf("expected QueryTimeoutSec 60, got %d", cfg.QueryTimeoutSec)
}
}

func TestLoad_DefaultPageSizeOverride(t *testing.T) {
os.Setenv("RTSA_AUDIT_DEFAULT_PAGE_SIZE", "50")
defer os.Unsetenv("RTSA_AUDIT_DEFAULT_PAGE_SIZE")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.DefaultPageSize != 50 {
t.Errorf("expected DefaultPageSize 50, got %d", cfg.DefaultPageSize)
}
}

func TestLoad_InvalidQueryTimeout(t *testing.T) {
os.Setenv("RTSA_AUDIT_QUERY_TIMEOUT_SEC", "notanint")
defer os.Unsetenv("RTSA_AUDIT_QUERY_TIMEOUT_SEC")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid query timeout")
}
}

func TestLoad_InvalidDefaultPageSize(t *testing.T) {
os.Setenv("RTSA_AUDIT_DEFAULT_PAGE_SIZE", "bad")
defer os.Unsetenv("RTSA_AUDIT_DEFAULT_PAGE_SIZE")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid page size")
}
}

func TestLoad_HealthPortOverride(t *testing.T) {
os.Setenv("RTSA_HEALTH_PORT", "9999")
defer os.Unsetenv("RTSA_HEALTH_PORT")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.HealthPort != 9999 {
t.Errorf("expected HealthPort 9999, got %d", cfg.HealthPort)
}
}

func TestLoad_InvalidHealthPort(t *testing.T) {
os.Setenv("RTSA_HEALTH_PORT", "bad")
defer os.Unsetenv("RTSA_HEALTH_PORT")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid health port")
}
}

func TestLoad_InvalidFlushPeriod(t *testing.T) {
os.Setenv("RTSA_AUDIT_FLUSH_PERIOD_SEC", "bad")
defer os.Unsetenv("RTSA_AUDIT_FLUSH_PERIOD_SEC")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid flush period")
}
}

func TestLoad_InvalidMaxRangeDays(t *testing.T) {
os.Setenv("RTSA_AUDIT_MAX_RANGE_DAYS", "bad")
defer os.Unsetenv("RTSA_AUDIT_MAX_RANGE_DAYS")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid max range days")
}
}
