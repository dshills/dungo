package content

import "fmt"

// Point represents a 2D position in tile coordinates.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// String returns a human-readable representation of a Point.
func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// Spawn represents an enemy spawn point in a room.
// Enemies are placed based on room.Difficulty values.
type Spawn struct {
	ID         string  `json:"id"`         // Unique spawn identifier
	RoomID     string  `json:"roomId"`     // Room containing this spawn
	Position   Point   `json:"position"`   // Spawn location in tile coords
	EnemyType  string  `json:"enemyType"`  // Type of enemy to spawn
	Count      int     `json:"count"`      // Number of enemies at this spawn
	PatrolPath []Point `json:"patrolPath"` // Optional patrol waypoints
}

// String returns a human-readable representation of a Spawn.
func (s Spawn) String() string {
	return fmt.Sprintf("Spawn[%s: %dx %s in %s at %s]",
		s.ID, s.Count, s.EnemyType, s.RoomID, s.Position)
}

// Validate checks if the spawn data is valid.
func (s *Spawn) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("spawn ID cannot be empty")
	}
	if s.RoomID == "" {
		return fmt.Errorf("spawn %s: RoomID cannot be empty", s.ID)
	}
	if s.EnemyType == "" {
		return fmt.Errorf("spawn %s: EnemyType cannot be empty", s.ID)
	}
	if s.Count <= 0 {
		return fmt.Errorf("spawn %s: Count must be > 0, got %d", s.ID, s.Count)
	}
	return nil
}

// Loot represents an item pickup in a room.
// Loot is distributed based on room.Reward values.
type Loot struct {
	ID       string `json:"id"`       // Unique loot identifier
	RoomID   string `json:"roomId"`   // Room containing this loot
	Position Point  `json:"position"` // Loot location in tile coords
	ItemType string `json:"itemType"` // Type of item (e.g., "gold", "potion", "key")
	Value    int    `json:"value"`    // Treasure value or quantity
	Required bool   `json:"required"` // Whether this is required for progression (e.g., keys)
}

// String returns a human-readable representation of Loot.
func (l Loot) String() string {
	required := ""
	if l.Required {
		required = " [REQUIRED]"
	}
	return fmt.Sprintf("Loot[%s: %s (value=%d) in %s at %s%s]",
		l.ID, l.ItemType, l.Value, l.RoomID, l.Position, required)
}

// Validate checks if the loot data is valid.
func (l *Loot) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("loot ID cannot be empty")
	}
	if l.RoomID == "" {
		return fmt.Errorf("loot %s: RoomID cannot be empty", l.ID)
	}
	if l.ItemType == "" {
		return fmt.Errorf("loot %s: ItemType cannot be empty", l.ID)
	}
	if l.Value < 0 {
		return fmt.Errorf("loot %s: Value must be >= 0, got %d", l.ID, l.Value)
	}
	return nil
}

// PuzzleInstance represents a puzzle encounter in a room.
// Puzzles are typically placed in rooms with Archetype == ArchetypePuzzle.
type PuzzleInstance struct {
	ID           string        `json:"id"`           // Unique puzzle identifier
	RoomID       string        `json:"roomId"`       // Room containing this puzzle
	Type         string        `json:"type"`         // Puzzle type (e.g., "lever", "rune_sequence")
	Requirements []Requirement `json:"requirements"` // What's needed to solve
	Provides     []Capability  `json:"provides"`     // What solving grants
	Difficulty   float64       `json:"difficulty"`   // Puzzle difficulty 0.0-1.0
}

// Requirement represents a prerequisite to solve a puzzle or enter a room.
type Requirement struct {
	Type  string `json:"type"`  // "key", "ability", "item"
	Value string `json:"value"` // Specific requirement
}

// Capability represents what a puzzle or room provides.
type Capability struct {
	Type  string `json:"type"`  // "key", "ability", "item"
	Value string `json:"value"` // What's provided
}

// String returns a human-readable representation of a PuzzleInstance.
func (p PuzzleInstance) String() string {
	return fmt.Sprintf("Puzzle[%s: %s in %s, difficulty=%.2f]",
		p.ID, p.Type, p.RoomID, p.Difficulty)
}

// Validate checks if the puzzle data is valid.
func (p *PuzzleInstance) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("puzzle ID cannot be empty")
	}
	if p.RoomID == "" {
		return fmt.Errorf("puzzle %s: RoomID cannot be empty", p.ID)
	}
	if p.Type == "" {
		return fmt.Errorf("puzzle %s: Type cannot be empty", p.ID)
	}
	if p.Difficulty < 0.0 || p.Difficulty > 1.0 {
		return fmt.Errorf("puzzle %s: Difficulty must be in [0.0, 1.0], got %f", p.ID, p.Difficulty)
	}
	return nil
}

// SecretInstance represents a hidden discovery in a room.
// Secrets are placed in secret rooms or as hidden elements.
type SecretInstance struct {
	ID       string   `json:"id"`       // Unique secret identifier
	RoomID   string   `json:"roomId"`   // Room containing this secret
	Type     string   `json:"type"`     // Secret type (e.g., "hidden_door", "treasure")
	Position Point    `json:"position"` // Secret location in tile coords
	Clues    []string `json:"clues"`    // Hints to find this secret
}

// String returns a human-readable representation of a SecretInstance.
func (s SecretInstance) String() string {
	return fmt.Sprintf("Secret[%s: %s in %s at %s, %d clues]",
		s.ID, s.Type, s.RoomID, s.Position, len(s.Clues))
}

// Validate checks if the secret data is valid.
func (s *SecretInstance) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("secret ID cannot be empty")
	}
	if s.RoomID == "" {
		return fmt.Errorf("secret %s: RoomID cannot be empty", s.ID)
	}
	if s.Type == "" {
		return fmt.Errorf("secret %s: Type cannot be empty", s.ID)
	}
	return nil
}
