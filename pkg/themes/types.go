package themes

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// ThemePack represents a complete theme with encounter tables, loot tables,
// tilesets, and decorators for dungeon generation.
//
// Themes provide the visual and gameplay flavor for dungeons. A ThemePack includes:
//   - Tilesets: Graphics and tile mappings for rendering
//   - EncounterTables: Enemy type distributions by difficulty
//   - LootTables: Item drops by room type
//   - Decorators: Environmental decoration rules
//
// Themes are loaded from YAML files and can be mixed/blended in a single dungeon.
// The content placement stage queries themes to determine what enemies and items
// to place based on room difficulty and archetype.
//
// Example themes: "dungeon", "crypt", "forest", "castle", "caves"
type ThemePack struct {
	Name            string           `yaml:"name" json:"name"`
	Description     string           `yaml:"description" json:"description"`
	Tilesets        []Tileset        `yaml:"tilesets" json:"tilesets"`
	EncounterTables []EncounterTable `yaml:"encounter_tables" json:"encounter_tables"`
	LootTables      []LootTable      `yaml:"loot_tables" json:"loot_tables"`
	Decorators      []Decorator      `yaml:"decorators" json:"decorators"`
}

// Tileset defines graphics and tile mappings for rendering.
type Tileset struct {
	Name       string `yaml:"name" json:"name"`
	Path       string `yaml:"path" json:"path"`
	TileWidth  int    `yaml:"tile_width" json:"tile_width"`
	TileHeight int    `yaml:"tile_height" json:"tile_height"`
}

// EncounterTable maps a difficulty level to weighted enemy spawns.
type EncounterTable struct {
	Difficulty float64         `yaml:"difficulty" json:"difficulty"`
	Entries    []WeightedEntry `yaml:"entries" json:"entries"`
}

// LootTable maps a room type to weighted item drops.
type LootTable struct {
	RoomType string          `yaml:"room_type" json:"room_type"`
	Entries  []WeightedEntry `yaml:"entries" json:"entries"`
}

// WeightedEntry represents an entry with a selection weight.
type WeightedEntry struct {
	Type   string `yaml:"type" json:"type"`
	Weight int    `yaml:"weight" json:"weight"`
}

// Decorator defines environmental decoration rules.
type Decorator struct {
	Type    string  `yaml:"type" json:"type"`
	Density float64 `yaml:"density" json:"density"`
}

// LoadThemeFromFile loads a theme pack from a YAML file.
// Returns error if file cannot be read or YAML is invalid.
func LoadThemeFromFile(path string) (*ThemePack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading theme file: %w", err)
	}

	var theme ThemePack
	if err := yaml.Unmarshal(data, &theme); err != nil {
		return nil, fmt.Errorf("parsing theme YAML: %w", err)
	}

	if err := ValidateThemePack(&theme); err != nil {
		return nil, err
	}

	return &theme, nil
}

// LoadThemeFromDirectory loads a theme pack from a directory containing theme.yml.
// Returns error if directory doesn't exist or theme.yml is invalid.
func LoadThemeFromDirectory(dir string) (*ThemePack, error) {
	// Try theme.yml first, then theme.yaml
	themePath := filepath.Join(dir, "theme.yml")
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		themePath = filepath.Join(dir, "theme.yaml")
		if _, err := os.Stat(themePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("theme file not found in directory: %s", dir)
		}
	}

	return LoadThemeFromFile(themePath)
}

// ValidateThemePack checks if a theme pack has all required fields and valid data.
// Returns error describing validation failures.
func ValidateThemePack(theme *ThemePack) error {
	if theme.Name == "" {
		return errors.New("name is required")
	}

	if len(theme.Tilesets) == 0 {
		return errors.New("at least one tileset is required")
	}

	// Validate encounter tables
	for _, table := range theme.EncounterTables {
		if table.Difficulty < 0.0 || table.Difficulty > 1.0 {
			return errors.New("difficulty must be between 0.0 and 1.0")
		}
		for _, entry := range table.Entries {
			if entry.Type == "" {
				return errors.New("encounter entry type is required")
			}
			if entry.Weight <= 0 {
				return errors.New("weight must be positive")
			}
		}
	}

	// Validate loot tables
	for _, table := range theme.LootTables {
		for _, entry := range table.Entries {
			if entry.Type == "" {
				return errors.New("loot entry type is required")
			}
			if entry.Weight <= 0 {
				return errors.New("weight must be positive")
			}
		}
	}

	return nil
}

// GetEncountersForDifficulty returns encounter tables appropriate for the given difficulty.
// Uses nearest-bracket selection: finds the nearest bracket(s) to interpolate between.
// - Exact match: returns that bracket only
// - Between brackets: returns both (lower and upper) for interpolation
// - Below all brackets: returns lowest bracket
// - Above all brackets: returns highest bracket
func (tp *ThemePack) GetEncountersForDifficulty(difficulty float64) []EncounterTable {
	if len(tp.EncounterTables) == 0 {
		return nil
	}

	// Sort tables by difficulty to find nearest brackets
	sorted := make([]EncounterTable, len(tp.EncounterTables))
	copy(sorted, tp.EncounterTables)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Difficulty < sorted[j].Difficulty
	})

	// Find exact match first
	for i, table := range sorted {
		if table.Difficulty == difficulty {
			return []EncounterTable{sorted[i]}
		}
	}

	// Find the two nearest brackets (one below, one above)
	var lower, upper *EncounterTable
	var lowerIdx, upperIdx int = -1, -1

	for i, table := range sorted {
		if table.Difficulty < difficulty {
			lower = &sorted[i]
			lowerIdx = i
		} else if table.Difficulty > difficulty && upper == nil {
			upper = &sorted[i]
			upperIdx = i
			break
		}
	}

	// Edge case: difficulty is below all brackets
	if lower == nil {
		return []EncounterTable{sorted[0]}
	}

	// Edge case: difficulty is above all brackets
	if upper == nil {
		return []EncounterTable{sorted[len(sorted)-1]}
	}

	// Normal case: return both brackets for interpolation
	return []EncounterTable{sorted[lowerIdx], sorted[upperIdx]}
}

// GetLootTableForRoomType returns the loot table for a specific room type.
// Returns nil if no loot table exists for the room type.
func (tp *ThemePack) GetLootTableForRoomType(roomType string) *LootTable {
	for i := range tp.LootTables {
		if tp.LootTables[i].RoomType == roomType {
			return &tp.LootTables[i]
		}
	}
	return nil
}

// SelectWeightedEntry performs weighted random selection from a list of entries.
// Returns nil if entries is empty. Uses provided RNG for deterministic selection.
func SelectWeightedEntry(entries []WeightedEntry, rng *rand.Rand) *WeightedEntry {
	if len(entries) == 0 {
		return nil
	}

	// Calculate total weight
	totalWeight := 0
	for _, entry := range entries {
		totalWeight += entry.Weight
	}

	if totalWeight == 0 {
		return nil
	}

	// Generate random value in [0, totalWeight)
	r := rng.Intn(totalWeight)

	// Find the entry that corresponds to this value
	cumulative := 0
	for i := range entries {
		cumulative += entries[i].Weight
		if r < cumulative {
			return &entries[i]
		}
	}

	// Fallback to last entry (shouldn't reach here)
	return &entries[len(entries)-1]
}
