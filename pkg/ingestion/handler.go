// CLASSIFICATION: UNCLASSIFIED
package ingestion

import (
"context"
"fmt"
"sync"
"sync/atomic"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/google/uuid"
"go.uber.org/zap"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

// sensorStats holds per-sensor statistics.
type sensorStats struct {
totalReceived int64
totalAccepted int64
totalRejected int64
lastSeen      time.Time
mu            sync.Mutex
}

// Handler implements ingestionv1.IngestionServiceServer.
type Handler struct {
ingestionv1.UnimplementedIngestionServiceServer
validator   Validator
normalizer  Normalizer
producer    Producer
dlqProducer Producer
logger      *zap.Logger
cfg         *Config
stats       sync.Map // map[string]*sensorStats
}

// NewHandler creates a new gRPC ingestion handler.
func NewHandler(v Validator, n Normalizer, p Producer, dlq Producer, logger *zap.Logger, cfg *Config) *Handler {
return &Handler{
validator:   v,
normalizer:  n,
producer:    p,
dlqProducer: dlq,
logger:      logger,
cfg:         cfg,
}
}

// IngestSingleObservation handles a single observation ingestion request.
func (h *Handler) IngestSingleObservation(ctx context.Context, obs *ingestionv1.SensorObservation) (*ingestionv1.IngestionAck, error) {
if obs == nil {
return nil, status.Error(codes.InvalidArgument, "observation must not be nil")
}

// Check classification ceiling
if obs.GetClassification() > h.cfg.MaxClassification {
return nil, status.Errorf(codes.PermissionDenied,
"classification: data level %s exceeds service ceiling %s",
obs.GetClassification().String(), h.cfg.MaxClassification.String())
}

// Assign observation ID
if obs.ObservationId == "" {
obs.ObservationId = uuid.New().String()
}

// Normalize
h.normalizer.Normalize(obs)

// Validate
result := h.validator.Validate(obs)

h.updateStats(obs.GetSensorId(), result.Valid)

if !result.Valid {
reason := "validation failed"
if len(result.Errors) > 0 {
reason = fmt.Sprintf("validation failed: %s", result.Errors[0].Message)
}
h.logger.Warn("observation rejected",
zap.String("service", h.cfg.ServiceName),
zap.String("sensor_id", obs.GetSensorId()),
zap.String("observation_id", obs.GetObservationId()),
zap.String("reason", reason),
)
if err := h.dlqProducer.Produce(ctx, obs); err != nil {
h.logger.Error("dlq produce failed", zap.Error(err))
}
return &ingestionv1.IngestionAck{
ObservationId:   obs.GetObservationId(),
Accepted:        false,
RejectionReason: reason,
}, nil
}

// Produce to output topic
if err := h.producer.Produce(ctx, obs); err != nil {
return nil, status.Errorf(codes.Internal,
"ingestion: failed to produce observation %s: %v", obs.GetObservationId(), err)
}

h.logger.Info("observation accepted",
zap.String("service", h.cfg.ServiceName),
zap.String("sensor_id", obs.GetSensorId()),
zap.String("observation_id", obs.GetObservationId()),
)

return &ingestionv1.IngestionAck{
ObservationId: obs.GetObservationId(),
Accepted:      true,
}, nil
}

// IngestSensorData handles client-streaming ingestion.
func (h *Handler) IngestSensorData(stream ingestionv1.IngestionService_IngestSensorDataServer) error {
var totalReceived, accepted, rejected int64
var rejections []*ingestionv1.RejectionDetail

for {
obs, err := stream.Recv()
if err != nil {
break
}
totalReceived++

ack, err := h.IngestSingleObservation(stream.Context(), obs)
if err != nil {
// Classification or internal error — stop stream
return err
}
if ack.GetAccepted() {
accepted++
} else {
rejected++
rejections = append(rejections, &ingestionv1.RejectionDetail{
ObservationId: ack.GetObservationId(),
Reason:        ack.GetRejectionReason(),
})
}
}

return stream.SendAndClose(&ingestionv1.IngestSummary{
TotalReceived: totalReceived,
Accepted:      accepted,
Rejected:      rejected,
Rejections:    rejections,
})
}

// GetSensorStatus returns statistics for a sensor.
func (h *Handler) GetSensorStatus(ctx context.Context, req *ingestionv1.GetSensorStatusRequest) (*ingestionv1.SensorStatusResponse, error) {
	sensorID := req.GetSensorId()
	v, ok := h.stats.Load(sensorID)
	if !ok {
		return &ingestionv1.SensorStatusResponse{
			SensorId:  sensorID,
			Connected: false,
		}, nil
	}
	s := v.(*sensorStats)
	s.mu.Lock()
	lastSeen := s.lastSeen
	s.mu.Unlock()
	resp := &ingestionv1.SensorStatusResponse{
		SensorId:            sensorID,
		SensorType:          commonv1.SensorType(0),
		Connected:           true,
		TotalReceived:       atomic.LoadInt64(&s.totalReceived),
		TotalAccepted:       atomic.LoadInt64(&s.totalAccepted),
		TotalRejected:       atomic.LoadInt64(&s.totalRejected),
		LastObservationTime: timestamppb.New(lastSeen),
	}

	if h.cfg.Coverage != nil {
		resp.Coverage = h.cfg.Coverage
	}

	return resp, nil
}

func (h *Handler) updateStats(sensorID string, accepted bool) {
	v, _ := h.stats.LoadOrStore(sensorID, &sensorStats{})
	s := v.(*sensorStats)
	atomic.AddInt64(&s.totalReceived, 1)
	if accepted {
		atomic.AddInt64(&s.totalAccepted, 1)
	} else {
		atomic.AddInt64(&s.totalRejected, 1)
	}
	s.mu.Lock()
	s.lastSeen = time.Now()
	s.mu.Unlock()
}

// ListSensorStatuses returns a list of all active sensor statistics.
func (h *Handler) ListSensorStatuses(ctx context.Context, req *ingestionv1.ListSensorStatusesRequest) (*ingestionv1.ListSensorStatusesResponse, error) {
	var sensors []*ingestionv1.SensorStatusResponse
	now := time.Now()

	h.stats.Range(func(key, value any) bool {
		sensorID := key.(string)
		s := value.(*sensorStats)

		s.mu.Lock()
		lastSeen := s.lastSeen
		s.mu.Unlock()

		// Apply active_within_seconds filter if specified
		if req.ActiveWithinSeconds > 0 {
			if now.Sub(lastSeen).Seconds() > float64(req.ActiveWithinSeconds) {
				return true // Continue iteration
			}
		}

		statusResponse := &ingestionv1.SensorStatusResponse{
			SensorId:            sensorID,
			SensorType:          commonv1.SensorType(0), // Would map from config
			Connected:           true,
			TotalReceived:       atomic.LoadInt64(&s.totalReceived),
			TotalAccepted:       atomic.LoadInt64(&s.totalAccepted),
			TotalRejected:       atomic.LoadInt64(&s.totalRejected),
			LastObservationTime: timestamppb.New(lastSeen),
		}

		if h.cfg.Coverage != nil {
			statusResponse.Coverage = h.cfg.Coverage
		}

		sensors = append(sensors, statusResponse)
		return true
	})

	return &ingestionv1.ListSensorStatusesResponse{
		Sensors: sensors,
	}, nil
}
