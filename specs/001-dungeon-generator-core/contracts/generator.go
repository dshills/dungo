// Package contracts defines the core interfaces for the dungeon generator.
// This is a design document, not executable code.
package contracts

import (
	"context"
)

// Generator is the main entry point for dungeon generation.
// Implementations must be deterministic: same Config+seed → identical Artifact.
type Generator interface {
	// Generate creates a complete dungeon from configuration.
	// Returns error if hard constraints cannot be satisfied after retry limit.
	//
	// Contract:
	// - Must be deterministic (same input → same output)
	// - Must validate Config before processing
	// - Must enforce all hard constraints or fail
	// - Must produce ValidationReport as part of Artifact
	// - Context cancellation stops generation and returns partial Artifact
	Generate(ctx context.Context, cfg Config) (*Artifact, error)
}

// GraphSynthesizer generates the abstract dungeon graph (ADG).
//
// Phase: A (first stage of pipeline)
// Input: Config, RNG
// Output: Graph with rooms, connectors, and satisfied constraints
type GraphSynthesizer interface {
	// Synthesize creates an ADG satisfying all hard constraints.
	//
	// Contract:
	// - Must use provided RNG for all randomized decisions
	// - Must validate all hard constraints before returning
	// - Graph must have exactly one Start and one Boss room
	// - All rooms must be connected (unless Config.AllowDisconnected)
	// - Must enforce key-before-lock reachability
	// - Returns error if constraints unsatisfiable after retries
	Synthesize(ctx context.Context, rng RNG, cfg Config) (*Graph, error)

	// Name returns the strategy identifier (e.g., "grammar", "template", "optimizer").
	Name() string
}

// Embedder maps the abstract graph to spatial coordinates.
//
// Phase: B (second stage of pipeline)
// Input: Graph, RNG, Config
// Output: Layout with room poses and corridor paths
type Embedder interface {
	// Embed assigns spatial positions to all rooms and routes corridors.
	//
	// Contract:
	// - Must use provided RNG for all randomized decisions
	// - Must ensure no room bounding boxes overlap
	// - Must create feasible corridor paths (within length/bend limits)
	// - All poses must be within reasonable bounds
	// - Returns error if spatial constraints unsatisfiable
	Embed(ctx context.Context, rng RNG, g *Graph, cfg Config) (*Layout, error)

	// Name returns the embedder identifier (e.g., "force-directed", "orthogonal", "packing").
	Name() string
}

// Carver rasterizes the layout into a tile map.
//
// Phase: C (third stage of pipeline)
// Input: Layout, Graph (for room types), RNG, Config
// Output: TileMap with layered tiles
type Carver interface {
	// Carve generates tile layers from spatial layout.
	//
	// Contract:
	// - Must use provided RNG for randomized tile choices
	// - Must create layers: floor, walls, doors, decor
	// - Must stamp room footprints matching poses
	// - Must route corridors matching paths
	// - Door tiles placed at corridor/room junctions
	// - Returns error if tile map cannot fit layout
	Carve(ctx context.Context, rng RNG, layout *Layout, g *Graph, cfg Config) (*TileMap, error)

	// Name returns the carver identifier (e.g., "stamper", "cellular", "noise").
	Name() string
}

// ContentPass populates the dungeon with gameplay elements.
//
// Phase: D (fourth stage of pipeline)
// Input: TileMap, Graph, RNG, Config
// Output: Content with spawns, loot, puzzles, secrets
type ContentPass interface {
	// Populate places enemies, treasure, puzzles, and secrets.
	//
	// Contract:
	// - Must use provided RNG for all randomized decisions
	// - Must respect room capacity limits
	// - Must guarantee required items (keys) before their locks
	// - Must distribute loot along critical path and optionals
	// - Encounter difficulty matches room.Difficulty
	// - Returns error if content placement fails
	Populate(ctx context.Context, rng RNG, tm *TileMap, g *Graph, cfg Config) (*Content, error)

	// Name returns the content pass identifier (e.g., "budget-based", "wave-based").
	Name() string
}

// Validator checks dungeon correctness and calculates metrics.
//
// Phase: E (final stage of pipeline)
// Input: Complete Artifact, Config
// Output: ValidationReport
type Validator interface {
	// Validate verifies all constraints and computes metrics.
	//
	// Contract:
	// - Must check all hard constraints (connectivity, key reachability, no overlaps)
	// - Must calculate all metrics (branching, path length, pacing, cycles)
	// - Must produce detailed error messages for any failures
	// - Returns report regardless of pass/fail (report.Passed indicates result)
	Validate(ctx context.Context, artifact *Artifact, cfg Config) (*ValidationReport, error)
}

// RNG provides deterministic random number generation.
//
// All pipeline stages use RNG instead of global rand to ensure determinism.
// Each stage receives a sub-seed derived from master seed + stage name + config hash.
type RNG interface {
	// Uint64 returns a random uint64.
	Uint64() uint64

	// Intn returns a random int in [0, n).
	// Panics if n <= 0.
	Intn(n int) int

	// Float64 returns a random float64 in [0.0, 1.0).
	Float64() float64

	// Shuffle randomizes the order of elements in a slice.
	Shuffle(slice interface{})

	// Seed returns the current RNG seed for this stage.
	Seed() uint64
}

// Exporter converts Artifact to various output formats.
type Exporter interface {
	// ExportJSON serializes the complete artifact to JSON.
	ExportJSON(artifact *Artifact) ([]byte, error)

	// ExportTMJ exports tile map to Tiled TMJ format.
	ExportTMJ(tileMap *TileMap) ([]byte, error)

	// ExportSVG generates SVG visualization of the ADG.
	ExportSVG(graph *Graph, opts SVGOptions) ([]byte, error)
}

// ThemeLoader loads theme packs from filesystem.
type ThemeLoader interface {
	// LoadTheme reads a theme pack by name.
	//
	// Contract:
	// - Must validate theme pack structure
	// - Must load all required assets (encounter tables, loot tables)
	// - Returns error if theme pack is malformed or missing required data
	LoadTheme(name string) (*ThemePack, error)

	// ListThemes returns all available theme pack names.
	ListThemes() ([]string, error)
}

// ConstraintEvaluator executes constraint DSL expressions.
type ConstraintEvaluator interface {
	// Evaluate checks if a constraint is satisfied.
	//
	// Contract:
	// - Must parse and execute DSL expression
	// - Returns true if constraint satisfied, false otherwise
	// - Returns error if DSL syntax invalid
	Evaluate(constraint *Constraint, graph *Graph, layout *Layout) (bool, error)

	// Score calculates optimization score for soft constraints.
	// Returns float64 in [0.0, 1.0] where 1.0 is perfect satisfaction.
	Score(constraint *Constraint, graph *Graph, layout *Layout) (float64, error)
}

// Notes on Contract Design:
//
// 1. All pipeline stages take context.Context for cancellation support.
// 2. All pipeline stages use injected RNG for determinism.
// 3. Interfaces are narrow and focused on single responsibilities.
// 4. Errors are returned explicitly (no panics except programmer errors).
// 5. All methods document their contracts and invariants.
// 6. Generator is the only public-facing interface; others are internal.
//
// Implementation Strategy:
//
// 1. Implement DefaultGenerator that orchestrates pipeline stages.
// 2. Each stage implementation is independently testable.
// 3. Use registry pattern for pluggable strategies:
//    - synthesis.Register("grammar", &GrammarSynthesizer{})
//    - embedding.Register("force-directed", &ForceDirectedEmbedder{})
// 4. Config specifies which strategies to use by name.
// 5. All implementations are pure functions (no global state).
