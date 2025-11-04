package carving

// Point represents a 2D coordinate.
type Point struct {
	X, Y int
}

// Rect represents a rectangular bounds.
type Rect struct {
	X, Y, Width, Height int
}

// Pose represents the spatial position and orientation of a room.
type Pose struct {
	X           int
	Y           int
	Rotation    int    // Degrees: 0, 90, 180, 270
	FootprintID string // Reference to room template shape
}

// Path represents a polyline path (for corridors).
type Path struct {
	Points []Point
}

// Layout contains the spatial embedding of the dungeon graph.
// It maps abstract rooms to concrete positions and corridors to paths.
type Layout struct {
	Poses         map[string]Pose // Room ID → position/rotation
	CorridorPaths map[string]Path // Connector ID → polyline path
	Bounds        Rect            // Overall dungeon extents
}

// TileMap is the rasterized dungeon with layered tiles.
// It represents the final 2D grid representation of the dungeon.
type TileMap struct {
	Width      int               // Grid width in tiles
	Height     int               // Grid height in tiles
	TileWidth  int               // Tile width in pixels
	TileHeight int               // Tile height in pixels
	Layers     map[string]*Layer // Named tile/object layers
}

// Layer represents a single layer in the tile map.
// Can be either a tile layer (grid) or an object layer (entities).
type Layer struct {
	ID      int      // Layer identifier
	Name    string   // Layer name
	Type    string   // "tilelayer" or "objectgroup"
	Visible bool     // Visibility flag
	Opacity float64  // Layer opacity (0.0-1.0)
	Data    []uint32 // Tile data (for tile layers), flat row-major array
	Objects []Object // Objects (for object layers)
}

// Object represents an entity in an object layer.
type Object struct {
	ID         int                    // Object identifier
	Name       string                 // Object name
	Type       string                 // Object type
	X          float64                // X position
	Y          float64                // Y position
	Width      float64                // Object width
	Height     float64                // Object height
	Rotation   float64                // Rotation in degrees
	GID        uint32                 // Tile GID (if tile object)
	Visible    bool                   // Visibility flag
	Properties map[string]interface{} // Custom properties
}

// RoomSize defines the abstract size class of a room.
type RoomSize int

const (
	SizeXS RoomSize = iota // Tiny corridor
	SizeS                  // Small chamber
	SizeM                  // Medium hall
	SizeL                  // Large room
	SizeXL                 // Boss arena
)

// ConnectorType defines the connection mechanism between rooms.
type ConnectorType int

const (
	TypeDoor ConnectorType = iota
	TypeCorridor
	TypeLadder
	TypeTeleporter
	TypeHidden
	TypeOneWay
)

// Gate represents an optional gating requirement for a connector.
type Gate struct {
	Type  string // "key", "puzzle", "ability"
	Value string // Specific gate (e.g., "silver_key", "runes_3")
}

// Graph represents the Abstract Dungeon Graph for carving purposes.
// This is a minimal interface to avoid import cycles with dungeon package.
type Graph interface {
	GetRoom(id string) Room
	GetConnector(id string) Connector
	GetRoomIDs() []string
	GetConnectorIDs() []string
}

// Room represents room data needed for carving.
type Room interface {
	GetID() string
	GetSize() RoomSize
}

// Connector represents connector data needed for carving.
type Connector interface {
	GetID() string
	GetFrom() string
	GetTo() string
	GetType() ConnectorType
	GetGate() *Gate
}
