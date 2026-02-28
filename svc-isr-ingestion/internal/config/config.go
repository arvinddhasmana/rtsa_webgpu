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

// Config extends BaseConfig with ISR-specific settings.
type Config struct {
	pkgconfig.BaseConfig

	// Output topic for validated ISR observations.
	OutputTopic string // RTSA_ISR_OUTPUT_TOPIC (default: "sensors.isr.observations")
	// DLQ topic for invalid messages.
	DLQTopic string // RTSA_ISR_DLQ_TOPIC (default: "dlq.sensors.isr")
	// Minimum polygon vertices required.
	MinPolygonVertices int // RTSA_ISR_MIN_POLYGON_VERTS (default: 3)
	// Optional static sensor coverage geometry.
	Coverage *ingestionv1.SensorCoverage
}

// MustLoad loads and validates ISR ingestion config.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("config: failed to load isr config: %v", err))
	}
	return cfg
}

// Load reads ISR-specific configuration from environment variables.
func Load() (*Config, error) {
	base, err := pkgconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("config: base config: %w", err)
	}

	cfg := &Config{
		BaseConfig:         *base,
		OutputTopic:        getEnvOrDefault("RTSA_ISR_OUTPUT_TOPIC", "sensors.isr.observations"),
		DLQTopic:           getEnvOrDefault("RTSA_ISR_DLQ_TOPIC", "dlq.sensors.isr"),
		MinPolygonVertices: parseInt("RTSA_ISR_MIN_POLYGON_VERTS", 3),
	}

	// Parse coverage
	if rangeNmStr := os.Getenv("RTSA_ISR_RANGE_NM"); rangeNmStr != "" {
		if rangeNm, err := strconv.ParseFloat(rangeNmStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.RangeNm = &rangeNm
		}
	}
	if bStartStr := os.Getenv("RTSA_ISR_BEARING_START"); bStartStr != "" {
		if bStart, err := strconv.ParseFloat(bStartStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingStartDegrees = &bStart
		}
	}
	if bEndStr := os.Getenv("RTSA_ISR_BEARING_END"); bEndStr != "" {
		if bEnd, err := strconv.ParseFloat(bEndStr, 64); err == nil {
			if cfg.Coverage == nil {
				cfg.Coverage = &ingestionv1.SensorCoverage{}
			}
			cfg.Coverage.BearingEndDegrees = &bEnd
		}
	}
	if latStr := os.Getenv("RTSA_ISR_LAT"); latStr != "" {
		if lonStr := os.Getenv("RTSA_ISR_LON"); lonStr != "" {
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

func getEnvOrDefault(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
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
