// CLASSIFICATION: UNCLASSIFIED
package producer

import (
	"context"

	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type AlertProducer struct {
	producer *redpanda.Producer
	topic    string
	logger   *zap.Logger
}

func NewAlertProducer(producer *redpanda.Producer, topic string, logger *zap.Logger) *AlertProducer {
	return &AlertProducer{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}
}

func (p *AlertProducer) Produce(ctx context.Context, alert *inferencev1.SpatialAlert) error {
	// Use protojson to marshal
	data, err := protojson.Marshal(alert)
	if err != nil {
		return err
	}

	return p.producer.Produce(ctx, p.topic, []byte(alert.AlertId), data, "UNCLASSIFIED", "")
}
