// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/rtsa_webgpu/tests/benchmark

go 1.24.0

toolchain go1.24.12

require (
	github.com/arvinddhasmana/rtsa_webgpu/gen/go v0.0.0
	google.golang.org/protobuf v1.36.10
)

require (
	go.opentelemetry.io/otel v1.41.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.79.1 // indirect
)

replace github.com/arvinddhasmana/rtsa_webgpu/gen/go => ../../gen/go
