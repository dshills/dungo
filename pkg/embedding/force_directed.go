package embedding

import (
	"fmt"
	"math"
	"sort"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// ForceDirectedEmbedder uses a force-directed layout algorithm to position rooms.
// It simulates spring forces (connected rooms attract) and repulsion forces
// (all rooms repel) to find a stable configuration, then quantizes to a grid
// and routes corridors.
type ForceDirectedEmbedder struct {
	config *Config
}

// NewForceDirectedEmbedder creates a force-directed embedder with the given config.
func NewForceDirectedEmbedder(config *Config) *ForceDirectedEmbedder {
	if config == nil {
		config = DefaultConfig()
	}
	return &ForceDirectedEmbedder{config: config}
}

// Name returns the identifier for this embedder.
func (e *ForceDirectedEmbedder) Name() string {
	return "force_directed"
}

// Embed performs force-directed layout of the graph.
func (e *ForceDirectedEmbedder) Embed(g *graph.Graph, rng *rng.RNG) (*Layout, error) {
	if g == nil {
		return nil, fmt.Errorf("cannot embed nil graph")
	}
	if rng == nil {
		return nil, fmt.Errorf("cannot embed with nil RNG")
	}

	if len(g.Rooms) == 0 {
		return nil, fmt.Errorf("cannot embed graph with no rooms")
	}

	// Phase 1: Initialize room positions randomly
	positions := e.initializePositions(g, rng)

	// Phase 2: Run force-directed simulation
	if err := e.simulateForces(g, positions, rng); err != nil {
		return nil, fmt.Errorf("force simulation failed: %w", err)
	}

	// Phase 3: Quantize to grid
	e.quantizeToGrid(positions)

	// Phase 4: Resolve overlaps
	if err := e.resolveOverlaps(g, positions, rng); err != nil {
		return nil, fmt.Errorf("overlap resolution failed: %w", err)
	}

	// Phase 5: Convert to Layout with Poses
	// Use sorted room IDs for deterministic layout construction
	layout := NewLayout()
	layout.Algorithm = e.Name()
	layout.Seed = rng.Seed()

	// Sort room IDs for deterministic order
	roomIDs := make([]string, 0, len(positions))
	for id := range positions {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	for _, roomID := range roomIDs {
		pos := positions[roomID]
		room := g.Rooms[roomID]
		width, height := SizeToGridDimensions(room.Size)

		pose := &Pose{
			X:        pos.x,
			Y:        pos.y,
			Width:    width,
			Height:   height,
			Rotation: 0, // Can be enhanced later for rotation support
		}

		if err := layout.AddPose(roomID, pose); err != nil {
			return nil, fmt.Errorf("failed to add pose: %w", err)
		}
	}

	// Phase 6: Route corridors between connected rooms
	if err := e.routeCorridors(g, layout, rng); err != nil {
		return nil, fmt.Errorf("corridor routing failed: %w", err)
	}

	// Phase 7: Compute final bounds
	layout.ComputeBounds()

	// Phase 8: Validate the embedding
	if err := ValidateEmbedding(layout, g, e.config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return layout, nil
}

// position tracks continuous 2D position and velocity during simulation.
type position struct {
	x, y   float64 // Current position
	vx, vy float64 // Current velocity
}

// initializePositions places rooms at random positions in a circle.
// CRITICAL: Uses sorted room IDs to ensure deterministic initialization.
func (e *ForceDirectedEmbedder) initializePositions(g *graph.Graph, rng *rng.RNG) map[string]*position {
	positions := make(map[string]*position, len(g.Rooms))

	// Sort room IDs for deterministic iteration order
	roomIDs := make([]string, 0, len(g.Rooms))
	for roomID := range g.Rooms {
		roomIDs = append(roomIDs, roomID)
	}
	sort.Strings(roomIDs)

	// Initialize positions in deterministic order
	for _, roomID := range roomIDs {
		// Random angle and radius for circular initial placement
		angle := rng.Float64() * 2 * math.Pi
		radius := rng.Float64() * e.config.InitialSpread

		positions[roomID] = &position{
			x:  radius * math.Cos(angle),
			y:  radius * math.Sin(angle),
			vx: 0,
			vy: 0,
		}
	}

	return positions
}

// simulateForces runs the force-directed simulation.
// CRITICAL: Uses sorted room IDs throughout to ensure deterministic force calculations.
func (e *ForceDirectedEmbedder) simulateForces(g *graph.Graph, positions map[string]*position, rng *rng.RNG) error {
	dt := 0.1 // Time step

	// Create sorted room IDs once for deterministic iteration
	roomIDs := make([]string, 0, len(positions))
	for id := range positions {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	for iter := 0; iter < e.config.MaxIterations; iter++ {
		// Calculate forces for each room
		forces := make(map[string]struct{ fx, fy float64 }, len(positions))

		// Initialize forces to zero in deterministic order
		for _, roomID := range roomIDs {
			forces[roomID] = struct{ fx, fy float64 }{0, 0}
		}

		// Apply spring forces (attraction between connected rooms)
		for _, conn := range g.Connectors {
			fromPos := positions[conn.From]
			toPos := positions[conn.To]

			dx := toPos.x - fromPos.x
			dy := toPos.y - fromPos.y
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > 0.001 { // Avoid division by zero
				// Spring force: F = k * distance
				forceMag := e.config.SpringConstant * dist
				fx := forceMag * dx / dist
				fy := forceMag * dy / dist

				// Apply to both rooms (Newton's third law)
				fromForce := forces[conn.From]
				fromForce.fx += fx
				fromForce.fy += fy
				forces[conn.From] = fromForce

				toForce := forces[conn.To]
				toForce.fx -= fx
				toForce.fy -= fy
				forces[conn.To] = toForce
			}
		}

		// Apply repulsion forces (all rooms repel each other)
		// Use the sorted roomIDs from above for deterministic pair iteration
		for i := 0; i < len(roomIDs); i++ {
			for j := i + 1; j < len(roomIDs); j++ {
				id1 := roomIDs[i]
				id2 := roomIDs[j]
				pos1 := positions[id1]
				pos2 := positions[id2]

				dx := pos2.x - pos1.x
				dy := pos2.y - pos1.y
				distSq := dx*dx + dy*dy

				if distSq > 0.001 { // Avoid division by zero
					dist := math.Sqrt(distSq)

					// Repulsion force: F = k / distance^2
					forceMag := e.config.RepulsionConstant / distSq
					fx := forceMag * dx / dist
					fy := forceMag * dy / dist

					// Apply to both rooms
					force1 := forces[id1]
					force1.fx -= fx
					force1.fy -= fy
					forces[id1] = force1

					force2 := forces[id2]
					force2.fx += fx
					force2.fy += fy
					forces[id2] = force2
				}
			}
		}

		// Update velocities and positions with damping in deterministic order
		maxMovement := 0.0
		for _, roomID := range roomIDs {
			pos := positions[roomID]
			force := forces[roomID]

			// Update velocity: v = v * damping + F * dt
			pos.vx = pos.vx*e.config.DampingFactor + force.fx*dt
			pos.vy = pos.vy*e.config.DampingFactor + force.fy*dt

			// Update position: p = p + v * dt
			pos.x += pos.vx * dt
			pos.y += pos.vy * dt

			// Track maximum movement for stability check
			movement := math.Sqrt(pos.vx*pos.vx + pos.vy*pos.vy)
			if movement > maxMovement {
				maxMovement = movement
			}
		}

		// Check for stability (early exit if movement is small)
		if maxMovement < e.config.StabilityThreshold {
			break
		}
	}

	return nil
}

// quantizeToGrid snaps positions to the grid.
func (e *ForceDirectedEmbedder) quantizeToGrid(positions map[string]*position) {
	if e.config.GridQuantization <= 0 {
		return // No quantization
	}

	for _, pos := range positions {
		pos.x = math.Round(pos.x/e.config.GridQuantization) * e.config.GridQuantization
		pos.y = math.Round(pos.y/e.config.GridQuantization) * e.config.GridQuantization
		pos.vx = 0
		pos.vy = 0
	}
}

// resolveOverlaps uses an iterative algorithm to separate overlapping rooms.
func (e *ForceDirectedEmbedder) resolveOverlaps(g *graph.Graph, positions map[string]*position, rng *rng.RNG) error {
	maxAttempts := 200

	for attempt := 0; attempt < maxAttempts; attempt++ {
		overlaps := e.findOverlaps(g, positions)
		if len(overlaps) == 0 {
			return nil // Success
		}

		// Resolve all overlaps in this iteration
		for _, overlap := range overlaps {
			e.separateRooms(g, positions, overlap.id1, overlap.id2)
		}

		// Re-quantize after separation
		e.quantizeToGrid(positions)

		// Add small deterministic perturbation to help escape local minima
		if attempt%20 == 19 {
			// Sort room IDs for deterministic iteration order
			roomIDs := make([]string, 0, len(positions))
			for id := range positions {
				roomIDs = append(roomIDs, id)
			}
			sort.Strings(roomIDs)

			for _, id := range roomIDs {
				pos := positions[id]
				pos.x += (rng.Float64() - 0.5) * e.config.GridQuantization
				pos.y += (rng.Float64() - 0.5) * e.config.GridQuantization
			}
		}
	}

	// Failed to resolve all overlaps
	overlaps := e.findOverlaps(g, positions)
	if len(overlaps) > 0 {
		return fmt.Errorf("failed to resolve %d overlaps after %d attempts", len(overlaps), maxAttempts)
	}

	return nil
}

type overlap struct {
	id1, id2 string
}

// findOverlaps detects all pairs of rooms with overlapping bounding boxes.
func (e *ForceDirectedEmbedder) findOverlaps(g *graph.Graph, positions map[string]*position) []overlap {
	overlaps := []overlap{}

	roomIDs := make([]string, 0, len(positions))
	for id := range positions {
		roomIDs = append(roomIDs, id)
	}

	for i := 0; i < len(roomIDs); i++ {
		for j := i + 1; j < len(roomIDs); j++ {
			id1 := roomIDs[i]
			id2 := roomIDs[j]

			if e.roomsOverlap(g, positions, id1, id2) {
				overlaps = append(overlaps, overlap{id1, id2})
			}
		}
	}

	return overlaps
}

// roomsOverlap checks if two rooms have overlapping bounding boxes.
func (e *ForceDirectedEmbedder) roomsOverlap(g *graph.Graph, positions map[string]*position, id1, id2 string) bool {
	pos1 := positions[id1]
	pos2 := positions[id2]
	room1 := g.Rooms[id1]
	room2 := g.Rooms[id2]

	w1, h1 := SizeToGridDimensions(room1.Size)
	w2, h2 := SizeToGridDimensions(room2.Size)

	// Bounding boxes
	minX1, minY1 := pos1.x, pos1.y
	maxX1, maxY1 := pos1.x+float64(w1), pos1.y+float64(h1)

	minX2, minY2 := pos2.x, pos2.y
	maxX2, maxY2 := pos2.x+float64(w2), pos2.y+float64(h2)

	// Check for overlap with spacing consideration
	spacing := e.config.MinRoomSpacing
	if maxX1+spacing <= minX2 || maxX2+spacing <= minX1 {
		return false
	}
	if maxY1+spacing <= minY2 || maxY2+spacing <= minY1 {
		return false
	}

	return true
}

// separateRooms pushes two overlapping rooms apart along the shortest axis.
func (e *ForceDirectedEmbedder) separateRooms(g *graph.Graph, positions map[string]*position, id1, id2 string) {
	pos1 := positions[id1]
	pos2 := positions[id2]
	room1 := g.Rooms[id1]
	room2 := g.Rooms[id2]

	w1, h1 := SizeToGridDimensions(room1.Size)
	w2, h2 := SizeToGridDimensions(room2.Size)

	// Calculate bounding boxes
	minX1, minY1 := pos1.x, pos1.y
	maxX1, maxY1 := pos1.x+float64(w1), pos1.y+float64(h1)
	minX2, minY2 := pos2.x, pos2.y
	maxX2, maxY2 := pos2.x+float64(w2), pos2.y+float64(h2)

	// Calculate overlap in each dimension
	overlapX := math.Min(maxX1, maxX2) - math.Max(minX1, minX2)
	overlapY := math.Min(maxY1, maxY2) - math.Max(minY1, minY2)

	// Add spacing requirement
	requiredSpacing := e.config.MinRoomSpacing

	// Separate along the axis with smaller overlap (easier to fix)
	if overlapX < overlapY {
		// Separate horizontally
		separation := (overlapX + requiredSpacing) / 2
		if pos1.x < pos2.x {
			pos1.x -= separation
			pos2.x += separation
		} else {
			pos1.x += separation
			pos2.x -= separation
		}
	} else {
		// Separate vertically
		separation := (overlapY + requiredSpacing) / 2
		if pos1.y < pos2.y {
			pos1.y -= separation
			pos2.y += separation
		} else {
			pos1.y += separation
			pos2.y -= separation
		}
	}
}

// routeCorridors creates Manhattan paths between all connected rooms.
func (e *ForceDirectedEmbedder) routeCorridors(g *graph.Graph, layout *Layout, rng *rng.RNG) error {
	for connID, conn := range g.Connectors {
		fromPose := layout.Poses[conn.From]
		toPose := layout.Poses[conn.To]

		// Get room centers
		fromX, fromY := fromPose.Center()
		toX, toY := toPose.Center()

		// Use A* for pathfinding
		path, err := e.findPath(layout, fromX, fromY, toX, toY)
		if err != nil {
			// Fallback to simple Manhattan path
			path = e.manhattanPath(fromX, fromY, toX, toY)
		}

		if err := layout.AddPath(connID, path); err != nil {
			return fmt.Errorf("failed to add path for %s: %w", connID, err)
		}
	}

	return nil
}

// manhattanPath creates a simple L-shaped path between two points.
func (e *ForceDirectedEmbedder) manhattanPath(x1, y1, x2, y2 float64) *Path {
	// Create L-shaped path (horizontal then vertical, or vice versa based on distance)
	points := []Point{
		{X: x1, Y: y1},
	}

	// Choose route based on which dimension has larger distance
	dx := math.Abs(x2 - x1)
	dy := math.Abs(y2 - y1)

	if dx > dy {
		// Go horizontal first, then vertical
		points = append(points, Point{X: x2, Y: y1})
	} else {
		// Go vertical first, then horizontal
		points = append(points, Point{X: x1, Y: y2})
	}

	points = append(points, Point{X: x2, Y: y2})

	return &Path{Points: points}
}

// findPath uses A* to find a path avoiding room bounding boxes.
func (e *ForceDirectedEmbedder) findPath(layout *Layout, x1, y1, x2, y2 float64) (*Path, error) {
	// Simple A* implementation for grid pathfinding
	// For now, just use Manhattan path as A* is complex
	// This can be enhanced later with proper A* implementation

	// Create waypoints that route around obstacles
	path := e.manhattanPath(x1, y1, x2, y2)

	// Validate path doesn't exceed constraints
	if path.Length() > e.config.CorridorMaxLength {
		return nil, fmt.Errorf("path too long: %.1f > %.1f", path.Length(), e.config.CorridorMaxLength)
	}

	if path.BendCount() > e.config.CorridorMaxBends {
		return nil, fmt.Errorf("path has too many bends: %d > %d", path.BendCount(), e.config.CorridorMaxBends)
	}

	return path, nil
}

// Register the force-directed embedder
func init() {
	Register("force_directed", func(config *Config) Embedder {
		return NewForceDirectedEmbedder(config)
	})
}
