package dungeon

import (
	"context"
	"fmt"
	"math"

	"github.com/dshills/dungo/pkg/carving"
	"github.com/dshills/dungo/pkg/content"
	"github.com/dshills/dungo/pkg/embedding"
	"github.com/dshills/dungo/pkg/rng"
	"github.com/dshills/dungo/pkg/synthesis"
)

// Corridor length scaling constants
const (
	// corridorScaleMultiplier is the multiplier applied to sqrt(roomCount) for corridor length calculation.
	// Set to 59 based on empirical analysis of pathological seed 0x4400f4, which showed
	// force-directed layouts can produce corridors up to 51*sqrt(N) units. The value 59 provides
	// a 15% safety margin above the observed maximum.
	corridorScaleMultiplier = 59.0

	// corridorMinLength is the minimum corridor length for small dungeons.
	corridorMinLength = 100.0

	// corridorMaxLength is the maximum corridor length cap for very large dungeons.
	corridorMaxLength = 600.0
)

// Generator is the main entry point for procedural dungeon generation.
// Implementations must be deterministic: same Config+seed produces identical Artifact.
// This ensures reproducibility for seeded generation, testing, and debugging.
//
// All generators orchestrate a multi-stage pipeline:
//  1. Graph synthesis - creates abstract room/connector graph
//  2. Spatial embedding - assigns 2D positions and layouts
//  3. Tile carving - rasterizes rooms and corridors to tile grid
//  4. Content population - places enemies, loot, puzzles
//  5. Validation - checks constraints and computes metrics
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
	//
	// Example:
	//   gen := dungeon.NewGenerator()
	//   cfg := &dungeon.Config{Seed: 12345, Size: dungeon.SizeCfg{RoomsMin: 20, RoomsMax: 30}}
	//   artifact, err := gen.Generate(ctx, cfg)
	Generate(ctx context.Context, cfg *Config) (*Artifact, error)
}

// Validator validates dungeon generation results and computes metrics.
// This interface is defined here to avoid import cycles with the validation package.
type Validator interface {
	// Validate checks all constraints and computes metrics for the generated dungeon.
	// Returns a detailed ValidationReport with constraint results and metrics.
	// Returns error if validation process itself fails (not constraint failures).
	Validate(ctx context.Context, artifact *Artifact, cfg *Config) (*ValidationReport, error)
}

// DefaultGenerator implements Generator interface.
// It orchestrates the five-stage pipeline:
// 1. Graph synthesis (abstract dungeon graph)
// 2. Spatial embedding (layout with poses)
// 3. Tile carving (rasterized map)
// 4. Content population (enemies, loot, puzzles)
// 5. Validation (metrics and constraint checking)
type DefaultGenerator struct {
	synthesizer     synthesis.GraphSynthesizer
	embeddingConfig *embedding.Config // Base config, will be adjusted per dungeon
	carver          carving.Carver
	contentPass     content.ContentPass
	validator       Validator
}

// NewGenerator creates a new dungeon generator with default implementations.
// Note: You must call SetValidator() after creation to set the validator,
// or use NewGeneratorWithValidator() directly with a validator instance.
func NewGenerator() Generator {
	// Create base embedding config that will be adjusted per dungeon size
	// These are base values optimized for general use
	embeddingCfg := embedding.DefaultConfig()
	embeddingCfg.MinRoomSpacing = 1.0      // Decrease from 2.0 to 1.0 for tighter layouts
	embeddingCfg.CorridorMaxBends = 6      // Increase from 4 to 6 for more routing flexibility
	embeddingCfg.MaxIterations = 1000      // Increase for better convergence on large graphs
	embeddingCfg.CorridorMaxLength = 100.0 // Base value, will be scaled per dungeon

	return &DefaultGenerator{
		synthesizer:     synthesis.Get("grammar"),
		embeddingConfig: embeddingCfg,
		carver:          carving.NewDefaultCarver(16, 16), // 16x16 pixel tiles
		contentPass:     content.NewDefaultContentPass(),
		validator:       nil, // Must be set via SetValidator or use NewGeneratorWithValidator
	}
}

// NewGeneratorWithValidator creates a generator with a custom validator.
func NewGeneratorWithValidator(validator Validator) Generator {
	gen := NewGenerator().(*DefaultGenerator)
	gen.validator = validator
	return gen
}

// SetValidator sets the validator for this generator.
// This method allows setting the validator after construction to avoid import cycles.
func (g *DefaultGenerator) SetValidator(validator Validator) {
	g.validator = validator
}

// Generate creates a complete dungeon.
// Orchestrates all five pipeline stages with deterministic RNG seeding.
// nolint:gocyclo // Complexity acceptable: pipeline orchestration with multiple stages
func (g *DefaultGenerator) Generate(ctx context.Context, cfg *Config) (*Artifact, error) {
	// Stage 0: Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Compute config hash for RNG derivation
	configHash := cfg.Hash()

	// Create stage-specific RNGs: H(master_seed, stage_name, config_hash)
	synthesisRNG := rng.NewRNG(cfg.Seed, "synthesis", configHash)
	embeddingRNG := rng.NewRNG(cfg.Seed, "embedding", configHash)
	// carvingRNG := rng.NewRNG(cfg.Seed, "carving", configHash) // TODO: Use when carving needs RNG
	contentRNG := rng.NewRNG(cfg.Seed, "content", configHash)

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Stage A: Graph Synthesis
	synthesisCfg := &synthesis.Config{
		Seed:          cfg.Seed,
		RoomsMin:      cfg.Size.RoomsMin,
		RoomsMax:      cfg.Size.RoomsMax,
		BranchingAvg:  cfg.Branching.Avg,
		BranchingMax:  cfg.Branching.Max,
		SecretDensity: cfg.SecretDensity,
		OptionalRatio: cfg.OptionalRatio,
		Keys:          make([]synthesis.KeyConfig, len(cfg.Keys)),
		Pacing: synthesis.PacingConfig{
			Curve:        string(cfg.Pacing.Curve),
			Variance:     cfg.Pacing.Variance,
			CustomPoints: cfg.Pacing.CustomPoints,
		},
		Themes: cfg.Themes,
	}
	for i, k := range cfg.Keys {
		synthesisCfg.Keys[i] = synthesis.KeyConfig{
			Name:  k.Name,
			Count: k.Count,
		}
	}

	adgInternal, err := g.synthesizer.Synthesize(ctx, synthesisRNG, synthesisCfg)
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}

	// Wrap internal graph.Graph in dungeon.Graph
	adg := &Graph{
		Graph: adgInternal,
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Stage B: Spatial Embedding
	// Create embedder with parameters scaled to dungeon size
	embedderCfg := *g.embeddingConfig // Copy base config
	roomCount := len(adgInternal.Rooms)
	embedderCfg.CorridorMaxLength = calculateCorridorMaxLength(roomCount)

	// For medium-to-large dungeons (>25 rooms), adjust force balance to keep layout more compact
	// This prevents excessively spread-out layouts that exceed corridor length limits
	// Pathological seeds can create very spread-out layouts even with small dungeon counts
	if roomCount > 25 {
		// More aggressive force balancing:
		// - Increase spring constant (attraction) up to 10x
		// - Decrease repulsion constant down to 0.2x
		// This shifts the force balance dramatically for larger dungeons
		scaleFactor := 1.0 + (float64(roomCount-25) / 10.0) // Faster scaling: 1.0→6.5 for 25→80 rooms
		if scaleFactor > 10.0 {
			scaleFactor = 10.0 // Cap at 10x
		}
		embedderCfg.SpringConstant *= scaleFactor

		// Inversely scale repulsion - reduce it as dungeons get larger
		repulsionScale := 1.0 / (1.0 + float64(roomCount-25)/50.0)
		if repulsionScale < 0.2 {
			repulsionScale = 0.2 // Floor at 0.2x (don't eliminate repulsion completely)
		}
		embedderCfg.RepulsionConstant *= repulsionScale
	}

	embedder, err := embedding.Get("force_directed", &embedderCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	layoutInternal, err := embedder.Embed(adgInternal, embeddingRNG)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	// Normalize embedding layout BEFORE converting to center coordinates
	// This ensures we have Width/Height info to properly handle bounds
	normalizeEmbeddingLayout(layoutInternal)

	// CRITICAL: Recompute bounds after normalization since poses have been shifted
	// This is correct by design: normalization translates all positions to be non-negative,
	// which changes the overall bounding box. ComputeBounds() recalculates the final bounds
	// after translation to get accurate min/max coordinates for the normalized layout.
	layoutInternal.ComputeBounds()

	// Convert embedding.Layout to dungeon.Layout (corner → center coordinates)
	layout := convertEmbeddingLayout(layoutInternal)

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Stage C: Carving
	// Create graph adapter for carving
	graphAdapter := carving.NewGraphAdapter(adgInternal.Rooms, adgInternal.Connectors)

	// Convert dungeon.Layout to carving.Layout
	carvingLayout := convertToCarvingLayout(layout)

	tileMapInternal, err := g.carver.Carve(ctx, graphAdapter, carvingLayout)
	if err != nil {
		return nil, fmt.Errorf("carving failed: %w", err)
	}

	// Convert carving.TileMap to dungeon.TileMap
	tileMap := convertCarvingTileMap(tileMapInternal)

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Stage D: Content Population
	contentInternal, err := g.contentPass.Place(ctx, adgInternal, contentRNG)
	if err != nil {
		return nil, fmt.Errorf("content failed: %w", err)
	}

	// Convert content.Content to dungeon.Content
	contentData := convertContent(contentInternal)

	// Create artifact before validation
	artifact := &Artifact{
		ADG:     adg,
		Layout:  layout,
		TileMap: tileMap,
		Content: contentData,
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Stage E: Validation
	report, err := g.validator.Validate(ctx, artifact, cfg)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Add metrics and debug info to artifact
	artifact.Metrics = report.Metrics
	artifact.Debug = &DebugArtifacts{
		Report: report,
	}

	// Check if hard constraints were satisfied
	if !report.Passed {
		return nil, fmt.Errorf("hard constraints not satisfied: %v", report.Errors)
	}

	return artifact, nil
}

// convertEmbeddingLayout converts embedding.Layout to dungeon.Layout
func convertEmbeddingLayout(el *embedding.Layout) *Layout {
	if el == nil {
		return nil
	}

	layout := &Layout{
		Poses:         make(map[string]Pose),
		CorridorPaths: make(map[string]Path),
		Bounds: Rect{
			X:      int(el.Bounds.MinX),
			Y:      int(el.Bounds.MinY),
			Width:  int(el.Bounds.Width()),
			Height: int(el.Bounds.Height()),
		},
	}

	// Convert poses
	// embedding.Pose has X,Y as corner (top-left) coordinates with Width/Height
	// dungeon.Pose needs X,Y as center coordinates for carving
	// CRITICAL: Translate to be relative to bounds origin since tile map starts at (0,0)
	minX := el.Bounds.MinX
	minY := el.Bounds.MinY
	for roomID, pose := range el.Poses {
		layout.Poses[roomID] = Pose{
			X:           int(pose.X - minX + float64(pose.Width)/2),  // Translate and convert corner to center
			Y:           int(pose.Y - minY + float64(pose.Height)/2), // Translate and convert corner to center
			Rotation:    pose.Rotation,
			FootprintID: pose.FootprintID,
		}
	}

	// Convert corridor paths
	// CRITICAL: Also translate corridor paths to be relative to bounds origin
	for connID, path := range el.CorridorPaths {
		points := make([]Point, len(path.Points))
		for i, pt := range path.Points {
			points[i] = Point{
				X: int(pt.X - minX),
				Y: int(pt.Y - minY),
			}
		}
		layout.CorridorPaths[connID] = Path{Points: points}
	}

	return layout
}

// convertToCarvingLayout converts dungeon.Layout to carving.Layout.
// Note: Both dungeon.Layout and carving.Layout use center coordinates for poses,
// so this is a straightforward type conversion.
func convertToCarvingLayout(dl *Layout) *carving.Layout {
	if dl == nil {
		return nil
	}

	carvingLayout := &carving.Layout{
		Poses:         make(map[string]carving.Pose),
		CorridorPaths: make(map[string]carving.Path),
		Bounds: carving.Rect{
			X:      dl.Bounds.X,
			Y:      dl.Bounds.Y,
			Width:  dl.Bounds.Width,
			Height: dl.Bounds.Height,
		},
	}

	// Convert poses - coordinates are already in center format from convertEmbeddingLayout
	for roomID, pose := range dl.Poses {
		carvingLayout.Poses[roomID] = carving.Pose{
			X:           pose.X,
			Y:           pose.Y,
			Rotation:    pose.Rotation,
			FootprintID: pose.FootprintID,
		}
	}

	// Convert corridor paths
	for connID, path := range dl.CorridorPaths {
		points := make([]carving.Point, len(path.Points))
		for i, pt := range path.Points {
			points[i] = carving.Point{
				X: pt.X,
				Y: pt.Y,
			}
		}
		carvingLayout.CorridorPaths[connID] = carving.Path{Points: points}
	}

	return carvingLayout
}

// convertCarvingTileMap converts carving.TileMap to dungeon.TileMap
func convertCarvingTileMap(ct *carving.TileMap) *TileMap {
	if ct == nil {
		return nil
	}

	tileMap := &TileMap{
		Width:      ct.Width,
		Height:     ct.Height,
		TileWidth:  ct.TileWidth,
		TileHeight: ct.TileHeight,
		Layers:     make(map[string]*Layer),
	}

	// Convert layers
	for name, layer := range ct.Layers {
		objects := make([]Object, len(layer.Objects))
		for i, obj := range layer.Objects {
			objects[i] = Object{
				ID:         obj.ID,
				Name:       obj.Name,
				Type:       obj.Type,
				X:          obj.X,
				Y:          obj.Y,
				Width:      obj.Width,
				Height:     obj.Height,
				Rotation:   obj.Rotation,
				GID:        obj.GID,
				Visible:    obj.Visible,
				Properties: obj.Properties,
			}
		}

		tileMap.Layers[name] = &Layer{
			ID:      layer.ID,
			Name:    layer.Name,
			Type:    layer.Type,
			Visible: layer.Visible,
			Opacity: layer.Opacity,
			Data:    layer.Data,
			Objects: objects,
		}
	}

	return tileMap
}

// convertContent converts content.Content to dungeon.Content
func convertContent(cc *content.Content) *Content {
	if cc == nil {
		return nil
	}

	dungeonContent := &Content{
		Spawns:  make([]Spawn, len(cc.Spawns)),
		Loot:    make([]Loot, len(cc.Loot)),
		Puzzles: make([]PuzzleInstance, len(cc.Puzzles)),
		Secrets: make([]SecretInstance, len(cc.Secrets)),
	}

	// Convert spawns
	for i, spawn := range cc.Spawns {
		patrolPath := make([]Point, len(spawn.PatrolPath))
		for j, pt := range spawn.PatrolPath {
			patrolPath[j] = Point{X: pt.X, Y: pt.Y}
		}

		dungeonContent.Spawns[i] = Spawn{
			ID:         spawn.ID,
			RoomID:     spawn.RoomID,
			Position:   Point{X: spawn.Position.X, Y: spawn.Position.Y},
			EnemyType:  spawn.EnemyType,
			Count:      spawn.Count,
			PatrolPath: patrolPath,
		}
	}

	// Convert loot
	for i, loot := range cc.Loot {
		dungeonContent.Loot[i] = Loot{
			ID:       loot.ID,
			RoomID:   loot.RoomID,
			Position: Point{X: loot.Position.X, Y: loot.Position.Y},
			ItemType: loot.ItemType,
			Value:    loot.Value,
			Required: loot.Required,
		}
	}

	// Convert puzzles
	for i, puzzle := range cc.Puzzles {
		requirements := make([]Requirement, len(puzzle.Requirements))
		for j, req := range puzzle.Requirements {
			requirements[j] = Requirement{
				Type:  req.Type,
				Value: req.Value,
			}
		}

		provides := make([]Capability, len(puzzle.Provides))
		for j, cap := range puzzle.Provides {
			provides[j] = Capability{
				Type:  cap.Type,
				Value: cap.Value,
			}
		}

		dungeonContent.Puzzles[i] = PuzzleInstance{
			ID:           puzzle.ID,
			RoomID:       puzzle.RoomID,
			Type:         puzzle.Type,
			Requirements: requirements,
			Provides:     provides,
			Difficulty:   puzzle.Difficulty,
		}
	}

	// Convert secrets
	for i, secret := range cc.Secrets {
		clues := make([]string, len(secret.Clues))
		copy(clues, secret.Clues)

		dungeonContent.Secrets[i] = SecretInstance{
			ID:       secret.ID,
			RoomID:   secret.RoomID,
			Type:     secret.Type,
			Position: Point{X: secret.Position.X, Y: secret.Position.Y},
			Clues:    clues,
		}
	}

	return dungeonContent
}

// normalizeEmbeddingLayout translates all positions in an embedding.Layout to ensure
// that all room corners (and thus final carved positions) are non-negative.
// This must be called BEFORE converting to center coordinates, while we still have Width/Height.
func normalizeEmbeddingLayout(layout *embedding.Layout) {
	if layout == nil || len(layout.Poses) == 0 {
		return
	}

	// Find minimum corner positions (X, Y) across all poses
	// For embedding.Layout, X,Y are already corner positions
	var minX, minY float64
	first := true

	for _, pose := range layout.Poses {
		if first {
			minX = pose.X
			minY = pose.Y
			first = false
		} else {
			if pose.X < minX {
				minX = pose.X
			}
			if pose.Y < minY {
				minY = pose.Y
			}
		}
	}

	// Also check corridor paths
	for _, path := range layout.CorridorPaths {
		for _, pt := range path.Points {
			if pt.X < minX {
				minX = pt.X
			}
			if pt.Y < minY {
				minY = pt.Y
			}
		}
	}

	// If either minimum is negative, translate everything
	if minX < 0 || minY < 0 {
		offsetX := 0.0
		offsetY := 0.0
		if minX < 0 {
			offsetX = -minX
		}
		if minY < 0 {
			offsetY = -minY
		}

		// Translate all poses
		for roomID, pose := range layout.Poses {
			pose.X += offsetX
			pose.Y += offsetY
			layout.Poses[roomID] = pose
		}

		// Translate all corridor paths
		for connID, path := range layout.CorridorPaths {
			for i := range path.Points {
				path.Points[i].X += offsetX
				path.Points[i].Y += offsetY
			}
			layout.CorridorPaths[connID] = path
		}

		// Update bounds
		layout.Bounds.MinX += offsetX
		layout.Bounds.MinY += offsetY
		layout.Bounds.MaxX += offsetX
		layout.Bounds.MaxY += offsetY
	}
}

// calculateCorridorMaxLength computes an appropriate corridor max length based on dungeon size.
// The formula scales with the expected spatial extent of the dungeon:
//   - For N rooms, force-directed layout with InitialSpread=100 creates a layout
//     roughly proportional to sqrt(N) in each dimension
//   - Use sqrt(N) * 59 to scale corridor length with dungeon spatial extent
//   - Increased from 20→41→59 based on empirical analysis of pathological seeds
//   - Pathological cases can create corridors up to 51*sqrt(N) units
//   - This gives ~148 for 5 rooms, ~295 for 25 rooms, ~590 for 100 rooms, ~600 (max) for 103+ rooms
//   - Minimum of 100 for small dungeons, maximum of 600 for very large dungeons
//   - Combined with spring constant scaling (for dungeons >25 rooms) to keep layouts compact
func calculateCorridorMaxLength(roomCount int) float64 {
	if roomCount <= 0 {
		return corridorMinLength
	}

	// Scale with square root: max_length = sqrt(N) * corridorScaleMultiplier
	// See constants defined at top of file for rationale and empirical analysis.
	length := math.Sqrt(float64(roomCount)) * corridorScaleMultiplier

	// Apply bounds
	if length < corridorMinLength {
		length = corridorMinLength
	}
	if length > corridorMaxLength {
		length = corridorMaxLength
	}

	return length
}

// normalizeLayout translates all positions to ensure they are non-negative.
// This is necessary because force-directed embedding can produce negative coordinates.
// DEPRECATED: Use normalizeEmbeddingLayout instead, which handles coordinate systems correctly.
// nolint:unused
func normalizeLayout(layout *Layout) {
	if layout == nil || len(layout.Poses) == 0 {
		return
	}

	// Find minimum X and Y across all poses
	minX, minY := 0, 0
	first := true

	for _, pose := range layout.Poses {
		if first {
			minX = pose.X
			minY = pose.Y
			first = false
		} else {
			if pose.X < minX {
				minX = pose.X
			}
			if pose.Y < minY {
				minY = pose.Y
			}
		}
	}

	// Also check corridor paths
	for _, path := range layout.CorridorPaths {
		for _, pt := range path.Points {
			if pt.X < minX {
				minX = pt.X
			}
			if pt.Y < minY {
				minY = pt.Y
			}
		}
	}

	// If either minimum is negative, translate everything
	if minX < 0 || minY < 0 {
		offsetX := 0
		offsetY := 0
		if minX < 0 {
			offsetX = -minX
		}
		if minY < 0 {
			offsetY = -minY
		}

		// Translate all poses
		for roomID, pose := range layout.Poses {
			layout.Poses[roomID] = Pose{
				X:           pose.X + offsetX,
				Y:           pose.Y + offsetY,
				Rotation:    pose.Rotation,
				FootprintID: pose.FootprintID,
			}
		}

		// Translate all corridor paths
		for connID, path := range layout.CorridorPaths {
			newPoints := make([]Point, len(path.Points))
			for i, pt := range path.Points {
				newPoints[i] = Point{
					X: pt.X + offsetX,
					Y: pt.Y + offsetY,
				}
			}
			layout.CorridorPaths[connID] = Path{Points: newPoints}
		}

		// Recalculate bounds after normalization
		minX, minY, maxX, maxY := 0, 0, 0, 0
		first := true

		for _, pose := range layout.Poses {
			if first {
				minX = pose.X
				minY = pose.Y
				maxX = pose.X
				maxY = pose.Y
				first = false
			} else {
				if pose.X < minX {
					minX = pose.X
				}
				if pose.Y < minY {
					minY = pose.Y
				}
				if pose.X > maxX {
					maxX = pose.X
				}
				if pose.Y > maxY {
					maxY = pose.Y
				}
			}
		}

		// Check corridor paths for bounds
		for _, path := range layout.CorridorPaths {
			for _, pt := range path.Points {
				if pt.X < minX {
					minX = pt.X
				}
				if pt.Y < minY {
					minY = pt.Y
				}
				if pt.X > maxX {
					maxX = pt.X
				}
				if pt.Y > maxY {
					maxY = pt.Y
				}
			}
		}

		layout.Bounds = Rect{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
		}
	}
}
