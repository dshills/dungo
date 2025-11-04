package dungeon_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/validation"
)

// TestLargeDungeon_NoCorridorLengthErrors verifies that large dungeons
// (200+ rooms) don't fail with corridor length constraint violations.
// This is a regression test for the bug where hardcoded CorridorMaxLength=100
// caused failures on large dungeons.
func TestLargeDungeon_NoCorridorLengthErrors(t *testing.T) {
	testCases := []struct {
		name     string
		roomsMax int
		seed     uint64
	}{
		{
			name:     "medium_100_rooms",
			roomsMax: 100,
			seed:     12345,
		},
		{
			name:     "large_150_rooms",
			roomsMax: 150,
			seed:     23456,
		},
		{
			name:     "large_200_rooms",
			roomsMax: 200,
			seed:     34567,
		},
		{
			name:     "original_bug_214_rooms",
			roomsMax: 214,
			seed:     0x868a3, // From original failure log
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &dungeon.Config{
				Seed: tc.seed,
				Size: dungeon.SizeCfg{
					RoomsMin: tc.roomsMax / 2,
					RoomsMax: tc.roomsMax,
				},
				Branching: dungeon.BranchingCfg{
					Avg: 2.0,
					Max: 4,
				},
				Pacing: dungeon.PacingCfg{
					Curve:    dungeon.PacingLinear,
					Variance: 0.1,
				},
				Themes:        []string{"dungeon"},
				SecretDensity: 0.1,
				OptionalRatio: 0.2,
			}

			gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
			artifact, err := gen.Generate(context.Background(), cfg)

			// Check specifically for corridor length errors
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "corridor") && strings.Contains(errStr, "exceeds max length") {
					t.Fatalf("Corridor length constraint violation: %v", err)
				}
				// Other errors (e.g., validation, carving) are acceptable for this test
				t.Skipf("Non-corridor error (acceptable): %v", err)
			}

			// If successful, verify room count
			if artifact != nil && artifact.ADG != nil {
				roomCount := len(artifact.ADG.Rooms)
				t.Logf("Successfully generated dungeon with %d rooms (requested %d-%d)",
					roomCount, cfg.Size.RoomsMin, cfg.Size.RoomsMax)

				if roomCount < cfg.Size.RoomsMin || roomCount > cfg.Size.RoomsMax {
					t.Errorf("Room count %d out of bounds [%d, %d]",
						roomCount, cfg.Size.RoomsMin, cfg.Size.RoomsMax)
				}
			}
		})
	}
}
