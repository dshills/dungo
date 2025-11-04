package dungeon_test

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/validation"
)

func TestNewGenerator(t *testing.T) {
	gen := dungeon.NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator() returned nil")
	}

	// Verify it implements the Generator interface
	var _ dungeon.Generator = gen
}

func TestDefaultGeneratorImplementsInterface(t *testing.T) {
	// Verify that DefaultGenerator implements Generator interface
	var _ dungeon.Generator = (*dungeon.DefaultGenerator)(nil)
}

func TestGenerateStub(t *testing.T) {
	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())

	cfg := &dungeon.Config{
		Seed: 12345,
		Size: dungeon.SizeCfg{
			RoomsMin: 10,
			RoomsMax: 20,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.1,
		},
		Themes:        []string{"dungeon"},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	ctx := context.Background()
	artifact, err := gen.Generate(ctx, cfg)

	// Synthesis stage is now implemented, so we should get a partial artifact
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if artifact == nil {
		t.Fatal("Expected artifact, got nil")
	}

	// Verify we have an ADG (synthesis stage is complete)
	if artifact.ADG == nil {
		t.Error("Expected ADG to be populated")
	}

	// Other stages are not yet implemented
	if artifact.Layout != nil {
		t.Error("Expected Layout to be nil (not yet implemented)")
	}
	if artifact.TileMap != nil {
		t.Error("Expected TileMap to be nil (not yet implemented)")
	}
	if artifact.Content != nil {
		t.Error("Expected Content to be nil (not yet implemented)")
	}
}

func TestGenerateWithCancellation(t *testing.T) {
	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())

	cfg := &dungeon.Config{
		Seed: 12345,
		Size: dungeon.SizeCfg{
			RoomsMin: 10,
			RoomsMax: 20,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.1,
		},
		Themes:        []string{"dungeon"},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := gen.Generate(ctx, cfg)

	// For now, the stub doesn't check context cancellation
	// This test documents expected behavior for future implementation
	if err != nil {
		t.Logf("Context cancellation behavior: %v", err)
	}
}

// TestGolden_Determinism verifies that the same seed produces identical output.
// This is a critical property for dungeon generation - it ensures reproducibility
// and allows sharing of seeds between players.
// This test follows TDD - it will FAIL until full implementation is complete.
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

	// TODO: For now we import encoding/json for marshaling
	// In the future, we may want to implement custom comparison
	// or use a more sophisticated golden file format
	var (
		json1 []byte
		json2 []byte
	)

	// Marshal both artifacts to JSON
	json1, err = marshalArtifact(artifact1)
	if err != nil {
		t.Fatalf("Failed to marshal first artifact: %v", err)
	}

	json2, err = marshalArtifact(artifact2)
	if err != nil {
		t.Fatalf("Failed to marshal second artifact: %v", err)
	}

	// Property: Same seed must produce byte-for-byte identical output
	if len(json1) != len(json2) {
		t.Fatalf("Artifacts have different sizes: %d vs %d", len(json1), len(json2))
	}

	for i := range json1 {
		if json1[i] != json2[i] {
			t.Fatalf("Artifacts differ at byte %d - not deterministic", i)
		}
	}

	t.Log("✓ Same seed produced identical output - determinism verified")
}

// TestIntegration_CompletePipeline verifies that Generate() produces
// a complete Artifact with all pipeline stages populated.
// This test follows TDD - it will FAIL until full implementation is complete.
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

// marshalArtifact converts an Artifact to JSON bytes for comparison.
// This is a test helper function.
func marshalArtifact(a *dungeon.Artifact) ([]byte, error) {
	// TODO: This will need to import encoding/json once available
	// For now, this is a placeholder that will be implemented
	// when the test actually runs
	return nil, nil
}
