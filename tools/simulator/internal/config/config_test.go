// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
"os"
"testing"
"time"

"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
// Clear any env vars that might affect defaults.
for _, key := range []string{
"SIM_RADAR_ENDPOINT", "SIM_SURFACE_ENTITIES", "SIM_ANOMALY_RATE",
"SIM_UPDATE_INTERVAL_MS", "SIM_TLS_ENABLED",
} {
os.Unsetenv(key) //nolint:errcheck
}

cfg, err := config.Load()
if err != nil {
t.Fatalf("Load() returned error: %v", err)
}

if cfg.RadarEndpoint != "localhost:50051" {
t.Errorf("expected default RadarEndpoint localhost:50051, got %q", cfg.RadarEndpoint)
}
if cfg.SurfaceEntityCount != 20 {
t.Errorf("expected default surface count 20, got %d", cfg.SurfaceEntityCount)
}
if cfg.AirEntityCount != 10 {
t.Errorf("expected default air count 10, got %d", cfg.AirEntityCount)
}
if cfg.SubEntityCount != 5 {
t.Errorf("expected default sub count 5, got %d", cfg.SubEntityCount)
}
if cfg.AnomalyRate != 0.05 {
t.Errorf("expected default anomaly rate 0.05, got %f", cfg.AnomalyRate)
}
if cfg.UpdateInterval != 1000*time.Millisecond {
t.Errorf("expected default update interval 1s, got %v", cfg.UpdateInterval)
}
}

func TestLoad_EnvOverrides(t *testing.T) {
os.Setenv("SIM_RADAR_ENDPOINT", "radar-svc:50051")
os.Setenv("SIM_SURFACE_ENTITIES", "50")
os.Setenv("SIM_ANOMALY_RATE", "0.20")
os.Setenv("SIM_UPDATE_INTERVAL_MS", "500")
defer func() {
os.Unsetenv("SIM_RADAR_ENDPOINT")
os.Unsetenv("SIM_SURFACE_ENTITIES")
os.Unsetenv("SIM_ANOMALY_RATE")
os.Unsetenv("SIM_UPDATE_INTERVAL_MS")
}()

cfg, err := config.Load()
if err != nil {
t.Fatalf("Load() returned error: %v", err)
}

if cfg.RadarEndpoint != "radar-svc:50051" {
t.Errorf("expected RadarEndpoint radar-svc:50051, got %q", cfg.RadarEndpoint)
}
if cfg.SurfaceEntityCount != 50 {
t.Errorf("expected 50 surface entities, got %d", cfg.SurfaceEntityCount)
}
if cfg.AnomalyRate != 0.20 {
t.Errorf("expected anomaly rate 0.20, got %f", cfg.AnomalyRate)
}
if cfg.UpdateInterval != 500*time.Millisecond {
t.Errorf("expected 500ms update interval, got %v", cfg.UpdateInterval)
}
}

func TestLoad_NegativeEntityCount(t *testing.T) {
os.Setenv("SIM_SURFACE_ENTITIES", "-5")
defer os.Unsetenv("SIM_SURFACE_ENTITIES")

_, err := config.Load()
if err == nil {
t.Error("expected error for negative surface entity count")
}
}

func TestLoad_InvalidAnomalyRate(t *testing.T) {
os.Setenv("SIM_ANOMALY_RATE", "1.5")
defer os.Unsetenv("SIM_ANOMALY_RATE")

_, err := config.Load()
if err == nil {
t.Error("expected error for anomaly rate > 1.0")
}
}

func TestLoad_TLSMissingFiles(t *testing.T) {
os.Setenv("SIM_TLS_ENABLED", "true")
defer os.Unsetenv("SIM_TLS_ENABLED")

_, err := config.Load()
if err == nil {
t.Error("expected error for TLS enabled without cert/key/CA files")
}
}
