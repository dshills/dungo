package dungeon

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/graph"
)

func TestArtifactStructure(t *testing.T) {
	// Test that Artifact can be constructed with all fields
	artifact := &Artifact{
		ADG: &Graph{
			Graph: graph.NewGraph(12345),
		},
		Layout: &Layout{
			Poses:         make(map[string]Pose),
			CorridorPaths: make(map[string]Path),
			Bounds:        Rect{X: 0, Y: 0, Width: 100, Height: 100},
		},
		TileMap: &TileMap{
			Width:      100,
			Height:     100,
			TileWidth:  16,
			TileHeight: 16,
			Layers:     make(map[string]*Layer),
		},
		Content: &Content{
			Spawns:  []Spawn{},
			Loot:    []Loot{},
			Puzzles: []PuzzleInstance{},
			Secrets: []SecretInstance{},
		},
		Metrics: &Metrics{
			BranchingFactor:   2.5,
			PathLength:        10,
			CycleCount:        2,
			PacingDeviation:   0.1,
			SecretFindability: 0.8,
		},
		Debug: &DebugArtifacts{
			ADGSVG:    []byte("svg data"),
			LayoutPNG: []byte("png data"),
			Report: &ValidationReport{
				Passed: true,
			},
		},
	}

	// artifact is created with a literal, so it cannot be nil
	// This check is redundant and causes staticcheck SA4031

	if artifact.ADG.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", artifact.ADG.Seed)
	}

	if artifact.Metrics.BranchingFactor != 2.5 {
		t.Errorf("Expected branching factor 2.5, got %f", artifact.Metrics.BranchingFactor)
	}

	if !artifact.Debug.Report.Passed {
		t.Error("Expected validation report to be passed")
	}
}

func TestPoseStructure(t *testing.T) {
	pose := Pose{
		X:           10,
		Y:           20,
		Rotation:    90,
		FootprintID: "room_template_1",
	}

	if pose.X != 10 || pose.Y != 20 {
		t.Errorf("Expected position (10, 20), got (%d, %d)", pose.X, pose.Y)
	}

	if pose.Rotation != 90 {
		t.Errorf("Expected rotation 90, got %d", pose.Rotation)
	}
}

func TestContentStructure(t *testing.T) {
	content := &Content{
		Spawns: []Spawn{
			{
				ID:        "spawn_1",
				RoomID:    "room_1",
				Position:  Point{X: 5, Y: 5},
				EnemyType: "goblin",
				Count:     3,
			},
		},
		Loot: []Loot{
			{
				ID:       "loot_1",
				RoomID:   "room_1",
				Position: Point{X: 10, Y: 10},
				ItemType: "gold",
				Value:    100,
				Required: false,
			},
		},
	}

	if len(content.Spawns) != 1 {
		t.Errorf("Expected 1 spawn, got %d", len(content.Spawns))
	}

	if content.Spawns[0].EnemyType != "goblin" {
		t.Errorf("Expected enemy type 'goblin', got %s", content.Spawns[0].EnemyType)
	}

	if len(content.Loot) != 1 {
		t.Errorf("Expected 1 loot item, got %d", len(content.Loot))
	}

	if content.Loot[0].Value != 100 {
		t.Errorf("Expected loot value 100, got %d", content.Loot[0].Value)
	}
}

// createTestArtifact creates a minimal test artifact.
func createTestArtifact() *Artifact {
	return &Artifact{
		Metrics: &Metrics{
			BranchingFactor:   1.5,
			PathLength:        8,
			CycleCount:        1,
			PacingDeviation:   0.05,
			SecretFindability: 0.75,
		},
		Content: &Content{
			Spawns: []Spawn{
				{
					ID:        "spawn1",
					RoomID:    "room1",
					Position:  Point{X: 10, Y: 10},
					EnemyType: "goblin",
					Count:     2,
				},
			},
			Loot: []Loot{
				{
					ID:       "loot1",
					RoomID:   "room1",
					Position: Point{X: 5, Y: 5},
					ItemType: "key",
					Value:    50,
					Required: true,
				},
			},
		},
	}
}

// TestArtifactExportJSON tests JSON export.
func TestArtifactExportJSON(t *testing.T) {
	artifact := createTestArtifact()

	data, err := artifact.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportJSON returned empty data")
	}

	// Verify valid JSON
	var result Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}
}

// TestArtifactExportJSONCompact tests compact JSON export.
func TestArtifactExportJSONCompact(t *testing.T) {
	artifact := createTestArtifact()

	compact, err := artifact.ExportJSONCompact()
	if err != nil {
		t.Fatalf("ExportJSONCompact failed: %v", err)
	}

	pretty, err := artifact.ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Compact should be smaller
	if len(compact) >= len(pretty) {
		t.Errorf("Compact JSON (%d bytes) should be smaller than pretty JSON (%d bytes)",
			len(compact), len(pretty))
	}
}

// TestArtifactSaveJSON tests saving to file.
func TestArtifactSaveJSON(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "artifact.json")

	if err := artifact.SaveJSON(path); err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Verify valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var result Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("File contains invalid JSON: %v", err)
	}
}

// TestArtifactSaveJSONCompact tests saving compact JSON.
func TestArtifactSaveJSONCompact(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	compactPath := filepath.Join(tmpDir, "compact.json")
	prettyPath := filepath.Join(tmpDir, "pretty.json")

	if err := artifact.SaveJSONCompact(compactPath); err != nil {
		t.Fatalf("SaveJSONCompact failed: %v", err)
	}

	if err := artifact.SaveJSON(prettyPath); err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}

	// Compare sizes
	compactInfo, _ := os.Stat(compactPath)
	prettyInfo, _ := os.Stat(prettyPath)

	if compactInfo.Size() >= prettyInfo.Size() {
		t.Errorf("Compact file (%d bytes) should be smaller than pretty file (%d bytes)",
			compactInfo.Size(), prettyInfo.Size())
	}
}

// TestArtifactExportTMJ tests TMJ export (not yet implemented).
func TestArtifactExportTMJ(t *testing.T) {
	artifact := createTestArtifact()

	_, err := artifact.ExportTMJ()
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

// TestArtifactSaveTMJ tests TMJ save (not yet implemented).
func TestArtifactSaveTMJ(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "dungeon.tmj")

	err := artifact.SaveTMJ(path)
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

// TestArtifactExportSVG tests SVG export (not yet implemented).
func TestArtifactExportSVG(t *testing.T) {
	artifact := createTestArtifact()
	opts := DefaultSVGOptions()

	_, err := artifact.ExportSVG(opts)
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

// TestArtifactSaveSVG tests SVG save (not yet implemented).
func TestArtifactSaveSVG(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "dungeon.svg")
	opts := DefaultSVGOptions()

	err := artifact.SaveSVG(path, opts)
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

// TestDefaultSVGOptions tests default SVG options.
func TestDefaultSVGOptions(t *testing.T) {
	opts := DefaultSVGOptions()

	if opts.Width != 1024 {
		t.Errorf("Expected width 1024, got %d", opts.Width)
	}

	if opts.Height != 768 {
		t.Errorf("Expected height 768, got %d", opts.Height)
	}

	if !opts.ShowDifficulty {
		t.Error("Expected ShowDifficulty to be true")
	}

	if !opts.ShowLegend {
		t.Error("Expected ShowLegend to be true")
	}

	if opts.Scale != 1.0 {
		t.Errorf("Expected scale 1.0, got %f", opts.Scale)
	}
}

// TestArtifactJSONRoundTrip tests export and re-import.
func TestArtifactJSONRoundTrip(t *testing.T) {
	original := createTestArtifact()

	data, err := original.ExportJSON()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var restored Artifact
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify metrics
	if restored.Metrics.BranchingFactor != original.Metrics.BranchingFactor {
		t.Errorf("BranchingFactor mismatch: got %f, want %f",
			restored.Metrics.BranchingFactor, original.Metrics.BranchingFactor)
	}

	if restored.Metrics.PathLength != original.Metrics.PathLength {
		t.Errorf("PathLength mismatch: got %d, want %d",
			restored.Metrics.PathLength, original.Metrics.PathLength)
	}

	// Verify content
	if len(restored.Content.Spawns) != len(original.Content.Spawns) {
		t.Errorf("Spawns count mismatch: got %d, want %d",
			len(restored.Content.Spawns), len(original.Content.Spawns))
	}

	if len(restored.Content.Loot) != len(original.Content.Loot) {
		t.Errorf("Loot count mismatch: got %d, want %d",
			len(restored.Content.Loot), len(original.Content.Loot))
	}
}
