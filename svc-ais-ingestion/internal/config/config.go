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

// Config extends BaseConfig with AIS/BFT-specific settings.
type Config struct {
	pkgconfig.BaseConfig

	// Output topic for validated AIS position observations.
	OutputTopic string // RTSA_AIS_OUTPUT_TOPIC (default: "sensors.ais.positions")
	// DLQ topic for invalid messages.
	DLQTopic string // RTSA_AIS_DLQ_TOPIC (default: "dlq.sensors.ais")
	// Maximum speed jump in knots for spoofing detection.
	MaxSpeedJumpKnots float64 // RTSA_AIS_MAX_SPEED_JUMP_KNOTS (default: 50)
	// Optional static sensor coverage geometry.
	Coverage *ingestionv1.SensorCoverage
}

// MustLoad loads and validates AIS ingestion config.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config: failed to load ais config: %v", err))
	}
	return cfg
}

// Load reads AIS-specific configuration from environment variables.
func Load() (*Config, error) {
	base, err := pkgconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("config: base config: %w", err)
	}

	cfg := &Config{
		BaseConfig:        *base,
		OutputTopic:       getEnvOrDefault("RTSA_AIS_OUTPUT_TOPIC", "sensors.ais.positions"),
		DLQTopic:          getEnvOrDefault("RTSA_AIS_DLQ_TOPIC", "dlq.sensors.ais"),
		MaxSpeedJumpKnots: parseFloat("RTSA_AIS_MAX_SPEED_JUMP_KNOTS", 50.0),
	}

	// Parse coverage
	if rangeNmStr := os.Getenv("RTSA_AIS_RANGE_NM"); rangeNmStr != "" {
		if rangeNm, err := strconv.ParseFloat(rangeNmStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.RangeNm = &rangeNm
		}
	}
	if bStartStr := os.Getenv("RTSA_AIS_BEARING_START"); bStartStr != "" {
		if bStart, err := strconv.ParseFloat(bStartStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingStartDegrees = &bStart
		}
	}
	if bEndStr := os.Getenv("RTSA_AIS_BEARING_END"); bEndStr != "" {
		if bEnd, err := strconv.ParseFloat(bEndStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingEndDegrees = &bEnd
		}
	}
	if latStr := os.Getenv("RTSA_AIS_LAT"); latStr != "" {
		if lonStr := os.Getenv("RTSA_AIS_LON"); lonStr != "" {
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
