// Package rng provides deterministic random number generation for the dungeon generator.
//
// # Overview
//
// The RNG type ensures reproducible dungeon generation by deriving stage-specific
// seeds from a master seed. This allows each pipeline stage (graph synthesis,
// embedding, carving, content placement) to have independent random sequences
// while maintaining overall determinism.
//
// # Sub-Seed Derivation
//
// Each RNG derives its seed using SHA-256:
//
//	seed_stage = H(masterSeed, stageName, configHash)
//
// where:
//   - masterSeed: Top-level seed for entire generation
//   - stageName: Pipeline stage identifier (e.g., "graph_synthesis")
//   - configHash: Hash of configuration parameters
//
// This ensures:
//  1. Same inputs always produce same RNG sequence (determinism)
//  2. Different stages get independent random sequences (isolation)
//  3. Config changes result in different sequences (sensitivity)
//
// # Usage
//
// Create an RNG for each pipeline stage:
//
//	configHash := sha256.Sum256([]byte(configJSON))
//	graphRNG := rng.NewRNG(masterSeed, "graph_synthesis", configHash[:])
//	embedRNG := rng.NewRNG(masterSeed, "embedding", configHash[:])
//
// Use the RNG for all random decisions in that stage:
//
//	roomCount := graphRNG.IntRange(10, 50)
//	difficulty := graphRNG.Float64Range(0.3, 0.8)
//	if graphRNG.Bool() {
//	    // spawn optional room
//	}
//
// # Thread Safety
//
// RNG instances are NOT thread-safe. Each goroutine should use its own RNG
// instance. Create stage-specific RNGs before spawning goroutines and pass
// them explicitly.
//
// # Performance
//
// The underlying math/rand.Rand is highly efficient:
//   - Uint64(): ~2ns per call
//   - Intn():   ~3ns per call
//   - Float64(): ~2ns per call
//
// Creating a new RNG costs ~8µs due to SHA-256 computation.
// Reuse RNG instances within a stage for best performance.
package rng
