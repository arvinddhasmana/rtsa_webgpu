// CLASSIFICATION: UNCLASSIFIED
package consumer_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-fusion-engine/internal/consumer"
	"go.uber.org/zap"
)

func TestNewSensorConsumer_NoBrokers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := consumer.NewSensorConsumer(nil, "grp", []string{"sensors.radar"}, logger)
	if err == nil {
		t.Error("expected error for empty broker list")
	}
}
