package export

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// TestGenerateSampleSVG creates a sample SVG file for manual inspection.
// Run with: go test -v -run TestGenerateSampleSVG ./pkg/export
// The output will be written to a temporary directory and the path will be printed.
func TestGenerateSampleSVG(t *testing.T) {
	// Create a more interesting test graph
	g := graph.NewGraph(123456)

	// Add diverse rooms
	rooms := []struct {
		id         string
		archetype  graph.RoomArchetype
		size       graph.RoomSize
		difficulty float64
	}{
		{"entrance", graph.ArchetypeStart, graph.SizeM, 0.0},
		{"hub_main", graph.ArchetypeHub, graph.SizeL, 0.2},
		{"treasure_1", graph.ArchetypeTreasure, graph.SizeS, 0.3},
		{"puzzle_room", graph.ArchetypePuzzle, graph.SizeM, 0.4},
		{"secret_passage", graph.ArchetypeSecret, graph.SizeXS, 0.5},
		{"vendor_shop", graph.ArchetypeVendor, graph.SizeS, 0.3},
		{"shrine", graph.ArchetypeShrine, graph.SizeM, 0.4},
		{"checkpoint", graph.ArchetypeCheckpoint, graph.SizeM, 0.6},
		{"optional_wing", graph.ArchetypeOptional, graph.SizeL, 0.7},
		{"mini_boss", graph.ArchetypeCorridor, graph.SizeM, 0.8},
		{"boss_arena", graph.ArchetypeBoss, graph.SizeXL, 1.0},
	}

	for _, r := range rooms {
		if err := g.AddRoom(&graph.Room{
			ID:         r.id,
			Archetype:  r.archetype,
			Size:       r.size,
			Difficulty: r.difficulty,
			Reward:     r.difficulty * 0.8,
		}); err != nil {
			t.Fatalf("Failed to add room: %v", err)
		}
	}

	// Add diverse connections
	connections := []struct {
		from string
		to   string
		typ  graph.ConnectorType
		vis  graph.VisibilityType
		gate *graph.Gate
	}{
		{"entrance", "hub_main", graph.TypeDoor, graph.VisibilityNormal, nil},
		{"hub_main", "treasure_1", graph.TypeCorridor, graph.VisibilityNormal, nil},
		{"hub_main", "puzzle_room", graph.TypeDoor, graph.VisibilityNormal, nil},
		{"hub_main", "vendor_shop", graph.TypeDoor, graph.VisibilityNormal, nil},
		{"hub_main", "secret_passage", graph.TypeHidden, graph.VisibilitySecret, nil},
		{"puzzle_room", "shrine", graph.TypeDoor, graph.VisibilityNormal, &graph.Gate{Type: "key", Value: "silver"}},
		{"shrine", "checkpoint", graph.TypeDoor, graph.VisibilityNormal, nil},
		{"checkpoint", "optional_wing", graph.TypeDoor, graph.VisibilityNormal, nil},
		{"checkpoint", "mini_boss", graph.TypeCorridor, graph.VisibilityNormal, nil},
		{"mini_boss", "boss_arena", graph.TypeDoor, graph.VisibilityNormal, &graph.Gate{Type: "key", Value: "gold"}},
		{"optional_wing", "boss_arena", graph.TypeOneWay, graph.VisibilityNormal, nil},
		{"secret_passage", "treasure_1", graph.TypeHidden, graph.VisibilitySecret, nil},
	}

	connID := 0
	for _, c := range connections {
		connID++
		if err := g.AddConnector(&graph.Connector{
			ID:            string(rune('A' + connID - 1)),
			From:          c.from,
			To:            c.to,
			Type:          c.typ,
			Gate:          c.gate,
			Cost:          1.0,
			Visibility:    c.vis,
			Bidirectional: c.typ != graph.TypeOneWay,
		}); err != nil {
			t.Fatalf("Failed to add connector: %v", err)
		}
	}

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
		Metrics: &dungeon.Metrics{
			BranchingFactor: 2.1,
			PathLength:      7,
			CycleCount:      1,
			PacingDeviation: 0.12,
		},
	}

	// Generate SVG with all features enabled
	opts := DefaultSVGOptions()
	opts.Title = "Sample Dungeon Graph Visualization"
	opts.ShowLabels = true
	opts.ColorByType = true
	opts.ShowHeatmap = false // Disable to see node colors better
	opts.ShowLegend = true
	opts.ShowStats = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "sample_dungeon.svg")
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		t.Fatalf("Failed to write SVG file: %v", err)
	}

	t.Logf("Sample SVG written to: %s", outputPath)
	t.Logf("Open it in a browser to view the visualization")
	t.Logf("SVG size: %d bytes", len(data))

	// Also test with heatmap enabled
	opts.ShowHeatmap = true
	opts.Title = "Sample Dungeon Graph with Difficulty Heatmap"
	dataHeatmap, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG with heatmap failed: %v", err)
	}

	heatmapPath := filepath.Join(tmpDir, "sample_dungeon_heatmap.svg")
	if err := os.WriteFile(heatmapPath, dataHeatmap, 0644); err != nil {
		t.Fatalf("Failed to write heatmap SVG file: %v", err)
	}

	t.Logf("Sample SVG with heatmap written to: %s", heatmapPath)
}
