package dungeon_test

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/validation"
	"pgregory.net/rapid"
)

// TestProperty_RoomCountBounds is a property-based test that verifies
// generated dungeons have room counts within Config.Size bounds.
// This test verifies the synthesis stage is working correctly.
func TestProperty_RoomCountBounds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random but valid configuration parameters
		roomsMin := rapid.IntRange(10, 100).Draw(t, "roomsMin")
		roomsMax := rapid.IntRange(roomsMin, 300).Draw(t, "roomsMax")

		branchingAvg := rapid.Float64Range(1.5, 3.0).Draw(t, "branchingAvg")
		branchingMax := rapid.IntRange(2, 5).Draw(t, "branchingMax")

		pacingVariance := rapid.Float64Range(0.0, 0.3).Draw(t, "pacingVariance")

		secretDensity := rapid.Float64Range(0.0, 0.3).Draw(t, "secretDensity")
		optionalRatio := rapid.Float64Range(0.1, 0.4).Draw(t, "optionalRatio")

		// Build configuration
		cfg := &dungeon.Config{
			Seed: rapid.Uint64().Draw(t, "seed"),
			Size: dungeon.SizeCfg{
				RoomsMin: roomsMin,
				RoomsMax: roomsMax,
			},
			Branching: dungeon.BranchingCfg{
				Avg: branchingAvg,
				Max: branchingMax,
			},
			Pacing: dungeon.PacingCfg{
				Curve:    dungeon.PacingLinear,
				Variance: pacingVariance,
			},
			Themes:        []string{"dungeon"},
			SecretDensity: secretDensity,
			OptionalRatio: optionalRatio,
		}

		// Generate dungeon
		gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
		artifact, err := gen.Generate(context.Background(), cfg)

		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		if artifact == nil || artifact.ADG == nil {
			t.Fatal("Generate() returned nil artifact or ADG")
		}

		// Property: Room count must be within configured bounds
		roomCount := len(artifact.ADG.Rooms)
		if roomCount < cfg.Size.RoomsMin {
			t.Fatalf("Room count %d is less than minimum %d", roomCount, cfg.Size.RoomsMin)
		}
		if roomCount > cfg.Size.RoomsMax {
			t.Fatalf("Room count %d exceeds maximum %d", roomCount, cfg.Size.RoomsMax)
		}
	})
}

// TestProperty_StartAndBossRooms is a property-based test that verifies
// all synthesized dungeons must have exactly one Start room and at least one Boss room.
// This is a fundamental requirement for any dungeon graph.
func TestProperty_StartAndBossRooms(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random but valid configuration
		cfg := &dungeon.Config{
			Seed: rapid.Uint64().Draw(t, "seed"),
			Size: dungeon.SizeCfg{
				RoomsMin: rapid.IntRange(10, 50).Draw(t, "roomsMin"),
				RoomsMax: rapid.IntRange(50, 100).Draw(t, "roomsMax"),
			},
			Branching: dungeon.BranchingCfg{
				Avg: rapid.Float64Range(1.5, 3.0).Draw(t, "branchingAvg"),
				Max: rapid.IntRange(2, 5).Draw(t, "branchingMax"),
			},
			Pacing: dungeon.PacingCfg{
				Curve:    dungeon.PacingLinear,
				Variance: rapid.Float64Range(0.0, 0.3).Draw(t, "pacingVariance"),
			},
			Themes:        []string{"dungeon"},
			SecretDensity: rapid.Float64Range(0.0, 0.3).Draw(t, "secretDensity"),
			OptionalRatio: rapid.Float64Range(0.1, 0.4).Draw(t, "optionalRatio"),
		}

		// Generate dungeon
		gen := dungeon.NewGeneratorWithValidator(validation.NewValidator())
		artifact, err := gen.Generate(context.Background(), cfg)

		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		if artifact == nil || artifact.ADG == nil {
			t.Fatal("Generate() returned nil artifact or ADG")
		}

		g := artifact.ADG.Graph

		// Count rooms by archetype
		startCount := 0
		bossCount := 0
		var startRoom *graph.Room
		var bossRoom *graph.Room

		for _, room := range g.Rooms {
			switch room.Archetype {
			case graph.ArchetypeStart:
				startCount++
				startRoom = room
			case graph.ArchetypeBoss:
				bossCount++
				if bossRoom == nil {
					bossRoom = room
				}
			}
		}

		// Property 1: Must have exactly one Start room
		if startCount != 1 {
			t.Fatalf("graph must have exactly 1 Start room, got %d", startCount)
		}

		// Property 2: Must have at least one Boss room
		if bossCount < 1 {
			t.Fatalf("graph must have at least 1 Boss room, got %d", bossCount)
		}

		// Property 3: Start room must have a path to at least one Boss room
		if startRoom != nil && bossRoom != nil {
			path, err := g.GetPath(startRoom.ID, bossRoom.ID)
			if err != nil {
				t.Fatalf("no path from Start to Boss: %v", err)
			}
			if len(path) < 2 {
				t.Fatalf("path from Start to Boss must have at least 2 rooms, got %d", len(path))
			}
		}

		// Property 4: Start room should be Start archetype (sanity check)
		if startRoom != nil && startRoom.Archetype != graph.ArchetypeStart {
			t.Fatalf("Start room has wrong archetype: %v", startRoom.Archetype)
		}

		// Property 5: Boss room should be Boss archetype (sanity check)
		if bossRoom != nil && bossRoom.Archetype != graph.ArchetypeBoss {
			t.Fatalf("Boss room has wrong archetype: %v", bossRoom.Archetype)
		}

		t.Logf("Validated: 1 Start room, %d Boss room(s), path exists", bossCount)
	})
}
