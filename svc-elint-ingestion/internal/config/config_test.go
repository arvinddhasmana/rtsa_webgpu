// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"testing"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-elint-ingestion/internal/config"
)

func TestLoad_Exhaustive(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-elint-ingestion")
	t.Setenv("RTSA_ELINT_OUTPUT_TOPIC", "out")
	t.Setenv("RTSA_ELINT_DLQ_TOPIC", "dlq")
	t.Setenv("RTSA_ELINT_MAX_CEP_METERS", "1000")
	t.Setenv("RTSA_ELINT_RANGE_NM", "100")
	t.Setenv("RTSA_ELINT_LAT", "1.0")
	t.Setenv("RTSA_ELINT_LON", "2.0")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputTopic != "out" || cfg.DLQTopic != "dlq" {
		t.Error("topic mismatch")
	}
	if cfg.MaxCEPMeters != 1000 {
		t.Error("value mismatch")
	}
	if cfg.Coverage == nil || *cfg.Coverage.RangeNm != 100 || cfg.Coverage.SensorPosition.Latitude != 1.0 {
		t.Error("coverage mismatch")
	}
}

func TestLoad_InvalidParsing(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-elint-ingestion")
	t.Setenv("RTSA_ELINT_MAX_CEP_METERS", "bad")
	t.Setenv("RTSA_ELINT_RANGE_NM", "bad")
	t.Setenv("RTSA_ELINT_LAT", "bad")
	t.Setenv("RTSA_ELINT_LON", "bad")

	cfg, _ := config.Load()
	if cfg.MaxCEPMeters != 50000.0 {
		t.Errorf("expected default 50000, got %v", cfg.MaxCEPMeters)
	}
}

func TestMustLoad(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-elint-ingestion")
	config.MustLoad()
}
