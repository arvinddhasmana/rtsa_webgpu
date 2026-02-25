// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"
)

// Config holds all runtime configuration for svc-alert.
// All values are sourced from environment variables with RTSA_ prefix.
type Config struct {
// Server ports
GRPCPort    int
HealthPort  int
MetricsPort int

// Redpanda / Kafka configuration
RedpandaBrokers []string
ConsumerGroup   string
Topics          []string

// Priority queue
MaxQueueSize int

// TLS
TLSCACert     string
TLSServerCert string
TLSServerKey  string

// Observability
LogLevel    string
LogFormat   string
ServiceName string
OTELEndpoint string
}

// Load reads configuration from environment variables.
// Returns an error if any required value is invalid.
func Load() (*Config, error) {
cfg := &Config{
GRPCPort:        50051,
HealthPort:      8081,
MetricsPort:     9090,
RedpandaBrokers: []string{"localhost:19092"},
ConsumerGroup:   "svc-alert",
Topics: []string{
"alerts.anomaly.critical",
"alerts.anomaly.elevated",
"alerts.anomaly.watch",
},
MaxQueueSize:  10000,
TLSCACert:     "./certs/dev/ca.crt",
TLSServerCert: "./certs/dev/server.crt",
TLSServerKey:  "./certs/dev/server.key",
LogLevel:      "info",
LogFormat:     "json",
ServiceName:   "svc-alert",
OTELEndpoint:  "localhost:4317",
}

if v := os.Getenv("RTSA_GRPC_PORT"); v != "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("[config].[Load]: invalid RTSA_GRPC_PORT %q: %w", v, err)
}
cfg.GRPCPort = p
}

if v := os.Getenv("RTSA_HEALTH_PORT"); v != "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("[config].[Load]: invalid RTSA_HEALTH_PORT %q: %w", v, err)
}
cfg.HealthPort = p
}

if v := os.Getenv("RTSA_METRICS_PORT"); v != "" {
p, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("[config].[Load]: invalid RTSA_METRICS_PORT %q: %w", v, err)
}
cfg.MetricsPort = p
}

if v := os.Getenv("RTSA_REDPANDA_BROKERS"); v != "" {
cfg.RedpandaBrokers = strings.Split(v, ",")
}

if v := os.Getenv("RTSA_ALERT_CONSUMER_GROUP"); v != "" {
cfg.ConsumerGroup = v
}

if v := os.Getenv("RTSA_ALERT_TOPICS"); v != "" {
cfg.Topics = strings.Split(v, ",")
}

if v := os.Getenv("RTSA_ALERT_MAX_QUEUE_SIZE"); v != "" {
n, err := strconv.Atoi(v)
if err != nil {
return nil, fmt.Errorf("[config].[Load]: invalid RTSA_ALERT_MAX_QUEUE_SIZE %q: %w", v, err)
}
cfg.MaxQueueSize = n
}

if v := os.Getenv("RTSA_TLS_CA_CERT"); v != "" {
cfg.TLSCACert = v
}
if v := os.Getenv("RTSA_TLS_SERVER_CERT"); v != "" {
cfg.TLSServerCert = v
}
if v := os.Getenv("RTSA_TLS_SERVER_KEY"); v != "" {
cfg.TLSServerKey = v
}

if v := os.Getenv("RTSA_LOG_LEVEL"); v != "" {
cfg.LogLevel = v
}
if v := os.Getenv("RTSA_LOG_FORMAT"); v != "" {
cfg.LogFormat = v
}
if v := os.Getenv("RTSA_SERVICE_NAME"); v != "" {
cfg.ServiceName = v
}
if v := os.Getenv("RTSA_OTEL_ENDPOINT"); v != "" {
cfg.OTELEndpoint = v
}

if err := cfg.validate(); err != nil {
return nil, err
}

return cfg, nil
}

// validate ensures configuration values are sensible.
func (c *Config) validate() error {
if c.GRPCPort < 1 || c.GRPCPort > 65535 {
return fmt.Errorf("[config].[validate]: GRPCPort %d out of range", c.GRPCPort)
}
if c.HealthPort < 1 || c.HealthPort > 65535 {
return fmt.Errorf("[config].[validate]: HealthPort %d out of range", c.HealthPort)
}
if c.MetricsPort < 1 || c.MetricsPort > 65535 {
return fmt.Errorf("[config].[validate]: MetricsPort %d out of range", c.MetricsPort)
}
if len(c.RedpandaBrokers) == 0 {
return fmt.Errorf("[config].[validate]: RedpandaBrokers must not be empty")
}
if c.MaxQueueSize < 1 {
return fmt.Errorf("[config].[validate]: MaxQueueSize must be >= 1")
}
return nil
}
