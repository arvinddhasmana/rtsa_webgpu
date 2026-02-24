// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/pkg

go 1.22.0

require (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
	github.com/google/uuid v1.6.0
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.33.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
)

replace github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
