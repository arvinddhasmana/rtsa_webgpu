// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"
)

// Config holds all runtime configuration for svc-query.
// All values are sourced from environment variables with RTSA_ prefix.
type Config struct {
GRPCPort           int
HealthPort         int
MetricsPort        int
ClickHouseDSN      string
ClickHouseDatabase string
MaxQueryRangeDays  int
MaxResultRows      int
QueryTimeoutSec    int
DefaultPageSize    int
MaxPageSize        int
RedpandaBrokers    []string
ServiceName        string
LogLevel           string
LogFormat          string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
cfg := &Config{
GRPCPort:           50072,
HealthPort:         8082,
MetricsPort:        9092,
ClickHouseDSN:      "clickhouse://default:@localhost:9000/rtsa",
ClickHouseDatabase: "rtsa",
MaxQueryRangeDays:  30,
MaxResultRows:      100000,
QueryTimeoutSec:    30,
DefaultPageSize:    100,
MaxPageSize:        1000,
RedpandaBrokers:    []string{"localhost:19092"},
ServiceName:        "svc-query",
LogLevel:           "info",
LogFormat:          "json",
}

if v := os.Getenv("RTSA_QUERY_GRPC_PORT"); v != "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_QUERY_GRPC_PORT %q: %w", v, err)
}
cfg.GRPCPort = p
}
if v := os.Getenv("RTSA_GRPC_PORT"); v != "" && os.Getenv("RTSA_QUERY_GRPC_PORT") == "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_GRPC_PORT %q: %w", v, err)
}
cfg.GRPCPort = p
}
if v := os.Getenv("RTSA_HEALTH_PORT"); v != "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_HEALTH_PORT %q: %w", v, err)
}
cfg.HealthPort = p
}
if v := os.Getenv("RTSA_METRICS_PORT"); v != "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_METRICS_PORT %q: %w", v, err)
}
cfg.MetricsPort = p
}
if v := os.Getenv("RTSA_CLICKHOUSE_DSN"); v != "" {
cfg.ClickHouseDSN = v
}
if v := os.Getenv("RTSA_CLICKHOUSE_DATABASE"); v != "" {
cfg.ClickHouseDatabase = v
}
if v := os.Getenv("RTSA_QUERY_MAX_RANGE_DAYS"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_QUERY_MAX_RANGE_DAYS %q: %w", v, err)
}
cfg.MaxQueryRangeDays = n
}
if v := os.Getenv("RTSA_QUERY_MAX_ROWS"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_QUERY_MAX_ROWS %q: %w", v, err)
}
cfg.MaxResultRows = n
}
if v := os.Getenv("RTSA_QUERY_TIMEOUT_SEC"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_QUERY_TIMEOUT_SEC %q: %w", v, err)
}
cfg.QueryTimeoutSec = n
}
if v := os.Getenv("RTSA_QUERY_DEFAULT_PAGE_SIZE"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_QUERY_DEFAULT_PAGE_SIZE %q: %w", v, err)
}
cfg.DefaultPageSize = n
}
if v := os.Getenv("RTSA_QUERY_MAX_PAGE_SIZE"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config: invalid RTSA_QUERY_MAX_PAGE_SIZE %q: %w", v, err)
}
cfg.MaxPageSize = n
}
if v := os.Getenv("RTSA_REDPANDA_BROKERS"); v != "" {
cfg.RedpandaBrokers = strings.Split(v, ",")
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

return cfg, nil
}
