// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/tests/benchmark

go 1.24.0

toolchain go1.24.12

require (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
	github.com/arvinddhasmana/RTSA_VS_Opus/svc-fusion-engine v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.67.0 // indirect
)

replace (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../../gen/go
	github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../../pkg
	github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection => ../../svc-anomaly-detection
	github.com/arvinddhasmana/RTSA_VS_Opus/svc-fusion-engine => ../../svc-fusion-engine
)
