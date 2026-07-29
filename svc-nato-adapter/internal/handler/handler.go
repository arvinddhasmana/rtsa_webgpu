// CLASSIFICATION: UNCLASSIFIED
// Package handler provides the noop gRPC handler for NatoAdapterService.
//
// Feature: FEAT-15 NATO Interoperability
// UC: UC011 NATO Adapter
// Requirements: CR-NATO-001, CR-NATO-002, CR-NATO-003, CR-NATO-004, CR-NATO-005
package handler

import (
"context"

natov1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/nato/v1"
"go.uber.org/zap"
)

// NatoAdapterHandler is a noop implementation of NatoAdapterServiceServer.
type NatoAdapterHandler struct {
natov1.UnimplementedNatoAdapterServiceServer
logger *zap.Logger
}

// New creates a new NatoAdapterHandler.
func New(logger *zap.Logger) *NatoAdapterHandler {
return &NatoAdapterHandler{logger: logger}
}

// ExportTracks is a noop implementation that logs the request and returns success.
func (h *NatoAdapterHandler) ExportTracks(
ctx context.Context,
req *natov1.ExportTracksRequest,
) (*natov1.ExportTracksResponse, error) {
h.logger.Info("noop ExportTracks called",
zap.Int("track_count", len(req.GetTrackIds())),
zap.String("destination", req.GetDestinationPartner()),
)
return &natov1.ExportTracksResponse{
Accepted: true,
Message:  "noop — export acknowledged but not transmitted",
}, nil
}

// ImportTracks is a noop implementation that logs the request and returns success.
func (h *NatoAdapterHandler) ImportTracks(
ctx context.Context,
req *natov1.ImportTracksRequest,
) (*natov1.ImportTracksResponse, error) {
h.logger.Info("noop ImportTracks called",
zap.Int("payload_bytes", len(req.GetRawPayload())),
zap.String("source", req.GetSourcePartner()),
)
return &natov1.ImportTracksResponse{
TracksImported: 0,
Message:        "noop — import acknowledged but not processed",
}, nil
}
