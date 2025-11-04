package rng

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
)

// RNG provides deterministic random number generation for a pipeline stage.
// Each stage derives its own seed from the master seed to ensure isolation
// and reproducibility. The derivation follows the formula:
//
//	seed_stage = H(masterSeed, stageName, configHash)
//
// where H is SHA-256 and the first 8 bytes are used as the uint64 seed.
//
// All methods are deterministic given the same initial seed, making dungeons
// reproducible across runs with identical inputs.
type RNG struct {
	seed      uint64
	stageName string
	source    *rand.Rand
}

// NewRNG creates a stage-specific RNG by deriving a sub-seed from the master seed.
// The derivation uses SHA-256 to combine:
//   - masterSeed: The top-level seed for the entire generation process
//   - stageName: Identifies the pipeline stage (e.g., "graph_synthesis", "embedding")
//   - configHash: Hash of the configuration to ensure different configs yield different results
//
// This ensures that:
//  1. Same inputs always produce the same RNG sequence (determinism)
//  2. Different stages get independent random sequences (isolation)
//  3. Config changes result in different sequences (sensitivity)
func NewRNG(masterSeed uint64, stageName string, configHash []byte) *RNG {
	// Derive sub-seed using SHA-256
	h := sha256.New()

	// Write master seed as big-endian bytes
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], masterSeed)
	h.Write(buf[:])

	// Write stage name to differentiate pipeline stages
	h.Write([]byte(stageName))

	// Write config hash to ensure config changes affect randomness
	h.Write(configHash)

	// Extract first 8 bytes of hash as uint64 seed
	hash := h.Sum(nil)
	derivedSeed := binary.BigEndian.Uint64(hash[:8])

	return &RNG{
		seed:      derivedSeed,
		stageName: stageName,
		source:    rand.New(rand.NewSource(int64(derivedSeed))),
	}
}

// Uint64 returns a pseudo-random 64-bit unsigned integer.
// The sequence is deterministic based on the RNG's seed.
func (r *RNG) Uint64() uint64 {
	return r.source.Uint64()
}

// Intn returns a pseudo-random integer in [0, n).
// It panics if n <= 0.
func (r *RNG) Intn(n int) int {
	if n <= 0 {
		panic("rng: Intn argument must be positive")
	}
	return r.source.Intn(n)
}

// Float64 returns a pseudo-random float64 in [0.0, 1.0).
func (r *RNG) Float64() float64 {
	return r.source.Float64()
}

// Shuffle pseudo-randomizes the order of elements in slice.
// The shuffle is deterministic based on the RNG's seed.
func (r *RNG) Shuffle(n int, swap func(i, j int)) {
	r.source.Shuffle(n, swap)
}

// Seed returns the derived seed for this RNG.
// This is useful for debugging and logging which seed was used for a stage.
func (r *RNG) Seed() uint64 {
	return r.seed
}

// StageName returns the stage name this RNG was created for.
// This is useful for debugging and logging.
func (r *RNG) StageName() string {
	return r.stageName
}

// IntRange returns a pseudo-random integer in [min, max].
// It panics if min > max.
func (r *RNG) IntRange(min, max int) int {
	if min > max {
		panic("rng: IntRange min must be <= max")
	}
	if min == max {
		return min
	}
	return min + r.source.Intn(max-min+1)
}

// Float64Range returns a pseudo-random float64 in [min, max).
// It panics if min >= max.
func (r *RNG) Float64Range(min, max float64) float64 {
	if min >= max {
		panic("rng: Float64Range min must be < max")
	}
	return min + r.source.Float64()*(max-min)
}

// Bool returns a pseudo-random boolean value.
func (r *RNG) Bool() bool {
	return r.source.Intn(2) == 1
}

// WeightedChoice selects an index from weights using weighted random selection.
// Weights must be non-negative. Returns -1 if all weights are zero or weights is empty.
func (r *RNG) WeightedChoice(weights []float64) int {
	if len(weights) == 0 {
		return -1
	}

	// Calculate total weight
	total := 0.0
	for _, w := range weights {
		if w < 0 {
			panic("rng: WeightedChoice weights must be non-negative")
		}
		total += w
	}

	if total == 0 {
		return -1
	}

	// Generate random value in [0, total)
	randVal := r.Float64() * total

	// Find the weighted index
	cumulative := 0.0
	for i, w := range weights {
		cumulative += w
		if randVal < cumulative {
			return i
		}
	}

	// Should not reach here, but return last index if we do
	return len(weights) - 1
}
