// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for the training pipeline (IT15):
// model candidate publish → models.anomaly.published Redpanda topic validation.
//
// The training pipeline publishes JSON-encoded ModelCandidate messages to the
// models.anomaly.published topic.  This test validates message production,
// topic routing, and header propagation via the shared Redpanda container.
package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"github.com/arvinddhasmana/rtsa_webgpu/tests/integration/testutil"
	"github.com/twmb/franz-go/pkg/kgo"
)

// modelCandidatePayload mirrors the JSON structure produced by svc-training.
// Kept local so the test has no compile-time dependency on svc-training.
type modelCandidatePayload struct {
	ModelID   string    `json:"model_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// TestIT15_TrainingModelPublish_CandidateOnTopic validates:
//  1. A model candidate JSON payload can be produced to models.anomaly.published
//  2. Classification and source-service headers are correctly attached
//  3. The message is consumable and deserialisable with all required fields
func TestIT15_TrainingModelPublish_CandidateOnTopic(t *testing.T) {
	testutil.SkipUnlessEnabled(t)

	env := testutil.SetupRedpandaOnly(t)
	defer env.Teardown()

	producer := env.NewKafkaProducer(t)
	ctx := context.Background()

	candidate := modelCandidatePayload{
		ModelID:   "noop-v1.0.0-it15",
		Status:    "candidate",
		Timestamp: time.Now().UTC(),
	}

	payload, err := json.Marshal(candidate)
	if err != nil {
		t.Fatalf("IT15: marshal model candidate: %v", err)
	}

	headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-training", "", "v1")

	r := producer.ProduceSync(ctx, &kgo.Record{
		Topic:   "models.anomaly.published",
		Key:     []byte(candidate.ModelID),
		Value:   payload,
		Headers: headers,
	})
	if r.FirstErr() != nil {
		t.Fatalf("IT15: produce model candidate: %v", r.FirstErr())
	}

	// Consume and verify.
	consumer := env.NewKafkaConsumer(t, "it15-training-group", "models.anomaly.published")
	received := testutil.WaitForTopicMessages(t, consumer, 1, 20*time.Second)
	if len(received) == 0 {
		t.Fatal("IT15: no messages on models.anomaly.published")
	}

	rec := received[0]

	// Verify headers.
	testutil.AssertHeaderPresent(t, rec, redpanda.HeaderClassification)
	testutil.AssertHeaderValue(t, rec, redpanda.HeaderSourceService, "svc-training")

	// Verify key == model_id.
	if string(rec.Key) != candidate.ModelID {
		t.Errorf("IT15: key=%q, want model_id=%q", string(rec.Key), candidate.ModelID)
	}

	// Verify payload deserialisable with required fields.
	var decoded modelCandidatePayload
	if err := json.Unmarshal(rec.Value, &decoded); err != nil {
		t.Fatalf("IT15: deserialise model candidate: %v", err)
	}
	if decoded.ModelID == "" {
		t.Error("IT15: model_id is empty")
	}
	if decoded.Status == "" {
		t.Error("IT15: status is empty")
	}
	if decoded.Timestamp.IsZero() {
		t.Error("IT15: timestamp is zero")
	}

	t.Logf("IT15 PASS: model candidate on models.anomaly.published (model_id=%s, status=%s)",
		decoded.ModelID, decoded.Status)
}
