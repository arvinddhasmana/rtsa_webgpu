// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"context"
	"testing"

	natov1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/nato/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-nato-adapter/internal/handler"
	"go.uber.org/zap"
)

func TestHandler_ExportTracks(t *testing.T) {
	logger := zap.NewNop()
	h := handler.New(logger)
	req := &natov1.ExportTracksRequest{
		TrackIds:           []string{"TRK-1", "TRK-2"},
		DestinationPartner: "NATO-SOUTH",
	}
	resp, err := h.ExportTracks(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Accepted {
		t.Error("expected accepted=true")
	}
}

func TestHandler_ImportTracks(t *testing.T) {
	logger := zap.NewNop()
	h := handler.New(logger)
	req := &natov1.ImportTracksRequest{
		SourcePartner: "NATO-NORTH",
		RawPayload:    []byte("fake-payload"),
	}
	resp, err := h.ImportTracks(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TracksImported != 0 {
		t.Errorf("expected 0 imported, got %d", resp.TracksImported)
	}
}
