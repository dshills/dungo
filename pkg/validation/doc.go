// Package validation provides constraint checking and metrics calculation for dungeon generation.
//
// This package implements the validation stage (Stage 5) of the dungeon generation pipeline.
// It validates hard constraints (must pass), soft constraints (optimization targets),
// and computes quality metrics for generated dungeons.
//
// # Hard Constraints
//
// Hard constraints must be satisfied for a dungeon to be considered valid:
//
//   - Connectivity: All rooms must be reachable from any starting point
//   - Key Reachability: Keys must be obtainable before their locks
//   - No Overlaps: Rooms must not overlap in spatial layout
//   - Path Bounds: Start-to-Boss path must be within reasonable length
//
// # Soft Constraints
//
// Soft constraints are optimization targets that should be met but won't fail generation:
//
//   - Pacing Deviation: Difficulty should follow the configured curve
//   - Branching Factor: Connectivity should match target average
//
// # Metrics
//
// The validation stage computes several quality metrics:
//
//   - Branching Factor: Average connections per room
//   - Path Length: Shortest path from Start to Boss
//   - Cycle Count: Number of cycles in the graph
//   - Pacing Deviation: L2 distance from target difficulty curve
//   - Secret Findability: Heuristic score for secret discoverability (future)
//
// # Usage Example
//
//	validator := validation.NewValidator()
//	report, err := validator.Validate(ctx, artifact, cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if !report.Passed {
//	    log.Printf("Validation failed: %v", report.Errors)
//	}
//
//	log.Printf("Metrics: %+v", report.Metrics)
//	log.Printf("Summary:\n%s", validation.Summary(report))
package validation
