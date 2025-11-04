package validation

import (
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// TestCheckPacingDeviation_AllCurveTypes verifies pacing validation works with all curve types.
func TestCheckPacingDeviation_AllCurveTypes(t *testing.T) {
	tests := []struct {
		name       string
		curveType  dungeon.PacingCurve
		customPts  [][2]float64
		wantMetric string
	}{
		{
			name:       "linear_curve",
			curveType:  dungeon.PacingLinear,
			wantMetric: "Pacing",
		},
		{
			name:       "s_curve",
			curveType:  dungeon.PacingSCurve,
			wantMetric: "Pacing",
		},
		{
			name:       "exponential_curve",
			curveType:  dungeon.PacingExponential,
			wantMetric: "Pacing",
		},
		{
			name:       "custom_curve",
			curveType:  dungeon.PacingCustom,
			customPts:  [][2]float64{{0.0, 0.0}, {0.5, 0.3}, {1.0, 1.0}},
			wantMetric: "Pacing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test graph with a linear path
			g := createLinearTestGraph(10)

			// Create config with the specific curve type
			cfg := &dungeon.Config{
				Pacing: dungeon.PacingCfg{
					Curve:        tt.curveType,
					Variance:     0.2,
					CustomPoints: tt.customPts,
				},
			}

			// Run the pacing deviation check
			result := CheckPacingDeviation(g, cfg)

			// Verify result structure
			if result.Constraint.Kind != tt.wantMetric {
				t.Errorf("CheckPacingDeviation() kind = %q, want %q", result.Constraint.Kind, tt.wantMetric)
			}

			if result.Constraint.Severity != "soft" {
				t.Errorf("CheckPacingDeviation() severity = %q, want %q", result.Constraint.Severity, "soft")
			}

			// Score should be in valid range [0.0, 1.0]
			if result.Score < 0.0 || result.Score > 1.0 {
				t.Errorf("CheckPacingDeviation() score = %v, want [0.0, 1.0]", result.Score)
			}

			// Result should have details
			if result.Details == "" {
				t.Errorf("CheckPacingDeviation() details is empty")
			}

			t.Logf("%s: score=%.3f, details=%s", tt.name, result.Score, result.Details)
		})
	}
}

// TestCheckPacingDeviation_MissingRooms verifies handling of missing Start/Boss.
func TestCheckPacingDeviation_MissingRooms(t *testing.T) {
	tests := []struct {
		name          string
		setupGraph    func() *graph.Graph
		wantDeviation float64 // Maximum deviation expected
	}{
		{
			name: "missing_start",
			setupGraph: func() *graph.Graph {
				g := createLinearTestGraph(5)
				// Remove Start archetype
				for _, room := range g.Rooms {
					if room.Archetype == graph.ArchetypeStart {
						room.Archetype = graph.ArchetypeOptional
					}
				}
				return g
			},
			wantDeviation: 1.0, // Should return max deviation
		},
		{
			name: "missing_boss",
			setupGraph: func() *graph.Graph {
				g := createLinearTestGraph(5)
				// Remove Boss archetype
				for _, room := range g.Rooms {
					if room.Archetype == graph.ArchetypeBoss {
						room.Archetype = graph.ArchetypeOptional
					}
				}
				return g
			},
			wantDeviation: 1.0, // Should return max deviation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setupGraph()

			cfg := &dungeon.Config{
				Pacing: dungeon.PacingCfg{
					Curve:    dungeon.PacingLinear,
					Variance: 0.2,
				},
			}

			// Calculate deviation directly
			deviation := CalculatePacingDeviation(g, cfg)

			if deviation != tt.wantDeviation {
				t.Errorf("CalculatePacingDeviation() = %v, want %v", deviation, tt.wantDeviation)
			}

			// Check via constraint checker
			result := CheckPacingDeviation(g, cfg)
			expectedScore := 1.0 - tt.wantDeviation

			if result.Score != expectedScore {
				t.Errorf("CheckPacingDeviation().Score = %v, want %v", result.Score, expectedScore)
			}
		})
	}
}

// TestCheckPacingDeviation_PerfectAdherence verifies score calculation.
func TestCheckPacingDeviation_PerfectAdherence(t *testing.T) {
	// Create a graph where difficulties match the linear curve exactly
	g := &graph.Graph{
		Rooms:      make(map[string]*graph.Room),
		Adjacency:  make(map[string][]string),
		Connectors: make(map[string]*graph.Connector),
	}

	n := 10
	for i := 0; i < n; i++ {
		id := roomID(i)
		archetype := graph.ArchetypeOptional
		if i == 0 {
			archetype = graph.ArchetypeStart
		} else if i == n-1 {
			archetype = graph.ArchetypeBoss
		}

		// Set difficulty to match linear curve exactly
		difficulty := float64(i) / float64(n-1)

		g.Rooms[id] = &graph.Room{
			ID:         id,
			Archetype:  archetype,
			Difficulty: difficulty,
		}

		// Create linear path
		g.Adjacency[id] = make([]string, 0)
		if i > 0 {
			prevID := roomID(i - 1)
			g.Adjacency[id] = append(g.Adjacency[id], prevID)
			g.Adjacency[prevID] = append(g.Adjacency[prevID], id)

			connID := prevID + "_" + id
			g.Connectors[connID] = &graph.Connector{
				ID:            connID,
				From:          prevID,
				To:            id,
				Type:          graph.TypeCorridor,
				Cost:          1.0,
				Visibility:    graph.VisibilityNormal,
				Bidirectional: true,
			}
		}
	}

	cfg := &dungeon.Config{
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.2,
		},
	}

	deviation := CalculatePacingDeviation(g, cfg)

	// With perfect adherence, deviation should be very close to 0
	if deviation > 0.01 {
		t.Errorf("Perfect linear adherence has deviation %v, want near 0.0", deviation)
	}

	result := CheckPacingDeviation(g, cfg)

	// Score should be very high (near 1.0)
	if result.Score < 0.99 {
		t.Errorf("Perfect adherence score = %v, want near 1.0", result.Score)
	}

	t.Logf("Perfect adherence: deviation=%.6f, score=%.3f", deviation, result.Score)
}

// TestCheckPacingDeviation_CustomCurveInterpolation verifies custom curve support.
func TestCheckPacingDeviation_CustomCurveInterpolation(t *testing.T) {
	// Create a graph with custom difficulties matching a specific curve
	g := createLinearTestGraph(10)

	// Define custom curve: flat start, steep end
	customPoints := [][2]float64{
		{0.0, 0.0},
		{0.6, 0.2}, // Slow increase
		{1.0, 1.0}, // Rapid increase at end
	}

	cfg := &dungeon.Config{
		Pacing: dungeon.PacingCfg{
			Curve:        dungeon.PacingCustom,
			Variance:     0.3,
			CustomPoints: customPoints,
		},
	}

	result := CheckPacingDeviation(g, cfg)

	// Should successfully compute deviation with custom curve
	if result.Score < 0.0 || result.Score > 1.0 {
		t.Errorf("Custom curve score = %v, want [0.0, 1.0]", result.Score)
	}

	t.Logf("Custom curve: score=%.3f, details=%s", result.Score, result.Details)
}

// Helper functions

func createLinearTestGraph(n int) *graph.Graph {
	g := &graph.Graph{
		Rooms:      make(map[string]*graph.Room),
		Adjacency:  make(map[string][]string),
		Connectors: make(map[string]*graph.Connector),
	}

	for i := 0; i < n; i++ {
		id := roomID(i)
		archetype := graph.ArchetypeOptional
		if i == 0 {
			archetype = graph.ArchetypeStart
		} else if i == n-1 {
			archetype = graph.ArchetypeBoss
		}

		g.Rooms[id] = &graph.Room{
			ID:         id,
			Archetype:  archetype,
			Difficulty: float64(i) / float64(n),
			Size:       graph.SizeM,
		}

		g.Adjacency[id] = make([]string, 0)
		if i > 0 {
			prevID := roomID(i - 1)
			g.Adjacency[id] = append(g.Adjacency[id], prevID)
			g.Adjacency[prevID] = append(g.Adjacency[prevID], id)

			connID := prevID + "_" + id
			g.Connectors[connID] = &graph.Connector{
				ID:            connID,
				From:          prevID,
				To:            id,
				Type:          graph.TypeCorridor,
				Cost:          1.0,
				Visibility:    graph.VisibilityNormal,
				Bidirectional: true,
			}
		}
	}

	return g
}

func roomID(i int) string {
	return "room" + string(rune('0'+i))
}
