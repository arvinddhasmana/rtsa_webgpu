// CLASSIFICATION: UNCLASSIFIED
package classification

import "fmt"

// Level represents a data classification level.
type Level int32

const (
LevelUnspecified  Level = 0
LevelUnclassified Level = 1
LevelProtectedA   Level = 2
LevelProtectedB   Level = 3
LevelProtectedC   Level = 4
LevelSecret       Level = 5
)

// Guard enforces classification policies.
type Guard struct{}

// NewGuard creates a new classification guard.
func NewGuard() *Guard { return &Guard{} }

// Enforce returns an error if dataLevel exceeds requiredLevel.
func (g *Guard) Enforce(dataLevel, requiredLevel Level) error {
if dataLevel > requiredLevel {
return fmt.Errorf("[classification.Guard.Enforce]: data level %d exceeds required level %d", dataLevel, requiredLevel)
}
return nil
}

// MaxLevel returns the highest classification level among the provided levels.
func (g *Guard) MaxLevel(levels ...Level) Level {
var max Level
for _, l := range levels {
if l > max {
max = l
}
}
return max
}
