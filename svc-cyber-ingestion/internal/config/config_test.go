// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/config"
)

func TestLoad_Exhaustive(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-cyber-ingestion")
	t.Setenv("RTSA_CYBER_OUTPUT_TOPIC", "out")
	t.Setenv("RTSA_CYBER_DLQ_TOPIC", "dlq")
	t.Setenv("RTSA_CYBER_DEDUP_CACHE_SIZE", "5000")
	t.Setenv("RTSA_CYBER_RANGE_NM", "100")
	t.Setenv("RTSA_CYBER_BEARING_START", "10")
	t.Setenv("RTSA_CYBER_BEARING_END", "20")
	t.Setenv("RTSA_CYBER_LAT", "1.0")
	t.Setenv("RTSA_CYBER_LON", "2.0")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputTopic != "out" || cfg.DLQTopic != "dlq" {
		t.Error("topic mismatch")
	}
	if cfg.DedupCacheSize != 5000 {
		t.Error("value mismatch")
	}
	if cfg.Coverage == nil || *cfg.Coverage.RangeNm != 100 || cfg.Coverage.SensorPosition.Latitude != 1.0 {
		t.Error("coverage mismatch")
	}
	if *cfg.Coverage.BearingStartDegrees != 10 || *cfg.Coverage.BearingEndDegrees != 20 {
		t.Error("bearing mismatch")
	}
}

func TestLoad_InvalidParsing(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-cyber-ingestion")
	t.Setenv("RTSA_CYBER_DEDUP_CACHE_SIZE", "bad")
	t.Setenv("RTSA_CYBER_RANGE_NM", "bad")
	t.Setenv("RTSA_CYBER_BEARING_START", "bad")
	t.Setenv("RTSA_CYBER_BEARING_END", "bad")
	t.Setenv("RTSA_CYBER_LAT", "bad")
	t.Setenv("RTSA_CYBER_LON", "bad")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DedupCacheSize != 1000 {
		t.Errorf("expected default 1000, got %d", cfg.DedupCacheSize)
	}
	// Defaults for coverage fields when parsing fails should lead to not setting them IF parseFloat is used with specific sentinels.
}

func TestMustLoad(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-cyber-ingestion")
	config.MustLoad()
}
