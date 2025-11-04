# Quickstart: Graph-Based Dungeon Generator

**Feature**: 001-dungeon-generator-core
**Date**: 2025-11-04
**Audience**: Developers integrating the dungeon generator into their game

## Overview

This guide walks you through using the dungeon generator to create your first procedurally generated dungeon. The generator produces complete dungeons with rooms, connections, tile maps, and gameplay content from a seed and configuration file.

---

## Installation

```bash
# Add dependency to your project
go get github.com/dshills/dungo

# Or add to go.mod manually:
# require github.com/dshills/dungo v0.1.0
```

---

## Quick Example: Generate Your First Dungeon

### Step 1: Create a Configuration File

Save as `dungeons/small_crypt.yaml`:

```yaml
seed: 12345
size:
  roomsMin: 25
  roomsMax: 35
branching:
  avg: 1.7
  max: 3
pacing:
  curve: "S_CURVE"
  variance: 0.15
themes:
  - crypt
keys:
  - name: silver
    count: 1
  - name: gold
    count: 1
secretDensity: 0.15
optionalRatio: 0.20
constraints:
  - kind: Connectivity
    severity: Hard
    expr: "isConnected()"
  - kind: KeyLock
    severity: Hard
    expr: "keyBeforeLock('silver')"
  - kind: KeyLock
    severity: Hard
    expr: "keyBeforeLock('gold')"
  - kind: Pacing
    severity: Soft
    expr: "monotoneIncrease(start, boss, 0.6)"
```

### Step 2: Generate the Dungeon

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dshills/dungo/pkg/dungeon"
)

func main() {
	// Load configuration
	cfg, err := dungeon.LoadConfig("dungeons/small_crypt.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Create generator
	gen := dungeon.NewGenerator()

	// Generate dungeon
	ctx := context.Background()
	artifact, err := gen.Generate(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Check validation
	if !artifact.Debug.Report.Passed {
		fmt.Println("Warning: Dungeon has constraint violations")
		for _, err := range artifact.Debug.Report.Errors {
			fmt.Println("  -", err)
		}
	}

	// Print statistics
	fmt.Printf("Generated dungeon with:\n")
	fmt.Printf("  Rooms: %d\n", len(artifact.ADG.Rooms))
	fmt.Printf("  Connections: %d\n", len(artifact.ADG.Connectors))
	fmt.Printf("  Path to boss: %d rooms\n", artifact.Metrics.PathLength)
	fmt.Printf("  Branching factor: %.2f\n", artifact.Metrics.BranchingFactor)

	// Export to JSON
	data, _ := json.MarshalIndent(artifact, "", "  ")
	os.WriteFile("output/dungeon.json", data, 0644)

	fmt.Println("Dungeon saved to output/dungeon.json")
}
```

### Step 3: Run the Generator

```bash
mkdir -p output
go run main.go

# Output:
# Generated dungeon with:
#   Rooms: 32
#   Connections: 54
#   Path to boss: 12 rooms
#   Branching factor: 1.69
# Dungeon saved to output/dungeon.json
```

---

## Export Formats

### Export to Tiled TMJ Format

```go
import "github.com/dshills/dungo/pkg/export"

// Export tile map for game engine
exporter := export.NewExporter()
tmjData, err := exporter.ExportTMJ(artifact.TileMap)
if err != nil {
	log.Fatal(err)
}

os.WriteFile("output/dungeon.tmj", tmjData, 0644)
```

Load in Tiled Editor: `File → Open → output/dungeon.tmj`

### Generate Debug Visualization (SVG)

```go
// SVG visualization of room connectivity
svgData, err := exporter.ExportSVG(artifact.ADG, export.SVGOptions{
	Width:       1200,
	Height:      800,
	ShowLabels:  true,
	ColorByType: true,
	ShowHeatmap: true,
	ShowLegend:  true,
})
if err != nil {
	log.Fatal(err)
}

os.WriteFile("output/dungeon_debug.svg", svgData, 0644)
```

Open in browser to see the room graph with color-coded nodes.

---

## Configuration Options

### Dungeon Size

Control the number of rooms:

```yaml
size:
  roomsMin: 10    # Minimum: 10
  roomsMax: 100   # Maximum: 300
```

Small dungeons (10-30 rooms): Quick, linear, good for tutorials
Medium dungeons (40-80 rooms): Balanced, branching, standard gameplay
Large dungeons (100-300 rooms): Complex, non-linear, late-game content

### Branching Complexity

Control how rooms connect:

```yaml
branching:
  avg: 1.5   # Linear (mostly corridors)
  avg: 1.8   # Moderate branching (some choices)
  avg: 2.5   # High branching (many paths)
  max: 3     # Maximum connections per room
```

### Difficulty Pacing

Shape how difficulty increases:

```yaml
pacing:
  curve: "LINEAR"      # Steady increase
  curve: "S_CURVE"     # Slow start, steep middle, plateau
  curve: "EXPONENTIAL" # Rapid difficulty spike
  variance: 0.15       # Allowed deviation (0.0-0.3)
```

Custom pacing curve:

```yaml
pacing:
  curve: "CUSTOM"
  customPoints:
    - [0.0, 0.1]  # Start easy
    - [0.3, 0.4]  # Gradual rise
    - [0.7, 0.8]  # Spike
    - [1.0, 1.0]  # Max at boss
```

### Themes and Biomes

Use single or multiple themes:

```yaml
# Single theme
themes:
  - crypt

# Multiple themes (smooth transitions)
themes:
  - crypt
  - fungal
  - arcane
```

Available themes (v1):
- `crypt` - Undead, stone, gothic
- `fungal` - Mushrooms, spores, caves
- `arcane` - Magic, crystals, ruins

### Keys and Locks

Create gated progression:

```yaml
keys:
  - name: silver
    count: 2   # Two silver keys in dungeon
  - name: gold
    count: 1   # One gold key (rare)
```

Keys are automatically placed before their locks.

### Secret and Optional Content

Control exploration density:

```yaml
secretDensity: 0.15   # 15% of rooms are hidden/secret
optionalRatio: 0.20   # 20% of rooms are optional (off critical path)
```

---

## Deterministic Generation

### Same Seed = Same Dungeon

```go
cfg1 := dungeon.Config{Seed: 99999, /* ... */}
cfg2 := dungeon.Config{Seed: 99999, /* ... */}

artifact1, _ := gen.Generate(ctx, cfg1)
artifact2, _ := gen.Generate(ctx, cfg2)

// artifact1 and artifact2 are byte-for-byte identical
```

### Seed Management

```go
// Auto-generate seed if not specified
cfg := dungeon.Config{Seed: 0} // Will use time-based seed

// Use player-provided seed for sharing
cfg := dungeon.Config{Seed: 12345} // Players can share "12345"

// Generate from string (e.g., level name)
import "hash/fnv"

func seedFromString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

cfg := dungeon.Config{Seed: seedFromString("Forest-Temple-1")}
```

---

## Accessing Generated Data

### Room Information

```go
for id, room := range artifact.ADG.Rooms {
	fmt.Printf("Room %s: %s (%s)\n", id, room.Archetype, room.Size)
	fmt.Printf("  Difficulty: %.2f\n", room.Difficulty)
	fmt.Printf("  Reward: %.2f\n", room.Reward)

	// Check if room provides a key
	for _, cap := range room.Provides {
		if cap.Type == "key" {
			fmt.Printf("  Provides: %s\n", cap.Value)
		}
	}
}
```

### Connections and Paths

```go
for id, conn := range artifact.ADG.Connectors {
	fmt.Printf("Connection %s: %s → %s (%s)\n",
		id, conn.From, conn.To, conn.Type)

	// Check if connection is gated
	if conn.Gate != nil {
		fmt.Printf("  Requires: %s (%s)\n",
			conn.Gate.Value, conn.Gate.Type)
	}
}
```

### Tile Map Layers

```go
tileMap := artifact.TileMap

// Access floor tiles
floorLayer := tileMap.Layers["floor"]
for y := 0; y < tileMap.Height; y++ {
	for x := 0; x < tileMap.Width; x++ {
		tileID := floorLayer.Data[y*tileMap.Width + x]
		if tileID != 0 {
			// Place floor tile at (x, y)
		}
	}
}

// Access entities (object layer)
entityLayer := tileMap.Layers["entities"]
for _, obj := range entityLayer.Objects {
	if obj.Type == "SpawnPoint" {
		// Spawn enemy at (obj.X, obj.Y)
	}
}
```

### Enemy Spawns and Loot

```go
// Process enemy spawns
for _, spawn := range artifact.Content.Spawns {
	roomID := spawn.RoomID
	enemyType := spawn.EnemyType
	position := spawn.Position

	// Spawn 'spawn.Count' enemies of 'enemyType' at 'position' in room 'roomID'
	fmt.Printf("Spawn %d x %s at room %s (%d, %d)\n",
		spawn.Count, enemyType, roomID, position.X, position.Y)
}

// Process loot
for _, loot := range artifact.Content.Loot {
	if loot.Required {
		// This is a key or required item
		fmt.Printf("Required item '%s' in room %s\n", loot.ItemType, loot.RoomID)
	} else {
		// Optional treasure
		fmt.Printf("Treasure '%s' worth %d gold in room %s\n",
			loot.ItemType, loot.Value, loot.RoomID)
	}
}
```

---

## Custom Theme Packs

### Create Your Own Theme

1. Create theme directory: `themes/mypack/`
2. Create theme manifest: `themes/mypack/theme.yaml`

```yaml
name: mypack
tilesets:
  floor: tiles/mypack_floor.png
  walls: tiles/mypack_walls.png
  decor: tiles/mypack_decor.png
encounterTables:
  - difficulty: 0.3
    entries:
      - type: "goblin"
        weight: 10
      - type: "orc"
        weight: 5
  - difficulty: 0.7
    entries:
      - type: "troll"
        weight: 8
      - type: "ogre"
        weight: 3
lootTables:
  - difficulty: 0.5
    entries:
      - type: "gold_pile"
        amount: "50-100"
        weight: 10
      - type: "health_potion"
        weight: 5
decorators:
  - condition: "room.archetype == 'Treasure'"
    actions:
      - "place_chest"
      - "place_sparkles"
```

3. Use in config:

```yaml
themes:
  - mypack
```

---

## Testing and Validation

### Check Constraint Satisfaction

```go
report := artifact.Debug.Report

if !report.Passed {
	fmt.Println("Hard constraints failed:")
	for _, result := range report.HardConstraintResults {
		if !result.Satisfied {
			fmt.Printf("  - %s: %s\n",
				result.Constraint.Expr, result.Details)
		}
	}
}

// Check soft constraint optimization
for _, result := range report.SoftConstraintResults {
	fmt.Printf("Soft constraint '%s': score %.2f\n",
		result.Constraint.Expr, result.Score)
}
```

### Generate Test Dungeons

```go
// Generate 100 dungeons and collect statistics
seeds := make([]uint64, 100)
for i := range seeds {
	seeds[i] = uint64(i * 1000)
}

avgRooms := 0.0
avgBranching := 0.0

for _, seed := range seeds {
	cfg.Seed = seed
	artifact, err := gen.Generate(ctx, cfg)
	if err != nil {
		continue
	}

	avgRooms += float64(len(artifact.ADG.Rooms))
	avgBranching += artifact.Metrics.BranchingFactor
}

avgRooms /= 100
avgBranching /= 100

fmt.Printf("Across 100 dungeons:\n")
fmt.Printf("  Average rooms: %.1f\n", avgRooms)
fmt.Printf("  Average branching: %.2f\n", avgBranching)
```

---

## Performance Tips

### Reuse Generator Instance

```go
// Create once, reuse for multiple generations
gen := dungeon.NewGenerator()

for i := 0; i < 1000; i++ {
	cfg.Seed = uint64(i)
	artifact, _ := gen.Generate(ctx, cfg)
	// Process artifact
}
```

### Limit Room Count for Speed

Small dungeons generate faster:

```yaml
# Fast: < 10ms
size: { roomsMin: 10, roomsMax: 20 }

# Medium: < 50ms
size: { roomsMin: 30, roomsMax: 60 }

# Large: < 200ms
size: { roomsMin: 80, roomsMax: 150 }
```

### Cancel Long Generations

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

artifact, err := gen.Generate(ctx, cfg)
if err == context.DeadlineExceeded {
	fmt.Println("Generation took too long, using fallback dungeon")
}
```

---

## Common Patterns

### Daily Challenge Dungeon

```go
// Generate deterministic "daily dungeon" from date
func dailySeed() uint64 {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	return seedFromString(dateStr)
}

cfg := dungeon.Config{
	Seed: dailySeed(),
	// ... other settings
}
```

### Player-Seeded Dungeon

```go
// Let players enter a seed code
func generateFromPlayerCode(code string) (*dungeon.Artifact, error) {
	seed := seedFromString(code)
	cfg := dungeon.Config{Seed: seed, /* ... */}
	return gen.Generate(context.Background(), cfg)
}

// Player enters "MySecretDungeon"
artifact, _ := generateFromPlayerCode("MySecretDungeon")
```

### Difficulty Tiers

```go
easyConfig := dungeon.Config{
	Size: dungeon.SizeCfg{RoomsMin: 15, RoomsMax: 25},
	Pacing: dungeon.PacingCfg{Curve: "LINEAR", Variance: 0.1},
	// ...
}

hardConfig := dungeon.Config{
	Size: dungeon.SizeCfg{RoomsMin: 60, RoomsMax: 100},
	Pacing: dungeon.PacingCfg{Curve: "EXPONENTIAL", Variance: 0.25},
	// ...
}
```

---

## Next Steps

1. **Customize Configuration**: Experiment with different pacing curves, themes, and room counts
2. **Create Theme Packs**: Design your own encounter tables and loot distributions
3. **Integrate with Game Engine**: Parse TMJ files or use JSON artifact data
4. **Add Custom Constraints**: Extend the DSL with game-specific rules
5. **Benchmark Performance**: Test with your target hardware and room counts

---

## Troubleshooting

### "Hard constraint failed: not all rooms reachable"

- Reduce `branching.avg` to ensure better connectivity
- Increase `roomsMax` to give more space for path finding
- Check that `keys` configuration doesn't create impossible gates

### "Generation timeout"

- Reduce `roomsMax` for faster generation
- Simplify constraints (fewer hard constraints)
- Use simpler themes (fewer encounter table entries)

### "Tileset not found: crypt"

- Ensure theme packs exist in `themes/` directory
- Check `theme.yaml` manifest file
- Verify theme name matches config

### "Key-before-lock constraint violated"

- Ensure `keyBeforeLock` constraint is marked as `Hard` severity
- Check that key count matches lock count in config
- Verify key names are consistent (case-sensitive)

---

## Support

- Documentation: [https://github.com/dshills/dungo/docs](https://github.com/dshills/dungo/docs)
- Examples: [https://github.com/dshills/dungo/examples](https://github.com/dshills/dungo/examples)
- Issues: [https://github.com/dshills/dungo/issues](https://github.com/dshills/dungo/issues)

---

**Ready to generate dungeons? Start with the quick example above and explore from there!**
