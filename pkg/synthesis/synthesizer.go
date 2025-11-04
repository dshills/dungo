package synthesis

import (
	"context"
	"fmt"
	"sync"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// Config contains the configuration parameters needed for graph synthesis.
// This is a subset of dungeon.Config focused on synthesis concerns.
type Config struct {
	Seed          uint64
	RoomsMin      int
	RoomsMax      int
	BranchingAvg  float64
	BranchingMax  int
	SecretDensity float64
	OptionalRatio float64
	Keys          []KeyConfig
	Pacing        PacingConfig // Difficulty curve configuration
	Themes        []string     // Theme names for biome assignment
}

// PacingConfig defines the difficulty curve for the dungeon.
type PacingConfig struct {
	Curve        string       // LINEAR, S_CURVE, EXPONENTIAL, or CUSTOM
	Variance     float64      // Random variance (0.0-0.3)
	CustomPoints [][2]float64 // For CUSTOM curve
}

// KeyConfig defines a key/lock configuration.
type KeyConfig struct {
	Name  string
	Count int
}

// GraphSynthesizer is the interface for all graph synthesis strategies.
// Implementations must be deterministic: same RNG+Config produces identical Graph.
//
// Graph synthesis is the FIRST stage in the dungeon generation pipeline.
// It creates the Abstract Dungeon Graph (ADG) - a purely topological representation
// with no spatial information. The ADG defines room types, connections, difficulty
// progression, and key-lock puzzles.
//
// Available implementations:
//   - "grammar" (GrammarSynthesizer): Production rule-based, flexible, hub-and-spoke
//   - "template" (TemplateSynthesizer): Template-stitching, predictable, architectural
//
// Contract:
// - Must use provided RNG for all randomness
// - Must create exactly 1 Start room and 1 Boss room
// - Must ensure graph connectivity (all rooms reachable from Start)
// - Must enforce key-before-lock constraints
// - Must respect room count bounds from Config
// - Must satisfy all hard constraints or return error
type GraphSynthesizer interface {
	// Synthesize generates an Abstract Dungeon Graph from configuration.
	// Returns error if hard constraints cannot be satisfied after retries.
	Synthesize(ctx context.Context, rng *rng.RNG, cfg *Config) (*graph.Graph, error)

	// Name returns the synthesizer's identifier for registration.
	Name() string
}

// Registry manages available graph synthesizers.
var (
	synthesizersMu sync.RWMutex
	synthesizers   = make(map[string]GraphSynthesizer)
)

// Register adds a synthesizer to the global registry.
// Panics if name is already registered.
func Register(name string, s GraphSynthesizer) {
	synthesizersMu.Lock()
	defer synthesizersMu.Unlock()

	if _, exists := synthesizers[name]; exists {
		panic(fmt.Sprintf("synthesizer %q already registered", name))
	}

	synthesizers[name] = s
}

// Get retrieves a registered synthesizer by name.
// Returns nil if not found.
func Get(name string) GraphSynthesizer {
	synthesizersMu.RLock()
	defer synthesizersMu.RUnlock()

	return synthesizers[name]
}

// List returns all registered synthesizer names.
func List() []string {
	synthesizersMu.RLock()
	defer synthesizersMu.RUnlock()

	names := make([]string, 0, len(synthesizers))
	for name := range synthesizers {
		names = append(names, name)
	}
	return names
}
