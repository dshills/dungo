package themes_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/themes"
)

// T114: Unit test for theme YAML parsing
// Tests that theme packs can be loaded from YAML files with proper structure and validation

func TestLoadThemeFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, theme *themes.ThemePack)
	}{
		{
			name: "valid complete theme pack",
			yamlData: `
name: test-theme
description: A test theme for unit testing
tilesets:
  - name: dungeon_tiles
    path: assets/tilesets/dungeon.png
    tile_width: 32
    tile_height: 32
encounter_tables:
  - difficulty: 0.0
    entries:
      - type: goblin
        weight: 10
      - type: rat
        weight: 15
  - difficulty: 0.5
    entries:
      - type: orc
        weight: 8
      - type: goblin_shaman
        weight: 5
  - difficulty: 1.0
    entries:
      - type: troll
        weight: 5
      - type: dark_knight
        weight: 3
loot_tables:
  - room_type: treasure
    entries:
      - type: gold_coins
        weight: 20
      - type: magic_sword
        weight: 5
      - type: health_potion
        weight: 15
  - room_type: boss
    entries:
      - type: legendary_weapon
        weight: 1
      - type: artifact
        weight: 2
decorators:
  - type: torch
    density: 0.3
  - type: skull
    density: 0.1
`,
			wantErr: false,
			validate: func(t *testing.T, theme *themes.ThemePack) {
				if theme.Name != "test-theme" {
					t.Errorf("expected name 'test-theme', got %q", theme.Name)
				}
				if len(theme.Tilesets) != 1 {
					t.Errorf("expected 1 tileset, got %d", len(theme.Tilesets))
				}
				if len(theme.EncounterTables) != 3 {
					t.Errorf("expected 3 encounter tables, got %d", len(theme.EncounterTables))
				}
				if len(theme.LootTables) != 2 {
					t.Errorf("expected 2 loot tables, got %d", len(theme.LootTables))
				}
				if len(theme.Decorators) != 2 {
					t.Errorf("expected 2 decorators, got %d", len(theme.Decorators))
				}
			},
		},
		{
			name: "missing required name field",
			yamlData: `
description: Theme without name
tilesets:
  - name: test
    path: test.png
`,
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing tilesets",
			yamlData: `
name: no-tilesets
description: Theme without tilesets
`,
			wantErr: true,
			errMsg:  "at least one tileset is required",
		},
		{
			name: "encounter table with missing type",
			yamlData: `
name: bad-encounters
tilesets:
  - name: test
    path: test.png
encounter_tables:
  - difficulty: 0.5
    entries:
      - weight: 10
`,
			wantErr: true,
			errMsg:  "encounter entry type is required",
		},
		{
			name: "encounter table with invalid weight",
			yamlData: `
name: bad-weight
tilesets:
  - name: test
    path: test.png
encounter_tables:
  - difficulty: 0.5
    entries:
      - type: goblin
        weight: -5
`,
			wantErr: true,
			errMsg:  "weight must be positive",
		},
		{
			name: "invalid difficulty value out of range",
			yamlData: `
name: bad-difficulty
tilesets:
  - name: test
    path: test.png
encounter_tables:
  - difficulty: 1.5
    entries:
      - type: goblin
        weight: 10
`,
			wantErr: true,
			errMsg:  "difficulty must be between 0.0 and 1.0",
		},
		{
			name: "loot table with weighted entries",
			yamlData: `
name: loot-test
tilesets:
  - name: test
    path: test.png
loot_tables:
  - room_type: treasure
    entries:
      - type: common_item
        weight: 50
      - type: rare_item
        weight: 10
      - type: legendary_item
        weight: 1
`,
			wantErr: false,
			validate: func(t *testing.T, theme *themes.ThemePack) {
				if len(theme.LootTables) != 1 {
					t.Fatalf("expected 1 loot table, got %d", len(theme.LootTables))
				}
				loot := theme.LootTables[0]
				if loot.RoomType != "treasure" {
					t.Errorf("expected room_type 'treasure', got %q", loot.RoomType)
				}
				if len(loot.Entries) != 3 {
					t.Errorf("expected 3 loot entries, got %d", len(loot.Entries))
				}
				// Verify weights are preserved
				weights := []int{50, 10, 1}
				for i, entry := range loot.Entries {
					if entry.Weight != weights[i] {
						t.Errorf("entry %d: expected weight %d, got %d", i, weights[i], entry.Weight)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test YAML to temporary file
			tmpDir := t.TempDir()
			themePath := filepath.Join(tmpDir, "theme.yml")
			err := os.WriteFile(themePath, []byte(tt.yamlData), 0644)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Attempt to load theme
			theme, err := themes.LoadThemeFromFile(themePath)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			// Check success case
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Run custom validation if provided
			if tt.validate != nil {
				tt.validate(t, theme)
			}
		})
	}
}

// T117: Integration test for custom theme loading
// Tests complete workflow from filesystem to dungeon generation

func TestLoadThemeFromDirectory(t *testing.T) {
	// Create a temporary theme directory structure
	tmpDir := t.TempDir()
	themeDir := filepath.Join(tmpDir, "custom-theme")
	err := os.Mkdir(themeDir, 0755)
	if err != nil {
		t.Fatalf("failed to create theme directory: %v", err)
	}

	// Write a complete theme pack
	themeYAML := `
name: custom-integration-theme
description: Theme for integration testing
tilesets:
  - name: custom_tiles
    path: assets/custom.png
    tile_width: 16
    tile_height: 16
encounter_tables:
  - difficulty: 0.0
    entries:
      - type: custom_minion
        weight: 20
  - difficulty: 0.3
    entries:
      - type: custom_warrior
        weight: 15
  - difficulty: 0.7
    entries:
      - type: custom_elite
        weight: 10
  - difficulty: 1.0
    entries:
      - type: custom_boss
        weight: 5
loot_tables:
  - room_type: treasure
    entries:
      - type: custom_treasure
        weight: 10
decorators:
  - type: custom_prop
    density: 0.2
`
	themePath := filepath.Join(themeDir, "theme.yml")
	err = os.WriteFile(themePath, []byte(themeYAML), 0644)
	if err != nil {
		t.Fatalf("failed to write theme file: %v", err)
	}

	// Load the theme
	theme, err := themes.LoadThemeFromDirectory(themeDir)
	if err != nil {
		t.Fatalf("failed to load theme from directory: %v", err)
	}

	// Verify theme loaded correctly
	if theme.Name != "custom-integration-theme" {
		t.Errorf("expected theme name 'custom-integration-theme', got %q", theme.Name)
	}

	// Test that encounters can be selected by difficulty
	t.Run("custom encounters at low difficulty", func(t *testing.T) {
		encounters := theme.GetEncountersForDifficulty(0.1)
		if len(encounters) == 0 {
			t.Fatal("expected encounters for difficulty 0.1")
		}
		// Should get custom_minion at low difficulty
		hasCustomMinion := false
		for _, table := range encounters {
			for _, entry := range table.Entries {
				if entry.Type == "custom_minion" {
					hasCustomMinion = true
					break
				}
			}
		}
		if !hasCustomMinion {
			t.Error("expected custom_minion encounter at low difficulty")
		}
	})

	t.Run("custom encounters at high difficulty", func(t *testing.T) {
		encounters := theme.GetEncountersForDifficulty(0.9)
		if len(encounters) == 0 {
			t.Fatal("expected encounters for difficulty 0.9")
		}
		// Should get custom_boss or custom_elite at high difficulty
		hasHighLevelEnemy := false
		for _, table := range encounters {
			for _, entry := range table.Entries {
				if entry.Type == "custom_boss" || entry.Type == "custom_elite" {
					hasHighLevelEnemy = true
					break
				}
			}
		}
		if !hasHighLevelEnemy {
			t.Error("expected high-level custom enemy at high difficulty")
		}
	})

	t.Run("custom loot tables accessible", func(t *testing.T) {
		loot := theme.GetLootTableForRoomType("treasure")
		if loot == nil {
			t.Fatal("expected loot table for treasure room")
		}
		if len(loot.Entries) == 0 {
			t.Error("expected loot entries in treasure table")
		}
		// Verify custom_treasure exists
		hasCustomTreasure := false
		for _, entry := range loot.Entries {
			if entry.Type == "custom_treasure" {
				hasCustomTreasure = true
				break
			}
		}
		if !hasCustomTreasure {
			t.Error("expected custom_treasure in loot table")
		}
	})
}

func TestLoadThemeFromDirectory_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")

	_, err := themes.LoadThemeFromDirectory(nonExistentDir)
	if err == nil {
		t.Fatal("expected error loading from nonexistent directory")
	}
}

func TestValidateThemePack(t *testing.T) {
	tests := []struct {
		name    string
		theme   *themes.ThemePack
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid theme pack",
			theme: &themes.ThemePack{
				Name:        "valid",
				Description: "A valid theme",
				Tilesets: []themes.Tileset{
					{Name: "test", Path: "test.png", TileWidth: 32, TileHeight: 32},
				},
				EncounterTables: []themes.EncounterTable{
					{
						Difficulty: 0.5,
						Entries: []themes.WeightedEntry{
							{Type: "enemy", Weight: 10},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			theme: &themes.ThemePack{
				Description: "Missing name",
				Tilesets: []themes.Tileset{
					{Name: "test", Path: "test.png"},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty tilesets",
			theme: &themes.ThemePack{
				Name:     "no-tiles",
				Tilesets: []themes.Tileset{},
			},
			wantErr: true,
			errMsg:  "at least one tileset is required",
		},
		{
			name: "nil tilesets",
			theme: &themes.ThemePack{
				Name:     "nil-tiles",
				Tilesets: nil,
			},
			wantErr: true,
			errMsg:  "at least one tileset is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := themes.ValidateThemePack(tt.theme)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected validation error but got none")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}
