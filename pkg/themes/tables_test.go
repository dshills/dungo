package themes_test

import (
	"math/rand"
	"testing"

	"github.com/dshills/dungo/pkg/themes"
)

// T115: Unit test for encounter table selection
// Tests that encounter tables are selected by difficulty with proper interpolation

func TestGetEncountersForDifficulty(t *testing.T) {
	// Create a theme pack with multiple difficulty brackets
	theme := &themes.ThemePack{
		Name:        "test-encounters",
		Description: "Testing encounter selection",
		Tilesets: []themes.Tileset{
			{Name: "test", Path: "test.png", TileWidth: 32, TileHeight: 32},
		},
		EncounterTables: []themes.EncounterTable{
			{
				Difficulty: 0.0,
				Entries: []themes.WeightedEntry{
					{Type: "rat", Weight: 20},
					{Type: "goblin", Weight: 15},
				},
			},
			{
				Difficulty: 0.3,
				Entries: []themes.WeightedEntry{
					{Type: "orc", Weight: 12},
					{Type: "skeleton", Weight: 10},
				},
			},
			{
				Difficulty: 0.7,
				Entries: []themes.WeightedEntry{
					{Type: "ogre", Weight: 8},
					{Type: "dark_mage", Weight: 6},
				},
			},
			{
				Difficulty: 1.0,
				Entries: []themes.WeightedEntry{
					{Type: "dragon", Weight: 3},
					{Type: "demon_lord", Weight: 2},
				},
			},
		},
	}

	tests := []struct {
		name            string
		difficulty      float64
		expectedTypes   []string // Expected enemy types at this difficulty
		unexpectedTypes []string // Enemy types that should not appear
	}{
		{
			name:            "very low difficulty",
			difficulty:      0.05,
			expectedTypes:   []string{"rat", "goblin"},
			unexpectedTypes: []string{"dragon", "demon_lord", "ogre"},
		},
		{
			name:            "exact low difficulty bracket",
			difficulty:      0.3,
			expectedTypes:   []string{"orc", "skeleton"},
			unexpectedTypes: []string{"dragon", "demon_lord"},
		},
		{
			name:            "mid difficulty between brackets",
			difficulty:      0.5,
			expectedTypes:   []string{"orc", "skeleton", "ogre", "dark_mage"},
			unexpectedTypes: []string{"rat", "goblin", "dragon"},
		},
		{
			name:            "high difficulty",
			difficulty:      0.7,
			expectedTypes:   []string{"ogre", "dark_mage"},
			unexpectedTypes: []string{"rat", "goblin"},
		},
		{
			name:            "very high difficulty",
			difficulty:      0.9,
			expectedTypes:   []string{"dragon", "demon_lord", "ogre"},
			unexpectedTypes: []string{"rat", "goblin"},
		},
		{
			name:            "maximum difficulty",
			difficulty:      1.0,
			expectedTypes:   []string{"dragon", "demon_lord"},
			unexpectedTypes: []string{"rat", "goblin", "orc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encounters := theme.GetEncountersForDifficulty(tt.difficulty)

			if len(encounters) == 0 {
				t.Fatal("expected at least one encounter table")
			}

			// Collect all available enemy types
			availableTypes := make(map[string]bool)
			for _, table := range encounters {
				for _, entry := range table.Entries {
					availableTypes[entry.Type] = true
				}
			}

			// Verify expected types are present
			for _, expectedType := range tt.expectedTypes {
				if !availableTypes[expectedType] {
					t.Errorf("expected enemy type %q at difficulty %.2f", expectedType, tt.difficulty)
				}
			}

			// Verify unexpected types are absent
			for _, unexpectedType := range tt.unexpectedTypes {
				if availableTypes[unexpectedType] {
					t.Errorf("unexpected enemy type %q at difficulty %.2f", unexpectedType, tt.difficulty)
				}
			}
		})
	}
}

func TestGetEncountersForDifficulty_Interpolation(t *testing.T) {
	// Test that difficulty values between brackets get appropriate enemies
	theme := &themes.ThemePack{
		Name: "interpolation-test",
		Tilesets: []themes.Tileset{
			{Name: "test", Path: "test.png"},
		},
		EncounterTables: []themes.EncounterTable{
			{
				Difficulty: 0.0,
				Entries: []themes.WeightedEntry{
					{Type: "weak", Weight: 10},
				},
			},
			{
				Difficulty: 0.5,
				Entries: []themes.WeightedEntry{
					{Type: "medium", Weight: 10},
				},
			},
			{
				Difficulty: 1.0,
				Entries: []themes.WeightedEntry{
					{Type: "strong", Weight: 10},
				},
			},
		},
	}

	// Test difficulty 0.25 (between 0.0 and 0.5)
	encounters := theme.GetEncountersForDifficulty(0.25)
	if len(encounters) == 0 {
		t.Fatal("expected encounters for difficulty 0.25")
	}

	// Should get weak or medium enemies, or both
	availableTypes := make(map[string]bool)
	for _, table := range encounters {
		for _, entry := range table.Entries {
			availableTypes[entry.Type] = true
		}
	}

	hasWeakOrMedium := availableTypes["weak"] || availableTypes["medium"]
	if !hasWeakOrMedium {
		t.Error("expected weak or medium enemies at difficulty 0.25")
	}

	// Should NOT get strong enemies
	if availableTypes["strong"] {
		t.Error("unexpected strong enemy at difficulty 0.25")
	}
}

func TestGetEncountersForDifficulty_EdgeCases(t *testing.T) {
	theme := &themes.ThemePack{
		Name: "edge-cases",
		Tilesets: []themes.Tileset{
			{Name: "test", Path: "test.png"},
		},
		EncounterTables: []themes.EncounterTable{
			{
				Difficulty: 0.5,
				Entries: []themes.WeightedEntry{
					{Type: "enemy", Weight: 10},
				},
			},
		},
	}

	tests := []struct {
		name       string
		difficulty float64
		wantEmpty  bool
	}{
		{name: "negative difficulty", difficulty: -0.5, wantEmpty: false},
		{name: "zero difficulty", difficulty: 0.0, wantEmpty: false},
		{name: "difficulty above 1.0", difficulty: 1.5, wantEmpty: false},
		{name: "normal difficulty", difficulty: 0.5, wantEmpty: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encounters := theme.GetEncountersForDifficulty(tt.difficulty)
			isEmpty := len(encounters) == 0
			if isEmpty != tt.wantEmpty {
				t.Errorf("difficulty %.2f: expected empty=%v, got empty=%v", tt.difficulty, tt.wantEmpty, isEmpty)
			}
		})
	}
}

// T116: Unit test for loot table weighted selection
// Tests weighted random selection from loot tables

func TestSelectLootFromTable(t *testing.T) {
	lootTable := &themes.LootTable{
		RoomType: "treasure",
		Entries: []themes.WeightedEntry{
			{Type: "common_item", Weight: 50},
			{Type: "rare_item", Weight: 10},
			{Type: "legendary_item", Weight: 1},
		},
	}

	// Test deterministic selection with fixed RNG
	rng := rand.New(rand.NewSource(12345))

	// Perform many selections to test distribution
	iterations := 1000
	counts := make(map[string]int)

	for i := 0; i < iterations; i++ {
		item := themes.SelectWeightedEntry(lootTable.Entries, rng)
		if item == nil {
			t.Fatal("SelectWeightedEntry returned nil")
		}
		counts[item.Type]++
	}

	// Verify all items appeared at least once
	if counts["common_item"] == 0 {
		t.Error("common_item never selected")
	}
	if counts["rare_item"] == 0 {
		t.Error("rare_item never selected")
	}
	if counts["legendary_item"] == 0 {
		t.Error("legendary_item never selected")
	}

	// Verify common items are most frequent (roughly)
	if counts["common_item"] <= counts["rare_item"] {
		t.Error("common_item should appear more often than rare_item")
	}
	if counts["rare_item"] <= counts["legendary_item"] {
		t.Error("rare_item should appear more often than legendary_item")
	}

	// Check approximate distribution (with tolerance for randomness)
	totalWeight := 50 + 10 + 1 // 61
	expectedCommonRatio := float64(50) / float64(totalWeight)
	actualCommonRatio := float64(counts["common_item"]) / float64(iterations)

	// Allow 10% deviation from expected ratio
	tolerance := 0.1
	if actualCommonRatio < expectedCommonRatio-tolerance || actualCommonRatio > expectedCommonRatio+tolerance {
		t.Errorf("common_item ratio %.3f outside expected range %.3f±%.3f",
			actualCommonRatio, expectedCommonRatio, tolerance)
	}
}

func TestSelectWeightedEntry_Deterministic(t *testing.T) {
	entries := []themes.WeightedEntry{
		{Type: "item_a", Weight: 10},
		{Type: "item_b", Weight: 20},
		{Type: "item_c", Weight: 30},
	}

	// Test that same seed produces same results
	seed := int64(42)

	rng1 := rand.New(rand.NewSource(seed))
	item1 := themes.SelectWeightedEntry(entries, rng1)

	rng2 := rand.New(rand.NewSource(seed))
	item2 := themes.SelectWeightedEntry(entries, rng2)

	if item1 == nil || item2 == nil {
		t.Fatal("SelectWeightedEntry returned nil")
	}

	if item1.Type != item2.Type {
		t.Errorf("same seed produced different results: %q vs %q", item1.Type, item2.Type)
	}

	// Test that different seeds can produce different results
	rng3 := rand.New(rand.NewSource(99))
	results := make(map[string]bool)
	for i := 0; i < 50; i++ {
		item := themes.SelectWeightedEntry(entries, rng3)
		if item != nil {
			results[item.Type] = true
		}
	}

	// Should get variety across multiple selections
	if len(results) < 2 {
		t.Error("expected variety in weighted selection results")
	}
}

func TestSelectWeightedEntry_AllItemsSelectable(t *testing.T) {
	// Test with very different weights to ensure all items can be selected
	entries := []themes.WeightedEntry{
		{Type: "common", Weight: 1000},
		{Type: "uncommon", Weight: 100},
		{Type: "rare", Weight: 10},
		{Type: "ultra_rare", Weight: 1},
	}

	rng := rand.New(rand.NewSource(999))
	selected := make(map[string]bool)

	// Run enough iterations to likely hit all items
	maxIterations := 10000
	for i := 0; i < maxIterations; i++ {
		item := themes.SelectWeightedEntry(entries, rng)
		if item != nil {
			selected[item.Type] = true
		}
		// Early exit if we've seen all items
		if len(selected) == len(entries) {
			t.Logf("All items selected after %d iterations", i+1)
			break
		}
	}

	// Verify all items were selected at least once
	for _, entry := range entries {
		if !selected[entry.Type] {
			t.Errorf("item %q with weight %d was never selected", entry.Type, entry.Weight)
		}
	}
}

func TestSelectWeightedEntry_EmptyEntries(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	result := themes.SelectWeightedEntry([]themes.WeightedEntry{}, rng)
	if result != nil {
		t.Error("expected nil for empty entries")
	}
}

func TestSelectWeightedEntry_SingleEntry(t *testing.T) {
	entries := []themes.WeightedEntry{
		{Type: "only_item", Weight: 10},
	}

	rng := rand.New(rand.NewSource(0))

	// Should always return the single item
	for i := 0; i < 100; i++ {
		item := themes.SelectWeightedEntry(entries, rng)
		if item == nil {
			t.Fatal("expected item, got nil")
		}
		if item.Type != "only_item" {
			t.Errorf("expected 'only_item', got %q", item.Type)
		}
	}
}

func TestGetLootTableForRoomType(t *testing.T) {
	theme := &themes.ThemePack{
		Name: "loot-test",
		Tilesets: []themes.Tileset{
			{Name: "test", Path: "test.png"},
		},
		LootTables: []themes.LootTable{
			{
				RoomType: "treasure",
				Entries: []themes.WeightedEntry{
					{Type: "gold", Weight: 20},
				},
			},
			{
				RoomType: "boss",
				Entries: []themes.WeightedEntry{
					{Type: "artifact", Weight: 5},
				},
			},
			{
				RoomType: "secret",
				Entries: []themes.WeightedEntry{
					{Type: "rare_gem", Weight: 10},
				},
			},
		},
	}

	tests := []struct {
		roomType    string
		expectTable bool
		expectType  string
	}{
		{"treasure", true, "gold"},
		{"boss", true, "artifact"},
		{"secret", true, "rare_gem"},
		{"nonexistent", false, ""},
		{"", false, ""},
	}

	for _, tt := range tests {
		t.Run("room_type_"+tt.roomType, func(t *testing.T) {
			table := theme.GetLootTableForRoomType(tt.roomType)

			if tt.expectTable && table == nil {
				t.Fatalf("expected loot table for room type %q", tt.roomType)
			}
			if !tt.expectTable && table != nil {
				t.Fatalf("expected nil for room type %q, got table", tt.roomType)
			}

			if tt.expectTable {
				// Verify the table contains expected loot
				hasExpectedType := false
				for _, entry := range table.Entries {
					if entry.Type == tt.expectType {
						hasExpectedType = true
						break
					}
				}
				if !hasExpectedType {
					t.Errorf("loot table missing expected type %q", tt.expectType)
				}
			}
		})
	}
}
