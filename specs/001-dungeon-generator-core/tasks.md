# Tasks: Graph-Based Dungeon Generator

**Input**: Design documents from `/specs/001-dungeon-generator-core/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are MANDATORY per project constitution. All tasks MUST include test-first workflow (write tests → verify they fail → implement → verify they pass).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Single Go library project**: `pkg/` at repository root
- Paths shown below use pkg/ structure from plan.md
- Tests co-located with implementation (_test.go files)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize go.mod with dependencies (gopkg.in/yaml.v3, github.com/ajstarks/svgo, pgregory.net/rapid)
- [x] T002 [P] Create pkg/dungeon directory with package structure
- [x] T003 [P] Create pkg/graph directory with package structure
- [x] T004 [P] Create pkg/rng directory with package structure
- [x] T005 [P] Create pkg/synthesis directory with package structure
- [x] T006 [P] Create pkg/embedding directory with package structure
- [x] T007 [P] Create pkg/carving directory with package structure
- [x] T008 [P] Create pkg/content directory with package structure
- [x] T009 [P] Create pkg/validation directory with package structure
- [x] T010 [P] Create testdata/seeds directory for test configurations
- [x] T011 [P] Create testdata/golden directory for expected outputs
- [x] T012 [P] Create testdata/schemas directory for JSON schemas
- [x] T013 [P] Create themes/crypt directory for first theme pack
- [x] T014 [P] Create themes/fungal directory for second theme pack
- [x] T015 [P] Create themes/arcane directory for third theme pack
- [x] T016 Create cmd/dungeongen directory for CLI tool

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T017 [P] Implement RNG type with seed derivation in pkg/rng/rng.go (crypto/sha256 for H(master, stage, config))
- [x] T018 [P] Write unit tests for RNG determinism in pkg/rng/rng_test.go
- [x] T019 [P] Implement Config type with validation in pkg/dungeon/config.go
- [x] T020 [P] Implement YAML config parsing in pkg/dungeon/config.go (gopkg.in/yaml.v3)
- [x] T021 [P] Write unit tests for Config validation in pkg/dungeon/config_test.go
- [x] T022 [P] Implement Room type with enums in pkg/graph/room.go
- [x] T023 [P] Implement Connector type with enums in pkg/graph/connector.go
- [x] T024 [P] Implement Constraint type and DSL parser in pkg/graph/constraint.go
- [x] T025 Implement Graph type with operations (AddRoom, AddConnector, GetPath) in pkg/graph/graph.go
- [x] T026 Write unit tests for Graph operations in pkg/graph/graph_test.go
- [x] T027 [P] Implement Artifact type structure in pkg/dungeon/artifact.go
- [x] T028 [P] Implement basic Generator interface stub in pkg/dungeon/dungeon.go
- [x] T029 Create test configurations in testdata/seeds/small_crypt.yaml
- [x] T030 [P] Create test configurations in testdata/seeds/medium_fungal.yaml
- [x] T031 [P] Create test configurations in testdata/seeds/large_dual_biome.yaml

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Generate Deterministic Dungeons (Priority: P1) 🎯 MVP

**Goal**: Implement complete end-to-end dungeon generation pipeline that produces deterministic, reproducible dungeons from seed and configuration. This is the core MVP.

**Independent Test**: Generate dungeon from fixed seed, regenerate with same seed, verify byte-for-byte identical JSON output.

### Tests for User Story 1 (MANDATORY - TDD) ⚠️

> **CRITICAL: Write these tests FIRST, ensure they FAIL before implementation (Test-Driven Development)**

- [x] T032 [P] [US1] Property test for graph connectivity in pkg/graph/graph_test.go (use pgregory.net/rapid)
- [x] T033 [P] [US1] Property test for Start and Boss room presence in pkg/synthesis/synthesis_test.go
- [x] T034 [P] [US1] Property test for key-before-lock reachability in pkg/graph/graph_test.go
- [x] T035 [P] [US1] Property test for room count bounds in pkg/synthesis/synthesis_test.go
- [x] T036 [P] [US1] Golden test for determinism (same seed → identical output) in pkg/dungeon/dungeon_test.go
- [x] T037 [P] [US1] Integration test for complete pipeline in pkg/dungeon/dungeon_test.go

### Implementation for User Story 1

#### Synthesis Stage

- [x] T038 [P] [US1] Implement GraphSynthesizer interface and registry in pkg/synthesis/synthesizer.go
- [x] T039 [US1] Implement GrammarSynthesizer with production rules in pkg/synthesis/grammar.go
- [x] T040 [US1] Add Start-Boss-Mid core trio generation in pkg/synthesis/grammar.go
- [x] T041 [US1] Add ExpandHub production rule in pkg/synthesis/grammar.go
- [x] T042 [US1] Add InsertKeyLoop production rule for key placement in pkg/synthesis/grammar.go
- [x] T043 [US1] Add BranchOptional production rule for optional rooms in pkg/synthesis/grammar.go
- [x] T044 [US1] Implement hard constraint validation (connectivity, degree bounds) in pkg/synthesis/grammar.go
- [x] T045 [US1] Write unit tests for GrammarSynthesizer in pkg/synthesis/synthesis_test.go

#### Embedding Stage

- [x] T046 [P] [US1] Implement Embedder interface and registry in pkg/embedding/embedder.go
- [x] T047 [P] [US1] Implement Pose and Layout types in pkg/embedding/layout.go
- [x] T048 [US1] Implement ForceDirectedEmbedder with spring model in pkg/embedding/force_directed.go
- [x] T049 [US1] Add room overlap prevention in pkg/embedding/force_directed.go
- [x] T050 [US1] Add corridor routing with Manhattan paths in pkg/embedding/force_directed.go
- [x] T051 [US1] Implement spatial constraint validation (no overlaps, corridor feasibility) in pkg/embedding/embedder.go
- [x] T052 [US1] Write unit tests for ForceDirectedEmbedder in pkg/embedding/embedding_test.go

#### Carving Stage

- [x] T053 [P] [US1] Implement Carver interface in pkg/carving/carver.go
- [x] T054 [P] [US1] Implement TileMap and Layer types in pkg/carving/tilemap.go
- [x] T055 [US1] Implement room stamper with footprint rasterization in pkg/carving/stamper.go
- [x] T056 [US1] Implement corridor router with polyline rasterization in pkg/carving/corridor.go
- [x] T057 [US1] Add door placement at room/corridor junctions in pkg/carving/carver.go
- [x] T058 [US1] Create basic tile layers (floor, walls, doors) in pkg/carving/carver.go
- [x] T059 [US1] Write unit tests for tile map generation in pkg/carving/carving_test.go

#### Content Stage

- [x] T060 [P] [US1] Implement ContentPass interface in pkg/content/content.go
- [x] T061 [P] [US1] Implement Spawn, Loot, Puzzle types in pkg/content/types.go
- [x] T062 [US1] Implement encounter spawner with capacity limits in pkg/content/encounter.go
- [x] T063 [US1] Implement loot distributor with budget allocation in pkg/content/loot.go
- [x] T064 [US1] Add required item placement (keys) before locks in pkg/content/loot.go
- [x] T065 [US1] Write unit tests for content placement in pkg/content/content_test.go

#### Validation Stage

- [x] T066 [P] [US1] Implement Validator interface in pkg/validation/validator.go
- [x] T067 [P] [US1] Implement ValidationReport and ConstraintResult types in pkg/validation/report.go
- [x] T068 [US1] Implement hard constraint checkers (connectivity, key reachability, no overlaps) in pkg/validation/constraints.go
- [x] T069 [US1] Implement metrics calculation (branching, path length, cycles) in pkg/validation/metrics.go
- [x] T070 [US1] Write unit tests for validation in pkg/validation/validation_test.go

#### Pipeline Integration

- [x] T071 [US1] Implement DefaultGenerator orchestrating all stages in pkg/dungeon/dungeon.go
- [x] T072 [US1] Add RNG sub-seed derivation for each stage in pkg/dungeon/dungeon.go
- [x] T073 [US1] Add context cancellation support in pkg/dungeon/dungeon.go
- [x] T074 [US1] Add error handling and retry logic for constraint failures in pkg/dungeon/dungeon.go
- [x] T075 [US1] Write integration test verifying full pipeline in pkg/dungeon/dungeon_test.go
- [x] T076 [US1] Run quality gates (golangci-lint run, go test ./...)

**Checkpoint**: At this point, User Story 1 should be fully functional - basic dungeon generation works end-to-end with determinism verified

---

## Phase 4: User Story 2 - Configure Dungeon Characteristics (Priority: P2)

**Goal**: Add configuration support for pacing curves, branching complexity, and theme selection. Enable game designers to control dungeon feel without code.

**Independent Test**: Create configs with different pacing curves (LINEAR, S_CURVE, EXPONENTIAL), verify generated dungeons match specifications.

### Tests for User Story 2 (MANDATORY - TDD) ⚠️

> **CRITICAL: Write these tests FIRST, ensure they FAIL before implementation (Test-Driven Development)**

- [ ] T077 [P] [US2] Property test for pacing curve adherence in pkg/synthesis/synthesis_test.go
- [ ] T078 [P] [US2] Property test for branching factor bounds in pkg/graph/graph_test.go
- [ ] T079 [P] [US2] Unit test for S-curve difficulty distribution in pkg/synthesis/pacing_test.go
- [ ] T080 [P] [US2] Unit test for theme assignment and clustering in pkg/synthesis/themes_test.go
- [ ] T081 [US2] Golden test for different pacing curves produce distinct difficulty distributions in pkg/dungeon/dungeon_test.go

### Implementation for User Story 2

- [ ] T082 [P] [US2] Implement pacing curve calculators (LINEAR, S_CURVE, EXPONENTIAL) in pkg/synthesis/pacing.go
- [ ] T083 [P] [US2] Implement custom pacing curve from points in pkg/synthesis/pacing.go
- [ ] T084 [US2] Add difficulty assignment to rooms based on pacing curve in pkg/synthesis/grammar.go
- [ ] T085 [US2] Implement branching factor control in grammar rules in pkg/synthesis/grammar.go
- [ ] T086 [US2] Implement theme assignment with clustering in pkg/synthesis/themes.go
- [ ] T087 [US2] Add multi-theme support with smooth transitions in pkg/synthesis/themes.go
- [ ] T088 [US2] Add variance tolerance to pacing curve in pkg/synthesis/pacing.go
- [ ] T089 [US2] Write unit tests for pacing module in pkg/synthesis/pacing_test.go
- [ ] T090 [US2] Write unit tests for theme clustering in pkg/synthesis/themes_test.go
- [ ] T091 [US2] Update soft constraint validation for pacing deviation in pkg/validation/constraints.go
- [ ] T092 [US2] Run quality gates (golangci-lint run, go test ./...)

**Checkpoint**: At this point, User Stories 1 AND 2 work independently - deterministic generation with configurable characteristics

---

## Phase 5: User Story 3 - Export Dungeon Data (Priority: P3)

**Goal**: Add export functionality for JSON, TMJ (Tiled), and SVG formats. Enable integration with game engines and visual debugging.

**Independent Test**: Generate dungeon, export to all formats, load each in appropriate tools (JSON parser, Tiled editor, browser), verify data integrity.

### Tests for User Story 3 (MANDATORY - TDD) ⚠️

- [ ] T093 [P] [US3] Unit test for JSON serialization round-trip in pkg/dungeon/export_test.go
- [ ] T094 [P] [US3] Unit test for TMJ export structure validation in pkg/export/tmj_test.go
- [ ] T095 [P] [US3] Unit test for SVG generation with all elements in pkg/export/svg_test.go
- [ ] T096 [US3] Integration test for Tiled editor compatibility in pkg/export/tmj_test.go
- [ ] T097 [US3] Golden test for SVG visualization consistency in pkg/export/svg_test.go

### Implementation for User Story 3

- [ ] T098 [P] Create pkg/export package for export functionality
- [ ] T099 [P] [US3] Implement JSON export for complete Artifact in pkg/export/json.go
- [ ] T100 [P] [US3] Implement TMJ map builder with layer support in pkg/export/tmj.go
- [ ] T101 [US3] Implement TMJ tile layer export (floor, walls, doors) in pkg/export/tmj.go
- [ ] T102 [US3] Implement TMJ object layer export (entities, triggers) in pkg/export/tmj.go
- [ ] T103 [US3] Add TMJ tileset reference generation in pkg/export/tmj.go
- [ ] T104 [US3] Add TMJ compression support (gzip) in pkg/export/tmj.go
- [ ] T105 [P] [US3] Implement SVG exporter using github.com/ajstarks/svgo in pkg/export/svg.go
- [ ] T106 [US3] Add node-edge graph visualization to SVG in pkg/export/svg.go
- [ ] T107 [US3] Add color-coding by room archetype in SVG in pkg/export/svg.go
- [ ] T108 [US3] Add difficulty heatmap overlay to SVG in pkg/export/svg.go
- [ ] T109 [US3] Add legend and annotations to SVG in pkg/export/svg.go
- [ ] T110 [P] [US3] Implement ValidationReport JSON export in pkg/validation/export.go
- [ ] T111 [P] [US3] Write unit tests for all exporters in pkg/export/export_test.go
- [ ] T112 [US3] Add export methods to Artifact in pkg/dungeon/artifact.go
- [ ] T113 [US3] Run quality gates (golangci-lint run, go test ./...)

**Checkpoint**: At this point, User Stories 1, 2, AND 3 work independently - generation with multiple export formats

---

## Phase 6: User Story 4 - Custom Content Packs (Priority: P4)

**Goal**: Enable theme pack extensibility with custom encounter tables, loot distributions, and decorators. Users can define their own content without modifying code.

**Independent Test**: Create custom theme pack YAML, generate dungeon with that theme, verify custom content appears correctly.

### Tests for User Story 4 (MANDATORY - TDD) ⚠️

- [ ] T114 [P] [US4] Unit test for theme pack YAML parsing in pkg/themes/loader_test.go
- [ ] T115 [P] [US4] Unit test for encounter table selection by difficulty in pkg/themes/tables_test.go
- [ ] T116 [P] [US4] Unit test for loot table weighted selection in pkg/themes/tables_test.go
- [ ] T117 [US4] Integration test for custom theme pack loading and usage in pkg/themes/loader_test.go

### Implementation for User Story 4

- [ ] T118 [P] Create pkg/themes package for theme pack management
- [ ] T119 [P] [US4] Implement ThemePack type structure in pkg/themes/theme.go
- [ ] T120 [P] [US4] Implement ThemeLoader interface in pkg/themes/loader.go
- [ ] T121 [US4] Add YAML theme pack parser in pkg/themes/loader.go
- [ ] T122 [US4] Implement encounter table lookup by difficulty in pkg/themes/tables.go
- [ ] T123 [US4] Implement loot table lookup by difficulty in pkg/themes/tables.go
- [ ] T124 [US4] Implement weighted random selection from tables in pkg/themes/tables.go
- [ ] T125 [US4] Add theme pack validation (required fields, valid structure) in pkg/themes/validator.go
- [ ] T126 [P] [US4] Create crypt theme pack YAML in themes/crypt/theme.yaml
- [ ] T127 [P] [US4] Create fungal theme pack YAML in themes/fungal/theme.yaml
- [ ] T128 [P] [US4] Create arcane theme pack YAML in themes/arcane/theme.yaml
- [ ] T129 [US4] Add theme pack integration to content placement in pkg/content/encounter.go
- [ ] T130 [US4] Update loot distributor to use theme loot tables in pkg/content/loot.go
- [ ] T131 [US4] Write unit tests for theme loading in pkg/themes/loader_test.go
- [ ] T132 [US4] Write unit tests for table selection in pkg/themes/tables_test.go
- [ ] T133 [US4] Run quality gates (golangci-lint run, go test ./...)

**Checkpoint**: All user stories complete - full feature set working independently

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and final quality assurance

- [ ] T134 [P] Add godoc comments to all exported types and functions
- [ ] T135 [P] Implement TemplateSynthesizer as alternative strategy in pkg/synthesis/template.go
- [ ] T136 [P] Implement OrthogonalEmbedder as alternative strategy in pkg/embedding/orthogonal.go
- [ ] T137 [P] Add benchmark tests for graph synthesis in pkg/synthesis/synthesis_bench_test.go
- [ ] T138 [P] Add benchmark tests for spatial embedding in pkg/embedding/embedding_bench_test.go
- [ ] T139 [P] Add benchmark tests for full generation pipeline in pkg/dungeon/dungeon_bench_test.go
- [ ] T140 Verify performance targets (<50ms graph+embedding, <200ms total) with benchmarks
- [ ] T141 [P] Add memory profiling and verify <50MB target
- [ ] T142 [P] Implement CLI tool in cmd/dungeongen/main.go for command-line generation
- [ ] T143 [P] Add CLI flags for config file, output format, seed override in cmd/dungeongen/main.go
- [ ] T144 Create example usage in cmd/dungeongen/README.md
- [ ] T145 [P] Add fuzz tests for edge cases (0 rooms, 1000 rooms, conflicting constraints) in pkg/synthesis/fuzz_test.go
- [ ] T146 [P] Implement simulated agent for secret findability testing in pkg/validation/agent.go
- [ ] T147 Update CLAUDE.md with implementation-specific guidance
- [ ] T148 Run final quality gates on all packages (golangci-lint run, go test ./...)
- [ ] T149 Generate golden test snapshots for 5 fixed seeds (testdata/golden/)
- [ ] T150 Run quickstart.md examples to verify documentation accuracy

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories - **MVP CORE**
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Extends US1 synthesis stage but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Operates on US1 artifacts but independently testable
- **User Story 4 (P4)**: Can start after Foundational (Phase 2) - Extends US1 content stage but independently testable

### Within Each User Story

- Tests MUST be written FIRST and FAIL before implementation (TDD non-negotiable)
- Core types before algorithms
- Interfaces before implementations
- Individual stage implementations before pipeline integration
- Unit tests pass before integration tests
- Quality gates pass before marking story complete

### Parallel Opportunities

- All Setup tasks (T002-T016) marked [P] can run in parallel
- All Foundational base types (T017-T024) marked [P] can run in parallel within Phase 2
- Once Foundational phase completes, all 4 user stories can start in parallel (if team capacity allows)
- Within each story, tests marked [P] can run in parallel
- Within each story, independent implementation tasks marked [P] can run in parallel
- Different user stories can be worked on in parallel by different agents

---

## Parallel Execution Examples

### Phase 1: Setup (All Parallel)

```bash
# Launch all directory creation tasks together:
Task: "Create pkg/dungeon directory with package structure"
Task: "Create pkg/graph directory with package structure"
Task: "Create pkg/rng directory with package structure"
Task: "Create pkg/synthesis directory with package structure"
Task: "Create pkg/embedding directory with package structure"
Task: "Create pkg/carving directory with package structure"
Task: "Create pkg/content directory with package structure"
Task: "Create pkg/validation directory with package structure"
```

### Phase 2: Foundational (Some Parallel)

```bash
# Launch all base type implementations together:
Task: "Implement RNG type with seed derivation in pkg/rng/rng.go"
Task: "Write unit tests for RNG determinism in pkg/rng/rng_test.go"
Task: "Implement Config type with validation in pkg/dungeon/config.go"
Task: "Implement YAML config parsing in pkg/dungeon/config.go"
Task: "Write unit tests for Config validation in pkg/dungeon/config_test.go"
Task: "Implement Room type with enums in pkg/graph/room.go"
Task: "Implement Connector type with enums in pkg/graph/connector.go"
Task: "Implement Constraint type and DSL parser in pkg/graph/constraint.go"
```

### Phase 3: User Story 1 Tests (All Parallel)

```bash
# Launch all tests for User Story 1 together (MANDATORY - TDD):
Task: "Property test for graph connectivity in pkg/graph/graph_test.go"
Task: "Property test for Start and Boss room presence in pkg/synthesis/synthesis_test.go"
Task: "Property test for key-before-lock reachability in pkg/graph/graph_test.go"
Task: "Property test for room count bounds in pkg/synthesis/synthesis_test.go"
Task: "Golden test for determinism in pkg/dungeon/dungeon_test.go"
Task: "Integration test for complete pipeline in pkg/dungeon/dungeon_test.go"
```

### Phase 3: User Story 1 Stage Implementations (Some Parallel)

```bash
# Launch interface definitions together:
Task: "Implement GraphSynthesizer interface and registry in pkg/synthesis/synthesizer.go"
Task: "Implement Embedder interface and registry in pkg/embedding/embedder.go"
Task: "Implement Pose and Layout types in pkg/embedding/layout.go"
Task: "Implement Carver interface in pkg/carving/carver.go"
Task: "Implement TileMap and Layer types in pkg/carving/tilemap.go"
Task: "Implement ContentPass interface in pkg/content/content.go"
Task: "Implement Spawn, Loot, Puzzle types in pkg/content/types.go"
```

### Multiple User Stories in Parallel

```bash
# After Foundational complete, launch all user stories concurrently:
Task: "Implement User Story 1 - Complete synthesis pipeline"
Task: "Implement User Story 2 - Pacing and configuration"
Task: "Implement User Story 3 - Export formats"
Task: "Implement User Story 4 - Theme pack system"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T016)
2. Complete Phase 2: Foundational (T017-T031) - **CRITICAL GATE**
3. Complete Phase 3: User Story 1 (T032-T076)
4. **STOP and VALIDATE**: Test US1 independently with golden tests
5. Demo working deterministic dungeon generation

**MVP Deliverable**: Basic dungeon generator that creates reproducible dungeons with rooms, connections, tiles, and content from seed values.

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (31 tasks)
2. Add User Story 1 (45 tasks) → Test independently → **MVP RELEASE**
3. Add User Story 2 (16 tasks) → Test independently → **Configuration Release**
4. Add User Story 3 (21 tasks) → Test independently → **Export Release**
5. Add User Story 4 (20 tasks) → Test independently → **Extensibility Release**
6. Complete Polish (17 tasks) → **v1.0 Release**

### Parallel Team Strategy

With multiple developers/agents:

1. Team completes Setup + Foundational together (31 tasks, ~2-3 hours)
2. Once Foundational is done:
   - Agent A: User Story 1 (synthesis + embedding stages) (25 tasks)
   - Agent B: User Story 1 (carving + content stages) (20 tasks)
   - Or: All 4 user stories in parallel after US1 tests pass
3. Stories integrate independently at Foundational interfaces

---

## Task Summary

- **Total Tasks**: 150
- **Setup Phase**: 16 tasks (all parallelizable)
- **Foundational Phase**: 15 tasks (11 parallelizable)
- **User Story 1 (P1 - MVP)**: 45 tasks (core generation pipeline)
  - Tests: 6 tasks (all parallel)
  - Synthesis: 8 tasks
  - Embedding: 7 tasks
  - Carving: 7 tasks
  - Content: 6 tasks
  - Validation: 5 tasks
  - Integration: 6 tasks
- **User Story 2 (P2)**: 16 tasks (configuration flexibility)
  - Tests: 5 tasks (4 parallel)
  - Implementation: 11 tasks
- **User Story 3 (P3)**: 21 tasks (export formats)
  - Tests: 5 tasks (4 parallel)
  - Implementation: 16 tasks
- **User Story 4 (P4)**: 20 tasks (theme packs)
  - Tests: 4 tasks (3 parallel)
  - Implementation: 16 tasks
- **Polish Phase**: 17 tasks (15 parallelizable)

### Parallel Execution Potential

- **Maximum parallelism**: 60+ tasks can run in parallel at various stages
- **Phase 1**: All 16 tasks in parallel
- **Phase 2**: 11 tasks in parallel
- **Phase 3-6**: Tests (4-6 per story), base implementations (8-12 per story)
- **Phase 7**: 15 tasks in parallel

### Test Coverage

- **Total Test Tasks**: 20 explicit test tasks (13% of total)
- **Test Types**: Property (12), Unit (20+), Golden (5), Integration (4), Fuzz (2), Benchmark (3)
- **TDD Enforcement**: All implementation follows write-test-first pattern per constitution

---

## Notes

- [P] tasks = different files, no dependencies on incomplete work
- [Story] label maps task to specific user story for traceability
- Each user story independently completable and testable
- Tests written FIRST, verify FAIL, then implement (TDD mandatory)
- Quality gates run after each story (golangci-lint + go test)
- Stop at any checkpoint to validate story independently
- MVP = Phase 1 + Phase 2 + Phase 3 (User Story 1) = 76 tasks = Core working system
