// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"context"
"testing"
"time"

"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewQueryGuardrail_valid(t *testing.T) {
g, err := NewQueryGuardrail(30, 100000, 30, 100, 1000)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if g.MaxRangeDays != 30 {
t.Errorf("MaxRangeDays = %d, want 30", g.MaxRangeDays)
}
if g.DefaultPageSize != 100 {
t.Errorf("DefaultPageSize = %d, want 100", g.DefaultPageSize)
}
}

func TestNewQueryGuardrail_invalidParams(t *testing.T) {
tests := []struct {
name           string
rangeDays      int
maxRows        int
timeoutSec     int
defaultPageSz  int
maxPageSz      int
}{
{"zero rangeDays", 0, 100000, 30, 100, 1000},
{"negative rangeDays", -1, 100000, 30, 100, 1000},
{"zero maxRows", 30, 0, 30, 100, 1000},
{"zero timeout", 30, 100000, 0, 100, 1000},
{"zero defaultPageSize", 30, 100000, 30, 0, 1000},
{"zero maxPageSize", 30, 100000, 30, 100, 0},
}
for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
_, err := NewQueryGuardrail(tc.rangeDays, tc.maxRows, tc.timeoutSec, tc.defaultPageSz, tc.maxPageSz)
if err == nil {
t.Fatal("expected error, got nil")
}
})
}
}

func TestValidateTimeRange(t *testing.T) {
g, _ := NewQueryGuardrail(30, 100000, 30, 100, 1000)

now := time.Now().UTC()
dayAgo := now.Add(-24 * time.Hour)
future := now.Add(time.Hour)

tests := []struct {
name     string
start    *timestamppb.Timestamp
end      *timestamppb.Timestamp
wantCode codes.Code
}{
{
name:     "valid range",
start:    timestamppb.New(dayAgo),
end:      timestamppb.New(now),
wantCode: codes.OK,
},
{
name:     "nil start",
start:    nil,
end:      timestamppb.New(now),
wantCode: codes.InvalidArgument,
},
{
name:     "nil end",
start:    timestamppb.New(dayAgo),
end:      nil,
wantCode: codes.InvalidArgument,
},
{
name:     "end equals start",
start:    timestamppb.New(now),
end:      timestamppb.New(now),
wantCode: codes.InvalidArgument,
},
{
name:     "end before start",
start:    timestamppb.New(now),
end:      timestamppb.New(dayAgo),
wantCode: codes.InvalidArgument,
},
{
name:     "range exceeds 30 days",
start:    timestamppb.New(now.Add(-31 * 24 * time.Hour)),
end:      timestamppb.New(now),
wantCode: codes.InvalidArgument,
},
{
name:     "range exactly 30 days",
start:    timestamppb.New(now.Add(-30*24*time.Hour + time.Second)),
end:      timestamppb.New(now),
wantCode: codes.OK,
},
{
name:     "future end allowed",
start:    timestamppb.New(now),
end:      timestamppb.New(future),
wantCode: codes.OK,
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
err := g.ValidateTimeRange(tc.start, tc.end)
if tc.wantCode == codes.OK {
if err != nil {
t.Errorf("unexpected error: %v", err)
}
return
}
if err == nil {
t.Fatal("expected error, got nil")
}
st, ok := status.FromError(err)
if !ok {
t.Fatalf("expected gRPC status error, got: %v", err)
}
if st.Code() != tc.wantCode {
t.Errorf("code = %v, want %v", st.Code(), tc.wantCode)
}
})
}
}

func TestEnforceRowLimit(t *testing.T) {
g, _ := NewQueryGuardrail(30, 1000, 30, 100, 500)

tests := []struct {
name      string
requested int
want      int
}{
{"zero to max", 0, 1000},
{"negative to max", -1, 1000},
{"exceeds max to max", 9999, 1000},
{"below max", 500, 500},
{"exactly max", 1000, 1000},
}
for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
got := g.EnforceRowLimit(tc.requested)
if got != tc.want {
t.Errorf("EnforceRowLimit(%d) = %d, want %d", tc.requested, got, tc.want)
}
})
}
}

func TestEnforcePageSize(t *testing.T) {
g, _ := NewQueryGuardrail(30, 100000, 30, 100, 500)

tests := []struct {
name      string
requested int
want      int
}{
{"zero uses default", 0, 100},
{"negative uses default", -1, 100},
{"exceeds max uses max", 9999, 500},
{"below max returned as-is", 200, 200},
{"exactly max returned", 500, 500},
}
for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
got := g.EnforcePageSize(tc.requested)
if got != tc.want {
t.Errorf("EnforcePageSize(%d) = %d, want %d", tc.requested, got, tc.want)
}
})
}
}

func TestQueryContext(t *testing.T) {
g, _ := NewQueryGuardrail(30, 1000, 1, 100, 500) // 1 second timeout

ctx, cancel := g.QueryContext(context.Background())
defer cancel()

deadline, ok := ctx.Deadline()
if !ok {
t.Fatal("expected deadline to be set")
}
if time.Until(deadline) > 2*time.Second {
t.Errorf("deadline too far in future: %v", time.Until(deadline))
}
}
