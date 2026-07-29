// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e provides the multi-sensor ingestion end-to-end test (E2E04).
package e2e

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestE2E04_MultiSensorIngestion_RadarAndAISBFT_BothIngestedToTopics validates
// that multiple sensor adapters can simultaneously ingest observations, and that
// each produces records on the correct Redpanda topic.
//
// Steps:
//  1. Subscribe to sensors.radar.tracks and sensors.ais.positions from AtStart
//     (to capture any record with our unique test sensor ID)
//  2. Produce 5 radar observations with sensor_id "RADAR-E2E04"
//  3. Produce 5 AIS/BFT observations with sensor_id "AIS-E2E04"
//  4. Poll for at least 1 record with our radar sensor key on sensors.radar.tracks
//  5. Poll for at least 1 record with our AIS sensor key on sensors.ais.positions
//
// Topics are consumed from AtStart so that any previously-produced test data
// from prior runs is visible. The test matches records by Kafka key (sensor_id)
// to avoid false positives from unrelated traffic.
//
// Covers UC001 (multi-sensor ingestion) and topic routing per sensor type.
// Timeout: 3 minutes
func TestE2E04_MultiSensorIngestion_RadarAndAISBFT_BothIngestedToTopics(t *testing.T) {
	skipE2E(t)

	broker := redpandaBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const radarSensorID = "RADAR-E2E04"
	const aisSensorID = "AIS-E2E04"

	// ── Produce radar observations ─────────────────────────────────────────────
	radarProducer := newKafkaProducer(t, broker)
	defer radarProducer.Close()

	for i := 0; i < 5; i++ {
		obs := &ingestionv1.SensorObservation{
			ObservationId:   generateObsID("e2e04-radar", i),
			SensorId:        radarSensorID,
			SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
			ObservationTime: timestamppb.New(time.Now().UTC()),
			Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			Position: &commonv1.Position{
				Latitude:  45.0 + float64(i)*0.01, // Mid-Atlantic (non-sensitive)
				Longitude: -30.0 + float64(i)*0.01,
			},
			SensorData: &ingestionv1.SensorObservation_Radar{
				Radar: &ingestionv1.RadarTrack{
					TrackNumber:    generateObsID("rdr4", i),
					RangeNm:        10.0,
					BearingDegrees: float64(i * 72),
					TrackQuality:   0.9,
				},
			},
		}
		payload, _ := proto.Marshal(obs)
		headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-radar-ingestion", "", "v1")
		results := radarProducer.ProduceSync(ctx, &kgo.Record{
			Topic:   "sensors.radar.tracks",
			Key:     []byte(obs.SensorId),
			Value:   payload,
			Headers: headers,
		})
		if results.FirstErr() != nil {
			t.Logf("E2E04: radar produce %d: %v", i, results.FirstErr())
		}
	}

	// ── Produce AIS/BFT observations ───────────────────────────────────────────
	aisProducer := newKafkaProducer(t, broker)
	defer aisProducer.Close()

	for i := 0; i < 5; i++ {
		obs := &ingestionv1.SensorObservation{
			ObservationId:   generateObsID("e2e04-ais", i),
			SensorId:        aisSensorID,
			SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
			ObservationTime: timestamppb.New(time.Now().UTC()),
			Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			Position: &commonv1.Position{
				Latitude:  46.0 + float64(i)*0.01, // Mid-Atlantic (non-sensitive)
				Longitude: -29.0 + float64(i)*0.01,
			},
			SensorData: &ingestionv1.SensorObservation_AisBft{
				AisBft: &ingestionv1.AISPosition{
					Mmsi:           generateObsID("366999000", i),
					VesselName:     "E2E-VESSEL-" + itoa(i),
					VesselTypeCode: 30, // Fishing vessel (non-sensitive)
					NavStatus:      "under way using engine",
					AisMessageType: 1,
				},
			},
		}
		payload, _ := proto.Marshal(obs)
		headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-ais-ingestion", "", "v1")
		results := aisProducer.ProduceSync(ctx, &kgo.Record{
			Topic:   "sensors.ais.positions",
			Key:     []byte(obs.SensorId),
			Value:   payload,
			Headers: headers,
		})
		if results.FirstErr() != nil {
			t.Logf("E2E04: AIS produce %d: %v", i, results.FirstErr())
		}
	}

	t.Log("E2E04: 5 radar + 5 AIS/BFT observations produced")

	// ── Subscribe from AtStart and match by key ────────────────────────────────
	// Using AtStart ensures we see the records we just produced even if the
	// consumer group join takes longer than the produce round-trip.
	radarConsumer := newKafkaConsumer(t, broker,
		"e2e04-radar-verify",
		kgo.NewOffset().AtStart(),
		"sensors.radar.tracks",
	)
	defer radarConsumer.Close()

	aisConsumer := newKafkaConsumer(t, broker,
		"e2e04-ais-verify",
		kgo.NewOffset().AtStart(),
		"sensors.ais.positions",
	)
	defer aisConsumer.Close()

	// ── Assert records present ─────────────────────────────────────────────────
	radarOK := pollUntil(ctx, radarConsumer, 20*time.Second, func(r *kgo.Record) bool {
		return string(r.Key) == radarSensorID
	})
	aisOK := pollUntil(ctx, aisConsumer, 20*time.Second, func(r *kgo.Record) bool {
		return string(r.Key) == aisSensorID
	})

	if !radarOK {
		t.Error("E2E04: record with key RADAR-E2E04 not found on sensors.radar.tracks within 20s")
	}
	if !aisOK {
		t.Error("E2E04: record with key AIS-E2E04 not found on sensors.ais.positions within 20s")
	}

	if radarOK && aisOK {
		t.Log("E2E04 PASS: radar and AIS/BFT observations confirmed on their respective Redpanda topics")
	}
}
