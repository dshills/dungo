# Embedding Package

The `embedding` package provides spatial layout algorithms for transforming Abstract Dungeon Graphs (ADG) into 2D spatial layouts. This is the **SECOND stage** of the dungeon generation pipeline.

## Overview

The embedding stage takes a logical graph of rooms and connections (from the synthesis stage) and assigns concrete spatial coordinates to all elements. It ensures:

- **No overlaps**: Room bounding boxes do not intersect
- **Feasible corridors**: All corridors fit within length and bend constraints
- **Minimum spacing**: Rooms maintain required separation
- **Determinism**: Same seed produces identical layouts

## Architecture

### Core Types

#### Layout
The complete spatial embedding result:
```go
type Layout struct {
    Poses         map[string]*Pose  // Room ID → spatial placement
    CorridorPaths map[string]*Path  // Connector ID → corridor route
    Bounds        Rect              // Overall bounding box
    Seed          uint64            // RNG seed used
    Algorithm     string            // Embedder used
}
```

#### Pose
Spatial placement of a room:
```go
type Pose struct {
    X, Y        float64  // Grid position
    Width       int      // Bounding box width
    Height      int      // Bounding box height
    Rotation    int      // Rotation in degrees (0, 90, 180, 270)
    FootprintID string   // Template identifier
}
```

#### Path
Corridor route between rooms:
```go
type Path struct {
    Points        []Point  // Polyline points
    DoorPositions []int    // Where to place doors
}
```

### Embedder Interface

All embedding algorithms implement:
```go
type Embedder interface {
    Embed(g *graph.Graph, rng *rng.RNG) (*Layout, error)
    Name() string
}
```

## Force-Directed Embedder

The default embedder uses a force-directed layout algorithm inspired by graph visualization techniques.

### Algorithm Steps

1. **Initialize**: Place rooms at random positions in a circle
2. **Simulate Forces**: Iteratively apply physics-based forces:
   - **Spring forces**: Connected rooms attract each other
   - **Repulsion forces**: All rooms repel each other
3. **Stabilize**: Continue until movement is minimal or max iterations reached
4. **Quantize**: Snap positions to grid coordinates
5. **Resolve Overlaps**: Iteratively separate overlapping rooms
6. **Route Corridors**: Create Manhattan paths between connected rooms
7. **Validate**: Check all constraints are satisfied

### Force Model

**Spring Force (Attraction)**:
```
F = k_spring * distance
```
Pulls connected rooms together based on their separation.

**Repulsion Force**:
```
F = k_repulsion / distance²
```
Pushes all room pairs apart, stronger when closer.

**Velocity Update**:
```
v = v * damping + F * dt
```
Damping prevents oscillation and helps convergence.

### Configuration

```go
config := &embedding.Config{
    MaxIterations:      500,    // Force simulation limit
    CorridorMaxLength:  50.0,   // Maximum corridor length
    CorridorMaxBends:   4,      // Maximum bends per corridor
    MinRoomSpacing:     2.0,    // Minimum gap between rooms
    GridQuantization:   1.0,    // Grid cell size
    SpringConstant:     0.5,    // Attraction strength
    RepulsionConstant:  500.0,  // Repulsion strength
    DampingFactor:      0.8,    // Movement damping [0-1]
    StabilityThreshold: 0.1,    // Early stop threshold
    InitialSpread:      100.0,  // Initial random radius
}
```

### Room Size Mapping

Abstract room sizes map to grid dimensions:
- **XS** (Tiny corridor): 3x3
- **S** (Small chamber): 5x5
- **M** (Medium hall): 8x8
- **L** (Large room): 12x12
- **XL** (Boss arena): 16x16

## Usage

### Basic Usage

```go
import (
    "github.com/dshills/dungo/pkg/embedding"
    "github.com/dshills/dungo/pkg/graph"
    "github.com/dshills/dungo/pkg/rng"
)

// Create graph (from synthesis stage)
g := graph.NewGraph(12345)
// ... add rooms and connectors ...

// Create RNG for determinism
configHash := []byte("my_config")
rngInstance := rng.NewRNG(12345, "embedding", configHash)

// Get embedder
config := embedding.DefaultConfig()
embedder, err := embedding.Get("force_directed", config)
if err != nil {
    panic(err)
}

// Perform embedding
layout, err := embedder.Embed(g, rngInstance)
if err != nil {
    panic(err)
}

// Access results
for roomID, pose := range layout.Poses {
    fmt.Printf("Room %s at (%.1f, %.1f)\n", roomID, pose.X, pose.Y)
}

// Validate constraints
err = embedding.ValidateEmbedding(layout, g, config)
```

### Custom Configuration

```go
config := &embedding.Config{
    MaxIterations:     1000,  // More iterations for complex graphs
    CorridorMaxLength: 100.0, // Longer corridors allowed
    MinRoomSpacing:    3.0,   // More spacing between rooms
    // ... other settings ...
}

embedder := embedding.NewForceDirectedEmbedder(config)
layout, err := embedder.Embed(g, rng)
```

### Registering Custom Embedders

```go
// Define your embedder
type MyEmbedder struct {
    config *embedding.Config
}

func (e *MyEmbedder) Name() string {
    return "my_algorithm"
}

func (e *MyEmbedder) Embed(g *graph.Graph, rng *rng.RNG) (*embedding.Layout, error) {
    // Your implementation
}

// Register it
func init() {
    embedding.Register("my_algorithm", func(config *embedding.Config) embedding.Embedder {
        return &MyEmbedder{config: config}
    })
}
```

## Validation

The package provides comprehensive validation:

```go
err := embedding.ValidateEmbedding(layout, graph, config)
```

Checks:
- All rooms have poses
- All connectors have paths
- No room overlaps
- Corridors within length limits
- Corridors within bend limits
- Minimum spacing maintained

## Determinism

The embedder is fully deterministic when using the same:
- Graph structure
- RNG seed
- Configuration

This ensures reproducible dungeons across runs.

## Performance

**Time Complexity**: O(N² × I) where:
- N = number of rooms
- I = number of iterations (typically 100-500)

**Space Complexity**: O(N + E) where:
- N = number of rooms
- E = number of connectors

**Typical Performance**:
- 10 rooms: <10ms
- 50 rooms: ~50ms
- 100 rooms: ~200ms

## Future Enhancements

Planned for future versions:
- **Orthogonal Embedder**: Grid-aligned layout with BFS layering
- **Template-Based Embedder**: Pre-designed room arrangements
- **A* Corridor Routing**: Obstacle-avoiding pathfinding
- **3D Support**: Multi-level dungeons with vertical connections
- **Room Rotation**: Non-axis-aligned placements
- **Irregular Shapes**: Non-rectangular room footprints

## Integration

This package integrates with:
- **Input**: `pkg/graph` (Abstract Dungeon Graph)
- **Output**: Used by `pkg/carving` (tile map generation)
- **Utility**: `pkg/rng` (deterministic randomness)

## Testing

The package includes:
- Unit tests for all core types
- Integration tests for complete embedding
- Determinism tests verifying reproducibility
- 84.6% code coverage

Run tests:
```bash
go test ./pkg/embedding/...
```

Run with coverage:
```bash
go test ./pkg/embedding/... -cover
```

## Examples

See `/examples/embedding_example.go` for a complete working example demonstrating:
- Graph creation
- Embedder configuration
- Layout generation
- Result visualization
- Validation

## API Documentation

For detailed API documentation, see:
```bash
go doc github.com/dshills/dungo/pkg/embedding
```

## References

- Force-Directed Graph Drawing: Fruchterman & Reingold (1991)
- Graph Layout Algorithms: Tamassia (2013)
- Procedural Dungeon Generation: Shaker, Togelius & Nelson (2016)
