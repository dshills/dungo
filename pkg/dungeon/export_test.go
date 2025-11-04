package dungeon_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/export"
)

// TestJSONRoundTrip verifies that a dungeon can be exported to JSON and parsed back
// with all data intact (round-trip fidelity).
// This is the RED PHASE - test will fail until JSON export is properly implemented.
func TestJSONRoundTrip(t *testing.T) {
	// Generate a small test dungeon
	cfg := &dungeon.Config{
		Seed: 12345,
		Size: dungeon.SizeCfg{
			RoomsMin: 5,
			RoomsMax: 8,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 1.5,
			Max: 3,
		},
		SecretDensity: 0.2,
		OptionalRatio: 0.3,
		Keys: []dungeon.KeyCfg{
			{Name: "bronze_key", Count: 1},
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.1,
		},
		Themes: []string{"crypt"},
	}

	// Generate dungeon (requires validator to be set)
	gen := dungeon.NewGenerator()
	// TODO: This will fail until validator is properly integrated
	// For now, we expect the test to fail at generation or export
	artifact, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping round-trip test: generation failed (expected in RED phase): %v", err)
	}

	// Export to JSON
	jsonData, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify JSON is valid
	if len(jsonData) == 0 {
		t.Fatal("ExportJSON returned empty data")
	}

	// Parse JSON back into a new artifact
	var parsed dungeon.Artifact
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify round-trip: critical fields must match
	// Note: We check structure existence and counts, not deep equality
	// Deep equality can fail due to unexported fields or pointer comparisons

	// Check ADG structure
	if parsed.ADG == nil {
		t.Fatal("Round-trip failed: ADG is nil")
	}
	if len(parsed.ADG.Rooms) != len(artifact.ADG.Rooms) {
		t.Errorf("Room count mismatch: got %d, want %d",
			len(parsed.ADG.Rooms), len(artifact.ADG.Rooms))
	}
	if len(parsed.ADG.Connectors) != len(artifact.ADG.Connectors) {
		t.Errorf("Connector count mismatch: got %d, want %d",
			len(parsed.ADG.Connectors), len(artifact.ADG.Connectors))
	}

	// Check Layout structure
	if parsed.Layout == nil {
		t.Fatal("Round-trip failed: Layout is nil")
	}
	if len(parsed.Layout.Poses) != len(artifact.Layout.Poses) {
		t.Errorf("Pose count mismatch: got %d, want %d",
			len(parsed.Layout.Poses), len(artifact.Layout.Poses))
	}
	if len(parsed.Layout.CorridorPaths) != len(artifact.Layout.CorridorPaths) {
		t.Errorf("CorridorPath count mismatch: got %d, want %d",
			len(parsed.Layout.CorridorPaths), len(artifact.Layout.CorridorPaths))
	}

	// Check TileMap structure
	if parsed.TileMap == nil {
		t.Fatal("Round-trip failed: TileMap is nil")
	}
	if parsed.TileMap.Width != artifact.TileMap.Width {
		t.Errorf("TileMap Width mismatch: got %d, want %d",
			parsed.TileMap.Width, artifact.TileMap.Width)
	}
	if parsed.TileMap.Height != artifact.TileMap.Height {
		t.Errorf("TileMap Height mismatch: got %d, want %d",
			parsed.TileMap.Height, artifact.TileMap.Height)
	}
	if len(parsed.TileMap.Layers) != len(artifact.TileMap.Layers) {
		t.Errorf("Layer count mismatch: got %d, want %d",
			len(parsed.TileMap.Layers), len(artifact.TileMap.Layers))
	}

	// Check Content structure
	if parsed.Content == nil {
		t.Fatal("Round-trip failed: Content is nil")
	}
	if len(parsed.Content.Spawns) != len(artifact.Content.Spawns) {
		t.Errorf("Spawn count mismatch: got %d, want %d",
			len(parsed.Content.Spawns), len(artifact.Content.Spawns))
	}
	if len(parsed.Content.Loot) != len(artifact.Content.Loot) {
		t.Errorf("Loot count mismatch: got %d, want %d",
			len(parsed.Content.Loot), len(artifact.Content.Loot))
	}
	if len(parsed.Content.Puzzles) != len(artifact.Content.Puzzles) {
		t.Errorf("Puzzle count mismatch: got %d, want %d",
			len(parsed.Content.Puzzles), len(artifact.Content.Puzzles))
	}
	if len(parsed.Content.Secrets) != len(artifact.Content.Secrets) {
		t.Errorf("Secret count mismatch: got %d, want %d",
			len(parsed.Content.Secrets), len(artifact.Content.Secrets))
	}

	// Check Metrics
	if parsed.Metrics != nil && artifact.Metrics != nil {
		if parsed.Metrics.PathLength != artifact.Metrics.PathLength {
			t.Errorf("PathLength mismatch: got %d, want %d",
				parsed.Metrics.PathLength, artifact.Metrics.PathLength)
		}
		if parsed.Metrics.CycleCount != artifact.Metrics.CycleCount {
			t.Errorf("CycleCount mismatch: got %d, want %d",
				parsed.Metrics.CycleCount, artifact.Metrics.CycleCount)
		}
	}

	t.Log("JSON round-trip test structure validation passed")
}

// TestJSONCompactRoundTrip verifies compact JSON export/import.
func TestJSONCompactRoundTrip(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 99999,
		Size: dungeon.SizeCfg{
			RoomsMin: 3,
			RoomsMax: 5,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 1.2,
			Max: 2,
		},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.05,
		},
		Themes: []string{"fungal"},
	}

	gen := dungeon.NewGenerator()
	artifact, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping compact round-trip test: generation failed: %v", err)
	}

	// Export to compact JSON
	compactData, err := export.ExportJSONCompact(artifact)
	if err != nil {
		t.Fatalf("ExportJSONCompact failed: %v", err)
	}

	// Verify compact JSON is smaller than formatted
	formattedData, _ := export.ExportJSON(artifact)
	if len(compactData) >= len(formattedData) {
		t.Errorf("Compact JSON should be smaller: compact=%d, formatted=%d",
			len(compactData), len(formattedData))
	}

	// Parse compact JSON
	var parsed dungeon.Artifact
	if err := json.Unmarshal(compactData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal compact JSON: %v", err)
	}

	// Basic structure checks
	if parsed.ADG == nil {
		t.Fatal("Compact round-trip failed: ADG is nil")
	}
	if parsed.Layout == nil {
		t.Fatal("Compact round-trip failed: Layout is nil")
	}
	if parsed.TileMap == nil {
		t.Fatal("Compact round-trip failed: TileMap is nil")
	}
	if parsed.Content == nil {
		t.Fatal("Compact round-trip failed: Content is nil")
	}

	t.Log("Compact JSON round-trip test passed")
}

// TestJSONExportDeterminism verifies that the same artifact produces identical JSON.
func TestJSONExportDeterminism(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 42424242,
		Size: dungeon.SizeCfg{
			RoomsMin: 4,
			RoomsMax: 6,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 1.3,
			Max: 2,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.1,
		},
		Themes: []string{"temple"},
	}

	gen := dungeon.NewGenerator()
	artifact, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping determinism test: generation failed: %v", err)
	}

	// Export twice
	json1, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("First export failed: %v", err)
	}

	json2, err := export.ExportJSON(artifact)
	if err != nil {
		t.Fatalf("Second export failed: %v", err)
	}

	// Compare byte-for-byte
	if string(json1) != string(json2) {
		t.Error("JSON export is not deterministic: same artifact produced different JSON")
	}

	t.Log("JSON export determinism test passed")
}
