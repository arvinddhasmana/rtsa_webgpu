// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-ew-ingestion")
os.Unsetenv("RTSA_EW_OUTPUT_TOPIC")
os.Unsetenv("RTSA_EW_DLQ_TOPIC")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "sensors.ew.intercepts" {
t.Errorf("expected default output topic, got %s", cfg.OutputTopic)
}
if cfg.DLQTopic != "dlq.sensors.ew" {
t.Errorf("expected default DLQ topic, got %s", cfg.DLQTopic)
}
if cfg.MaxFrequencyMHz != 40000.0 {
t.Errorf("expected 40000.0 max frequency, got %f", cfg.MaxFrequencyMHz)
}
if cfg.MaxFutureOffsetSec != 300 {
t.Errorf("expected 300s future offset, got %d", cfg.MaxFutureOffsetSec)
}
	if cfg.MaxPastOffsetSec != 86400 {
		t.Errorf("expected 86400s past offset, got %d", cfg.MaxPastOffsetSec)
	}
}

func TestLoad_CustomValues(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-ew-ingestion")
t.Setenv("RTSA_EW_OUTPUT_TOPIC", "custom.ew.topic")
t.Setenv("RTSA_EW_DLQ_TOPIC", "custom.dlq.topic")
t.Setenv("RTSA_EW_MAX_FREQUENCY_MHZ", "20000.0")
t.Setenv("RTSA_EW_MAX_FUTURE_OFFSET_SEC", "600")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "custom.ew.topic" {
t.Errorf("expected custom output topic, got %s", cfg.OutputTopic)
}
if cfg.MaxFrequencyMHz != 20000.0 {
t.Errorf("expected 20000.0 max frequency, got %f", cfg.MaxFrequencyMHz)
}
if cfg.MaxFutureOffsetSec != 600 {
t.Errorf("expected 600 future offset, got %d", cfg.MaxFutureOffsetSec)
}
}

func TestLoad_MissingBaseConfig(t *testing.T) {
os.Unsetenv("RTSA_SERVICE_NAME")
_, err := config.Load()
if err == nil {
t.Fatal("expected error for missing RTSA_SERVICE_NAME")
}
}

func TestLoad_InvalidFloat(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-ew-ingestion")
t.Setenv("RTSA_EW_MAX_FREQUENCY_MHZ", "notafloat")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MaxFrequencyMHz != 40000.0 {
t.Errorf("expected default 40000.0 on invalid float, got %f", cfg.MaxFrequencyMHz)
}
}

func TestLoad_InvalidInt(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-ew-ingestion")
t.Setenv("RTSA_EW_MAX_FUTURE_OFFSET_SEC", "notanint")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MaxFutureOffsetSec != 300 {
t.Errorf("expected default 300 on invalid int, got %d", cfg.MaxFutureOffsetSec)
}
}
