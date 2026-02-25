// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-isr-ingestion/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-isr-ingestion")
os.Unsetenv("RTSA_ISR_OUTPUT_TOPIC")
os.Unsetenv("RTSA_ISR_DLQ_TOPIC")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "sensors.isr.observations" {
t.Errorf("expected default output topic, got %s", cfg.OutputTopic)
}
if cfg.DLQTopic != "dlq.sensors.isr" {
t.Errorf("expected default DLQ topic, got %s", cfg.DLQTopic)
}
if cfg.MinPolygonVertices != 3 {
t.Errorf("expected 3 min polygon vertices, got %d", cfg.MinPolygonVertices)
}
}

func TestLoad_CustomValues(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-isr-ingestion")
t.Setenv("RTSA_ISR_OUTPUT_TOPIC", "custom.isr.topic")
t.Setenv("RTSA_ISR_DLQ_TOPIC", "custom.dlq.topic")
t.Setenv("RTSA_ISR_MIN_POLYGON_VERTS", "5")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "custom.isr.topic" {
t.Errorf("expected custom output topic, got %s", cfg.OutputTopic)
}
if cfg.MinPolygonVertices != 5 {
t.Errorf("expected 5 min polygon vertices, got %d", cfg.MinPolygonVertices)
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
t.Setenv("RTSA_SERVICE_NAME", "svc-isr-ingestion")
t.Setenv("RTSA_ISR_MIN_POLYGON_VERTS", "notanint")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MinPolygonVertices != 3 {
t.Errorf("expected default 3 on invalid int, got %d", cfg.MinPolygonVertices)
}
}
