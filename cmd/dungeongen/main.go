package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/export"
	"github.com/dshills/dungo/pkg/validation"
)

const (
	version = "1.0.0"
)

// CLI flags
var (
	configPath = flag.String("config", "", "Path to YAML configuration file (required)")
	outputDir  = flag.String("output", ".", "Output directory for generated files")
	format     = flag.String("format", "json", "Export format: json, tmj, svg, or all")
	seedFlag   = flag.Uint64("seed", 0, "Override the seed from config (0 = use config seed)")
	verbose    = flag.Bool("verbose", false, "Enable verbose output")
	versionF   = flag.Bool("version", false, "Print version and exit")
	help       = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	// Handle version flag
	if *versionF {
		fmt.Printf("dungeongen version %s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if *help {
		printHelp()
		os.Exit(0)
	}

	// Validate required flags
	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -config flag is required")
		printUsage()
		os.Exit(1)
	}

	// Validate format
	validFormats := map[string]bool{
		"json": true,
		"tmj":  true,
		"svg":  true,
		"all":  true,
	}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q, must be one of: json, tmj, svg, all\n", *format)
		os.Exit(1)
	}

	// Run the generator
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// nolint:gocyclo // Complexity acceptable: CLI argument handling and output formatting
func run() error {
	ctx := context.Background()

	// Load configuration
	if *verbose {
		fmt.Printf("Loading configuration from %s\n", *configPath)
	}

	cfg, err := dungeon.LoadConfig(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override seed if specified
	if *seedFlag != 0 {
		if *verbose {
			fmt.Printf("Overriding seed from %d to %d\n", cfg.Seed, *seedFlag)
		}
		cfg.Seed = *seedFlag
	}

	if *verbose {
		fmt.Printf("Using seed: %d\n", cfg.Seed)
		fmt.Printf("Room count: %d-%d\n", cfg.Size.RoomsMin, cfg.Size.RoomsMax)
		fmt.Printf("Themes: %v\n", cfg.Themes)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create generator with validator
	validator := validation.NewValidator()
	gen := dungeon.NewGeneratorWithValidator(validator)

	// Generate dungeon
	start := time.Now()
	if *verbose {
		fmt.Println("Generating dungeon...")
	}

	artifact, err := gen.Generate(ctx, cfg)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	elapsed := time.Since(start)
	if *verbose {
		fmt.Printf("Generation completed in %v\n", elapsed)
		printStats(artifact)
	}

	// Export to requested format(s)
	baseName := fmt.Sprintf("dungeon_%d", cfg.Seed)

	if *format == "json" || *format == "all" {
		if err := exportJSON(artifact, baseName); err != nil {
			return err
		}
	}

	if *format == "tmj" || *format == "all" {
		if err := exportTMJ(artifact, baseName); err != nil {
			return err
		}
	}

	if *format == "svg" || *format == "all" {
		if err := exportSVG(artifact, baseName); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully generated dungeon (seed=%d) in %v\n", cfg.Seed, elapsed)
	return nil
}

// exportJSON exports the artifact to JSON format
func exportJSON(artifact *dungeon.Artifact, baseName string) error {
	filename := filepath.Join(*outputDir, baseName+".json")
	if *verbose {
		fmt.Printf("Exporting JSON to %s\n", filename)
	}

	if err := export.SaveJSONToFile(artifact, filename); err != nil {
		return fmt.Errorf("failed to export JSON: %w", err)
	}

	if *verbose {
		info, _ := os.Stat(filename)
		fmt.Printf("  Wrote %d bytes\n", info.Size())
	}

	return nil
}

// exportTMJ exports the artifact to Tiled Map JSON format
func exportTMJ(artifact *dungeon.Artifact, baseName string) error {
	filename := filepath.Join(*outputDir, baseName+".tmj")
	if *verbose {
		fmt.Printf("Exporting TMJ to %s\n", filename)
	}

	// Compress tile data for efficiency
	compress := true
	if err := export.SaveArtifactToTMJFile(artifact, filename, compress); err != nil {
		return fmt.Errorf("failed to export TMJ: %w", err)
	}

	if *verbose {
		info, _ := os.Stat(filename)
		fmt.Printf("  Wrote %d bytes\n", info.Size())
	}

	return nil
}

// exportSVG exports the artifact to SVG visualization format
func exportSVG(artifact *dungeon.Artifact, baseName string) error {
	filename := filepath.Join(*outputDir, baseName+".svg")
	if *verbose {
		fmt.Printf("Exporting SVG to %s\n", filename)
	}

	opts := export.DefaultSVGOptions()
	opts.Title = fmt.Sprintf("Dungeon (seed=%d)", artifact.ADG.Graph.Seed)

	if err := export.SaveSVGToFile(artifact, filename, opts); err != nil {
		return fmt.Errorf("failed to export SVG: %w", err)
	}

	if *verbose {
		info, _ := os.Stat(filename)
		fmt.Printf("  Wrote %d bytes\n", info.Size())
	}

	return nil
}

// printStats prints dungeon statistics
func printStats(artifact *dungeon.Artifact) {
	fmt.Println("\nDungeon Statistics:")
	fmt.Printf("  Rooms: %d\n", len(artifact.ADG.Graph.Rooms))
	fmt.Printf("  Connectors: %d\n", len(artifact.ADG.Graph.Connectors))

	if artifact.TileMap != nil {
		fmt.Printf("  Tile Map: %dx%d tiles\n", artifact.TileMap.Width, artifact.TileMap.Height)
	}

	if artifact.Content != nil {
		fmt.Printf("  Spawns: %d\n", len(artifact.Content.Spawns))
		fmt.Printf("  Loot: %d\n", len(artifact.Content.Loot))
		fmt.Printf("  Puzzles: %d\n", len(artifact.Content.Puzzles))
		fmt.Printf("  Secrets: %d\n", len(artifact.Content.Secrets))
	}

	if artifact.Metrics != nil {
		fmt.Println("\nMetrics:")
		fmt.Printf("  BranchingFactor: %.3f\n", artifact.Metrics.BranchingFactor)
		fmt.Printf("  PathLength: %d\n", artifact.Metrics.PathLength)
		fmt.Printf("  CycleCount: %d\n", artifact.Metrics.CycleCount)
		fmt.Printf("  PacingDeviation: %.3f\n", artifact.Metrics.PacingDeviation)
		fmt.Printf("  SecretFindability: %.3f\n", artifact.Metrics.SecretFindability)
	}

	if artifact.Debug != nil && artifact.Debug.Report != nil {
		report := artifact.Debug.Report
		fmt.Printf("\nValidation: %s\n", validationStatus(report.Passed))
		if len(report.Warnings) > 0 {
			fmt.Printf("  Warnings: %d\n", len(report.Warnings))
		}
		if len(report.Errors) > 0 {
			fmt.Printf("  Errors: %d\n", len(report.Errors))
		}
	}
}

// validationStatus returns a colored status string
func validationStatus(passed bool) string {
	if passed {
		return "✓ PASSED"
	}
	return "✗ FAILED"
}

// printUsage prints basic usage information
func printUsage() {
	fmt.Fprintln(os.Stderr, "\nUsage: dungeongen -config <config.yaml> [options]")
	fmt.Fprintln(os.Stderr, "\nRun 'dungeongen -help' for detailed help")
}

// printHelp prints detailed help information
func printHelp() {
	fmt.Printf("dungeongen version %s\n\n", version)
	fmt.Println("A command-line tool for generating procedural dungeons.")
	fmt.Println("\nUsage:")
	fmt.Println("  dungeongen -config <config.yaml> [options]")
	fmt.Println("\nRequired Flags:")
	fmt.Println("  -config string")
	fmt.Println("        Path to YAML configuration file")
	fmt.Println("\nOptional Flags:")
	fmt.Println("  -output string")
	fmt.Println("        Output directory for generated files (default: current directory)")
	fmt.Println("  -format string")
	fmt.Println("        Export format: json, tmj, svg, or all (default: json)")
	fmt.Println("  -seed uint")
	fmt.Println("        Override the seed from config (0 = use config seed) (default: 0)")
	fmt.Println("  -verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -version")
	fmt.Println("        Print version and exit")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  # Generate dungeon with default JSON export")
	fmt.Println("  dungeongen -config dungeon.yaml")
	fmt.Println("\n  # Generate with custom seed and all export formats")
	fmt.Println("  dungeongen -config dungeon.yaml -seed 12345 -format all -output ./out")
	fmt.Println("\n  # Generate SVG visualization with verbose output")
	fmt.Println("  dungeongen -config dungeon.yaml -format svg -verbose")
	fmt.Println("\nConfiguration File:")
	fmt.Println("  The YAML configuration file specifies dungeon parameters including:")
	fmt.Println("  - Seed (for deterministic generation)")
	fmt.Println("  - Size constraints (roomsMin, roomsMax)")
	fmt.Println("  - Branching parameters (avg, max connections)")
	fmt.Println("  - Pacing curve (LINEAR, S_CURVE, EXPONENTIAL, CUSTOM)")
	fmt.Println("  - Themes (crypt, fungal, arcane, etc.)")
	fmt.Println("  - Keys/locks, constraints, and more")
	fmt.Println("\n  See the project documentation for detailed configuration schema.")
}
