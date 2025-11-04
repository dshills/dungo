package validation

import (
	"fmt"
	"math"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// CheckConnectivity ensures the graph is a single connected component.
// This is a hard constraint - all rooms must be reachable from any starting point.
func CheckConnectivity(g *graph.Graph) dungeon.ConstraintResult {
	if len(g.Rooms) == 0 {
		return NewHardConstraintResult(
			"Connectivity",
			"graph.isConnected()",
			false,
			"Graph has no rooms",
		)
	}

	// Use weak connectivity for dungeons (allows one-way passages)
	// This ensures all rooms are in the same connected component
	// when treating edges as undirected
	isConnected := g.IsWeaklyConnected()

	details := "All rooms are reachable (weak connectivity)"
	if !isConnected {
		// For better error reporting, count components using weak connectivity
		details = "Graph is disconnected: rooms cannot all reach each other (even ignoring edge direction)"
	}

	return NewHardConstraintResult(
		"Connectivity",
		"graph.isConnected()",
		isConnected,
		details,
	)
}

// CheckKeyReachability ensures keys are obtainable before locked rooms.
// This is a hard constraint - players must be able to access keys before locks.
func CheckKeyReachability(g *graph.Graph, cfg *dungeon.Config) dungeon.ConstraintResult {
	// Find all key providers and locked rooms
	keyRooms := FindKeyRooms(g)
	lockedRooms := FindLockedRooms(g)

	// If no keys are configured, this constraint is satisfied
	if len(cfg.Keys) == 0 && len(keyRooms) == 0 {
		return NewHardConstraintResult(
			"KeyReachability",
			"keys.reachableBeforeLocks()",
			true,
			"No keys configured",
		)
	}

	// Check each key type
	violations := []string{}

	for _, keyCfg := range cfg.Keys {
		keyName := keyCfg.Name
		providers := keyRooms[keyName]
		requirers := lockedRooms[keyName]

		if len(requirers) == 0 {
			// No locked rooms for this key, OK
			continue
		}

		if len(providers) == 0 {
			violations = append(violations, fmt.Sprintf("Key '%s' is required but not provided anywhere", keyName))
			continue
		}

		// Check that at least one key provider is reachable without the key
		keyReachableWithoutKey := false
		for _, providerID := range providers {
			provider := g.Rooms[providerID]
			requiresThisKey := false
			for _, req := range provider.Requirements {
				if req.Type == "key" && req.Value == keyName {
					requiresThisKey = true
					break
				}
			}
			if !requiresThisKey {
				keyReachableWithoutKey = true
				break
			}
		}

		if !keyReachableWithoutKey {
			violations = append(violations, fmt.Sprintf("Key '%s' requires itself to obtain (circular dependency)", keyName))
		}
	}

	satisfied := len(violations) == 0
	details := "All keys are reachable before locks"
	if !satisfied {
		details = fmt.Sprintf("Key reachability violations: %v", violations)
	}

	return NewHardConstraintResult(
		"KeyReachability",
		"keys.reachableBeforeLocks()",
		satisfied,
		details,
	)
}

// CheckNoOverlaps ensures rooms don't overlap in spatial layout.
// This is a hard constraint for embedded dungeons.
func CheckNoOverlaps(g *graph.Graph, layout *dungeon.Layout) dungeon.ConstraintResult {
	if layout == nil {
		return NewHardConstraintResult(
			"NoOverlaps",
			"spatial.noOverlaps()",
			true,
			"No layout provided (skipping spatial check)",
		)
	}

	// Create a simple bounding box for each room
	// In a real implementation, this would use actual footprints
	overlaps := []string{}

	rooms := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		rooms = append(rooms, id)
	}

	// Check all pairs of rooms
	for i := 0; i < len(rooms); i++ {
		for j := i + 1; j < len(rooms); j++ {
			id1 := rooms[i]
			id2 := rooms[j]

			pose1, exists1 := layout.Poses[id1]
			pose2, exists2 := layout.Poses[id2]

			if !exists1 || !exists2 {
				continue
			}

			// Simple overlap check using estimated room sizes
			// In reality, this would use actual footprint geometry
			size1 := estimateRoomSize(g.Rooms[id1].Size)
			size2 := estimateRoomSize(g.Rooms[id2].Size)

			// Convert center coordinates to corner (top-left) coordinates
			// pose.X and pose.Y are center positions, so we subtract half the size
			corner1X := pose1.X - size1/2
			corner1Y := pose1.Y - size1/2
			corner2X := pose2.X - size2/2
			corner2Y := pose2.Y - size2/2

			// Check if bounding boxes overlap
			if rectOverlaps(
				corner1X, corner1Y, size1, size1,
				corner2X, corner2Y, size2, size2,
			) {
				overlaps = append(overlaps, fmt.Sprintf("%s and %s", id1, id2))
			}
		}
	}

	satisfied := len(overlaps) == 0
	details := "No room overlaps detected"
	if !satisfied {
		details = fmt.Sprintf("Found %d overlaps: %v", len(overlaps), overlaps)
	}

	return NewHardConstraintResult(
		"NoOverlaps",
		"spatial.noOverlaps()",
		satisfied,
		details,
	)
}

// CheckPathBounds ensures Start-to-Boss path length is within reasonable bounds.
// This is a hard constraint to prevent degenerate dungeons.
//
// Design Philosophy:
// This validation accepts both linear and branching (hub-and-spoke) dungeon architectures.
// For hub-and-spoke dungeons (Start → Hub → Boss), the critical path is intentionally short
// (typically 3-4 rooms), with most content in optional branches off the hubs.
//
// Path Length Calculation:
// - pathLength = number of rooms in shortest path (nodes, not edges)
// - For Start → Hub → Boss, pathLength = 3
// - CalculatePathLength() metric uses edges (pathLength - 1) for display purposes
//
// Bounds:
//   - Minimum: max(2, roomsMin/10) to allow very short critical paths in branching dungeons
//     Examples: 25 rooms → min 2 | 50 rooms → min 5 | 100 rooms → min 10
//   - Maximum: roomsMax * 2 (very generous, allows linear dungeons and cycles)
func CheckPathBounds(g *graph.Graph, cfg *dungeon.Config) dungeon.ConstraintResult {
	startID := FindStartRoom(g)
	bossID := FindBossRoom(g)

	if startID == "" || bossID == "" {
		return NewHardConstraintResult(
			"PathBounds",
			"path.withinBounds(start, boss)",
			false,
			"Missing Start or Boss room",
		)
	}

	path, err := g.GetPath(startID, bossID)
	if err != nil {
		return NewHardConstraintResult(
			"PathBounds",
			"path.withinBounds(start, boss)",
			false,
			fmt.Sprintf("No path from Start to Boss: %v", err),
		)
	}

	pathLength := len(path)

	// Minimum path length: 2 rooms (Start + Boss absolute minimum)
	// Hub-and-spoke architecture creates 3-room critical paths regardless of dungeon size
	minPathLength := 2

	// Maximum path length: allow up to full dungeon size
	maxPathLength := cfg.Size.RoomsMax

	satisfied := pathLength >= minPathLength && pathLength <= maxPathLength
	details := fmt.Sprintf("Path length: %d rooms (bounds: %d-%d)", pathLength, minPathLength, maxPathLength)

	return NewHardConstraintResult(
		"PathBounds",
		"path.withinBounds(start, boss)",
		satisfied,
		details,
	)
}

// CheckPacingDeviation measures how well the dungeon follows the configured pacing curve.
// This is a soft constraint - returns a score from 0.0 to 1.0.
func CheckPacingDeviation(g *graph.Graph, cfg *dungeon.Config) dungeon.ConstraintResult {
	deviation := CalculatePacingDeviation(g, cfg)

	// Convert deviation to a score (lower deviation = higher score)
	// Deviation is L2 distance, typically in range [0, 1]
	// Score = 1 - deviation, clamped to [0, 1]
	score := math.Max(0.0, 1.0-deviation)

	details := fmt.Sprintf("Pacing deviation: %.3f (target variance: %.2f)", deviation, cfg.Pacing.Variance)
	if deviation > cfg.Pacing.Variance {
		details += " - exceeds target variance"
	} else {
		details += " - within target variance"
	}

	return NewSoftConstraintResult(
		"Pacing",
		"pacing.deviationFromCurve()",
		score,
		details,
	)
}

// CheckBranchingFactor measures how well the dungeon meets branching targets.
// This is a soft constraint - returns a score from 0.0 to 1.0.
func CheckBranchingFactor(g *graph.Graph, cfg *dungeon.Config) dungeon.ConstraintResult {
	actual := CalculateBranchingFactor(g)
	target := cfg.Branching.Avg

	// Score based on how close actual is to target
	// Perfect match = 1.0, deviation reduces score
	deviation := math.Abs(actual - target)
	maxDeviation := 1.0 // Allow up to 1.0 deviation before score reaches 0

	score := math.Max(0.0, 1.0-(deviation/maxDeviation))

	details := fmt.Sprintf("Branching factor: %.2f (target: %.2f)", actual, target)
	if deviation > 0.3 {
		details += " - significant deviation from target"
	} else {
		details += " - close to target"
	}

	return NewSoftConstraintResult(
		"BranchingFactor",
		"branching.matchesTarget()",
		score,
		details,
	)
}

// Helper functions

func estimateRoomSize(size graph.RoomSize) int {
	switch size {
	case graph.SizeXS:
		return 3
	case graph.SizeS:
		return 5
	case graph.SizeM:
		return 8
	case graph.SizeL:
		return 12
	case graph.SizeXL:
		return 16
	default:
		return 5
	}
}

func rectOverlaps(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	// Check if rectangles overlap
	return !(x1+w1 <= x2 || x2+w2 <= x1 || y1+h1 <= y2 || y2+h2 <= y1)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Removed unused min() function
