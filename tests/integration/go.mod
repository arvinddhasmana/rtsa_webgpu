// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/tests/integration

go 1.22.0

require (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0-00010101000000-000000000000
	github.com/arvinddhasmana/RTSA_VS_Opus/svc-track v0.0.0-00010101000000-000000000000
	github.com/testcontainers/testcontainers-go v0.31.0
	github.com/twmb/franz-go v1.17.0
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.33.0
)

replace (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../../gen/go
	github.com/arvinddhasmana/RTSA_VS_Opus/svc-track => ../../svc-track
)
