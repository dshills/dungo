package content

import (
	"context"
	"fmt"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// ContentPass populates rooms with gameplay elements: enemies, loot, and puzzles.
// This is the FOURTH stage in the dungeon generation pipeline.
//
// Content placement logic:
//  1. Find all key-locked connectors
//  2. Place keys in rooms before locks (on path from Start)
//  3. Distribute treasure based on room.Reward values
//  4. Spawn enemies based on room.Difficulty values
//  5. Respect capacity limits (e.g., max 10 enemies per room)
//
// The ContentPass uses room properties from the graph (Difficulty, Reward, Archetype)
// to make placement decisions. Keys are placed to satisfy the key-before-lock
// constraint automatically. Enemy spawns scale with difficulty progression.
//
// The pass uses the provided RNG to ensure deterministic placement
// that is reproducible given the same seed.
type ContentPass interface {
	// Place populates the graph with content based on room properties.
	// Returns a Content container with all placed spawns, loot, and puzzles.
	// The RNG must be used for all randomness to ensure determinism.
	Place(ctx context.Context, g *graph.Graph, rng *rng.RNG) (*Content, error)
}

// Content is the complete output of content placement.
// It contains all gameplay elements distributed throughout the dungeon.
type Content struct {
	Spawns  []Spawn          `json:"spawns"`  // Enemy spawn points
	Loot    []Loot           `json:"loot"`    // Item pickups
	Puzzles []PuzzleInstance `json:"puzzles"` // Puzzle encounters
	Secrets []SecretInstance `json:"secrets"` // Hidden discoveries
}

// NewContent creates an empty Content container.
func NewContent() *Content {
	return &Content{
		Spawns:  make([]Spawn, 0),
		Loot:    make([]Loot, 0),
		Puzzles: make([]PuzzleInstance, 0),
		Secrets: make([]SecretInstance, 0),
	}
}

// Validate checks if the content placement is valid.
func (c *Content) Validate(g *graph.Graph) error {
	// Check that all spawns reference valid rooms
	for _, spawn := range c.Spawns {
		if _, exists := g.Rooms[spawn.RoomID]; !exists {
			return fmt.Errorf("spawn %s references non-existent room %s", spawn.ID, spawn.RoomID)
		}
	}

	// Check that all loot references valid rooms
	for _, loot := range c.Loot {
		if _, exists := g.Rooms[loot.RoomID]; !exists {
			return fmt.Errorf("loot %s references non-existent room %s", loot.ID, loot.RoomID)
		}
	}

	// Check that all puzzles reference valid rooms
	for _, puzzle := range c.Puzzles {
		if _, exists := g.Rooms[puzzle.RoomID]; !exists {
			return fmt.Errorf("puzzle %s references non-existent room %s", puzzle.ID, puzzle.RoomID)
		}
	}

	// Check that all secrets reference valid rooms
	for _, secret := range c.Secrets {
		if _, exists := g.Rooms[secret.RoomID]; !exists {
			return fmt.Errorf("secret %s references non-existent room %s", secret.ID, secret.RoomID)
		}
	}

	return nil
}

// String returns a human-readable summary of content.
func (c *Content) String() string {
	return fmt.Sprintf("Content[Spawns=%d, Loot=%d, Puzzles=%d, Secrets=%d]",
		len(c.Spawns), len(c.Loot), len(c.Puzzles), len(c.Secrets))
}

// DefaultContentPass is the standard implementation of ContentPass.
// It places content in a balanced way based on room properties.
type DefaultContentPass struct {
	maxEnemiesPerRoom int  // Capacity limit for enemies
	lootBudgetBase    int  // Base treasure value
	keyPlacementFirst bool // Whether to place keys before general loot
}

// NewDefaultContentPass creates a DefaultContentPass with default settings.
func NewDefaultContentPass() *DefaultContentPass {
	return &DefaultContentPass{
		maxEnemiesPerRoom: 10,
		lootBudgetBase:    1000,
		keyPlacementFirst: true,
	}
}

// Place implements ContentPass by orchestrating all content placement.
func (d *DefaultContentPass) Place(ctx context.Context, g *graph.Graph, rng *rng.RNG) (*Content, error) {
	content := NewContent()

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Step 1: Find key-locked connectors and place keys before locks
	if d.keyPlacementFirst {
		if err := placeRequiredKeys(g, content, rng); err != nil {
			return nil, fmt.Errorf("placing required keys: %w", err)
		}
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Step 2: Distribute treasure based on room.Reward values
	if err := distributeLoot(g, content, d.lootBudgetBase, rng); err != nil {
		return nil, fmt.Errorf("distributing loot: %w", err)
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Step 3: Spawn enemies based on room.Difficulty values
	if err := spawnEnemies(g, content, d.maxEnemiesPerRoom, rng); err != nil {
		return nil, fmt.Errorf("spawning enemies: %w", err)
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Step 4: Place puzzles in puzzle rooms
	if err := placePuzzles(g, content, rng); err != nil {
		return nil, fmt.Errorf("placing puzzles: %w", err)
	}

	// Validate the result
	if err := content.Validate(g); err != nil {
		return nil, fmt.Errorf("content validation failed: %w", err)
	}

	return content, nil
}

// WithMaxEnemiesPerRoom sets the capacity limit for enemies in a room.
func (d *DefaultContentPass) WithMaxEnemiesPerRoom(max int) *DefaultContentPass {
	d.maxEnemiesPerRoom = max
	return d
}

// WithLootBudget sets the base loot budget for treasure distribution.
func (d *DefaultContentPass) WithLootBudget(budget int) *DefaultContentPass {
	d.lootBudgetBase = budget
	return d
}
