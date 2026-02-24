// CLASSIFICATION: UNCLASSIFIED
module github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion

go 1.22.0

require (
github.com/arvinddhasmana/RTSA_VS_Opus/gen/go v0.0.0
github.com/arvinddhasmana/RTSA_VS_Opus/pkg v0.0.0
)

replace (
github.com/arvinddhasmana/RTSA_VS_Opus/gen/go => ../gen/go
github.com/arvinddhasmana/RTSA_VS_Opus/pkg => ../pkg
)
