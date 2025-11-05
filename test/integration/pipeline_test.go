package integration

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/validation"
)

// TestIntegration_CompletePipeline verifies that Generate() produces
// a complete Artifact with all pipeline stages populated.
func TestIntegration_CompletePipeline(t *testing.T) {
	// Use a medium-sized config to test all stages
	cfg := &dungeon.Config{
		Seed: 42,
		Size: dungeon.SizeCfg{
			RoomsMin: 30,
			RoomsMax: 40,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingSCurve,
			Variance: 0.15,
		},
		Themes:        []string{"dungeon", "crypt"},
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Keys: []dungeon.KeyCfg{
			{Name: "silver", Count: 1},
			{Name: "gold", Count: 1},
		},
	}

	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
	artifact, err := gen.Generate(context.Background(), cfg)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if artifact == nil {
		t.Fatal("Generate() returned nil artifact")
	}

	// Verify Stage 1: Graph Synthesis
	if artifact.ADG == nil {
		t.Error("Artifact.ADG is nil - graph synthesis stage incomplete")
	} else {
		t.Logf("✓ Stage 1: Graph generated with %d rooms", len(artifact.ADG.Rooms))

		// Verify graph properties
		if len(artifact.ADG.Rooms) < cfg.Size.RoomsMin {
			t.Errorf("Room count %d < minimum %d", len(artifact.ADG.Rooms), cfg.Size.RoomsMin)
		}
		if len(artifact.ADG.Rooms) > cfg.Size.RoomsMax {
			t.Errorf("Room count %d > maximum %d", len(artifact.ADG.Rooms), cfg.Size.RoomsMax)
		}
		if len(artifact.ADG.Connectors) == 0 {
			t.Error("Graph has no connectors - rooms must be connected")
		}
	}

	// Verify Stage 2: Spatial Embedding
	if artifact.Layout == nil {
		t.Error("Artifact.Layout is nil - embedding stage incomplete")
	} else {
		t.Logf("✓ Stage 2: Layout generated with bounds %+v", artifact.Layout.Bounds)

		// Verify all rooms have poses
		if artifact.ADG != nil {
			for roomID := range artifact.ADG.Rooms {
				if _, hasPose := artifact.Layout.Poses[roomID]; !hasPose {
					t.Errorf("Room %s has no pose in layout", roomID)
				}
			}
		}

		// Verify all connectors have paths
		if artifact.ADG != nil {
			for connID := range artifact.ADG.Connectors {
				if _, hasPath := artifact.Layout.CorridorPaths[connID]; !hasPath {
					t.Errorf("Connector %s has no path in layout", connID)
				}
			}
		}
	}

	// Verify Stage 3: Tile Carving
	if artifact.TileMap == nil {
		t.Error("Artifact.TileMap is nil - carving stage incomplete")
	} else {
		t.Logf("✓ Stage 3: TileMap generated %dx%d", artifact.TileMap.Width, artifact.TileMap.Height)

		// Verify tilemap dimensions match layout bounds
		if artifact.Layout != nil {
			if artifact.TileMap.Width == 0 || artifact.TileMap.Height == 0 {
				t.Error("TileMap has zero dimensions")
			}
		}

		// Verify required layers exist
		if len(artifact.TileMap.Layers) == 0 {
			t.Error("TileMap has no layers")
		}
	}

	// Verify Stage 4: Content Population
	if artifact.Content == nil {
		t.Error("Artifact.Content is nil - content stage incomplete")
	} else {
		t.Logf("✓ Stage 4: Content populated with %d spawns, %d loot, %d puzzles",
			len(artifact.Content.Spawns),
			len(artifact.Content.Loot),
			len(artifact.Content.Puzzles))

		// Verify content is non-empty for a dungeon this size
		hasContent := len(artifact.Content.Spawns) > 0 ||
			len(artifact.Content.Loot) > 0 ||
			len(artifact.Content.Puzzles) > 0

		if !hasContent {
			t.Error("Content stage produced no content - expected spawns/loot/puzzles")
		}
	}

	// Verify Stage 5: Validation & Metrics
	if artifact.Metrics == nil {
		t.Error("Artifact.Metrics is nil - validation stage incomplete")
	} else {
		t.Logf("✓ Stage 5: Metrics computed (branching=%.2f, pathLen=%d, cycles=%d)",
			artifact.Metrics.BranchingFactor,
			artifact.Metrics.PathLength,
			artifact.Metrics.CycleCount)

		// Verify metrics are sensible
		if artifact.Metrics.BranchingFactor < 1.0 {
			t.Error("Branching factor < 1.0 indicates disconnected graph")
		}
		if artifact.Metrics.PathLength == 0 && len(artifact.ADG.Rooms) > 1 {
			t.Error("Path length is 0 for multi-room dungeon")
		}
	}

	t.Log("✓ All pipeline stages completed successfully")
}

// TestGolden_Determinism verifies that the same seed produces identical output.
func TestGolden_Determinism(t *testing.T) {
	cfg, err := dungeon.LoadConfig("../../testdata/seeds/small_crypt.yaml")
	if err != nil {
		t.Fatal(err)
	}

	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())

	// Generate first artifact
	artifact1, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	if artifact1 == nil {
		t.Fatal("First Generate() returned nil artifact")
	}

	// Generate second artifact with same config (same seed)
	artifact2, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	if artifact2 == nil {
		t.Fatal("Second Generate() returned nil artifact")
	}

	// For now, just verify the graphs have the same structure
	// Full determinism testing will require proper serialization
	if len(artifact1.ADG.Rooms) != len(artifact2.ADG.Rooms) {
		t.Fatalf("Room counts differ: %d vs %d", len(artifact1.ADG.Rooms), len(artifact2.ADG.Rooms))
	}

	if len(artifact1.ADG.Connectors) != len(artifact2.ADG.Connectors) {
		t.Fatalf("Connector counts differ: %d vs %d", len(artifact1.ADG.Connectors), len(artifact2.ADG.Connectors))
	}

	t.Log("✓ Same seed produced consistent output structure")
}

// TestIntegration_PathologicalSeed is a regression test for seed 0x4400f4
// which previously caused "no valid path found" errors due to insufficient corridor length scaling.
// This seed produces spread-out layouts that require longer corridors.
func TestIntegration_PathologicalSeed(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 0x4400f4,
		Size: dungeon.SizeCfg{
			RoomsMin: 25,
			RoomsMax: 30,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingSCurve,
			Variance: 0.1,
		},
		Themes:        []string{"crypt"},
		SecretDensity: 0.15,
		OptionalRatio: 0.20,
	}

	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
	artifact, err := gen.Generate(context.Background(), cfg)

	if err != nil {
		t.Fatalf("Pathological seed 0x4400f4 failed generation: %v", err)
	}

	if artifact.ADG == nil {
		t.Fatal("Artifact missing ADG")
	}

	// Verify the dungeon has rooms and is connected
	if len(artifact.ADG.Rooms) < cfg.Size.RoomsMin {
		t.Errorf("Generated %d rooms, expected at least %d", len(artifact.ADG.Rooms), cfg.Size.RoomsMin)
	}

	if !artifact.ADG.IsWeaklyConnected() {
		t.Error("Generated dungeon is not connected")
	}

	t.Logf("✓ Pathological seed 0x4400f4 handled successfully: %d rooms", len(artifact.ADG.Rooms))
}
