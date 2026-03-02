// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"
	"log/slog"
	"os"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testMetrics() *metrics.Metrics {
	reg := prometheus.NewRegistry()
	return metrics.New(reg)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// makeTrack builds a FusedTrack for handler tests.
func makeTrack(id string, status commonv1.TrackStatus, et commonv1.EntityType, cls commonv1.ClassificationLevel, confidence float64) *entityv1.FusedTrack {
	return &entityv1.FusedTrack{
		TrackId:         id,
		Status:          status,
		EntityType:      et,
		Classification:  cls,
		ConfidenceScore: confidence,
		EstimatedPosition: &commonv1.Position{
			Latitude:  45.4215,
			Longitude: -75.6972,
		},
		UpdatedAt: timestamppb.New(time.Now()),
	}
}

// mockServerStream simulates a gRPC server stream for testing StreamTracks.
type mockServerStream struct {
	sent   []*entityv1.TrackUpdate
	ctx    context.Context
	sendFn func(*entityv1.TrackUpdate) error
}

func (m *mockServerStream) Send(u *entityv1.TrackUpdate) error {
	if m.sendFn != nil {
		return m.sendFn(u)
	}
	m.sent = append(m.sent, u)
	return nil
}

func (m *mockServerStream) Context() context.Context     { return m.ctx }
func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) RecvMsg(_ interface{}) error  { return nil }
func (m *mockServerStream) SendMsg(v interface{}) error {
	if u, ok := v.(*entityv1.TrackUpdate); ok {
		return m.Send(u)
	}
	return nil
}
