package synthesis

import (
	"math"

	"github.com/dshills/dungo/pkg/rng"
)

// PacingCurve evaluates a difficulty curve at a given progress point.
// Progress is in [0.0, 1.0] and returns difficulty in [0.0, 1.0].
type PacingCurve interface {
	// Evaluate computes the expected difficulty at the given progress point.
	Evaluate(progress float64) float64
}

// LinearCurve represents a linear difficulty progression.
type LinearCurve struct{}

// Evaluate returns progress as-is for linear progression.
func (c *LinearCurve) Evaluate(progress float64) float64 {
	return clamp(progress)
}

// SCurve represents an S-shaped difficulty curve using a logistic function.
// Provides smooth acceleration and deceleration.
type SCurve struct {
	Steepness float64 // Controls the steepness of the curve (default: 10.0)
}

// NewSCurve creates an S-curve with default steepness.
func NewSCurve() *SCurve {
	return &SCurve{Steepness: 10.0}
}

// Evaluate returns difficulty using a normalized logistic function.
func (c *SCurve) Evaluate(progress float64) float64 {
	progress = clamp(progress)

	// Logistic function: f(x) = 1 / (1 + e^(-k*(x-0.5)))
	// Centered at 0.5 for symmetric S-curve
	k := c.Steepness
	sigmoid := 1.0 / (1.0 + math.Exp(-k*(progress-0.5)))

	// Normalize to [0,1] range
	minVal := 1.0 / (1.0 + math.Exp(k*0.5))
	maxVal := 1.0 / (1.0 + math.Exp(-k*0.5))
	normalized := (sigmoid - minVal) / (maxVal - minVal)

	return clamp(normalized)
}

// ExponentialCurve represents an exponential difficulty progression.
// Provides slow start with rapid increase toward the end.
type ExponentialCurve struct {
	Exponent float64 // Controls steepness (default: 2.0 for quadratic)
}

// NewExponentialCurve creates an exponential curve with default exponent.
func NewExponentialCurve() *ExponentialCurve {
	return &ExponentialCurve{Exponent: 2.0}
}

// Evaluate returns difficulty using power function.
func (c *ExponentialCurve) Evaluate(progress float64) float64 {
	progress = clamp(progress)
	return math.Pow(progress, c.Exponent)
}

// CustomCurve represents a user-defined piecewise linear curve.
// Interpolates between provided control points.
type CustomCurve struct {
	Points [][2]float64 // Control points as [progress, difficulty] pairs
}

// NewCustomCurve creates a custom curve from control points.
// Points must be sorted by progress and have at least 2 entries.
func NewCustomCurve(points [][2]float64) (*CustomCurve, error) {
	if len(points) < 2 {
		return nil, ErrInsufficientPoints
	}

	// Verify points are sorted and in valid range
	for i, point := range points {
		if point[0] < 0.0 || point[0] > 1.0 {
			return nil, ErrInvalidProgress
		}
		if point[1] < 0.0 || point[1] > 1.0 {
			return nil, ErrInvalidDifficulty
		}
		if i > 0 && point[0] <= points[i-1][0] {
			return nil, ErrUnsortedPoints
		}
	}

	return &CustomCurve{Points: points}, nil
}

// Evaluate returns difficulty using linear interpolation between control points.
func (c *CustomCurve) Evaluate(progress float64) float64 {
	progress = clamp(progress)

	if len(c.Points) == 0 {
		return progress // Fallback to linear
	}

	// Before first point
	if progress <= c.Points[0][0] {
		return c.Points[0][1]
	}

	// After last point
	if progress >= c.Points[len(c.Points)-1][0] {
		return c.Points[len(c.Points)-1][1]
	}

	// Find the interval and interpolate
	for i := 0; i < len(c.Points)-1; i++ {
		if progress >= c.Points[i][0] && progress <= c.Points[i+1][0] {
			x0, y0 := c.Points[i][0], c.Points[i][1]
			x1, y1 := c.Points[i+1][0], c.Points[i+1][1]

			// Linear interpolation
			t := (progress - x0) / (x1 - x0)
			return y0 + t*(y1-y0)
		}
	}

	// Shouldn't reach here, but return linear as fallback
	return progress
}

// clamp ensures a value stays within [0.0, 1.0].
func clamp(v float64) float64 {
	if v < 0.0 {
		return 0.0
	}
	if v > 1.0 {
		return 1.0
	}
	return v
}

// EvaluateWithVariance computes difficulty with random variance applied.
// Base difficulty comes from the curve, then variance is added as a random offset.
// variance controls the magnitude: ± variance (clamped to [0.0, 0.3]).
//
// This allows rooms to deviate from the ideal pacing curve, creating more
// natural variation while keeping difficulty roughly on track.
func EvaluateWithVariance(curve PacingCurve, progress float64, variance float64, rng *rng.RNG) float64 {
	// Get base difficulty from curve
	baseDifficulty := curve.Evaluate(progress)

	// Clamp variance to valid range [0.0, 0.3]
	variance = clamp(variance)
	if variance > 0.3 {
		variance = 0.3
	}

	// If variance is effectively zero, return base difficulty
	if variance < 1e-9 {
		return baseDifficulty
	}

	// Apply random offset: [-variance, +variance]
	offset := rng.Float64Range(-variance, variance)
	difficulty := baseDifficulty + offset

	// Clamp final difficulty to [0.0, 1.0]
	return clamp(difficulty)
}

// Error definitions
var (
	ErrInsufficientPoints = &PacingError{Message: "custom curve requires at least 2 points"}
	ErrInvalidProgress    = &PacingError{Message: "progress must be in range [0.0, 1.0]"}
	ErrInvalidDifficulty  = &PacingError{Message: "difficulty must be in range [0.0, 1.0]"}
	ErrUnsortedPoints     = &PacingError{Message: "custom points must be sorted by progress"}
	ErrUnknownCurveType   = &PacingError{Message: "unknown pacing curve type"}
)

// PacingError represents an error in pacing curve configuration.
type PacingError struct {
	Message string
}

func (e *PacingError) Error() string {
	return e.Message
}
