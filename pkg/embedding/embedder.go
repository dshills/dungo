package embedding

import (
	"fmt"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// Embedder transforms an Abstract Dungeon Graph into a spatial layout.
// It assigns 2D coordinates and dimensions to all rooms, and creates
// corridor paths between connected rooms.
//
// Embedding is the SECOND stage in the dungeon generation pipeline.
// It takes the abstract graph from synthesis and produces a 2D spatial layout
// ready for tile carving. The embedder must handle:
//   - Room placement (avoiding overlaps)
//   - Corridor routing (respecting length/bend constraints)
//   - Spatial optimization (compact, playable layouts)
//
// Available implementations:
//   - "force_directed" (ForceDirectedEmbedder): Physics simulation, organic layouts
//   - "orthogonal" (OrthogonalEmbedder): Grid-based, Manhattan corridors, roguelike style
//
// Embedders must be deterministic: given the same graph and RNG state,
// they must produce identical layouts.
type Embedder interface {
	// Embed takes a graph and produces a spatial layout.
	// The RNG must be used for all randomness to ensure determinism.
	// Returns an error if a valid embedding cannot be found.
	Embed(g *graph.Graph, rng *rng.RNG) (*Layout, error)

	// Name returns the identifier for this embedder algorithm.
	Name() string
}

// Config holds spatial embedding constraints and parameters.
type Config struct {
	// MaxIterations is the maximum number of layout iterations
	MaxIterations int

	// CorridorMaxLength is the maximum corridor length in grid units
	CorridorMaxLength float64

	// CorridorMaxBends is the maximum number of bends allowed in a corridor
	CorridorMaxBends int

	// MinRoomSpacing is the minimum gap between room bounding boxes
	MinRoomSpacing float64

	// GridQuantization determines grid cell size (0 = continuous, >0 = snap to grid)
	GridQuantization float64

	// ForceDirected specific parameters
	SpringConstant     float64 // Attraction strength for connected rooms
	RepulsionConstant  float64 // Repulsion strength for all rooms
	DampingFactor      float64 // Movement damping (0.0-1.0)
	StabilityThreshold float64 // Stop when max movement < threshold
	InitialSpread      float64 // Initial random placement spread
}

// DefaultConfig returns a config with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		MaxIterations:      500,
		CorridorMaxLength:  50.0,
		CorridorMaxBends:   4,
		MinRoomSpacing:     2.0,
		GridQuantization:   1.0,
		SpringConstant:     0.5,
		RepulsionConstant:  500.0,
		DampingFactor:      0.8,
		StabilityThreshold: 0.1,
		InitialSpread:      100.0,
	}
}

// Validate checks if the config has valid values.
func (c *Config) Validate() error {
	if c.MaxIterations <= 0 {
		return fmt.Errorf("MaxIterations must be > 0, got %d", c.MaxIterations)
	}
	if c.CorridorMaxLength <= 0 {
		return fmt.Errorf("CorridorMaxLength must be > 0, got %f", c.CorridorMaxLength)
	}
	if c.CorridorMaxBends < 0 {
		return fmt.Errorf("CorridorMaxBends must be >= 0, got %d", c.CorridorMaxBends)
	}
	if c.MinRoomSpacing < 0 {
		return fmt.Errorf("MinRoomSpacing must be >= 0, got %f", c.MinRoomSpacing)
	}
	if c.GridQuantization < 0 {
		return fmt.Errorf("GridQuantization must be >= 0, got %f", c.GridQuantization)
	}
	if c.DampingFactor < 0 || c.DampingFactor > 1 {
		return fmt.Errorf("DampingFactor must be in [0, 1], got %f", c.DampingFactor)
	}
	if c.StabilityThreshold < 0 {
		return fmt.Errorf("StabilityThreshold must be >= 0, got %f", c.StabilityThreshold)
	}
	return nil
}

// Registry holds registered embedder implementations.
var registry = make(map[string]func(*Config) Embedder)

// Register adds an embedder factory to the registry.
// The factory function takes a Config and returns an Embedder instance.
func Register(name string, factory func(*Config) Embedder) {
	if factory == nil {
		panic(fmt.Sprintf("embedder: Register factory for %s is nil", name))
	}
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("embedder: Register called twice for %s", name))
	}
	registry[name] = factory
}

// Get retrieves an embedder by name and initializes it with the given config.
// Returns an error if the embedder is not registered.
func Get(name string, config *Config) (Embedder, error) {
	factory, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("embedder %q not registered", name)
	}

	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return factory(config), nil
}

// List returns the names of all registered embedders.
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// SizeToGridDimensions converts a graph.RoomSize to grid dimensions.
// This provides default dimensions based on the abstract size class.
func SizeToGridDimensions(size graph.RoomSize) (width, height int) {
	switch size {
	case graph.SizeXS:
		return 3, 3 // Tiny corridor
	case graph.SizeS:
		return 5, 5 // Small chamber
	case graph.SizeM:
		return 8, 8 // Medium hall
	case graph.SizeL:
		return 12, 12 // Large room
	case graph.SizeXL:
		return 16, 16 // Boss arena
	default:
		return 8, 8 // Default to medium
	}
}

// ValidateEmbedding performs spatial constraint validation on a layout.
// It checks for overlaps, corridor feasibility, and other spatial requirements.
func ValidateEmbedding(layout *Layout, g *graph.Graph, config *Config) error {
	// First validate layout structure
	if err := layout.Validate(g); err != nil {
		return err
	}

	// Check corridor constraints
	for connID, path := range layout.CorridorPaths {
		// Check length constraint
		length := path.Length()
		if length > config.CorridorMaxLength {
			return fmt.Errorf("corridor %s exceeds max length: %.1f > %.1f",
				connID, length, config.CorridorMaxLength)
		}

		// Check bend constraint
		bends := path.BendCount()
		if bends > config.CorridorMaxBends {
			return fmt.Errorf("corridor %s exceeds max bends: %d > %d",
				connID, bends, config.CorridorMaxBends)
		}
	}

	// Check minimum room spacing if configured
	if config.MinRoomSpacing > 0 {
		rooms := make([]*Pose, 0, len(layout.Poses))
		roomIDs := make([]string, 0, len(layout.Poses))
		for id, pose := range layout.Poses {
			rooms = append(rooms, pose)
			roomIDs = append(roomIDs, id)
		}

		for i := 0; i < len(rooms); i++ {
			for j := i + 1; j < len(rooms); j++ {
				spacing := minSpacing(rooms[i], rooms[j])
				if spacing < config.MinRoomSpacing {
					return fmt.Errorf("rooms %s and %s too close: spacing %.1f < %.1f",
						roomIDs[i], roomIDs[j], spacing, config.MinRoomSpacing)
				}
			}
		}
	}

	return nil
}

// minSpacing calculates the minimum distance between two room bounding boxes.
// Returns 0 if they overlap, otherwise returns the minimum gap.
func minSpacing(p1, p2 *Pose) float64 {
	minX1, minY1, maxX1, maxY1 := p1.Bounds()
	minX2, minY2, maxX2, maxY2 := p2.Bounds()

	// Calculate horizontal and vertical distances
	var dx, dy float64

	if maxX1 <= minX2 {
		dx = minX2 - maxX1
	} else if maxX2 <= minX1 {
		dx = minX1 - maxX2
	} else {
		dx = 0 // Overlapping in X
	}

	if maxY1 <= minY2 {
		dy = minY2 - maxY1
	} else if maxY2 <= minY1 {
		dy = minY1 - maxY2
	} else {
		dy = 0 // Overlapping in Y
	}

	// If overlapping in both dimensions, spacing is 0
	if dx == 0 && dy == 0 {
		return 0
	}

	// Return minimum of the two distances (for axis-aligned spacing)
	if dx == 0 {
		return dy
	}
	if dy == 0 {
		return dx
	}

	// Both have gaps, return minimum
	if dx < dy {
		return dx
	}
	return dy
}
