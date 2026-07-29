// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func validTimeRange() *commonv1.TimeRange {
	now := time.Now()
	return &commonv1.TimeRange{
		StartTime: timestamppb.New(now.Add(-1 * time.Hour)),
		EndTime:   timestamppb.New(now),
	}
}
