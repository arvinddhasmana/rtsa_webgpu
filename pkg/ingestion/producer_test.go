// CLASSIFICATION: UNCLASSIFIED
package ingestion_test

import (
	"context"
	"testing"

	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

func TestLogProducer_Produce(t *testing.T) {
	p := ingestion.NewLogProducer("test-topic")
	defer p.Close()

	tests := []struct {
		name        string
		obs         *ingestionv1.SensorObservation
		expectError bool
	}{
		{
			name: "valid observation",
			obs: &ingestionv1.SensorObservation{
				ObservationId: "obs-1",
				SensorId:      "sns-1",
			},
			expectError: false,
		},
		{
			name:        "nil observation",
			obs:         nil,
			expectError: false, // protobuf V2 marshal succeeds on nil typed ptr
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Produce(context.Background(), tt.obs)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err != nil)
			}
		})
	}
}

func TestNoopProducer_Produce(t *testing.T) {
	p := &ingestion.NoopProducer{}
	defer p.Close()

	err := p.Produce(context.Background(), &ingestionv1.SensorObservation{})
	if err != nil {
		t.Errorf("expected no error from NoopProducer, got %v", err)
	}
}
