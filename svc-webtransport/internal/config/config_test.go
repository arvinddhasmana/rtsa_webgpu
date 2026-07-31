package config
// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-webtransport/internal/config"
)

// setEnv sets the given environment variables for the duration of the test and
// clears everything else that Load reads so tests are hermetic.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	keys := []string{
		"RTSA_SERVICE_NAME", "RTSA_ENV", "RTSA_WT_LISTEN_ADDR",
		"RTSA_TLS_SERVER_CERT", "RTSA_TLS_SERVER_KEY", "RTSA_TLS_CA_CERT",
		"RTSA_WT_JWT_SECRET", "RTSA_WT_ALLOWED_ORIGINS", "RTSA_WT_MAX_SESSIONS",
		"RTSA_WT_DATAGRAM_BATCH", "RTSA_REDPANDA_BROKERS", "RTSA_REDPANDA_TLS_ENABLED",
		"RTSA_CONSUMER_GROUP", "RTSA_WT_TOPICS", "RTSA_WT_START_OFFSET",
		"RTSA_WT_SUBSCRIBER_BUFFER", "RTSA_HEALTH_PORT", "RTSA_METRICS_PORT",
		"RTSA_OTEL_ENDPOINT", "RTSA_LOG_LEVEL",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestLoad_Defaults(t *testing.T) {
	setEnv(t, map[string]string{"RTSA_WT_JWT_SECRET": "dev-secret"})

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.ServiceName != "svc-webtransport" {
		t.Errorf("ServiceName = %q, want svc-webtransport", cfg.ServiceName)
	}
	if cfg.WTListenAddr != ":4443" {
		t.Errorf("WTListenAddr = %q, want :4443", cfg.WTListenAddr)
	}
	if cfg.DatagramBatchSize != 9 {
		t.Errorf("DatagramBatchSize = %d, want 9", cfg.DatagramBatchSize)
	}
	if cfg.StartOffset != "latest" {
		t.Errorf("StartOffset = %q, want latest", cfg.StartOffset)
	}
	if len(cfg.Topics) != 1 || cfg.Topics[0] != "tracks.fused" {
		t.Errorf("Topics = %v, want [tracks.fused]", cfg.Topics)
	}
	if cfg.HealthPort != ":8081" {
		t.Errorf("HealthPort = %q, want :8081", cfg.HealthPort)
	}
	if string(cfg.JWTSecret) != "dev-secret" {
		t.Errorf("JWTSecret = %q, want dev-secret", string(cfg.JWTSecret))
	}
}

func TestLoad_MissingJWTSecret_FailsClosed(t *testing.T) {
	setEnv(t, map[string]string{})

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error when RTSA_WT_JWT_SECRET is unset, got nil")
	}
}

func TestLoad_InvalidDatagramBatch(t *testing.T) {
	setEnv(t, map[string]string{
		"RTSA_WT_JWT_SECRET":     "s",
		"RTSA_WT_DATAGRAM_BATCH": "42",
	})

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for out-of-range datagram batch, got nil")
	}
}

func TestLoad_InvalidStartOffset(t *testing.T) {
	setEnv(t, map[string]string{
		"RTSA_WT_JWT_SECRET":   "s",
		"RTSA_WT_START_OFFSET": "sideways",
	})

	if _, err := config.Load(); err == nil {
		t.Fatal("expected error for invalid start offset, got nil")
	}
}

func TestLoad_ParsesCSVLists(t *testing.T) {
	setEnv(t, map[string]string{
		"RTSA_WT_JWT_SECRET":      "s",
		"RTSA_REDPANDA_BROKERS":   "a:9092, b:9092 ,c:9092",
		"RTSA_WT_ALLOWED_ORIGINS": "https://cop.example.mil, https://ops.example.mil",
		"RTSA_WT_TOPICS":          "tracks.fused.air,tracks.fused.surface",
	})

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(cfg.RedpandaBrokers) != 3 {
		t.Errorf("RedpandaBrokers = %v, want 3 entries", cfg.RedpandaBrokers)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins = %v, want 2 entries", cfg.AllowedOrigins)
	}
	if len(cfg.Topics) != 2 {
		t.Errorf("Topics = %v, want 2 entries", cfg.Topics)
	}
}
