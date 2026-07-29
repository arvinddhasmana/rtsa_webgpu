// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"strings"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-elint-ingestion/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func validElintObs() *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        "ELINT-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ObservationTime: timestamppb.New(time.Now()),
		SensorData: &ingestionv1.SensorObservation_ElintComint{
			ElintComint: &ingestionv1.ELINTDetection{
				EmitterId:             "EM-123",
				RadarType:             "GENERIC",
				FrequencyMhz:          1200.0,
				CepMeters:             50.0,
				Confidence:            0.9,
				ContentClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
				ScanType:              "circular",
			},
		},
	}
}

func TestELINTValidator_Exhaustive(t *testing.T) {
	v := domain.NewValidator()

	t.Run("Valid", func(t *testing.T) {
		if !v.Validate(validElintObs()).Valid {
			t.Error("expected valid")
		}
	})

	t.Run("SensorID_Empty", func(t *testing.T) {
		obs := validElintObs()
		obs.SensorId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty sensor id")
		}
	})

	t.Run("SensorID_Long", func(t *testing.T) {
		obs := validElintObs()
		obs.SensorId = strings.Repeat("a", 129)
		if v.Validate(obs).Valid {
			t.Error("expected invalid: long sensor id")
		}
	})

	t.Run("WrongSensorType", func(t *testing.T) {
		obs := validElintObs()
		obs.SensorType = commonv1.SensorType_SENSOR_TYPE_RADAR
		if v.Validate(obs).Valid {
			t.Error("expected invalid: wrong sensor type")
		}
	})

	t.Run("ObservationTime_Nil", func(t *testing.T) {
		obs := validElintObs()
		obs.ObservationTime = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: nil time")
		}
	})

	t.Run("ObservationTime_Future", func(t *testing.T) {
		obs := validElintObs()
		obs.ObservationTime = timestamppb.New(time.Now().Add(10 * time.Minute))
		if v.Validate(obs).Valid {
			t.Error("expected invalid: future time")
		}
	})

	t.Run("ObservationTime_Past", func(t *testing.T) {
		obs := validElintObs()
		obs.ObservationTime = timestamppb.New(time.Now().Add(-48 * time.Hour))
		if v.Validate(obs).Valid {
			t.Error("expected invalid: past time")
		}
	})

	t.Run("Classification_Unspecified", func(t *testing.T) {
		obs := validElintObs()
		obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
		if v.Validate(obs).Valid {
			t.Error("expected invalid: unspecified classification")
		}
	})

	t.Run("MissingPayload", func(t *testing.T) {
		obs := validElintObs()
		obs.SensorData = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: missing payload")
		}
	})

	t.Run("EmitterID_Empty", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().EmitterId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty emitter id")
		}
	})

	t.Run("RadarType_Empty", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().RadarType = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty radar type")
		}
	})

	t.Run("Frequency_Low", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().FrequencyMhz = 0.1
		if v.Validate(obs).Valid {
			t.Error("expected invalid: freq too low")
		}
	})

	t.Run("Frequency_High", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().FrequencyMhz = 50000.0
		if v.Validate(obs).Valid {
			t.Error("expected invalid: freq too high")
		}
	})

	t.Run("CEPMeters_NonPositive", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().CepMeters = -1
		if v.Validate(obs).Valid {
			t.Error("expected invalid: negative cep")
		}
	})

	t.Run("Confidence_Low", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().Confidence = -0.1
		if v.Validate(obs).Valid {
			t.Error("expected invalid: negative confidence")
		}
	})

	t.Run("Confidence_High", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().Confidence = 1.1
		if v.Validate(obs).Valid {
			t.Error("expected invalid: confidence > 1")
		}
	})

	t.Run("ContentClassification_Unspecified", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().ContentClassification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
		if v.Validate(obs).Valid {
			t.Error("expected invalid: unspecified content classification")
		}
	})

	t.Run("Classification_Ceiling", func(t *testing.T) {
		obs := validElintObs()
		obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
		obs.GetElintComint().ContentClassification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
		if v.Validate(obs).Valid {
			t.Error("expected invalid: classification ceiling violation")
		}
	})

	t.Run("ScanType_Invalid", func(t *testing.T) {
		obs := validElintObs()
		obs.GetElintComint().ScanType = "invalid"
		if v.Validate(obs).Valid {
			t.Error("expected invalid: scan type")
		}
	})

	t.Run("Position_Latitude", func(t *testing.T) {
		obs := validElintObs()
		obs.Position = &commonv1.Position{Latitude: 100}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: latitude")
		}
	})

	t.Run("Position_Longitude", func(t *testing.T) {
		obs := validElintObs()
		obs.Position = &commonv1.Position{Longitude: 200}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: longitude")
		}
	})
}
