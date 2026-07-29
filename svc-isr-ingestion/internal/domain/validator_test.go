// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"strings"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-isr-ingestion/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func validIsrObs() *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        "ISR-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_ISR,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ObservationTime: timestamppb.New(time.Now()),
		SensorData: &ingestionv1.SensorObservation_Isr{
			Isr: &ingestionv1.ISRObservation{
				PlatformId: "GlobalHawk",
				SensorName: "EO",
				ImageId:    "IMG-123",
				CoveragePolygon: []*commonv1.Position{
					{Latitude: 0, Longitude: 0},
					{Latitude: 1, Longitude: 0},
					{Latitude: 1, Longitude: 1},
				},
			},
		},
	}
}

func TestISRValidator_Exhaustive(t *testing.T) {
	v := domain.NewValidator()

	t.Run("Valid", func(t *testing.T) {
		if !v.Validate(validIsrObs()).Valid {
			t.Error("expected valid")
		}
	})

	t.Run("SensorID_Empty", func(t *testing.T) {
		obs := validIsrObs()
		obs.SensorId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty sensor id")
		}
	})

	t.Run("SensorID_Long", func(t *testing.T) {
		obs := validIsrObs()
		obs.SensorId = strings.Repeat("a", 129)
		if v.Validate(obs).Valid {
			t.Error("expected invalid: long sensor id")
		}
	})

	t.Run("WrongSensorType", func(t *testing.T) {
		obs := validIsrObs()
		obs.SensorType = commonv1.SensorType_SENSOR_TYPE_RADAR
		if v.Validate(obs).Valid {
			t.Error("expected invalid: wrong sensor type")
		}
	})

	t.Run("ObservationTime_Nil", func(t *testing.T) {
		obs := validIsrObs()
		obs.ObservationTime = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: nil time")
		}
	})

	t.Run("ObservationTime_Future", func(t *testing.T) {
		obs := validIsrObs()
		obs.ObservationTime = timestamppb.New(time.Now().Add(10 * time.Minute))
		if v.Validate(obs).Valid {
			t.Error("expected invalid: future time")
		}
	})

	t.Run("Classification_Unspecified", func(t *testing.T) {
		obs := validIsrObs()
		obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
		if v.Validate(obs).Valid {
			t.Error("expected invalid: unspecified classification")
		}
	})

	t.Run("MissingPayload", func(t *testing.T) {
		obs := validIsrObs()
		obs.SensorData = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: missing payload")
		}
	})

	t.Run("PlatformID_Empty", func(t *testing.T) {
		obs := validIsrObs()
		obs.GetIsr().PlatformId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty platform id")
		}
	})

	t.Run("SensorName_Empty", func(t *testing.T) {
		obs := validIsrObs()
		obs.GetIsr().SensorName = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty sensor name")
		}
	})

	t.Run("ImageID_Empty", func(t *testing.T) {
		obs := validIsrObs()
		obs.GetIsr().ImageId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty image id")
		}
	})

	t.Run("Polygon_Invalid", func(t *testing.T) {
		obs := validIsrObs()
		obs.GetIsr().CoveragePolygon = []*commonv1.Position{
			{Latitude: 0, Longitude: 0},
		}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: too few vertices")
		}
	})

	t.Run("Detection_InvalidConfidence", func(t *testing.T) {
		obs := validIsrObs()
		obs.GetIsr().Detections = []*ingestionv1.ISRDetection{
			{Confidence: 2.0},
		}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: confidence")
		}
	})

	t.Run("Position_Latitude", func(t *testing.T) {
		obs := validIsrObs()
		obs.Position = &commonv1.Position{Latitude: 100}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: latitude")
		}
	})

	t.Run("Position_Longitude", func(t *testing.T) {
		obs := validIsrObs()
		obs.Position = &commonv1.Position{Longitude: 200}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: longitude")
		}
	})
}
