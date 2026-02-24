// CLASSIFICATION: UNCLASSIFIED

// Package config loads and validates svc-query configuration from environment variables.
// No hardcoded secrets or connection strings.
package config

import (
"fmt"
"os"
"strconv"
)

// Config holds all runtime configuration for the query service.
// Every field is populated from an environment variable with a safe default.
// Required fields return an error when missing.
type Config struct {
// GRPCPort is the port for the gRPC server.
GRPCPort int
// ClickHouseDSN is the ClickHouse connection string.
// Format: clickhouse://user:password@host:port/database
// Required.
ClickHouseDSN string
// ClickHouseDatabase is the ClickHouse database name.
ClickHouseDatabase string
// MaxQueryRangeDays is the maximum allowed time range in days for any query.
MaxQueryRangeDays int
// MaxResultRows is the maximum number of rows returned per query.
MaxResultRows int
// QueryTimeoutSec is the per-query execution timeout in seconds.
QueryTimeoutSec int
// DefaultPageSize is the page size when not specified in the request.
DefaultPageSize int
// MaxPageSize is the maximum allowed page size.
MaxPageSize int
// RedpandaBrokers is a comma-separated list of Redpanda broker addresses.
// Required.
RedpandaBrokers string
// ServiceName is the service identifier for telemetry and audit events.
ServiceName string
// OTelEndpoint is the OpenTelemetry collector gRPC endpoint.
OTelEndpoint string
// TLSCertFile is the path to the server TLS certificate for gRPC.
// Required.
TLSCertFile string
// TLSKeyFile is the path to the server TLS private key for gRPC.
// Required.
TLSKeyFile string
// TLSCAFile is the path to the CA certificate for mTLS client verification.
// Required.
TLSCAFile string
// ClickHouseTLSCertFile is the client cert for ClickHouse mTLS (optional).
ClickHouseTLSCertFile string
// ClickHouseTLSKeyFile is the client key for ClickHouse mTLS (optional).
ClickHouseTLSKeyFile string
// ClickHouseTLSCAFile is the CA cert for ClickHouse TLS (optional).
ClickHouseTLSCAFile string
// LogLevel is the minimum log level: debug, info, warn, error.
LogLevel string
}

// Load reads configuration from environment variables.
// Returns an error if any required variable is absent or invalid.
func Load() (*Config, error) {
cfg := &Config{
GRPCPort:           50072,
ClickHouseDatabase: "rtsa",
MaxQueryRangeDays:  30,
MaxResultRows:      100_000,
QueryTimeoutSec:    30,
DefaultPageSize:    100,
MaxPageSize:        1000,
ServiceName:        "svc-query",
OTelEndpoint:       "otel-collector:4317",
LogLevel:           "info",
}

if v := os.Getenv("RTSA_QUERY_GRPC_PORT"); v != "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config.Load: RTSA_QUERY_GRPC_PORT invalid: %w", err)
}
cfg.GRPCPort = p
}

cfg.ClickHouseDSN = os.Getenv("RTSA_CLICKHOUSE_DSN")
if cfg.ClickHouseDSN == "" {
return nil, fmt.Errorf("config.Load: RTSA_CLICKHOUSE_DSN is required")
}

if v := os.Getenv("RTSA_CLICKHOUSE_DATABASE"); v != "" {
cfg.ClickHouseDatabase = v
}

cfg.RedpandaBrokers = os.Getenv("RTSA_REDPANDA_BROKERS")
if cfg.RedpandaBrokers == "" {
return nil, fmt.Errorf("config.Load: RTSA_REDPANDA_BROKERS is required")
}

cfg.TLSCertFile = os.Getenv("RTSA_TLS_SERVER_CERT")
if cfg.TLSCertFile == "" {
return nil, fmt.Errorf("config.Load: RTSA_TLS_SERVER_CERT is required")
}

cfg.TLSKeyFile = os.Getenv("RTSA_TLS_SERVER_KEY")
if cfg.TLSKeyFile == "" {
return nil, fmt.Errorf("config.Load: RTSA_TLS_SERVER_KEY is required")
}

cfg.TLSCAFile = os.Getenv("RTSA_TLS_CA_CERT")
if cfg.TLSCAFile == "" {
return nil, fmt.Errorf("config.Load: RTSA_TLS_CA_CERT is required")
}

// ClickHouse TLS — optional
cfg.ClickHouseTLSCertFile = os.Getenv("RTSA_CLICKHOUSE_TLS_CERT")
cfg.ClickHouseTLSKeyFile = os.Getenv("RTSA_CLICKHOUSE_TLS_KEY")
cfg.ClickHouseTLSCAFile = os.Getenv("RTSA_CLICKHOUSE_TLS_CA")

if v := os.Getenv("RTSA_SERVICE_NAME"); v != "" {
cfg.ServiceName = v
}
if v := os.Getenv("RTSA_OTEL_ENDPOINT"); v != "" {
cfg.OTelEndpoint = v
}
if v := os.Getenv("RTSA_LOG_LEVEL"); v != "" {
cfg.LogLevel = v
}
if v := os.Getenv("RTSA_QUERY_MAX_RANGE_DAYS"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config.Load: RTSA_QUERY_MAX_RANGE_DAYS invalid: %w", err)
}
cfg.MaxQueryRangeDays = n
}
if v := os.Getenv("RTSA_QUERY_MAX_ROWS"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config.Load: RTSA_QUERY_MAX_ROWS invalid: %w", err)
}
cfg.MaxResultRows = n
}
if v := os.Getenv("RTSA_QUERY_TIMEOUT_SEC"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config.Load: RTSA_QUERY_TIMEOUT_SEC invalid: %w", err)
}
cfg.QueryTimeoutSec = n
}
if v := os.Getenv("RTSA_QUERY_DEFAULT_PAGE_SIZE"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config.Load: RTSA_QUERY_DEFAULT_PAGE_SIZE invalid: %w", err)
}
cfg.DefaultPageSize = n
}
if v := os.Getenv("RTSA_QUERY_MAX_PAGE_SIZE"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("config.Load: RTSA_QUERY_MAX_PAGE_SIZE invalid: %w", err)
}
cfg.MaxPageSize = n
}

return cfg, nil
}
