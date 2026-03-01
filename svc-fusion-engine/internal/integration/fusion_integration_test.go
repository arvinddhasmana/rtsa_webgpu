// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration_test contains integration tests for the fusion engine.
// These tests require a running Redpanda instance and are guarded by the
// RTSA_INTEGRATION_TESTS=true environment variable.
package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func skipUnlessIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
		t.Skip("integration tests disabled (set RTSA_INTEGRATION_TESTS=true)")
	}
}

func brokers() []string {
	b := os.Getenv("RTSA_REDPANDA_BROKERS")
	if b == "" {
		return []string{"localhost:19092"}
	}
	return []string{b}
}

func makeRadarObs(sensorID string, lat, lon float64) *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        sensorID,
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(time.Now()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		Position: &commonv1.Position{
			Latitude:  lat,
			Longitude: lon,
		},
	}
}

// IT01 — Produce radar observation → expect a FusedTrack on tracks.fused.surface
func TestIT01_ProduceRadarObs_ConsumeFusedTrack(t *testing.T) {
	skipUnlessIntegration(t)

	b := brokers()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Produce a radar observation to sensors.radar
	producer, err := kgo.NewClient(kgo.SeedBrokers(b...))
	if err != nil {
		t.Fatalf("create producer: %v", err)
	}
	defer producer.Close()

	obs := makeRadarObs("RADAR-INTEG-01", 45.0, -60.0)
	payload, _ := protojson.MarshalOptions{UseProtoNames: true}.Marshal(obs)
	producer.ProduceSync(ctx, &kgo.Record{
		Topic: "sensors.radar.tracks",
		Key:   []byte(obs.SensorId),
		Value: payload,
	})

	// Consume from tracks.fused.surface with a short timeout
	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(b...),
		kgo.ConsumerGroup("integ-test-it01"),
		kgo.ConsumeTopics("tracks.fused.surface"),
	)
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}
	defer consumer.Close()

	deadline := time.After(20 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timeout: no FusedTrack received on tracks.fused.surface")
		default:
		}
		fetches := consumer.PollRecords(ctx, 10)
		if fetches.NumRecords() > 0 {
			t.Logf("IT01 PASS: received FusedTrack on tracks.fused.surface")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
