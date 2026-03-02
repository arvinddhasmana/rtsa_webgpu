// CLASSIFICATION: UNCLASSIFIED
package config_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-isr-ingestion/internal/config"
)

func TestLoad_Exhaustive(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-isr-ingestion")
	t.Setenv("RTSA_ISR_OUTPUT_TOPIC", "out")
	t.Setenv("RTSA_ISR_DLQ_TOPIC", "dlq")
	t.Setenv("RTSA_ISR_MIN_POLYGON_VERTS", "5")
	t.Setenv("RTSA_ISR_RANGE_NM", "100")
	t.Setenv("RTSA_ISR_BEARING_START", "10")
	t.Setenv("RTSA_ISR_BEARING_END", "20")
	t.Setenv("RTSA_ISR_LAT", "1.0")
	t.Setenv("RTSA_ISR_LON", "2.0")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputTopic != "out" || cfg.DLQTopic != "dlq" {
		t.Error("topic mismatch")
	}
	if cfg.MinPolygonVertices != 5 {
		t.Error("value mismatch")
	}
	if cfg.Coverage == nil || *cfg.Coverage.RangeNm != 100 || cfg.Coverage.SensorPosition.Latitude != 1.0 {
		t.Error("coverage mismatch")
	}
}

func TestLoad_InvalidParsing(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-isr-ingestion")
	t.Setenv("RTSA_ISR_MIN_POLYGON_VERTS", "bad")
	t.Setenv("RTSA_ISR_RANGE_NM", "bad")
	t.Setenv("RTSA_ISR_BEARING_START", "bad")
	t.Setenv("RTSA_ISR_BEARING_END", "bad")
	t.Setenv("RTSA_ISR_LAT", "bad")
	t.Setenv("RTSA_ISR_LON", "bad")

	cfg, _ := config.Load()
	if cfg.MinPolygonVertices != 3 {
		t.Errorf("expected default 3, got %d", cfg.MinPolygonVertices)
	}
}

func TestMustLoad(t *testing.T) {
	t.Setenv("RTSA_SERVICE_NAME", "svc-isr-ingestion")
	config.MustLoad()
}
