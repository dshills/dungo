# Implementation Plan: Graph-Based Dungeon Generator

**Branch**: `001-dungeon-generator-core` | **Date**: 2025-11-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-dungeon-generator-core/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a deterministic, constraint-driven dungeon generator that creates playable dungeons from seed values and configuration files. The system follows a five-stage pipeline (Graph Synthesis → Spatial Embedding → Carving → Content Placement → Validation) producing complete dungeons with rooms, connections, tile maps, and gameplay content. Primary focus is deterministic generation enabling reproducible procedural content that can be shared and replayed. Technical approach uses pure functional pipeline stages with stage-specific RNG seeding, property-based constraint validation, and pluggable strategies for graph synthesis and content placement.

## Technical Context

**Language/Version**: Go 1.25.3
**Primary Dependencies**: Standard library (encoding/json, crypto/sha256), gopkg.in/yaml.v3 (config parsing), github.com/ajstarks/svgo (SVG generation), pgregory.net/rapid (property testing)
**Storage**: File-based (config files: JSON/YAML, output artifacts: JSON/SVG/TMJ, golden test snapshots in testdata/)
**Testing**: Go standard library testing, pgregory.net/rapid (property-based tests), golden test pattern with testdata/
**Target Platform**: Cross-platform library (Linux, macOS, Windows), embedded in game engine build pipelines
**Project Type**: Single Go library project with package-per-stage architecture
**Performance Goals**: 60-room dungeon graph+embedding <50ms, full generation <200ms, O(N log N) complexity
**Constraints**: <50MB memory per generation, deterministic (same seed → identical output), single-threaded (v1), offline-capable (no network dependencies)
**Scale/Scope**: 10-300 rooms per dungeon, 5-stage pipeline, 8 core packages, support for 3+ theme packs, 11 room archetypes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Test-First Development ✅
- **Status**: PASS
- **Evidence**: Feature design explicitly requires TDD workflow with unit, property, golden, integration, and fuzz tests
- **Packages to test**: graph, synthesis, embedding, carving, content, validation, rng, dungeon (8 packages)
- **Test types planned**: Unit (all packages), Property (constraints), Golden (determinism), Integration (simulated agent), Fuzz (boundary conditions)

### Principle II: Quality Gates ✅
- **Status**: PASS
- **Pre-commit gates**: golangci-lint run (zero errors), go test ./... (100% pass), gofmt -l . (empty output)
- **Linters configured**: gofmt, govet, staticcheck, errcheck, gosimple, ineffassign, unused, typecheck, gocyclo (max 15), misspell
- **Configuration**: .golangci.yml already exists with project-specific rules

### Principle III: Deterministic Design ✅
- **Status**: PASS
- **Requirements aligned**: FR-001 mandates identical seed → identical output, SC-001 requires byte-for-byte reproducibility
- **Design approach**: Single master seed, stage-specific sub-seeds via H(master, stage_name, config_hash), pure pipeline functions
- **RNG strategy**: Stage-local RNG (pkg/rng), never global state, all randomized decisions derive from stage seeds

### Principle IV: Performance Consciousness ✅
- **Status**: PASS
- **Targets defined**: <50ms graph+embedding for 60 rooms, <200ms full generation, <50MB memory
- **Success criteria**: SC-002 (25-35 rooms <100ms), SC-003 (50-70 rooms <200ms), SC-010 (<50MB for 100 rooms)
- **Complexity goals**: O(N log N) synthesis and embedding, O(cells) linear carving
- **Monitoring**: Benchmarks planned via `go test -bench ./...`

### Principle V: Concurrent Execution ✅
- **Status**: PASS
- **Development workflow**: Use concurrent agents for parallel research, independent package development, parallel test execution
- **Package independence**: 8 packages designed to be independently developed and tested (graph, synthesis, embedding, carving, content, validation, rng, dungeon)
- **Task marking**: Implementation tasks will be marked with [P] for parallelizable work

### Summary: All Gates PASS ✅

No constitution violations. Project aligns with all five core principles. No complexity tracking required.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
pkg/
├── dungeon/          # Core generator interface, Config, Artifact types
│   ├── dungeon.go
│   ├── config.go
│   ├── artifact.go
│   └── dungeon_test.go
├── graph/            # ADG (Abstract Dungeon Graph) data structures
│   ├── room.go       # Room node with archetype, size, tags, difficulty
│   ├── connector.go  # Edge with type, gate, cost, visibility
│   ├── constraint.go # Constraint types and DSL evaluators
│   ├── graph.go      # Graph structure and operations
│   └── graph_test.go
├── rng/              # Deterministic RNG with sub-seed derivation
│   ├── rng.go        # Stage-local RNG, seed derivation via H(master, stage, config)
│   └── rng_test.go
├── synthesis/        # Graph synthesis strategies
│   ├── synthesizer.go    # Interface and registry
│   ├── grammar.go        # Grammar-based synthesis
│   ├── template.go       # Template stitching
│   ├── optimizer.go      # Search/optimize approach
│   └── synthesis_test.go
├── embedding/        # Spatial layout algorithms
│   ├── embedder.go       # Interface and registry
│   ├── force_directed.go # Force-directed pre-layout
│   ├── orthogonal.go     # Orthogonal graph drawing
│   ├── packing.go        # Room template packing
│   └── embedding_test.go
├── carving/          # Tile map generation from spatial layout
│   ├── carver.go         # Interface and tile map structures
│   ├── stamper.go        # Room footprint stamping
│   ├── corridor.go       # Corridor routing
│   └── carving_test.go
├── content/          # Content placement (enemies, loot, puzzles)
│   ├── content.go        # Interface and content structures
│   ├── encounter.go      # Enemy spawner with encounter tables
│   ├── loot.go           # Loot distributor with budget allocation
│   ├── puzzle.go         # Puzzle injector
│   ├── secret.go         # Secret manager
│   └── content_test.go
└── validation/       # Constraint validation and metrics
    ├── validator.go      # Interface and validation report
    ├── constraints.go    # Hard constraint checkers
    ├── metrics.go        # Scoring and metrics calculation
    └── validation_test.go

testdata/             # Golden test snapshots
├── seeds/            # Fixed seed configurations
├── golden/           # Expected JSON/SVG outputs
└── schemas/          # Versioned schema definitions

themes/               # Theme pack data (3+ packs)
├── crypt/
│   ├── theme.yaml    # Theme pack manifest
│   ├── encounters.yaml
│   ├── loot.yaml
│   └── tiles/
├── fungal/
│   └── [same structure]
└── arcane/
    └── [same structure]

cmd/                  # Optional CLI tools for testing/demo
└── dungeongen/
    └── main.go       # CLI wrapper for generator
```

**Structure Decision**: Single Go library project using package-per-stage architecture as mandated by constitution. Each pipeline stage (synthesis, embedding, carving, content, validation) is an independent package enabling parallel development and testing. Core types live in `pkg/dungeon`, shared RNG in `pkg/rng`, graph structures in `pkg/graph`. Theme packs stored as data files in `themes/`. Golden test artifacts in `testdata/` with versioned schemas. Optional CLI tool in `cmd/` for demo/integration testing.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

N/A - No constitution violations detected.

---

## Post-Design Constitution Re-Check

*Performed after Phase 1 design (research, data-model, contracts, quickstart complete)*

### Principle I: Test-First Development ✅
- **Status**: PASS (CONFIRMED)
- **Design artifacts support TDD**: Contracts define clear interfaces for testing, data model includes validation rules, quickstart includes testing examples
- **Test coverage planned**: All 8 packages have corresponding _test.go files in structure
- **Property tests designed**: research.md specifies pgregory.net/rapid with concrete graph property examples
- **Golden tests designed**: testdata/ structure defined with seeds/ and golden/ directories

### Principle II: Quality Gates ✅
- **Status**: PASS (CONFIRMED)
- **Pre-commit workflow documented**: quickstart.md includes quality gate commands
- **Linter integration planned**: .golangci.yml configuration exists and is referenced
- **Test execution defined**: contracts specify how to run `go test ./...`

### Principle III: Deterministic Design ✅
- **Status**: PASS (CONFIRMED)
- **RNG package designed**: pkg/rng with sub-seed derivation via H(master, stage, config)
- **Pure functions enforced**: All interfaces take RNG parameter, no global state
- **Determinism testable**: Golden tests verify byte-for-byte reproducibility

### Principle IV: Performance Consciousness ✅
- **Status**: PASS (CONFIRMED)
- **Benchmarks planned**: research.md includes `go test -bench` examples
- **Performance targets reconfirmed**: <50ms graph+embedding, <200ms total, <50MB memory
- **Complexity documented**: data-model.md specifies O(N log N) operations

### Principle V: Concurrent Execution ✅
- **Status**: PASS (CONFIRMED)
- **Package independence verified**: 8 packages with minimal inter-dependencies enable parallel development
- **Research conducted in parallel**: Used 3 concurrent research agents for SVG, property testing, TMJ format
- **Test parallelization enabled**: Standard Go test runner supports `go test ./... -parallel N`

### Post-Design Summary: All Principles PASS ✅

Design artifacts (research.md, data-model.md, contracts/, quickstart.md) fully support all constitutional principles. No violations introduced during design phase. Ready for Phase 2 task generation.

---

## Phase Completion Status

- ✅ **Phase 0 (Research)**: research.md complete - all technical decisions made (SVG: ajstarks/svgo, Property testing: pgregory.net/rapid, TMJ: custom implementation)
- ✅ **Phase 1 (Design)**: data-model.md, contracts/, quickstart.md complete
- ✅ **Agent Context**: CLAUDE.md updated with technology stack
- ✅ **Constitution Re-Check**: All principles pass post-design validation

**Next Step**: Use `/speckit.tasks` to generate actionable task list from design artifacts.
