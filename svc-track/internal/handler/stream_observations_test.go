// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"
	"log/slog"
	"os"
	"testing"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"google.golang.org/grpc/metadata"
)

// mockSensorStream implements ServerStreamingServer[SensorObservationUpdate]
type mockSensorStream struct {
	ctx    context.Context
	sent   []*entityv1.SensorObservationUpdate
	sendFn func(*entityv1.SensorObservationUpdate) error
}

func (m *mockSensorStream) Send(u *entityv1.SensorObservationUpdate) error {
	if m.sendFn != nil {
		return m.sendFn(u)
	}
	m.sent = append(m.sent, u)
	return nil
}

func (m *mockSensorStream) Context() context.Context     { return m.ctx }
func (m *mockSensorStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockSensorStream) SendHeader(metadata.MD) error { return nil }
func (m *mockSensorStream) SetTrailer(metadata.MD)       {}
func (m *mockSensorStream) RecvMsg(_ interface{}) error  { return nil }
func (m *mockSensorStream) SendMsg(v interface{}) error {
	if u, ok := v.(*entityv1.SensorObservationUpdate); ok {
		return m.Send(u)
	}
	return nil
}

func testSensorLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

type mockSubscriber struct {
	subFn   func() (uint64, <-chan *ingestionv1.SensorObservation)
	unsubFn func(uint64)
}

func (m *mockSubscriber) Subscribe() (uint64, <-chan *ingestionv1.SensorObservation) {
	if m.subFn != nil {
		return m.subFn()
	}
	return 0, nil
}

func (m *mockSubscriber) Unsubscribe(id uint64) {
	if m.unsubFn != nil {
		m.unsubFn(id)
	}
}

func TestSensorStreamHandler_StreamSensorObservations(t *testing.T) {
	logger := testSensorLogger()

	tests := []struct {
		name         string
		req          *entityv1.StreamSensorObservationsRequest
		observations []*ingestionv1.SensorObservation
		expected     int // Number of observations expected to pass filters
	}{
		{
			name: "No filters, all pass",
			req: &entityv1.StreamSensorObservationsRequest{
				ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			},
			observations: []*ingestionv1.SensorObservation{
				{ObservationId: "1", Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
				{ObservationId: "2", Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET},
			},
			expected: 2,
		},
		{
			name: "Clearance level filter blocks SECRET",
			req: &entityv1.StreamSensorObservationsRequest{
				ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			},
			observations: []*ingestionv1.SensorObservation{
				{ObservationId: "1", Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
				{ObservationId: "2", Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET},
			},
			expected: 1,
		},
		{
			name: "Sensor type filter allows only RADAR",
			req: &entityv1.StreamSensorObservationsRequest{
				ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
				SensorTypes:    []commonv1.SensorType{commonv1.SensorType_SENSOR_TYPE_RADAR},
			},
			observations: []*ingestionv1.SensorObservation{
				{ObservationId: "1", SensorType: commonv1.SensorType_SENSOR_TYPE_RADAR},
				{ObservationId: "2", SensorType: commonv1.SensorType_SENSOR_TYPE_EW_SIGINT},
			},
			expected: 1,
		},
		{
			name: "Bounding box filter excludes out-of-bounds",
			req: &entityv1.StreamSensorObservationsRequest{
				ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
				BoundingBox: &commonv1.BoundingBox{
					MinLatitude:  40.0,
					MaxLatitude:  50.0,
					MinLongitude: -80.0,
					MaxLongitude: -70.0,
				},
			},
			observations: []*ingestionv1.SensorObservation{
				{
					ObservationId: "1", // Inside
					Position:      &commonv1.Position{Latitude: 45.0, Longitude: -75.0},
				},
				{
					ObservationId: "2", // Outside Latitude
					Position:      &commonv1.Position{Latitude: 55.0, Longitude: -75.0},
				},
				{
					ObservationId: "3", // No Position
				},
			},
			expected: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obsCh := make(chan *ingestionv1.SensorObservation, len(tc.observations))
			for _, obs := range tc.observations {
				obsCh <- obs
			}
			close(obsCh) // Important: close to unblock the handler loop after processing

			sub := &mockSubscriber{
				subFn: func() (uint64, <-chan *ingestionv1.SensorObservation) {
					return 1, obsCh
				},
			}

			h := NewSensorStreamHandler(sub, logger)
			stream := &mockSensorStream{ctx: context.Background()}

			// Run stream handler
			err := h.StreamSensorObservations(tc.req, stream)
			if err != nil {
				t.Fatalf("StreamSensorObservations error: %v", err)
			}

			if len(stream.sent) != tc.expected {
				t.Errorf("expected %d sent observations, got %d", tc.expected, len(stream.sent))
			}
		})
	}
}
