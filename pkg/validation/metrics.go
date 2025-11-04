package validation

import (
	"math"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// CalculateBranchingFactor computes the average number of connections per room.
// This is the sum of all edges (counting each edge once) divided by the number of rooms.
func CalculateBranchingFactor(g *graph.Graph) float64 {
	if len(g.Rooms) == 0 {
		return 0.0
	}

	// Count total edges
	// Each connector represents an edge
	totalEdges := len(g.Connectors)

	// Average connections per room
	// Note: Each edge connects 2 rooms, so total degree = 2 * edges
	// Average degree = (2 * edges) / rooms
	avgDegree := float64(2*totalEdges) / float64(len(g.Rooms))

	return avgDegree
}

// CalculatePathLength computes the length of the critical path from Start to Boss.
// Returns 0 if Start or Boss rooms don't exist, or if no path exists.
func CalculatePathLength(g *graph.Graph) int {
	startID := FindStartRoom(g)
	bossID := FindBossRoom(g)

	if startID == "" || bossID == "" {
		return 0
	}

	path, err := g.GetPath(startID, bossID)
	if err != nil {
		return 0
	}

	// Path length is the number of edges, which is (number of nodes - 1)
	return len(path) - 1
}

// CountCycles counts the number of cycles in the graph.
// Uses DFS-based cycle detection. Returns the total number of distinct cycles found.
func CountCycles(g *graph.Graph) int {
	cycles := g.GetCycles()
	return len(cycles)
}

// CalculatePacingDeviation measures how well room difficulties follow the configured pacing curve.
// Returns the L2 (Euclidean) distance between actual and target difficulty distribution.
// Lower values indicate better adherence to the pacing curve.
func CalculatePacingDeviation(g *graph.Graph, cfg *dungeon.Config) float64 {
	startID := FindStartRoom(g)
	bossID := FindBossRoom(g)

	if startID == "" || bossID == "" {
		return 1.0 // Maximum deviation if Start or Boss missing
	}

	// Get the critical path from Start to Boss
	path, err := g.GetPath(startID, bossID)
	if err != nil {
		return 1.0 // Maximum deviation if no path exists
	}

	if len(path) < 2 {
		return 0.0 // Too short to measure deviation
	}

	// Calculate expected difficulty at each step along the path
	sumSquaredError := 0.0
	for i, roomID := range path {
		room := g.Rooms[roomID]
		if room == nil {
			continue
		}

		// Progress along the path (0.0 to 1.0)
		progress := float64(i) / float64(len(path)-1)

		// Expected difficulty based on pacing curve
		expected := calculateExpectedDifficulty(progress, cfg.Pacing)

		// Actual difficulty from room
		actual := room.Difficulty

		// Squared error
		error := expected - actual
		sumSquaredError += error * error
	}

	// L2 distance (root mean squared error)
	if len(path) == 0 {
		return 0.0
	}

	rmse := math.Sqrt(sumSquaredError / float64(len(path)))
	return rmse
}

// calculateExpectedDifficulty computes the expected difficulty at a given progress point
// based on the configured pacing curve.
func calculateExpectedDifficulty(progress float64, pacing dungeon.PacingCfg) float64 {
	switch pacing.Curve {
	case dungeon.PacingLinear:
		return progress

	case dungeon.PacingSCurve:
		// S-curve using logistic function
		// Maps [0,1] to [0,1] with smooth acceleration and deceleration
		// f(x) = 1 / (1 + e^(-k*(x-0.5)))
		// Normalized to [0,1] range
		k := 10.0 // Steepness parameter
		x := progress
		sigmoid := 1.0 / (1.0 + math.Exp(-k*(x-0.5)))
		// Normalize so f(0) ≈ 0 and f(1) ≈ 1
		minVal := 1.0 / (1.0 + math.Exp(k*0.5))
		maxVal := 1.0 / (1.0 + math.Exp(-k*0.5))
		normalized := (sigmoid - minVal) / (maxVal - minVal)
		return normalized

	case dungeon.PacingExponential:
		// Exponential curve: y = x^2
		// Slow start, rapid increase toward end
		return progress * progress

	case dungeon.PacingCustom:
		// Interpolate from custom points
		return interpolateCustomCurve(progress, pacing.CustomPoints)

	default:
		// Default to linear if unknown
		return progress
	}
}

// interpolateCustomCurve performs linear interpolation between custom pacing points.
func interpolateCustomCurve(progress float64, points [][2]float64) float64 {
	if len(points) == 0 {
		return progress
	}

	// Find the two points to interpolate between
	if progress <= points[0][0] {
		return points[0][1]
	}
	if progress >= points[len(points)-1][0] {
		return points[len(points)-1][1]
	}

	// Binary search for the interval
	for i := 0; i < len(points)-1; i++ {
		if progress >= points[i][0] && progress <= points[i+1][0] {
			// Linear interpolation between points[i] and points[i+1]
			x0, y0 := points[i][0], points[i][1]
			x1, y1 := points[i+1][0], points[i+1][1]

			// Interpolate
			t := (progress - x0) / (x1 - x0)
			return y0 + t*(y1-y0)
		}
	}

	// Shouldn't reach here, but return linear as fallback
	return progress
}

// CalculateAverageDifficulty computes the mean difficulty across all rooms.
func CalculateAverageDifficulty(g *graph.Graph) float64 {
	if len(g.Rooms) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, room := range g.Rooms {
		sum += room.Difficulty
	}

	return sum / float64(len(g.Rooms))
}

// CalculateDifficultyStdDev computes the standard deviation of room difficulties.
func CalculateDifficultyStdDev(g *graph.Graph) float64 {
	if len(g.Rooms) == 0 {
		return 0.0
	}

	mean := CalculateAverageDifficulty(g)

	sumSquaredDiff := 0.0
	for _, room := range g.Rooms {
		diff := room.Difficulty - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(g.Rooms))
	return math.Sqrt(variance)
}

// GetDegreeDistribution returns a map of degree (number of connections) to count of rooms.
// Useful for analyzing branching patterns.
func GetDegreeDistribution(g *graph.Graph) map[int]int {
	distribution := make(map[int]int)

	for roomID := range g.Rooms {
		degree := len(g.Adjacency[roomID])
		distribution[degree]++
	}

	return distribution
}

// GetMaxDegree returns the maximum number of connections any room has.
func GetMaxDegree(g *graph.Graph) int {
	maxDegree := 0

	for roomID := range g.Rooms {
		degree := len(g.Adjacency[roomID])
		if degree > maxDegree {
			maxDegree = degree
		}
	}

	return maxDegree
}

// GetMinDegree returns the minimum number of connections any room has.
func GetMinDegree(g *graph.Graph) int {
	if len(g.Rooms) == 0 {
		return 0
	}

	minDegree := int(^uint(0) >> 1) // Max int

	for roomID := range g.Rooms {
		degree := len(g.Adjacency[roomID])
		if degree < minDegree {
			minDegree = degree
		}
	}

	return minDegree
}

// CalculateDiameter computes the longest shortest path between any two rooms.
// This represents the maximum distance across the dungeon.
func CalculateDiameter(g *graph.Graph) int {
	if len(g.Rooms) == 0 {
		return 0
	}

	maxDist := 0

	// For each pair of rooms, find shortest path
	rooms := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		rooms = append(rooms, id)
	}

	for i := 0; i < len(rooms); i++ {
		for j := i + 1; j < len(rooms); j++ {
			path, err := g.GetPath(rooms[i], rooms[j])
			if err != nil {
				continue
			}
			dist := len(path) - 1
			if dist > maxDist {
				maxDist = dist
			}
		}
	}

	return maxDist
}
