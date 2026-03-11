// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert

go 1.24.0

toolchain go1.24.12

require (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
	github.com/arvinddhasmana/RTSA_VS_Opus/pkg v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.19.0
	github.com/twmb/franz-go v1.17.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.67.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
	go.opentelemetry.io/otel v1.41.0 // indirect
	go.opentelemetry.io/otel/trace v1.41.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
)

replace (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
	github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../pkg
)
