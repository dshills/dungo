package dungeon

import (
	"fmt"
	"strings"

	"github.com/dshills/dungo/pkg/graph"
)

// RenderText creates a basic text representation of the dungeon for debugging.
// Returns a multi-line string showing the dungeon structure.
func (a *Artifact) RenderText() string {
	if a == nil || a.ADG == nil {
		return "No dungeon data available"
	}

	var sb strings.Builder

	// Header
	sb.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║              DUNGEON GENERATOR - TEXT VIEW                 ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")

	// Dungeon Statistics
	sb.WriteString("📊 STATISTICS:\n")
	sb.WriteString(fmt.Sprintf("   Rooms: %d\n", len(a.ADG.Rooms)))
	sb.WriteString(fmt.Sprintf("   Connections: %d\n", len(a.ADG.Connectors)))
	if a.Metrics != nil {
		sb.WriteString(fmt.Sprintf("   Branching Factor: %.2f\n", a.Metrics.BranchingFactor))
		sb.WriteString(fmt.Sprintf("   Path to Boss: %d rooms\n", a.Metrics.PathLength))
		sb.WriteString(fmt.Sprintf("   Cycles: %d\n", a.Metrics.CycleCount))
	}
	sb.WriteString("\n")

	// Room List
	sb.WriteString("🏰 ROOMS:\n")
	for id, room := range a.ADG.Rooms {
		symbol := getRoomSymbol(room.Archetype)
		sb.WriteString(fmt.Sprintf("   %s %s [%s] (Difficulty: %.1f, Reward: %.1f)\n",
			symbol, id, room.Archetype, room.Difficulty, room.Reward))

		// Show what this room provides
		if len(room.Provides) > 0 {
			sb.WriteString("      Provides: ")
			for i, cap := range room.Provides {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s:%s", cap.Type, cap.Value))
			}
			sb.WriteString("\n")
		}

		// Show requirements
		if len(room.Requirements) > 0 {
			sb.WriteString("      Requires: ")
			for i, req := range room.Requirements {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s:%s", req.Type, req.Value))
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Connections
	sb.WriteString("🚪 CONNECTIONS:\n")
	for id, conn := range a.ADG.Connectors {
		arrow := "↔"
		if !conn.Bidirectional {
			arrow = "→"
		}
		sb.WriteString(fmt.Sprintf("   %s: %s %s %s [%s]",
			id, conn.From, arrow, conn.To, conn.Type))

		// Show if gated
		if conn.Gate != nil {
			sb.WriteString(fmt.Sprintf(" 🔒 Requires %s:%s", conn.Gate.Type, conn.Gate.Value))
		}

		// Show if secret
		if conn.Visibility == graph.VisibilitySecret {
			sb.WriteString(" 🔍 Secret")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Tile Map Preview (if available)
	if a.TileMap != nil {
		sb.WriteString(a.renderTileMapPreview())
	}

	// Content Summary (if available)
	if a.Content != nil {
		sb.WriteString("⚔️  CONTENT:\n")
		sb.WriteString(fmt.Sprintf("   Enemies: %d spawns\n", len(a.Content.Spawns)))
		sb.WriteString(fmt.Sprintf("   Loot: %d items\n", len(a.Content.Loot)))
		sb.WriteString(fmt.Sprintf("   Puzzles: %d\n", len(a.Content.Puzzles)))
		sb.WriteString(fmt.Sprintf("   Secrets: %d\n", len(a.Content.Secrets)))

		// Show required items (keys)
		requiredItems := 0
		for _, loot := range a.Content.Loot {
			if loot.Required {
				requiredItems++
			}
		}
		if requiredItems > 0 {
			sb.WriteString(fmt.Sprintf("   Required Items (Keys): %d\n", requiredItems))
		}
		sb.WriteString("\n")
	}

	// Validation Status
	if a.Debug != nil && a.Debug.Report != nil {
		sb.WriteString("✓ VALIDATION:\n")
		if a.Debug.Report.Passed {
			sb.WriteString("   Status: ✅ PASSED - All hard constraints satisfied\n")
		} else {
			sb.WriteString("   Status: ❌ FAILED - Constraint violations detected\n")
			for _, err := range a.Debug.Report.Errors {
				sb.WriteString(fmt.Sprintf("      - %s\n", err))
			}
		}
		if len(a.Debug.Report.Warnings) > 0 {
			sb.WriteString("   Warnings:\n")
			for _, warn := range a.Debug.Report.Warnings {
				sb.WriteString(fmt.Sprintf("      - %s\n", warn))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderTileMapPreview creates a small ASCII art preview of the tile map.
func (a *Artifact) renderTileMapPreview() string {
	var sb strings.Builder

	sb.WriteString("🗺️  TILE MAP PREVIEW:\n")
	sb.WriteString(fmt.Sprintf("   Dimensions: %dx%d tiles (%dx%d pixels)\n",
		a.TileMap.Width, a.TileMap.Height,
		a.TileMap.Width*a.TileMap.TileWidth, a.TileMap.Height*a.TileMap.TileHeight))

	// Get floor layer
	floorLayer, hasFloor := a.TileMap.Layers["floor"]
	wallLayer, hasWall := a.TileMap.Layers["walls"]

	if !hasFloor && !hasWall {
		sb.WriteString("   (No tile data available)\n\n")
		return sb.String()
	}

	// Create small preview (limit to 60x30 for readability)
	maxPreviewWidth := 60
	maxPreviewHeight := 30
	scaleX := 1
	scaleY := 1

	if a.TileMap.Width > maxPreviewWidth {
		scaleX = a.TileMap.Width / maxPreviewWidth
		if scaleX < 1 {
			scaleX = 1
		}
	}
	if a.TileMap.Height > maxPreviewHeight {
		scaleY = a.TileMap.Height / maxPreviewHeight
		if scaleY < 1 {
			scaleY = 1
		}
	}

	previewWidth := a.TileMap.Width / scaleX
	previewHeight := a.TileMap.Height / scaleY

	if scaleX > 1 || scaleY > 1 {
		sb.WriteString(fmt.Sprintf("   (Scaled %dx for display)\n", max(scaleX, scaleY)))
	}
	sb.WriteString("\n")

	// Render preview
	for y := 0; y < previewHeight && y*scaleY < a.TileMap.Height; y++ {
		sb.WriteString("   ")
		for x := 0; x < previewWidth && x*scaleX < a.TileMap.Width; x++ {
			// Sample tile at scaled position
			tileX := x * scaleX
			tileY := y * scaleY
			idx := tileY*a.TileMap.Width + tileX

			char := ' '
			if hasWall && wallLayer.Type == "tilelayer" && int(idx) < len(wallLayer.Data) && wallLayer.Data[idx] != 0 {
				char = '█' // Wall
			} else if hasFloor && floorLayer.Type == "tilelayer" && int(idx) < len(floorLayer.Data) && floorLayer.Data[idx] != 0 {
				char = '·' // Floor
			}

			// Check for doors
			if doorLayer, hasDoor := a.TileMap.Layers["doors"]; hasDoor {
				if doorLayer.Type == "tilelayer" && int(idx) < len(doorLayer.Data) && doorLayer.Data[idx] != 0 {
					char = '+' // Door
				}
			}

			sb.WriteRune(char)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n   Legend: █=Wall  ·=Floor  +=Door\n\n")
	return sb.String()
}

// getRoomSymbol returns an emoji symbol for the room archetype.
func getRoomSymbol(archetype graph.RoomArchetype) string {
	switch archetype {
	case graph.ArchetypeStart:
		return "🟢"
	case graph.ArchetypeBoss:
		return "👑"
	case graph.ArchetypeTreasure:
		return "💰"
	case graph.ArchetypePuzzle:
		return "🧩"
	case graph.ArchetypeHub:
		return "🎯"
	case graph.ArchetypeCorridor:
		return "🚶"
	case graph.ArchetypeSecret:
		return "🔍"
	case graph.ArchetypeOptional:
		return "⭐"
	case graph.ArchetypeVendor:
		return "🛒"
	case graph.ArchetypeShrine:
		return "⛩️"
	case graph.ArchetypeCheckpoint:
		return "💾"
	default:
		return "❓"
	}
}

// RenderTextSimple creates a simplified text representation showing just the graph structure.
func (a *Artifact) RenderTextSimple() string {
	if a == nil || a.ADG == nil {
		return "No dungeon data"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Dungeon: %d rooms, %d connections\n",
		len(a.ADG.Rooms), len(a.ADG.Connectors)))

	// Find start room
	var startID string
	for id, room := range a.ADG.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startID = id
			break
		}
	}

	if startID == "" {
		sb.WriteString("No Start room found\n")
		return sb.String()
	}

	// Simple traversal from start
	sb.WriteString(fmt.Sprintf("\nPath from Start (%s):\n", startID))
	visited := make(map[string]bool)
	a.renderPath(&sb, startID, 0, visited)

	return sb.String()
}

// renderPath recursively renders the dungeon graph as a tree.
func (a *Artifact) renderPath(sb *strings.Builder, roomID string, depth int, visited map[string]bool) {
	if visited[roomID] {
		return
	}
	visited[roomID] = true

	room := a.ADG.Rooms[roomID]
	indent := strings.Repeat("  ", depth)
	symbol := getRoomSymbol(room.Archetype)

	sb.WriteString(fmt.Sprintf("%s%s %s [%s]\n", indent, symbol, roomID, room.Archetype))

	// Show adjacent rooms
	if adj, ok := a.ADG.Adjacency[roomID]; ok {
		for _, nextID := range adj {
			if !visited[nextID] {
				a.renderPath(sb, nextID, depth+1, visited)
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
