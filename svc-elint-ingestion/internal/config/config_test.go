// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-elint-ingestion")
os.Unsetenv("RTSA_ELINT_OUTPUT_TOPIC")
os.Unsetenv("RTSA_ELINT_DLQ_TOPIC")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "sensors.elint.detections" {
t.Errorf("expected default output topic, got %s", cfg.OutputTopic)
}
if cfg.DLQTopic != "dlq.sensors.elint" {
t.Errorf("expected default DLQ topic, got %s", cfg.DLQTopic)
}
if cfg.MaxCEPMeters != 50000.0 {
t.Errorf("expected 50000.0 max CEP, got %f", cfg.MaxCEPMeters)
}
}

func TestLoad_CustomValues(t *testing.T) {
t.Setenv("RTSA_SERVICE_NAME", "svc-elint-ingestion")
t.Setenv("RTSA_ELINT_OUTPUT_TOPIC", "custom.elint.topic")
t.Setenv("RTSA_ELINT_DLQ_TOPIC", "custom.dlq.topic")
t.Setenv("RTSA_ELINT_MAX_CEP_METERS", "10000.0")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.OutputTopic != "custom.elint.topic" {
t.Errorf("expected custom output topic, got %s", cfg.OutputTopic)
}
if cfg.MaxCEPMeters != 10000.0 {
t.Errorf("expected 10000.0 max CEP, got %f", cfg.MaxCEPMeters)
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
t.Setenv("RTSA_SERVICE_NAME", "svc-elint-ingestion")
t.Setenv("RTSA_ELINT_MAX_CEP_METERS", "notafloat")

cfg, err := config.Load()
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if cfg.MaxCEPMeters != 50000.0 {
t.Errorf("expected default 50000.0 on invalid float, got %f", cfg.MaxCEPMeters)
}
}
