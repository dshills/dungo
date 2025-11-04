package graph

import "fmt"

// RoomArchetype defines the type of room in the dungeon.
type RoomArchetype int

const (
	ArchetypeStart RoomArchetype = iota
	ArchetypeBoss
	ArchetypeTreasure
	ArchetypePuzzle
	ArchetypeHub
	ArchetypeCorridor
	ArchetypeSecret
	ArchetypeOptional
	ArchetypeVendor
	ArchetypeShrine
	ArchetypeCheckpoint
)

// String returns the string representation of a RoomArchetype.
func (r RoomArchetype) String() string {
	switch r {
	case ArchetypeStart:
		return "Start"
	case ArchetypeBoss:
		return "Boss"
	case ArchetypeTreasure:
		return "Treasure"
	case ArchetypePuzzle:
		return "Puzzle"
	case ArchetypeHub:
		return "Hub"
	case ArchetypeCorridor:
		return "Corridor"
	case ArchetypeSecret:
		return "Secret"
	case ArchetypeOptional:
		return "Optional"
	case ArchetypeVendor:
		return "Vendor"
	case ArchetypeShrine:
		return "Shrine"
	case ArchetypeCheckpoint:
		return "Checkpoint"
	default:
		return fmt.Sprintf("Unknown(%d)", r)
	}
}

// RoomSize defines the abstract size class of a room.
type RoomSize int

const (
	SizeXS RoomSize = iota // Tiny corridor
	SizeS                  // Small chamber
	SizeM                  // Medium hall
	SizeL                  // Large room
	SizeXL                 // Boss arena
)

// String returns the string representation of a RoomSize.
func (s RoomSize) String() string {
	switch s {
	case SizeXS:
		return "XS"
	case SizeS:
		return "S"
	case SizeM:
		return "M"
	case SizeL:
		return "L"
	case SizeXL:
		return "XL"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// Requirement represents a prerequisite to enter a room.
type Requirement struct {
	Type  string `json:"type"`  // "key", "ability", "item"
	Value string `json:"value"` // Specific requirement (e.g., "silver_key", "double_jump")
}

// Capability represents what a room grants to the player.
type Capability struct {
	Type  string `json:"type"`  // "key", "ability", "item"
	Value string `json:"value"` // What's provided
}

// Room represents a node in the Abstract Dungeon Graph.
type Room struct {
	ID           string            `json:"id"`
	Archetype    RoomArchetype     `json:"archetype"`
	Size         RoomSize          `json:"size"`
	Tags         map[string]string `json:"tags,omitempty"`
	Difficulty   float64           `json:"difficulty"` // 0.0-1.0
	Reward       float64           `json:"reward"`     // 0.0-1.0
	Requirements []Requirement     `json:"requirements,omitempty"`
	Provides     []Capability      `json:"provides,omitempty"`
	DegreeMin    *int              `json:"degreeMin,omitempty"` // Optional connection bounds
	DegreeMax    *int              `json:"degreeMax,omitempty"`
}

// Validate checks if the room data is valid.
func (r *Room) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("room ID cannot be empty")
	}

	if r.Difficulty < 0.0 || r.Difficulty > 1.0 {
		return fmt.Errorf("room %s: difficulty must be in [0.0, 1.0], got %f", r.ID, r.Difficulty)
	}

	if r.Reward < 0.0 || r.Reward > 1.0 {
		return fmt.Errorf("room %s: reward must be in [0.0, 1.0], got %f", r.ID, r.Reward)
	}

	if r.DegreeMin != nil && r.DegreeMax != nil {
		if *r.DegreeMin > *r.DegreeMax {
			return fmt.Errorf("room %s: DegreeMin (%d) must be <= DegreeMax (%d)", r.ID, *r.DegreeMin, *r.DegreeMax)
		}
	}

	return nil
}

// String returns a human-readable representation of the Room.
func (r *Room) String() string {
	return fmt.Sprintf("Room[%s: %s %s, Difficulty=%.2f, Reward=%.2f]",
		r.ID, r.Archetype, r.Size, r.Difficulty, r.Reward)
}
