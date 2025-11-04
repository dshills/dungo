# Research: Graph-Based Dungeon Generator

**Feature**: 001-dungeon-generator-core
**Date**: 2025-11-04
**Status**: Complete

## Overview

This document resolves technical uncertainties identified in the implementation plan, specifically addressing library choices for SVG generation, property-based testing, and Tiled TMJ export.

---

## Decision 1: SVG Generation Library

### Decision: github.com/ajstarks/svgo

### Rationale:
- **Pure Go**: Zero cgo dependencies, essential for cross-platform compatibility
- **Lightweight**: Single-purpose library with no bloat, only what we need
- **Complete SVG 1.1 support**: All primitives needed for debug visualizations (shapes, paths, gradients, groups, filters)
- **Active maintenance**: Recent commits in 2024, responsive to issues
- **Perfect control level**: Low-level API provides pixel-perfect control for custom dungeon visualization
- **Flexible output**: Writes to any `io.Writer` (files, HTTP, stdout)
- **Proven in production**: 2,200+ stars, used by many projects

### Alternatives Considered:

**gonum.org/v1/plot** - Rejected because:
- Designed for statistical plots, not spatial graphs
- Heavier abstraction makes simple room-corridor drawings complex
- Font embedding complexity adds unnecessary overhead
- Overkill for dungeon visualization needs

**github.com/llgcode/draw2d** - Rejected because:
- Has cgo dependencies via freetype-go (violates pure Go requirement)
- Heavier weight with unused features (Cairo API, OpenGL)
- Less active maintenance
- More complex API for simple SVG generation

### Implementation Notes:

**Key Features for Dungeon Visualization:**
- Node-edge graphs: rooms (circles/rectangles), connections (lines with styles)
- Color-coding: room types (start=green, boss=red, treasure=orange, etc.)
- Heatmaps: semi-transparent rectangles for difficulty/density zones
- Annotations: labels, legends, debug info
- Layers: use SVG groups (`<g>`) for organizing rooms, corridors, overlays

**Code Pattern:**
```go
import svg "github.com/ajstarks/svgo"

canvas := svg.New(outputFile)
canvas.Start(width, height)

// Background
canvas.Rect(0, 0, width, height, "fill:#1a1a1a")

// Edges (draw first, below nodes)
canvas.Line(x1, y1, x2, y2, "stroke:#666666;stroke-width:2")

// Nodes (draw last, on top)
canvas.Circle(x, y, radius, "fill:#00ff00;stroke:#ffffff;stroke-width:2")

canvas.End()
```

**Gotchas:**
- SVG Y-axis increases downward (top-left origin)
- String-based styling (consider helper functions for type safety)
- Z-order is paint order (edges before nodes so nodes appear on top)
- Text anchor is bottom-left by default (use `text-anchor:middle` for centered text)

**Installation:**
```bash
go get github.com/ajstarks/svgo
```

---

## Decision 2: Property-Based Testing Framework

### Decision: pgregory.net/rapid

### Rationale:
- **Active maintenance**: Latest release v1.2.0 (February 2025)
- **Automatic shrinking**: Reduces failing test cases to minimal examples (critical for debugging complex graph constraints)
- **Modern Go design**: Uses generics for type-safe generators
- **Zero external dependencies**: Pure Go, no overhead
- **Perfect for graph properties**: Ideal for testing connectivity, reachability, bounds, constraint satisfaction
- **7,000+ dependent projects**: Battle-tested in production

### Alternatives Considered:

**github.com/leanovate/gopter** (Strong alternative):
- Mature (v0.2.8, 2020) but less actively maintained
- QuickCheck/ScalaCheck-inspired functional API
- Rich generator ecosystem
- Good for teams familiar with functional property-based testing patterns
- **Why rapid won**: More active maintenance, modern API, better shrinking

**testing/quick** (Standard library):
- Feature-frozen, no new development
- No automatic shrinking (huge disadvantage)
- Good for simple checks only, not suitable for complex constraint testing
- **Why rapid won**: Shrinking is essential for dungeon graph debugging

### Implementation Notes:

**Property Test Categories:**

1. **Graph Structure Properties:**
   - Connectivity: All rooms reachable from Start
   - Degree bounds: Average branching factor within configured range
   - Cycle presence: At least min cycles, at most max cycles
   - Path bounds: Start→Boss path within minLen..maxLen

2. **Constraint Properties:**
   - Key-before-lock: For each lock L, path exists Start→Key(L)→...→L
   - Room count: Generated dungeons have roomsMin ≤ N ≤ roomsMax
   - Required archetypes: Start and Boss always present
   - No orphans: All rooms reachable (except explicitly disconnected teleport motifs)

3. **Determinism Properties:**
   - Identical seeds produce byte-for-byte identical output
   - Same configuration with different seeds produces structurally similar dungeons

4. **Spatial Properties:**
   - No room overlaps: All bounding boxes non-intersecting
   - Corridor feasibility: All corridors within length and bend limits
   - Grid validity: All coordinates within dungeon bounds

**Code Pattern:**
```go
import "pgregory.net/rapid"

func TestGraphConnectivity(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random dungeon configuration
        seed := rapid.Uint64().Draw(t, "seed")
        roomCount := rapid.IntRange(10, 100).Draw(t, "roomCount")

        cfg := generateConfig(seed, roomCount)
        dungeon, err := generator.Generate(context.Background(), cfg)

        if err != nil {
            t.Fatalf("generation failed: %v", err)
        }

        // Property: All rooms reachable from Start
        startRoom := findRoom(dungeon, RoomTypeStart)
        reachable := bfs(dungeon.Graph, startRoom.ID)

        if len(reachable) != roomCount {
            t.Fatalf("not all rooms reachable: got %d, want %d",
                len(reachable), roomCount)
        }
    })
}
```

**Test Organization:**
```
pkg/graph/graph_test.go         # Graph structure property tests
pkg/synthesis/synthesis_test.go  # Synthesis constraint tests
pkg/embedding/embedding_test.go  # Spatial property tests
pkg/validation/validation_test.go # Validation correctness tests
testdata/properties/             # Shared property test fixtures
```

**Installation:**
```bash
go get pgregory.net/rapid
```

**CI Integration:**
```bash
# Run property tests with extended checks
go test -v -rapidchecks=1000 ./...

# Run with seed for reproducibility
go test -v -rapidseed=12345 ./pkg/graph
```

---

## Decision 3: Tiled TMJ Export Approach

### Decision: Custom Implementation using encoding/json

### Rationale:
- **No existing Go libraries for TMJ export**: All available libraries focus on TMX (XML) import, not TMJ (JSON) export
- **Standard library sufficient**: `encoding/json` provides everything needed
- **Full control**: Custom implementation allows optimization for dungeon-specific needs
- **No external dependencies**: Aligns with project philosophy
- **Minimal code**: Approximately 500 lines for complete implementation
- **Future-proof**: Can add features as needed without library constraints

### Alternatives Considered:

**Existing Libraries** (All rejected):
- `github.com/lafriks/go-tiled`: TMX only, focused on import/rendering
- `github.com/salviati/go-tmx`: TMX parsing only
- `github.com/fardog/tmx`: Basic TMX parser
- None support TMJ (JSON) export

**Why Custom Implementation**:
- TMJ format is straightforward JSON structure
- Dungeon generator already has all spatial data needed
- Direct control over layer organization and optimization
- Can add compression (gzip) and validation as needed

### TMJ Format Overview:

**Structure:**
```json
{
  "type": "map",
  "version": "1.10",
  "tiledversion": "1.11.0",
  "width": 100,
  "height": 100,
  "tilewidth": 32,
  "tileheight": 32,
  "orientation": "orthogonal",
  "layers": [ /* tile layers, object layers */ ],
  "tilesets": [ /* tileset definitions */ ],
  "properties": [ /* custom metadata */ ]
}
```

**Layer Types:**
1. **Tile Layer**: Grid-based tiles (floors, walls)
2. **Object Layer**: Entities, triggers, spawn points
3. **Image Layer**: Background images
4. **Group Layer**: Hierarchical organization

**Recommended Layer Structure for Dungeons:**
1. floor - Base terrain tiles
2. floor_decor - Carpet, cracks, details
3. walls - Wall tiles
4. wall_decor - Banners, sconces, decorations
5. doors - Door tiles (interactive)
6. furniture - Tables, chests, props
7. triggers - Invisible zones (object layer)
8. hazards - Traps, spikes (object layer)
9. entities - Spawn points (object layer)
10. collision - Collision shapes (object layer)
11. lighting - Light sources (object layer)
12. overlay - Fog of war, discovered areas

### Implementation Structure:

**Core Types:**
```go
type Map struct {
    Type          string     `json:"type"`
    Version       string     `json:"version"`
    TiledVersion  string     `json:"tiledversion"`
    Width         int        `json:"width"`
    Height        int        `json:"height"`
    TileWidth     int        `json:"tilewidth"`
    TileHeight    int        `json:"tileheight"`
    Orientation   string     `json:"orientation"`
    RenderOrder   string     `json:"renderorder"`
    Layers        []Layer    `json:"layers"`
    Tilesets      []Tileset  `json:"tilesets"`
    Properties    []Property `json:"properties,omitempty"`
}

type Layer struct {
    ID         int        `json:"id"`
    Name       string     `json:"name"`
    Type       string     `json:"type"` // "tilelayer", "objectgroup"
    Visible    bool       `json:"visible"`
    Opacity    float64    `json:"opacity"`
    X          int        `json:"x"`
    Y          int        `json:"y"`
    Width      int        `json:"width"`
    Height     int        `json:"height"`
    Data       []uint32   `json:"data,omitempty"`      // For tile layers
    Objects    []Object   `json:"objects,omitempty"`   // For object layers
    Encoding   string     `json:"encoding,omitempty"`  // "csv", "base64"
    Compression string    `json:"compression,omitempty"` // "", "gzip", "zlib"
}

type Object struct {
    ID       int        `json:"id"`
    Name     string     `json:"name"`
    Type     string     `json:"type"`
    X        float64    `json:"x"`
    Y        float64    `json:"y"`
    Width    float64    `json:"width"`
    Height   float64    `json:"height"`
    Rotation float64    `json:"rotation"`
    GID      uint32     `json:"gid,omitempty"`      // For tile objects
    Visible  bool       `json:"visible"`
    Properties []Property `json:"properties,omitempty"`
}
```

**Builder Pattern:**
```go
builder := NewMapBuilder(width, height, 32, 32)

// Add tileset
builder.AddTileset("dungeon_tiles", "themes/crypt/tiles.png", 32, 32, 256)

// Add tile layers
floorLayer := builder.AddTileLayer("floor", width, height)
floorLayer.SetTile(x, y, tileID)

// Add object layer
entityLayer := builder.AddObjectLayer("entities")
entityLayer.AddObject("spawn_player", "SpawnPoint", x, y, 32, 32, nil)

// Export
tmjData, err := builder.Build()
json.NewEncoder(file).Encode(tmjData)
```

### Best Practices:

**Tileset Organization:**
- Use single 256x256 tileset atlas per theme (8x8 grid of 32x32 tiles = 64 tiles)
- Organize by category: floors (0-15), walls (16-31), doors (32-39), decor (40-63)
- First tile (ID 0) is always empty/transparent

**GID Calculation:**
- GID = tileset.firstgid + local_tile_id
- GID 0 means empty cell
- Bitwise flags for flips: horizontal (0x80000000), vertical (0x40000000), diagonal (0x20000000)

**Compression:**
- Small dungeons (<50x50): Use "csv" encoding for human readability
- Large dungeons (>100x100): Use "gzip" compression
- Always provide uncompressed version for debugging

**Validation:**
- Verify all tile IDs are within tileset bounds
- Check object coordinates are within map bounds
- Validate layer dimensions match map dimensions
- Ensure GIDs reference existing tilesets

**Performance:**
- Pre-allocate layer data arrays: `make([]uint32, width*height)`
- Use efficient tile indexing: `data[y*width + x]`
- Batch tile updates before export
- Consider streaming for very large dungeons

### Implementation Phases:

**Phase 1: Core Export (MVP)**
- Map structure with basic properties
- Tile layer export with CSV encoding
- Object layer export for entities
- Single tileset support

**Phase 2: Advanced Features**
- Gzip compression for large maps
- Multiple tileset support
- Custom property types
- Layer groups for organization

**Phase 3: Integration**
- Convert dungeon Artifact to TMJ
- Map room archetypes to tile patterns
- Export content placement as objects
- Generate metadata properties

**Phase 4: Enhancement**
- Import capability for round-trip editing
- Optimization (RLE, delta compression)
- Validation framework
- CLI tool for batch export

### Installation:
```bash
# No external dependencies needed
# Use standard library only:
import "encoding/json"
import "compress/gzip"
```

### References:
- Tiled TMJ Format Spec: https://doc.mapeditor.org/en/stable/reference/json-map-format/
- Tiled Editor: https://www.mapeditor.org/

---

## Decision 4: Golden Test Framework

### Decision: Use Standard Library with testdata/

### Rationale:
- Go's standard library testing is sufficient for golden tests
- Simple pattern: generate output, compare to snapshot
- No external dependencies needed
- Integrates with existing test infrastructure

### Pattern:
```go
func TestGoldenDungeon(t *testing.T) {
    seed := uint64(12345)
    cfg := loadConfig("testdata/configs/small_crypt.yaml")

    artifact, err := generator.Generate(context.Background(), cfg)
    if err != nil {
        t.Fatal(err)
    }

    // Export to JSON
    output, err := json.MarshalIndent(artifact, "", "  ")
    if err != nil {
        t.Fatal(err)
    }

    goldenPath := "testdata/golden/small_crypt_12345.json"

    if *update {
        // Update golden file
        os.WriteFile(goldenPath, output, 0644)
        return
    }

    // Compare to golden
    expected, err := os.ReadFile(goldenPath)
    if err != nil {
        t.Fatal(err)
    }

    if !bytes.Equal(output, expected) {
        t.Errorf("output differs from golden file\nRun with -update to update golden files")
    }
}
```

### Directory Structure:
```
testdata/
├── configs/           # Test configurations
│   ├── small_crypt.yaml
│   ├── large_fungal.yaml
│   └── dual_biome.yaml
├── golden/            # Expected outputs
│   ├── small_crypt_12345.json
│   ├── small_crypt_12345.svg
│   └── large_fungal_99999.json
└── schemas/           # JSON schemas for validation
    ├── artifact_v1.schema.json
    └── config_v1.schema.json
```

---

## Summary of Resolved Uncertainties

| Uncertainty | Decision | Library/Approach |
|-------------|----------|------------------|
| SVG generation | Use github.com/ajstarks/svgo | Pure Go, active, complete features |
| Property testing | Use pgregory.net/rapid | Modern, automatic shrinking, perfect for graphs |
| Tiled TMJ export | Custom implementation | Standard library only, full control |
| Golden testing | Standard library pattern | No external dependencies |

All technical uncertainties resolved. Ready to proceed to Phase 1 design.

---

## Additional Research Notes

### Performance Considerations:
- RNG must use crypto/sha256 for deterministic sub-seed derivation
- Consider sync.Pool for frequently allocated structures (rooms, edges)
- Profile early with `go test -bench -memprofile`
- Target <50ms for graph+embedding, <200ms total

### Testing Strategy:
- Unit tests: 100% coverage for core types and pure functions
- Property tests: Focus on constraint satisfaction across random inputs
- Golden tests: Snapshot 5-10 fixed seeds with different configurations
- Integration tests: Simulated agent verifying boss/key findability
- Fuzz tests: Push boundaries (0 rooms, 1000 rooms, conflicting constraints)

### CI/CD Integration:
```bash
# Pre-commit hooks
golangci-lint run
go test ./...
gofmt -l .

# CI pipeline
go test -v -race ./...
go test -bench ./... -benchmem
go test -rapidchecks=10000 ./... # Extended property checks
```

---

**Status**: All research complete. No outstanding clarifications. Ready for Phase 1 design (data-model.md, contracts/, quickstart.md).
