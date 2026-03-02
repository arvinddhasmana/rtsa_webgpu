// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/config"
)

func TestLoad_Exhaustive(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-ew-ingestion")
	t.Setenv("RTSA_EW_OUTPUT_TOPIC", "out")
	t.Setenv("RTSA_EW_DLQ_TOPIC", "dlq")
	t.Setenv("RTSA_EW_MAX_FREQUENCY_MHZ", "50000")
	t.Setenv("RTSA_EW_MAX_FUTURE_OFFSET_SEC", "120")
	t.Setenv("RTSA_EW_MAX_PAST_OFFSET_SEC", "3600")
	t.Setenv("RTSA_EW_RANGE_NM", "100")
	t.Setenv("RTSA_EW_BEARING_START", "10")
	t.Setenv("RTSA_EW_BEARING_END", "20")
	t.Setenv("RTSA_EW_LAT", "1.0")
	t.Setenv("RTSA_EW_LON", "2.0")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputTopic != "out" || cfg.DLQTopic != "dlq" {
		t.Error("topic mismatch")
	}
	if cfg.MaxFrequencyMHz != 50000 || cfg.MaxFutureOffsetSec != 120 || cfg.MaxPastOffsetSec != 3600 {
		t.Error("value mismatch")
	}
	if cfg.Coverage == nil || *cfg.Coverage.RangeNm != 100 || cfg.Coverage.SensorPosition.Latitude != 1.0 {
		t.Error("coverage mismatch")
	}
}

func TestLoad_InvalidParsing(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-ew-ingestion")
	t.Setenv("RTSA_EW_MAX_FREQUENCY_MHZ", "bad")
	t.Setenv("RTSA_EW_MAX_FUTURE_OFFSET_SEC", "bad")
	t.Setenv("RTSA_EW_MAX_PAST_OFFSET_SEC", "bad")
	t.Setenv("RTSA_EW_RANGE_NM", "bad")
	t.Setenv("RTSA_EW_BEARING_START", "bad")
	t.Setenv("RTSA_EW_BEARING_END", "bad")
	t.Setenv("RTSA_EW_LAT", "bad")
	t.Setenv("RTSA_EW_LON", "bad")

	cfg, _ := config.Load()
	if cfg.MaxFrequencyMHz != 40000.0 {
		t.Errorf("expected default 40000, got %v", cfg.MaxFrequencyMHz)
	}
}

func TestMustLoad(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-ew-ingestion")
	config.MustLoad()
}
