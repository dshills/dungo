// Package contracts defines core data types.
// This is a design document showing expected structure, not executable code.
package contracts

// Config specifies all dungeon generation parameters.
type Config struct {
	Seed              uint64
	Size              SizeCfg
	Branching         BranchingCfg
	Pacing            PacingCfg
	Themes            []string
	Keys              []KeyCfg
	Constraints       []Constraint
	AllowDisconnected bool
	SecretDensity     float64
	OptionalRatio     float64
}

type SizeCfg struct {
	RoomsMin int
	RoomsMax int
}

type BranchingCfg struct {
	Avg float64
	Max int
}

type PacingCfg struct {
	Curve        string // "LINEAR", "S_CURVE", "EXPONENTIAL", "CUSTOM"
	Variance     float64
	CustomPoints [][2]float64 // For CUSTOM curve
}

type KeyCfg struct {
	Name  string
	Count int
}

// Artifact is the complete output of generation.
type Artifact struct {
	ADG     *Graph
	Layout  *Layout
	TileMap *TileMap
	Content *Content
	Metrics *Metrics
	Debug   *DebugArtifacts
}

// Graph is the Abstract Dungeon Graph.
type Graph struct {
	Rooms      map[string]*Room
	Connectors map[string]*Connector
	Adjacency  map[string][]string
	Seed       uint64
	Metadata   map[string]interface{}
}

// Room is a node in the ADG.
type Room struct {
	ID           string
	Archetype    RoomArchetype
	Size         RoomSize
	Tags         map[string]string
	Difficulty   float64
	Reward       float64
	Requirements []Requirement
	Provides     []Capability
	DegreeMin    int
	DegreeMax    int
}

type RoomArchetype int

const (
	Start RoomArchetype = iota
	Boss
	Treasure
	Puzzle
	Hub
	Corridor
	Secret
	Optional
	Vendor
	Shrine
	Checkpoint
)

type RoomSize int

const (
	XS RoomSize = iota // Tiny corridor
	S                  // Small chamber
	M                  // Medium hall
	L                  // Large room
	XL                 // Boss arena
)

// Connector is an edge in the ADG.
type Connector struct {
	ID            string
	From          string
	To            string
	Type          ConnectorType
	Gate          *Gate
	Cost          float64
	Visibility    VisibilityType
	Bidirectional bool
}

type ConnectorType int

const (
	Door ConnectorType = iota
	CorridorConn
	Ladder
	Teleporter
	Hidden
	OneWay
)

type VisibilityType int

const (
	Normal VisibilityType = iota
	SecretVis
	Illusory
)

type Gate struct {
	Type  string // "key", "puzzle", "ability"
	Value string // Specific requirement
}

// Requirement is a prerequisite to enter/use something.
type Requirement struct {
	Type  string
	Value string
}

// Capability is something granted by a room/item.
type Capability struct {
	Type  string
	Value string
}

// Constraint is a rule to satisfy or optimize.
type Constraint struct {
	Kind     ConstraintKind
	Severity ConstraintSeverity
	Expr     string
	Priority int
}

type ConstraintKind int

const (
	Connectivity ConstraintKind = iota
	Degree
	KeyLock
	Pacing
	Theme
	Spatial
	Cycle
	PathLen
	BranchFactor
	SecretDensity
	Optionality
	LootBudget
)

type ConstraintSeverity int

const (
	Hard ConstraintSeverity = iota // Must pass
	Soft                           // Optimize
)

// Layout contains spatial embedding data.
type Layout struct {
	Poses         map[string]Pose
	CorridorPaths map[string]Path
	Bounds        Rect
}

type Pose struct {
	X           int
	Y           int
	Rotation    int
	FootprintID string
}

type Path struct {
	Points []Point
}

type Point struct {
	X, Y int
}

type Rect struct {
	X, Y, Width, Height int
}

// TileMap is the rasterized dungeon.
type TileMap struct {
	Width      int
	Height     int
	TileWidth  int
	TileHeight int
	Layers     map[string]*Layer
}

type Layer struct {
	ID      int
	Name    string
	Type    string // "tilelayer", "objectgroup"
	Visible bool
	Opacity float64
	Data    []uint32 // For tile layers
	Objects []Object // For object layers
}

type Object struct {
	ID         int
	Name       string
	Type       string
	X          float64
	Y          float64
	Width      float64
	Height     float64
	Rotation   float64
	GID        uint32
	Visible    bool
	Properties map[string]interface{}
}

// Content contains gameplay elements.
type Content struct {
	Spawns  []Spawn
	Loot    []Loot
	Puzzles []PuzzleInstance
	Secrets []SecretInstance
}

type Spawn struct {
	ID         string
	RoomID     string
	Position   Point
	EnemyType  string
	Count      int
	PatrolPath []Point
}

type Loot struct {
	ID       string
	RoomID   string
	Position Point
	ItemType string
	Value    int
	Required bool
}

type PuzzleInstance struct {
	ID           string
	RoomID       string
	Type         string
	Requirements []Requirement
	Provides     []Capability
	Difficulty   float64
}

type SecretInstance struct {
	ID       string
	RoomID   string
	Type     string
	Position Point
	Clues    []string
}

// Metrics contains generation statistics.
type Metrics struct {
	BranchingFactor   float64
	PathLength        int
	CycleCount        int
	PacingDeviation   float64
	SecretFindability float64
}

// ValidationReport contains validation results.
type ValidationReport struct {
	Passed                bool
	HardConstraintResults []ConstraintResult
	SoftConstraintResults []ConstraintResult
	Metrics               *Metrics
	Warnings              []string
	Errors                []string
}

type ConstraintResult struct {
	Constraint *Constraint
	Satisfied  bool
	Score      float64
	Details    string
}

// DebugArtifacts contains optional debug outputs.
type DebugArtifacts struct {
	ADGSVG    []byte
	LayoutPNG []byte
	Report    *ValidationReport
}

// ThemePack contains content for a biome.
type ThemePack struct {
	Name            string
	Tilesets        map[string]*Tileset
	RoomShapes      map[RoomSize][]*RoomTemplate
	EncounterTables map[float64][]WeightedEntry
	LootTables      map[float64][]WeightedEntry
	Decorators      []DecoratorRule
}

type Tileset struct {
	Name       string
	Image      string
	TileWidth  int
	TileHeight int
	TileCount  int
	FirstGID   uint32
}

type RoomTemplate struct {
	ID     string
	Width  int
	Height int
	Shape  string // "rect", "oval", "irregular"
	Data   []byte // Footprint mask
}

type WeightedEntry struct {
	Type   string
	Weight int
	Data   map[string]interface{}
}

type DecoratorRule struct {
	Condition string
	Actions   []string
}

// SVGOptions configures SVG export.
type SVGOptions struct {
	Width       int
	Height      int
	ShowLabels  bool
	ColorByType bool
	ShowHeatmap bool
	ShowLegend  bool
}
