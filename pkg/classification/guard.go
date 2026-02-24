// CLASSIFICATION: UNCLASSIFIED
package classification

import (
"fmt"
"strings"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
)

// LevelOrder maps ClassificationLevel to numeric order for comparison.
var LevelOrder = map[commonv1.ClassificationLevel]int{
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED:  0,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED: 1,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A:  2,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B:  3,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C:  4,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET:       5,
}

// Guard enforces classification level ceiling for a service.
type Guard struct {
maxLevel commonv1.ClassificationLevel
}

// NewGuard creates a Guard with the specified maximum classification ceiling.
func NewGuard(maxLevel commonv1.ClassificationLevel) *Guard {
return &Guard{maxLevel: maxLevel}
}

// Check returns nil if dataLevel ≤ maxLevel, error otherwise.
func (g *Guard) Check(dataLevel commonv1.ClassificationLevel) error {
if LevelOrder[dataLevel] > LevelOrder[g.maxLevel] {
return fmt.Errorf("classification: data level %s exceeds service ceiling %s",
LevelToString(dataLevel), LevelToString(g.maxLevel))
}
return nil
}

// CanAccess returns true if callerClearance >= dataLevel.
func CanAccess(callerClearance, dataLevel commonv1.ClassificationLevel) bool {
return LevelOrder[callerClearance] >= LevelOrder[dataLevel]
}

// Max returns the higher of two classification levels.
func Max(a, b commonv1.ClassificationLevel) commonv1.ClassificationLevel {
if LevelOrder[a] >= LevelOrder[b] {
return a
}
return b
}

// MaxAll returns the highest classification level among all provided.
func MaxAll(levels ...commonv1.ClassificationLevel) commonv1.ClassificationLevel {
if len(levels) == 0 {
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
}
max := levels[0]
for _, l := range levels[1:] {
max = Max(max, l)
}
return max
}

// LevelToString converts a ClassificationLevel to its string representation.
func LevelToString(level commonv1.ClassificationLevel) string {
switch level {
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED:
return "UNCLASSIFIED"
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A:
return "PROTECTED_A"
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B:
return "PROTECTED_B"
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C:
return "PROTECTED_C"
case commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET:
return "SECRET"
default:
return "UNSPECIFIED"
}
}

// StringToLevel converts a string to ClassificationLevel.
func StringToLevel(s string) commonv1.ClassificationLevel {
switch strings.ToUpper(strings.TrimSpace(s)) {
case "UNCLASSIFIED":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
case "PROTECTED_A":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A
case "PROTECTED_B":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B
case "PROTECTED_C":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C
case "SECRET":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
default:
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
}
}
