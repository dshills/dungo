package content

import (
	"fmt"
	"sort"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
	"github.com/dshills/dungo/pkg/themes"
)

// enemyTable maps difficulty ranges to enemy types.
// This is a simple lookup table; in a full implementation,
// this would come from a theme pack or configuration.
var enemyTable = map[string]struct {
	minDifficulty float64
	maxDifficulty float64
}{
	"rat":      {0.0, 0.2},
	"spider":   {0.1, 0.3},
	"skeleton": {0.2, 0.5},
	"goblin":   {0.3, 0.6},
	"orc":      {0.5, 0.8},
	"troll":    {0.7, 0.9},
	"dragon":   {0.8, 1.0},
}

// spawnEnemies places enemy spawns in rooms based on their difficulty values.
// Respects capacity limits and distributes enemies proportionally to difficulty.
//
// Algorithm:
//  1. Skip Start, Boss (special handling), Treasure, Vendor, Shrine rooms
//  2. For each eligible room, calculate enemy count from difficulty
//  3. Select enemy type(s) matching difficulty range (using theme pack if available)
//  4. Place spawn points with dummy positions (actual positions require layout)
//  5. Respect maxEnemiesPerRoom capacity limit
//
// Theme Integration:
//   - If room has "biome" tag, load corresponding theme pack
//   - Use theme's encounter table for difficulty-based enemy selection
//   - Fall back to default enemy table if theme not found or no biome tag
func spawnEnemies(g *graph.Graph, content *Content, maxEnemiesPerRoom int, rng *rng.RNG) error {
	return spawnEnemiesWithThemes(g, content, maxEnemiesPerRoom, rng, nil)
}

// spawnEnemiesWithThemes is the internal implementation that supports theme pack injection.
// themeLoader can be nil for default behavior.
func spawnEnemiesWithThemes(g *graph.Graph, content *Content, maxEnemiesPerRoom int, rng *rng.RNG, themeLoader *themes.Loader) error {
	spawnID := 0

	// Sort room IDs for deterministic iteration
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	for _, roomID := range roomIDs {
		room := g.Rooms[roomID]

		// Skip rooms where enemies don't belong
		if shouldSkipEnemyPlacement(room) {
			continue
		}

		// Calculate enemy count based on difficulty
		// difficulty 0.0 = 0 enemies, difficulty 1.0 = maxEnemiesPerRoom
		enemyCount := int(room.Difficulty * float64(maxEnemiesPerRoom))
		if enemyCount == 0 && room.Difficulty > 0.0 {
			enemyCount = 1 // At least 1 enemy if room has any difficulty
		}

		// Cap at maximum
		if enemyCount > maxEnemiesPerRoom {
			enemyCount = maxEnemiesPerRoom
		}

		// Skip rooms with no enemies - don't create invalid spawns
		if enemyCount == 0 {
			continue
		}

		// Select enemy type based on difficulty (with theme support)
		enemyType := selectEnemyTypeWithTheme(room, rng, themeLoader)

		// Create spawn point
		// Position is placeholder (0,0) - actual position requires layout stage
		spawn := Spawn{
			ID:         fmt.Sprintf("spawn_%d", spawnID),
			RoomID:     roomID,
			Position:   Point{X: 0, Y: 0}, // Placeholder - needs layout
			EnemyType:  enemyType,
			Count:      enemyCount,
			PatrolPath: nil, // Can be added later based on room layout
		}

		if err := spawn.Validate(); err != nil {
			return fmt.Errorf("invalid spawn: %w", err)
		}

		content.Spawns = append(content.Spawns, spawn)
		spawnID++
	}

	return nil
}

// shouldSkipEnemyPlacement determines if a room should not have enemy spawns.
func shouldSkipEnemyPlacement(room *graph.Room) bool {
	switch room.Archetype {
	case graph.ArchetypeStart:
		return true // Start room should be safe
	case graph.ArchetypeTreasure:
		return true // Treasure rooms are rewards, not combat
	case graph.ArchetypeVendor:
		return true // Vendors are safe zones
	case graph.ArchetypeShrine:
		return true // Shrines are safe zones
	case graph.ArchetypeCheckpoint:
		return true // Checkpoints are safe zones
	default:
		return false
	}
}

// selectEnemyTypeWithTheme chooses an enemy type appropriate for the room's difficulty.
// Attempts to use theme pack if room has biome tag, falls back to default table.
func selectEnemyTypeWithTheme(room *graph.Room, rng *rng.RNG, themeLoader *themes.Loader) string {
	// Try theme-based selection if loader available
	if themeLoader != nil && room.Tags != nil {
		if biome, ok := room.Tags["biome"]; ok && biome != "" {
			// Load theme pack for this biome
			themePack, err := themeLoader.Load(biome)
			if err == nil && themePack != nil {
				// Use theme adapter for selection
				enemyType := themes.SelectEncounterFromTheme(themePack, room.Difficulty, rng)
				if enemyType != "" {
					return enemyType
				}
			}
			// Theme load failed or no matching table - fall through to default
		}
	}

	// Fall back to default enemy selection
	return selectEnemyType(room.Difficulty, rng)
}

// selectEnemyType chooses an enemy type appropriate for the given difficulty.
// Uses weighted random selection from enemies matching the difficulty range.
// This is the default fallback when no theme pack is available.
func selectEnemyType(difficulty float64, rng *rng.RNG) string {
	// Sort enemy types for deterministic iteration
	enemyTypes := make([]string, 0, len(enemyTable))
	for et := range enemyTable {
		enemyTypes = append(enemyTypes, et)
	}
	sort.Strings(enemyTypes)

	// Build list of eligible enemies
	eligible := make([]string, 0, len(enemyTable))
	weights := make([]float64, 0, len(enemyTable))

	for _, enemyType := range enemyTypes {
		diffRange := enemyTable[enemyType]
		// Check if this enemy type is appropriate for this difficulty
		if difficulty >= diffRange.minDifficulty && difficulty <= diffRange.maxDifficulty {
			eligible = append(eligible, enemyType)

			// Weight enemies toward center of their range
			center := (diffRange.minDifficulty + diffRange.maxDifficulty) / 2
			dist := (difficulty - center)
			if dist < 0 {
				dist = -dist
			}
			weight := 1.0 - dist
			if weight < 0.1 {
				weight = 0.1
			}
			weights = append(weights, weight)
		}
	}

	// If no eligible enemies (shouldn't happen with our table), default to skeleton
	if len(eligible) == 0 {
		return "skeleton"
	}

	// Use weighted random selection
	index := rng.WeightedChoice(weights)
	if index < 0 || index >= len(eligible) {
		return eligible[0]
	}

	return eligible[index]
}

// placePuzzles places puzzle instances in puzzle rooms.
// Each puzzle room gets one puzzle matching its difficulty.
func placePuzzles(g *graph.Graph, content *Content, rng *rng.RNG) error {
	puzzleID := 0

	// Sort room IDs for deterministic iteration
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	for _, roomID := range roomIDs {
		room := g.Rooms[roomID]
		if room.Archetype != graph.ArchetypePuzzle {
			continue
		}

		// Select puzzle type based on difficulty
		puzzleType := selectPuzzleType(room.Difficulty, rng)

		// Create puzzle instance
		puzzle := PuzzleInstance{
			ID:           fmt.Sprintf("puzzle_%d", puzzleID),
			RoomID:       roomID,
			Type:         puzzleType,
			Requirements: make([]Requirement, 0),
			Provides:     make([]Capability, 0),
			Difficulty:   room.Difficulty,
		}

		// If room provides capabilities, add them to puzzle
		for _, cap := range room.Provides {
			puzzle.Provides = append(puzzle.Provides, Capability{
				Type:  cap.Type,
				Value: cap.Value,
			})
		}

		// If room has requirements, add them to puzzle
		for _, req := range room.Requirements {
			puzzle.Requirements = append(puzzle.Requirements, Requirement{
				Type:  req.Type,
				Value: req.Value,
			})
		}

		if err := puzzle.Validate(); err != nil {
			return fmt.Errorf("invalid puzzle: %w", err)
		}

		content.Puzzles = append(content.Puzzles, puzzle)
		puzzleID++
	}

	return nil
}

// puzzleTypes maps difficulty ranges to puzzle types.
var puzzleTypes = []struct {
	name          string
	minDifficulty float64
	maxDifficulty float64
}{
	{"lever", 0.0, 0.3},
	{"rune_sequence", 0.2, 0.6},
	{"pressure_plate", 0.3, 0.7},
	{"light_beam", 0.5, 0.9},
	{"cipher", 0.7, 1.0},
}

// selectPuzzleType chooses a puzzle type appropriate for the given difficulty.
func selectPuzzleType(difficulty float64, rng *rng.RNG) string {
	eligible := make([]string, 0, len(puzzleTypes))
	weights := make([]float64, 0, len(puzzleTypes))

	for _, pt := range puzzleTypes {
		if difficulty >= pt.minDifficulty && difficulty <= pt.maxDifficulty {
			eligible = append(eligible, pt.name)

			// Weight puzzles toward center of their range
			center := (pt.minDifficulty + pt.maxDifficulty) / 2
			dist := difficulty - center
			if dist < 0 {
				dist = -dist
			}
			weight := 1.0 - dist
			if weight < 0.1 {
				weight = 0.1
			}
			weights = append(weights, weight)
		}
	}

	if len(eligible) == 0 {
		return "lever" // Default fallback
	}

	index := rng.WeightedChoice(weights)
	if index < 0 || index >= len(eligible) {
		return eligible[0]
	}

	return eligible[index]
}
