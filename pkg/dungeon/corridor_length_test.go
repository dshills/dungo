package dungeon

import (
	"math"
	"testing"
)

func TestCalculateCorridorMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		roomCount int
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "zero rooms",
			roomCount: 0,
			wantMin:   100.0,
			wantMax:   100.0,
		},
		{
			name:      "small dungeon (25 rooms)",
			roomCount: 25,
			wantMin:   295.0,
			wantMax:   295.0,
		},
		{
			name:      "medium dungeon (50 rooms)",
			roomCount: 50,
			wantMin:   417.0,
			wantMax:   418.0,
		},
		{
			name:      "medium dungeon (69 rooms)",
			roomCount: 69,
			wantMin:   490.0,
			wantMax:   491.0,
		},
		{
			name:      "large dungeon (100 rooms)",
			roomCount: 100,
			wantMin:   590.0,
			wantMax:   590.0,
		},
		{
			name:      "large dungeon (200 rooms)",
			roomCount: 200,
			wantMin:   600.0,
			wantMax:   600.0,
		},
		{
			name:      "very large dungeon (214 rooms) - original bug report",
			roomCount: 214,
			wantMin:   600.0,
			wantMax:   600.0,
		},
		{
			name:      "huge dungeon (625 rooms) - should hit max",
			roomCount: 625,
			wantMin:   600.0,
			wantMax:   600.0,
		},
		{
			name:      "massive dungeon (1000 rooms) - should hit max",
			roomCount: 1000,
			wantMin:   600.0,
			wantMax:   600.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCorridorMaxLength(tt.roomCount)

			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateCorridorMaxLength(%d) = %.1f, want in range [%.1f, %.1f]",
					tt.roomCount, got, tt.wantMin, tt.wantMax)
			}

			// Verify it follows the sqrt(N) * 59 formula (within bounds)
			expected := math.Sqrt(float64(tt.roomCount)) * 59.0
			if expected < 100.0 {
				expected = 100.0
			}
			if expected > 600.0 {
				expected = 600.0
			}

			if math.Abs(got-expected) > 0.1 {
				t.Errorf("calculateCorridorMaxLength(%d) = %.1f, formula gives %.1f (diff > 0.1)",
					tt.roomCount, got, expected)
			}
		})
	}
}

// TestCorridorMaxLength_ScalingProperty verifies that corridor max length
// increases monotonically with room count (except when hitting the max cap).
func TestCorridorMaxLength_ScalingProperty(t *testing.T) {
	prev := calculateCorridorMaxLength(1)

	for rooms := 10; rooms <= 500; rooms += 10 {
		curr := calculateCorridorMaxLength(rooms)

		// Should either increase or stay at cap
		if curr < prev {
			t.Errorf("calculateCorridorMaxLength(%d) = %.1f < previous %.1f (not monotonic)",
				rooms, curr, prev)
		}

		prev = curr
	}
}
