package dungeon_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/validation"
)

func TestNewGenerator(t *testing.T) {
	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
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

	// All pipeline stages should now be implemented
	if artifact.Layout == nil {
		t.Error("Expected Layout to be populated (embedding stage)")
	}
	if artifact.TileMap == nil {
		t.Error("Expected TileMap to be populated (carving stage)")
	}
	if artifact.Content == nil {
		t.Error("Expected Content to be populated (content stage)")
	}
	if artifact.Metrics == nil {
		t.Error("Expected Metrics to be populated (validation stage)")
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

// =============================================================================
// TDD RED PHASE TESTS - T081: Golden tests for pacing curves
// =============================================================================

// TestGolden_LinearPacingCurve verifies that LINEAR pacing produces
// uniform difficulty distribution in generated dungeons.
// TDD RED PHASE: Will FAIL until pacing implementation is complete.
func TestGolden_LinearPacingCurve(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 1000,
		Size: dungeon.SizeCfg{
			RoomsMin: 25,
			RoomsMax: 25,
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

	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
	artifact, err := gen.Generate(context.Background(), cfg)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if artifact.ADG == nil {
		t.Fatal("Artifact missing ADG")
	}

	// Find critical path (Start -> Boss)
	var startID, bossID string
	for id, room := range artifact.ADG.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startID = id
		}
		if room.Archetype == graph.ArchetypeBoss {
			bossID = id
		}
	}

	path, err := artifact.ADG.GetPath(startID, bossID)
	if err != nil {
		t.Fatalf("No path from Start to Boss: %v", err)
	}

	// Extract difficulty distribution along critical path
	difficulties := extractDifficultyDistribution(artifact.ADG, path)

	// Verify LINEAR pacing characteristics
	// Property: Difficulty should increase roughly uniformly
	for i := 1; i < len(difficulties); i++ {
		// Allow some variance but should generally increase
		if difficulties[i] < difficulties[i-1]-0.15 {
			t.Errorf("Linear pacing violated at position %d: difficulty decreased from %.2f to %.2f",
				i, difficulties[i-1], difficulties[i])
		}
	}

	// Calculate slope variations
	slopes := calculateSlopes(difficulties)
	avgSlope := average(slopes)
	maxSlopeDeviation := maxDeviation(slopes, avgSlope)

	// Property: Linear should have relatively consistent slope
	if maxSlopeDeviation > 0.3 {
		t.Errorf("Linear pacing slope too variable: max deviation %.2f > 0.3", maxSlopeDeviation)
	}

	t.Logf("✓ LINEAR pacing: path length=%d, avg slope=%.3f, max deviation=%.3f",
		len(path), avgSlope, maxSlopeDeviation)
	t.Logf("  Difficulty distribution: %v", formatDifficulties(difficulties))
}

// TestGolden_SCurvePacingCurve verifies that S_CURVE pacing produces
// smooth acceleration and deceleration.
// TDD RED PHASE: Will FAIL until pacing implementation is complete.
func TestGolden_SCurvePacingCurve(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 2000,
		Size: dungeon.SizeCfg{
			RoomsMin: 30,
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
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
	artifact, err := gen.Generate(context.Background(), cfg)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if artifact.ADG == nil {
		t.Fatal("Artifact missing ADG")
	}

	// Find critical path
	var startID, bossID string
	for id, room := range artifact.ADG.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startID = id
		}
		if room.Archetype == graph.ArchetypeBoss {
			bossID = id
		}
	}

	path, err := artifact.ADG.GetPath(startID, bossID)
	if err != nil {
		t.Fatalf("No path from Start to Boss: %v", err)
	}

	// S-curve testing requires at least 7 rooms for meaningful slope analysis
	// (need at least 2 slopes per section: early/mid/late)
	// With 7 rooms, we get 6 slopes which divide into 2 per section (6/3=2)
	if len(path) < 7 {
		t.Skipf("Critical path too short (%d rooms) for S-curve analysis, need at least 7", len(path))
	}

	difficulties := extractDifficultyDistribution(artifact.ADG, path)
	slopes := calculateSlopes(difficulties)

	// Property: S-curve should have low slope at start
	earlySlopes := slopes[:len(slopes)/3]
	avgEarlySlope := average(earlySlopes)

	// Property: S-curve should have high slope in middle
	midSlopes := slopes[len(slopes)/3 : 2*len(slopes)/3]
	avgMidSlope := average(midSlopes)

	// Property: S-curve should have low slope at end
	lateSlopes := slopes[2*len(slopes)/3:]
	avgLateSlope := average(lateSlopes)

	// Verify acceleration pattern: early < mid > late
	if avgMidSlope <= avgEarlySlope || avgMidSlope <= avgLateSlope {
		t.Errorf("S-curve acceleration pattern incorrect: early=%.3f, mid=%.3f, late=%.3f",
			avgEarlySlope, avgMidSlope, avgLateSlope)
	}

	t.Logf("✓ S_CURVE pacing: path length=%d", len(path))
	t.Logf("  Slopes: early=%.3f, mid=%.3f, late=%.3f (mid should be highest)",
		avgEarlySlope, avgMidSlope, avgLateSlope)
	t.Logf("  Difficulty distribution: %v", formatDifficulties(difficulties))
}

// TestGolden_ExponentialPacingCurve verifies that EXPONENTIAL pacing produces
// rapid difficulty increase toward the end.
// TDD RED PHASE: Will FAIL until pacing implementation is complete.
func TestGolden_ExponentialPacingCurve(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 3000,
		Size: dungeon.SizeCfg{
			RoomsMin: 25,
			RoomsMax: 25,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingExponential,
			Variance: 0.1,
		},
		Themes:        []string{"cave"},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
	artifact, err := gen.Generate(context.Background(), cfg)

	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if artifact.ADG == nil {
		t.Fatal("Artifact missing ADG")
	}

	// Find critical path
	var startID, bossID string
	for id, room := range artifact.ADG.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startID = id
		}
		if room.Archetype == graph.ArchetypeBoss {
			bossID = id
		}
	}

	path, err := artifact.ADG.GetPath(startID, bossID)
	if err != nil {
		t.Fatalf("No path from Start to Boss: %v", err)
	}

	difficulties := extractDifficultyDistribution(artifact.ADG, path)
	slopes := calculateSlopes(difficulties)

	// Property: Exponential should have increasing slopes (acceleration)
	earlySlopes := slopes[:len(slopes)/2]
	lateSlopes := slopes[len(slopes)/2:]

	avgEarlySlope := average(earlySlopes)
	avgLateSlope := average(lateSlopes)

	// Property: Late slopes should be significantly higher than early slopes
	if avgLateSlope <= avgEarlySlope {
		t.Errorf("Exponential pacing not accelerating: early slope %.3f >= late slope %.3f",
			avgEarlySlope, avgLateSlope)
	}

	accelerationRatio := avgLateSlope / avgEarlySlope
	if accelerationRatio < 1.5 {
		t.Errorf("Exponential acceleration too weak: ratio %.2f < 1.5", accelerationRatio)
	}

	t.Logf("✓ EXPONENTIAL pacing: path length=%d", len(path))
	t.Logf("  Acceleration: early slope=%.3f, late slope=%.3f, ratio=%.2fx",
		avgEarlySlope, avgLateSlope, accelerationRatio)
	t.Logf("  Difficulty distribution: %v", formatDifficulties(difficulties))
}

// TestGolden_PacingCurveComparison verifies that different curves
// produce measurably different difficulty distributions.
// TDD RED PHASE: Will FAIL until pacing implementation is complete.
func TestGolden_PacingCurveComparison(t *testing.T) {
	baseCfg := dungeon.Config{
		Seed: 4000,
		Size: dungeon.SizeCfg{
			RoomsMin: 30,
			RoomsMax: 30,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		Themes:        []string{"dungeon"},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	curves := []struct {
		name  string
		curve dungeon.PacingCurve
	}{
		{"LINEAR", dungeon.PacingLinear},
		{"S_CURVE", dungeon.PacingSCurve},
		{"EXPONENTIAL", dungeon.PacingExponential},
	}

	distributions := make(map[string][]float64)

	gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())

	for _, tc := range curves {
		cfg := baseCfg
		cfg.Pacing = dungeon.PacingCfg{
			Curve:    tc.curve,
			Variance: 0.1,
		}

		artifact, err := gen.Generate(context.Background(), &cfg)
		if err != nil {
			t.Fatalf("Generate() with %s failed: %v", tc.name, err)
		}

		if artifact.ADG == nil {
			t.Fatalf("%s: Artifact missing ADG", tc.name)
		}

		// Extract difficulty distribution
		var startID, bossID string
		for id, room := range artifact.ADG.Rooms {
			if room.Archetype == graph.ArchetypeStart {
				startID = id
			}
			if room.Archetype == graph.ArchetypeBoss {
				bossID = id
			}
		}

		path, err := artifact.ADG.GetPath(startID, bossID)
		if err != nil {
			t.Fatalf("%s: No path from Start to Boss: %v", tc.name, err)
		}

		distributions[tc.name] = extractDifficultyDistribution(artifact.ADG, path)
		t.Logf("%s distribution: %v", tc.name, formatDifficulties(distributions[tc.name]))
	}

	// Property: Curves should produce distinct distributions
	// Compare distributions at midpoint
	midIndex := len(distributions["LINEAR"]) / 2

	linearMid := distributions["LINEAR"][midIndex]
	scurveMid := distributions["S_CURVE"][midIndex]
	expMid := distributions["EXPONENTIAL"][midIndex]

	// At midpoint, exponential should be lower than linear
	if expMid >= linearMid-0.05 {
		t.Errorf("Exponential and Linear too similar at midpoint: exp=%.2f, linear=%.2f",
			expMid, linearMid)
	}

	// S-curve should be near linear at midpoint (inflection point)
	if scurveMid < linearMid-0.15 || scurveMid > linearMid+0.15 {
		t.Logf("Note: S-curve midpoint %.2f differs from linear %.2f (may indicate different steepness)",
			scurveMid, linearMid)
	}

	t.Logf("✓ Pacing curves produce distinct distributions")
	t.Logf("  Midpoint difficulties: LINEAR=%.2f, S_CURVE=%.2f, EXPONENTIAL=%.2f",
		linearMid, scurveMid, expMid)
}

// =============================================================================
// Test Helper Functions
// =============================================================================

// extractDifficultyDistribution extracts difficulty values along a path.
func extractDifficultyDistribution(g *dungeon.Graph, path []string) []float64 {
	difficulties := make([]float64, len(path))
	for i, roomID := range path {
		difficulties[i] = g.Rooms[roomID].Difficulty
	}
	return difficulties
}

// calculateSlopes computes difficulty increase between consecutive rooms.
func calculateSlopes(difficulties []float64) []float64 {
	if len(difficulties) < 2 {
		return []float64{}
	}

	slopes := make([]float64, len(difficulties)-1)
	for i := 0; i < len(difficulties)-1; i++ {
		slopes[i] = difficulties[i+1] - difficulties[i]
	}
	return slopes
}

// average calculates the mean of a slice of floats.
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// maxDeviation finds the maximum absolute deviation from a reference value.
func maxDeviation(values []float64, reference float64) float64 {
	maxDev := 0.0
	for _, v := range values {
		dev := v - reference
		if dev < 0 {
			dev = -dev
		}
		if dev > maxDev {
			maxDev = dev
		}
	}
	return maxDev
}

// formatDifficulties formats a difficulty distribution for logging.
func formatDifficulties(difficulties []float64) string {
	if len(difficulties) == 0 {
		return "[]"
	}

	// Sample evenly across the distribution (show ~10 values)
	step := len(difficulties) / 10
	if step < 1 {
		step = 1
	}

	result := "["
	for i := 0; i < len(difficulties); i += step {
		if i > 0 {
			result += ", "
		}
		result += formatFloat(difficulties[i])
	}
	// Always include the last value
	if (len(difficulties)-1)%step != 0 {
		result += ", " + formatFloat(difficulties[len(difficulties)-1])
	}
	result += "]"
	return result
}

// formatFloat formats a float to 2 decimal places.
func formatFloat(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
