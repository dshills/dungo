package embedding

import (
	"fmt"

	"github.com/dshills/dungo/pkg/graph"
)

// Pose represents the spatial placement of a room in 2D space.
// It includes position, rotation, and the selected footprint template.
type Pose struct {
	// X coordinate in grid units
	X float64 `json:"x"`

	// Y coordinate in grid units
	Y float64 `json:"y"`

	// Rotation in degrees (0, 90, 180, 270)
	Rotation int `json:"rotation"`

	// Width of the room's bounding box in grid units
	Width int `json:"width"`

	// Height of the room's bounding box in grid units
	Height int `json:"height"`

	// FootprintID identifies which template/shape was used
	FootprintID string `json:"footprintId,omitempty"`
}

// Bounds returns the axis-aligned bounding box for this pose.
// Returns (minX, minY, maxX, maxY) in grid units.
func (p *Pose) Bounds() (float64, float64, float64, float64) {
	return p.X, p.Y, p.X + float64(p.Width), p.Y + float64(p.Height)
}

// Center returns the center point of the room's bounding box.
func (p *Pose) Center() (float64, float64) {
	return p.X + float64(p.Width)/2, p.Y + float64(p.Height)/2
}

// Overlaps checks if this pose's bounding box intersects with another pose.
func (p *Pose) Overlaps(other *Pose) bool {
	minX1, minY1, maxX1, maxY1 := p.Bounds()
	minX2, minY2, maxX2, maxY2 := other.Bounds()

	// No overlap if one is completely to the left/right/above/below the other
	if maxX1 <= minX2 || maxX2 <= minX1 {
		return false
	}
	if maxY1 <= minY2 || maxY2 <= minY1 {
		return false
	}

	return true
}

// Validate checks if the pose has valid values.
func (p *Pose) Validate() error {
	if p.Width <= 0 {
		return fmt.Errorf("pose width must be > 0, got %d", p.Width)
	}
	if p.Height <= 0 {
		return fmt.Errorf("pose height must be > 0, got %d", p.Height)
	}
	if p.Rotation%90 != 0 || p.Rotation < 0 || p.Rotation >= 360 {
		return fmt.Errorf("pose rotation must be 0, 90, 180, or 270, got %d", p.Rotation)
	}
	return nil
}

// String returns a human-readable representation of the Pose.
func (p *Pose) String() string {
	return fmt.Sprintf("Pose[(%0.1f, %0.1f) %dx%d rot=%d]",
		p.X, p.Y, p.Width, p.Height, p.Rotation)
}

// Point represents a 2D coordinate point.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Path represents a polyline path through grid coordinates.
// Used for corridors connecting rooms.
type Path struct {
	// Points in the polyline, in order
	Points []Point `json:"points"`

	// DoorPositions specifies where doors should be placed along the path.
	// Indices refer to segments between points (segment i connects Points[i] and Points[i+1])
	DoorPositions []int `json:"doorPositions,omitempty"`
}

// Length returns the Manhattan distance of the path.
func (p *Path) Length() float64 {
	if len(p.Points) < 2 {
		return 0
	}

	length := 0.0
	for i := 0; i < len(p.Points)-1; i++ {
		dx := p.Points[i+1].X - p.Points[i].X
		dy := p.Points[i+1].Y - p.Points[i].Y
		// Manhattan distance for each segment
		length += abs(dx) + abs(dy)
	}
	return length
}

// BendCount returns the number of direction changes in the path.
func (p *Path) BendCount() int {
	if len(p.Points) < 3 {
		return 0
	}

	bends := 0
	for i := 1; i < len(p.Points)-1; i++ {
		// Check if direction changes at point i
		dx1 := p.Points[i].X - p.Points[i-1].X
		dy1 := p.Points[i].Y - p.Points[i-1].Y
		dx2 := p.Points[i+1].X - p.Points[i].X
		dy2 := p.Points[i+1].Y - p.Points[i].Y

		// Direction changes if the delta changes
		if (dx1 == 0 && dx2 != 0) || (dx1 != 0 && dx2 == 0) ||
			(dy1 == 0 && dy2 != 0) || (dy1 != 0 && dy2 == 0) {
			bends++
		}
	}
	return bends
}

// Validate checks if the path is valid.
func (p *Path) Validate() error {
	if len(p.Points) < 2 {
		return fmt.Errorf("path must have at least 2 points, got %d", len(p.Points))
	}

	// Check that door positions are valid
	maxSegment := len(p.Points) - 2
	for _, doorPos := range p.DoorPositions {
		if doorPos < 0 || doorPos > maxSegment {
			return fmt.Errorf("door position %d out of range [0, %d]", doorPos, maxSegment)
		}
	}

	return nil
}

// Rect represents an axis-aligned bounding rectangle.
type Rect struct {
	MinX float64 `json:"minX"`
	MinY float64 `json:"minY"`
	MaxX float64 `json:"maxX"`
	MaxY float64 `json:"maxY"`
}

// Width returns the width of the rectangle.
func (r *Rect) Width() float64 {
	return r.MaxX - r.MinX
}

// Height returns the height of the rectangle.
func (r *Rect) Height() float64 {
	return r.MaxY - r.MinY
}

// Contains checks if a point is inside the rectangle.
func (r *Rect) Contains(x, y float64) bool {
	return x >= r.MinX && x <= r.MaxX && y >= r.MinY && y <= r.MaxY
}

// Layout represents the complete spatial embedding of an Abstract Dungeon Graph.
// It maps room IDs to their spatial poses and connector IDs to corridor paths.
type Layout struct {
	// Poses maps room IDs to their spatial placements
	Poses map[string]*Pose `json:"poses"`

	// CorridorPaths maps connector IDs to their polyline paths
	CorridorPaths map[string]*Path `json:"corridorPaths"`

	// Bounds is the overall bounding box containing all rooms and corridors
	Bounds Rect `json:"bounds"`

	// Seed is the RNG seed used for this layout (for debugging)
	Seed uint64 `json:"seed,omitempty"`

	// Algorithm identifies which embedder produced this layout
	Algorithm string `json:"algorithm,omitempty"`
}

// NewLayout creates an empty layout with initialized maps.
func NewLayout() *Layout {
	return &Layout{
		Poses:         make(map[string]*Pose),
		CorridorPaths: make(map[string]*Path),
	}
}

// AddPose adds a room pose to the layout.
func (l *Layout) AddPose(roomID string, pose *Pose) error {
	if pose == nil {
		return fmt.Errorf("cannot add nil pose for room %s", roomID)
	}

	if err := pose.Validate(); err != nil {
		return fmt.Errorf("invalid pose for room %s: %w", roomID, err)
	}

	l.Poses[roomID] = pose
	return nil
}

// AddPath adds a corridor path to the layout.
func (l *Layout) AddPath(connectorID string, path *Path) error {
	if path == nil {
		return fmt.Errorf("cannot add nil path for connector %s", connectorID)
	}

	if err := path.Validate(); err != nil {
		return fmt.Errorf("invalid path for connector %s: %w", connectorID, err)
	}

	l.CorridorPaths[connectorID] = path
	return nil
}

// ComputeBounds calculates the bounding box that contains all rooms and corridors.
func (l *Layout) ComputeBounds() {
	if len(l.Poses) == 0 {
		l.Bounds = Rect{0, 0, 0, 0}
		return
	}

	// Initialize with first room bounds
	var initialized bool
	for _, pose := range l.Poses {
		minX, minY, maxX, maxY := pose.Bounds()
		if !initialized {
			l.Bounds = Rect{minX, minY, maxX, maxY}
			initialized = true
		} else {
			l.Bounds.MinX = min(l.Bounds.MinX, minX)
			l.Bounds.MinY = min(l.Bounds.MinY, minY)
			l.Bounds.MaxX = max(l.Bounds.MaxX, maxX)
			l.Bounds.MaxY = max(l.Bounds.MaxY, maxY)
		}
	}

	// Expand bounds to include corridor paths
	for _, path := range l.CorridorPaths {
		for _, pt := range path.Points {
			l.Bounds.MinX = min(l.Bounds.MinX, pt.X)
			l.Bounds.MinY = min(l.Bounds.MinY, pt.Y)
			l.Bounds.MaxX = max(l.Bounds.MaxX, pt.X)
			l.Bounds.MaxY = max(l.Bounds.MaxY, pt.Y)
		}
	}
}

// Validate checks that the layout is valid for the given graph.
func (l *Layout) Validate(g *graph.Graph) error {
	// Check that all rooms have poses
	for roomID := range g.Rooms {
		if _, exists := l.Poses[roomID]; !exists {
			return fmt.Errorf("missing pose for room %s", roomID)
		}
	}

	// Check that all connectors have paths
	for connID := range g.Connectors {
		if _, exists := l.CorridorPaths[connID]; !exists {
			return fmt.Errorf("missing path for connector %s", connID)
		}
	}

	// Check for room overlaps
	rooms := make([]*Pose, 0, len(l.Poses))
	roomIDs := make([]string, 0, len(l.Poses))
	for id, pose := range l.Poses {
		rooms = append(rooms, pose)
		roomIDs = append(roomIDs, id)
	}

	for i := 0; i < len(rooms); i++ {
		for j := i + 1; j < len(rooms); j++ {
			if rooms[i].Overlaps(rooms[j]) {
				return fmt.Errorf("rooms %s and %s have overlapping bounding boxes",
					roomIDs[i], roomIDs[j])
			}
		}
	}

	return nil
}

// Helper functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
