// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion

go 1.24.0

toolchain go1.24.12

require (
github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
github.com/arvinddhasmana/RTSA_VS_Opus/pkg v0.0.0
github.com/google/uuid v1.6.0
go.opentelemetry.io/otel/trace v1.39.0
go.uber.org/zap v1.27.1
google.golang.org/grpc v1.67.0
google.golang.org/protobuf v1.34.2
)

require (
github.com/beorn7/perks v1.0.1 // indirect
github.com/cenkalti/backoff/v4 v4.3.0 // indirect
github.com/cespare/xxhash/v2 v2.3.0 // indirect
github.com/go-logr/logr v1.4.3 // indirect
github.com/go-logr/stdr v1.2.2 // indirect
github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
github.com/klauspost/compress v1.18.3 // indirect
github.com/pierrec/lz4/v4 v4.1.25 // indirect
github.com/prometheus/client_golang v1.19.0 // indirect
github.com/prometheus/client_model v0.6.0 // indirect
github.com/prometheus/common v0.48.0 // indirect
github.com/prometheus/procfs v0.12.0 // indirect
github.com/twmb/franz-go v1.17.0 // indirect
github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
go.opentelemetry.io/auto/sdk v1.2.1 // indirect
go.opentelemetry.io/otel v1.39.0 // indirect
go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0 // indirect
go.opentelemetry.io/otel/exporters/prometheus v0.46.0 // indirect
go.opentelemetry.io/otel/metric v1.39.0 // indirect
go.opentelemetry.io/otel/sdk v1.39.0 // indirect
go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
go.opentelemetry.io/proto/otlp v1.1.0 // indirect
go.uber.org/multierr v1.11.0 // indirect
golang.org/x/crypto v0.47.0 // indirect
golang.org/x/net v0.49.0 // indirect
golang.org/x/sys v0.40.0 // indirect
golang.org/x/text v0.33.0 // indirect
google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142 // indirect
google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
)

replace (
github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../pkg
)
