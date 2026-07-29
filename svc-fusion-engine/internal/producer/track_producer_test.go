// CLASSIFICATION: UNCLASSIFIED
package producer_test

import (
	"context"
	"testing"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/producer"
	"go.uber.org/zap"
)

func TestNewTrackProducer_NoBrokers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := producer.NewTrackProducer(context.Background(), nil, "tracks.fused", logger)
	if err == nil {
		t.Error("expected error for empty broker list")
	}
}
