// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"strings"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ew-ingestion/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func validEwObs() *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        "EW-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_EW_SIGINT,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ObservationTime: timestamppb.New(time.Now()),
		SensorData: &ingestionv1.SensorObservation_EwSigint{
			EwSigint: &ingestionv1.EWIntercept{
				EmitterId:      "EM-123",
				FrequencyMhz:   3000.0,
				BearingDegrees: 45.0,
				Confidence:     0.9,
				ModulationType: "AM",
			},
		},
	}
}

func TestEWValidator_Exhaustive(t *testing.T) {
	v := domain.NewValidator()

	t.Run("Valid", func(t *testing.T) {
		if !v.Validate(validEwObs()).Valid {
			t.Error("expected valid")
		}
	})

	t.Run("SensorID_Empty", func(t *testing.T) {
		obs := validEwObs()
		obs.SensorId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty sensor id")
		}
	})

	t.Run("SensorID_Long", func(t *testing.T) {
		obs := validEwObs()
		obs.SensorId = strings.Repeat("a", 129)
		if v.Validate(obs).Valid {
			t.Error("expected invalid: long sensor id")
		}
	})

	t.Run("WrongSensorType", func(t *testing.T) {
		obs := validEwObs()
		obs.SensorType = commonv1.SensorType_SENSOR_TYPE_RADAR
		if v.Validate(obs).Valid {
			t.Error("expected invalid: wrong sensor type")
		}
	})

	t.Run("ObservationTime_Nil", func(t *testing.T) {
		obs := validEwObs()
		obs.ObservationTime = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: nil time")
		}
	})

	t.Run("ObservationTime_Future", func(t *testing.T) {
		obs := validEwObs()
		obs.ObservationTime = timestamppb.New(time.Now().Add(10 * time.Minute))
		if v.Validate(obs).Valid {
			t.Error("expected invalid: future time")
		}
	})

	t.Run("Classification_Unspecified", func(t *testing.T) {
		obs := validEwObs()
		obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
		if v.Validate(obs).Valid {
			t.Error("expected invalid: unspecified classification")
		}
	})

	t.Run("MissingPayload", func(t *testing.T) {
		obs := validEwObs()
		obs.SensorData = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: missing payload")
		}
	})

	t.Run("EmitterID_Empty", func(t *testing.T) {
		obs := validEwObs()
		obs.GetEwSigint().EmitterId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty emitter id")
		}
	})

	t.Run("Frequency_OutOfRange", func(t *testing.T) {
		obs := validEwObs()
		obs.GetEwSigint().FrequencyMhz = 50000.0
		if v.Validate(obs).Valid {
			t.Error("expected invalid: freq too high")
		}
	})

	t.Run("Bearing_OutOfRange", func(t *testing.T) {
		obs := validEwObs()
		obs.GetEwSigint().BearingDegrees = 400.0
		if v.Validate(obs).Valid {
			t.Error("expected invalid: bearing")
		}
	})

	t.Run("Confidence_OutOfRange", func(t *testing.T) {
		obs := validEwObs()
		obs.GetEwSigint().Confidence = 1.5
		if v.Validate(obs).Valid {
			t.Error("expected invalid: confidence")
		}
	})

	t.Run("Position_Latitude", func(t *testing.T) {
		obs := validEwObs()
		obs.Position = &commonv1.Position{Latitude: 100}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: latitude")
		}
	})

	t.Run("Position_Longitude", func(t *testing.T) {
		obs := validEwObs()
		obs.Position = &commonv1.Position{Longitude: 200}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: longitude")
		}
	})
}
