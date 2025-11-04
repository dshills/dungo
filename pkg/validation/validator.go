package validation

import (
	"context"
	"fmt"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// DefaultValidator implements the dungeon.Validator interface with comprehensive checks.
type DefaultValidator struct {
	// Configuration options could be added here in the future
}

// NewValidator creates a new validator with default settings.
// Returns a dungeon.Validator implementation.
func NewValidator() dungeon.Validator {
	return &DefaultValidator{}
}

// Validate performs comprehensive validation of the dungeon artifact.
// It checks all hard and soft constraints, computes metrics, and returns a detailed report.
func (v *DefaultValidator) Validate(ctx context.Context, artifact *dungeon.Artifact, cfg *dungeon.Config) (*dungeon.ValidationReport, error) {
	if artifact == nil {
		return nil, fmt.Errorf("artifact cannot be nil")
	}
	if artifact.ADG == nil || artifact.ADG.Graph == nil {
		return nil, fmt.Errorf("artifact must have a valid graph")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	report := &dungeon.ValidationReport{
		Passed:                true,
		HardConstraintResults: []dungeon.ConstraintResult{},
		SoftConstraintResults: []dungeon.ConstraintResult{},
		Warnings:              []string{},
		Errors:                []string{},
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Step 1: Check hard constraints
	if err := v.checkHardConstraints(ctx, artifact, cfg, report); err != nil {
		return nil, fmt.Errorf("hard constraint checking failed: %w", err)
	}

	// Step 2: Check soft constraints
	if err := v.checkSoftConstraints(ctx, artifact, cfg, report); err != nil {
		return nil, fmt.Errorf("soft constraint checking failed: %w", err)
	}

	// Step 3: Compute metrics
	metrics, err := v.computeMetrics(ctx, artifact, cfg)
	if err != nil {
		return nil, fmt.Errorf("metrics computation failed: %w", err)
	}
	report.Metrics = metrics

	// Set overall pass/fail based on hard constraints
	report.Passed = len(report.Errors) == 0

	return report, nil
}

// checkHardConstraints validates all hard constraints that must be satisfied.
func (v *DefaultValidator) checkHardConstraints(ctx context.Context, artifact *dungeon.Artifact, cfg *dungeon.Config, report *dungeon.ValidationReport) error {
	// Check connectivity
	if result := CheckConnectivity(artifact.ADG.Graph); !result.Satisfied {
		report.Passed = false
		report.Errors = append(report.Errors, result.Details)
		report.HardConstraintResults = append(report.HardConstraintResults, result)
	} else {
		report.HardConstraintResults = append(report.HardConstraintResults, result)
	}

	// Check key reachability
	if result := CheckKeyReachability(artifact.ADG.Graph, cfg); !result.Satisfied {
		report.Passed = false
		report.Errors = append(report.Errors, result.Details)
		report.HardConstraintResults = append(report.HardConstraintResults, result)
	} else {
		report.HardConstraintResults = append(report.HardConstraintResults, result)
	}

	// Check no overlaps (spatial)
	if artifact.Layout != nil {
		if result := CheckNoOverlaps(artifact.ADG.Graph, artifact.Layout); !result.Satisfied {
			report.Passed = false
			report.Errors = append(report.Errors, result.Details)
			report.HardConstraintResults = append(report.HardConstraintResults, result)
		} else {
			report.HardConstraintResults = append(report.HardConstraintResults, result)
		}
	}

	// Check path bounds
	if result := CheckPathBounds(artifact.ADG.Graph, cfg); !result.Satisfied {
		report.Passed = false
		report.Errors = append(report.Errors, result.Details)
		report.HardConstraintResults = append(report.HardConstraintResults, result)
	} else {
		report.HardConstraintResults = append(report.HardConstraintResults, result)
	}

	return nil
}

// checkSoftConstraints validates all soft constraints (optimization targets).
func (v *DefaultValidator) checkSoftConstraints(ctx context.Context, artifact *dungeon.Artifact, cfg *dungeon.Config, report *dungeon.ValidationReport) error {
	// Check pacing deviation
	if result := CheckPacingDeviation(artifact.ADG.Graph, cfg); result.Score < 0.8 {
		report.Warnings = append(report.Warnings, result.Details)
		report.SoftConstraintResults = append(report.SoftConstraintResults, result)
	} else {
		report.SoftConstraintResults = append(report.SoftConstraintResults, result)
	}

	// Check branching factor
	if result := CheckBranchingFactor(artifact.ADG.Graph, cfg); result.Score < 0.8 {
		report.Warnings = append(report.Warnings, result.Details)
		report.SoftConstraintResults = append(report.SoftConstraintResults, result)
	} else {
		report.SoftConstraintResults = append(report.SoftConstraintResults, result)
	}

	return nil
}

// computeMetrics calculates all quality metrics for the dungeon.
func (v *DefaultValidator) computeMetrics(ctx context.Context, artifact *dungeon.Artifact, cfg *dungeon.Config) (*dungeon.Metrics, error) {
	g := artifact.ADG.Graph // This is now *graph.Graph from the embedded field

	metrics := &dungeon.Metrics{
		BranchingFactor:   CalculateBranchingFactor(g),
		PathLength:        CalculatePathLength(g),
		CycleCount:        CountCycles(g),
		PacingDeviation:   CalculatePacingDeviation(g, cfg),
		SecretFindability: 0.0, // TODO: Implement in future iterations
	}

	return metrics, nil
}

// FindStartRoom locates the Start room in the graph.
// Returns the room ID or empty string if not found.
func FindStartRoom(g *graph.Graph) string {
	for id, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			return id
		}
	}
	return ""
}

// FindBossRoom locates the Boss room in the graph.
// Returns the room ID or empty string if not found.
func FindBossRoom(g *graph.Graph) string {
	for id, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeBoss {
			return id
		}
	}
	return ""
}

// FindKeyRooms locates all rooms that provide keys.
// Returns a map of key name to room IDs that provide that key.
func FindKeyRooms(g *graph.Graph) map[string][]string {
	keyRooms := make(map[string][]string)

	for id, room := range g.Rooms {
		for _, cap := range room.Provides {
			if cap.Type == "key" {
				keyRooms[cap.Value] = append(keyRooms[cap.Value], id)
			}
		}
	}

	return keyRooms
}

// FindLockedRooms locates all rooms that require keys.
// Returns a map of key name to room IDs that require that key.
func FindLockedRooms(g *graph.Graph) map[string][]string {
	lockedRooms := make(map[string][]string)

	for id, room := range g.Rooms {
		for _, req := range room.Requirements {
			if req.Type == "key" {
				lockedRooms[req.Value] = append(lockedRooms[req.Value], id)
			}
		}
	}

	return lockedRooms
}
