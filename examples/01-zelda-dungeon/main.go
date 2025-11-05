package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/export"
	"github.com/dshills/dungo/pkg/validation"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
)

func main() {
	flag.Parse()

	fmt.Println("🗝️  Generating Classic Zelda-Style Dungeon...")
	fmt.Println()

	// Load configuration
	cfg, err := dungeon.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Seed: 0x%X\n", cfg.Seed)
	fmt.Printf("Target rooms: %d-%d\n", cfg.Size.RoomsMin, cfg.Size.RoomsMax)
	fmt.Printf("Keys: %d types\n", len(cfg.Keys))
	fmt.Println()

	// Create generator with validator
	validator := validation.NewValidator()
	gen := dungeon.NewGeneratorWithValidator(validator)

	// Generate dungeon
	ctx := context.Background()
	artifact, err := gen.Generate(ctx, cfg)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	// Print statistics
	fmt.Println("✨ Generation Complete!")
	fmt.Println()
	printDungeonStats(artifact)

	// Export all formats to current directory
	baseName := fmt.Sprintf("zelda_dungeon_%d", cfg.Seed)

	// Get absolute paths for output
	absBaseName, err := filepath.Abs(baseName)
	if err != nil {
		absBaseName = baseName
	}

	if err := export.SaveJSONToFile(artifact, absBaseName+".json"); err != nil {
		log.Printf("Warning: Failed to save JSON: %v", err)
	}

	if err := export.SaveArtifactToTMJFile(artifact, absBaseName+".tmj", true); err != nil {
		log.Printf("Warning: Failed to save TMJ: %v", err)
	}

	opts := export.DefaultSVGOptions()
	opts.Title = "Classic Zelda Dungeon"
	if err := export.SaveSVGToFile(artifact, absBaseName+".svg", opts); err != nil {
		log.Printf("Warning: Failed to save SVG: %v", err)
	}

	fmt.Printf("\n📁 Files saved: %s.{json,tmj,svg}\n", absBaseName)
	fmt.Println()
	fmt.Println("💡 Try changing the seed in config.yaml for different layouts!")
	fmt.Println("💡 Open the .svg file to visualize the dungeon graph")
	fmt.Println("💡 Import the .tmj file into Tiled Map Editor")
}

func printDungeonStats(artifact *dungeon.Artifact) {
	g := artifact.ADG.Graph

	fmt.Printf("📊 Dungeon Stats:\n")
	fmt.Printf("   Rooms: %d\n", len(g.Rooms))
	fmt.Printf("   Corridors: %d\n", len(g.Connectors))

	// Count room types
	roomTypes := make(map[string]int)
	for _, room := range g.Rooms {
		roomTypes[room.Archetype.String()]++
	}

	fmt.Printf("\n🚪 Room Types:\n")
	for roomType, count := range roomTypes {
		fmt.Printf("   %s: %d\n", roomType, count)
	}

	// Count keys and locks
	if artifact.Content != nil {
		keyCount := 0
		lockCount := 0
		for _, puzzle := range artifact.Content.Puzzles {
			if puzzle.Type == "key" {
				keyCount++
			} else if puzzle.Type == "lock" {
				lockCount++
			}
		}
		fmt.Printf("\n🗝️  Key/Lock System:\n")
		fmt.Printf("   Keys: %d\n", keyCount)
		fmt.Printf("   Locks: %d\n", lockCount)
		fmt.Printf("   Secrets: %d\n", len(artifact.Content.Secrets))
	}

	// Validation status
	if artifact.Debug != nil && artifact.Debug.Report != nil {
		report := artifact.Debug.Report
		status := "✓ PASSED"
		if !report.Passed {
			status = "✗ FAILED"
		}
		fmt.Printf("\n✅ Validation: %s\n", status)

		if len(report.Warnings) > 0 {
			fmt.Printf("   ⚠️  Warnings: %d\n", len(report.Warnings))
		}
		if len(report.Errors) > 0 {
			fmt.Printf("   ❌ Errors: %d\n", len(report.Errors))
		}
	}

	// Metrics
	if artifact.Metrics != nil {
		fmt.Printf("\n📈 Quality Metrics:\n")
		fmt.Printf("   Branching Factor: %.2f\n", artifact.Metrics.BranchingFactor)
		fmt.Printf("   Critical Path: %d rooms\n", artifact.Metrics.PathLength)
		fmt.Printf("   Loops: %d\n", artifact.Metrics.CycleCount)
		fmt.Printf("   Pacing Score: %.1f%%\n", (1.0-artifact.Metrics.PacingDeviation)*100)
	}
}
