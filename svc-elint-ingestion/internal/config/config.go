// CLASSIFICATION: UNCLASSIFIED
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	pkgconfig "github.com/arvinddhasmana/RTSA_VS_Opus/pkg/config"
)

// Config extends BaseConfig with ELINT/COMINT-specific settings.
type Config struct {
	pkgconfig.BaseConfig

	// Output topic for validated ELINT detections.
	OutputTopic string // RTSA_ELINT_OUTPUT_TOPIC (default: "sensors.elint.detections")
	// DLQ topic for invalid messages.
	DLQTopic string // RTSA_ELINT_DLQ_TOPIC (default: "dlq.sensors.elint")
	// Maximum CEP (circular error probable) in meters.
	MaxCEPMeters float64 // RTSA_ELINT_MAX_CEP_METERS (default: 50000)
	// Optional static sensor coverage geometry.
	Coverage *ingestionv1.SensorCoverage
}

// MustLoad loads and validates ELINT ingestion config.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config: failed to load elint config: %v", err))
	}
	return cfg
}

// Load reads ELINT-specific configuration from environment variables.
func Load() (*Config, error) {
	base, err := pkgconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("config: base config: %w", err)
	}

	cfg := &Config{
		BaseConfig:   *base,
		OutputTopic:  getEnvOrDefault("RTSA_ELINT_OUTPUT_TOPIC", "sensors.elint.detections"),
		DLQTopic:     getEnvOrDefault("RTSA_ELINT_DLQ_TOPIC", "dlq.sensors.elint"),
		MaxCEPMeters: parseFloat("RTSA_ELINT_MAX_CEP_METERS", 50000.0),
	}

	// Parse coverage
	if rangeNmStr := os.Getenv("RTSA_ELINT_RANGE_NM"); rangeNmStr != "" {
		if rangeNm, err := strconv.ParseFloat(rangeNmStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.RangeNm = &rangeNm
		}
	}
	if bStartStr := os.Getenv("RTSA_ELINT_BEARING_START"); bStartStr != "" {
		if bStart, err := strconv.ParseFloat(bStartStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingStartDegrees = &bStart
		}
	}
	if bEndStr := os.Getenv("RTSA_ELINT_BEARING_END"); bEndStr != "" {
		if bEnd, err := strconv.ParseFloat(bEndStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingEndDegrees = &bEnd
		}
	}
	if latStr := os.Getenv("RTSA_ELINT_LAT"); latStr != "" {
		if lonStr := os.Getenv("RTSA_ELINT_LON"); lonStr != "" {
			lat, er1 := strconv.ParseFloat(latStr, 64)
			lon, er2 := strconv.ParseFloat(lonStr, 64)
			if er1 == nil && er2 == nil {
				if cfg.Coverage == nil {
					cfg.Coverage = &ingestionv1.SensorCoverage{}
				}
				cfg.Coverage.SensorPosition = &commonv1.Position{
					Latitude:  lat,
					Longitude: lon,
				}
			}
		}
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func parseFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil {
		return defaultVal
	}
	return f
}
