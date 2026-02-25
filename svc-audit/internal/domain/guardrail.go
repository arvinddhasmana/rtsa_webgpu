// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"context"
"fmt"
"time"

"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

// QueryGuardrail enforces safety limits on ClickHouse queries.
type QueryGuardrail struct {
MaxRangeDays  int
MaxResultRows int
TimeoutSec    int
}

// NewQueryGuardrail creates a new guardrail with the given limits.
func NewQueryGuardrail(maxRangeDays, maxResultRows, timeoutSec int) *QueryGuardrail {
return &QueryGuardrail{
MaxRangeDays:  maxRangeDays,
MaxResultRows: maxResultRows,
TimeoutSec:    timeoutSec,
}
}

// ValidateTimeRange ensures the query time range does not exceed MaxRangeDays.
// Returns INVALID_ARGUMENT if the range is too wide or end is before start.
func (g *QueryGuardrail) ValidateTimeRange(start, end *timestamppb.Timestamp) error {
if start == nil || end == nil {
return status.Error(codes.InvalidArgument, "guardrail: time_range start and end are required")
}
s := start.AsTime()
e := end.AsTime()
if e.Before(s) {
return status.Error(codes.InvalidArgument, "guardrail: time_range end must be after start")
}
maxDuration := time.Duration(g.MaxRangeDays) * 24 * time.Hour
if e.Sub(s) > maxDuration {
return status.Error(codes.InvalidArgument,
fmt.Sprintf("guardrail: time_range exceeds maximum of %d days", g.MaxRangeDays))
}
return nil
}

// EnforceRowLimit returns the effective row limit.
// If requested is 0 or exceeds MaxResultRows, MaxResultRows is returned.
func (g *QueryGuardrail) EnforceRowLimit(requested int) int {
if requested <= 0 || requested > g.MaxResultRows {
return g.MaxResultRows
}
return requested
}

// QueryContext returns a context with the configured query timeout applied.
func (g *QueryGuardrail) QueryContext(parent context.Context) (context.Context, context.CancelFunc) {
return context.WithTimeout(parent, time.Duration(g.TimeoutSec)*time.Second)
}
