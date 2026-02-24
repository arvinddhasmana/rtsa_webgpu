// CLASSIFICATION: UNCLASSIFIED
package sensor

import (
"math/rand"
"sync"
)

var (
globalRNG   *rand.Rand
globalRNGMu sync.Mutex
)

// SetRNG configures the package-level RNG used by all sensor generators.
// Must be called before generating observations. Safe for concurrent use.
func SetRNG(r *rand.Rand) {
globalRNGMu.Lock()
defer globalRNGMu.Unlock()
globalRNG = r
}

// rng returns the package-level RNG, initialising with a default seed if needed.
func rng() *rand.Rand {
globalRNGMu.Lock()
defer globalRNGMu.Unlock()
if globalRNG == nil {
globalRNG = rand.New(rand.NewSource(0)) //nolint:gosec
}
return globalRNG
}
