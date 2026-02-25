// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit

go 1.24.1

toolchain go1.24.12

replace (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
	github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../pkg
)

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.43.0
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
	github.com/twmb/franz-go v1.17.0
	go.uber.org/zap v1.27.1
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/ClickHouse/ch-go v0.71.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.64.0 // indirect
)
