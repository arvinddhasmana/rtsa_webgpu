// CLASSIFICATION: UNCLASSIFIED
package ingestion

import (
"fmt"
"os"
"strconv"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
)

// Config holds common ingestion service configuration.
type Config struct {
GRPCPort          int
RedpandaBrokers   string
OutputTopic       string
DLQTopic          string
ServiceName       string
MaxClassification commonv1.ClassificationLevel
LogLevel          string
}

// MustLoad loads config from environment variables, panics on missing required vars.
func MustLoad(serviceName, defaultOutputTopic, defaultDLQTopic string, defaultPort int) *Config {
cfg := &Config{
ServiceName:     serviceName,
OutputTopic:     getEnvOrDefault("OUTPUT_TOPIC", defaultOutputTopic),
DLQTopic:        getEnvOrDefault("DLQ_TOPIC", defaultDLQTopic),
RedpandaBrokers: getEnvOrDefault("REDPANDA_BROKERS", "localhost:9092"),
LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
}
portStr := getEnvOrDefault("GRPC_PORT", strconv.Itoa(defaultPort))
port, err := strconv.Atoi(portStr)
if err != nil {
panic(fmt.Sprintf("invalid GRPC_PORT: %s", portStr))
}
cfg.GRPCPort = port

maxClass := getEnvOrDefault("MAX_CLASSIFICATION", "CLASSIFICATION_LEVEL_SECRET")
level, ok := commonv1.ClassificationLevel_value[maxClass]
if !ok {
panic(fmt.Sprintf("invalid MAX_CLASSIFICATION: %s", maxClass))
}
cfg.MaxClassification = commonv1.ClassificationLevel(level)
return cfg
}

func getEnvOrDefault(key, def string) string {
if v := os.Getenv(key); v != "" {
return v
}
return def
}
