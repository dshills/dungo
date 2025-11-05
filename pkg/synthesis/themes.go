package synthesis

import (
	"fmt"
	"sort"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// assignThemes assigns theme tags to all rooms in the graph based on the theme configuration.
//
// Theme Assignment Strategies:
// - Single theme: All rooms get the same "biome:themename" tag
// - Multi-theme: Creates clusters using graph regions with smooth transitions
//
// For multi-theme dungeons, the algorithm:
// 1. Selects theme seed rooms (distributed across the graph)
// 2. Grows regions from seed rooms using breadth-first expansion
// 3. Prefers assigning same theme to connected rooms for smooth transitions
// 4. Tags each room with "biome:themename"
func assignThemes(g *graph.Graph, themes []string, rng *rng.RNG) error {
	if len(themes) == 0 {
		return fmt.Errorf("at least one theme must be specified")
	}

	if len(g.Rooms) == 0 {
		return nil // No rooms to assign themes to
	}

	// Single theme: assign to all rooms
	if len(themes) == 1 {
		return assignSingleTheme(g, themes[0])
	}

	// Multi-theme: create clusters with smooth transitions
	return assignMultiTheme(g, themes, rng)
}

// assignSingleTheme assigns the same theme to all rooms.
func assignSingleTheme(g *graph.Graph, theme string) error {
	for _, room := range g.Rooms {
		if room.Tags == nil {
			room.Tags = make(map[string]string)
		}
		room.Tags["biome"] = theme
	}
	return nil
}

// assignMultiTheme creates theme clusters across the graph with smooth transitions.
// Uses region growing from seed rooms to create cohesive themed areas.
func assignMultiTheme(g *graph.Graph, themes []string, rng *rng.RNG) error {
	// Track which rooms have been assigned a theme
	assigned := make(map[string]bool)
	themeAssignments := make(map[string]string) // roomID -> theme

	// Step 1: Select seed rooms for each theme
	// Distribute seeds roughly evenly across the graph
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}

	// Sort room IDs for deterministic order before shuffling
	sort.Strings(roomIDs)

	// Shuffle room IDs for random seed selection
	rng.Shuffle(len(roomIDs), func(i, j int) {
		roomIDs[i], roomIDs[j] = roomIDs[j], roomIDs[i]
	})

	// Select one seed room per theme
	seeds := make(map[string]string) // theme -> roomID
	for i, theme := range themes {
		if i < len(roomIDs) {
			seeds[theme] = roomIDs[i]
			assigned[roomIDs[i]] = true
			themeAssignments[roomIDs[i]] = theme
		}
	}

	// Step 2: Grow regions from seed rooms using breadth-first expansion
	// Process themes in shuffled order to avoid bias
	themeOrder := make([]string, len(themes))
	copy(themeOrder, themes)
	rng.Shuffle(len(themeOrder), func(i, j int) {
		themeOrder[i], themeOrder[j] = themeOrder[j], themeOrder[i]
	})

	// Expand regions until all rooms are assigned
	maxIterations := len(g.Rooms) * 2 // Safety limit
	iteration := 0

	for len(assigned) < len(g.Rooms) && iteration < maxIterations {
		iteration++

		// Try to expand each theme's region
		for _, theme := range themeOrder {
			// Find frontier rooms for this theme (neighbors of assigned rooms)
			frontier := findFrontier(g, theme, themeAssignments, assigned)

			if len(frontier) == 0 {
				continue
			}

			// Randomly select a frontier room to add to this theme
			frontierRoom := frontier[rng.Intn(len(frontier))]
			assigned[frontierRoom] = true
			themeAssignments[frontierRoom] = theme
		}
	}

	// Step 3: Assign any remaining unassigned rooms (shouldn't happen, but safety check)
	// Use sorted iteration for deterministic RNG consumption
	unassignedIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		unassignedIDs = append(unassignedIDs, id)
	}
	sort.Strings(unassignedIDs)

	for _, roomID := range unassignedIDs {
		if !assigned[roomID] {
			// Assign to the theme of a neighboring room, or random theme
			neighborTheme := findNeighborTheme(g, roomID, themeAssignments)
			if neighborTheme == "" {
				neighborTheme = themes[rng.Intn(len(themes))]
			}
			themeAssignments[roomID] = neighborTheme
		}
	}

	// Step 4: Apply theme assignments to room tags
	for roomID, theme := range themeAssignments {
		room := g.Rooms[roomID]
		if room.Tags == nil {
			room.Tags = make(map[string]string)
		}
		room.Tags["biome"] = theme
	}

	return nil
}

// findFrontier returns rooms adjacent to the current theme's territory but not yet assigned.
// These are candidates for expanding the theme region.
func findFrontier(g *graph.Graph, theme string, assignments map[string]string, assigned map[string]bool) []string {
	frontier := make([]string, 0)
	visited := make(map[string]bool)

	// Find all rooms currently assigned to this theme
	for roomID, roomTheme := range assignments {
		if roomTheme != theme {
			continue
		}

		// Check neighbors of this room
		for _, neighborID := range g.Adjacency[roomID] {
			// Skip if already assigned or already in frontier
			if assigned[neighborID] || visited[neighborID] {
				continue
			}

			frontier = append(frontier, neighborID)
			visited[neighborID] = true
		}
	}

	return frontier
}

// findNeighborTheme returns the theme of a random neighbor, or empty string if no neighbors have themes.
// Used as a fallback to ensure smooth transitions.
func findNeighborTheme(g *graph.Graph, roomID string, assignments map[string]string) string {
	neighbors := g.Adjacency[roomID]
	if len(neighbors) == 0 {
		return ""
	}

	// Return the theme of the first neighbor that has one
	for _, neighborID := range neighbors {
		if theme, ok := assignments[neighborID]; ok {
			return theme
		}
	}

	return ""
}
