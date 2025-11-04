# Graph-Based Dungeon Generator — Technical Specification (v1.0)

> **Purpose:** Define a deterministic, extensible, and testable system that generates playable dungeons by first synthesizing an abstract **graph** of rooms and constraints, then embedding that graph into a **spatial layout**, and finally emitting **tile maps** and **gameplay content** (enemies, loot, locks/keys, secrets) that adhere to the constraints.

---

## 1. Scope

This specification covers:
- Abstract dungeon graph modeling (rooms, edges, constraints, tags, layers).
- Graph synthesis strategies (grammars, templates, and constraint solving).
- Spatial embedding and layout algorithms (grid and non-grid).
- Content placement systems (combat, loot, puzzles, secrets).
- Difficulty pacing, theme packs, biome integration.
- Determinism & seeding; reproducibility.
- Configuration formats, APIs, validation, metrics, and testing.
- Performance/security considerations; extensibility points.

Out of scope for v1.0:
- Real-time adaptive generation mid-run (v2.0).
- 3D geometry export (v1 focuses on 2D tile/room graphs; 2.5D allowed via metadata).
- Live player modeling/telemetry feedback loops (hooks provided).

---

## 2. Terminology

- **ADG (Abstract Dungeon Graph):** A labeled multigraph \(G = (V, E)\) where nodes are rooms and edges are connections (doors, corridors, teleporters).
- **Room:** Node in ADG with attributes: type, size class, tags, requirements, capacity, reward value.
- **Connector:** Edge in ADG with attributes: directionality, gating (lock/key), hazards, traversal cost.
- **Constraint:** Predicate over rooms/edges/subgraphs (e.g., connectivity, degree bounds, key-before-lock, monotonic difficulty, biome compatibility).
- **Embedding:** Mapping from ADG to spatial layout (coordinates, rotation, tile masks) respecting constraints.
- **Theme Pack:** Collection of tilesets, decorators, content tables, and rule overrides.
- **Pacing Curve:** Target intensity over path length (combat, puzzles, rest).

---

## 3. Goals & Non-Goals

### 3.1 Goals
- Deterministic generation from a seed (same inputs → same output).
- Separation of concerns: **Graph → Embed → Carve → Decorate** pipeline.
- Highly testable with unit, property-based, and golden tests.
- Modular: swappable strategies and content packs.
- Constraint-driven: hard (must) and soft (optimize) constraints.
- Scalable: supports 10–300 rooms; O(N log N) typical.

### 3.2 Non-Goals
- Photorealistic rendering; engine-specific effects.
- Authoring tools UI (provide schemas and hooks only).

---

## 4. High-Level Architecture

```
[Seed, Config, Theme]
        │
        ▼
  (A) Graph Synthesis ──▶ ADG (rooms/edges + constraints)
        │
        ▼
  (B) Spatial Embedding ──▶ Room placements, corridors, grid/coord map
        │
        ▼
  (C) Carving & Topology ──▶ Tile map (walls, floors, doors), portals
        │
        ▼
  (D) Content Pass ──▶ Enemies, loot, keys/locks, secrets, hazards
        │
        ▼
  (E) Validation & Scoring ──▶ Metrics, assertions, debug artifacts
```

Each stage is pure and deterministic given its inputs, facilitating caching and stepwise testing.

---

## 5. Data Model

### 5.1 Core Types (Abstract)
- **Room**
  - `id: string`
  - `archetype: enum {Start, Boss, Treasure, Puzzle, Hub, Corridor, Secret, Optional, Vendor, Shrine, Checkpoint}`
  - `size: enum {XS, S, M, L, XL}` (abstract; mapped to tiles later)
  - `tags: set<string>` (e.g., biome/element: `crypt`, `fungal`, `arcane`)
  - `difficulty: float in [0,1]` (target local intensity)
  - `reward: float in [0,1]`
  - `requirements: set<Requirement>` (e.g., requires ability `double_jump`)
  - `provides: set<Capability>` (e.g., provides `silver_key`)
  - `bounds: degMin/degMax`

- **Connector (Edge)**
  - `id: string`
  - `u, v: room-id`
  - `type: enum {Door, Corridor, Ladder, Teleporter, Hidden, OneWay}`
  - `gate: Gate?` (e.g., `key=silver`, `puzzle=runes_3`)
  - `cost: float` (pathfinding weight)
  - `visibility: enum {Normal, Secret, Illusory}`

- **Constraint**
  - `kind: enum {Connectivity, Degree, KeyLock, Pacing, Theme, Spatial, Cycle, PathLen, BranchFactor, SecretDensity, Optionality, LootBudget}`
  - `severity: enum {Hard, Soft}`
  - `expr: DSL` (see §12)

- **ThemePack**
  - `tilesets: {biome→tileset}`
  - `roomShapes: catalog of parametric templates`
  - `encounterTables: weighted lists`
  - `lootTables: weighted lists`
  - `decorators: rules → sprite/prop placements`

### 5.2 JSON Schema (Config Excerpt)
```json
{
  "seed": 123456789,
  "size": {"roomsMin": 35, "roomsMax": 60},
  "pacing": {"curve": "S_CURVE", "variance": 0.15},
  "branching": {"avg": 1.7, "max": 3},
  "keys": [
    {"name": "silver", "count": 2},
    {"name": "gold", "count": 1}
  ],
  "themes": ["crypt", "fungal"],
  "constraints": [
    {"kind": "Connectivity", "severity": "Hard", "expr": "isConnected()"},
    {"kind": "KeyLock", "severity": "Hard", "expr": "keyBeforeLock('silver')"},
    {"kind": "Pacing", "severity": "Soft", "expr": "monotoneIncrease(start, boss, slope=0.6)"}
  ]
}
```

---

## 6. Graph Synthesis

### 6.1 Approaches
- **Grammar-Based:** Start with a core trio (Start–Mid–Boss). Apply production rules \(P_i: subgraph → subgraph\) with probabilities conditioned on difficulty and branching targets.
- **Template Stitching:** Pre-authored motifs (e.g., key-loop, hub-spoke, optional treasure spur) combined via constraint solver.
- **Search/Optimize:** Initialize random graph then hill-climb or simulated anneal to maximize a fitness function with penalties for soft constraint violations.

### 6.2 Hard Constraints (must pass before embedding)
- Single connected component unless `allowDisconnected=true` for teleport motifs.
- Degree bounds per room and globally (avg branching).
- Key-before-lock reachability for all lock gates.
- At least one viable path Start→Boss within `minLen..maxLen`.
- Optional/secret rooms off main path must remain discoverable (not isolated without clue).

### 6.3 Soft Constraints (optimized)
- Pacing adherence to curve.
- Thematic clustering (biome continuity with occasional contrast).
- Cycle density within band (to avoid tree boredom and maze frustration).
- Secret density and reward pacing.

### 6.4 Production Rules (Examples)
- `ExpandHub(h, k)` adds `k` spokes with {Puzzle, Combat, Treasure} distribution.
- `InsertKeyLoop(path, keyName)` creates detour subgraph placing key on side path before lock on main trunk.
- `BranchOptional(room, rarity)` adds optional spur with probability tied to rarity.

---

## 7. Spatial Embedding

### 7.1 Objectives
Map ADG to 2D coordinates and orientations while minimizing crossings, respecting room size/shape, and ensuring corridors fit within bounds.

### 7.2 Strategies
- **Force-Directed Pre-Layout:** Compute continuous positions minimizing edge length and overlap penalties, then quantize to grid.
- **Orthogonal Graph Drawing:** BFS layering from Start, assign lanes, then route Manhattan corridors.
- **Packing with Room Templates:** Choose a room footprint from templates per room size; place via A* packing with collision checks.

### 7.3 Constraints
- Non-overlap of room bounding boxes.
- Corridor feasibility (length ≤ `L_max`, bend count ≤ `B_max`).
- Spatial themes (e.g., crypt catacombs prefer orthogonal, fungal caverns prefer organic noise fields).

### 7.4 Output
- Vertex `pose`: `(x, y, rot, footprint-id)`
- Edge `path`: polyline grid cells with door positions.

---

## 8. Carving & Topology

- Start from empty solid grid (or noise-filled field for caves).
- **Carve Rooms:** Stamp footprints; mark floors/walls/doors.
- **Route Corridors:** Rasterize edge paths; insert doors/gates, stairs, ladders as per Connector type.
- **Post-Process:** Wall thickening, pillar placement, dead-end dressing, loop breakers/creators (if allowed).
- **Navmesh/Graph:** Produce navigation graph for AI (cells or portals).

Tile IDs and metadata are theme-dependent but use a consistent logical layer model: `floor`, `wall`, `door`, `decor`, `trigger`, `hazard`.

---

## 9. Content Placement

### 9.1 Systems
- **Encounter Spawner:** Assigns enemy sets per room using encounter tables keyed by biome and difficulty. Respects capacity and line-of-sight.
- **Loot Distributor:** Budget-based allocation along critical path and optionals; guarantees required items before their gates.
- **Puzzle Injector:** Inserts puzzle widgets per archetype; exposes `requirements` and `provides` for gating logic.
- **Secret Manager:** Generates hidden connectors/rooms adhering to `secretDensity` and clue rules (visual hints, map fragments).

### 9.2 Pacing Control
- Compute main path via shortest path weighted by gate costs.
- Fit local difficulty to target curve using per-room modifiers and enemy composition.
- Insert rest/safe rooms at intervals.

---

## 10. Determinism & Seeding

- Single **master seed** + stage-specific sub-seeds via hash derivation: `seed_stage = H(master, stage_name, config_hash)`.
- All randomized decisions derive from stage-local RNG to ensure reproducibility and cacheability.

---

## 11. Validation & Metrics

**Hard Assertions:**
- Graph connectivity; Start and Boss exist; path length bounds.
- Key reachability (for each lock L, exists path Start→Key(L)→...→L).
- Spatial overlaps = 0; corridor feasibility.

**Metrics (for scoring and regression):**
- Branching factor (avg, variance), cycle count, optional ratio.
- Path length to boss, detour cost distribution.
- Pacing deviation (L2 against curve).
- Secret findability score (via heuristic or simulated agent).

Output a `report.json` and debug artifacts (SVG of ADG, PNG of layout with overlays).

---

## 12. Constraint DSL (Minimal v1)

A small, pure, composable DSL evaluated against the ADG and embedded layout:

- Predicates: `isConnected()`, `hasPath(a,b,lenMin?,lenMax?)`, `degreeInRange(tag, min, max)`, `keyBeforeLock(k)`, `cycleCountIn(min,max)`, `branchingAvgIn(min,max)`.
- Spatial: `noOverlap()`, `maxCorridorBends(b)`, `maxCorridorLen(l)`.
- Pacing: `monotoneIncrease(start,boss,slope?)`, `peakNear(tag='Boss', radius=3)`.
- Composition: `and(...)`, `or(...)`, `not(...)`.

Constraints serialize as JSON or a simple s-expression string; engine compiles to evaluators.

---

## 13. Configuration Files

### 13.1 YAML Example
```yaml
seed: 987654321
size: { roomsMin: 45, roomsMax: 70 }
branching: { avg: 1.8, max: 3 }
pacing: { curve: S_CURVE, variance: 0.1 }
biomes: [crypt, arcane]
keys:
  - { name: silver, count: 2 }
  - { name: gold, count: 1 }
constraints:
  - { kind: Connectivity, severity: Hard, expr: "isConnected()" }
  - { kind: KeyLock, severity: Hard, expr: "keyBeforeLock('gold')" }
  - { kind: Pacing, severity: Soft, expr: "monotoneIncrease(start,boss,0.5)" }
```

---

## 14. APIs (Engine-Facing)

### 14.1 Language-Agnostic Interface
- `Generate(config) -> DungeonArtifact` (pure function)
- `Validate(artifact) -> ValidationReport`
- `Visualize(artifact, mode) -> Image/SVG`

### 14.2 Go Interface (reference)
```go
// Core entry points
package dungeon

type Config struct {
    Seed        uint64
    Size        SizeCfg
    Branching   BranchingCfg
    Pacing      PacingCfg
    Themes      []string
    Keys        []KeyCfg
    Constraints []Constraint
}

type Artifact struct {
    ADG        Graph
    Layout     Layout
    TileMap    TileMap
    Content    Content
    Metrics    Metrics
    Debug      DebugArtifacts
}

type Generator interface {
    Generate(ctx context.Context, cfg Config) (Artifact, error)
}

// Pluggable strategies
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

---

## 15. Extensibility Points

- **Themes:** Add new tilesets, decorators, encounter & loot tables.
- **Motifs:** Register new graph templates and grammar rules.
- **Gates:** Define new gate types (e.g., ability, rune count, boss token).
- **Biomes:** Attach biome-specific spatial preferences and hazards.
- **Evaluators:** Extend DSL with custom predicates.

Registration via simple plugin mechanism (Go: init-time registration into maps; or runtime via config reflection).

---

## 16. Testing Strategy

- **Unit Tests:** Rooms, edges, constraints, seed stability.
- **Property Tests:** Connectivity, key-before-lock, degree bounds for random seeds.
- **Golden Tests:** Snapshot SVG/JSON for fixed seeds and configs; diff on regressions.
- **Fuzzing:** Push synthesis and embedding until constraint violations surface.
- **Simulated Agent:** Basic planner to verify findability of boss and required keys.

Artifacts stored under `testdata/` with versioned schema.

---

## 17. Performance Targets (v1)

- 60-room dungeon in < 50 ms on modern desktop (single thread) for graph + embedding; < 200 ms including carve/content.
- Memory < 50 MB peak per generation for 100-room target.
- Complexity: synthesis O(N log N), embedding O(N log N + E) with efficient packing; carving O(cells) linear.

---

## 18. Failure Modes & Recovery

- If soft constraints cannot be satisfied within budgeted iterations, return best-scoring artifact with warnings.
- If hard constraints fail after `R` retries (default 5), bubble explicit error with minimal repro bundle (seed, config hash, partial graphs).

---

## 19. Output Formats

- **ADG:** JSON (nodes/edges with attributes) + SVG visualization.
- **Layout:** JSON with room poses and corridor polylines.
- **TileMap:** RLE-compressed layer arrays or Tiled (TMJ) export.
- **Content:** JSON lists of spawns, items, triggers.
- **Report:** Validation metrics and warnings.

---

## 20. Security & Safety

- All inputs validated against JSON/YAML schemas.
- Time/iteration caps to avoid unbounded search.
- No dynamic code execution from content packs (data-only, or sandboxed WASM for advanced rules).

---

## 21. Examples

### 21.1 Minimal Config → Small Dungeon
- 25–30 rooms, one key/lock loop, S-curve pacing, crypt theme.
- Expected: Start → Hub → Key Loop → Midboss → Boss, with 2–3 optional spurs and 1 secret.

### 21.2 Large Hub-Spoke with Two Keys
- 60–80 rooms, dual biomes (crypt/arcane), gold locks gate late-game.
- Expected: Two hubs, inter-hub loop, silver key mid, gold key late, treasure rooms off branches.

---

## 22. Roadmap

- **v1.1:** 3D cell complexes, verticality (z-layers), elevators, drop-throughs.
- **v1.2:** Adaptive difficulty from playtest telemetry; mission objectives overlay.
- **v2.0:** Real-time generation during gameplay; co-op constraints.

---

## 23. Acceptance Criteria

1. Given the same seed and config, output artifact hashes are identical (excluding timestamps and non-deterministic debug fields).
2. Validation report passes all hard constraints.
3. Performance targets met for small/medium configs on reference hardware.
4. At least three theme packs (crypt, fungal, arcane) with distinct visual and pacing signatures.
5. Documentation includes schema files, API signatures, and 5 golden seeds with SVG/PNG outputs.

---

## 24. Appendix: Fitness Function (Reference)

Objective to maximize:

```
score = w_conn * f_conn
      + w_pace * (1 - L2(pacing, target))
      + w_cycles * inBand(cycleCount)
      + w_branch * inBand(branchFactor)
      + w_secret * secretScore
      + w_theme * themeCohesion
```

Subject to: all hard constraints satisfied.

---

## 25. Appendix: Debug Artifacts

- `ADG.svg`: nodes colored by archetype; edges by gate type; keys/locks annotated.
- `layout.png`: heatmap overlay for difficulty; arrows for one-ways; hatching for secrets.
- `report.json`: machine-readable metrics and violations.

---

**End of v1.0**

