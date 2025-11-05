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

	fmt.Println("⚔️  Generating Dark Souls Challenge Dungeon...")
	fmt.Println("    (You Died)")
	fmt.Println()

	// Load configuration
	cfg, err := dungeon.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Seed: 0x%X (The cursed seed...)\n", cfg.Seed)
	fmt.Printf("Target rooms: %d-%d\n", cfg.Size.RoomsMin, cfg.Size.RoomsMax)
	fmt.Printf("Pacing: CUSTOM curve (brutal mid-game)\n")
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
	fmt.Println("✨ Challenge Dungeon Generated!")
	fmt.Println()
	printChallengeStats(artifact, cfg)

	// Export all formats
	baseName := fmt.Sprintf("darksouls_challenge_%d", cfg.Seed)

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
	opts.Title = "Dark Souls Challenge Dungeon"
	opts.ShowLabels = true
	if err := export.SaveSVGToFile(artifact, absBaseName+".svg", opts); err != nil {
		log.Printf("Warning: Failed to save SVG: %v", err)
	}

	fmt.Printf("\n📁 Files saved: %s.{json,tmj,svg}\n", absBaseName)
	fmt.Println()
	fmt.Println("⚠️  WARNING: This dungeon is brutally difficult")
	fmt.Println("💀 Peak difficulty at 40-60% progression")
	fmt.Println("🎯 Finding secrets is crucial for survival")
	fmt.Println()
	fmt.Println("💡 Try seed 666 vs 777 for different challenges")
	fmt.Println("💡 Check the SVG to plan your route carefully")
}

func printChallengeStats(artifact *dungeon.Artifact, cfg *dungeon.Config) {
	g := artifact.ADG.Graph

	fmt.Printf("📊 Challenge Stats:\n")
	fmt.Printf("   Rooms: %d\n", len(g.Rooms))
	fmt.Printf("   Corridors: %d\n", len(g.Connectors))

	// Show custom difficulty curve
	if cfg.Pacing.Curve == dungeon.PacingCustom && len(cfg.Pacing.CustomPoints) > 0 {
		fmt.Printf("\n📈 Difficulty Curve:\n")
		for i, point := range cfg.Pacing.CustomPoints {
			progress := int(point[0] * 100)
			difficulty := int(point[1] * 100)
			bar := generateDifficultyBar(point[1], 20)
			fmt.Printf("   %3d%% progress: %s %d%%\n", progress, bar, difficulty)

			// Highlight the brutal section
			if i > 0 && point[1] >= 0.8 {
				fmt.Printf("        └─ ☠️  BRUTAL SECTION\n")
			}
		}
	}

	// Critical path analysis
	if artifact.Metrics != nil {
		fmt.Printf("\n⚔️  Combat Analysis:\n")
		fmt.Printf("   Critical path: %d rooms\n", artifact.Metrics.PathLength)
		fmt.Printf("   Branching factor: %.2f\n", artifact.Metrics.BranchingFactor)

		// Estimate difficulty
		avgDifficulty := 0.0
		if len(cfg.Pacing.CustomPoints) > 0 {
			for _, point := range cfg.Pacing.CustomPoints {
				avgDifficulty += point[1]
			}
			avgDifficulty /= float64(len(cfg.Pacing.CustomPoints))
			fmt.Printf("   Average difficulty: %.0f%%\n", avgDifficulty*100)
		}

		// Pacing quality
		pacingScore := (1.0 - artifact.Metrics.PacingDeviation) * 100
		fmt.Printf("   Pacing adherence: %.1f%%\n", pacingScore)
	}

	// Content distribution
	if artifact.Content != nil {
		fmt.Printf("\n🗡️  Encounters:\n")
		fmt.Printf("   Enemy spawns: %d\n", len(artifact.Content.Spawns))
		fmt.Printf("   Loot drops: %d\n", len(artifact.Content.Loot))
		fmt.Printf("   Keys/Locks: %d\n", len(artifact.Content.Puzzles))
		fmt.Printf("   Hidden secrets: %d\n", len(artifact.Content.Secrets))

		if len(artifact.Content.Secrets) > 0 {
			fmt.Printf("        └─ 💎 Secrets provide crucial advantages!\n")
		}
	}

	// Validation
	if artifact.Debug != nil && artifact.Debug.Report != nil {
		report := artifact.Debug.Report
		status := "✓ PASSED"
		if !report.Passed {
			status = "✗ FAILED"
		}
		fmt.Printf("\n✅ Validation: %s\n", status)
	}
}

// generateDifficultyBar creates a visual bar for difficulty
func generateDifficultyBar(difficulty float64, width int) string {
	filled := int(difficulty * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			if difficulty >= 0.8 {
				bar += "█" // Brutal
			} else if difficulty >= 0.5 {
				bar += "▓" // Hard
			} else {
				bar += "░" // Easy
			}
		} else {
			bar += "·"
		}
	}

	return bar
}
