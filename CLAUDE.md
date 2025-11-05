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

**REQUIRED** per constitution v1.1.1: Run mcp-pr before committing.

**Security First**: Never review files with secrets, credentials, or PII.

```bash
# Stage your changes first
git add <files>

# Run code review on staged changes (in Claude Code)
# Use mcp-pr tool: mcp__mcp-pr__review_staged with OpenAI provider

# Save evidence for PR compliance
# The review will show findings with severity levels

# Address findings:
# - HIGH severity: MUST fix before commit
# - MEDIUM severity: Document as tech debt if not immediately fixable
# - LOW/INFO: Optional improvements
```

**Exemptions**:
- Trivial changes (<10 lines)
- Automated commits (e.g., dependency updates)
- Emergency hotfixes (with post-commit review)

**Data Privacy**:
- Never review files containing: API keys, passwords, tokens, certificates, PII
- Review only code in `pkg/`, `cmd/`, test files
- Exclude: `.env`, `*_secret.go`, `credentials/`, `secrets/`

**API Key Handling**:
- Store OpenAI API key in environment variable or keychain
- Never commit API keys to repository
- Rotate keys regularly per security policy

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

## Implementation Lessons Learned

### Coordinate System Management
**Critical Insight**: Different pipeline stages use different coordinate systems. Careful conversion is essential.

- **embedding.Layout**: Uses corner coordinates (X, Y = top-left) with Width/Height
- **dungeon.Layout**: Uses center coordinates (X, Y = center point) for carving
- **carving.Layout**: Also uses center coordinates for pose-based stamping

**Best Practice**: Always normalize embedding layouts BEFORE converting to center coordinates. Call `normalizeEmbeddingLayout()` while you still have Width/Height info to handle room bounds correctly.

### Corridor Length Scaling
Corridor max length must scale with dungeon size to avoid embedding failures:
- Small dungeons (10-25 rooms): ~100-295 units
- Medium dungeons (25-100 rooms): ~295-590 units
- Large dungeons (100+ rooms): 590-600 units (capped)

Formula: `maxLength = sqrt(roomCount) * 59`, clamped to [100, 600]

**Updated progression: 20 → 41 → 59** (195% total increase from original)
- Initial multiplier of 20 caused failures in 11-52% of cases
- Increased to 41 based on typical force-directed spreading
- Final increase to 59 needed to handle pathological seeds that create very spread-out layouts
- Pathological seed 0x4400f4 demonstrated corridors up to 74*sqrt(N) units in worst cases

### Force Balancing for Compact Layouts
**Critical**: The default force-directed parameters (SpringConstant=0.5, RepulsionConstant=500.0) create a 1000:1 ratio favoring repulsion, causing very spread-out layouts.

**Solution**: For dungeons >25 rooms, dynamically adjust force balance:
- **Increase spring constant (attraction)**: Scale up to 10x for large dungeons
  - Formula: `scaleFactor = 1.0 + (roomCount-25) / 10.0`, capped at 10x
  - For 32 rooms: 1.7x, for 80 rooms: 6.5x, for 100+ rooms: 10x
- **Decrease repulsion constant**: Scale down to 0.2x for large dungeons
  - Formula: `repulsionScale = 1.0 / (1.0 + (roomCount-25)/50.0)`, floored at 0.2x
  - For 32 rooms: 0.88x, for 80 rooms: 0.52x, for 100+ rooms: 0.2x

This dual scaling dramatically shifts the force balance, keeping layouts compact enough to satisfy corridor length constraints while still maintaining natural spacing between rooms.

### Validation Integration
The validator must be set AFTER generator construction to avoid import cycles:
```go
validator := validation.NewValidator()
gen := dungeon.NewGeneratorWithValidator(validator)
```

Alternative: Use `SetValidator()` method if constructing generator first.

### Graph Synthesis Constraints
When implementing grammar-based synthesis:
- Always create exactly 1 Start and 1 Boss room
- Ensure connectivity check before returning (BFS from Start)
- Respect degree bounds both per-room and globally
- Handle key-before-lock constraints during graph construction, not post-hoc
- Use retry logic with seed perturbation if constraints fail

### Force-Directed Embedding Determinism
**CRITICAL**: Go map iteration order is randomized, causing non-deterministic layouts!

**Problem**: Iterating over maps in force-directed simulation leads to different floating-point accumulation patterns, producing wildly different layouts from the same seed.

**Solution**: Always use sorted slices when iterating over rooms/positions:
```go
// BAD - Non-deterministic!
for roomID := range g.Rooms {
    // ...
}

// GOOD - Deterministic
roomIDs := make([]string, 0, len(g.Rooms))
for id := range g.Rooms {
    roomIDs = append(roomIDs, id)
}
sort.Strings(roomIDs)
for _, roomID := range roomIDs {
    // ...
}
```

**Key locations requiring sorted iteration**:
- `initializePositions`: Room initialization order affects RNG sequence
- `simulateForces`: Force calculation order affects floating-point accumulation
- `Embed`: Layout construction order (less critical but good practice)
- `resolveOverlaps`: Perturbation order must be deterministic

### Testing Strategy That Worked
1. **Unit tests**: Focus on individual functions, use table-driven tests
2. **Property tests**: Use rapid for randomized testing with invariants
3. **Integration tests**: Test full pipeline with fixed seeds
4. **Golden tests**: Store SVG/JSON snapshots, diff on changes
5. **Fuzz tests**: Stress edge cases (0 rooms, 1000 rooms, conflicting constraints)
6. **Agent simulation**: Verify boss/key reachability with pathfinding

### Export Format Compatibility
- **JSON**: Straightforward serialization, good for debugging
- **TMJ**: Requires careful adherence to Tiled spec (compression, layer IDs, chunk format)
- **SVG**: Use viewBox for proper scaling, include legend for interpretability

### Performance Gotchas
- Force-directed embedding: Iterations matter more than you think (1000+ for large graphs)
- Room spacing: Too tight (< 1.0) causes carving overlaps; too loose wastes space
- Corridor pathfinding: A* is overkill; simple Manhattan distance works well
- Theme table lookups: Cache compiled tables, don't parse YAML on every access

### CLI Design Patterns
- Always provide `-verbose` flag for debugging
- Support `-seed` override for reproducibility during testing
- Allow multiple export formats with `-format all`
- Print statistics by default (room count, generation time, validation status)
- Make `-config` required, provide helpful error messages

## Recent Changes
- 001-dungeon-generator-core: Added Go 1.25.3 + Standard library (encoding/json, crypto/sha256), gopkg.in/yaml.v3 (config parsing), github.com/ajstarks/svgo (SVG generation), pgregory.net/rapid (property testing)
- Implemented full pipeline: synthesis → embedding → carving → content → validation
- Added CLI tool (cmd/dungeongen) with multi-format export support
- Created fuzz tests for edge cases and agent-based pathfinding validation
