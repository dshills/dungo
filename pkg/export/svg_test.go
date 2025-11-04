package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// T095: Test basic SVG export functionality
func TestExportSVG_Basic(t *testing.T) {
	artifact := createSVGTestArtifact(t)

	opts := DefaultSVGOptions()
	opts.Title = "Test Dungeon"

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ExportSVG returned empty data")
	}

	// Verify it's valid SVG
	svgStr := string(data)
	if !strings.Contains(svgStr, "<svg") {
		t.Error("Output does not contain <svg> tag")
	}
	if !strings.Contains(svgStr, "</svg>") {
		t.Error("Output does not contain closing </svg> tag")
	}
}

// T095: Test SVG export with nil artifact
func TestExportSVG_NilArtifact(t *testing.T) {
	opts := DefaultSVGOptions()
	_, err := ExportSVG(nil, opts)
	if err == nil {
		t.Error("Expected error for nil artifact, got nil")
	}
}

// T095: Test SVG export with nil graph
func TestExportSVG_NilGraph(t *testing.T) {
	artifact := &dungeon.Artifact{}
	opts := DefaultSVGOptions()
	_, err := ExportSVG(artifact, opts)
	if err == nil {
		t.Error("Expected error for artifact with nil ADG, got nil")
	}
}

// T095: Test SVG export integration with existing test helper
func TestExportSVG_WithStandardArtifact(t *testing.T) {
	artifact := createTestArtifactForSVG(t)

	opts := DefaultSVGOptions()
	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ExportSVG returned empty data")
	}
}

// T095: Test default options
func TestDefaultSVGOptions(t *testing.T) {
	opts := DefaultSVGOptions()

	if opts.Width <= 0 {
		t.Errorf("Width should be positive, got %d", opts.Width)
	}
	if opts.Height <= 0 {
		t.Errorf("Height should be positive, got %d", opts.Height)
	}
	if opts.NodeRadius <= 0 {
		t.Errorf("NodeRadius should be positive, got %d", opts.NodeRadius)
	}
	if opts.EdgeWidth <= 0 {
		t.Errorf("EdgeWidth should be positive, got %d", opts.EdgeWidth)
	}
	if !opts.ShowLabels {
		t.Error("ShowLabels should be true by default")
	}
	if !opts.ColorByType {
		t.Error("ColorByType should be true by default")
	}
	if !opts.ShowLegend {
		t.Error("ShowLegend should be true by default")
	}
}

// T095: Test option validation (invalid dimensions should be corrected)
func TestExportSVG_InvalidOptions(t *testing.T) {
	artifact := createTestArtifactForSVG(t)

	opts := SVGOptions{
		Width:      -100,
		Height:     -100,
		NodeRadius: -10,
		EdgeWidth:  -5,
		Margin:     -20,
	}

	// Should not error even with invalid options (should use defaults)
	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG should handle invalid options gracefully: %v", err)
	}
	if len(data) == 0 {
		t.Error("Should still produce output with invalid options")
	}
}

// T106: Test node-edge graph visualization
func TestExportSVG_NodeEdgeGraph(t *testing.T) {
	artifact := createTestArtifactForSVG(t)

	opts := DefaultSVGOptions()
	opts.ShowLabels = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	svgStr := string(data)

	// Check for circles (nodes)
	if !strings.Contains(svgStr, "<circle") {
		t.Error("SVG should contain circle elements for nodes")
	}

	// Check for lines (edges)
	if !strings.Contains(svgStr, "<line") {
		t.Error("SVG should contain line elements for edges")
	}

	// Check for text labels
	if !strings.Contains(svgStr, "<text") {
		t.Error("SVG should contain text elements for labels")
	}
}

// T107: Test color-coding by room archetype
func TestExportSVG_ColorByArchetype(t *testing.T) {
	artifact := createTestArtifactForSVG(t)

	opts := DefaultSVGOptions()
	opts.ColorByType = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	svgStr := string(data)

	// Verify different colors are used
	// Start room should be green (#48bb78)
	if !strings.Contains(svgStr, "#48bb78") {
		t.Error("Expected green color for Start room")
	}

	// Boss room should be red (#f56565)
	if !strings.Contains(svgStr, "#f56565") {
		t.Error("Expected red color for Boss room")
	}

	// Treasure room should be gold (#ffd700)
	if !strings.Contains(svgStr, "#ffd700") {
		t.Error("Expected gold color for Treasure room")
	}
}

// T107: Test color-coding disabled
func TestExportSVG_ColorByArchetypeDisabled(t *testing.T) {
	artifact := createTestArtifactForSVG(t)

	opts := DefaultSVGOptions()
	opts.ColorByType = false

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	// Should still produce valid SVG
	if len(data) == 0 {
		t.Error("Should produce output even with ColorByType disabled")
	}
}

// T108: Test difficulty heatmap overlay
func TestExportSVG_DifficultyHeatmap(t *testing.T) {
	artifact := createTestArtifactForSVG(t)

	opts := DefaultSVGOptions()
	opts.ShowHeatmap = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	svgStr := string(data)

	// Check for heatmap colors
	// Should have blue, green, yellow, or red based on difficulty
	hasHeatmapColors := strings.Contains(svgStr, "#3b82f6") || // Blue
		strings.Contains(svgStr, "#10b981") || // Green
		strings.Contains(svgStr, "#f59e0b") || // Yellow
		strings.Contains(svgStr, "#ef4444") // Red

	if !hasHeatmapColors {
		t.Error("Expected heatmap colors in output")
	}
}

// T108: Test heatmap with varying difficulty levels
func TestExportSVG_HeatmapDifficultyRange(t *testing.T) {
	// Create artifact with rooms at different difficulty levels
	g := graph.NewGraph(12345)

	// Low difficulty
	g.AddRoom(&graph.Room{
		ID:         "room_easy",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Difficulty: 0.1,
	})

	// Medium difficulty
	g.AddRoom(&graph.Room{
		ID:         "room_medium",
		Archetype:  graph.ArchetypeHub,
		Size:       graph.SizeM,
		Difficulty: 0.5,
	})

	// High difficulty
	g.AddRoom(&graph.Room{
		ID:         "room_hard",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeL,
		Difficulty: 0.9,
	})

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}

	opts := DefaultSVGOptions()
	opts.ShowHeatmap = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should produce output with difficulty heatmap")
	}
}

// T109: Test legend rendering
func TestExportSVG_Legend(t *testing.T) {
	artifact := createTestArtifactForSVG(t)

	opts := DefaultSVGOptions()
	opts.ShowLegend = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	svgStr := string(data)

	// Check for legend content
	legendItems := []string{
		"Room Types",
		"Start",
		"Boss",
		"Treasure",
		"Puzzle",
		"Connections",
		"Door",
		"Corridor",
	}

	for _, item := range legendItems {
		if !strings.Contains(svgStr, item) {
			t.Errorf("Legend should contain '%s'", item)
		}
	}
}

// T109: Test title and stats rendering
func TestExportSVG_TitleAndStats(t *testing.T) {
	artifact := createTestArtifactForSVG(t)
	artifact.Metrics = &dungeon.Metrics{
		BranchingFactor: 2.5,
		PathLength:      10,
		CycleCount:      2,
		PacingDeviation: 0.123,
	}

	opts := DefaultSVGOptions()
	opts.Title = "Test Dungeon Graph"
	opts.ShowStats = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	svgStr := string(data)

	// Check for title
	if !strings.Contains(svgStr, "Test Dungeon Graph") {
		t.Error("SVG should contain the title")
	}

	// Check for stats
	if !strings.Contains(svgStr, "Rooms:") {
		t.Error("SVG should contain room count")
	}
	if !strings.Contains(svgStr, "Connectors:") {
		t.Error("SVG should contain connector count")
	}
	if !strings.Contains(svgStr, "Seed:") {
		t.Error("SVG should contain seed")
	}

	// Check for metrics
	if !strings.Contains(svgStr, "Branch Factor:") {
		t.Error("SVG should contain branch factor")
	}
	if !strings.Contains(svgStr, "Path Length:") {
		t.Error("SVG should contain path length")
	}
}

// T109: Test annotations without metrics
func TestExportSVG_StatsWithoutMetrics(t *testing.T) {
	artifact := createTestArtifactForSVG(t)
	artifact.Metrics = nil // No metrics

	opts := DefaultSVGOptions()
	opts.ShowStats = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	// Should not error even without metrics
	if len(data) == 0 {
		t.Error("Should produce output even without metrics")
	}
}

// T097: Test SVG export integration
func TestExportSVG_FullIntegration(t *testing.T) {
	artifact := createComplexSVGArtifact(t)

	opts := DefaultSVGOptions()
	opts.Title = "Complex Dungeon Test"
	opts.ShowLabels = true
	opts.ColorByType = true
	opts.ShowHeatmap = false
	opts.ShowLegend = true
	opts.ShowStats = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ExportSVG returned empty data")
	}

	// Optionally write to file for manual inspection
	if os.Getenv("WRITE_TEST_OUTPUT") == "1" {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "test_dungeon.svg")
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			t.Logf("Could not write test output: %v", err)
		} else {
			t.Logf("Test SVG written to: %s", outputPath)
		}
	}
}

// T097: Test all connector types
func TestExportSVG_AllConnectorTypes(t *testing.T) {
	g := graph.NewGraph(54321)

	// Create rooms
	rooms := []struct {
		id        string
		archetype graph.RoomArchetype
	}{
		{"room1", graph.ArchetypeStart},
		{"room2", graph.ArchetypeHub},
		{"room3", graph.ArchetypeTreasure},
		{"room4", graph.ArchetypePuzzle},
		{"room5", graph.ArchetypeBoss},
	}

	for _, r := range rooms {
		g.AddRoom(&graph.Room{
			ID:        r.id,
			Archetype: r.archetype,
			Size:      graph.SizeM,
		})
	}

	// Create connectors of different types
	connectorTypes := []graph.ConnectorType{
		graph.TypeDoor,
		graph.TypeCorridor,
		graph.TypeLadder,
		graph.TypeTeleporter,
		graph.TypeHidden,
		graph.TypeOneWay,
	}

	for i, connType := range connectorTypes {
		from := rooms[i%len(rooms)].id
		to := rooms[(i+1)%len(rooms)].id
		g.AddConnector(&graph.Connector{
			ID:            string(rune('A' + i)),
			From:          from,
			To:            to,
			Type:          connType,
			Cost:          1.0,
			Bidirectional: connType != graph.TypeOneWay,
		})
	}

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}

	opts := DefaultSVGOptions()
	opts.ColorByType = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should produce output with all connector types")
	}
}

// T097: Test gates on connectors
func TestExportSVG_GatedConnectors(t *testing.T) {
	g := graph.NewGraph(99999)

	g.AddRoom(&graph.Room{ID: "start", Archetype: graph.ArchetypeStart, Size: graph.SizeM})
	g.AddRoom(&graph.Room{ID: "locked", Archetype: graph.ArchetypeTreasure, Size: graph.SizeM})

	g.AddConnector(&graph.Connector{
		ID:   "gated_conn",
		From: "start",
		To:   "locked",
		Type: graph.TypeDoor,
		Gate: &graph.Gate{
			Type:  "key",
			Value: "silver_key",
		},
		Cost:          1.0,
		Bidirectional: true,
	})

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}

	opts := DefaultSVGOptions()
	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	svgStr := string(data)

	// Should have gate indicator (gold circle)
	if !strings.Contains(svgStr, "#ffd700") {
		t.Error("Expected gate indicator color in output")
	}
}

// T097: Test empty graph
func TestExportSVG_EmptyGraph(t *testing.T) {
	g := graph.NewGraph(11111)
	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}

	opts := DefaultSVGOptions()
	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG should handle empty graph: %v", err)
	}

	// Should produce valid SVG even with no rooms
	if len(data) == 0 {
		t.Error("Should produce output even for empty graph")
	}
}

// Helper: Create a basic test artifact with a few rooms and connections
func createTestArtifactForSVG(t *testing.T) *dungeon.Artifact {
	t.Helper()

	g := graph.NewGraph(42)

	// Add rooms
	g.AddRoom(&graph.Room{
		ID:         "start",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Difficulty: 0.1,
	})

	g.AddRoom(&graph.Room{
		ID:         "treasure",
		Archetype:  graph.ArchetypeTreasure,
		Size:       graph.SizeS,
		Difficulty: 0.5,
	})

	g.AddRoom(&graph.Room{
		ID:         "boss",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeXL,
		Difficulty: 0.9,
	})

	// Add connectors
	g.AddConnector(&graph.Connector{
		ID:            "conn1",
		From:          "start",
		To:            "treasure",
		Type:          graph.TypeDoor,
		Cost:          1.0,
		Bidirectional: true,
	})

	g.AddConnector(&graph.Connector{
		ID:            "conn2",
		From:          "treasure",
		To:            "boss",
		Type:          graph.TypeCorridor,
		Cost:          1.5,
		Bidirectional: true,
	})

	return &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
		Metrics: &dungeon.Metrics{
			BranchingFactor: 2.0,
			PathLength:      3,
			CycleCount:      0,
			PacingDeviation: 0.05,
		},
	}
}

// Helper: Create a more complex artifact for integration testing
func createComplexSVGArtifact(t *testing.T) *dungeon.Artifact {
	t.Helper()

	g := graph.NewGraph(98765)

	// Create a variety of room types
	roomData := []struct {
		id         string
		archetype  graph.RoomArchetype
		size       graph.RoomSize
		difficulty float64
	}{
		{"start", graph.ArchetypeStart, graph.SizeM, 0.1},
		{"hub1", graph.ArchetypeHub, graph.SizeL, 0.2},
		{"treasure1", graph.ArchetypeTreasure, graph.SizeS, 0.3},
		{"puzzle1", graph.ArchetypePuzzle, graph.SizeM, 0.4},
		{"hub2", graph.ArchetypeHub, graph.SizeL, 0.5},
		{"secret1", graph.ArchetypeSecret, graph.SizeXS, 0.6},
		{"optional1", graph.ArchetypeOptional, graph.SizeM, 0.7},
		{"checkpoint", graph.ArchetypeCheckpoint, graph.SizeM, 0.75},
		{"boss", graph.ArchetypeBoss, graph.SizeXL, 1.0},
	}

	for _, rd := range roomData {
		g.AddRoom(&graph.Room{
			ID:         rd.id,
			Archetype:  rd.archetype,
			Size:       rd.size,
			Difficulty: rd.difficulty,
		})
	}

	// Create connections
	connections := []struct {
		from  string
		to    string
		cType graph.ConnectorType
		vis   graph.VisibilityType
	}{
		{"start", "hub1", graph.TypeDoor, graph.VisibilityNormal},
		{"hub1", "treasure1", graph.TypeCorridor, graph.VisibilityNormal},
		{"hub1", "puzzle1", graph.TypeDoor, graph.VisibilityNormal},
		{"puzzle1", "hub2", graph.TypeCorridor, graph.VisibilityNormal},
		{"hub2", "checkpoint", graph.TypeDoor, graph.VisibilityNormal},
		{"checkpoint", "boss", graph.TypeDoor, graph.VisibilityNormal},
		{"hub1", "secret1", graph.TypeHidden, graph.VisibilitySecret},
		{"hub2", "optional1", graph.TypeDoor, graph.VisibilityNormal},
		{"optional1", "boss", graph.TypeOneWay, graph.VisibilityNormal},
	}

	for i, conn := range connections {
		g.AddConnector(&graph.Connector{
			ID:            string(rune('A' + i)),
			From:          conn.from,
			To:            conn.to,
			Type:          conn.cType,
			Visibility:    conn.vis,
			Cost:          1.0,
			Bidirectional: conn.cType != graph.TypeOneWay,
		})
	}

	return &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
		Metrics: &dungeon.Metrics{
			BranchingFactor: 2.3,
			PathLength:      6,
			CycleCount:      1,
			PacingDeviation: 0.08,
		},
	}
}

// Helper: Create minimal SVG test artifact
func createSVGTestArtifact(t *testing.T) *dungeon.Artifact {
	t.Helper()

	g := graph.NewGraph(123)
	g.AddRoom(&graph.Room{
		ID:        "room1",
		Archetype: graph.ArchetypeStart,
		Size:      graph.SizeM,
	})
	g.AddRoom(&graph.Room{
		ID:        "room2",
		Archetype: graph.ArchetypeBoss,
		Size:      graph.SizeM,
	})
	g.AddConnector(&graph.Connector{
		ID:            "conn1",
		From:          "room1",
		To:            "room2",
		Type:          graph.TypeDoor,
		Cost:          1.0,
		Bidirectional: true,
	})

	return &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}
}

// =====================================================================
// TDD RED PHASE TESTS: T095 and T097
// These tests define requirements for SVG export functionality.
// They will fail until proper implementation is complete.
// =====================================================================

// T095: Test SVG contains all required graph visualization elements
func TestSVGGeneration_AllElements(t *testing.T) {
	artifact := createComplexSVGArtifact(t)

	opts := DefaultSVGOptions()
	opts.ShowLabels = true
	opts.ShowLegend = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("ExportSVG failed: %v", err)
	}

	svg := string(data)

	t.Run("SVGTags", func(t *testing.T) {
		if !strings.Contains(svg, "<svg") {
			t.Error("SVG should contain opening <svg tag")
		}
		if !strings.Contains(svg, "</svg>") {
			t.Error("SVG should contain closing </svg> tag")
		}
	})

	t.Run("NodesForAllRooms", func(t *testing.T) {
		roomCount := len(artifact.ADG.Rooms)
		// Should have nodes (circles or other shapes) for each room
		nodeCount := strings.Count(svg, "<circle") +
			strings.Count(svg, "<rect") +
			strings.Count(svg, "<polygon")

		if nodeCount < roomCount {
			t.Errorf("Expected at least %d room nodes, found %d",
				roomCount, nodeCount)
		}
	})

	t.Run("EdgesForAllConnectors", func(t *testing.T) {
		connectorCount := len(artifact.ADG.Connectors)
		// Should have edges (lines or paths) for each connector
		edgeCount := strings.Count(svg, "<line") +
			strings.Count(svg, "<path") +
			strings.Count(svg, "<polyline")

		if edgeCount < connectorCount {
			t.Errorf("Expected at least %d connector edges, found %d",
				connectorCount, edgeCount)
		}
	})

	t.Run("LegendPresent", func(t *testing.T) {
		// Legend should explain symbols
		if !strings.Contains(svg, "<text") {
			t.Error("SVG should contain <text> elements for legend")
		}

		legendTerms := []string{"Start", "Boss", "Treasure"}
		for _, term := range legendTerms {
			if !strings.Contains(svg, term) {
				t.Errorf("Legend should contain term: %s", term)
			}
		}
	})

	t.Run("TitleAnnotations", func(t *testing.T) {
		// Should have metadata
		hasTitle := strings.Contains(svg, "<title") || strings.Contains(svg, "Dungeon")
		if !hasTitle {
			t.Error("SVG should contain title or dungeon identifier")
		}
	})
}

// T097: Test SVG consistency - same artifact produces identical SVG
func TestSVGGoldenConsistency(t *testing.T) {
	// Fixed seed for deterministic generation
	seed := uint64(424242)

	// Create same artifact twice
	createArtifact := func() *dungeon.Artifact {
		g := graph.NewGraph(seed)

		rooms := []struct {
			id        string
			archetype graph.RoomArchetype
			size      graph.RoomSize
			diff      float64
		}{
			{"start", graph.ArchetypeStart, graph.SizeM, 0.1},
			{"hall", graph.ArchetypeHub, graph.SizeL, 0.3},
			{"treasure", graph.ArchetypeTreasure, graph.SizeS, 0.5},
			{"boss", graph.ArchetypeBoss, graph.SizeXL, 0.9},
		}

		for _, r := range rooms {
			g.AddRoom(&graph.Room{
				ID:         r.id,
				Archetype:  r.archetype,
				Size:       r.size,
				Difficulty: r.diff,
			})
		}

		g.AddConnector(&graph.Connector{
			ID:            "conn1",
			From:          "start",
			To:            "hall",
			Type:          graph.TypeDoor,
			Cost:          1.0,
			Bidirectional: true,
		})
		g.AddConnector(&graph.Connector{
			ID:            "conn2",
			From:          "hall",
			To:            "treasure",
			Type:          graph.TypeCorridor,
			Cost:          1.5,
			Bidirectional: true,
		})
		g.AddConnector(&graph.Connector{
			ID:            "conn3",
			From:          "hall",
			To:            "boss",
			Type:          graph.TypeDoor,
			Cost:          2.0,
			Bidirectional: true,
		})

		return &dungeon.Artifact{
			ADG: &dungeon.Graph{Graph: g},
			Metrics: &dungeon.Metrics{
				BranchingFactor: 1.5,
				PathLength:      3,
				CycleCount:      0,
			},
		}
	}

	artifact1 := createArtifact()
	artifact2 := createArtifact()

	opts := DefaultSVGOptions()
	opts.Title = "Golden Test Dungeon"
	opts.ShowLabels = true
	opts.ShowLegend = true
	opts.ShowStats = true

	// Export both
	svg1, err := ExportSVG(artifact1, opts)
	if err != nil {
		t.Fatalf("First SVG export failed: %v", err)
	}

	svg2, err := ExportSVG(artifact2, opts)
	if err != nil {
		t.Fatalf("Second SVG export failed: %v", err)
	}

	// Compare byte-for-byte
	if string(svg1) != string(svg2) {
		t.Error("SVG export is not deterministic: same artifact produced different SVG")
		t.Logf("SVG 1 length: %d bytes", len(svg1))
		t.Logf("SVG 2 length: %d bytes", len(svg2))

		// Show first difference
		lines1 := strings.Split(string(svg1), "\n")
		lines2 := strings.Split(string(svg2), "\n")
		for i := 0; i < len(lines1) && i < len(lines2); i++ {
			if lines1[i] != lines2[i] {
				t.Logf("First diff at line %d:", i+1)
				t.Logf("  SVG1: %s", lines1[i])
				t.Logf("  SVG2: %s", lines2[i])
				break
			}
		}
	}

	t.Log("SVG consistency test passed - deterministic output verified")
}

// T097: Test golden file comparison for regression detection
func TestSVGGoldenFile(t *testing.T) {
	artifact := createComplexSVGArtifact(t)

	opts := DefaultSVGOptions()
	opts.Title = "Golden Test Dungeon"
	opts.ShowLabels = true
	opts.ColorByType = true
	opts.ShowLegend = true
	opts.ShowStats = true

	data, err := ExportSVG(artifact, opts)
	if err != nil {
		t.Fatalf("SVG export failed: %v", err)
	}

	// Golden file path
	goldenDir := filepath.Join("..", "..", "testdata", "golden")
	goldenPath := filepath.Join(goldenDir, "dungeon_complex.svg")

	// Check if we should update golden files
	updateGolden := os.Getenv("UPDATE_GOLDEN") == "1"

	if updateGolden {
		// Create directory if needed
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatalf("Failed to create golden directory: %v", err)
		}

		// Write golden file
		if err := os.WriteFile(goldenPath, data, 0644); err != nil {
			t.Fatalf("Failed to write golden file: %v", err)
		}

		t.Logf("Golden file updated: %s", goldenPath)
		return
	}

	// Read and compare golden file
	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("Golden file does not exist: %s (run with UPDATE_GOLDEN=1 to create)",
				goldenPath)
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare
	if string(data) != string(goldenData) {
		t.Errorf("SVG output differs from golden file")
		t.Logf("Golden: %s", goldenPath)
		t.Logf("Generated: %d bytes, Golden: %d bytes", len(data), len(goldenData))

		// Show line differences
		genLines := strings.Split(string(data), "\n")
		goldLines := strings.Split(string(goldenData), "\n")

		if len(genLines) != len(goldLines) {
			t.Logf("Line count: generated=%d, golden=%d", len(genLines), len(goldLines))
		}

		for i := 0; i < len(genLines) && i < len(goldLines); i++ {
			if genLines[i] != goldLines[i] {
				t.Logf("First diff at line %d:", i+1)
				t.Logf("  Generated: %s", genLines[i])
				t.Logf("  Golden:    %s", goldLines[i])
				break
			}
		}
	}

	t.Log("SVG matches golden file - no regressions detected")
}
