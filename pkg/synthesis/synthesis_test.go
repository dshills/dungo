package synthesis

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
	"pgregory.net/rapid"
)

// TestProperty_PacingCurveAdherence verifies that generated dungeons follow
// the configured difficulty curve within the specified variance tolerance.
// This property-based test will FAIL until pacing implementation is complete.
//
// For each pacing curve type (LINEAR, S_CURVE, EXPONENTIAL), we:
// 1. Generate a dungeon
// 2. Measure actual difficulty distribution along Start→Boss critical path
// 3. Verify deviation from expected curve is within variance tolerance
//
// TDD RED PHASE: This test documents the required pacing behavior.
func TestProperty_PacingCurveAdherence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	// Test each pacing curve type
	curveTypes := []string{
		"LINEAR",
		"S_CURVE",
		"EXPONENTIAL",
	}

	for _, curveType := range curveTypes {
		curveType := curveType // Capture for closure
		t.Run(curveType, func(t *testing.T) {
			rapid.Check(t, func(rt *rapid.T) {
				// Generate random but valid configuration
				seed := rapid.Uint64().Draw(rt, "seed")
				roomCount := rapid.IntRange(15, 50).Draw(rt, "roomCount")
				variance := rapid.Float64Range(0.05, 0.25).Draw(rt, "variance")

				cfg := &Config{
					Seed:          seed,
					RoomsMin:      roomCount,
					RoomsMax:      roomCount,
					BranchingAvg:  2.0,
					BranchingMax:  4,
					SecretDensity: 0.1,
					OptionalRatio: 0.2,
					Pacing: PacingConfig{
						Curve:    curveType,
						Variance: variance,
					},
					Themes: []string{"dungeon"},
				}

				// Create RNG and synthesizer
				rngInst := rng.NewRNG(seed, "test", []byte("test"))
				synthesizer := Get("grammar") // Assuming grammar synthesizer exists
				if synthesizer == nil {
					rt.Skip("grammar synthesizer not registered")
					return
				}

				// Generate the graph
				ctx := context.Background()
				g, err := synthesizer.Synthesize(ctx, rngInst, cfg)
				if err != nil {
					rt.Fatalf("synthesis failed: %v", err)
				}

				// Find Start and Boss rooms
				var startID, bossID string
				for id, room := range g.Rooms {
					if room.Archetype == graph.ArchetypeStart {
						startID = id
					}
					if room.Archetype == graph.ArchetypeBoss {
						bossID = id
					}
				}

				if startID == "" || bossID == "" {
					rt.Fatalf("missing Start or Boss room")
				}

				// Get critical path from Start to Boss
				path, err := g.GetPath(startID, bossID)
				if err != nil {
					rt.Fatalf("no path from Start to Boss: %v", err)
				}

				if len(path) < 2 {
					rt.Fatalf("path too short: %d rooms", len(path))
				}

				// Measure actual difficulty distribution along path
				actualDifficulties := make([]float64, len(path))
				for i, roomID := range path {
					actualDifficulties[i] = g.Rooms[roomID].Difficulty
				}

				// Calculate expected difficulties based on curve type
				expectedDifficulties := make([]float64, len(path))
				for i := range path {
					progress := float64(i) / float64(len(path)-1)
					expectedDifficulties[i] = calculateExpectedDifficulty(progress, curveType)
				}

				// Verify deviation is within variance tolerance
				// Use Mean Absolute Deviation (MAD) as our metric
				mad := calculateMAD(actualDifficulties, expectedDifficulties)

				// Property: MAD should be within variance tolerance
				// TDD: This will FAIL until pacing logic is implemented
				if mad > variance {
					rt.Fatalf("pacing curve deviation too high: MAD=%.3f > variance=%.3f\nPath: %v\nActual: %v\nExpected: %v",
						mad, variance, path, actualDifficulties, expectedDifficulties)
				}

				rt.Logf("✓ %s curve adhered to (MAD=%.3f, variance=%.3f, pathLen=%d)",
					curveType, mad, variance, len(path))
			})
		})
	}
}

// TestProperty_CustomPacingCurve verifies that CUSTOM curves interpolate
// correctly between user-defined control points.
// TDD RED PHASE: This will FAIL until custom curve interpolation is implemented.
func TestProperty_CustomPacingCurve(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	rapid.Check(t, func(rt *rapid.T) {
		// Generate random custom control points
		numPoints := rapid.IntRange(2, 5).Draw(rt, "numPoints")
		customPoints := make([][2]float64, numPoints)

		// First point must be at progress=0.0
		startDiff := rapid.Float64Range(0.0, 0.3).Draw(rt, "startDiff")
		customPoints[0] = [2]float64{0.0, startDiff}

		// Last point must be at progress=1.0
		endDiff := rapid.Float64Range(0.7, 1.0).Draw(rt, "endDiff")
		customPoints[numPoints-1] = [2]float64{1.0, endDiff}

		// Intermediate points at increasing progress values
		// Divide the [0,1] progress range into segments
		for i := 1; i < numPoints-1; i++ {
			// Calculate bounds for this intermediate point
			prevProgress := customPoints[i-1][0]
			// For intermediate point i, ensure there's room for remaining points
			remainingPoints := (numPoints - 1) - i
			// Each remaining point needs at least 0.05 space, plus keep away from 1.0
			maxProgress := 1.0 - float64(remainingPoints+1)*0.05

			// Ensure valid range by constraining progress properly
			minProg := prevProgress + 0.05
			if minProg >= maxProgress {
				// Not enough space, skip this test case
				rt.Skip("cannot generate valid progress points")
				return
			}

			progress := rapid.Float64Range(minProg, maxProgress).Draw(rt, fmt.Sprintf("prog_%d", i))

			// Difficulty can be anywhere in [0,1] - interpolation will handle it
			difficulty := rapid.Float64Range(0.0, 1.0).Draw(rt, fmt.Sprintf("diff_%d", i))
			customPoints[i] = [2]float64{progress, difficulty}
		}

		seed := rapid.Uint64().Draw(rt, "seed")
		roomCount := rapid.IntRange(20, 50).Draw(rt, "roomCount")

		cfg := &Config{
			Seed:          seed,
			RoomsMin:      roomCount,
			RoomsMax:      roomCount,
			BranchingAvg:  2.0,
			BranchingMax:  4,
			SecretDensity: 0.1,
			OptionalRatio: 0.2,
			Pacing: PacingConfig{
				Curve:        "CUSTOM",
				Variance:     0.1,
				CustomPoints: customPoints,
			},
			Themes: []string{"dungeon"},
		}

		// Create RNG and synthesizer
		rngInst := rng.NewRNG(seed, "test", []byte("test"))
		synthesizer := Get("grammar")
		if synthesizer == nil {
			rt.Skip("grammar synthesizer not registered")
			return
		}

		// Generate graph
		ctx := context.Background()
		g, err := synthesizer.Synthesize(ctx, rngInst, cfg)
		if err != nil {
			rt.Fatalf("synthesis failed: %v", err)
		}

		// Find critical path
		var startID, bossID string
		for id, room := range g.Rooms {
			if room.Archetype == graph.ArchetypeStart {
				startID = id
			}
			if room.Archetype == graph.ArchetypeBoss {
				bossID = id
			}
		}

		path, err := g.GetPath(startID, bossID)
		if err != nil {
			rt.Fatalf("no path from Start to Boss: %v", err)
		}

		// Verify difficulties follow custom interpolation
		for i, roomID := range path {
			progress := float64(i) / float64(len(path)-1)
			actual := g.Rooms[roomID].Difficulty
			expected := interpolateCustomCurve(progress, customPoints)

			// Allow some tolerance for interpolation
			if math.Abs(actual-expected) > 0.2 {
				rt.Fatalf("custom curve interpolation failed at progress=%.2f: actual=%.2f, expected=%.2f",
					progress, actual, expected)
			}
		}

		rt.Logf("✓ Custom curve with %d points interpolated correctly", numPoints)
	})
}

// calculateExpectedDifficulty returns the expected difficulty at a given
// progress point (0.0-1.0) for a specific pacing curve type.
func calculateExpectedDifficulty(progress float64, curve string) float64 {
	switch curve {
	case "LINEAR":
		// Linear: y = x
		return progress

	case "S_CURVE":
		// S-curve using logistic function (matching our SCurve implementation)
		k := 10.0 // Steepness
		sigmoid := 1.0 / (1.0 + math.Exp(-k*(progress-0.5)))
		minVal := 1.0 / (1.0 + math.Exp(k*0.5))
		maxVal := 1.0 / (1.0 + math.Exp(-k*0.5))
		return (sigmoid - minVal) / (maxVal - minVal)

	case "EXPONENTIAL":
		// Exponential: y = x² (matching our ExponentialCurve)
		return math.Pow(progress, 2.0)

	default:
		return progress // Fallback to linear
	}
}

// interpolateCustomCurve performs linear interpolation between custom control points.
func interpolateCustomCurve(progress float64, points [][2]float64) float64 {
	if len(points) == 0 {
		return progress
	}

	// Find the two control points to interpolate between
	for i := 0; i < len(points)-1; i++ {
		if progress >= points[i][0] && progress <= points[i+1][0] {
			// Linear interpolation between points[i] and points[i+1]
			x0, y0 := points[i][0], points[i][1]
			x1, y1 := points[i+1][0], points[i+1][1]

			if x1 == x0 {
				return y0
			}

			t := (progress - x0) / (x1 - x0)
			return y0 + t*(y1-y0)
		}
	}

	// If we're beyond the last point, return its difficulty
	return points[len(points)-1][1]
}

// calculateMAD computes Mean Absolute Deviation between two slices.
func calculateMAD(actual, expected []float64) float64 {
	if len(actual) != len(expected) || len(actual) == 0 {
		return 0.0
	}

	sum := 0.0
	for i := range actual {
		sum += math.Abs(actual[i] - expected[i])
	}

	return sum / float64(len(actual))
}
