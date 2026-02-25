// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-cyber-ingestion")
os.Unsetenv("RTSA_CYBER_OUTPUT_TOPIC")
os.Unsetenv("RTSA_CYBER_DLQ_TOPIC")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "sensors.cyber.iocs" {
t.Errorf("expected default output topic, got %s", cfg.OutputTopic)
}
if cfg.DLQTopic != "dlq.sensors.cyber" {
t.Errorf("expected default DLQ topic, got %s", cfg.DLQTopic)
}
if cfg.DedupCacheSize != 1000 {
t.Errorf("expected 1000 dedup cache size, got %d", cfg.DedupCacheSize)
}
}

func TestLoad_CustomValues(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-cyber-ingestion")
t.Setenv("RTSA_CYBER_OUTPUT_TOPIC", "custom.cyber.topic")
t.Setenv("RTSA_CYBER_DLQ_TOPIC", "custom.dlq.topic")
t.Setenv("RTSA_CYBER_DEDUP_CACHE_SIZE", "2000")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "custom.cyber.topic" {
t.Errorf("expected custom output topic, got %s", cfg.OutputTopic)
}
if cfg.DedupCacheSize != 2000 {
t.Errorf("expected 2000 dedup cache size, got %d", cfg.DedupCacheSize)
}
}

func TestLoad_MissingBaseConfig(t *testing.T) {
os.Unsetenv("RTSA_SERVICE_NAME")
_, err := config.Load()
if err == nil {
t.Fatal("expected error for missing RTSA_SERVICE_NAME")
}
}

func TestLoad_InvalidInt(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-cyber-ingestion")
t.Setenv("RTSA_CYBER_DEDUP_CACHE_SIZE", "notanint")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.DedupCacheSize != 1000 {
t.Errorf("expected default 1000 on invalid int, got %d", cfg.DedupCacheSize)
}
}

func TestMustLoad_Success(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-cyber-ingestion")

defer func() {
if r := recover(); r != nil {
t.Errorf("MustLoad should not panic with valid config: %v", r)
}
}()

cfg := config.MustLoad()
if cfg == nil {
t.Error("expected non-nil config")
}
}
