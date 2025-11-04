package carving

import (
	"context"
	"fmt"
)

// Carver converts spatial layouts into rasterized tile maps.
// It stamps room footprints, routes corridors, and places doors.
type Carver interface {
	// Carve transforms a Layout into a TileMap with multiple layers.
	// The graph is used to determine room sizes and connector types.
	Carve(ctx context.Context, g Graph, layout *Layout) (*TileMap, error)
}

// CarverRegistry manages available carver implementations.
type CarverRegistry struct {
	carvers map[string]Carver
}

// NewCarverRegistry creates a new empty registry.
func NewCarverRegistry() *CarverRegistry {
	return &CarverRegistry{
		carvers: make(map[string]Carver),
	}
}

// Register adds a carver implementation to the registry.
func (r *CarverRegistry) Register(name string, carver Carver) error {
	if name == "" {
		return fmt.Errorf("carver name cannot be empty")
	}
	if carver == nil {
		return fmt.Errorf("carver cannot be nil")
	}
	if _, exists := r.carvers[name]; exists {
		return fmt.Errorf("carver %q already registered", name)
	}
	r.carvers[name] = carver
	return nil
}

// Get retrieves a carver by name.
func (r *CarverRegistry) Get(name string) (Carver, error) {
	carver, exists := r.carvers[name]
	if !exists {
		return nil, fmt.Errorf("carver %q not found", name)
	}
	return carver, nil
}

// List returns all registered carver names.
func (r *CarverRegistry) List() []string {
	names := make([]string, 0, len(r.carvers))
	for name := range r.carvers {
		names = append(names, name)
	}
	return names
}

// TileType represents different tile categories for the tile map.
type TileType uint32

const (
	TileEmpty TileType = iota // Empty/void space
	TileFloor                 // Walkable floor
	TileWall                  // Solid wall
	TileDoor                  // Traversable door
)

// String returns the string representation of a TileType.
func (t TileType) String() string {
	switch t {
	case TileEmpty:
		return "Empty"
	case TileFloor:
		return "Floor"
	case TileWall:
		return "Wall"
	case TileDoor:
		return "Door"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// DefaultCarver is a basic implementation of the Carver interface.
type DefaultCarver struct {
	tileWidth  int
	tileHeight int
}

// NewDefaultCarver creates a new carver with the specified tile dimensions.
func NewDefaultCarver(tileWidth, tileHeight int) *DefaultCarver {
	if tileWidth <= 0 {
		tileWidth = 16
	}
	if tileHeight <= 0 {
		tileHeight = 16
	}
	return &DefaultCarver{
		tileWidth:  tileWidth,
		tileHeight: tileHeight,
	}
}

// Carve implements the Carver interface.
func (c *DefaultCarver) Carve(ctx context.Context, g Graph, layout *Layout) (*TileMap, error) {
	if g == nil {
		return nil, fmt.Errorf("graph cannot be nil")
	}
	if layout == nil {
		return nil, fmt.Errorf("layout cannot be nil")
	}

	// Calculate tile map dimensions from layout bounds
	width := layout.Bounds.Width
	height := layout.Bounds.Height
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid layout bounds: %dx%d", width, height)
	}

	// Create tile map with layers
	tm := &TileMap{
		Width:      width,
		Height:     height,
		TileWidth:  c.tileWidth,
		TileHeight: c.tileHeight,
		Layers:     make(map[string]*Layer),
	}

	// Initialize layers
	floorLayer := &Layer{
		ID:      0,
		Name:    "floor",
		Type:    "tilelayer",
		Visible: true,
		Opacity: 1.0,
		Data:    make([]uint32, width*height),
	}
	wallLayer := &Layer{
		ID:      1,
		Name:    "walls",
		Type:    "tilelayer",
		Visible: true,
		Opacity: 1.0,
		Data:    make([]uint32, width*height),
	}
	doorLayer := &Layer{
		ID:      2,
		Name:    "doors",
		Type:    "objectgroup",
		Visible: true,
		Opacity: 1.0,
		Objects: []Object{},
	}

	tm.Layers["floor"] = floorLayer
	tm.Layers["walls"] = wallLayer
	tm.Layers["doors"] = doorLayer

	// Stamp room footprints
	stamper := NewStamper(width, height)
	for roomID, pose := range layout.Poses {
		room := g.GetRoom(roomID)
		if room == nil {
			return nil, fmt.Errorf("room %s not found in graph", roomID)
		}

		if err := stamper.StampRoom(room, pose, floorLayer.Data); err != nil {
			return nil, fmt.Errorf("stamping room %s: %w", roomID, err)
		}
	}

	// Route corridors
	router := NewCorridorRouter(width, height)
	for connID, path := range layout.CorridorPaths {
		conn := g.GetConnector(connID)
		if conn == nil {
			return nil, fmt.Errorf("connector %s not found in graph", connID)
		}

		if err := router.RouteCorridor(path, floorLayer.Data); err != nil {
			return nil, fmt.Errorf("routing corridor %s: %w", connID, err)
		}
	}

	// Generate walls around floors
	c.generateWalls(floorLayer.Data, wallLayer.Data, width, height)

	// Place doors at room/corridor junctions
	c.placeDoors(g, layout, floorLayer.Data, doorLayer)

	return tm, nil
}

// generateWalls creates walls around all floor tiles.
func (c *DefaultCarver) generateWalls(floorData, wallData []uint32, width, height int) {
	// For each floor tile, add walls to adjacent empty tiles
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if floorData[idx] != uint32(TileFloor) {
				continue
			}

			// Check all 8 neighbors
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					if nx < 0 || nx >= width || ny < 0 || ny >= height {
						continue
					}

					nidx := ny*width + nx
					// If neighbor is empty and not already a wall, make it a wall
					if floorData[nidx] == uint32(TileEmpty) && wallData[nidx] == uint32(TileEmpty) {
						wallData[nidx] = uint32(TileWall)
					}
				}
			}
		}
	}
}

// placeDoors places door objects at room/corridor junctions.
func (c *DefaultCarver) placeDoors(g Graph, layout *Layout, floorData []uint32, doorLayer *Layer) {
	doorID := 1

	// For each connector, find the junction point and place a door
	for connID, path := range layout.CorridorPaths {
		conn := g.GetConnector(connID)
		if conn == nil || conn.GetType() != TypeCorridor {
			continue
		}

		// Find junction points at start and end of corridor
		if len(path.Points) < 2 {
			continue
		}

		// Place door at corridor entrance
		startPoint := path.Points[0]
		doorObj := Object{
			ID:       doorID,
			Name:     fmt.Sprintf("door_%s_start", connID),
			Type:     "door",
			X:        float64(startPoint.X * c.tileWidth),
			Y:        float64(startPoint.Y * c.tileHeight),
			Width:    float64(c.tileWidth),
			Height:   float64(c.tileHeight),
			Rotation: 0,
			Visible:  true,
			Properties: map[string]interface{}{
				"connector_id": connID,
				"from_room":    conn.GetFrom(),
				"to_room":      conn.GetTo(),
				"gate":         conn.GetGate(),
			},
		}
		doorLayer.Objects = append(doorLayer.Objects, doorObj)
		doorID++

		// Place door at corridor exit
		endPoint := path.Points[len(path.Points)-1]
		doorObj = Object{
			ID:       doorID,
			Name:     fmt.Sprintf("door_%s_end", connID),
			Type:     "door",
			X:        float64(endPoint.X * c.tileWidth),
			Y:        float64(endPoint.Y * c.tileHeight),
			Width:    float64(c.tileWidth),
			Height:   float64(c.tileHeight),
			Rotation: 0,
			Visible:  true,
			Properties: map[string]interface{}{
				"connector_id": connID,
				"from_room":    conn.GetFrom(),
				"to_room":      conn.GetTo(),
				"gate":         conn.GetGate(),
			},
		}
		doorLayer.Objects = append(doorLayer.Objects, doorObj)
		doorID++
	}
}
