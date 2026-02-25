// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/tests/e2e

go 1.22.0

require (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
	github.com/arvinddhasmana/RTSA_VS_Opus/pkg v0.0.0-00010101000000-000000000000
	github.com/twmb/franz-go v1.17.0
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.64.0 // indirect
)

replace (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../../gen/go
	github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../../pkg
)
