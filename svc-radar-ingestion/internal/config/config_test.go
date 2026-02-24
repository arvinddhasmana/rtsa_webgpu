// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-radar-ingestion")
os.Unsetenv("RTSA_RADAR_OUTPUT_TOPIC")
os.Unsetenv("RTSA_RADAR_DLQ_TOPIC")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "sensors.radar.tracks" {
t.Errorf("expected default output topic, got %s", cfg.OutputTopic)
}
if cfg.DLQTopic != "dlq.sensors.radar" {
t.Errorf("expected default DLQ topic, got %s", cfg.DLQTopic)
}
if cfg.MaxFutureOffsetSec != 300 {
t.Errorf("expected 300s future offset, got %d", cfg.MaxFutureOffsetSec)
}
if cfg.MaxPastOffsetSec != 86400 {
t.Errorf("expected 86400s past offset, got %d", cfg.MaxPastOffsetSec)
}
}

func TestLoad_CustomValues(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-radar-ingestion")
t.Setenv("RTSA_RADAR_OUTPUT_TOPIC", "custom.radar.topic")
t.Setenv("RTSA_RADAR_DLQ_TOPIC", "custom.dlq.topic")
t.Setenv("RTSA_RADAR_MAX_SURFACE_SPEED", "500.0")
t.Setenv("RTSA_RADAR_MAX_FUTURE_OFFSET", "600")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "custom.radar.topic" {
t.Errorf("expected custom output topic, got %s", cfg.OutputTopic)
}
if cfg.MaxSurfaceSpeedKnots != 500.0 {
t.Errorf("expected 500.0 speed, got %f", cfg.MaxSurfaceSpeedKnots)
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
t.Setenv("RTSA_SERVICE_NAME", "svc")
t.Setenv("RTSA_RADAR_MAX_SURFACE_SPEED", "notafloat")

// Should use default value, not error
cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MaxSurfaceSpeedKnots != 999.0 {
t.Errorf("expected default 999.0 on invalid float, got %f", cfg.MaxSurfaceSpeedKnots)
}
}

func TestLoad_InvalidInt(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc")
t.Setenv("RTSA_RADAR_MAX_FUTURE_OFFSET", "notanint")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MaxFutureOffsetSec != 300 {
t.Errorf("expected default 300 on invalid int, got %d", cfg.MaxFutureOffsetSec)
}
}
