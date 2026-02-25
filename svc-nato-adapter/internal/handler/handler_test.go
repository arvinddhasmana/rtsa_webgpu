// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
"context"
"strings"
"testing"

natov1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/nato/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-nato-adapter/internal/handler"
"go.uber.org/zap"
)

func TestNew(t *testing.T) {
h := handler.New(zap.NewNop())
if h == nil {
t.Fatal("New() returned nil")
}
}

func TestExportTracks(t *testing.T) {
tests := []struct {
name        string
req         *natov1.ExportTracksRequest
wantAccepted bool
wantNoop    bool
}{
{
name: "empty request",
req:  &natov1.ExportTracksRequest{},
wantAccepted: true,
wantNoop:    true,
},
{
name: "with track ids and destination",
req: &natov1.ExportTracksRequest{
TrackIds:           []string{"track-001", "track-002"},
DestinationPartner: "NATO_PARTNER_ALPHA",
},
wantAccepted: true,
wantNoop:    true,
},
{
name: "multiple tracks",
req: &natov1.ExportTracksRequest{
TrackIds: []string{"t1", "t2", "t3", "t4", "t5"},
},
wantAccepted: true,
wantNoop:    true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
h := handler.New(zap.NewNop())
resp, err := h.ExportTracks(context.Background(), tt.req)
if err != nil {
t.Fatalf("ExportTracks() error = %v", err)
}
if resp == nil {
t.Fatal("ExportTracks() returned nil response")
}
if resp.GetAccepted() != tt.wantAccepted {
t.Errorf("ExportTracks() Accepted = %v, want %v", resp.GetAccepted(), tt.wantAccepted)
}
if tt.wantNoop && !strings.Contains(resp.GetMessage(), "noop") {
t.Errorf("ExportTracks() Message = %q, want contains 'noop'", resp.GetMessage())
}
})
}
}

func TestImportTracks(t *testing.T) {
tests := []struct {
name              string
req               *natov1.ImportTracksRequest
wantTracksImported int32
wantNoop          bool
}{
{
name:              "empty request",
req:               &natov1.ImportTracksRequest{},
wantTracksImported: 0,
wantNoop:          true,
},
{
name: "with payload and source",
req: &natov1.ImportTracksRequest{
RawPayload:    []byte("STANAG5516_BINARY_BLOB"),
SourcePartner: "NATO_PARTNER_BRAVO",
},
wantTracksImported: 0,
wantNoop:          true,
},
{
name: "large payload",
req: &natov1.ImportTracksRequest{
RawPayload: make([]byte, 65536),
},
wantTracksImported: 0,
wantNoop:          true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
h := handler.New(zap.NewNop())
resp, err := h.ImportTracks(context.Background(), tt.req)
if err != nil {
t.Fatalf("ImportTracks() error = %v", err)
}
if resp == nil {
t.Fatal("ImportTracks() returned nil response")
}
if resp.GetTracksImported() != tt.wantTracksImported {
t.Errorf("ImportTracks() TracksImported = %v, want %v", resp.GetTracksImported(), tt.wantTracksImported)
}
if tt.wantNoop && !strings.Contains(resp.GetMessage(), "noop") {
t.Errorf("ImportTracks() Message = %q, want contains 'noop'", resp.GetMessage())
}
})
}
}
