package validation

import (
	"fmt"
	"strings"

	"github.com/dshills/dungo/pkg/dungeon"
)

// NewValidationReport creates a new empty validation report.
func NewValidationReport() *dungeon.ValidationReport {
	return &dungeon.ValidationReport{
		Passed:                true,
		HardConstraintResults: []dungeon.ConstraintResult{},
		SoftConstraintResults: []dungeon.ConstraintResult{},
		Warnings:              []string{},
		Errors:                []string{},
	}
}

// NewConstraintResult creates a constraint result for a specific constraint check.
func NewConstraintResult(constraint *dungeon.Constraint, satisfied bool, score float64, details string) dungeon.ConstraintResult {
	return dungeon.ConstraintResult{
		Constraint: constraint,
		Satisfied:  satisfied,
		Score:      score,
		Details:    details,
	}
}

// NewHardConstraintResult creates a result for a hard constraint.
// Hard constraints are pass/fail (score is 1.0 or 0.0).
func NewHardConstraintResult(kind, expr string, satisfied bool, details string) dungeon.ConstraintResult {
	score := 0.0
	if satisfied {
		score = 1.0
	}

	return dungeon.ConstraintResult{
		Constraint: &dungeon.Constraint{
			Kind:     kind,
			Severity: "hard",
			Expr:     expr,
		},
		Satisfied: satisfied,
		Score:     score,
		Details:   details,
	}
}

// NewSoftConstraintResult creates a result for a soft constraint.
// Soft constraints have a continuous score from 0.0 to 1.0.
func NewSoftConstraintResult(kind, expr string, score float64, details string) dungeon.ConstraintResult {
	return dungeon.ConstraintResult{
		Constraint: &dungeon.Constraint{
			Kind:     kind,
			Severity: "soft",
			Expr:     expr,
		},
		Satisfied: score > 0.5, // Consider satisfied if score > 0.5
		Score:     score,
		Details:   details,
	}
}

// Summary returns a human-readable summary of the validation report.
func Summary(report *dungeon.ValidationReport) string {
	var b strings.Builder

	b.WriteString("=== Validation Report ===\n\n")

	// Overall status
	if report.Passed {
		b.WriteString("Status: PASSED\n")
	} else {
		b.WriteString("Status: FAILED\n")
	}

	// Metrics summary
	if report.Metrics != nil {
		b.WriteString("\n=== Metrics ===\n")
		b.WriteString(fmt.Sprintf("Branching Factor: %.2f\n", report.Metrics.BranchingFactor))
		b.WriteString(fmt.Sprintf("Path Length: %d\n", report.Metrics.PathLength))
		b.WriteString(fmt.Sprintf("Cycle Count: %d\n", report.Metrics.CycleCount))
		b.WriteString(fmt.Sprintf("Pacing Deviation: %.3f\n", report.Metrics.PacingDeviation))
		b.WriteString(fmt.Sprintf("Secret Findability: %.2f\n", report.Metrics.SecretFindability))
	}

	// Hard constraints
	b.WriteString("\n=== Hard Constraints ===\n")
	passedHard := 0
	for _, result := range report.HardConstraintResults {
		if result.Satisfied {
			passedHard++
		}
	}
	b.WriteString(fmt.Sprintf("Passed: %d/%d\n", passedHard, len(report.HardConstraintResults)))

	for i, result := range report.HardConstraintResults {
		status := "PASS"
		if !result.Satisfied {
			status = "FAIL"
		}
		b.WriteString(fmt.Sprintf("  %d. [%s] %s: %s\n", i+1, status, result.Constraint.Kind, result.Details))
	}

	// Soft constraints
	b.WriteString("\n=== Soft Constraints ===\n")
	if len(report.SoftConstraintResults) == 0 {
		b.WriteString("None evaluated\n")
	} else {
		for i, result := range report.SoftConstraintResults {
			b.WriteString(fmt.Sprintf("  %d. %s (score: %.2f): %s\n",
				i+1, result.Constraint.Kind, result.Score, result.Details))
		}
	}

	// Errors
	if len(report.Errors) > 0 {
		b.WriteString("\n=== Errors ===\n")
		for i, err := range report.Errors {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err))
		}
	}

	// Warnings
	if len(report.Warnings) > 0 {
		b.WriteString("\n=== Warnings ===\n")
		for i, warn := range report.Warnings {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, warn))
		}
	}

	return b.String()
}

// HasErrors returns true if the report contains any hard constraint failures.
func HasErrors(report *dungeon.ValidationReport) bool {
	return len(report.Errors) > 0
}

// HasWarnings returns true if the report contains any soft constraint warnings.
func HasWarnings(report *dungeon.ValidationReport) bool {
	return len(report.Warnings) > 0
}

// GetFailedConstraints returns all failed hard constraints.
func GetFailedConstraints(report *dungeon.ValidationReport) []dungeon.ConstraintResult {
	failed := []dungeon.ConstraintResult{}
	for _, result := range report.HardConstraintResults {
		if !result.Satisfied {
			failed = append(failed, result)
		}
	}
	return failed
}

// GetLowScoringConstraints returns soft constraints with score below threshold.
func GetLowScoringConstraints(report *dungeon.ValidationReport, threshold float64) []dungeon.ConstraintResult {
	lowScoring := []dungeon.ConstraintResult{}
	for _, result := range report.SoftConstraintResults {
		if result.Score < threshold {
			lowScoring = append(lowScoring, result)
		}
	}
	return lowScoring
}
