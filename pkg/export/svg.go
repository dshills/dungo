package export

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"

	svg "github.com/ajstarks/svgo"
	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// SVGOptions configures SVG visualization export.
type SVGOptions struct {
	Width       int    // Canvas width in pixels
	Height      int    // Canvas height in pixels
	ShowLabels  bool   // Show room ID labels
	ColorByType bool   // Color nodes by room archetype
	ShowHeatmap bool   // Show difficulty heatmap overlay
	ShowLegend  bool   // Show legend explaining colors/symbols
	NodeRadius  int    // Radius of room nodes (default: 20)
	EdgeWidth   int    // Width of connector lines (default: 2)
	Margin      int    // Canvas margin in pixels (default: 50)
	Title       string // Optional title for the visualization
	ShowStats   bool   // Show dungeon statistics
}

// DefaultSVGOptions returns sensible default SVG export options.
func DefaultSVGOptions() SVGOptions {
	return SVGOptions{
		Width:       1200,
		Height:      900,
		ShowLabels:  true,
		ColorByType: true,
		ShowHeatmap: false,
		ShowLegend:  true,
		NodeRadius:  20,
		EdgeWidth:   2,
		Margin:      60,
		Title:       "Dungeon Graph",
		ShowStats:   true,
	}
}

// ExportSVG generates an SVG visualization of the dungeon graph.
// Returns the SVG as a byte slice or an error if generation fails.
func ExportSVG(artifact *dungeon.Artifact, opts SVGOptions) ([]byte, error) {
	if artifact == nil {
		return nil, fmt.Errorf("artifact cannot be nil")
	}
	if artifact.ADG == nil || artifact.ADG.Graph == nil {
		return nil, fmt.Errorf("artifact must contain a valid ADG")
	}

	// Validate options
	if opts.Width <= 0 {
		opts.Width = 1200
	}
	if opts.Height <= 0 {
		opts.Height = 900
	}
	if opts.NodeRadius <= 0 {
		opts.NodeRadius = 20
	}
	if opts.EdgeWidth <= 0 {
		opts.EdgeWidth = 2
	}
	if opts.Margin <= 0 {
		opts.Margin = 60
	}

	// Create buffer for SVG output
	buf := new(bytes.Buffer)
	canvas := svg.New(buf)
	canvas.Start(opts.Width, opts.Height)

	// Add background
	canvas.Rect(0, 0, opts.Width, opts.Height, "fill:#1a1a2e")

	// Calculate layout positions for nodes
	positions := calculateLayout(artifact.ADG.Graph, opts)

	// Draw edges first (so they appear behind nodes)
	drawEdges(canvas, artifact.ADG.Graph, positions, opts)

	// Draw nodes
	drawNodes(canvas, artifact.ADG.Graph, positions, opts)

	// Draw labels if enabled
	if opts.ShowLabels {
		drawLabels(canvas, artifact.ADG.Graph, positions, opts)
	}

	// Draw heatmap overlay if enabled
	if opts.ShowHeatmap {
		drawHeatmap(canvas, artifact.ADG.Graph, positions, opts)
	}

	// Draw legend if enabled
	if opts.ShowLegend {
		drawLegend(canvas, opts)
	}

	// Draw title and stats if enabled
	if opts.Title != "" || opts.ShowStats {
		drawHeader(canvas, artifact, opts)
	}

	canvas.End()
	return buf.Bytes(), nil
}

// SaveSVGToFile generates an SVG visualization and saves it to a file.
// The file is created with 0644 permissions (readable by all, writable by owner).
func SaveSVGToFile(artifact *dungeon.Artifact, filepath string, opts SVGOptions) error {
	data, err := ExportSVG(artifact, opts)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// position represents a 2D coordinate.
type position struct {
	X, Y float64
}

// calculateLayout computes positions for all rooms using a force-directed layout.
// Returns a map from room ID to position.
func calculateLayout(g *graph.Graph, opts SVGOptions) map[string]position {
	positions := make(map[string]position)

	// If no rooms, return empty
	if len(g.Rooms) == 0 {
		return positions
	}

	// Calculate drawable area (accounting for margins and node size)
	drawWidth := float64(opts.Width - 2*opts.Margin - 2*opts.NodeRadius)
	drawHeight := float64(opts.Height - 2*opts.Margin - 2*opts.NodeRadius - 100) // Extra space for header/legend

	// Use a simple circular layout for now (can be enhanced with force-directed later)
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	// Center point
	centerX := float64(opts.Width) / 2
	centerY := float64(opts.Height-100) / 2 // Account for header space

	// Calculate radius based on number of rooms
	radius := math.Min(drawWidth, drawHeight) / 2.5

	// Position rooms in a circle
	angleStep := 2 * math.Pi / float64(len(roomIDs))
	for i, id := range roomIDs {
		angle := float64(i) * angleStep
		positions[id] = position{
			X: centerX + radius*math.Cos(angle),
			Y: centerY + radius*math.Sin(angle),
		}
	}

	return positions
}

// drawEdges renders all connectors as lines between rooms.
func drawEdges(canvas *svg.SVG, g *graph.Graph, positions map[string]position, opts SVGOptions) {
	// Sort connector IDs for deterministic output
	connectorIDs := make([]string, 0, len(g.Connectors))
	for id := range g.Connectors {
		connectorIDs = append(connectorIDs, id)
	}
	sort.Strings(connectorIDs)

	for _, connID := range connectorIDs {
		conn := g.Connectors[connID]
		fromPos, fromOK := positions[conn.From]
		toPos, toOK := positions[conn.To]

		if !fromOK || !toOK {
			continue // Skip if positions not found
		}

		// Determine edge color and style based on connector type
		color, style := getEdgeStyle(conn, opts)

		// Draw the line
		canvas.Line(
			int(fromPos.X), int(fromPos.Y),
			int(toPos.X), int(toPos.Y),
			fmt.Sprintf("stroke:%s;stroke-width:%d;%s", color, opts.EdgeWidth, style),
		)

		// Draw arrow if one-way
		if !conn.Bidirectional {
			drawArrow(canvas, fromPos, toPos, color, opts)
		}

		// Draw gate indicator if present
		if conn.Gate != nil {
			midX := (fromPos.X + toPos.X) / 2
			midY := (fromPos.Y + toPos.Y) / 2
			drawGate(canvas, midX, midY, conn.Gate, opts)
		}
	}
}

// getEdgeStyle returns color and SVG style string for a connector.
func getEdgeStyle(conn *graph.Connector, opts SVGOptions) (string, string) {
	baseColor := "#4a5568" // Default gray
	style := "opacity:0.8"

	if opts.ColorByType {
		switch conn.Type {
		case graph.TypeDoor:
			baseColor = "#48bb78" // Green
		case graph.TypeCorridor:
			baseColor = "#4299e1" // Blue
		case graph.TypeLadder:
			baseColor = "#ed8936" // Orange
		case graph.TypeTeleporter:
			baseColor = "#9f7aea" // Purple
		case graph.TypeHidden:
			baseColor = "#718096" // Gray
			style = "opacity:0.4;stroke-dasharray:5,5"
		case graph.TypeOneWay:
			baseColor = "#f56565" // Red
		}

		// Modify for secret visibility
		if conn.Visibility == graph.VisibilitySecret {
			style += ";stroke-dasharray:3,3"
		}
	}

	return baseColor, style
}

// drawArrow draws a small arrow at the midpoint of an edge.
func drawArrow(canvas *svg.SVG, from, to position, color string, opts SVGOptions) {
	// Calculate midpoint and angle
	midX := (from.X + to.X) / 2
	midY := (from.Y + to.Y) / 2

	dx := to.X - from.X
	dy := to.Y - from.Y
	angle := math.Atan2(dy, dx)

	// Arrow size
	arrowSize := 8.0

	// Calculate arrow points
	tip := position{
		X: midX + arrowSize*math.Cos(angle),
		Y: midY + arrowSize*math.Sin(angle),
	}
	left := position{
		X: midX + arrowSize*math.Cos(angle+2.8),
		Y: midY + arrowSize*math.Sin(angle+2.8),
	}
	right := position{
		X: midX + arrowSize*math.Cos(angle-2.8),
		Y: midY + arrowSize*math.Sin(angle-2.8),
	}

	// Draw arrow as polygon
	xs := []int{int(tip.X), int(left.X), int(right.X)}
	ys := []int{int(tip.Y), int(left.Y), int(right.Y)}
	canvas.Polygon(xs, ys, fmt.Sprintf("fill:%s", color))
}

// drawGate draws a small icon indicating a gated connection.
func drawGate(canvas *svg.SVG, x, y float64, gate *graph.Gate, opts SVGOptions) {
	// Draw a small circle for gate indicator
	canvas.Circle(int(x), int(y), 6, "fill:#ffd700;stroke:#000;stroke-width:1")

	// Add first letter of gate type
	glyph := "?"
	if len(gate.Type) > 0 {
		glyph = string(gate.Type[0])
	}
	canvas.Text(int(x), int(y+4), glyph,
		"text-anchor:middle;font-size:10px;font-weight:bold;fill:#000")
}

// drawNodes renders all rooms as colored circles.
func drawNodes(canvas *svg.SVG, g *graph.Graph, positions map[string]position, opts SVGOptions) {
	// Sort room IDs for deterministic output
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	for _, id := range roomIDs {
		room := g.Rooms[id]
		pos, ok := positions[id]
		if !ok {
			continue
		}

		// Get node color based on archetype
		color := getNodeColor(room.Archetype, opts)

		// Adjust size based on room size
		radius := getNodeRadius(room.Size, opts.NodeRadius)

		// Draw circle with stroke
		canvas.Circle(
			int(pos.X), int(pos.Y), radius,
			fmt.Sprintf("fill:%s;stroke:#fff;stroke-width:2;opacity:0.9", color),
		)

		// Draw inner circle for difficulty indication if heatmap not shown
		if !opts.ShowHeatmap {
			innerRadius := int(float64(radius) * 0.6)
			difficultyAlpha := 0.3 + (room.Difficulty * 0.7)
			canvas.Circle(
				int(pos.X), int(pos.Y), innerRadius,
				fmt.Sprintf("fill:#ff6b6b;opacity:%.2f", difficultyAlpha),
			)
		}
	}
}

// getNodeColor returns the color for a room based on its archetype.
func getNodeColor(archetype graph.RoomArchetype, opts SVGOptions) string {
	if !opts.ColorByType {
		return "#4a5568" // Default gray
	}

	switch archetype {
	case graph.ArchetypeStart:
		return "#48bb78" // Green
	case graph.ArchetypeBoss:
		return "#f56565" // Red
	case graph.ArchetypeTreasure:
		return "#ffd700" // Gold
	case graph.ArchetypePuzzle:
		return "#9f7aea" // Purple
	case graph.ArchetypeHub:
		return "#4299e1" // Blue
	case graph.ArchetypeCorridor:
		return "#718096" // Gray
	case graph.ArchetypeSecret:
		return "#805ad5" // Dark purple
	case graph.ArchetypeOptional:
		return "#38b2ac" // Teal
	case graph.ArchetypeVendor:
		return "#ed8936" // Orange
	case graph.ArchetypeShrine:
		return "#ecc94b" // Yellow
	case graph.ArchetypeCheckpoint:
		return "#4299e1" // Light blue
	default:
		return "#4a5568" // Gray
	}
}

// getNodeRadius returns the radius for a room based on its size.
func getNodeRadius(size graph.RoomSize, baseRadius int) int {
	multiplier := 1.0
	switch size {
	case graph.SizeXS:
		multiplier = 0.6
	case graph.SizeS:
		multiplier = 0.8
	case graph.SizeM:
		multiplier = 1.0
	case graph.SizeL:
		multiplier = 1.3
	case graph.SizeXL:
		multiplier = 1.6
	}
	return int(float64(baseRadius) * multiplier)
}

// drawLabels renders room ID labels near each node.
func drawLabels(canvas *svg.SVG, g *graph.Graph, positions map[string]position, opts SVGOptions) {
	// Sort room IDs for deterministic output
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	for _, id := range roomIDs {
		room := g.Rooms[id]
		pos, ok := positions[id]
		if !ok {
			continue
		}

		// Calculate label position (below the node)
		radius := getNodeRadius(room.Size, opts.NodeRadius)
		labelY := int(pos.Y) + radius + 15

		// Draw label with background for readability
		canvas.Text(
			int(pos.X), labelY, id,
			"text-anchor:middle;font-size:11px;font-family:monospace;fill:#e2e8f0;font-weight:500",
		)
	}
}

// drawHeatmap renders a difficulty heatmap overlay on nodes.
func drawHeatmap(canvas *svg.SVG, g *graph.Graph, positions map[string]position, opts SVGOptions) {
	// Sort room IDs for deterministic output
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	for _, id := range roomIDs {
		room := g.Rooms[id]
		pos, ok := positions[id]
		if !ok {
			continue
		}

		radius := getNodeRadius(room.Size, opts.NodeRadius)

		// Use a red-to-yellow gradient for difficulty
		// Low difficulty = cooler colors, high difficulty = hotter colors
		heatColor := getHeatmapColor(room.Difficulty)
		alpha := 0.4 + (room.Difficulty * 0.4)

		// Draw semi-transparent circle
		canvas.Circle(
			int(pos.X), int(pos.Y), radius+5,
			fmt.Sprintf("fill:%s;opacity:%.2f;stroke:none", heatColor, alpha),
		)
	}
}

// getHeatmapColor returns a color from cool to hot based on difficulty.
func getHeatmapColor(difficulty float64) string {
	// Interpolate from blue (low) through green/yellow to red (high)
	if difficulty < 0.25 {
		return "#3b82f6" // Blue
	} else if difficulty < 0.5 {
		return "#10b981" // Green
	} else if difficulty < 0.75 {
		return "#f59e0b" // Yellow
	} else {
		return "#ef4444" // Red
	}
}

// drawLegend renders a legend explaining the color coding.
func drawLegend(canvas *svg.SVG, opts SVGOptions) {
	legendX := opts.Width - opts.Margin - 180
	legendY := opts.Margin + 20

	// Legend background
	canvas.Rect(legendX-10, legendY-15, 190, 320,
		"fill:#2d3748;stroke:#4a5568;stroke-width:1;opacity:0.95;rx:5")

	// Legend title
	canvas.Text(legendX, legendY, "Room Types",
		"font-size:14px;font-weight:bold;fill:#e2e8f0")

	legendY += 25

	// Room type legend entries
	legendEntries := []struct {
		name  string
		color string
	}{
		{"Start", getNodeColor(graph.ArchetypeStart, opts)},
		{"Boss", getNodeColor(graph.ArchetypeBoss, opts)},
		{"Treasure", getNodeColor(graph.ArchetypeTreasure, opts)},
		{"Puzzle", getNodeColor(graph.ArchetypePuzzle, opts)},
		{"Hub", getNodeColor(graph.ArchetypeHub, opts)},
		{"Corridor", getNodeColor(graph.ArchetypeCorridor, opts)},
		{"Secret", getNodeColor(graph.ArchetypeSecret, opts)},
		{"Optional", getNodeColor(graph.ArchetypeOptional, opts)},
		{"Vendor", getNodeColor(graph.ArchetypeVendor, opts)},
		{"Shrine", getNodeColor(graph.ArchetypeShrine, opts)},
		{"Checkpoint", getNodeColor(graph.ArchetypeCheckpoint, opts)},
	}

	for _, entry := range legendEntries {
		canvas.Circle(legendX+8, legendY, 8, fmt.Sprintf("fill:%s;stroke:#fff;stroke-width:1", entry.color))
		canvas.Text(legendX+25, legendY+4, entry.name, "font-size:11px;fill:#cbd5e0")
		legendY += 22
	}

	// Connector type legend
	legendY += 15
	canvas.Text(legendX, legendY, "Connections",
		"font-size:14px;font-weight:bold;fill:#e2e8f0")
	legendY += 20

	canvas.Line(legendX, legendY, legendX+30, legendY, "stroke:#48bb78;stroke-width:2")
	canvas.Text(legendX+35, legendY+4, "Door", "font-size:11px;fill:#cbd5e0")
	legendY += 18

	canvas.Line(legendX, legendY, legendX+30, legendY, "stroke:#4299e1;stroke-width:2")
	canvas.Text(legendX+35, legendY+4, "Corridor", "font-size:11px;fill:#cbd5e0")
	legendY += 18

	canvas.Line(legendX, legendY, legendX+30, legendY, "stroke:#718096;stroke-width:2;stroke-dasharray:5,5")
	canvas.Text(legendX+35, legendY+4, "Hidden", "font-size:11px;fill:#cbd5e0")
	legendY += 18

	canvas.Line(legendX, legendY, legendX+30, legendY, "stroke:#f56565;stroke-width:2")
	canvas.Text(legendX+35, legendY+4, "One-Way", "font-size:11px;fill:#cbd5e0")
}

// drawHeader renders title and statistics at the top of the visualization.
func drawHeader(canvas *svg.SVG, artifact *dungeon.Artifact, opts SVGOptions) {
	headerY := 25

	// Draw title
	if opts.Title != "" {
		canvas.Text(opts.Width/2, headerY, opts.Title,
			"text-anchor:middle;font-size:20px;font-weight:bold;fill:#e2e8f0;font-family:sans-serif")
		headerY += 30
	}

	// Draw statistics
	if opts.ShowStats && artifact.ADG != nil && artifact.ADG.Graph != nil {
		g := artifact.ADG.Graph
		stats := fmt.Sprintf("Rooms: %d | Connectors: %d | Seed: %d",
			len(g.Rooms), len(g.Connectors), g.Seed)

		canvas.Text(opts.Width/2, headerY, stats,
			"text-anchor:middle;font-size:12px;fill:#a0aec0;font-family:monospace")

		// Add metrics if available
		if artifact.Metrics != nil {
			headerY += 20
			metricsStr := fmt.Sprintf("Branch Factor: %.2f | Path Length: %d | Cycles: %d | Pacing Δ: %.3f",
				artifact.Metrics.BranchingFactor,
				artifact.Metrics.PathLength,
				artifact.Metrics.CycleCount,
				artifact.Metrics.PacingDeviation)

			canvas.Text(opts.Width/2, headerY, metricsStr,
				"text-anchor:middle;font-size:11px;fill:#718096;font-family:monospace")
		}
	}
}
