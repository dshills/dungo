# Dungeon Generator Examples

This directory contains working examples demonstrating how to use the dungeon generator.

## Available Examples

### 1. Text Rendering Example

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

### 2. Embedding Example

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

## Running Examples

All examples can be run directly with `go run`:

```bash
# Text rendering (recommended to start)
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
