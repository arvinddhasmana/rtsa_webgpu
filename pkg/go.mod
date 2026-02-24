// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/pkg

go 1.22.0

require github.com/twmb/franz-go v1.17.0

require (
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
)

replace github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
