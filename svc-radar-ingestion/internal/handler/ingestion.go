// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/mapper"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/producer"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IngestionHandler implements IngestionService for radar data.
type IngestionHandler struct {
	ingestionv1.UnimplementedIngestionServiceServer

	validator    *domain.RadarValidator
	normalizer   *domain.RadarNormalizer
	enricher     *mapper.Enricher
	prod         *producer.ObservationProducer
	dlqProd      *producer.ObservationProducer
	auditEmitter *audit.Emitter
	logger       *zap.Logger
	coverage     *ingestionv1.SensorCoverage

	// Statistics (atomic for thread safety)
	totalReceived atomic.Int64
	totalAccepted atomic.Int64
	totalRejected atomic.Int64
	lastObsTime   atomic.Value // stores time.Time
	startTime     time.Time

	// Dynamic sensor tracking
	sensors sync.Map // map[string]*ingestion.SensorStateTracker
}


// NewIngestionHandler creates a new radar ingestion handler.
func NewIngestionHandler(
	validator *domain.RadarValidator,
	normalizer *domain.RadarNormalizer,
	enricher *mapper.Enricher,
	prod *producer.ObservationProducer,
	dlqProd *producer.ObservationProducer,
	auditEmitter *audit.Emitter,
	logger *zap.Logger,
	coverage *ingestionv1.SensorCoverage,
) *IngestionHandler {
	h := &IngestionHandler{
		validator:    validator,
		normalizer:   normalizer,
		enricher:     enricher,
		prod:         prod,
		dlqProd:      dlqProd,
		auditEmitter: auditEmitter,
		logger:       logger,
		coverage:     coverage,
		startTime:    time.Now(),
	}
	h.lastObsTime.Store(time.Time{})
	return h
}

// IngestSingleObservation handles unary radar observation ingestion.
func (h *IngestionHandler) IngestSingleObservation(ctx context.Context,
	obs *ingestionv1.SensorObservation) (*ingestionv1.IngestionAck, error) {

	// 1. Increment totalReceived counter
	h.totalReceived.Add(1)

	// Per-sensor tracker (replaces sensorState)
	t0 := time.Now()
	sID := obs.GetSensorId()
	newTracker := ingestion.NewSensorStateTracker()
	rawTracker, loaded := h.sensors.LoadOrStore(sID, newTracker)
	tracker := rawTracker.(*ingestion.SensorStateTracker)
	if !loaded {
		tracker.StartThroughputSampler(context.Background(), 30*time.Second)
	}

	// 2. Validate
	result := h.validator.Validate(obs)
	if !result.Valid {
		reason := "validation failed"
		if len(result.Errors) > 0 {
			reason = fmt.Sprintf("validation failed: %s", result.Errors[0].Message)
		}

		h.logger.Warn("observation rejected",
			zap.String("sensor_id", obs.GetSensorId()),
			zap.String("reason", reason))

		// Produce to DLQ
		if h.dlqProd != nil {
			if dlqErr := h.dlqProd.Produce(ctx, obs); dlqErr != nil {
				h.logger.Error("failed to produce to DLQ",
					zap.String("sensor_id", obs.GetSensorId()),
					zap.Error(dlqErr))
			}
		}

		tracker.RecordRejected(reason)
		h.totalRejected.Add(1)
		return &ingestionv1.IngestionAck{
			ObservationId:   obs.GetObservationId(),
			Accepted:        false,
			RejectionReason: reason,
		}, nil
	}

	// 3. Normalize
	normalized := h.normalizer.Normalize(obs)

// 4. Enrich (adds observation_id, checks classification ceiling)
if err := h.enricher.Enrich(ctx, normalized); err != nil {
h.logger.Warn("classification violation",
zap.String("sensor_id", obs.GetSensorId()),
zap.Error(err))
return nil, status.Errorf(codes.PermissionDenied,
"classification: data level %s exceeds service ceiling",
classification.LevelToString(obs.GetClassification()))
}

// 5. Produce to output topic
if h.prod != nil {
if err := h.prod.Produce(ctx, normalized); err != nil {
h.logger.Error("produce failed",
zap.String("sensor_id", obs.GetSensorId()),
zap.String("observation_id", normalized.ObservationId),
zap.Error(err))
return nil, status.Errorf(codes.Internal,
"ingestion: failed to produce observation %s: %v",
normalized.ObservationId, err)
}
}

// 6. Emit audit event
if h.auditEmitter != nil {
h.auditEmitter.Emit(ctx, audit.AuditParams{
EventType:           audit.EventObservationIngested,
ActorID:             obs.GetSensorId(),
ActorType:           auditv1.ActorType_ACTOR_TYPE_SERVICE,
ResourceType:        "observation",
ResourceID:          normalized.ObservationId,
Action:              auditv1.AuditAction_AUDIT_ACTION_INGEST,
ClassificationLevel: obs.GetClassification(),
})
}

	// 7. Increment totalAccepted
	tracker.RecordAccepted(time.Since(t0).Nanoseconds())
	h.totalAccepted.Add(1)
	h.lastObsTime.Store(time.Now().UTC())

	// 7.5 Extract and update coverage if present in metadata
	tracker.ExtractCoverage(obs.GetMetadata())

	h.lastObsTime.Store(time.Now().UTC())

h.logger.Info("observation accepted",
zap.String("sensor_id", obs.GetSensorId()),
zap.String("observation_id", normalized.ObservationId),
zap.String("sensor_type", "RADAR"),
zap.String("classification", classification.LevelToString(obs.GetClassification())))

// 8. Return ack
return &ingestionv1.IngestionAck{
ObservationId: normalized.ObservationId,
Accepted:      true,
}, nil
}

// IngestSensorData handles client-streaming radar observation ingestion.
// Does NOT fail the entire stream on individual observation failures.
func (h *IngestionHandler) IngestSensorData(
stream ingestionv1.IngestionService_IngestSensorDataServer) error {

ctx := stream.Context()
summary := &ingestionv1.IngestSummary{}
var rejections []*ingestionv1.RejectionDetail

for {
obs, err := stream.Recv()
if err != nil {
if err == io.EOF {
break
}
return status.Errorf(codes.Internal, "stream recv error: %v", err)
}

summary.TotalReceived++

ack, grpcErr := h.IngestSingleObservation(ctx, obs)
if grpcErr != nil {
// On gRPC-level error (e.g. classification violation): count as rejected
summary.Rejected++
rejections = append(rejections, &ingestionv1.RejectionDetail{
ObservationId: obs.GetObservationId(),
Reason:        grpcErr.Error(),
})
// Continue processing remaining observations
continue
}

if ack.GetAccepted() {
summary.Accepted++
} else {
summary.Rejected++
rejections = append(rejections, &ingestionv1.RejectionDetail{
ObservationId: ack.GetObservationId(),
Reason:        ack.GetRejectionReason(),
})
}
}

summary.Rejections = rejections
return stream.SendAndClose(summary)
}

// GetSensorStatus returns live statistics for the radar ingestion service.
func (h *IngestionHandler) GetSensorStatus(ctx context.Context,
req *ingestionv1.GetSensorStatusRequest) (*ingestionv1.SensorStatusResponse, error) {

lastTime := h.lastObsTime.Load().(time.Time)

resp := &ingestionv1.SensorStatusResponse{
	SensorId:      req.GetSensorId(),
	SensorType:    commonv1.SensorType_SENSOR_TYPE_RADAR,
	Connected:     true,
	TotalReceived: h.totalReceived.Load(),
	TotalAccepted: h.totalAccepted.Load(),
	TotalRejected: h.totalRejected.Load(),
}

	// Calculate throughput (obs/s)
	runtime := time.Since(h.startTime).Seconds()
	if runtime > 0 {
		resp.EventsPerSecond = float64(h.totalReceived.Load()) / runtime
	}

	if !lastTime.IsZero() {
		resp.LastObservationTime = timestamppb.New(lastTime)
	}
	if h.coverage != nil {
		resp.Coverage = h.coverage
	}

	return resp, nil
}

// ListSensorStatuses returns a list of all active radar sensor statistics.
func (h *IngestionHandler) ListSensorStatuses(ctx context.Context, req *ingestionv1.ListSensorStatusesRequest) (*ingestionv1.ListSensorStatusesResponse, error) {
	var sensors []*ingestionv1.SensorStatusResponse

	h.sensors.Range(func(key, value interface{}) bool {
		sID := key.(string)
		tracker := value.(*ingestion.SensorStateTracker)
		lastTime := tracker.LastObsTime()
		if req.ActiveWithinSeconds > 0 && !lastTime.IsZero() {
			if time.Since(lastTime).Seconds() > float64(req.ActiveWithinSeconds) {
				return true
			}
		}
		sensors = append(sensors, &ingestionv1.SensorStatusResponse{
			SensorId:            sID,
			SensorType:          commonv1.SensorType_SENSOR_TYPE_RADAR,
			Connected:           tracker.Connected(),
			TotalReceived:       tracker.TotalReceived(),
			TotalAccepted:       tracker.TotalAccepted(),
			TotalRejected:       tracker.TotalRejected(),
			EventsPerSecond:     tracker.EventsPerSecond(),
			LastObservationTime: timestamppb.New(lastTime),
			Coverage:            tracker.Coverage(),
		})
		return true
	})

	if len(sensors) == 0 && req.ActiveWithinSeconds == 0 {
		resp, _ := h.GetSensorStatus(ctx, &ingestionv1.GetSensorStatusRequest{SensorId: "radar-cluster-01"})
		if resp != nil {
			sensors = append(sensors, resp)
		}
	}

	return &ingestionv1.ListSensorStatusesResponse{Sensors: sensors}, nil
}

// GetSensorDiagnostic returns deep diagnostic data for a specific radar sensor.
func (h *IngestionHandler) GetSensorDiagnostic(ctx context.Context, req *ingestionv1.GetSensorDiagnosticRequest) (*ingestionv1.SensorDiagnosticResponse, error) {
	sID := req.GetSensorId()
	if sID == "" {
		return nil, status.Error(codes.InvalidArgument, "sensor_id is required")
	}
	raw, ok := h.sensors.Load(sID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "sensor %q not found or has not ingested data", sID)
	}
	tracker := raw.(*ingestion.SensorStateTracker)

	historySamples := req.GetHistorySamples()
	if historySamples <= 0 || historySamples > 60 {
		historySamples = 20
	}
	eventsLimit := req.GetRecentEventsLimit()
	if eventsLimit <= 0 || eventsLimit > 100 {
		eventsLimit = 20
	}

	resp := &ingestionv1.SensorDiagnosticResponse{
		SensorId:           sID,
		SensorType:         commonv1.SensorType_SENSOR_TYPE_RADAR,
		Connected:          tracker.Connected(),
		TotalReceived:      tracker.TotalReceived(),
		TotalAccepted:      tracker.TotalAccepted(),
		TotalRejected:      tracker.TotalRejected(),
		EventsPerSecond:    tracker.EventsPerSecond(),
		LatencyMs:          tracker.LatencyMs(),
		ValidationPassRate: tracker.ValidationPassRate(),
		ThroughputHistory:  tracker.SnapshotThroughput(int(historySamples)),
		DlqBreakdown:       tracker.DLQBreakdown(),
		RecentEvents:       tracker.SnapshotEvents(int(eventsLimit)),
		Coverage:           tracker.Coverage(),
	}
	if t := tracker.LastObsTime(); !t.IsZero() {
		resp.LastObservationTime = timestamppb.New(t)
	}
	if h.coverage != nil {
		resp.Coverage = h.coverage
	}
	return resp, nil
}
