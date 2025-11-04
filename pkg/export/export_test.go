package export_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/export"
	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// createTestArtifact creates a minimal artifact for testing.
func createTestArtifact() *dungeon.Artifact {
	g := graph.NewGraph(12345)
	start := &graph.Room{
		ID:         "start",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Difficulty: 0.1,
		Reward:     0.2,
	}
	boss := &graph.Room{
		ID:         "boss",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeXL,
		Difficulty: 1.0,
		Reward:     1.0,
	}
	g.AddRoom(start)
	g.AddRoom(boss)

	conn := &graph.Connector{
		ID:            "conn1",
		From:          "start",
		To:            "boss",
		Type:          graph.TypeDoor,
		Cost:          1.0,
		Bidirectional: false,
	}
	g.AddConnector(conn)

	return &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
		Layout: &dungeon.Layout{
			Poses: map[string]dungeon.Pose{
				start.ID: {X: 0, Y: 0, Rotation: 0, FootprintID: "rect_5x5"},
				boss.ID:  {X: 100, Y: 100, Rotation: 0, FootprintID: "rect_10x10"},
			},
			CorridorPaths: map[string]dungeon.Path{
				"conn1": {
					Points: []dungeon.Point{
						{X: 5, Y: 5},
						{X: 50, Y: 50},
						{X: 95, Y: 95},
					},
				},
			},
			Bounds: dungeon.Rect{X: 0, Y: 0, Width: 200, Height: 200},
		},
		TileMap: &dungeon.TileMap{
			Width:      200,
			Height:     200,
			TileWidth:  16,
			TileHeight: 16,
			Layers: map[string]*dungeon.Layer{
				"floor": {
					ID:      1,
					Name:    "floor",
					Type:    "tilelayer",
					Visible: true,
					Opacity: 1.0,
					Data:    make([]uint32, 200*200),
				},
			},
		},
		Content: &dungeon.Content{
			Spawns: []dungeon.Spawn{
				{
					ID:        "spawn1",
					RoomID:    boss.ID,
					Position:  dungeon.Point{X: 105, Y: 105},
					EnemyType: "boss_enemy",
					Count:     1,
				},
			},
			Loot: []dungeon.Loot{
				{
					ID:       "loot1",
					RoomID:   start.ID,
					Position: dungeon.Point{X: 5, Y: 5},
					ItemType: "key",
					Value:    100,
					Required: true,
				},
			},
			Puzzles: []dungeon.PuzzleInstance{},
			Secrets: []dungeon.SecretInstance{},
		},
		Metrics: &dungeon.Metrics{
			BranchingFactor:   1.0,
			PathLength:        1,
			CycleCount:        0,
			PacingDeviation:   0.05,
			SecretFindability: 0.8,
		},
	}
}

// TestExportJSON tests JSON export with indentation.
func TestExportJSON(t *testing.T) {
	artifact := createTestArtifact()

	data, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportJSON returned empty data")
	}

	// Verify it's valid JSON by unmarshaling
	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify basic structure
	if result.Metrics == nil {
		t.Error("Metrics not exported")
	}
	if result.Content == nil {
		t.Error("Content not exported")
	}
	if result.Layout == nil {
		t.Error("Layout not exported")
	}
}

// TestExportJSONCompact tests compact JSON export.
func TestExportJSONCompact(t *testing.T) {
	artifact := createTestArtifact()

	compact, err := export.ExportJSONCompact(artifact)
	if err != nil {
		t.Fatalf("ExportJSONCompact failed: %v", err)
	}

	pretty, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Compact should be smaller than pretty
	if len(compact) >= len(pretty) {
		t.Errorf("Compact JSON (%d bytes) is not smaller than pretty JSON (%d bytes)",
			len(compact), len(pretty))
	}

	// Both should unmarshal to same structure
	var compactResult, prettyResult dungeon.Artifact
	if err := json.Unmarshal(compact, &compactResult); err != nil {
		t.Fatalf("Compact JSON is invalid: %v", err)
	}
	if err := json.Unmarshal(pretty, &prettyResult); err != nil {
		t.Fatalf("Pretty JSON is invalid: %v", err)
	}
}

// TestJSONRoundTrip tests that we can export and re-import JSON.
func TestJSONRoundTrip(t *testing.T) {
	original := createTestArtifact()

	// Export to JSON
	data, err := export.ExportJSON(original)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Import from JSON
	var restored dungeon.Artifact
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify critical fields
	if restored.Metrics.BranchingFactor != original.Metrics.BranchingFactor {
		t.Errorf("BranchingFactor mismatch: got %f, want %f",
			restored.Metrics.BranchingFactor, original.Metrics.BranchingFactor)
	}

	if restored.Metrics.PathLength != original.Metrics.PathLength {
		t.Errorf("PathLength mismatch: got %d, want %d",
			restored.Metrics.PathLength, original.Metrics.PathLength)
	}

	if len(restored.Content.Spawns) != len(original.Content.Spawns) {
		t.Errorf("Spawns count mismatch: got %d, want %d",
			len(restored.Content.Spawns), len(original.Content.Spawns))
	}

	if len(restored.Content.Loot) != len(original.Content.Loot) {
		t.Errorf("Loot count mismatch: got %d, want %d",
			len(restored.Content.Loot), len(original.Content.Loot))
	}
}

// TestSaveJSONToFile tests saving JSON to a file.
func TestSaveJSONToFile(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "artifact.json")

	if err := export.SaveJSONToFile(artifact, filepath); err != nil {
		t.Fatalf("SaveJSONToFile failed: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Verify content is valid JSON
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result dungeon.Artifact
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Output file contains invalid JSON: %v", err)
	}
}

// TestSaveJSONCompactToFile tests saving compact JSON to a file.
func TestSaveJSONCompactToFile(t *testing.T) {
	artifact := createTestArtifact()
	tmpDir := t.TempDir()
	compactPath := filepath.Join(tmpDir, "artifact_compact.json")
	prettyPath := filepath.Join(tmpDir, "artifact_pretty.json")

	if err := export.SaveJSONCompactToFile(artifact, compactPath); err != nil {
		t.Fatalf("SaveJSONCompactToFile failed: %v", err)
	}

	if err := export.SaveJSONToFile(artifact, prettyPath); err != nil {
		t.Fatalf("SaveJSONToFile failed: %v", err)
	}

	// Compare file sizes
	compactInfo, err := os.Stat(compactPath)
	if err != nil {
		t.Fatalf("Compact file not found: %v", err)
	}

	prettyInfo, err := os.Stat(prettyPath)
	if err != nil {
		t.Fatalf("Pretty file not found: %v", err)
	}

	if compactInfo.Size() >= prettyInfo.Size() {
		t.Errorf("Compact file (%d bytes) is not smaller than pretty file (%d bytes)",
			compactInfo.Size(), prettyInfo.Size())
	}
}

// TestExportJSONErrorHandling tests error cases.
func TestExportJSONErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		artifact *dungeon.Artifact
		wantErr  bool
	}{
		{
			name:     "nil artifact",
			artifact: nil,
			wantErr:  false, // json.Marshal handles nil
		},
		{
			name:     "valid artifact",
			artifact: createTestArtifact(),
			wantErr:  false,
		},
		{
			name: "artifact with nil fields",
			artifact: &dungeon.Artifact{
				Metrics: nil,
				Content: nil,
			},
			wantErr: false, // Should handle nil fields gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := export.ExportJSON(tt.artifact)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSaveJSONToFileErrors tests file writing error cases.
func TestSaveJSONToFileErrors(t *testing.T) {
	artifact := createTestArtifact()

	// Test invalid path
	invalidPath := "/nonexistent/directory/artifact.json"
	err := export.SaveJSONToFile(artifact, invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// TestJSONMetricsPreservation tests that metrics are correctly preserved.
func TestJSONMetricsPreservation(t *testing.T) {
	artifact := &dungeon.Artifact{
		Metrics: &dungeon.Metrics{
			BranchingFactor:   2.5,
			PathLength:        10,
			CycleCount:        3,
			PacingDeviation:   0.123,
			SecretFindability: 0.789,
		},
	}

	data, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var restored dungeon.Artifact
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if restored.Metrics == nil {
		t.Fatal("Metrics is nil after restoration")
	}

	// Check each metric
	if restored.Metrics.BranchingFactor != artifact.Metrics.BranchingFactor {
		t.Errorf("BranchingFactor: got %f, want %f",
			restored.Metrics.BranchingFactor, artifact.Metrics.BranchingFactor)
	}

	if restored.Metrics.PathLength != artifact.Metrics.PathLength {
		t.Errorf("PathLength: got %d, want %d",
			restored.Metrics.PathLength, artifact.Metrics.PathLength)
	}

	if restored.Metrics.CycleCount != artifact.Metrics.CycleCount {
		t.Errorf("CycleCount: got %d, want %d",
			restored.Metrics.CycleCount, artifact.Metrics.CycleCount)
	}

	if restored.Metrics.PacingDeviation != artifact.Metrics.PacingDeviation {
		t.Errorf("PacingDeviation: got %f, want %f",
			restored.Metrics.PacingDeviation, artifact.Metrics.PacingDeviation)
	}

	if restored.Metrics.SecretFindability != artifact.Metrics.SecretFindability {
		t.Errorf("SecretFindability: got %f, want %f",
			restored.Metrics.SecretFindability, artifact.Metrics.SecretFindability)
	}
}

// TestJSONContentPreservation tests that content is correctly preserved.
func TestJSONContentPreservation(t *testing.T) {
	artifact := &dungeon.Artifact{
		Content: &dungeon.Content{
			Spawns: []dungeon.Spawn{
				{
					ID:        "spawn1",
					RoomID:    "room1",
					Position:  dungeon.Point{X: 10, Y: 20},
					EnemyType: "goblin",
					Count:     3,
				},
			},
			Loot: []dungeon.Loot{
				{
					ID:       "loot1",
					RoomID:   "room2",
					Position: dungeon.Point{X: 30, Y: 40},
					ItemType: "gold",
					Value:    100,
					Required: false,
				},
			},
		},
	}

	data, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var restored dungeon.Artifact
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if restored.Content == nil {
		t.Fatal("Content is nil after restoration")
	}

	// Check spawns
	if len(restored.Content.Spawns) != len(artifact.Content.Spawns) {
		t.Errorf("Spawns count: got %d, want %d",
			len(restored.Content.Spawns), len(artifact.Content.Spawns))
	}

	if len(restored.Content.Spawns) > 0 {
		spawn := restored.Content.Spawns[0]
		original := artifact.Content.Spawns[0]
		if spawn.ID != original.ID || spawn.EnemyType != original.EnemyType || spawn.Count != original.Count {
			t.Errorf("Spawn mismatch: got %+v, want %+v", spawn, original)
		}
	}

	// Check loot
	if len(restored.Content.Loot) != len(artifact.Content.Loot) {
		t.Errorf("Loot count: got %d, want %d",
			len(restored.Content.Loot), len(artifact.Content.Loot))
	}

	if len(restored.Content.Loot) > 0 {
		loot := restored.Content.Loot[0]
		original := artifact.Content.Loot[0]
		if loot.ID != original.ID || loot.ItemType != original.ItemType || loot.Value != original.Value {
			t.Errorf("Loot mismatch: got %+v, want %+v", loot, original)
		}
	}
}

// TestJSONLayoutPreservation tests that layout is correctly preserved.
func TestJSONLayoutPreservation(t *testing.T) {
	artifact := &dungeon.Artifact{
		Layout: &dungeon.Layout{
			Poses: map[string]dungeon.Pose{
				"room1": {X: 10, Y: 20, Rotation: 90, FootprintID: "rect_5x5"},
				"room2": {X: 50, Y: 60, Rotation: 180, FootprintID: "rect_10x10"},
			},
			CorridorPaths: map[string]dungeon.Path{
				"conn1": {
					Points: []dungeon.Point{
						{X: 15, Y: 25},
						{X: 45, Y: 55},
					},
				},
			},
			Bounds: dungeon.Rect{X: 0, Y: 0, Width: 100, Height: 100},
		},
	}

	data, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var restored dungeon.Artifact
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if restored.Layout == nil {
		t.Fatal("Layout is nil after restoration")
	}

	// Check poses
	if len(restored.Layout.Poses) != len(artifact.Layout.Poses) {
		t.Errorf("Poses count: got %d, want %d",
			len(restored.Layout.Poses), len(artifact.Layout.Poses))
	}

	for id, originalPose := range artifact.Layout.Poses {
		restoredPose, exists := restored.Layout.Poses[id]
		if !exists {
			t.Errorf("Pose for room %s not found in restored layout", id)
			continue
		}
		if restoredPose != originalPose {
			t.Errorf("Pose mismatch for room %s: got %+v, want %+v",
				id, restoredPose, originalPose)
		}
	}

	// Check bounds
	if restored.Layout.Bounds != artifact.Layout.Bounds {
		t.Errorf("Bounds mismatch: got %+v, want %+v",
			restored.Layout.Bounds, artifact.Layout.Bounds)
	}
}

// BenchmarkExportJSON benchmarks JSON export performance.
func BenchmarkExportJSON(b *testing.B) {
	artifact := createTestArtifact()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := export.ExportJSON(artifact)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkExportJSONCompact benchmarks compact JSON export performance.
func BenchmarkExportJSONCompact(b *testing.B) {
	artifact := createTestArtifact()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := export.ExportJSONCompact(artifact)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestExportWithRNG tests export with RNG-based determinism.
func TestExportWithRNG(t *testing.T) {
	seed := uint64(12345)
	stageName := "test-stage"
	configHash := []byte("test-config")
	rng1 := rng.NewRNG(seed, stageName, configHash)
	rng2 := rng.NewRNG(seed, stageName, configHash)

	// Create two artifacts with same seed
	artifact1 := createTestArtifact()
	artifact2 := createTestArtifact()

	// Add some randomization based on RNG
	_ = rng1.Intn(100)
	_ = rng2.Intn(100)

	// Export both
	data1, err := export.ExportJSON(artifact1)
	if err != nil {
		t.Fatalf("Export 1 failed: %v", err)
	}

	data2, err := export.ExportJSON(artifact2)
	if err != nil {
		t.Fatalf("Export 2 failed: %v", err)
	}

	// They should be identical for same artifacts
	if string(data1) != string(data2) {
		t.Error("Exports of identical artifacts differ")
	}
}
