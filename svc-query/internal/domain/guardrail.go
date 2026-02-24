// CLASSIFICATION: UNCLASSIFIED

// Package domain contains the core business rules for the query service,
// independent of any infrastructure or transport concerns.
package domain

import (
"context"
"fmt"
"time"

"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

// QueryGuardrail enforces safety limits on all ClickHouse queries.
// It prevents runaway queries from exhausting ClickHouse resources.
type QueryGuardrail struct {
// MaxRangeDays is the maximum allowed query time range.
MaxRangeDays int
// MaxResultRows is the hard ceiling on returned rows per query.
MaxResultRows int
// TimeoutSec is the per-query execution timeout.
TimeoutSec int
// DefaultPageSize is the page size used when the client does not specify one.
DefaultPageSize int
// MaxPageSize is the maximum allowed page size per request.
MaxPageSize int
}

// NewQueryGuardrail creates a guardrail with the given limits.
// All parameters must be positive.
func NewQueryGuardrail(maxRangeDays, maxResultRows, timeoutSec, defaultPageSize, maxPageSize int) (*QueryGuardrail, error) {
if maxRangeDays <= 0 {
return nil, fmt.Errorf("domain.NewQueryGuardrail: maxRangeDays must be positive, got %d", maxRangeDays)
}
if maxResultRows <= 0 {
return nil, fmt.Errorf("domain.NewQueryGuardrail: maxResultRows must be positive, got %d", maxResultRows)
}
if timeoutSec <= 0 {
return nil, fmt.Errorf("domain.NewQueryGuardrail: timeoutSec must be positive, got %d", timeoutSec)
}
if defaultPageSize <= 0 {
return nil, fmt.Errorf("domain.NewQueryGuardrail: defaultPageSize must be positive, got %d", defaultPageSize)
}
if maxPageSize <= 0 {
return nil, fmt.Errorf("domain.NewQueryGuardrail: maxPageSize must be positive, got %d", maxPageSize)
}
return &QueryGuardrail{
MaxRangeDays:    maxRangeDays,
MaxResultRows:   maxResultRows,
TimeoutSec:      timeoutSec,
DefaultPageSize: defaultPageSize,
MaxPageSize:     maxPageSize,
}, nil
}

// ValidateTimeRange checks that the requested time window is valid and within limits.
// Returns a gRPC INVALID_ARGUMENT status error on failure so it can be returned
// directly from handlers.
//
// Rules:
//   - start and end must both be present
//   - end must be after start
//   - range must not exceed MaxRangeDays
func (g *QueryGuardrail) ValidateTimeRange(start, end *timestamppb.Timestamp) error {
if start == nil {
return status.Error(codes.InvalidArgument, "time_range.start_time is required")
}
if end == nil {
return status.Error(codes.InvalidArgument, "time_range.end_time is required")
}

startT := start.AsTime()
endT := end.AsTime()

if !endT.After(startT) {
return status.Errorf(codes.InvalidArgument,
"time_range.end_time must be after start_time (start=%s, end=%s)",
startT.Format(time.RFC3339), endT.Format(time.RFC3339))
}

rangeDays := endT.Sub(startT).Hours() / 24
if rangeDays > float64(g.MaxRangeDays) {
return status.Errorf(codes.InvalidArgument,
"time range %.1f days exceeds maximum allowed %d days",
rangeDays, g.MaxRangeDays)
}

return nil
}

// EnforceRowLimit returns the effective row limit for a query.
// If requestedLimit is 0 or exceeds MaxResultRows, MaxResultRows is returned.
func (g *QueryGuardrail) EnforceRowLimit(requestedLimit int) int {
if requestedLimit <= 0 || requestedLimit > g.MaxResultRows {
return g.MaxResultRows
}
return requestedLimit
}

// EnforcePageSize returns the effective page size for a paginated request.
// If requestedSize is 0 or exceeds MaxPageSize, DefaultPageSize or MaxPageSize is returned.
func (g *QueryGuardrail) EnforcePageSize(requestedSize int) int {
if requestedSize <= 0 {
return g.DefaultPageSize
}
if requestedSize > g.MaxPageSize {
return g.MaxPageSize
}
return requestedSize
}

// QueryContext returns a context with the configured query timeout applied.
// The caller MUST call the returned CancelFunc when the query completes.
func (g *QueryGuardrail) QueryContext(parent context.Context) (context.Context, context.CancelFunc) {
return context.WithTimeout(parent, time.Duration(g.TimeoutSec)*time.Second)
}
