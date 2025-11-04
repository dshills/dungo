package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

func createTestArtifact() *dungeon.Artifact {
	// Create test spawns
	spawns := []dungeon.Spawn{
		{
			ID:        "spawn-001",
			RoomID:    "room-001",
			Position:  dungeon.Point{X: 10, Y: 15},
			EnemyType: "goblin",
			Count:     3,
		},
	}

	// Create test loot
	loot := []dungeon.Loot{
		{
			ID:       "loot-001",
			RoomID:   "room-001",
			Position: dungeon.Point{X: 5, Y: 8},
			ItemType: "gold_pouch",
			Value:    50,
			Required: false,
		},
	}

	// Create test puzzles
	puzzles := []dungeon.PuzzleInstance{
		{
			ID:           "puzzle-001",
			RoomID:       "room-002",
			Type:         "lever",
			Requirements: []dungeon.Requirement{{Type: "key", Value: "brass_key"}},
			Provides:     []dungeon.Capability{{Type: "door_unlock", Value: "door-002"}},
			Difficulty:   0.5,
		},
	}

	// Create test secrets
	secrets := []dungeon.SecretInstance{
		{
			ID:       "secret-001",
			RoomID:   "room-003",
			Type:     "hidden_passage",
			Position: dungeon.Point{X: 20, Y: 25},
			Clues:    []string{"A crack in the wall", "Strange air flow"},
		},
	}

	// Create graph
	adg := &dungeon.Graph{
		Graph: &graph.Graph{
			Rooms: map[string]*graph.Room{
				"room-001": {ID: "room-001", Archetype: graph.ArchetypeStart, Size: graph.SizeM},
				"room-002": {ID: "room-002", Archetype: graph.ArchetypeBoss, Size: graph.SizeXL},
			},
			Connectors: map[string]*graph.Connector{
				"conn-001": {ID: "conn-001", From: "room-001", To: "room-002", Bidirectional: true},
			},
		},
	}

	return &dungeon.Artifact{
		ADG: adg,
		Layout: &dungeon.Layout{
			Poses: map[string]dungeon.Pose{
				"room-001": {X: 0, Y: 0, Rotation: 0, FootprintID: "rect-small"},
				"room-002": {X: 10, Y: 10, Rotation: 0, FootprintID: "rect-medium"},
			},
			CorridorPaths: map[string]dungeon.Path{
				"conn-001": {Points: []dungeon.Point{{X: 5, Y: 5}, {X: 10, Y: 10}}},
			},
			Bounds: dungeon.Rect{X: 0, Y: 0, Width: 100, Height: 100},
		},
		TileMap: &dungeon.TileMap{
			Width:      50,
			Height:     50,
			TileWidth:  16,
			TileHeight: 16,
			Layers: map[string]*dungeon.Layer{
				"floor": {
					ID:      1,
					Name:    "floor",
					Type:    "tilelayer",
					Visible: true,
					Opacity: 1.0,
					Data:    []uint32{1, 1, 1, 2, 2},
				},
			},
		},
		Content: &dungeon.Content{
			Spawns:  spawns,
			Loot:    loot,
			Puzzles: puzzles,
			Secrets: secrets,
		},
		Metrics: &dungeon.Metrics{
			BranchingFactor:   2.5,
			PathLength:        12,
			CycleCount:        2,
			PacingDeviation:   0.15,
			SecretFindability: 0.7,
		},
	}
}

func TestExportJSON(t *testing.T) {
	artifact := createTestArtifact()

	data, err := ExportJSON(artifact)
	if err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportJSON() returned empty data")
	}

	// Verify it's valid JSON by unmarshaling
	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("ExportJSON() produced invalid JSON: %v", err)
	}

	// Verify key fields
	if result.ADG == nil {
		t.Error("ADG is nil after unmarshal")
	}
	if result.Layout == nil {
		t.Error("Layout is nil after unmarshal")
	}
	if result.Content == nil {
		t.Error("Content is nil after unmarshal")
	}
}

func TestExportJSONCompact(t *testing.T) {
	artifact := createTestArtifact()

	data, err := ExportJSONCompact(artifact)
	if err != nil {
		t.Fatalf("ExportJSONCompact() error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportJSONCompact() returned empty data")
	}

	// Verify it's valid JSON
	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("ExportJSONCompact() produced invalid JSON: %v", err)
	}

	// Compact should be smaller than formatted
	formatted, _ := ExportJSON(artifact)
	if len(data) >= len(formatted) {
		t.Errorf("Compact JSON is not smaller: compact=%d, formatted=%d", len(data), len(formatted))
	}
}

func TestSaveJSONToFile(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_artifact.json")

	err := SaveJSONToFile(artifact, filePath)
	if err != nil {
		t.Fatalf("SaveJSONToFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("SaveJSONToFile() did not create file")
	}

	// Read and verify content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Saved file contains invalid JSON: %v", err)
	}

	if result.ADG == nil {
		t.Error("Saved artifact ADG is nil")
	}
}

func TestSaveJSONCompactToFile(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_artifact_compact.json")

	err := SaveJSONCompactToFile(artifact, filePath)
	if err != nil {
		t.Fatalf("SaveJSONCompactToFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("SaveJSONCompactToFile() did not create file")
	}

	// Read and verify content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Saved file contains invalid JSON: %v", err)
	}

	// Compare sizes with formatted version
	formattedPath := filepath.Join(tmpDir, "test_formatted.json")
	SaveJSONToFile(artifact, formattedPath)
	formattedData, _ := os.ReadFile(formattedPath)

	if len(data) >= len(formattedData) {
		t.Errorf("Compact file is not smaller: compact=%d, formatted=%d", len(data), len(formattedData))
	}
}

func TestExportJSON_EmptyArtifact(t *testing.T) {
	artifact := &dungeon.Artifact{}

	data, err := ExportJSON(artifact)
	if err != nil {
		t.Fatalf("ExportJSON() with empty artifact error = %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportJSON() returned empty data for empty artifact")
	}

	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("ExportJSON() produced invalid JSON for empty artifact: %v", err)
	}
}

func TestExportJSON_ComplexMetadata(t *testing.T) {
	// Create artifact with nested graph structure
	adg := &dungeon.Graph{
		Graph: &graph.Graph{
			Rooms: map[string]*graph.Room{
				"room-001": {
					ID:        "room-001",
					Archetype: graph.ArchetypeStart,
					Size:      graph.SizeM,
					Tags:      map[string]string{"theme": "dark", "lighting": "dim"},
				},
			},
			Connectors: map[string]*graph.Connector{},
			Metadata: map[string]interface{}{
				"string":  "value",
				"number":  42,
				"boolean": true,
				"nested": map[string]interface{}{
					"key": "value",
				},
				"array": []interface{}{1, 2, 3},
			},
		},
	}

	artifact := &dungeon.Artifact{
		ADG: adg,
	}

	data, err := ExportJSON(artifact)
	if err != nil {
		t.Fatalf("ExportJSON() with complex metadata error = %v", err)
	}

	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("ExportJSON() produced invalid JSON: %v", err)
	}

	if result.ADG == nil || result.ADG.Metadata == nil {
		t.Fatal("Metadata was not preserved")
	}

	if result.ADG.Metadata["string"] != "value" {
		t.Errorf("String metadata mismatch: got %v, want 'value'", result.ADG.Metadata["string"])
	}
}

func TestSaveJSONToFile_InvalidPath(t *testing.T) {
	artifact := createTestArtifact()
	invalidPath := "/nonexistent/directory/that/does/not/exist/file.json"

	err := SaveJSONToFile(artifact, invalidPath)
	if err == nil {
		t.Fatal("SaveJSONToFile() should fail with invalid path")
	}
}

func TestExportJSON_RoundTrip(t *testing.T) {
	original := createTestArtifact()

	// Export to JSON
	data, err := ExportJSON(original)
	if err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}

	// Import back
	var restored dungeon.Artifact
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify top-level structure is preserved
	if restored.ADG == nil {
		t.Error("ADG is nil after round-trip")
	}
	if restored.Layout == nil {
		t.Error("Layout is nil after round-trip")
	}
	if restored.Content == nil {
		t.Error("Content is nil after round-trip")
	}
	if restored.Metrics == nil {
		t.Error("Metrics is nil after round-trip")
	}

	// Verify nested structures
	if restored.ADG != nil && original.ADG != nil {
		if len(restored.ADG.Rooms) != len(original.ADG.Rooms) {
			t.Errorf("Rooms count mismatch: got %v, want %v", len(restored.ADG.Rooms), len(original.ADG.Rooms))
		}
		if len(restored.ADG.Connectors) != len(original.ADG.Connectors) {
			t.Errorf("Connectors count mismatch: got %v, want %v", len(restored.ADG.Connectors), len(original.ADG.Connectors))
		}
	}

	if restored.Layout != nil && original.Layout != nil {
		if len(restored.Layout.Poses) != len(original.Layout.Poses) {
			t.Errorf("Poses count mismatch: got %v, want %v", len(restored.Layout.Poses), len(original.Layout.Poses))
		}
		if len(restored.Layout.CorridorPaths) != len(original.Layout.CorridorPaths) {
			t.Errorf("CorridorPaths count mismatch: got %v, want %v", len(restored.Layout.CorridorPaths), len(original.Layout.CorridorPaths))
		}
	}

	if restored.Content != nil && original.Content != nil {
		if len(restored.Content.Spawns) != len(original.Content.Spawns) {
			t.Errorf("Spawns count mismatch: got %v, want %v", len(restored.Content.Spawns), len(original.Content.Spawns))
		}
		if len(restored.Content.Loot) != len(original.Content.Loot) {
			t.Errorf("Loot count mismatch: got %v, want %v", len(restored.Content.Loot), len(original.Content.Loot))
		}
		if len(restored.Content.Puzzles) != len(original.Content.Puzzles) {
			t.Errorf("Puzzles count mismatch: got %v, want %v", len(restored.Content.Puzzles), len(original.Content.Puzzles))
		}
		if len(restored.Content.Secrets) != len(original.Content.Secrets) {
			t.Errorf("Secrets count mismatch: got %v, want %v", len(restored.Content.Secrets), len(original.Content.Secrets))
		}
	}

	if restored.Metrics != nil && original.Metrics != nil {
		if restored.Metrics.BranchingFactor != original.Metrics.BranchingFactor {
			t.Errorf("BranchingFactor mismatch: got %v, want %v", restored.Metrics.BranchingFactor, original.Metrics.BranchingFactor)
		}
		if restored.Metrics.PathLength != original.Metrics.PathLength {
			t.Errorf("PathLength mismatch: got %v, want %v", restored.Metrics.PathLength, original.Metrics.PathLength)
		}
	}
}
