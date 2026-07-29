// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/rtsa_webgpu/tests/e2e

go 1.24.0

toolchain go1.24.12

require (
	github.com/arvinddhasmana/rtsa_webgpu/gen/go v0.0.0
	github.com/arvinddhasmana/rtsa_webgpu/pkg v0.0.0-00010101000000-000000000000
	github.com/twmb/franz-go v1.17.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)

replace (
	github.com/arvinddhasmana/rtsa_webgpu/gen/go => ../../gen/go
	github.com/arvinddhasmana/rtsa_webgpu/pkg => ../../pkg
)
