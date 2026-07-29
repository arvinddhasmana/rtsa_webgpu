// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"os"
	"testing"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-radar-ingestion/internal/config"
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
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-radar-ingestion")
	t.Setenv("RTSA_RADAR_OUTPUT_TOPIC", "custom.radar")
	t.Setenv("RTSA_RADAR_MAX_SURFACE_SPEED", "100.0")
	t.Setenv("RTSA_RADAR_MAX_FUTURE_OFFSET", "600")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputTopic != "custom.radar" {
		t.Errorf("expected custom output topic, got %s", cfg.OutputTopic)
	}
	if cfg.MaxSurfaceSpeedKnots != 100.0 {
		t.Errorf("expected 100.0, got %f", cfg.MaxSurfaceSpeedKnots)
	}
	if cfg.MaxFutureOffsetSec != 600 {
		t.Errorf("expected 600, got %d", cfg.MaxFutureOffsetSec)
	}
}

func TestLoad_InvalidValues(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-radar-ingestion")
	t.Setenv("RTSA_RADAR_MAX_SURFACE_SPEED", "notafloat")
	t.Setenv("RTSA_RADAR_MAX_FUTURE_OFFSET", "notanint")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxSurfaceSpeedKnots != 999.0 {
		t.Errorf("expected default 999.0, got %f", cfg.MaxSurfaceSpeedKnots)
	}
	if cfg.MaxFutureOffsetSec != 300 {
		t.Errorf("expected default 300 on invalid int, got %d", cfg.MaxFutureOffsetSec)
	}
}

func TestLoad_Coverage(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-radar-ingestion")
	t.Setenv("RTSA_RADAR_RANGE_NM", "50.0")
	t.Setenv("RTSA_RADAR_BEARING_START", "10.0")
	t.Setenv("RTSA_RADAR_BEARING_END", "170.0")
	t.Setenv("RTSA_RADAR_LAT", "45.0")
	t.Setenv("RTSA_RADAR_LON", "-60.0")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Coverage == nil {
		t.Fatal("expected coverage")
	}
	if *cfg.Coverage.RangeNm != 50.0 {
		t.Errorf("expected 50.0, got %f", *cfg.Coverage.RangeNm)
	}
	if *cfg.Coverage.BearingStartDegrees != 10.0 {
		t.Errorf("expected 10.0, got %f", *cfg.Coverage.BearingStartDegrees)
	}
}

func TestLoad_InvalidCoverage(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-radar-ingestion")
	t.Setenv("RTSA_RADAR_RANGE_NM", "invalid")
	t.Setenv("RTSA_RADAR_LAT", "invalid")

	cfg, _ := config.Load()
	if cfg.Coverage != nil {
		t.Error("expected nil coverage for invalid input")
	}
}

func TestMustLoad(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-radar-ingestion")
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustLoad panicked: %v", r)
		}
	}()
	config.MustLoad()
}
