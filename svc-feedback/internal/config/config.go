// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"
)

// Config holds all runtime configuration for svc-feedback.
// Values are sourced exclusively from environment variables (RTSA_ prefix).
// No defaults embed secrets or connection credentials.
type Config struct {
GRPCPort          string
HealthPort        string
RedpandaBrokers   []string
RateLimitPerMin   int
LogLevel          string
ServiceName       string
}

// Load reads configuration from environment variables.
// Returns an error if any required value is invalid.
func Load() (*Config, error) {
cfg := &Config{
GRPCPort:        getEnv("RTSA_GRPC_PORT", "50051"),
HealthPort:      getEnv("RTSA_HEALTH_PORT", "8081"),
LogLevel:        getEnv("RTSA_LOG_LEVEL", "info"),
ServiceName:     getEnv("RTSA_SERVICE_NAME", "svc-feedback"),
}

// Parse broker list
brokerStr := getEnv("RTSA_REDPANDA_BROKERS", "localhost:19092")
cfg.RedpandaBrokers = splitBrokers(brokerStr)

// Parse rate limit
rateStr := getEnv("RTSA_RATE_LIMIT_PER_MINUTE", "10")
rate, err := strconv.Atoi(rateStr)
if err != nil {
return nil, fmt.Errorf("[config.Load]: invalid RTSA_RATE_LIMIT_PER_MINUTE %q: %w", rateStr, err)
}
if rate <= 0 {
return nil, fmt.Errorf("[config.Load]: RTSA_RATE_LIMIT_PER_MINUTE must be > 0, got %d", rate)
}
cfg.RateLimitPerMin = rate

return cfg, nil
}

// getEnv returns the value of the named environment variable,
// or the provided default if the variable is unset or empty.
func getEnv(key, defaultVal string) string {
if v := os.Getenv(key); v != "" {
return v
}
return defaultVal
}

// splitBrokers splits a comma-separated broker string into individual addresses.
func splitBrokers(s string) []string {
parts := strings.Split(s, ",")
out := make([]string, 0, len(parts))
for _, p := range parts {
p = strings.TrimSpace(p)
if p != "" {
out = append(out, p)
}
}
if len(out) == 0 {
out = []string{"localhost:19092"}
}
return out
}
