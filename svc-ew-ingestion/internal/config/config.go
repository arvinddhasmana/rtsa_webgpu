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

// Config extends BaseConfig with EW/SIGINT-specific settings.
type Config struct {
	pkgconfig.BaseConfig

	// Output topic for validated EW intercept observations.
	OutputTopic string // RTSA_EW_OUTPUT_TOPIC (default: "sensors.ew.intercepts")
	// DLQ topic for invalid messages.
	DLQTopic string // RTSA_EW_DLQ_TOPIC (default: "dlq.sensors.ew")
	// Maximum accepted frequency in MHz.
	MaxFrequencyMHz float64 // RTSA_EW_MAX_FREQUENCY_MHZ (default: 40000)
	// Maximum accepted future time offset (seconds).
	MaxFutureOffsetSec int // RTSA_EW_MAX_FUTURE_OFFSET_SEC (default: 300)
	// Maximum accepted past time offset (seconds).
	MaxPastOffsetSec int // RTSA_EW_MAX_PAST_OFFSET_SEC (default: 86400)
	// Optional static sensor coverage geometry.
	Coverage *ingestionv1.SensorCoverage
}

// MustLoad loads and validates EW ingestion config.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config: failed to load ew config: %v", err))
	}
	return cfg
}

// Load reads EW-specific configuration from environment variables.
func Load() (*Config, error) {
	base, err := pkgconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("config: base config: %w", err)
	}

	cfg := &Config{
		BaseConfig:         *base,
		OutputTopic:        getEnvOrDefault("RTSA_EW_OUTPUT_TOPIC", "sensors.ew.intercepts"),
		DLQTopic:           getEnvOrDefault("RTSA_EW_DLQ_TOPIC", "dlq.sensors.ew"),
		MaxFrequencyMHz:    parseFloat("RTSA_EW_MAX_FREQUENCY_MHZ", 40000.0),
		MaxFutureOffsetSec: parseInt("RTSA_EW_MAX_FUTURE_OFFSET_SEC", 300),
		MaxPastOffsetSec:   parseInt("RTSA_EW_MAX_PAST_OFFSET_SEC", 86400),
	}

	// Parse coverage
	if rangeNmStr := os.Getenv("RTSA_EW_RANGE_NM"); rangeNmStr != "" {
		if rangeNm, err := strconv.ParseFloat(rangeNmStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.RangeNm = &rangeNm
		}
	}
	if bStartStr := os.Getenv("RTSA_EW_BEARING_START"); bStartStr != "" {
		if bStart, err := strconv.ParseFloat(bStartStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingStartDegrees = &bStart
		}
	}
	if bEndStr := os.Getenv("RTSA_EW_BEARING_END"); bEndStr != "" {
		if bEnd, err := strconv.ParseFloat(bEndStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingEndDegrees = &bEnd
		}
	}
	if latStr := os.Getenv("RTSA_EW_LAT"); latStr != "" {
		if lonStr := os.Getenv("RTSA_EW_LON"); lonStr != "" {
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

func parseInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		return defaultVal
	}
	return n
}
