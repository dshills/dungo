package dungeon

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/dshills/dungo/pkg/graph"
)

// ErrNotImplemented is returned by export methods that are not yet implemented.
var ErrNotImplemented = errors.New("functionality not yet implemented")

// Requirement is an alias for graph.Requirement for convenience.
type Requirement = graph.Requirement

// Capability is an alias for graph.Capability for convenience.
type Capability = graph.Capability

// Artifact is the complete output of dungeon generation.
// It contains all pipeline outputs: graph structure, spatial layout,
// rasterized tiles, gameplay content, metrics, and optional debug data.
//
// An Artifact represents a fully-generated, validated dungeon ready for use.
// It can be serialized to JSON, exported to Tiled TMJ format, or processed
// by game engines. All data is deterministic based on the input Config seed.
//
// Fields:
//
//	ADG - Abstract Dungeon Graph (room/connector topology)
//	Layout - Spatial embedding (2D positions, corridors)
//	TileMap - Rasterized tile grid (for rendering)
//	Content - Gameplay elements (enemies, loot, puzzles)
//	Metrics - Quality measurements (branching, path length, etc.)
//	Debug - Optional validation reports and visualizations
type Artifact struct {
	ADG     *Graph
	Layout  *Layout
	TileMap *TileMap
	Content *Content
	Metrics *Metrics
	Debug   *DebugArtifacts
}

// Point represents a 2D coordinate.
type Point struct {
	X, Y int
}

// Rect represents a rectangular bounds.
type Rect struct {
	X, Y, Width, Height int
}

// Pose represents the spatial position and orientation of a room.
type Pose struct {
	X           int
	Y           int
	Rotation    int    // Degrees: 0, 90, 180, 270
	FootprintID string // Reference to room template shape
}

// Path represents a polyline path (for corridors).
type Path struct {
	Points []Point
}

// Layout contains the spatial embedding of the dungeon graph.
// It maps abstract rooms to concrete positions and corridors to paths.
type Layout struct {
	Poses         map[string]Pose // Room ID → position/rotation
	CorridorPaths map[string]Path // Connector ID → polyline path
	Bounds        Rect            // Overall dungeon extents
}

// TileMap is the rasterized dungeon with layered tiles.
// It represents the final 2D grid representation of the dungeon.
type TileMap struct {
	Width      int               // Grid width in tiles
	Height     int               // Grid height in tiles
	TileWidth  int               // Tile width in pixels
	TileHeight int               // Tile height in pixels
	Layers     map[string]*Layer // Named tile/object layers
}

// Layer represents a single layer in the tile map.
// Can be either a tile layer (grid) or an object layer (entities).
type Layer struct {
	ID      int      // Layer identifier
	Name    string   // Layer name
	Type    string   // "tilelayer" or "objectgroup"
	Visible bool     // Visibility flag
	Opacity float64  // Layer opacity (0.0-1.0)
	Data    []uint32 // Tile data (for tile layers), flat row-major array
	Objects []Object // Objects (for object layers)
}

// Object represents an entity in an object layer.
type Object struct {
	ID         int                    // Object identifier
	Name       string                 // Object name
	Type       string                 // Object type
	X          float64                // X position
	Y          float64                // Y position
	Width      float64                // Object width
	Height     float64                // Object height
	Rotation   float64                // Rotation in degrees
	GID        uint32                 // Tile GID (if tile object)
	Visible    bool                   // Visibility flag
	Properties map[string]interface{} // Custom properties
}

// Content contains all gameplay elements placed in the dungeon.
type Content struct {
	Spawns  []Spawn          // Enemy spawn points
	Loot    []Loot           // Treasure items
	Puzzles []PuzzleInstance // Interactive puzzles
	Secrets []SecretInstance // Hidden elements
}

// Spawn represents an enemy spawn point.
type Spawn struct {
	ID         string  // Unique identifier
	RoomID     string  // Parent room
	Position   Point   // Location within room
	EnemyType  string  // Reference to encounter table entry
	Count      int     // Number of enemies (1-10)
	PatrolPath []Point // Optional waypoints
}

// Loot represents a treasure item.
type Loot struct {
	ID       string // Unique identifier
	RoomID   string // Parent room
	Position Point  // Location within room
	ItemType string // Reference to loot table entry
	Value    int    // Gold/experience worth
	Required bool   // Required for progression (keys, abilities)
}

// PuzzleInstance represents an interactive challenge.
type PuzzleInstance struct {
	ID           string        // Unique identifier
	RoomID       string        // Parent room
	Type         string        // Puzzle mechanism ("lever", "rune", "statue")
	Requirements []Requirement // Prerequisites
	Provides     []Capability  // Grants on solve
	Difficulty   float64       // Challenge rating (0.0-1.0)
}

// SecretInstance represents a hidden element.
type SecretInstance struct {
	ID       string   // Unique identifier
	RoomID   string   // Parent room
	Type     string   // Secret type
	Position Point    // Location within room
	Clues    []string // Hints for discovery
}

// Metrics contains generation statistics and measurements.
type Metrics struct {
	BranchingFactor   float64 // Actual average connections per room
	PathLength        int     // Start→Boss path length
	CycleCount        int     // Number of graph cycles
	PacingDeviation   float64 // L2 distance from target difficulty curve
	SecretFindability float64 // Heuristic score (0.0-1.0)
}

// DebugArtifacts contains optional debug outputs.
// These are generated when debug mode is enabled in the configuration.
type DebugArtifacts struct {
	ADGSVG    []byte            // SVG visualization of graph
	LayoutPNG []byte            // Heatmap overlay image
	Report    *ValidationReport // Detailed validation metrics
}

// ValidationReport contains validation results and constraint satisfaction.
type ValidationReport struct {
	Passed                bool               // All hard constraints satisfied
	HardConstraintResults []ConstraintResult // Individual hard constraint results
	SoftConstraintResults []ConstraintResult // Individual soft constraint results
	Metrics               *Metrics           // Calculated statistics
	Warnings              []string           // Non-fatal issues
	Errors                []string           // Hard constraint failures
}

// ConstraintResult represents the result of evaluating a single constraint.
type ConstraintResult struct {
	Constraint *Constraint // The constraint that was checked
	Satisfied  bool        // Pass/fail
	Score      float64     // For soft constraints (0.0-1.0, higher is better)
	Details    string      // Explanation or violation info
}

// ExportJSON serializes the artifact to JSON with indentation.
// Returns formatted JSON with 2-space indentation for readability.
// This is a convenience method for export functionality.
func (a *Artifact) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(a, "", "  ")
}

// ExportJSONCompact serializes the artifact to JSON without indentation.
// Returns compact JSON suitable for storage or transmission.
func (a *Artifact) ExportJSONCompact() ([]byte, error) {
	return json.Marshal(a)
}

// SaveJSON exports the artifact to a JSON file with indentation.
// The file is created with 0644 permissions (readable by all, writable by owner).
func (a *Artifact) SaveJSON(path string) error {
	data, err := a.ExportJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// SaveJSONCompact exports the artifact to a compact JSON file.
// The file is created with 0644 permissions (readable by all, writable by owner).
func (a *Artifact) SaveJSONCompact(path string) error {
	data, err := a.ExportJSONCompact()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ExportTMJ exports the artifact to Tiled TMJ (JSON map) format.
// Note: TMJ export functionality is not yet implemented.
// This method is a placeholder for future TMJ export support.
func (a *Artifact) ExportTMJ() ([]byte, error) {
	// TODO: Implement TMJ export in pkg/export/tmj.go (tasks T100-T104)
	return nil, ErrNotImplemented
}

// SaveTMJ exports the artifact to a TMJ file.
// Note: TMJ export functionality is not yet implemented.
// This method is a placeholder for future TMJ export support.
func (a *Artifact) SaveTMJ(path string) error {
	// TODO: Implement TMJ export in pkg/export/tmj.go (tasks T100-T104)
	return ErrNotImplemented
}
