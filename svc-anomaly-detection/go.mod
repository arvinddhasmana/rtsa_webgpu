// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection

go 1.24.0

toolchain go1.24.12

replace (
	github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
	github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../pkg
)

require (
	github.com/google/uuid v1.6.0
	github.com/twmb/franz-go v1.17.0
	go.uber.org/zap v1.27.1
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
)
