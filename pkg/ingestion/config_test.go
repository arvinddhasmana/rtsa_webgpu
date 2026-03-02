// CLASSIFICATION: UNCLASSIFIED
package ingestion_test

import (
	"os"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

func TestMustLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg := ingestion.MustLoad("test-svc", "out", "dlq", 5000)

	if cfg.ServiceName != "test-svc" {
		t.Errorf("expected test-svc, got %s", cfg.ServiceName)
	}
	if cfg.OutputTopic != "out" {
		t.Errorf("expected out, got %s", cfg.OutputTopic)
	}
	if cfg.GRPCPort != 5000 {
		t.Errorf("expected 5000, got %d", cfg.GRPCPort)
	}
}

func TestMustLoad_EnvOverrides(t *testing.T) {
	os.Setenv("OUTPUT_TOPIC", "new-out")
	os.Setenv("DLQ_TOPIC", "new-dlq")
	os.Setenv("REDPANDA_BROKERS", "r1:9092")
	os.Setenv("GRPC_PORT", "6000")
	os.Setenv("MAX_CLASSIFICATION", "CLASSIFICATION_LEVEL_UNCLASSIFIED")
	os.Setenv("SENSOR_RANGE_NM", "50.5")
	os.Setenv("SENSOR_BEARING_START", "10")
	os.Setenv("SENSOR_BEARING_END", "20")
	os.Setenv("SENSOR_LAT", "45.0")
	os.Setenv("SENSOR_LON", "-75.0")

	defer os.Clearenv()

	cfg := ingestion.MustLoad("test-svc", "out", "dlq", 5000)

	if cfg.OutputTopic != "new-out" {
		t.Errorf("expected new-out, got %s", cfg.OutputTopic)
	}
	if cfg.GRPCPort != 6000 {
		t.Errorf("expected 6000, got %d", cfg.GRPCPort)
	}
	if cfg.Coverage == nil {
		t.Fatal("expected coverage to be populated")
	}
	if *cfg.Coverage.RangeNm != 50.5 {
		t.Errorf("expected 50.5, got %f", *cfg.Coverage.RangeNm)
	}
	if *cfg.Coverage.BearingStartDegrees != 10 {
		t.Errorf("expected 10, got %f", *cfg.Coverage.BearingStartDegrees)
	}
	if *cfg.Coverage.BearingEndDegrees != 20 {
		t.Errorf("expected 20, got %f", *cfg.Coverage.BearingEndDegrees)
	}
	if cfg.Coverage.SensorPosition.Latitude != 45.0 {
		t.Errorf("expected 45.0, got %f", cfg.Coverage.SensorPosition.Latitude)
	}
}

func TestMustLoad_Panics(t *testing.T) {
	t.Run("invalid port", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for invalid port")
			}
		}()
		os.Setenv("GRPC_PORT", "abc")
		defer os.Unsetenv("GRPC_PORT")
		ingestion.MustLoad("test", "o", "d", 0)
	})

	t.Run("invalid classification", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for invalid classification")
			}
		}()
		os.Setenv("MAX_CLASSIFICATION", "INVALID")
		defer os.Unsetenv("MAX_CLASSIFICATION")
		ingestion.MustLoad("test", "o", "d", 0)
	})
}

func TestMustLoad_InvalidFloats(t *testing.T) {
	os.Setenv("SENSOR_RANGE_NM", "invalid")
	os.Setenv("SENSOR_LAT", "invalid")
	os.Setenv("SENSOR_LON", "45.0")
	os.Setenv("SENSOR_BEARING_START", "invalid")
	os.Setenv("SENSOR_BEARING_END", "invalid")
	defer os.Clearenv()

	cfg := ingestion.MustLoad("test-svc", "out", "dlq", 5000)
	if cfg.Coverage != nil {
		if cfg.Coverage.RangeNm != nil {
			t.Error("expected RangeNm to be nil for invalid input")
		}
		if cfg.Coverage.BearingStartDegrees != nil {
			t.Error("expected BearingStartDegrees to be nil for invalid input")
		}
		if cfg.Coverage.BearingEndDegrees != nil {
			t.Error("expected BearingEndDegrees to be nil for invalid input")
		}
		if cfg.Coverage.SensorPosition != nil {
			t.Error("expected SensorPosition to be nil for invalid LAT")
		}
	}
}
func TestMustLoad_PartialEnv(t *testing.T) {
	os.Setenv("SENSOR_LAT", "45.0")
	// SENSOR_LON is missing
	defer os.Clearenv()

	cfg := ingestion.MustLoad("test-svc", "out", "dlq", 5000)
	if cfg.Coverage != nil && cfg.Coverage.SensorPosition != nil {
		t.Error("expected SensorPosition to be nil when LON is missing")
	}
}

func TestMustLoad_PartialBearing(t *testing.T) {
	os.Clearenv()
	os.Setenv("SENSOR_BEARING_START", "10")
	// SENSOR_BEARING_END is missing
	defer os.Clearenv()

	cfg := ingestion.MustLoad("test-svc", "out", "dlq", 5000)
	if cfg.Coverage == nil || cfg.Coverage.BearingStartDegrees == nil {
		t.Error("expected BearingStartDegrees to be set even if END is missing based on current impl")
	}
	if cfg.Coverage.BearingEndDegrees != nil {
		t.Error("expected BearingEndDegrees to be nil")
	}
}
