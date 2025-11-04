# Data Model: Graph-Based Dungeon Generator

**Feature**: 001-dungeon-generator-core
**Date**: 2025-11-04
**Status**: Complete

## Overview

This document defines the core data structures for the dungeon generator, organized by package. All types are designed for deterministic generation, constraint validation, and serialization.

---

## Package: pkg/dungeon

### Config

Specifies all generation parameters.

**Fields:**
- `Seed` (uint64): Master seed for deterministic generation
- `Size` (SizeCfg): Room count constraints
  - `RoomsMin` (int): Minimum room count (10-300)
  - `RoomsMax` (int): Maximum room count (10-300)
- `Branching` (BranchingCfg): Connectivity parameters
  - `Avg` (float64): Target average connections per room (1.5-3.0)
  - `Max` (int): Maximum connections for any single room (2-5)
- `Pacing` (PacingCfg): Difficulty curve configuration
  - `Curve` (string): Curve type ("LINEAR", "S_CURVE", "EXPONENTIAL", "CUSTOM")
  - `Variance` (float64): Allowed deviation from curve (0.0-0.3)
  - `CustomPoints` ([][2]float64): Optional custom pacing points (for CUSTOM curve)
- `Themes` ([]string): Biome/theme names (e.g., ["crypt", "fungal"])
- `Keys` ([]KeyCfg): Key/lock definitions
  - `Name` (string): Key identifier (e.g., "silver", "gold")
  - `Count` (int): Number of this key type (1-5)
- `Constraints` ([]Constraint): Hard and soft constraints
- `AllowDisconnected` (bool): Allow teleport motifs (default: false)
- `SecretDensity` (float64): Target ratio of secret rooms (0.0-0.3)
- `OptionalRatio` (float64): Target ratio of optional rooms (0.1-0.4)

**Validation Rules:**
- `RoomsMin` ≤ `RoomsMax`
- `RoomsMin` ≥ 10, `RoomsMax` ≤ 300
- `Branching.Avg` ≥ 1.0, `Branching.Max` ≥ 2
- `Seed` must be non-zero (0 treated as unset, auto-generated)
- At least one theme must be specified
- Pacing.Curve must be valid enum value

**Serialization:** JSON/YAML

---

### Artifact

Complete dungeon output from generation.

**Fields:**
- `ADG` (graph.Graph): Abstract dungeon graph
- `Layout` (Layout): Spatial embedding with poses
  - `Poses` (map[string]Pose): Room positions/rotations
  - `CorridorPaths` (map[string]Path): Polyline paths for edges
  - `Bounds` (Rect): Overall dungeon bounds
- `TileMap` (TileMap): Rasterized tile layers
  - `Width`, `Height` (int): Grid dimensions
  - `TileWidth`, `TileHeight` (int): Tile size in pixels
  - `Layers` (map[string]Layer): Named tile/object layers
- `Content` (Content): Placed gameplay elements
  - `Spawns` ([]Spawn): Enemy spawn points
  - `Loot` ([]Loot): Treasure locations
  - `Puzzles` ([]Puzzle): Puzzle configurations
  - `Secrets` ([]Secret): Hidden elements
- `Metrics` (Metrics): Generation statistics
  - `BranchingFactor` (float64): Actual avg branching
  - `PathLength` (int): Start→Boss path length
  - `CycleCount` (int): Number of graph cycles
  - `PacingDeviation` (float64): L2 distance from target curve
  - `SecretFindability` (float64): Heuristic score (0.0-1.0)
- `Debug` (DebugArtifacts): Optional debug outputs
  - `ADGSVG` ([]byte): SVG visualization of graph
  - `LayoutPNG` ([]byte): Heatmap overlay image
  - `Report` (ValidationReport): Detailed metrics

**Validation Rules:**
- `ADG` must pass all hard constraints
- `Layout.Bounds` must contain all room poses
- `TileMap` dimensions must be sufficient for layout
- `Content` must respect capacity limits per room
- All references (room IDs, connector IDs) must be valid

**Serialization:** JSON (complete artifact), individual components (SVG, TMJ, etc.)

---

## Package: pkg/graph

### Room

Node in the Abstract Dungeon Graph.

**Fields:**
- `ID` (string): Unique identifier (e.g., "R001", "R042")
- `Archetype` (RoomArchetype): Room type
  - Enum: `Start`, `Boss`, `Treasure`, `Puzzle`, `Hub`, `Corridor`, `Secret`, `Optional`, `Vendor`, `Shrine`, `Checkpoint`
- `Size` (RoomSize): Abstract size class
  - Enum: `XS` (tiny corridor), `S` (small chamber), `M` (medium hall), `L` (large room), `XL` (boss arena)
- `Tags` (map[string]string): Biome and metadata (e.g., "biome:crypt", "element:fire")
- `Difficulty` (float64): Target local intensity (0.0-1.0)
- `Reward` (float64): Treasure value (0.0-1.0)
- `Requirements` ([]Requirement): Prerequisites to enter
  - `Type` (string): "key", "ability", "item"
  - `Value` (string): Specific requirement (e.g., "silver_key", "double_jump")
- `Provides` ([]Capability): What this room grants
  - `Type` (string): "key", "ability", "item"
  - `Value` (string): What's provided
- `DegreeMin`, `DegreeMax` (int): Connection bounds (optional)

**Validation Rules:**
- `ID` must be unique within graph
- `Archetype` must be valid enum
- `Difficulty`, `Reward` in [0.0, 1.0]
- `DegreeMin` ≤ `DegreeMax` if specified
- Exactly one `Start` and one `Boss` per dungeon

**State Transitions:**
- Created → Positioned → Content Assigned → Validated

---

### Connector

Edge in the Abstract Dungeon Graph.

**Fields:**
- `ID` (string): Unique identifier (e.g., "E001", "C042")
- `From`, `To` (string): Room IDs (directed edge)
- `Type` (ConnectorType): Connection mechanism
  - Enum: `Door`, `Corridor`, `Ladder`, `Teleporter`, `Hidden`, `OneWay`
- `Gate` (*Gate): Optional gating requirement
  - `Type` (string): "key", "puzzle", "ability"
  - `Value` (string): Specific gate (e.g., "silver_key", "runes_3")
- `Cost` (float64): Pathfinding weight (1.0 = normal)
- `Visibility` (VisibilityType): Discovery mechanism
  - Enum: `Normal`, `Secret`, `Illusory`
- `Bidirectional` (bool): Can traverse in both directions (default: true)

**Validation Rules:**
- `ID` must be unique within graph
- `From` and `To` must reference existing rooms
- `From` ≠ `To` (no self-loops)
- `Type` must be valid enum
- `Cost` > 0.0
- If `Gate` is present, corresponding key/ability must exist in dungeon

---

### Constraint

Rule that must be satisfied or optimized.

**Fields:**
- `Kind` (ConstraintKind): Category of constraint
  - Enum: `Connectivity`, `Degree`, `KeyLock`, `Pacing`, `Theme`, `Spatial`, `Cycle`, `PathLen`, `BranchFactor`, `SecretDensity`, `Optionality`, `LootBudget`
- `Severity` (ConstraintSeverity): Enforcement level
  - Enum: `Hard` (must pass), `Soft` (optimize)
- `Expr` (string): DSL expression (e.g., "isConnected()", "keyBeforeLock('silver')")
- `Priority` (int): Order for constraint solving (higher = earlier)

**DSL Predicates:**
- Connectivity: `isConnected()`, `hasPath(a, b, minLen?, maxLen?)`
- Degree: `degreeInRange(tag, min, max)`, `branchingAvgIn(min, max)`
- Keys: `keyBeforeLock(keyName)`, `keysReachable(keyName)`
- Pacing: `monotoneIncrease(start, boss, slope?)`, `peakNear(tag, radius)`
- Spatial: `noOverlap()`, `maxCorridorBends(b)`, `maxCorridorLen(l)`
- Cycles: `cycleCountIn(min, max)`
- Composition: `and(...)`, `or(...)`, `not(...)`

**Validation Rules:**
- `Kind` must be valid enum
- `Severity` must be Hard or Soft
- `Expr` must parse successfully
- Hard constraints must never fail (or generation fails)

---

### Graph

Container for complete ADG.

**Fields:**
- `Rooms` (map[string]*Room): All rooms indexed by ID
- `Connectors` (map[string]*Connector): All edges indexed by ID
- `Adjacency` (map[string][]string): Adj list for pathfinding
- `Seed` (uint64): Seed used for this graph generation
- `Metadata` (map[string]interface{}): Extensibility

**Operations:**
- `AddRoom(room *Room) error`: Insert room, update indices
- `AddConnector(conn *Connector) error`: Insert edge, update adjacency
- `RemoveRoom(id string) error`: Delete room and its edges
- `GetPath(from, to string) ([]string, error)`: Find shortest path
- `IsConnected() bool`: Check single component
- `GetCycles() [][]string`: Detect all cycles
- `GetReachable(from string) map[string]bool`: BFS from node

**Validation Rules:**
- All connector From/To IDs must exist in Rooms
- Adjacency must be consistent with Connectors
- At least 2 rooms (Start + Boss minimum)

---

## Package: pkg/rng

### RNG

Deterministic random number generator with sub-seed derivation.

**Fields:**
- `seed` (uint64): Current seed state
- `stageName` (string): Name of pipeline stage ("synthesis", "embedding", etc.)
- `source` (*rand.Rand): Go stdlib random source

**Operations:**
- `NewRNG(masterSeed uint64, stageName string, configHash []byte) *RNG`: Derive stage-specific seed via `H(masterSeed, stageName, configHash)` using crypto/sha256
- `Uint64() uint64`: Generate random uint64
- `Intn(n int) int`: Generate random int in [0, n)
- `Float64() float64`: Generate random float64 in [0.0, 1.0)
- `Shuffle(slice interface{})`: Shuffle slice in-place
- `Choice(items []T) T`: Pick random item from slice

**Validation Rules:**
- Master seed must be non-zero
- Stage name must be non-empty
- Same inputs always produce same RNG sequence

**Rationale:** Pure determinism requires isolated RNG per pipeline stage. Global `rand` usage would break reproducibility.

---

## Package: pkg/synthesis

### SynthesisStrategy

Interface for graph generation approaches.

**Methods:**
- `Synthesize(ctx context.Context, rng *rng.RNG, cfg Config) (*graph.Graph, error)`
- `Name() string`: Strategy identifier ("grammar", "template", "optimizer")

**Implementations:**
- **GrammarSynthesizer**: Production rules with probabilities
- **TemplateSynthesizer**: Pre-authored motif stitching
- **OptimizerSynthesizer**: Search/simulated annealing

---

## Package: pkg/embedding

### Pose

Spatial position and orientation for a room.

**Fields:**
- `X`, `Y` (int): Grid coordinates
- `Rotation` (int): Degrees (0, 90, 180, 270)
- `FootprintID` (string): Reference to room template shape

---

### Layout

Complete spatial embedding.

**Fields:**
- `Poses` (map[string]Pose): Room ID → position
- `CorridorPaths` (map[string]Path): Connector ID → polyline
- `Bounds` (Rect): Overall dungeon extents

**Validation Rules:**
- All room IDs in Poses must exist in ADG
- No room bounding boxes may overlap
- All corridor paths must be feasible (within length/bend limits)

---

## Package: pkg/carving

### TileMap

Rasterized dungeon with layered tiles.

**Fields:**
- `Width`, `Height` (int): Grid dimensions
- `TileWidth`, `TileHeight` (int): Tile size in pixels
- `Layers` (map[string]*TileLayer): Named layers

**Layer Types:**
- **TileLayer**: Grid of tile IDs
  - `Data` ([]uint32): Flat array (row-major order)
  - Tile ID 0 = empty
- **ObjectLayer**: List of entities
  - `Objects` ([]Object): Spawn points, triggers, etc.

---

## Package: pkg/content

### Spawn

Enemy spawn point.

**Fields:**
- `ID` (string): Unique identifier
- `RoomID` (string): Parent room
- `Position` (Point): Location within room
- `EnemyType` (string): Reference to encounter table entry
- `Count` (int): Number of enemies (1-10)
- `PatrolPath` ([]Point): Optional waypoints

---

### Loot

Treasure item.

**Fields:**
- `ID` (string): Unique identifier
- `RoomID` (string): Parent room
- `Position` (Point): Location within room
- `ItemType` (string): Reference to loot table entry
- `Value` (int): Gold/experience worth
- `Required` (bool): Required for progression (keys, abilities)

---

### Puzzle

Interactive challenge.

**Fields:**
- `ID` (string): Unique identifier
- `RoomID` (string): Parent room
- `Type` (string): Puzzle mechanism ("lever", "rune", "statue")
- `Requirements` ([]Requirement): Prerequisites
- `Provides` ([]Capability): Grants on solve
- `Difficulty` (float64): Challenge rating (0.0-1.0)

---

## Package: pkg/validation

### ValidationReport

Detailed metrics and constraint satisfaction.

**Fields:**
- `Passed` (bool): All hard constraints satisfied
- `HardConstraintResults` ([]ConstraintResult): Individual check results
- `SoftConstraintResults` ([]ConstraintResult): Optimization scores
- `Metrics` (Metrics): Calculated statistics
- `Warnings` ([]string): Non-fatal issues
- `Errors` ([]string): Hard constraint failures

---

### ConstraintResult

Result of evaluating a single constraint.

**Fields:**
- `Constraint` (Constraint): The constraint that was checked
- `Satisfied` (bool): Pass/fail
- `Score` (float64): For soft constraints (0.0-1.0, higher is better)
- `Details` (string): Explanation or violation info

---

## Package: pkg/themes

### ThemePack

Complete content definition for a biome.

**Fields:**
- `Name` (string): Theme identifier ("crypt", "fungal", "arcane")
- `Tilesets` (map[string]*Tileset): Tile image collections
- `RoomShapes` (map[RoomSize][]*RoomTemplate): Shape options per size
- `EncounterTables` (map[float64][]WeightedEntry): Difficulty → enemy spawns
- `LootTables` (map[float64][]WeightedEntry): Difficulty → treasure
- `Decorators` ([]DecoratorRule): Prop placement rules

**File Format:** YAML

**Example:**
```yaml
name: crypt
tilesets:
  floor: tiles/crypt_floor.png
  walls: tiles/crypt_walls.png
encounterTables:
  - difficulty: 0.3
    entries:
      - { type: "skeleton", weight: 10 }
      - { type: "zombie", weight: 5 }
  - difficulty: 0.7
    entries:
      - { type: "wraith", weight: 8 }
      - { type: "lich", weight: 2 }
lootTables:
  - difficulty: 0.5
    entries:
      - { type: "gold_pile", amount: "50-100", weight: 10 }
      - { type: "health_potion", weight: 5 }
```

---

## Relationships

### Entity Relationship Diagram

```
Config
  ├── generates → Artifact
  │   ├── ADG (Graph)
  │   │   ├── contains → Rooms
  │   │   ├── contains → Connectors
  │   │   └── validates → Constraints
  │   ├── Layout
  │   │   ├── maps → Rooms to Poses
  │   │   └── maps → Connectors to Paths
  │   ├── TileMap
  │   │   ├── derived from → Layout
  │   │   └── uses → ThemePack
  │   ├── Content
  │   │   ├── Spawns (reference Rooms)
  │   │   ├── Loot (reference Rooms)
  │   │   └── Puzzles (reference Rooms)
  │   └── Metrics
  │       └── measures → Graph properties
  └── uses → ThemePack(s)
```

### Key Dependencies

- **Config → Graph**: Configuration drives graph synthesis
- **Graph → Layout**: ADG must exist before spatial embedding
- **Layout → TileMap**: Poses determine tile placement
- **Graph + TileMap → Content**: Content uses both structure and space
- **Artifact → ValidationReport**: Complete artifact is validated
- **ThemePack → TileMap + Content**: Themes provide visual/gameplay data

---

## Serialization Formats

### JSON (Primary)

All data structures implement `json.Marshaler` and `json.Unmarshaler`.

**Use Cases:**
- Configuration files (human-editable)
- Artifact export (complete dungeon data)
- Validation reports
- Debug artifacts

### YAML (Configuration Only)

Config and ThemePack use YAML for human-friendly authoring.

**Use Cases:**
- Dungeon generation configs
- Theme pack definitions

### TMJ (Tile Map JSON)

TileMap exports to Tiled TMJ format for integration with game engines.

**Layers Exported:**
1. floor (tile layer)
2. walls (tile layer)
3. doors (tile layer)
4. decor (tile layer)
5. entities (object layer - spawns, loot)
6. triggers (object layer - puzzles, secrets)
7. collision (object layer - navigation data)

### SVG (Visualization)

Debug artifacts include SVG visualizations:
- ADG graph (nodes colored by archetype, edges by type)
- Layout heatmap (difficulty zones, connectivity)

---

## Extensibility

### Custom Room Archetypes

Add new archetypes by extending the enum and theme pack definitions. Example: `Workshop`, `Library`, `Throne Room`.

### Custom Connectors

Extend connector types for game-specific mechanics. Example: `Zipline`, `WaterCurrent`, `Minecart`.

### Custom Constraints

Implement new DSL predicates in `pkg/graph/constraint.go`. Example: `minDistance(archetype1, archetype2, dist)` to enforce spatial separation.

### Custom Content

Theme packs support arbitrary encounter/loot types via string references. Game engine resolves these at runtime.

---

## Validation Summary

All entities include validation to ensure:
- IDs are unique and non-empty
- References (room IDs, connector IDs) resolve correctly
- Numeric fields are within specified bounds
- Enums are valid values
- Required fields are present
- Relationships are consistent (e.g., adjacency matches connectors)

Validation occurs at:
1. **Parse time**: JSON/YAML deserialization
2. **Construction time**: Builder pattern validation
3. **Generation time**: Constraint checking
4. **Export time**: Final artifact validation

---

**Status**: Data model complete. All entities defined with fields, validation rules, and relationships. Ready for contract generation and quickstart documentation.
