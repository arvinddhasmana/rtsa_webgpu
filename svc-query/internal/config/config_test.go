// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
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
if cfg.DefaultPageSize != 100 {
t.Errorf("expected default DefaultPageSize 100, got %d", cfg.DefaultPageSize)
}
}

func TestLoad_EnvOverrides(t *testing.T) {
os.Setenv("RTSA_QUERY_GRPC_PORT", "50099")
os.Setenv("RTSA_QUERY_MAX_RANGE_DAYS", "7")
defer func() {
os.Unsetenv("RTSA_QUERY_GRPC_PORT")
os.Unsetenv("RTSA_QUERY_MAX_RANGE_DAYS")
}()

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.GRPCPort != 50099 {
t.Errorf("expected GRPCPort 50099, got %d", cfg.GRPCPort)
}
if cfg.MaxQueryRangeDays != 7 {
t.Errorf("expected MaxQueryRangeDays 7, got %d", cfg.MaxQueryRangeDays)
}
}

func TestLoad_InvalidPort(t *testing.T) {
os.Setenv("RTSA_QUERY_GRPC_PORT", "notanint")
defer os.Unsetenv("RTSA_QUERY_GRPC_PORT")

_, err := config.Load()
if err == nil {
t.Error("expected error for invalid port, got nil")
}
}
