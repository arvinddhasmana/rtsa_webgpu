// CLASSIFICATION: UNCLASSIFIED
package classification

import (
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
)

// PropagateMax propagates the maximum classification level across a set of
// source levels. Used when aggregating data from multiple classified sources.
func PropagateMax(sources []commonv1.ClassificationLevel) commonv1.ClassificationLevel {
return MaxAll(sources...)
}

// IsValid returns true if the classification level is a known, non-UNSPECIFIED value.
func IsValid(level commonv1.ClassificationLevel) bool {
_, ok := LevelOrder[level]
return ok && level != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
}
