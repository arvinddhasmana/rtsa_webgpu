// CLASSIFICATION: UNCLASSIFIED
package classification_test

import (
	"testing"
	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
)

func TestPropagation(t *testing.T) {
	sources := []commonv1.ClassificationLevel{
		commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}
	if classification.PropagateMax(sources) != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
		t.Error("expected SECRET")
	}
}

func TestClassification_IsValid(t *testing.T) {
	if !classification.IsValid(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED) {
		t.Error("expected UNCLASSIFIED to be valid")
	}
	if classification.IsValid(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED) {
		t.Error("expected UNSPECIFIED to be invalid")
	}
	if classification.IsValid(commonv1.ClassificationLevel(999)) {
		t.Error("expected unknown level to be invalid")
	}
}
