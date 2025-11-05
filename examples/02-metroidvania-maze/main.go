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

	fmt.Println("🗺️  Generating Metroidvania-Style Labyrinth...")
	fmt.Println()

	// Load configuration
	cfg, err := dungeon.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Seed: 0x%X\n", cfg.Seed)
	fmt.Printf("Target rooms: %d-%d (Large dungeon!)\n", cfg.Size.RoomsMin, cfg.Size.RoomsMax)
	fmt.Printf("Branching: avg=%.1f, max=%d (Highly interconnected)\n",
		cfg.Branching.Avg, cfg.Branching.Max)
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
	printLabyrinthStats(artifact)

	// Export all formats
	baseName := fmt.Sprintf("metroidvania_maze_%d", cfg.Seed)

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
	opts.Title = "Metroidvania Labyrinth"
	opts.ShowLabels = true
	if err := export.SaveSVGToFile(artifact, absBaseName+".svg", opts); err != nil {
		log.Printf("Warning: Failed to save SVG: %v", err)
	}

	fmt.Printf("\n📁 Files saved: %s.{json,tmj,svg}\n", absBaseName)
	fmt.Println()
	fmt.Println("💡 This dungeon has multiple paths to the boss")
	fmt.Println("💡 Check the SVG to see the complex interconnections")
	fmt.Println("💡 Perfect for exploration-focused gameplay")
}

func printLabyrinthStats(artifact *dungeon.Artifact) {
	g := artifact.ADG.Graph

	fmt.Printf("📊 Labyrinth Stats:\n")
	fmt.Printf("   Rooms: %d\n", len(g.Rooms))
	fmt.Printf("   Corridors: %d\n", len(g.Connectors))

	// Calculate actual connectivity
	if len(g.Rooms) > 0 {
		avgConnectivity := float64(len(g.Connectors)*2) / float64(len(g.Rooms))
		fmt.Printf("   Avg connections/room: %.2f\n", avgConnectivity)
	}

	// Count room types
	roomTypes := make(map[string]int)
	for _, room := range g.Rooms {
		roomTypes[room.Archetype.String()]++
	}

	fmt.Printf("\n🏛️  Room Distribution:\n")
	fmt.Printf("   Hub rooms: %d\n", roomTypes["Hub"])
	fmt.Printf("   Treasure rooms: %d\n", roomTypes["Treasure"])
	fmt.Printf("   Secret rooms: %d\n", roomTypes["Secret"])
	fmt.Printf("   Optional rooms: %d\n", roomTypes["Optional"])

	// Metrics - focus on topology
	if artifact.Metrics != nil {
		fmt.Printf("\n🔄 Topology Analysis:\n")
		fmt.Printf("   Loops/Cycles: %d\n", artifact.Metrics.CycleCount)
		fmt.Printf("   Critical path: %d rooms\n", artifact.Metrics.PathLength)

		// Calculate exploration ratio
		explorationRatio := float64(len(g.Rooms)) / float64(artifact.Metrics.PathLength)
		fmt.Printf("   Exploration ratio: %.1fx (%.0f%% optional)\n",
			explorationRatio, (explorationRatio-1.0)*100)

		if artifact.Metrics.SecretFindability > 0 {
			fmt.Printf("   Secret findability: %.1f%%\n", artifact.Metrics.SecretFindability*100)
		}
	}

	// Content analysis
	if artifact.Content != nil {
		fmt.Printf("\n🎁 Content Distribution:\n")
		fmt.Printf("   Enemy spawns: %d\n", len(artifact.Content.Spawns))
		fmt.Printf("   Loot items: %d\n", len(artifact.Content.Loot))
		fmt.Printf("   Puzzles/Keys: %d\n", len(artifact.Content.Puzzles))
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
	}
}
