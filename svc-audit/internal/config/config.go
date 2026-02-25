// CLASSIFICATION: UNCLASSIFIED
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for svc-audit.
// All values are sourced from environment variables with RTSA_ prefix.
type Config struct {
	GRPCPort           int
	ClickHouseDSN      string
	ClickHouseDatabase string
	RedpandaBrokers    []string
	ConsumerGroup      string
	BatchSize          int
	FlushPeriodSec     int
	MaxQueryRangeDays  int
	MaxResultRows      int
	QueryTimeoutSec    int
	DefaultPageSize    int
	ServiceName        string
	LogLevel           string
	LogFormat          string
	HealthPort         int
	MetricsPort        int
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort:           50073,
		ClickHouseDSN:      "clickhouse://default:@localhost:9000/rtsa",
		ClickHouseDatabase: "rtsa",
		RedpandaBrokers:    []string{"localhost:19092"},
		ConsumerGroup:      "svc-audit",
		BatchSize:          500,
		FlushPeriodSec:     2,
		MaxQueryRangeDays:  90,
		MaxResultRows:      10000,
		QueryTimeoutSec:    30,
		DefaultPageSize:    100,
		ServiceName:        "svc-audit",
		LogLevel:           "info",
		LogFormat:          "json",
		HealthPort:         8083,
		MetricsPort:        9093,
	}

	if v := os.Getenv("RTSA_CLICKHOUSE_DSN"); v != "" {
		cfg.ClickHouseDSN = v
	}
	if v := os.Getenv("RTSA_CLICKHOUSE_DATABASE"); v != "" {
		cfg.ClickHouseDatabase = v
	}
	if v := os.Getenv("RTSA_REDPANDA_BROKERS"); v != "" {
		cfg.RedpandaBrokers = strings.Split(v, ",")
	}
	if v := os.Getenv("RTSA_AUDIT_CONSUMER_GROUP"); v != "" {
		cfg.ConsumerGroup = v
	}
	if v := os.Getenv("RTSA_AUDIT_BATCH_SIZE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid RTSA_AUDIT_BATCH_SIZE %q: %w", v, err)
		}
		cfg.BatchSize = n
	}
	if v := os.Getenv("RTSA_AUDIT_FLUSH_PERIOD_SEC"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid RTSA_AUDIT_FLUSH_PERIOD_SEC %q: %w", v, err)
		}
		cfg.FlushPeriodSec = n
	}
	if v := os.Getenv("RTSA_AUDIT_MAX_RANGE_DAYS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid RTSA_AUDIT_MAX_RANGE_DAYS %q: %w", v, err)
		}
		cfg.MaxQueryRangeDays = n
	}
	if v := os.Getenv("RTSA_SERVICE_NAME"); v != "" {
		cfg.ServiceName = v
	}
	if v := os.Getenv("RTSA_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("RTSA_LOG_FORMAT"); v != "" {
		cfg.LogFormat = v
	}
	if v := os.Getenv("RTSA_HEALTH_PORT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid RTSA_HEALTH_PORT %q: %w", v, err)
		}
		cfg.HealthPort = n
	}
	if v := os.Getenv("RTSA_AUDIT_GRPC_PORT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid RTSA_AUDIT_GRPC_PORT %q: %w", v, err)
		}
		cfg.GRPCPort = n
	}
	if v := os.Getenv("RTSA_AUDIT_QUERY_TIMEOUT_SEC"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid RTSA_AUDIT_QUERY_TIMEOUT_SEC %q: %w", v, err)
		}
		cfg.QueryTimeoutSec = n
	}
	if v := os.Getenv("RTSA_AUDIT_DEFAULT_PAGE_SIZE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: invalid RTSA_AUDIT_DEFAULT_PAGE_SIZE %q: %w", v, err)
		}
		cfg.DefaultPageSize = n
	}

	if len(cfg.RedpandaBrokers) == 0 {
		return nil, fmt.Errorf("config: RTSA_REDPANDA_BROKERS must not be empty")
	}

	return cfg, nil
}
