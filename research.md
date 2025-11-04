# Property-Based Testing Frameworks for Go: Dungeon Generator Research

## Executive Summary

This document evaluates property-based testing (PBT) frameworks for Go to test constraint satisfaction in a graph-based dungeon generator. After comprehensive research, **pgregory.net/rapid** is the primary recommendation, with gopter as a strong alternative for teams requiring more mature tooling.

---

## 1. Top 3 Recommended Frameworks

### 1.1 rapid (pgregory.net/rapid) - PRIMARY RECOMMENDATION

**Version:** v1.2.0 (Released: February 26, 2025)
**License:** Mozilla Public License 2.0
**Go Version:** 1.21+ (inferred from codebase)
**GitHub:** https://github.com/flyingmutant/rapid
**Stars:** 687+ | **Dependents:** 7,000+

#### Key Features
- **Imperative API with type-safe generators using generics** - Modern Go design leveraging type parameters
- **Fully automatic shrinking** - No manual shrinking code required; automatically minimizes failing test cases
- **Deterministic seed support** - `Example(seed)` method produces deterministic values for reproducibility
- **State machine testing** - Built-in `T.Repeat()` and `StateMachineActions()` helpers
- **Zero external dependencies** - Only uses Go standard library
- **Superior data generation** - Biased toward edge cases and "small" values

#### Strengths for Dungeon Testing
- Modern, clean API ideal for graph property testing
- Excellent for testing connectivity invariants with automatic minimization
- Built-in state machine support for testing graph mutations
- Active maintenance (latest release Feb 2025)
- Strong for testing complex constraint satisfaction

#### Example Generators
```go
// Collections
rapid.SliceOf(elementGen)
rapid.SliceOfDistinct(elementGen) // Unique elements
rapid.MapOf(keyGen, valueGen)
rapid.Permutation(slice)

// Custom types
rapid.Custom(func(t *rapid.T) MyType { ... })
rapid.Deferred(func() *Generator) // Recursive structures
```

---

### 1.2 gopter (github.com/leanovate/gopter)

**Version:** v0.2.8 (Released: June 15, 2020)
**License:** MIT
**Go Version:** Modern Go modules supported
**GitHub:** https://github.com/leanovate/gopter
**Stars:** 614+ | **Dependents:** 2,600+

#### Key Features
- **Inspired by ScalaCheck/QuickCheck** - Brings Haskell/Scala PBT patterns to Go
- **Integrated shrinking** - Shrinking built into generators
- **Rich generator library** - Extensive built-in generators in `gopter/gen` package
- **Stateful testing via commands package** - Mature state machine testing support
- **Arbitrary type generation** - Automatic generator composition via `gopter/arbitrary`

#### Strengths for Dungeon Testing
- Mature, battle-tested codebase (4+ years in production)
- Extensive documentation and examples
- Powerful command-based state machine testing
- Rich ecosystem of generators
- Proven track record finding real bugs

#### Package Structure
```
gopter/
├── gen/          # Common generators
├── prop/         # Property helpers
├── arbitrary/    # Automatic generator composition
└── commands/     # Stateful test helpers
```

#### Considerations
- Less active maintenance (last release 2020, but stable)
- More verbose API compared to rapid
- Requires more manual setup for custom generators

---

### 1.3 testing/quick (Go Standard Library)

**Version:** Included with Go 1.25+
**License:** BSD-3-Clause (Go license)
**Documentation:** https://pkg.go.dev/testing/quick

#### Key Features
- **Part of Go standard library** - No external dependencies
- **Basic property testing** - Simple `Check()` function
- **Minimal API** - Easy to learn and integrate
- **Seed support** - `quick.Config.Rand` accepts custom RNG with seed

#### Strengths for Dungeon Testing
- Zero setup required
- Good for simple property checks
- Guaranteed long-term stability
- Team familiarity (likely)

#### Limitations
- **Feature-frozen** - No new features being added
- **No automatic shrinking** - Manual shrinking required
- **Limited generators** - Only supports types implementing `Generate(rand, size)`
- **Minimal documentation**
- Not suitable for complex constraint testing

#### When to Use
- Quick prototyping of property tests
- Simple invariant checks where shrinking isn't critical
- When external dependencies are prohibited
- Educational purposes

---

## 2. Code Example: Testing Graph Connectivity Properties

### 2.1 Example using rapid (Recommended)

```go
package dungeon_test

import (
    "testing"
    "pgregory.net/rapid"
)

// Graph represents an adjacency list for dungeon rooms
type Graph struct {
    Nodes map[int][]int // node ID -> list of connected node IDs
}

// IsConnected checks if all nodes are reachable from start node 0
func (g *Graph) IsConnected() bool {
    if len(g.Nodes) == 0 {
        return true
    }

    visited := make(map[int]bool)
    g.dfs(0, visited)
    return len(visited) == len(g.Nodes)
}

func (g *Graph) dfs(node int, visited map[int]bool) {
    visited[node] = true
    for _, neighbor := range g.Nodes[node] {
        if !visited[neighbor] {
            g.dfs(neighbor, visited)
        }
    }
}

// HasPath checks if there's a path from start to end
func (g *Graph) HasPath(start, end int) bool {
    if start == end {
        return true
    }
    visited := make(map[int]bool)
    return g.hasPathDFS(start, end, visited)
}

func (g *Graph) hasPathDFS(current, target int, visited map[int]bool) bool {
    if current == target {
        return true
    }
    visited[current] = true
    for _, neighbor := range g.Nodes[current] {
        if !visited[neighbor] {
            if g.hasPathDFS(neighbor, target, visited) {
                return true
            }
        }
    }
    return false
}

// InBounds checks if all node IDs are within valid range
func (g *Graph) InBounds(minID, maxID int) bool {
    for node, neighbors := range g.Nodes {
        if node < minID || node > maxID {
            return false
        }
        for _, neighbor := range neighbors {
            if neighbor < minID || neighbor > maxID {
                return false
            }
        }
    }
    return true
}

// GraphGenerator creates a custom generator for connected graphs
func GraphGenerator(nodeCount, maxEdgesPerNode int) *rapid.Generator[*Graph] {
    return rapid.Custom(func(t *rapid.T) *Graph {
        numNodes := rapid.IntRange(2, nodeCount).Draw(t, "numNodes")

        g := &Graph{
            Nodes: make(map[int][]int),
        }

        // Initialize nodes
        for i := 0; i < numNodes; i++ {
            g.Nodes[i] = []int{}
        }

        // Ensure connectivity: create spanning tree first
        for i := 1; i < numNodes; i++ {
            parent := rapid.IntRange(0, i-1).Draw(t, "parent")
            g.Nodes[parent] = append(g.Nodes[parent], i)
            g.Nodes[i] = append(g.Nodes[i], parent) // bidirectional
        }

        // Add random additional edges
        for node := range g.Nodes {
            numExtraEdges := rapid.IntRange(0, maxEdgesPerNode-len(g.Nodes[node])).
                Draw(t, "extraEdges")

            for j := 0; j < numExtraEdges; j++ {
                target := rapid.IntRange(0, numNodes-1).Draw(t, "target")
                if target != node {
                    g.Nodes[node] = append(g.Nodes[node], target)
                }
            }
        }

        return g
    })
}

// Test: All generated graphs must be connected
func TestGraphConnectivity(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        g := GraphGenerator(20, 5).Draw(t, "graph")

        if !g.IsConnected() {
            t.Fatalf("generated graph is not connected: %+v", g.Nodes)
        }
    })
}

// Test: Transitive reachability property
func TestGraphTransitiveReachability(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        g := GraphGenerator(20, 5).Draw(t, "graph")

        // Pick three random nodes
        nodeIDs := make([]int, 0, len(g.Nodes))
        for id := range g.Nodes {
            nodeIDs = append(nodeIDs, id)
        }

        if len(nodeIDs) < 3 {
            return // Skip if not enough nodes
        }

        a := rapid.SampledFrom(nodeIDs).Draw(t, "nodeA")
        b := rapid.SampledFrom(nodeIDs).Draw(t, "nodeB")
        c := rapid.SampledFrom(nodeIDs).Draw(t, "nodeC")

        // Property: if path(a->b) and path(b->c), then path(a->c)
        if g.HasPath(a, b) && g.HasPath(b, c) {
            if !g.HasPath(a, c) {
                t.Fatalf("transitivity violated: path(%d->%d) and path(%d->%d) exist, but not path(%d->%d)",
                    a, b, b, c, a, c)
            }
        }
    })
}

// Test: Dungeon-specific constraint - bounds checking
func TestDungeonBounds(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        g := GraphGenerator(50, 4).Draw(t, "graph")

        // All room IDs must be in valid range [0, 299]
        if !g.InBounds(0, 299) {
            t.Fatalf("graph has nodes outside valid dungeon bounds [0, 299]")
        }
    })
}

// Test: Boss room must be reachable from start
func TestBossReachability(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        g := GraphGenerator(60, 4).Draw(t, "graph")

        startRoom := 0
        // Boss is typically the last room
        bossRoom := len(g.Nodes) - 1

        if !g.HasPath(startRoom, bossRoom) {
            t.Fatalf("boss room %d not reachable from start room %d", bossRoom, startRoom)
        }
    })
}

// Test: Reproducibility with seed
func TestGraphGenerationDeterminism(t *testing.T) {
    gen := GraphGenerator(15, 3)

    // Generate with seed
    seed := uint64(42)
    g1 := gen.Example(seed)
    g2 := gen.Example(seed)

    // Both should be identical
    if len(g1.Nodes) != len(g2.Nodes) {
        t.Fatalf("seed %d produced different graphs: %d vs %d nodes",
            seed, len(g1.Nodes), len(g2.Nodes))
    }

    // Deep comparison would go here
    // In practice, use a proper equality check or hash comparison
}
```

### 2.2 Example using gopter

```go
package dungeon_test

import (
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

// Graph type (same as above)
type Graph struct {
    Nodes map[int][]int
}

// Methods: IsConnected, HasPath, InBounds (same as above)

// Custom generator for connected graphs
func GenConnectedGraph(nodeCount, maxEdges int) gopter.Gen {
    return func(genParams *gopter.GenParameters) *gopter.GenResult {
        rng := genParams.Rng

        numNodes := rng.Intn(nodeCount-1) + 2 // At least 2 nodes

        g := &Graph{
            Nodes: make(map[int][]int),
        }

        // Initialize nodes
        for i := 0; i < numNodes; i++ {
            g.Nodes[i] = []int{}
        }

        // Create spanning tree for connectivity
        for i := 1; i < numNodes; i++ {
            parent := rng.Intn(i)
            g.Nodes[parent] = append(g.Nodes[parent], i)
            g.Nodes[i] = append(g.Nodes[i], parent)
        }

        // Add random edges
        for node := range g.Nodes {
            extraEdges := rng.Intn(maxEdges - len(g.Nodes[node]) + 1)
            for j := 0; j < extraEdges; j++ {
                target := rng.Intn(numNodes)
                if target != node {
                    g.Nodes[node] = append(g.Nodes[node], target)
                }
            }
        }

        return gopter.NewGenResult(g, gopter.NoShrinker)
    }
}

func TestGraphConnectivityGopter(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("all generated graphs are connected", prop.ForAll(
        func(g *Graph) bool {
            return g.IsConnected()
        },
        GenConnectedGraph(20, 5),
    ))

    properties.TestingRun(t)
}

func TestTransitiveReachabilityGopter(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("transitive reachability holds", prop.ForAll(
        func(g *Graph, a, b, c int) bool {
            if a >= len(g.Nodes) || b >= len(g.Nodes) || c >= len(g.Nodes) {
                return true // Skip invalid indices
            }

            // If path(a->b) and path(b->c), then path(a->c)
            if g.HasPath(a, b) && g.HasPath(b, c) {
                return g.HasPath(a, c)
            }
            return true
        },
        GenConnectedGraph(20, 5),
        gen.IntRange(0, 19),
        gen.IntRange(0, 19),
        gen.IntRange(0, 19),
    ))

    properties.TestingRun(t)
}

func TestDungeonBoundsGopter(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("all rooms within dungeon bounds", prop.ForAll(
        func(g *Graph) bool {
            return g.InBounds(0, 299)
        },
        GenConnectedGraph(50, 4),
    ))

    properties.TestingRun(t)
}

// Determinism test with seed
func TestDeterminismGopter(t *testing.T) {
    seed := int64(424242)

    params1 := gopter.DefaultGenParameters()
    params1.Rng.Seed(seed)

    params2 := gopter.DefaultGenParameters()
    params2.Rng.Seed(seed)

    gen := GenConnectedGraph(15, 3)

    result1 := gen(params1)
    result2 := gen(params2)

    g1 := result1.Result.(*Graph)
    g2 := result2.Result.(*Graph)

    if len(g1.Nodes) != len(g2.Nodes) {
        t.Fatalf("same seed produced different graphs")
    }
}
```

### 2.3 Example using testing/quick (Standard Library)

```go
package dungeon_test

import (
    "math/rand"
    "testing"
    "testing/quick"
)

// Graph type (same as above)
type Graph struct {
    Nodes map[int][]int
}

// Implement quick.Generator interface
func (g Graph) Generate(rng *rand.Rand, size int) reflect.Value {
    numNodes := rng.Intn(size) + 2

    graph := Graph{
        Nodes: make(map[int][]int),
    }

    // Initialize
    for i := 0; i < numNodes; i++ {
        graph.Nodes[i] = []int{}
    }

    // Create spanning tree
    for i := 1; i < numNodes; i++ {
        parent := rng.Intn(i)
        graph.Nodes[parent] = append(graph.Nodes[parent], i)
        graph.Nodes[i] = append(graph.Nodes[i], parent)
    }

    return reflect.ValueOf(graph)
}

func TestConnectivityQuick(t *testing.T) {
    config := &quick.Config{
        MaxCount: 100,
        Rand:     rand.New(rand.NewSource(42)), // Seeded for reproducibility
    }

    property := func(g Graph) bool {
        return g.IsConnected()
    }

    if err := quick.Check(property, config); err != nil {
        t.Error(err)
    }
}

// Note: testing/quick requires more boilerplate and doesn't support
// automatic shrinking, making debugging harder when properties fail
```

---

## 3. Comparison Matrix

| Feature | rapid | gopter | testing/quick |
|---------|-------|---------|---------------|
| **Maintenance Status** | Active (2025) | Stable (2020) | Frozen |
| **Go Version** | 1.21+ | 1.13+ | Any |
| **External Dependencies** | None | None | None |
| **API Style** | Modern/Imperative | Functional | Minimal |
| **Automatic Shrinking** | Yes (fully automatic) | Yes (manual) | No |
| **Seed Support** | Yes (`Example(seed)`) | Yes (`Rng.Seed()`) | Yes (`Config.Rand`) |
| **Custom Generators** | `Custom()`, `Deferred()` | Implement `Gen` interface | Implement `Generator` |
| **Built-in Generators** | Good (collections, strings) | Excellent (rich library) | Minimal |
| **State Machine Testing** | Built-in (`Repeat()`) | Via commands package | Manual |
| **Type Safety** | Excellent (generics) | Good | Limited |
| **Learning Curve** | Low | Medium | Low |
| **Documentation** | Good | Excellent | Basic |
| **Community/Examples** | Growing | Mature | Limited |
| **Graph Testing Suitability** | Excellent | Excellent | Fair |
| **Performance** | Fast | Fast (~58μs/test) | Fast |
| **GitHub Stars** | 687+ | 614+ | N/A (stdlib) |
| **Production Usage** | 7,000+ projects | 2,600+ projects | Widespread |
| **Error Messages** | Clear | Clear | Basic |
| **Test Minimization** | Automatic | Semi-automatic | Manual |
| **Constraint Testing** | Excellent | Excellent | Limited |

### Performance Notes

- **gopter:** Demonstrated ~346ms for 20,000 tests (~58μs per test)
- **rapid:** Comparable performance; optimized for edge case exploration
- **testing/quick:** Fastest but least feature-rich

### Use Case Suitability

| Use Case | rapid | gopter | testing/quick |
|----------|-------|---------|---------------|
| Graph connectivity | Excellent | Excellent | Fair |
| Constraint satisfaction | Excellent | Excellent | Limited |
| Complex properties | Excellent | Excellent | Poor |
| State machine testing | Excellent | Excellent | Manual |
| Seed reproducibility | Excellent | Excellent | Good |
| Integration testing | Good | Good | Fair |
| Quick prototyping | Excellent | Good | Excellent |

---

## 4. Rationale for Primary Recommendation

### Why rapid is the Primary Choice

#### 1. Modern Design Philosophy
- **Type-safe generics** - Leverages Go 1.18+ type parameters for compile-time safety
- **Imperative API** - More intuitive for Go developers than functional composition
- **Zero external dependencies** - Reduces supply chain risk and build complexity

#### 2. Automatic Test Case Minimization
Rapid's fully automatic shrinking is a game-changer:
- No manual shrinking code required
- Intelligent minimization based on internal bitstream tracking
- When a test fails, rapid automatically finds the minimal failing case
- Critical for debugging complex constraint violations in dungeon generation

**Example:** If a 50-room dungeon fails connectivity, rapid will automatically shrink to the minimal graph demonstrating the bug (perhaps just 3-4 rooms).

#### 3. Active Maintenance
- Latest release: **February 2025**
- Responsive maintainer (Peter Gregory)
- Adapts to latest Go features
- Growing adoption (7,000+ dependent projects)

#### 4. Superior for Graph Properties
- `Custom()` generator is perfect for building complex graph structures
- `Deferred()` supports recursive graph generators
- `SliceOfDistinct()` ideal for unique node IDs
- Built-in state machine testing via `Repeat()` for testing graph mutations

#### 5. Excellent Developer Experience
- Clear, concise API
- Good error messages
- Comprehensive documentation
- Active community

### When to Choose gopter Instead

Choose **gopter** if:
- Your team is familiar with QuickCheck/ScalaCheck patterns
- You need the most mature, battle-tested solution
- You prefer a rich ecosystem of built-in generators
- You're working on a project that already uses gopter
- You need extensive documentation and examples

### When to Choose testing/quick

Choose **testing/quick** if:
- You need basic property testing without external dependencies
- You're prototyping and don't need shrinking
- Your organization prohibits external dependencies
- You're testing simple invariants that rarely fail

---

## 5. Integration Recommendations for Dungeon Generator

### 5.1 Testing Strategy

Based on the technical specification, here are recommended property tests:

#### Graph Synthesis Phase
```go
// Test: Generated graphs are connected
func TestGraphConnectivity(t *testing.T)

// Test: Degree bounds respected
func TestDegreeConstraints(t *testing.T)

// Test: Key-before-lock reachability
func TestKeyLockOrdering(t *testing.T)

// Test: Start and Boss nodes exist
func TestCriticalNodesExist(t *testing.T)

// Test: Path length within bounds
func TestPathLengthConstraints(t *testing.T)
```

#### Spatial Embedding Phase
```go
// Test: No room overlaps
func TestNoRoomOverlaps(t *testing.T)

// Test: Corridors within max length
func TestCorridorFeasibility(t *testing.T)

// Test: All rooms within grid bounds
func TestSpatialBounds(t *testing.T)
```

#### Constraint Satisfaction
```go
// Test: All hard constraints satisfied
func TestHardConstraintsSatisfied(t *testing.T)

// Test: Soft constraints optimized (scoring)
func TestSoftConstraintScoring(t *testing.T)

// Test: Monotonic difficulty pacing
func TestDifficultyPacing(t *testing.T)
```

#### Determinism
```go
// Test: Same seed produces identical output
func TestDeterministicGeneration(t *testing.T)

// Test: Stage-specific seeds are stable
func TestStageSeedStability(t *testing.T)
```

### 5.2 Recommended Project Structure

```
dungo/
├── go.mod
├── graph/
│   ├── graph.go              # ADG data structures
│   ├── graph_test.go         # Unit tests
│   └── graph_property_test.go # Property-based tests
├── embed/
│   ├── embedder.go
│   └── embed_property_test.go
├── constraints/
│   ├── validator.go
│   └── validator_property_test.go
└── testdata/
    ├── seeds/                # Known good seeds
    └── golden/               # Golden test artifacts
```

### 5.3 go.mod Setup

```go
module github.com/dshills/dungo

go 1.25.3

require (
    pgregory.net/rapid v1.2.0
)
```

### 5.4 Example Property Test Suite

```go
// graph_property_test.go
package graph_test

import (
    "testing"
    "pgregory.net/rapid"
    "github.com/dshills/dungo/graph"
)

// TestGraphInvariants runs all graph property tests
func TestGraphInvariants(t *testing.T) {
    t.Run("connectivity", TestConnectivity)
    t.Run("reachability", TestReachability)
    t.Run("bounds", TestBounds)
    t.Run("degree_constraints", TestDegreeConstraints)
    t.Run("determinism", TestDeterminism)
}

func TestConnectivity(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        config := GenerateConfig(t)
        dungeon := graph.Generate(config)

        if !dungeon.ADG.IsConnected() {
            t.Fatalf("generated disconnected graph with config: %+v", config)
        }
    })
}

func GenerateConfig(t *rapid.T) graph.Config {
    return graph.Config{
        Seed:      rapid.Uint64().Draw(t, "seed"),
        RoomsMin:  rapid.IntRange(10, 35).Draw(t, "roomsMin"),
        RoomsMax:  rapid.IntRange(35, 60).Draw(t, "roomsMax"),
        // ... other config fields
    }
}
```

### 5.5 CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Property Tests

on: [push, pull_request]

jobs:
  property-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run property tests (quick)
        run: go test -rapid.checks=100 ./...

      - name: Run property tests (thorough, nightly)
        if: github.event_name == 'schedule'
        run: go test -rapid.checks=10000 ./...
```

---

## 6. Additional Resources

### Documentation
- **rapid:** https://pkg.go.dev/pgregory.net/rapid
- **gopter:** https://pkg.go.dev/github.com/leanovate/gopter
- **testing/quick:** https://pkg.go.dev/testing/quick

### Articles & Tutorials
- "Property-Based Testing in Go" (DZone): https://dzone.com/articles/property-based-testing-guide-go
- "Property Based Testing" (Gopher Academy): https://blog.gopheracademy.com/advent-2017/property-based-testing/
- "Gopter: Property Based Testing in Golang" (ITNEXT): https://itnext.io/gopter-property-based-testing-in-golang-b36728c7c6d7

### Example Projects
- rapid examples: https://github.com/flyingmutant/rapid/tree/master/example_test.go
- gopter examples: https://github.com/leanovate/gopter/tree/master/example_*_test.go

### Graph Algorithm Libraries (for reference)
- github.com/dominikbraun/graph - Modern Go graph library
- github.com/yourbasic/graph - Simple graph algorithms

---

## 7. Decision Summary

**For the dungo project, use pgregory.net/rapid** because:

1. **Modern Go idioms** - Generics, clean API, type-safe
2. **Automatic shrinking** - Critical for debugging complex graph failures
3. **Active maintenance** - Latest release Feb 2025
4. **Excellent for constraints** - Custom generators perfect for graph properties
5. **No dependencies** - Minimal supply chain risk
6. **Growing community** - 7,000+ projects, active development
7. **Great documentation** - Clear examples, good error messages

**Fallback:** gopter is a solid alternative if rapid doesn't meet specific needs, particularly if the team prefers functional composition or needs more built-in generators.

**Avoid:** testing/quick for this project due to lack of shrinking and limited features for complex constraint testing.

---

## 8. Next Steps

1. Add `pgregory.net/rapid v1.2.0` to go.mod
2. Create initial property test for graph connectivity
3. Implement custom graph generator using `rapid.Custom()`
4. Test key-before-lock constraint satisfaction
5. Verify determinism with seeded generation
6. Integrate property tests into CI/CD pipeline
7. Document failing test cases for regression suite

---

**Document Version:** 1.0
**Last Updated:** 2025-11-04
**Author:** Research Team
**Status:** Final Recommendation
