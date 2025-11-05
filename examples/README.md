# Dungeon Generator Examples

This directory contains working examples demonstrating how to use the dungeon generator.

## 🎮 Fun Game-Style Examples

Three complete examples showcasing different dungeon generation styles:

| Example | Style | Size | Difficulty | Run Command |
|---------|-------|------|------------|-------------|
| **01-zelda-dungeon** 🗝️ | Linear progression | 20-25 rooms | S-curve | `cd 01-zelda-dungeon && go run main.go` |
| **02-metroidvania-maze** 🗺️ | Open exploration | 60-80 rooms | Linear | `cd 02-metroidvania-maze && go run main.go` |
| **03-darksouls-challenge** ⚔️ | Brutal mid-game | 30-40 rooms | Custom | `cd 03-darksouls-challenge && go run main.go` |

### 1. Classic Zelda-Style Dungeon 🗝️

**Path**: `01-zelda-dungeon/`

Traditional dungeon with key/lock progression, treasure rooms, and structured difficulty.

**Features**:
- 3 key types (small, big, boss)
- S-curve pacing (easy → hard → fair boss)
- 15% secret rooms, 20% optional content
- Key-before-lock constraint enforcement

**Perfect for**: Roguelikes, puzzle dungeons, structured campaigns

### 2. Metroidvania-Style Labyrinth 🗺️

**Path**: `02-metroidvania-maze/`

Massive, interconnected dungeon with loops and multiple paths to every objective.

**Features**:
- 60-80 rooms (3x larger!)
- High connectivity (2.8 avg connections/room)
- 25% secret rooms, 35% optional content
- Multiple viable routes to boss

**Perfect for**: Exploration games, backtracking mechanics, speedruns

### 3. Dark Souls Challenge Dungeon ⚔️

**Path**: `03-darksouls-challenge/`

Brutally difficult with custom curve that spikes at mid-game.

**Features**:
- Custom difficulty (20% → 85% → 75%)
- Most content mandatory (85% critical path)
- High variance (±25% unpredictability)
- Strategic secret placement

**Perfect for**: Challenge runs, permadeath, skill testing

Each example generates `.json`, `.tmj` (Tiled), and `.svg` (visualization) files.

## 🛠️ Technical Examples

### 4. Text Rendering Example

**Path**: `text-render/main.go`

Demonstrates text-based dungeon visualization in the terminal.

**Run**:
```bash
go run examples/text-render/main.go
```

**Shows**:
- Room statistics and graph structure
- Emoji-based room type visualization
- ASCII art tile map preview
- Content summary (enemies, loot, keys)
- Key-lock relationships
- Validation status

### 5. Embedding Example

**Path**: `embedding/main.go`

Demonstrates spatial layout with the force-directed embedding algorithm.

**Run**:
```bash
go run examples/embedding/main.go
```

**Shows**:
- Graph creation with rooms and connections
- Spatial layout generation
- Room positions and dimensions
- Overlap prevention

## Configuration Files

Example configurations are in the `configs/` directory:

- `demo.yaml` - Simple 10-15 room dungeon with one key
- `basic_dungeon.yaml` - Standard configuration
- `custom_pacing.yaml` - Advanced pacing curve example

## Quick Start

**New users?** Start with the game-style examples to see the library in action:

```bash
# Classic Zelda-style dungeon (recommended first!)
cd examples/01-zelda-dungeon
go run main.go

# Or use make from project root
make run-zelda
make run-metroidvania
make run-darksouls

# Or build binaries
make examples
./bin/zelda-dungeon -config examples/01-zelda-dungeon/config.yaml
```

**Technical exploration?** Try the rendering and embedding examples:

```bash
# Text rendering
go run examples/text-render/main.go

# Spatial embedding
go run examples/embedding/main.go
```

## Creating Your Own Examples

Copy one of the existing examples and modify the configuration:

```go
package main

import (
    "context"
    "fmt"
    "github.com/dshills/dungo/pkg/dungeon"
    "github.com/dshills/dungo/pkg/validation"
)

func main() {
    // Load your config
    cfg, _ := dungeon.LoadConfig("path/to/your/config.yaml")

    // Create generator
    gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())

    // Generate
    artifact, _ := gen.Generate(context.Background(), cfg)

    // Visualize
    fmt.Println(artifact.RenderText())
}
```

## Next Steps

- See `../testdata/seeds/` for more configuration examples
- Check `../specs/001-dungeon-generator-core/quickstart.md` for full API guide
- Read `../CLAUDE.md` for development guidance
