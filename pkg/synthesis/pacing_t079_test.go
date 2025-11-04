package synthesis

import (
	"math"
	"testing"
)

// TestT079_SCurveDifficultyDistribution verifies S-curve difficulty distribution
// as specified in Task T079.
//
// This test validates that the S-curve produces the expected smooth acceleration
// and deceleration pattern with:
// - Slow start (early rooms easier)
// - Rapid middle transition (steepest part of curve)
// - Gentle ending (approaching max difficulty)
func TestT079_SCurveDifficultyDistribution(t *testing.T) {
	curve := NewSCurve() // Default steepness of 10.0

	tests := []struct {
		name        string
		progress    float64
		expectation string
		minDiff     float64
		maxDiff     float64
	}{
		{
			name:        "Start: very easy",
			progress:    0.0,
			expectation: "near zero difficulty",
			minDiff:     0.0,
			maxDiff:     0.01,
		},
		{
			name:        "Early quarter: still mostly easy",
			progress:    0.25,
			expectation: "low difficulty, slow increase",
			minDiff:     0.0,
			maxDiff:     0.15,
		},
		{
			name:        "Midpoint: exactly half difficulty",
			progress:    0.5,
			expectation: "symmetric midpoint",
			minDiff:     0.48,
			maxDiff:     0.52,
		},
		{
			name:        "Late quarter: mostly hard",
			progress:    0.75,
			expectation: "high difficulty, slowing increase",
			minDiff:     0.85,
			maxDiff:     1.0,
		},
		{
			name:        "End: maximum difficulty",
			progress:    1.0,
			expectation: "at or near max difficulty",
			minDiff:     0.99,
			maxDiff:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			difficulty := curve.Evaluate(tt.progress)

			if difficulty < tt.minDiff || difficulty > tt.maxDiff {
				t.Errorf("At progress=%v (%s): difficulty=%v, expected range [%v, %v]",
					tt.progress, tt.expectation, difficulty, tt.minDiff, tt.maxDiff)
			}

			t.Logf("✓ progress=%v → difficulty=%0.3f (%s)",
				tt.progress, difficulty, tt.expectation)
		})
	}
}

// TestT079_SCurveVsLinearComparison demonstrates that S-curve differs from linear.
func TestT079_SCurveVsLinearComparison(t *testing.T) {
	scurve := NewSCurve()
	linear := &LinearCurve{}

	// S-curve should be BELOW linear in early game
	earlyProgress := 0.3
	earlyS := scurve.Evaluate(earlyProgress)
	earlyLinear := linear.Evaluate(earlyProgress)

	if earlyS >= earlyLinear {
		t.Errorf("S-curve should be below linear at %v: S=%v, Linear=%v",
			earlyProgress, earlyS, earlyLinear)
	}
	t.Logf("✓ Early game (0.3): S-curve=%0.3f < Linear=%0.3f (easier start)",
		earlyS, earlyLinear)

	// S-curve should be ABOVE linear in late game
	lateProgress := 0.7
	lateS := scurve.Evaluate(lateProgress)
	lateLinear := linear.Evaluate(lateProgress)

	if lateS <= lateLinear {
		t.Errorf("S-curve should be above linear at %v: S=%v, Linear=%v",
			lateProgress, lateS, lateLinear)
	}
	t.Logf("✓ Late game (0.7): S-curve=%0.3f > Linear=%0.3f (steeper ramp)",
		lateS, lateLinear)
}

// TestT079_SCurveAccelerationRate verifies the S-curve has maximum
// acceleration around the midpoint.
func TestT079_SCurveAccelerationRate(t *testing.T) {
	curve := NewSCurve()

	// Calculate rate of change at different points
	epsilon := 0.01

	// Early rate of change
	earlyDiff := curve.Evaluate(0.2+epsilon) - curve.Evaluate(0.2-epsilon)
	earlyRate := earlyDiff / (2 * epsilon)

	// Middle rate of change
	midDiff := curve.Evaluate(0.5+epsilon) - curve.Evaluate(0.5-epsilon)
	midRate := midDiff / (2 * epsilon)

	// Late rate of change
	lateDiff := curve.Evaluate(0.8+epsilon) - curve.Evaluate(0.8-epsilon)
	lateRate := lateDiff / (2 * epsilon)

	// Middle should have highest rate of change
	if midRate <= earlyRate {
		t.Errorf("Middle rate (%v) should exceed early rate (%v)", midRate, earlyRate)
	}
	if midRate <= lateRate {
		t.Errorf("Middle rate (%v) should exceed late rate (%v)", midRate, lateRate)
	}

	t.Logf("✓ Rate of change: early=%0.3f, middle=%0.3f (max), late=%0.3f",
		earlyRate, midRate, lateRate)

	// Early and late rates should be similar (symmetry)
	if math.Abs(earlyRate-lateRate) > 0.1 {
		t.Errorf("S-curve not symmetric: early rate=%v, late rate=%v", earlyRate, lateRate)
	}
	t.Logf("✓ Symmetric: early and late rates similar (diff=%0.4f)",
		math.Abs(earlyRate-lateRate))
}

// TestT079_AllCurveTypesWork verifies all required curve types are implemented.
func TestT079_AllCurveTypesWork(t *testing.T) {
	testProgress := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	t.Run("LINEAR", func(t *testing.T) {
		curve := &LinearCurve{}
		for _, p := range testProgress {
			d := curve.Evaluate(p)
			if math.Abs(d-p) > 1e-9 {
				t.Errorf("Linear curve at %v: got %v, want %v", p, d, p)
			}
		}
		t.Log("✓ LINEAR curve working")
	})

	t.Run("S_CURVE", func(t *testing.T) {
		curve := NewSCurve()
		prev := -1.0
		for _, p := range testProgress {
			d := curve.Evaluate(p)
			if d < prev {
				t.Errorf("S-curve not monotonic: %v < %v", d, prev)
			}
			if d < 0.0 || d > 1.0 {
				t.Errorf("S-curve out of range: %v at progress %v", d, p)
			}
			prev = d
		}
		t.Log("✓ S_CURVE working")
	})

	t.Run("EXPONENTIAL", func(t *testing.T) {
		curve := NewExponentialCurve()
		prev := -1.0
		for _, p := range testProgress {
			d := curve.Evaluate(p)
			if d < prev {
				t.Errorf("Exponential curve not monotonic: %v < %v", d, prev)
			}
			if d < 0.0 || d > 1.0 {
				t.Errorf("Exponential curve out of range: %v at progress %v", d, p)
			}
			prev = d
		}
		t.Log("✓ EXPONENTIAL curve working")
	})

	t.Run("CUSTOM", func(t *testing.T) {
		points := [][2]float64{
			{0.0, 0.0},
			{0.5, 0.3},
			{1.0, 1.0},
		}
		curve, err := NewCustomCurve(points)
		if err != nil {
			t.Fatalf("NewCustomCurve failed: %v", err)
		}

		prev := -1.0
		for _, p := range testProgress {
			d := curve.Evaluate(p)
			if d < prev {
				t.Errorf("Custom curve not monotonic: %v < %v", d, prev)
			}
			if d < 0.0 || d > 1.0 {
				t.Errorf("Custom curve out of range: %v at progress %v", d, p)
			}
			prev = d
		}
		t.Log("✓ CUSTOM curve working")
	})
}
