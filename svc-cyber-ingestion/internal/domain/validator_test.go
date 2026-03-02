// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"strings"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const validSHA256 = "a948904f2f0f479b8f936378f543e5a9b5a1c0a3e6b8f9c2d0e4f6a8c0e2d4f6"

func validCyberObs() *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        "CYBER-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_CYBER,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ObservationTime: timestamppb.New(time.Now()),
		SensorData: &ingestionv1.SensorObservation_Cyber{
			Cyber: &ingestionv1.CyberIOC{
				StixId:     "indicator--12345678-1234-1234-1234-123456789012",
				IocType:    "ipv4-addr",
				IocValue:   "192.168.1.1",
				Confidence: 0.9,
				ValidFrom:  timestamppb.New(time.Now().Add(-1 * time.Hour)),
				SourceFeed: "CTI-FEED-01",
				DedupHash:  validSHA256,
			},
		},
	}
}

func TestCyberValidator_Exhaustive(t *testing.T) {
	v := domain.NewValidator()

	t.Run("Valid", func(t *testing.T) {
		if !v.Validate(validCyberObs()).Valid {
			t.Error("expected valid")
		}
	})

	t.Run("SensorID_Empty", func(t *testing.T) {
		obs := validCyberObs()
		obs.SensorId = ""
		if v.Validate(obs).Valid {
			t.Error("expected invalid: empty sensor id")
		}
	})

	t.Run("SensorID_Long", func(t *testing.T) {
		obs := validCyberObs()
		obs.SensorId = strings.Repeat("a", 129)
		if v.Validate(obs).Valid {
			t.Error("expected invalid: long sensor id")
		}
	})

	t.Run("WrongSensorType", func(t *testing.T) {
		obs := validCyberObs()
		obs.SensorType = commonv1.SensorType_SENSOR_TYPE_RADAR
		if v.Validate(obs).Valid {
			t.Error("expected invalid: wrong sensor type")
		}
	})

	t.Run("ObservationTime_Nil", func(t *testing.T) {
		obs := validCyberObs()
		obs.ObservationTime = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: nil time")
		}
	})

	t.Run("ObservationTime_Future", func(t *testing.T) {
		obs := validCyberObs()
		obs.ObservationTime = timestamppb.New(time.Now().Add(10 * time.Minute))
		if v.Validate(obs).Valid {
			t.Error("expected invalid: future time")
		}
	})

	t.Run("Classification_Unspecified", func(t *testing.T) {
		obs := validCyberObs()
		obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
		if v.Validate(obs).Valid {
			t.Error("expected invalid: unspecified classification")
		}
	})

	t.Run("MissingPayload", func(t *testing.T) {
		obs := validCyberObs()
		obs.SensorData = nil
		if v.Validate(obs).Valid {
			t.Error("expected invalid: missing payload")
		}
	})

	t.Run("STIXID_Invalid", func(t *testing.T) {
		obs := validCyberObs()
		obs.GetCyber().StixId = "bad-id"
		if v.Validate(obs).Valid {
			t.Error("expected invalid: STIX ID")
		}
	})

	t.Run("IOCType_Invalid", func(t *testing.T) {
		obs := validCyberObs()
		obs.GetCyber().IocType = "invalid"
		if v.Validate(obs).Valid {
			t.Error("expected invalid: IOC type")
		}
	})

	t.Run("IPv4_Invalid", func(t *testing.T) {
		obs := validCyberObs()
		obs.GetCyber().IocType = "ipv4-addr"
		obs.GetCyber().IocValue = "bad-ip"
		if v.Validate(obs).Valid {
			t.Error("expected invalid: IPv4")
		}
	})

	t.Run("SHA256_Invalid", func(t *testing.T) {
		obs := validCyberObs()
		obs.GetCyber().IocType = "file-sha256"
		obs.GetCyber().IocValue = "bad-hash"
		if v.Validate(obs).Valid {
			t.Error("expected invalid: SHA256")
		}
	})

	t.Run("DedupHash_Invalid", func(t *testing.T) {
		obs := validCyberObs()
		obs.GetCyber().DedupHash = "bad-hash"
		if v.Validate(obs).Valid {
			t.Error("expected invalid: DedupHash")
		}
	})

	t.Run("ValidFrom_Future", func(t *testing.T) {
		obs := validCyberObs()
		obs.GetCyber().ValidFrom = timestamppb.New(time.Now().Add(1 * time.Hour))
		if v.Validate(obs).Valid {
			t.Error("expected invalid: future valid from")
		}
	})

	t.Run("Position_Latitude", func(t *testing.T) {
		obs := validCyberObs()
		obs.Position = &commonv1.Position{Latitude: 100}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: latitude")
		}
	})

	t.Run("Position_Longitude", func(t *testing.T) {
		obs := validCyberObs()
		obs.Position = &commonv1.Position{Longitude: 200}
		if v.Validate(obs).Valid {
			t.Error("expected invalid: longitude")
		}
	})
}
