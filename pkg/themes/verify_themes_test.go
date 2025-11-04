package themes

import (
	"testing"
)

// TestLoadActualThemeFiles verifies that the actual theme YAML files in themes/ directory
// can be loaded and parsed correctly with the current Go struct definitions.
func TestLoadActualThemeFiles(t *testing.T) {
	themes := []struct {
		name string
		dir  string
	}{
		{"crypt", "../../themes/crypt"},
		{"fungal", "../../themes/fungal"},
		{"arcane", "../../themes/arcane"},
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			theme, err := LoadThemeFromDirectory(tt.dir)
			if err != nil {
				t.Fatalf("Failed to load %s theme: %v", tt.name, err)
			}

			// Verify basic structure
			if theme.Name != tt.name {
				t.Errorf("Expected name %q, got %q", tt.name, theme.Name)
			}

			if len(theme.Tilesets) == 0 {
				t.Error("Expected at least one tileset")
			}

			if len(theme.EncounterTables) == 0 {
				t.Error("Expected at least one encounter table")
			}

			if len(theme.LootTables) == 0 {
				t.Error("Expected at least one loot table")
			}

			if len(theme.Decorators) == 0 {
				t.Error("Expected at least one decorator")
			}

			// Verify tileset structure
			for i, tileset := range theme.Tilesets {
				if tileset.Name == "" {
					t.Errorf("Tileset %d has empty name", i)
				}
				if tileset.Path == "" {
					t.Errorf("Tileset %d has empty path", i)
				}
				if tileset.TileWidth <= 0 {
					t.Errorf("Tileset %d has invalid tile_width: %d", i, tileset.TileWidth)
				}
				if tileset.TileHeight <= 0 {
					t.Errorf("Tileset %d has invalid tile_height: %d", i, tileset.TileHeight)
				}
			}

			// Verify encounter tables
			for i, table := range theme.EncounterTables {
				if table.Difficulty < 0.0 || table.Difficulty > 1.0 {
					t.Errorf("Encounter table %d has invalid difficulty: %f", i, table.Difficulty)
				}
				if len(table.Entries) == 0 {
					t.Errorf("Encounter table %d has no entries", i)
				}
			}

			// Verify loot tables
			for i, table := range theme.LootTables {
				if table.RoomType == "" {
					t.Errorf("Loot table %d has empty room_type", i)
				}
				if len(table.Entries) == 0 {
					t.Errorf("Loot table %d has no entries", i)
				}
			}

			// Verify decorators
			for i, decorator := range theme.Decorators {
				if decorator.Type == "" {
					t.Errorf("Decorator %d has empty type", i)
				}
				if decorator.Density < 0.0 || decorator.Density > 1.0 {
					t.Errorf("Decorator %d has invalid density: %f", i, decorator.Density)
				}
			}

			t.Logf("Successfully loaded %s theme: %d tilesets, %d encounter tables, %d loot tables, %d decorators",
				theme.Name, len(theme.Tilesets), len(theme.EncounterTables), len(theme.LootTables), len(theme.Decorators))
		})
	}
}
