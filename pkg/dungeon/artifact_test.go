package dungeon

import (
	"testing"

	"github.com/dshills/dungo/pkg/graph"
)

func TestArtifactStructure(t *testing.T) {
	// Test that Artifact can be constructed with all fields
	artifact := &Artifact{
		ADG: &Graph{
			Graph: graph.NewGraph(12345),
		},
		Layout: &Layout{
			Poses:         make(map[string]Pose),
			CorridorPaths: make(map[string]Path),
			Bounds:        Rect{X: 0, Y: 0, Width: 100, Height: 100},
		},
		TileMap: &TileMap{
			Width:      100,
			Height:     100,
			TileWidth:  16,
			TileHeight: 16,
			Layers:     make(map[string]*Layer),
		},
		Content: &Content{
			Spawns:  []Spawn{},
			Loot:    []Loot{},
			Puzzles: []PuzzleInstance{},
			Secrets: []SecretInstance{},
		},
		Metrics: &Metrics{
			BranchingFactor:   2.5,
			PathLength:        10,
			CycleCount:        2,
			PacingDeviation:   0.1,
			SecretFindability: 0.8,
		},
		Debug: &DebugArtifacts{
			ADGSVG:    []byte("svg data"),
			LayoutPNG: []byte("png data"),
			Report: &ValidationReport{
				Passed: true,
			},
		},
	}

	// artifact is created with a literal, so it cannot be nil
	// This check is redundant and causes staticcheck SA4031

	if artifact.ADG.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", artifact.ADG.Seed)
	}

	if artifact.Metrics.BranchingFactor != 2.5 {
		t.Errorf("Expected branching factor 2.5, got %f", artifact.Metrics.BranchingFactor)
	}

	if !artifact.Debug.Report.Passed {
		t.Error("Expected validation report to be passed")
	}
}

func TestPoseStructure(t *testing.T) {
	pose := Pose{
		X:           10,
		Y:           20,
		Rotation:    90,
		FootprintID: "room_template_1",
	}

	if pose.X != 10 || pose.Y != 20 {
		t.Errorf("Expected position (10, 20), got (%d, %d)", pose.X, pose.Y)
	}

	if pose.Rotation != 90 {
		t.Errorf("Expected rotation 90, got %d", pose.Rotation)
	}
}

func TestContentStructure(t *testing.T) {
	content := &Content{
		Spawns: []Spawn{
			{
				ID:        "spawn_1",
				RoomID:    "room_1",
				Position:  Point{X: 5, Y: 5},
				EnemyType: "goblin",
				Count:     3,
			},
		},
		Loot: []Loot{
			{
				ID:       "loot_1",
				RoomID:   "room_1",
				Position: Point{X: 10, Y: 10},
				ItemType: "gold",
				Value:    100,
				Required: false,
			},
		},
	}

	if len(content.Spawns) != 1 {
		t.Errorf("Expected 1 spawn, got %d", len(content.Spawns))
	}

	if content.Spawns[0].EnemyType != "goblin" {
		t.Errorf("Expected enemy type 'goblin', got %s", content.Spawns[0].EnemyType)
	}

	if len(content.Loot) != 1 {
		t.Errorf("Expected 1 loot item, got %d", len(content.Loot))
	}

	if content.Loot[0].Value != 100 {
		t.Errorf("Expected loot value 100, got %d", content.Loot[0].Value)
	}
}
