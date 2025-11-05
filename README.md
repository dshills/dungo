# Dungo

**A deterministic, graph-based procedural dungeon generator for Go**

[![Go Version](https://img.shields.io/badge/Go-1.25.3-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Dungo is a sophisticated procedural dungeon generation library that creates playable, well-paced dungeons by synthesizing abstract graphs, embedding them spatially, and populating them with gameplay content. Built for game developers who need deterministic, configurable, and high-quality procedural content.

---

## Features

### Core Capabilities

- **Deterministic Generation**: Same seed always produces identical output, enabling caching, debugging, and reproducible designs
- **Graph-Based Architecture**: Separates logical structure from spatial layout for flexible, constraint-driven generation
- **Five-Stage Pipeline**: Graph synthesis → Spatial embedding → Tile carving → Content population → Validation
- **Rich Configuration**: YAML-based configs with extensive control over size, branching, pacing, themes, and constraints
- **Multiple Export Formats**: JSON (data), TMJ (Tiled map), SVG (visualization)

### Game Design Features

- **Pacing Curves**: Linear, S-curve, exponential, or custom difficulty progression from start to boss
- **Lock-and-Key Puzzles**: Automatic key placement with constraint solving ensures reachability
- **Secret Rooms**: Hidden areas with configurable density and reward scaling
- **Optional Content**: Side paths and optional rooms for exploration
- **Thematic Coherence**: Multi-theme support with biome clustering and content tables

### Technical Features

- **Constraint Validation**: Hard constraints (connectivity, key-before-lock) and soft constraints (pacing adherence)
- **Performance**: Generates 60-room dungeons in <50ms (graph + embedding), <200ms total
- **Extensible**: Plugin system for custom graph synthesizers, embedders, carvers, and content generators
- **Battle-Tested**: Comprehensive test suite with unit, property, golden, and agent-based validation tests

---

## Quick Start

### Installation

```bash
go get github.com/dshills/dungo
```

### CLI Tool

Install the command-line generator:

```bash
go install github.com/dshills/dungo/cmd/dungeongen@latest
```

Generate your first dungeon:

```bash
# Create a config file
cat > dungeon.yaml <<EOF
seed: 42
size:
  roomsMin: 20
  roomsMax: 30
branching:
  avg: 2.0
  max: 4
pacing:
  curve: "S_CURVE"
  variance: 0.15
themes:
  - crypt
keys:
  - name: silver
    count: 1
secretDensity: 0.15
optionalRatio: 0.25
EOF

# Generate with all export formats
dungeongen -config dungeon.yaml -format all -output ./output
```

This creates:
- `dungeon.json` - Full artifact data
- `dungeon.tmj` - Tiled map editor format
- `dungeon.svg` - Visual graph representation

### Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dshills/dungo/pkg/dungeon"
    "github.com/dshills/dungo/pkg/validation"
)

func main() {
    // Load configuration
    cfg, err := dungeon.LoadConfig("dungeon.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Create generator with validator
    validator := validation.NewValidator()
    gen := dungeon.NewGeneratorWithValidator(validator)

    // Generate dungeon
    artifact, err := gen.Generate(context.Background(), cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Access generated data
    fmt.Printf("Generated %d rooms\n", len(artifact.ADG.Graph.Rooms))
    fmt.Printf("Path to boss: %d rooms\n", artifact.Metrics.PathLength)
    fmt.Printf("Branching factor: %.2f\n", artifact.Metrics.BranchingFactor)

    // Check validation
    if artifact.Debug.Report.Passed {
        fmt.Println("✓ All constraints satisfied")
    }

    // Access specific data
    for _, room := range artifact.ADG.Graph.Rooms {
        fmt.Printf("Room %s: %s (difficulty: %.2f)\n",
            room.GetID(), room.GetArchetype(), room.GetDifficulty())
    }
}
```

---

## Configuration

Dungeons are configured via YAML files with extensive control over generation parameters.

### Basic Configuration

```yaml
seed: 12345              # For deterministic generation
size:
  roomsMin: 25           # Minimum room count
  roomsMax: 35           # Maximum room count
branching:
  avg: 1.7               # Average connections per room
  max: 3                 # Maximum connections per room
pacing:
  curve: "S_CURVE"       # LINEAR, S_CURVE, EXPONENTIAL, or CUSTOM
  variance: 0.15         # Difficulty variance (0.0-1.0)
themes:
  - crypt                # Theme for rooms and content
secretDensity: 0.15      # Proportion of secret rooms (0.0-1.0)
optionalRatio: 0.20      # Proportion of optional content (0.0-1.0)
```

### Keys and Locks

```yaml
keys:
  - name: silver         # Key identifier
    count: 1             # Number of keys to place
  - name: gold
    count: 1
```

Keys automatically generate lock gates that must be opened to progress. The system ensures keys are always reachable before their locks.

### Constraints

```yaml
constraints:
  - kind: Connectivity
    severity: hard       # hard = must pass, soft = optimize
    expr: "isConnected()"
  - kind: KeyLock
    severity: hard
    expr: "keyBeforeLock('silver')"
  - kind: Pacing
    severity: soft
    expr: "monotoneIncrease(start, boss, 0.8)"
```

### Pacing Curves

Control difficulty progression from start to boss:

- **LINEAR**: Steady linear increase
- **S_CURVE**: Easy start, steep middle, plateau at end (recommended)
- **EXPONENTIAL**: Exponential difficulty ramp
- **CUSTOM**: Define your own curve with control points

```yaml
pacing:
  curve: "CUSTOM"
  customCurve:
    - [0.0, 0.0]     # [progress, difficulty]
    - [0.3, 0.2]
    - [0.7, 0.7]
    - [1.0, 1.0]
  variance: 0.15
```

### Themes

Themes control visual style and content tables:

```yaml
themes:
  - crypt              # Undead, skeletons, dark atmosphere
  - fungal             # Spore clouds, mushroom creatures
  - arcane             # Magic traps, elementals, crystals
```

See [`pkg/themes/builtin/`](pkg/themes/builtin/) for available themes and content tables.

---

## Architecture

Dungo uses a five-stage deterministic pipeline:

```
┌──────────────────────────────────────────────────────────────┐
│ Stage 1: Graph Synthesis                                     │
│ Input:  Seed, Config, Constraints                            │
│ Output: Abstract Dungeon Graph (ADG)                         │
│ - Room nodes (Start, Boss, Treasure, Puzzle, etc.)          │
│ - Connectors (Door, Corridor, OneWay, Hidden)               │
│ - Constraint annotations (key-before-lock, pacing)          │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│ Stage 2: Spatial Embedding                                   │
│ Input:  ADG, Spatial preferences                             │
│ Output: Layout (room positions, corridor paths)             │
│ - Force-directed graph embedding                            │
│ - Room placement with overlap prevention                    │
│ - Corridor pathfinding (Manhattan, A*)                      │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│ Stage 3: Tile Carving                                        │
│ Input:  Layout, Room templates                              │
│ Output: TileMap (walls, floors, doors)                      │
│ - Room stamping (rectangles, ovals, L-shapes)               │
│ - Corridor carving with door placement                      │
│ - Wall generation around carved areas                       │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│ Stage 4: Content Population                                  │
│ Input:  TileMap, ADG, Theme tables                          │
│ Output: Content (spawns, loot, puzzles)                     │
│ - Enemy placement based on difficulty                       │
│ - Loot distribution with reward scaling                     │
│ - Puzzle and secret content injection                       │
└──────────────────────────────────────────────────────────────┘
                            ↓
┌──────────────────────────────────────────────────────────────┐
│ Stage 5: Validation & Metrics                                │
│ Input:  Complete Artifact                                    │
│ Output: Validation report, Metrics                           │
│ - Hard constraint verification                               │
│ - Soft constraint scoring                                    │
│ - Agent-based pathfinding tests                             │
│ - Difficulty curve analysis                                  │
└──────────────────────────────────────────────────────────────┘
```

Each stage is pure and deterministic given its inputs. Sub-seeds are derived from the master seed for each stage, ensuring reproducibility.

### Package Structure

```
dungo/
├── cmd/
│   └── dungeongen/        # CLI tool for generation
├── pkg/
│   ├── dungeon/           # Core generator interface and config
│   ├── graph/             # Abstract Dungeon Graph (ADG)
│   ├── synthesis/         # Graph synthesis strategies
│   ├── embedding/         # Spatial layout algorithms
│   ├── carving/           # Tile map generation
│   ├── content/           # Content placement
│   ├── validation/        # Constraint validation
│   ├── export/            # Export to JSON, TMJ, SVG
│   ├── themes/            # Theme system and content tables
│   └── rng/               # Deterministic RNG with sub-seeds
├── examples/              # Working code examples
├── testdata/              # Test configurations and golden files
└── specs/                 # Design specifications
```

---

## Examples

The [`examples/`](examples/) directory contains working demonstrations:

### 1. Quick Start

**Path**: `examples/quickstart/main.go`

Minimal example showing the basic generation workflow.

```bash
go run examples/quickstart/main.go
```

### 2. Text Rendering

**Path**: `examples/text-render/main.go`

Terminal visualization with ASCII art and emoji room icons.

```bash
go run examples/text-render/main.go
```

Shows:
- Room statistics and graph structure
- Emoji-based room type visualization (📍 Start, 💀 Boss, 💎 Treasure)
- ASCII tile map preview
- Content summary (enemies, loot, keys)
- Key-lock relationships
- Validation status

### 3. Spatial Embedding

**Path**: `examples/embedding/main.go`

Demonstrates the force-directed embedding algorithm in detail.

```bash
go run examples/embedding/main.go
```

Shows:
- Graph creation with rooms and connections
- Spatial layout generation with positions
- Room dimensions and overlap prevention
- Embedding quality metrics

### More Examples

See [`examples/README.md`](examples/README.md) for complete list and sample configurations.

---

## Development

### Prerequisites

- Go 1.25.3 or later
- Make (optional, for convenience commands)

### Building

```bash
# Build all packages
go build ./...

# Build CLI tool
go build -o bin/dungeongen cmd/dungeongen/main.go

# Install CLI to $GOPATH/bin
go install ./cmd/dungeongen
```

### Project Setup

```bash
# Clone repository
git clone https://github.com/dshills/dungo.git
cd dungo

# Install dependencies
go mod download

# Run tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

The project uses strict code quality standards:

```bash
# Run linter
golangci-lint run

# Format code
gofmt -w .

# Run pre-commit checks (recommended)
# Uses mcp-pr for automated code review
git add <files>
# Run code review in Claude Code or configure as pre-commit hook
```

See [`.golangci.yml`](.golangci.yml) for linter configuration.

---

## Testing

Dungo has a comprehensive test suite with multiple testing strategies.

### Test Types

#### 1. Unit Tests

```bash
# Run all tests
go test ./...

# Run specific package
go test ./pkg/graph
go test ./pkg/synthesis

# Verbose output
go test -v ./...
```

#### 2. Property Tests

Uses `pgregory.net/rapid` for randomized property testing:

```bash
# Run property tests
go test ./pkg/dungeon -run Property

# With more iterations
go test ./pkg/dungeon -run Property -rapid.checks=1000
```

#### 3. Golden Tests

Snapshot-based tests that detect regressions:

```bash
go test ./pkg/dungeon -run Golden
```

Golden files are stored in `testdata/` with SVG visualizations for manual inspection.

#### 4. Integration Tests

Full pipeline tests with validation:

```bash
go test ./test/integration
```

#### 5. Benchmarks

Performance benchmarks for each pipeline stage:

```bash
# Run all benchmarks
go test -bench=. ./...

# Specific benchmark
go test -bench=BenchmarkFullPipeline ./pkg/dungeon

# With memory stats
go test -bench=. -benchmem ./pkg/dungeon
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# Coverage summary
go test -cover ./...
```

Target: >80% coverage for core packages.

---

## Documentation

### Specifications

The [`specs/`](specs/) directory contains detailed design documentation:

- **[`dungo_specification.md`](specs/dungo_specification.md)**: Complete technical specification
- **[`001-dungeon-generator-core/`](specs/001-dungeon-generator-core/)**: Implementation contracts and quickstart guide

### Developer Guide

[`CLAUDE.md`](CLAUDE.md) provides guidance for contributors and AI-assisted development, including:

- Architecture overview
- Development commands
- Design constraints and patterns
- Implementation lessons learned
- Code style guidelines

### API Documentation

```bash
# Generate and view Go documentation
go doc github.com/dshills/dungo/pkg/dungeon
go doc github.com/dshills/dungo/pkg/graph

# Start local documentation server
godoc -http=:6060
# Visit http://localhost:6060/pkg/github.com/dshills/dungo/
```

---

## Performance

Dungo is optimized for fast generation with minimal memory usage.

### Benchmarks

Typical performance on modern hardware (M1 Mac):

| Dungeon Size | Rooms | Stage 1 (Graph) | Stage 2 (Embed) | Stage 3 (Carve) | Stage 4 (Content) | Total  |
|--------------|-------|-----------------|-----------------|-----------------|-------------------|--------|
| Small        | 10    | ~1ms            | ~2ms            | ~1ms            | ~1ms              | ~5ms   |
| Medium       | 30    | ~3ms            | ~8ms            | ~3ms            | ~2ms              | ~16ms  |
| Large        | 60    | ~8ms            | ~25ms           | ~8ms            | ~5ms              | ~46ms  |
| Very Large   | 100   | ~15ms           | ~60ms           | ~15ms           | ~10ms             | ~100ms |

Memory usage: <50MB per generation for typical dungeons.

### Optimization Tips

1. **Reuse Generators**: Create generator once, call `Generate()` multiple times
2. **Parallel Generation**: Generators are stateless - safe for concurrent use
3. **Cache Results**: Same seed always produces same output - cache by seed
4. **Limit Iterations**: Reduce embedding iterations for prototyping (trade quality for speed)

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Process

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes with tests
4. **Run** tests and lints: `go test ./... && golangci-lint run`
5. **Commit** with descriptive messages
6. **Push** to your fork
7. **Open** a Pull Request

### Code Standards

- Write tests for new features (aim for >80% coverage)
- Follow Go conventions and style guidelines
- Add documentation comments for exported types and functions
- Run `gofmt` before committing
- Keep cyclomatic complexity reasonable (< 15 per function)
- Add examples for new major features

### Areas for Contribution

- **New Themes**: Add theme packs with content tables
- **Graph Synthesizers**: Alternative synthesis algorithms (template stitching, constraint solvers)
- **Embedders**: New spatial layout algorithms (orthogonal, packing-based)
- **Content Generators**: Specialized content for puzzle types, boss encounters, etc.
- **Export Formats**: Additional export targets (Unity, Godot, custom engines)
- **Visualization**: Better debugging visualizations and analysis tools

See open issues for specific feature requests and bugs.

---

## Roadmap

### Version 1.1 (Current)

- ✅ Core five-stage pipeline
- ✅ Grammar-based graph synthesis
- ✅ Force-directed spatial embedding
- ✅ Lock-and-key puzzles with constraint solving
- ✅ Multi-format export (JSON, TMJ, SVG)
- ✅ Comprehensive test suite

### Version 1.2 (Planned)

- [ ] Template-based synthesis strategy
- [ ] Orthogonal graph embedding
- [ ] Room shape variants (L-shaped, circular, irregular)
- [ ] Advanced puzzle types (pressure plates, toggles)
- [ ] Multi-level dungeons (stairs, ladders)
- [ ] Additional themes (ice, lava, library, prison)

### Version 2.0 (Future)

- [ ] Real-time adaptive generation during gameplay
- [ ] Player telemetry feedback loops for pacing
- [ ] 3D geometry export
- [ ] Authoring tools and visual editor
- [ ] Procedural narrative elements

See [GitHub Issues](https://github.com/dshills/dungo/issues) for detailed planning.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

### Inspiration

- **Spelunky** by Derek Yu - Procedural generation with handcrafted feel
- **Binding of Isaac** - Room-based dungeon structure
- **Brogue** - Elegant constraint-based level generation
- **Diablo series** - Tileset-based procedural dungeons

### Research Papers

- Dormans, J. (2010). "Adventures in level design: Generating missions and spaces for action adventure games"
- Shaker, N., Togelius, J., & Nelson, M. J. (2016). "Procedural Content Generation in Games"
- Smith, G., Whitehead, J., & Mateas, M. (2011). "Tanagra: Reactive planning and constraint solving for mixed-initiative level design"

### Libraries

- [svgo](https://github.com/ajstarks/svgo) - SVG generation
- [yaml.v3](https://github.com/go-yaml/yaml) - YAML configuration parsing
- [rapid](https://pgregory.net/rapid) - Property-based testing

---

## Support

- **Documentation**: See [`specs/`](specs/) and [`CLAUDE.md`](CLAUDE.md)
- **Examples**: Check [`examples/`](examples/) directory
- **Issues**: [GitHub Issues](https://github.com/dshills/dungo/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dshills/dungo/discussions)

---

**Built with ❤️ for game developers who love procedural content generation**
