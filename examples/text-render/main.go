package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/validation"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║        Dungeon Generator - Text Rendering Example          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load configuration
	cfg, err := dungeon.LoadConfig("examples/configs/demo.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Generating dungeon with seed %d...\n\n", cfg.Seed)

	// Create generator with validator
	validator := validation.NewValidator()
	gen := dungeon.NewGeneratorWithValidator(validator)

	// Generate dungeon
	ctx := context.Background()
	artifact, err := gen.Generate(ctx, cfg)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	// Render full text representation
	fmt.Println(artifact.RenderText())

	// Also show simple graph view
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SIMPLE GRAPH VIEW:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(artifact.RenderTextSimple())
}
