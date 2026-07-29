// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e provides end-to-end tests for the RTSA system.
//
// E2E tests require a fully running RTSA Docker Compose stack.
// Start with: docker compose -f deploy/docker-compose.yml \
//
//	-f deploy/docker-compose.services.yml up -d --build --wait
//
// Initialize with: ./scripts/dev/init-topics.sh && ./scripts/dev/init-clickhouse.sh
// Run with:        RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./...
// Use script:      ./scripts/dev/test-e2e.sh
package e2e

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestE2E01_FullPipeline_SensorToFusedTrack_ProducesFusedOutput validates the
// complete data flow from sensor ingestion to fused track output:
//
//	Simulator → Ingestion → Redpanda → Fusion → Redpanda (tracks.fused.surface)
//
// Steps:
//  1. Produce 10 radar observations (JSON-encoded as the fusion engine expects)
//     to sensors.radar.tracks
//  2. Wait up to 60s for at least 1 fused track on tracks.fused.surface
//  3. Fail if fusion engine did not process any tracks
//
// Covers UC001 (sensor ingestion) and UC002 (track fusion).
// Timeout: 5 minutes
func TestE2E01_FullPipeline_SensorToFusedTrack_ProducesFusedOutput(t *testing.T) {
	skipE2E(t)

	broker := redpandaBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	producer := newKafkaProducer(t, broker)
	defer producer.Close()

	// Produce 10 radar observations from 10 surface entities.
	// The fusion engine uses protojson.Unmarshal, so observations must be JSON-encoded.
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	for i := 0; i < 10; i++ {
		obs := &ingestionv1.SensorObservation{
			ObservationId:   generateObsID("e2e01-surface", i),
			SensorId:        "RADAR-E2E01",
			SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
			ObservationTime: timestamppb.New(time.Now().UTC()),
			Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			Position: &commonv1.Position{
				Latitude:  45.0 + float64(i)*0.01, // Mid-Atlantic (non-sensitive)
				Longitude: -30.0 + float64(i)*0.01,
			},
			SensorData: &ingestionv1.SensorObservation_Radar{
				Radar: &ingestionv1.RadarTrack{
					TrackNumber:    generateObsID("rdr", i),
					RangeNm:        5.0,
					BearingDegrees: float64(i * 36),
					TrackQuality:   0.8,
				},
			},
		}

		// JSON-encode for the fusion engine (which uses protojson.Unmarshal).
		payload, err := marshaler.Marshal(obs)
		if err != nil {
			t.Fatalf("E2E01: marshal observation %d: %v", i, err)
		}
		headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-radar-ingestion", "", "v1")

		results := producer.ProduceSync(ctx, &kgo.Record{
			Topic:   "sensors.radar.tracks",
			Key:     []byte(obs.SensorId),
			Value:   payload,
			Headers: headers,
		})
		if results.FirstErr() != nil {
			t.Logf("E2E01: produce observation %d: %v", i, results.FirstErr())
		}
	}

	t.Log("E2E01: 10 JSON-encoded radar observations produced to sensors.radar.tracks")

	// Wait for fused tracks on tracks.fused.surface.
	consumer := newKafkaConsumer(t, broker,
		"e2e01-fused-consumer",
		kgo.NewOffset().AtStart(),
		"tracks.fused.surface",
	)
	defer consumer.Close()

	var fusedCount int
	ok := pollUntil(ctx, consumer, 60*time.Second, func(_ *kgo.Record) bool {
		fusedCount++
		return fusedCount >= 1
	})
	if !ok {
		t.Fatalf("E2E01: timeout: 0 fused tracks received on tracks.fused.surface after 60s — is svc-fusion-engine running?")
	}

	t.Logf("E2E01 PASS: %d fused track(s) received on tracks.fused.surface", fusedCount)
}
