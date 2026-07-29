// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"os"
	"testing"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GRPCPort != 50072 {
		t.Errorf("expected default GRPCPort 50072, got %d", cfg.GRPCPort)
	}
	if cfg.MaxQueryRangeDays != 30 {
		t.Errorf("expected default MaxQueryRangeDays 30, got %d", cfg.MaxQueryRangeDays)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("RTSA_QUERY_GRPC_PORT", "50099")
	os.Setenv("RTSA_QUERY_MAX_RANGE_DAYS", "7")
	os.Setenv("RTSA_HEALTH_PORT", "8083")
	os.Setenv("RTSA_METRICS_PORT", "9093")
	os.Setenv("RTSA_CLICKHOUSE_DSN", "ch://test")
	os.Setenv("RTSA_CLICKHOUSE_DATABASE", "test-db")
	os.Setenv("RTSA_QUERY_MAX_ROWS", "500")
	os.Setenv("RTSA_QUERY_TIMEOUT_SEC", "60")
	os.Setenv("RTSA_QUERY_DEFAULT_PAGE_SIZE", "50")
	os.Setenv("RTSA_QUERY_MAX_PAGE_SIZE", "200")
	os.Setenv("RTSA_REDPANDA_BROKERS", "r1:9092,r2:9092")
	os.Setenv("RTSA_SERVICE_NAME", "query-test")
	os.Setenv("RTSA_LOG_LEVEL", "debug")
	os.Setenv("RTSA_LOG_FORMAT", "text")

	defer os.Clearenv()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GRPCPort != 50099 {
		t.Errorf("expected GRPCPort 50099, got %d", cfg.GRPCPort)
	}
	if cfg.HealthPort != 8083 {
		t.Errorf("expected HealthPort 8083, got %d", cfg.HealthPort)
	}
	if cfg.MetricsPort != 9093 {
		t.Errorf("expected MetricsPort 9093, got %d", cfg.MetricsPort)
	}
	if cfg.MaxResultRows != 500 {
		t.Errorf("expected MaxResultRows 500, got %d", cfg.MaxResultRows)
	}
	if len(cfg.RedpandaBrokers) != 2 {
		t.Errorf("expected 2 brokers, got %d", len(cfg.RedpandaBrokers))
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected text format, got %s", cfg.LogFormat)
	}
}

func TestLoad_InvalidInts(t *testing.T) {
	tests := []struct {
		env   string
		value string
	}{
		{"RTSA_QUERY_GRPC_PORT", "abc"},
		{"RTSA_GRPC_PORT", "abc"},
		{"RTSA_HEALTH_PORT", "abc"},
		{"RTSA_METRICS_PORT", "abc"},
		{"RTSA_QUERY_MAX_RANGE_DAYS", "abc"},
		{"RTSA_QUERY_MAX_ROWS", "abc"},
		{"RTSA_QUERY_TIMEOUT_SEC", "abc"},
		{"RTSA_QUERY_DEFAULT_PAGE_SIZE", "abc"},
		{"RTSA_QUERY_MAX_PAGE_SIZE", "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.env, func(t *testing.T) {
			os.Clearenv()
			os.Setenv(tc.env, tc.value)
			_, err := config.Load()
			if err == nil {
				t.Errorf("expected error for %s=%s", tc.env, tc.value)
			}
		})
	}
}

func TestLoad_FallbackGRPCPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("RTSA_GRPC_PORT", "50001")
	cfg, _ := config.Load()
	if cfg.GRPCPort != 50001 {
		t.Errorf("expected fallback GRPCPort 50001, got %d", cfg.GRPCPort)
	}
}
