// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/rtsa_webgpu/tools/simulator

go 1.24.0

toolchain go1.24.12

require (
	github.com/arvinddhasmana/rtsa_webgpu/gen/go v0.0.0
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.10
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	go.opentelemetry.io/otel v1.41.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace (
	github.com/arvinddhasmana/rtsa_webgpu/gen/go => ../../gen/go
	github.com/arvinddhasmana/rtsa_webgpu/pkg => ../../pkg
)
