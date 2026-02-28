// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"log/slog"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"google.golang.org/grpc"
)

type SensorSubscriber interface {
	Subscribe() (uint64, <-chan *ingestionv1.SensorObservation)
	Unsubscribe(id uint64)
}

type SensorStreamHandler struct {
	subscriber SensorSubscriber
	logger     *slog.Logger
}

func NewSensorStreamHandler(sub SensorSubscriber, logger *slog.Logger) *SensorStreamHandler {
	return &SensorStreamHandler{
		subscriber: sub,
		logger:     logger,
	}
}

func (h *SensorStreamHandler) StreamSensorObservations(
	req *entityv1.StreamSensorObservationsRequest,
	stream grpc.ServerStreamingServer[entityv1.SensorObservationUpdate],
) error {
	id, ch := h.subscriber.Subscribe()
	defer h.subscriber.Unsubscribe(id)

	sensorTypeMap := make(map[commonv1.SensorType]bool)
	for _, st := range req.SensorTypes {
		sensorTypeMap[st] = true
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case obs, ok := <-ch:
			if !ok {
				return nil
			}

			if obs.Classification > req.ClearanceLevel {
				continue
			}

			if len(sensorTypeMap) > 0 && !sensorTypeMap[obs.SensorType] {
				continue
			}

			if req.BoundingBox != nil {
				if obs.Position == nil {
					continue
				}
				p := obs.Position
				bb := req.BoundingBox
				if p.Latitude < bb.MinLatitude || p.Latitude > bb.MaxLatitude ||
					p.Longitude < bb.MinLongitude || p.Longitude > bb.MaxLongitude {
					continue
				}
			}

			update := &entityv1.SensorObservationUpdate{
				Observation: obs,
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}
