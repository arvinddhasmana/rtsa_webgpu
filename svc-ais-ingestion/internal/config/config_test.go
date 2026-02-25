// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ais-ingestion/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-ais-ingestion")
os.Unsetenv("RTSA_AIS_OUTPUT_TOPIC")
os.Unsetenv("RTSA_AIS_DLQ_TOPIC")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "sensors.ais.positions" {
t.Errorf("expected default output topic, got %s", cfg.OutputTopic)
}
if cfg.DLQTopic != "dlq.sensors.ais" {
t.Errorf("expected default DLQ topic, got %s", cfg.DLQTopic)
}
if cfg.MaxSpeedJumpKnots != 50.0 {
t.Errorf("expected 50.0 max speed jump, got %f", cfg.MaxSpeedJumpKnots)
}
}

func TestLoad_CustomValues(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-ais-ingestion")
t.Setenv("RTSA_AIS_OUTPUT_TOPIC", "custom.ais.topic")
t.Setenv("RTSA_AIS_DLQ_TOPIC", "custom.dlq.topic")
t.Setenv("RTSA_AIS_MAX_SPEED_JUMP_KNOTS", "30.0")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "custom.ais.topic" {
t.Errorf("expected custom output topic, got %s", cfg.OutputTopic)
}
if cfg.MaxSpeedJumpKnots != 30.0 {
t.Errorf("expected 30.0 max speed jump, got %f", cfg.MaxSpeedJumpKnots)
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
t.Setenv("RTSA_SERVICE_NAME", "svc-ais-ingestion")
t.Setenv("RTSA_AIS_MAX_SPEED_JUMP_KNOTS", "notafloat")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MaxSpeedJumpKnots != 50.0 {
t.Errorf("expected default 50.0 on invalid float, got %f", cfg.MaxSpeedJumpKnots)
}
}
