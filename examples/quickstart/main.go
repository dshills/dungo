package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/validation"
)

func main() {
	// Load configuration - use an existing test config
	cfg, err := dungeon.LoadConfig("../../testdata/seeds/small_crypt.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Create generator with validator
	validator := validation.NewValidator()
	gen := dungeon.NewGeneratorWithValidator(validator)

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
	fmt.Printf("  Rooms: %d\n", len(artifact.ADG.Graph.Rooms))
	fmt.Printf("  Connections: %d\n", len(artifact.ADG.Graph.Connectors))
	fmt.Printf("  Path to boss: %d rooms\n", artifact.Metrics.PathLength)
	fmt.Printf("  Branching factor: %.2f\n", artifact.Metrics.BranchingFactor)

	// Export to JSON
	data, _ := json.MarshalIndent(artifact, "", "  ")
	os.MkdirAll("output", 0755)
	os.WriteFile("output/dungeon.json", data, 0644)

	fmt.Println("Dungeon saved to output/dungeon.json")
}
