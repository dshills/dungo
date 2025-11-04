# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**dungo** is a graph-based procedural dungeon generator library written in Go. The system generates playable dungeons by:
1. Synthesizing an abstract graph (ADG) of rooms and constraints
2. Embedding that graph into a spatial layout
3. Emitting tile maps and gameplay content (enemies, loot, locks/keys, secrets)

The project is in early development stage with specifications defined but no implementation yet.

## Architecture

### Pipeline Stages

The generation follows a deterministic pipeline:

```
[Seed, Config, Theme]
    ↓
(A) Graph Synthesis → ADG (rooms/edges + constraints)
    ↓
(B) Spatial Embedding → Room placements, corridors, grid/coord map
    ↓
(C) Carving & Topology → Tile map (walls, floors, doors), portals
    ↓
(D) Content Pass → Enemies, loot, keys/locks, secrets, hazards
    ↓
(E) Validation & Scoring → Metrics, assertions, debug artifacts
```

Each stage is pure and deterministic given its inputs.

### Core Package Structure

Expected package organization (from `specs/dungo_specification.md`):

- `pkg/dungeon` - Core generator interface and config types
- `pkg/graph` - ADG (Abstract Dungeon Graph) data structures
  - Room types: Start, Boss, Treasure, Puzzle, Hub, Corridor, Secret, etc.
  - Connector types: Door, Corridor, Ladder, Teleporter, Hidden, OneWay
  - Constraint system (connectivity, degree bounds, key-before-lock, pacing)
- `pkg/synthesis` - Graph synthesis strategies (grammar-based, template stitching, search/optimize)
- `pkg/embedding` - Spatial layout algorithms (force-directed, orthogonal, packing)
- `pkg/carving` - Tile map generation from spatial layout
- `pkg/content` - Content placement (encounters, loot, puzzles, secrets)
- `pkg/validation` - Constraint validation and metrics
- `pkg/rng` - Deterministic RNG with sub-seed derivation

### Key Interfaces

The specification defines these core interfaces (see `specs/dungo_specification.md` §14.2):

```go
type Generator interface {
    Generate(ctx context.Context, cfg Config) (Artifact, error)
}

type GraphSynthesizer interface {
    Synthesize(ctx context.Context, rng RNG, cfg Config) (Graph, error)
}

type Embedder interface {
    Embed(ctx context.Context, rng RNG, g Graph, cfg Config) (Layout, error)
}

type Carver interface {
    Carve(ctx context.Context, rng RNG, layout Layout, cfg Config) (TileMap, error)
}

type ContentPass interface {
    Populate(ctx context.Context, rng RNG, tm TileMap, g Graph, cfg Config) (Content, error)
}

type Validator interface {
    Validate(ctx context.Context, a Artifact, cfg Config) (ValidationReport, error)
}
```

## Development Commands

### Building
```bash
go build ./...
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./pkg/graph
go test ./pkg/synthesis

# Run with verbose output
go test -v ./...
```

### Code Review (Pre-Commit)
```bash
# REQUIRED: Run mcp-pr code review before committing
# Reviews unstaged changes using OpenAI for quality assurance
mcp__mcp-pr__review_unstaged with OpenAI provider

# Address findings:
# - HIGH severity: MUST fix before commit
# - MEDIUM severity: Document as tech debt if not immediately fixable
# - LOW/INFO: Optional improvements
```

Per constitution v1.1.0, code review is mandatory before all commits.

### Linting
```bash
# Run golangci-lint (configuration in .golangci.yml)
golangci-lint run

# Run with specific linters
golangci-lint run --enable-all

# Auto-fix issues where possible
golangci-lint run --fix
```

Configured linters: gofmt, govet, staticcheck, errcheck, gosimple, ineffassign, unused, typecheck, gocyclo, misspell

### Formatting
```bash
# Format all Go files
gofmt -w .

# Check formatting without modifying files
gofmt -l .
```

## Design Constraints

### Determinism (Critical)
- **Single seed must produce identical output** across runs
- Use sub-seeds for each stage: `seed_stage = H(master, stage_name, config_hash)`
- All randomized decisions must use stage-local RNG
- This enables caching and stepwise testing

### Hard Constraints (Must Pass)
These must be satisfied before proceeding to next stage:
- Single connected component (unless `allowDisconnected=true`)
- Degree bounds per room and globally
- Key-before-lock reachability for all lock gates
- At least one viable path Start→Boss within `minLen..maxLen`
- No spatial overlaps; corridor feasibility

### Soft Constraints (Optimized)
These are targets to optimize for:
- Pacing adherence to curve (S-curve, linear, custom)
- Thematic clustering (biome continuity)
- Cycle density (avoid tree boredom and maze frustration)
- Secret density and reward pacing

### Performance Targets (v1)
- 60-room dungeon in < 50ms for graph + embedding
- < 200ms including carve/content
- Memory < 50MB per generation
- Single-threaded operation

## Configuration Format

Dungeons are configured via JSON/YAML with:
- Seed (for determinism)
- Size constraints (roomsMin, roomsMax)
- Branching parameters (avg, max)
- Pacing curves (S_CURVE, LINEAR, etc.)
- Biomes/themes (crypt, fungal, arcane, etc.)
- Key/lock definitions
- Hard and soft constraints

See `specs/dungo_specification.md` §5.2 and §13 for schema details.

## Testing Strategy

Per specification §16:

### Test Types
1. **Unit Tests** - Rooms, edges, constraints, seed stability
2. **Property Tests** - Connectivity, key-before-lock, degree bounds for random seeds
3. **Golden Tests** - Snapshot SVG/JSON for fixed seeds; diff on regressions
4. **Fuzzing** - Push synthesis and embedding until violations surface
5. **Simulated Agent** - Basic planner to verify findability of boss and keys

### Test Data
- Store artifacts under `testdata/` with versioned schema
- Include SVG visualizations of ADG
- Include PNG layouts with overlays for debugging

## Specification-Driven Development

This project uses the `.specify` framework for spec-driven development:

### Slash Commands
Available in `.claude/commands/`:
- `/speckit.specify` - Create/update feature specifications
- `/speckit.plan` - Execute implementation planning
- `/speckit.tasks` - Generate actionable task lists
- `/speckit.implement` - Execute implementation plan
- `/speckit.analyze` - Cross-artifact consistency analysis
- `/speckit.clarify` - Identify underspecified areas
- `/speckit.checklist` - Generate custom checklists

### Workflow
When implementing features:
1. Check `specs/dungo_specification.md` for requirements
2. Use `/speckit.plan` to create implementation plan
3. Use `/speckit.tasks` to break down into actionable items
4. Implement following TDD practices
5. Use `/speckit.analyze` to verify consistency

## Code Style Notes

### Cyclomatic Complexity
- Default max complexity: 15
- Exception: `pkg/llvmgen/builder.go` allows higher complexity for instruction switch statements
- Keep functions focused and single-purpose

### Package Organization
- Use `pkg/` for library packages (not `internal/` initially)
- Each stage (synthesis, embedding, carving, content) is a separate package
- Common types in `pkg/dungeon`
- Utilities in respective packages (no shared `util` package)

### Error Handling
- Return errors, don't panic
- Wrap errors with context using `fmt.Errorf` with `%w`
- Validate inputs early in pipeline stages

### Context Usage
- All long-running operations take `context.Context`
- Check context cancellation in loops
- Use context for timeout and cancellation

## Key Design Patterns

### Constraint DSL
The project includes a minimal constraint DSL (§12) for expressing requirements:
- Predicates: `isConnected()`, `hasPath(a,b)`, `keyBeforeLock(k)`, etc.
- Spatial: `noOverlap()`, `maxCorridorBends(b)`, `maxCorridorLen(l)`
- Pacing: `monotoneIncrease(start,boss,slope)`, `peakNear(tag, radius)`
- Composition: `and(...)`, `or(...)`, `not(...)`

### Extensibility Points
- **Themes**: Add tilesets, decorators, encounter/loot tables
- **Motifs**: Register graph templates and grammar rules
- **Gates**: Define new gate types (ability, rune count, boss token)
- **Biomes**: Attach biome-specific spatial preferences
- **Evaluators**: Extend DSL with custom predicates

### Pluggable Strategies
Each pipeline stage uses interfaces allowing swappable implementations:
- Graph synthesis: grammar-based, template stitching, search/optimize
- Embedding: force-directed, orthogonal drawing, packing with templates
- Content: various encounter spawners, loot distributors, puzzle injectors

## Dependencies

Current (from go.mod):
- Go 1.25.3
- Module: `github.com/dshills/dungo`

Expected dependencies (not yet added):
- Testing: standard library `testing`, possibly `github.com/stretchr/testify`
- Property testing: possibly `github.com/leanovate/gopter`
- YAML/JSON: `gopkg.in/yaml.v3` for config
- Visualization: SVG generation library for debug artifacts

## Out of Scope (v1)

The following are explicitly out of scope for v1:
- Real-time adaptive generation mid-run (planned for v2)
- 3D geometry export (v1 focuses on 2D tile/room graphs)
- Live player modeling/telemetry feedback loops (hooks provided for v2)
- Authoring tools UI (provide schemas and hooks only)
- Photorealistic rendering or engine-specific effects

## Active Technologies
- Go 1.25.3 + Standard library (encoding/json, crypto/sha256), gopkg.in/yaml.v3 (config parsing), github.com/ajstarks/svgo (SVG generation), pgregory.net/rapid (property testing) (001-dungeon-generator-core)
- File-based (config files: JSON/YAML, output artifacts: JSON/SVG/TMJ, golden test snapshots in testdata/) (001-dungeon-generator-core)

## Recent Changes
- 001-dungeon-generator-core: Added Go 1.25.3 + Standard library (encoding/json, crypto/sha256), gopkg.in/yaml.v3 (config parsing), github.com/ajstarks/svgo (SVG generation), pgregory.net/rapid (property testing)
