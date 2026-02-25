// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for svc-track using testcontainers.
//
// These tests require Docker and spin up a real Redpanda broker to validate
// the full message flow: publish FusedTrack → consumer processes → cache updated.
//
// Run with: go test ./... -tags=integration -v -timeout=300s
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-UI-002, NFR-PERF-001
package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/consumer"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// redpandaContainer holds the testcontainers Redpanda instance.
type redpandaContainer struct {
	container testcontainers.Container
	brokers   []string
}

// startRedpanda starts a Redpanda container for the test suite.
func startRedpanda(ctx context.Context) (*redpandaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redpandadata/redpanda:v24.1.1",
		ExposedPorts: []string{"19092/tcp"},
		Cmd: []string{
			"redpanda",
			"start",
			"--overprovisioned",
			"--smp", "1",
			"--memory", "512M",
			"--reserve-memory", "0M",
			"--node-id", "0",
			"--check=false",
			"--kafka-addr", "PLAINTEXT://0.0.0.0:19092",
			"--advertise-kafka-addr", "PLAINTEXT://localhost:19092",
		},
		WaitingFor: wait.ForLog("Successfully started Redpanda").WithStartupTimeout(90 * time.Second),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("startRedpanda: GenericContainer: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("startRedpanda: Host: %w", err)
	}
	port, err := c.MappedPort(ctx, "19092")
	if err != nil {
		return nil, fmt.Errorf("startRedpanda: MappedPort: %w", err)
	}

	broker := fmt.Sprintf("%s:%s", host, port.Port())
	return &redpandaContainer{
		container: c,
		brokers:   []string{broker},
	}, nil
}

// TestFusedTrackConsumer_IntegrationFlow tests the full consumer → cache flow.
func TestFusedTrackConsumer_IntegrationFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Start Redpanda.
	rp, err := startRedpanda(ctx)
	if err != nil {
		t.Skipf("could not start Redpanda container (Docker may not be available): %v", err)
	}
	defer func() {
		if termErr := rp.container.Terminate(context.Background()); termErr != nil {
			t.Logf("container terminate error: %v", termErr)
		}
	}()

	t.Logf("Redpanda started at %v", rp.brokers)

	// Create producer to inject test messages.
	producer, err := kgo.NewClient(kgo.SeedBrokers(rp.brokers...))
	if err != nil {
		t.Fatalf("kgo.NewClient (producer): %v", err)
	}
	defer producer.Close()

	// Build test track.
	testTrack := &entityv1.FusedTrack{
		TrackId:         "integ-trk-001",
		EntityType:      commonv1.EntityType_ENTITY_TYPE_AIR,
		HostileClass:    commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
		Status:          commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		ConfidenceScore: 0.88,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		EstimatedPosition: &commonv1.Position{
			Latitude:  45.4215,
			Longitude: -75.6972,
		},
		UpdatedAt: timestamppb.New(time.Now()),
	}

	payload, err := proto.Marshal(testTrack)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}

	// Produce the test message.
	record := &kgo.Record{
		Topic: "tracks.fused.air",
		Key:   []byte(testTrack.TrackId),
		Value: payload,
	}
	results := producer.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		t.Fatalf("produce error: %v", err)
	}
	t.Log("test track produced to tracks.fused.air")

	// Create cache and consumer.
	cache := domain.NewTrackCache(100)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	c, err := consumer.NewFusedTrackConsumer(rp.brokers, "integ-test-group", cache, logger)
	if err != nil {
		t.Fatalf("NewFusedTrackConsumer: %v", err)
	}
	defer c.Close()

	// Run consumer until we see the track in the cache.
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	go func() {
		if runErr := c.Run(consumerCtx); runErr != nil {
			t.Logf("consumer.Run error: %v", runErr)
		}
	}()

	// Poll the cache until the track appears or timeout.
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(30 * time.Second)

	for {
		select {
		case <-ticker.C:
			if got := cache.Get("integ-trk-001"); got != nil {
				consumerCancel()
				t.Logf("track found in cache: id=%s, status=%s", got.TrackId, got.Status)
				if got.TrackId != "integ-trk-001" {
					t.Errorf("expected track_id=integ-trk-001, got %q", got.TrackId)
				}
				if got.ConfidenceScore != 0.88 {
					t.Errorf("expected confidence=0.88, got %f", got.ConfidenceScore)
				}
				return
			}
		case <-deadline:
			consumerCancel()
			t.Fatal("timeout: track not found in cache after 30 seconds")
		case <-ctx.Done():
			consumerCancel()
			t.Fatal("context cancelled before track found")
		}
	}
}

// TestTrackCache_FilterIntegration tests cache + filter engine together.
func TestTrackCache_FilterIntegration(t *testing.T) {
	cache := domain.NewTrackCache(50)
	filter := &domain.FilterEngine{}

	// Insert mixed tracks.
	tracks := []*entityv1.FusedTrack{
		{
			TrackId: "fi-air-uncl", EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
			Status: commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			ConfidenceScore: 0.9,
			EstimatedPosition: &commonv1.Position{Latitude: 45.4, Longitude: -75.6},
			UpdatedAt: timestamppb.Now(),
		},
		{
			TrackId: "fi-surface-protb", EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
			Status: commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
			ConfidenceScore: 0.7,
			EstimatedPosition: &commonv1.Position{Latitude: 46.0, Longitude: -76.0},
			UpdatedAt: timestamppb.Now(),
		},
		{
			TrackId: "fi-cyber-secret", EntityType: commonv1.EntityType_ENTITY_TYPE_CYBER,
			Status: commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			ConfidenceScore: 0.95,
			UpdatedAt: timestamppb.Now(),
		},
	}
	for _, tr := range tracks {
		cache.Put(tr)
	}

	// Filter with PROTECTED_B clearance — should see unclassified and PROTECTED_B.
	f := &domain.TrackFilter{
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	}
	result := cache.GetFiltered(f)
	if len(result) != 2 {
		t.Errorf("expected 2 tracks with PROTECTED_B clearance, got %d", len(result))
	}

	// Filter by entity type AIR + bbox.
	f2 := &domain.TrackFilter{
		EntityTypes:    []commonv1.EntityType{commonv1.EntityType_ENTITY_TYPE_AIR},
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
		BoundingBox: &commonv1.BoundingBox{
			MinLatitude: 44.0, MaxLatitude: 47.0,
			MinLongitude: -77.0, MaxLongitude: -74.0,
		},
	}
	result2 := filter.Apply(cache.GetAll(), f2)
	if len(result2) != 1 {
		t.Errorf("expected 1 AIR track in bbox, got %d", len(result2))
	}

	t.Log("Filter integration test passed")
}
