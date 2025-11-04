package validation

import (
	"fmt"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// TestCheckPathBounds_HubAndSpokeArchitecture verifies that validation accepts
// branching (hub-and-spoke) dungeon architectures where the critical path is short
// but most content is in optional branches.
func TestCheckPathBounds_HubAndSpokeArchitecture(t *testing.T) {
	tests := []struct {
		name        string
		totalRooms  int
		pathRooms   int // Number of rooms on critical path (nodes, not edges)
		minRooms    int
		maxRooms    int
		shouldPass  bool
		description string
	}{
		{
			name:        "tiny_hub_spoke_3_rooms",
			totalRooms:  15,
			pathRooms:   3, // Start → Hub → Boss
			minRooms:    10,
			maxRooms:    20,
			shouldPass:  true,
			description: "15-room dungeon with 3-room path (Start→Hub→Boss) should pass",
		},
		{
			name:        "small_hub_spoke_3_rooms",
			totalRooms:  25,
			pathRooms:   3, // Start → Hub → Boss
			minRooms:    25,
			maxRooms:    35,
			shouldPass:  true,
			description: "25-room dungeon with 3-room path should pass (min=2)",
		},
		{
			name:        "medium_hub_spoke_4_rooms",
			totalRooms:  30,
			pathRooms:   4, // Start → Hub1 → Hub2 → Boss
			minRooms:    30,
			maxRooms:    40,
			shouldPass:  true,
			description: "30-room dungeon with 4-room path should pass (min=3)",
		},
		{
			name:        "large_hub_spoke_5_rooms",
			totalRooms:  50,
			pathRooms:   5, // Start → Hub1 → Hub2 → Hub3 → Boss
			minRooms:    50,
			maxRooms:    60,
			shouldPass:  true,
			description: "50-room dungeon with 5-room path should pass (min=5)",
		},
		{
			name:        "degenerate_direct_path",
			totalRooms:  25,
			pathRooms:   2, // Start → Boss (no intermediate rooms)
			minRooms:    25,
			maxRooms:    35,
			shouldPass:  true,
			description: "2-room path should pass (meets minimum of 2)",
		},
		{
			name:        "linear_dungeon_many_rooms",
			totalRooms:  30,
			pathRooms:   30, // All rooms on critical path
			minRooms:    30,
			maxRooms:    40,
			shouldPass:  true,
			description: "Linear dungeon with all rooms on path should also pass",
		},
		{
			name:        "impossible_single_room",
			totalRooms:  1,
			pathRooms:   1, // Only Start room
			minRooms:    10,
			maxRooms:    20,
			shouldPass:  false,
			description: "Single room path should fail (needs at least 2: Start and Boss)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test graph with hub-and-spoke structure
			g := createHubSpokeGraph(tt.totalRooms, tt.pathRooms)

			cfg := &dungeon.Config{
				Size: dungeon.SizeCfg{
					RoomsMin: tt.minRooms,
					RoomsMax: tt.maxRooms,
				},
			}

			result := CheckPathBounds(g, cfg)

			if result.Satisfied != tt.shouldPass {
				t.Errorf("%s: %s\n  Expected satisfied=%v, got %v\n  Details: %s",
					tt.name,
					tt.description,
					tt.shouldPass,
					result.Satisfied,
					result.Details)
			}

			// Log bounds calculation
			expectedMin := max(2, tt.minRooms/10)
			t.Logf("Bounds calculation: min=%d (max(2, %d/10))", expectedMin, tt.minRooms)

			t.Logf("✓ %s: path=%d rooms, satisfied=%v",
				tt.name,
				tt.pathRooms,
				result.Satisfied)
		})
	}
}

// createHubSpokeGraph creates a test graph with hub-and-spoke architecture.
// pathRooms: number of rooms on the critical path (Start → Hubs → Boss)
// totalRooms: total rooms including optional branches
func createHubSpokeGraph(totalRooms, pathRooms int) *graph.Graph {
	g := &graph.Graph{
		Rooms:      make(map[string]*graph.Room),
		Adjacency:  make(map[string][]string),
		Connectors: make(map[string]*graph.Connector),
	}

	if totalRooms < 2 || pathRooms < 2 {
		// Degenerate case: just Start room
		g.Rooms["start"] = &graph.Room{
			ID:        "start",
			Archetype: graph.ArchetypeStart,
			Size:      graph.SizeM,
		}
		g.Adjacency["start"] = []string{}
		return g
	}

	// Create critical path: Start → Hub(s) → Boss
	pathIDs := make([]string, pathRooms)
	pathIDs[0] = "start"
	pathIDs[pathRooms-1] = "boss"

	// Create intermediate hubs
	for i := 1; i < pathRooms-1; i++ {
		pathIDs[i] = fmt.Sprintf("hub_%d", i)
	}

	// Add rooms for critical path
	for i, id := range pathIDs {
		archetype := graph.ArchetypeHub
		if i == 0 {
			archetype = graph.ArchetypeStart
		} else if i == pathRooms-1 {
			archetype = graph.ArchetypeBoss
		}

		g.Rooms[id] = &graph.Room{
			ID:        id,
			Archetype: archetype,
			Size:      graph.SizeM,
		}
		g.Adjacency[id] = []string{}
	}

	// Connect critical path
	for i := 0; i < len(pathIDs)-1; i++ {
		from := pathIDs[i]
		to := pathIDs[i+1]
		connID := fmt.Sprintf("conn_%s_%s", from, to)

		g.Connectors[connID] = &graph.Connector{
			ID:            connID,
			From:          from,
			To:            to,
			Type:          graph.TypeCorridor,
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		}

		g.Adjacency[from] = append(g.Adjacency[from], to)
		g.Adjacency[to] = append(g.Adjacency[to], from)
	}

	// Add optional branches (spokes) to reach totalRooms
	roomsToAdd := totalRooms - pathRooms
	for i := 0; i < roomsToAdd; i++ {
		// Attach to a hub (prefer middle hubs)
		// For 2-room path (Start→Boss), attach to Start
		// For 3+ room path, distribute across middle hubs
		hubIdx := 0 // Default to Start
		if pathRooms > 2 {
			hubIdx = 1 + (i % (pathRooms - 2))
			if hubIdx >= len(pathIDs)-1 {
				hubIdx = 1
			}
		}
		hubID := pathIDs[hubIdx]

		// Create optional room
		optionalID := fmt.Sprintf("optional_%d", i)
		g.Rooms[optionalID] = &graph.Room{
			ID:        optionalID,
			Archetype: graph.ArchetypeOptional,
			Size:      graph.SizeS,
		}
		g.Adjacency[optionalID] = []string{}

		// Connect to hub
		connID := fmt.Sprintf("conn_%s_%s", hubID, optionalID)
		g.Connectors[connID] = &graph.Connector{
			ID:            connID,
			From:          hubID,
			To:            optionalID,
			Type:          graph.TypeDoor,
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		}

		g.Adjacency[hubID] = append(g.Adjacency[hubID], optionalID)
		g.Adjacency[optionalID] = append(g.Adjacency[optionalID], hubID)
	}

	return g
}

// TestCheckPathBounds_DocumentedExamples provides concrete examples for documentation.
func TestCheckPathBounds_DocumentedExamples(t *testing.T) {
	t.Log("Hub-and-Spoke Architecture Examples")
	t.Log("=====================================")
	t.Log("")

	examples := []struct {
		totalRooms int
		pathRooms  int
		minRooms   int
		maxRooms   int
	}{
		{totalRooms: 25, pathRooms: 3, minRooms: 25, maxRooms: 35}, // Small
		{totalRooms: 30, pathRooms: 3, minRooms: 30, maxRooms: 40}, // Medium
		{totalRooms: 50, pathRooms: 5, minRooms: 50, maxRooms: 60}, // Large
	}

	for _, ex := range examples {
		g := createHubSpokeGraph(ex.totalRooms, ex.pathRooms)
		cfg := &dungeon.Config{
			Size: dungeon.SizeCfg{
				RoomsMin: ex.minRooms,
				RoomsMax: ex.maxRooms,
			},
		}

		result := CheckPathBounds(g, cfg)
		pathEdges := CalculatePathLength(g)

		t.Logf("Dungeon: %d total rooms, %d rooms on critical path (%d edges)",
			ex.totalRooms, ex.pathRooms, pathEdges)
		t.Logf("  Validation: %s", result.Details)
		t.Logf("  Optional branches: %d rooms (%.0f%% of total)",
			ex.totalRooms-ex.pathRooms,
			float64(ex.totalRooms-ex.pathRooms)/float64(ex.totalRooms)*100)
		t.Logf("")
	}
}
