package graph

import (
	"fmt"
	"strings"
)

// ConstraintKind defines the category of constraint being applied.
type ConstraintKind int

const (
	// ConstraintConnectivity ensures graph connectivity properties.
	ConstraintConnectivity ConstraintKind = iota
	// ConstraintDegree enforces connection count constraints.
	ConstraintDegree
	// ConstraintKeyLock ensures keys are obtained before locks.
	ConstraintKeyLock
	// ConstraintPacing enforces difficulty curve progression.
	ConstraintPacing
	// ConstraintTheme ensures thematic consistency.
	ConstraintTheme
	// ConstraintSpatial enforces spatial layout constraints.
	ConstraintSpatial
	// ConstraintCycle controls the number of cycles in the graph.
	ConstraintCycle
	// ConstraintPathLen enforces path length constraints.
	ConstraintPathLen
	// ConstraintBranchFactor controls branching complexity.
	ConstraintBranchFactor
	// ConstraintSecretDensity enforces secret room distribution.
	ConstraintSecretDensity
	// ConstraintOptionality controls optional room distribution.
	ConstraintOptionality
	// ConstraintLootBudget enforces treasure distribution limits.
	ConstraintLootBudget
)

// String returns the string representation of the ConstraintKind.
func (k ConstraintKind) String() string {
	switch k {
	case ConstraintConnectivity:
		return "Connectivity"
	case ConstraintDegree:
		return "Degree"
	case ConstraintKeyLock:
		return "KeyLock"
	case ConstraintPacing:
		return "Pacing"
	case ConstraintTheme:
		return "Theme"
	case ConstraintSpatial:
		return "Spatial"
	case ConstraintCycle:
		return "Cycle"
	case ConstraintPathLen:
		return "PathLen"
	case ConstraintBranchFactor:
		return "BranchFactor"
	case ConstraintSecretDensity:
		return "SecretDensity"
	case ConstraintOptionality:
		return "Optionality"
	case ConstraintLootBudget:
		return "LootBudget"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

// ConstraintSeverity defines the enforcement level of a constraint.
type ConstraintSeverity int

const (
	// SeverityHard means the constraint must be satisfied (generation fails if not).
	SeverityHard ConstraintSeverity = iota
	// SeveritySoft means the constraint should be optimized (best effort).
	SeveritySoft
)

// String returns the string representation of the ConstraintSeverity.
func (s ConstraintSeverity) String() string {
	switch s {
	case SeverityHard:
		return "Hard"
	case SeveritySoft:
		return "Soft"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// Constraint represents a rule that must be satisfied or optimized during generation.
type Constraint struct {
	// Kind is the category of constraint.
	Kind ConstraintKind `json:"kind"`
	// Severity is the enforcement level (Hard = must pass, Soft = optimize).
	Severity ConstraintSeverity `json:"severity"`
	// Expr is the DSL expression defining the constraint.
	Expr string `json:"expr"`
	// Priority determines the order for constraint solving (higher = earlier).
	Priority int `json:"priority"`
}

// Validate checks if the Constraint is well-formed.
func (c *Constraint) Validate() error {
	if c.Expr == "" {
		return fmt.Errorf("constraint expression cannot be empty")
	}

	// Basic syntax validation - full parsing will be done by the constraint solver
	if err := ValidateConstraintExpr(c.Expr); err != nil {
		return fmt.Errorf("invalid constraint expression: %w", err)
	}

	return nil
}

// String returns a human-readable representation of the Constraint.
func (c *Constraint) String() string {
	return fmt.Sprintf("Constraint[%s %s: %s (Priority=%d)]",
		c.Severity, c.Kind, c.Expr, c.Priority)
}

// ValidateConstraintExpr performs basic validation of a constraint DSL expression.
// This is a stub implementation that will be enhanced with full parsing logic.
func ValidateConstraintExpr(expr string) error {
	// Trim whitespace
	expr = strings.TrimSpace(expr)

	// Check for empty expression
	if expr == "" {
		return fmt.Errorf("empty expression")
	}

	// Basic syntax checks
	if !strings.Contains(expr, "(") && !strings.Contains(expr, ")") {
		// If no function call syntax, it might be invalid
		// (though some future predicates might not require parens)
		return fmt.Errorf("expression appears to be malformed (missing function call syntax)")
	}

	// Check for balanced parentheses
	depth := 0
	for _, ch := range expr {
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth < 0 {
				return fmt.Errorf("unbalanced parentheses (too many closing)")
			}
		}
	}
	if depth != 0 {
		return fmt.Errorf("unbalanced parentheses (unclosed)")
	}

	// TODO: Implement full DSL parser with proper tokenization
	// For now, this basic validation is sufficient

	return nil
}

// ParseConstraintExpr parses a constraint DSL expression into an AST.
// This is a stub implementation that will be expanded with full parsing logic.
func ParseConstraintExpr(expr string) (*ConstraintAST, error) {
	// Validate first
	if err := ValidateConstraintExpr(expr); err != nil {
		return nil, err
	}

	// TODO: Implement full recursive descent parser or use a parsing library
	// For now, return a simple stub AST
	ast := &ConstraintAST{
		Type: "function_call",
		Name: extractFunctionName(expr),
		Args: []string{}, // TODO: Extract arguments
	}

	return ast, nil
}

// ConstraintAST represents the abstract syntax tree of a parsed constraint expression.
// This is a simple stub that will be expanded as the DSL grows.
type ConstraintAST struct {
	Type string   // "function_call", "and", "or", "not", "literal"
	Name string   // Function name for function_call type
	Args []string // Arguments (will be more sophisticated in full implementation)
}

// extractFunctionName extracts the function name from a simple function call expression.
// This is a helper for the stub parser.
func extractFunctionName(expr string) string {
	expr = strings.TrimSpace(expr)
	idx := strings.Index(expr, "(")
	if idx == -1 {
		return expr
	}
	return strings.TrimSpace(expr[:idx])
}

// String returns a string representation of the AST.
func (a *ConstraintAST) String() string {
	if a.Type == "function_call" {
		return fmt.Sprintf("%s(%s)", a.Name, strings.Join(a.Args, ", "))
	}
	return fmt.Sprintf("%s: %s", a.Type, a.Name)
}
