package embedding

import (
	"fmt"
	"math"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// OrthogonalEmbedder uses a grid-based layout with BFS layering.
// This is a simpler, more predictable alternative to ForceDirectedEmbedder.
// It produces orthogonal (Manhattan-style) layouts similar to classic roguelikes.
//
// Algorithm:
//  1. BFS from Start room to assign layer indices
//  2. Assign rooms to grid cells based on layer and ordering
//  3. Route corridors using Manhattan paths (only horizontal/vertical)
//  4. Quantize all positions to grid alignment
//
// This embedder guarantees:
//   - No room overlaps (grid cells are exclusive)
//   - Short corridor paths (Manhattan distance)
//   - Predictable left-to-right or top-to-bottom progression
//   - Clean orthogonal aesthetic
//
// Trade-offs vs ForceDirectedEmbedder:
//   - Simpler, more predictable layouts
//   - Faster (no iterative simulation)
//   - Better for grid-based games
//   - Less organic, more artificial appearance
//   - May waste space with sparse layouts
type OrthogonalEmbedder struct {
	config *Config
}

// NewOrthogonalEmbedder creates an orthogonal grid-based embedder.
func NewOrthogonalEmbedder(config *Config) *OrthogonalEmbedder {
	if config == nil {
		config = DefaultConfig()
	}
	return &OrthogonalEmbedder{config: config}
}

// Name returns the identifier for this embedder.
func (e *OrthogonalEmbedder) Name() string {
	return "orthogonal"
}

// Embed performs orthogonal grid layout of the graph.
func (e *OrthogonalEmbedder) Embed(g *graph.Graph, rng *rng.RNG) (*Layout, error) {
	if g == nil {
		return nil, fmt.Errorf("cannot embed nil graph")
	}
	if rng == nil {
		return nil, fmt.Errorf("cannot embed with nil RNG")
	}
	if len(g.Rooms) == 0 {
		return nil, fmt.Errorf("cannot embed graph with no rooms")
	}

	// Phase 1: Find Start room
	startID := findStartRoom(g)
	if startID == "" {
		return nil, fmt.Errorf("no Start room found in graph")
	}

	// Phase 2: Assign layers via BFS from Start
	layers := e.assignLayers(g, startID)

	// Phase 3: Assign grid positions based on layers
	gridPositions := e.assignGridPositions(g, layers, rng)

	// Phase 4: Convert to Layout with Poses
	layout := NewLayout()
	layout.Algorithm = e.Name()
	layout.Seed = rng.Seed()

	for roomID, gridPos := range gridPositions {
		room := g.Rooms[roomID]
		width, height := SizeToGridDimensions(room.Size)

		// Convert grid position to world position
		// Use spacing to prevent rooms from being too close
		spacing := int(e.config.MinRoomSpacing)
		if spacing < 1 {
			spacing = 1
		}

		worldX := float64(gridPos.col * (12 + spacing)) // 12 is average room size
		worldY := float64(gridPos.row * (12 + spacing))

		pose := &Pose{
			X:        worldX,
			Y:        worldY,
			Width:    width,
			Height:   height,
			Rotation: 0, // Orthogonal embedder doesn't use rotation
		}

		if err := layout.AddPose(roomID, pose); err != nil {
			return nil, fmt.Errorf("failed to add pose: %w", err)
		}
	}

	// Phase 5: Route corridors using Manhattan paths
	if err := e.routeCorridors(g, layout); err != nil {
		return nil, fmt.Errorf("corridor routing failed: %w", err)
	}

	// Phase 6: Compute bounds
	layout.ComputeBounds()

	// Phase 7: Validate
	if err := ValidateEmbedding(layout, g, e.config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return layout, nil
}

// gridPosition represents a room's position in the grid.
type gridPosition struct {
	row, col int
}

// assignLayers performs BFS to assign each room to a layer.
// Layer 0 is the Start room, layer 1 is rooms adjacent to Start, etc.
func (e *OrthogonalEmbedder) assignLayers(g *graph.Graph, startID string) map[string]int {
	layers := make(map[string]int)
	queue := []string{startID}
	layers[startID] = 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentLayer := layers[current]

		// Process neighbors
		for _, neighborID := range g.Adjacency[current] {
			if _, visited := layers[neighborID]; !visited {
				layers[neighborID] = currentLayer + 1
				queue = append(queue, neighborID)
			}
		}
	}

	return layers
}

// assignGridPositions places rooms in a 2D grid based on their layers.
// Rooms in the same layer are placed in the same column (or row, depending on orientation).
func (e *OrthogonalEmbedder) assignGridPositions(g *graph.Graph, layers map[string]int, rng *rng.RNG) map[string]gridPosition {
	// Group rooms by layer
	maxLayer := 0
	layerGroups := make(map[int][]string)
	for roomID, layer := range layers {
		layerGroups[layer] = append(layerGroups[layer], roomID)
		if layer > maxLayer {
			maxLayer = layer
		}
	}

	// Assign positions: layer determines column, position in layer determines row
	positions := make(map[string]gridPosition)

	for layer := 0; layer <= maxLayer; layer++ {
		rooms := layerGroups[layer]
		if len(rooms) == 0 {
			continue
		}

		// Sort rooms in this layer for deterministic placement
		// (in real implementation, might sort by archetype or connectivity)
		// For simplicity, just place them in order
		for i, roomID := range rooms {
			positions[roomID] = gridPosition{
				row: i,
				col: layer,
			}
		}
	}

	return positions
}

// routeCorridors creates Manhattan-style corridors between connected rooms.
func (e *OrthogonalEmbedder) routeCorridors(g *graph.Graph, layout *Layout) error {
	for connID, conn := range g.Connectors {
		fromPose := layout.Poses[conn.From]
		toPose := layout.Poses[conn.To]

		if fromPose == nil || toPose == nil {
			return fmt.Errorf("missing pose for connector %s", connID)
		}

		// Get room centers
		fromX, fromY := fromPose.Center()
		toX, toY := toPose.Center()

		// Create Manhattan path (L-shaped)
		path := e.createManhattanPath(fromX, fromY, toX, toY)

		// Validate path length
		if path.Length() > e.config.CorridorMaxLength {
			// Try alternative routing (go vertical first instead of horizontal)
			altPath := e.createAlternateManhattanPath(fromX, fromY, toX, toY)
			if altPath.Length() <= e.config.CorridorMaxLength {
				path = altPath
			} else {
				return fmt.Errorf("corridor %s exceeds max length: %.1f > %.1f",
					connID, path.Length(), e.config.CorridorMaxLength)
			}
		}

		if err := layout.AddPath(connID, path); err != nil {
			return fmt.Errorf("failed to add path: %w", err)
		}
	}

	return nil
}

// createManhattanPath creates an L-shaped path: horizontal first, then vertical.
func (e *OrthogonalEmbedder) createManhattanPath(x1, y1, x2, y2 float64) *Path {
	points := []Point{
		{X: x1, Y: y1}, // Start
		{X: x2, Y: y1}, // Go horizontal
		{X: x2, Y: y2}, // Go vertical
	}

	// Simplify: remove middle point if already aligned
	if math.Abs(x1-x2) < 0.1 {
		// Vertically aligned, no horizontal segment needed
		points = []Point{
			{X: x1, Y: y1},
			{X: x2, Y: y2},
		}
	} else if math.Abs(y1-y2) < 0.1 {
		// Horizontally aligned, no vertical segment needed
		points = []Point{
			{X: x1, Y: y1},
			{X: x2, Y: y2},
		}
	}

	return &Path{Points: points}
}

// createAlternateManhattanPath creates an L-shaped path: vertical first, then horizontal.
func (e *OrthogonalEmbedder) createAlternateManhattanPath(x1, y1, x2, y2 float64) *Path {
	points := []Point{
		{X: x1, Y: y1}, // Start
		{X: x1, Y: y2}, // Go vertical
		{X: x2, Y: y2}, // Go horizontal
	}

	// Simplify: remove middle point if already aligned
	if math.Abs(x1-x2) < 0.1 {
		// Vertically aligned
		points = []Point{
			{X: x1, Y: y1},
			{X: x2, Y: y2},
		}
	} else if math.Abs(y1-y2) < 0.1 {
		// Horizontally aligned
		points = []Point{
			{X: x1, Y: y1},
			{X: x2, Y: y2},
		}
	}

	return &Path{Points: points}
}

// findStartRoom locates the Start room in the graph.
func findStartRoom(g *graph.Graph) string {
	for id, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			return id
		}
	}
	return ""
}

// Register the orthogonal embedder.
func init() {
	Register("orthogonal", func(config *Config) Embedder {
		return NewOrthogonalEmbedder(config)
	})
}
