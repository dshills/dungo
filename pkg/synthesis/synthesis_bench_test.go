package synthesis

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/rng"
)

// BenchmarkGrammarSynthesis benchmarks the grammar-based synthesizer
// for different dungeon sizes to verify performance targets.
func BenchmarkGrammarSynthesis(b *testing.B) {
	tests := []struct {
		name     string
		roomsMin int
		roomsMax int
	}{
		{"Small_25-35_rooms", 25, 35},
		{"Medium_50-70_rooms", 50, 70},
		{"Large_80-100_rooms", 80, 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create config
			cfg := &Config{
				Seed:          12345,
				RoomsMin:      tt.roomsMin,
				RoomsMax:      tt.roomsMax,
				BranchingAvg:  2.0,
				BranchingMax:  4,
				SecretDensity: 0.15,
				OptionalRatio: 0.25,
				Keys: []KeyConfig{
					{Name: "key_red", Count: 1},
					{Name: "key_blue", Count: 1},
				},
				Pacing: PacingConfig{
					Curve:    "S_CURVE",
					Variance: 0.15,
				},
				Themes: []string{"dungeon", "crypt", "cavern"},
			}

			// Get synthesizer
			synthesizer := Get("grammar")
			if synthesizer == nil {
				b.Fatal("grammar synthesizer not registered")
			}

			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use different seed for each iteration to avoid caching effects
				cfg.Seed = uint64(12345 + i)
				configHash := []byte("benchmark")
				rngInst := rng.NewRNG(cfg.Seed, "synthesis", configHash)

				g, err := synthesizer.Synthesize(ctx, rngInst, cfg)
				if err != nil {
					b.Fatalf("synthesis failed: %v", err)
				}

				// Verify we got a reasonable graph
				if len(g.Rooms) < tt.roomsMin || len(g.Rooms) > tt.roomsMax {
					b.Fatalf("unexpected room count: got %d, want [%d, %d]",
						len(g.Rooms), tt.roomsMin, tt.roomsMax)
				}
			}
		})
	}
}

// BenchmarkGrammarSynthesis_60Room benchmarks the specific target size
// mentioned in performance requirements (60-room dungeon).
func BenchmarkGrammarSynthesis_60Room(b *testing.B) {
	cfg := &Config{
		Seed:          12345,
		RoomsMin:      60,
		RoomsMax:      60,
		BranchingAvg:  2.0,
		BranchingMax:  4,
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Keys: []KeyConfig{
			{Name: "key_red", Count: 1},
			{Name: "key_blue", Count: 1},
		},
		Pacing: PacingConfig{
			Curve:    "S_CURVE",
			Variance: 0.15,
		},
		Themes: []string{"dungeon", "crypt", "cavern"},
	}

	synthesizer := Get("grammar")
	if synthesizer == nil {
		b.Fatal("grammar synthesizer not registered")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg.Seed = uint64(12345 + i)
		configHash := []byte("benchmark")
		rngInst := rng.NewRNG(cfg.Seed, "synthesis", configHash)

		g, err := synthesizer.Synthesize(ctx, rngInst, cfg)
		if err != nil {
			b.Fatalf("synthesis failed: %v", err)
		}

		if len(g.Rooms) != 60 {
			b.Fatalf("unexpected room count: got %d, want 60", len(g.Rooms))
		}
	}
}

// BenchmarkPacingAssignment benchmarks just the pacing assignment phase
// to measure overhead of difficulty curve calculation.
func BenchmarkPacingAssignment(b *testing.B) {
	tests := []struct {
		name      string
		curve     string
		roomCount int
	}{
		{"Linear_25rooms", "LINEAR", 25},
		{"SCurve_50rooms", "S_CURVE", 50},
		{"Exponential_70rooms", "EXPONENTIAL", 70},
		{"Custom_60rooms", "CUSTOM", 60},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cfg := &Config{
				Seed:          12345,
				RoomsMin:      tt.roomCount,
				RoomsMax:      tt.roomCount,
				BranchingAvg:  2.0,
				BranchingMax:  4,
				SecretDensity: 0.15,
				OptionalRatio: 0.25,
				Pacing: PacingConfig{
					Curve:    tt.curve,
					Variance: 0.15,
					// Add custom points for CUSTOM curve
					CustomPoints: [][2]float64{
						{0.0, 0.1},
						{0.3, 0.3},
						{0.7, 0.8},
						{1.0, 1.0},
					},
				},
				Themes: []string{"dungeon"},
			}

			synthesizer := Get("grammar")
			if synthesizer == nil {
				b.Fatal("grammar synthesizer not registered")
			}

			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cfg.Seed = uint64(12345 + i)
				configHash := []byte("benchmark")
				rngInst := rng.NewRNG(cfg.Seed, "synthesis", configHash)

				_, err := synthesizer.Synthesize(ctx, rngInst, cfg)
				if err != nil {
					b.Fatalf("synthesis failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkThemeAssignment benchmarks theme biome assignment.
func BenchmarkThemeAssignment(b *testing.B) {
	tests := []struct {
		name       string
		roomCount  int
		themeCount int
	}{
		{"1theme_25rooms", 25, 1},
		{"3themes_50rooms", 50, 3},
		{"5themes_70rooms", 70, 5},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			themes := make([]string, tt.themeCount)
			for i := 0; i < tt.themeCount; i++ {
				themes[i] = string(rune('A' + i)) // A, B, C, etc.
			}

			cfg := &Config{
				Seed:          12345,
				RoomsMin:      tt.roomCount,
				RoomsMax:      tt.roomCount,
				BranchingAvg:  2.0,
				BranchingMax:  4,
				SecretDensity: 0.15,
				OptionalRatio: 0.25,
				Pacing: PacingConfig{
					Curve:    "LINEAR",
					Variance: 0.15,
				},
				Themes: themes,
			}

			synthesizer := Get("grammar")
			if synthesizer == nil {
				b.Fatal("grammar synthesizer not registered")
			}

			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cfg.Seed = uint64(12345 + i)
				configHash := []byte("benchmark")
				rngInst := rng.NewRNG(cfg.Seed, "synthesis", configHash)

				_, err := synthesizer.Synthesize(ctx, rngInst, cfg)
				if err != nil {
					b.Fatalf("synthesis failed: %v", err)
				}
			}
		})
	}
}
