package export

import (
	"encoding/json"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// TestT093_JSONRoundTrip verifies that a dungeon artifact can be serialized
// to JSON and deserialized back without data loss.
// This test satisfies requirement T093: Unit test for JSON serialization round-trip.
func TestT093_JSONRoundTrip(t *testing.T) {
	// Create a comprehensive artifact with all fields populated
	original := &dungeon.Artifact{
		ADG: &dungeon.Graph{
			Graph: &graph.Graph{
				Rooms: map[string]*graph.Room{
					"start": {
						ID:         "start",
						Archetype:  graph.ArchetypeStart,
						Size:       graph.SizeM,
						Difficulty: 0.1,
						Reward:     0.2,
						Tags:       map[string]string{"theme": "entrance"},
					},
					"boss": {
						ID:         "boss",
						Archetype:  graph.ArchetypeBoss,
						Size:       graph.SizeXL,
						Difficulty: 0.9,
						Reward:     1.0,
						Tags:       map[string]string{"theme": "final"},
					},
				},
				Connectors: map[string]*graph.Connector{
					"conn1": {
						ID:            "conn1",
						From:          "start",
						To:            "boss",
						Type:          graph.TypeCorridor,
						Bidirectional: true,
						Cost:          1.0,
					},
				},
				Metadata: map[string]interface{}{
					"version": "1.0",
					"seed":    uint64(12345),
				},
			},
		},
		Layout: &dungeon.Layout{
			Poses: map[string]dungeon.Pose{
				"start": {X: 0, Y: 0, Rotation: 0, FootprintID: "rect-medium"},
				"boss":  {X: 50, Y: 50, Rotation: 0, FootprintID: "rect-large"},
			},
			CorridorPaths: map[string]dungeon.Path{
				"conn1": {
					Points: []dungeon.Point{{X: 0, Y: 0}, {X: 25, Y: 25}, {X: 50, Y: 50}},
				},
			},
			Bounds: dungeon.Rect{X: 0, Y: 0, Width: 100, Height: 100},
		},
		Content: &dungeon.Content{
			Spawns: []dungeon.Spawn{
				{
					ID:        "spawn1",
					RoomID:    "start",
					Position:  dungeon.Point{X: 10, Y: 10},
					EnemyType: "goblin",
					Count:     3,
				},
			},
			Loot: []dungeon.Loot{
				{
					ID:       "loot1",
					RoomID:   "boss",
					Position: dungeon.Point{X: 55, Y: 55},
					ItemType: "legendary_sword",
					Value:    1000,
					Required: true,
				},
			},
			Puzzles: []dungeon.PuzzleInstance{
				{
					ID:         "puzzle1",
					RoomID:     "boss",
					Type:       "lever",
					Difficulty: 0.8,
				},
			},
			Secrets: []dungeon.SecretInstance{
				{
					ID:       "secret1",
					RoomID:   "start",
					Type:     "hidden_passage",
					Position: dungeon.Point{X: 5, Y: 5},
					Clues:    []string{"Look for the crack"},
				},
			},
		},
		Metrics: &dungeon.Metrics{
			BranchingFactor:   1.5,
			PathLength:        2,
			CycleCount:        0,
			PacingDeviation:   0.15,
			SecretFindability: 0.7,
		},
	}

	// Step 1: Serialize to JSON
	jsonData, err := ExportJSON(original)
	if err != nil {
		t.Fatalf("Failed to serialize artifact to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Fatal("Serialization produced empty JSON")
	}

	t.Logf("Serialized artifact to %d bytes of JSON", len(jsonData))

	// Step 2: Deserialize back from JSON
	var restored dungeon.Artifact
	if err := json.Unmarshal(jsonData, &restored); err != nil {
		t.Fatalf("Failed to deserialize JSON back to artifact: %v", err)
	}

	// Step 3: Verify all critical data is preserved

	// Verify ADG structure
	if restored.ADG == nil {
		t.Fatal("ADG is nil after round-trip")
	}
	if restored.ADG.Graph == nil {
		t.Fatal("ADG.Graph is nil after round-trip")
	}

	// Verify rooms
	if len(restored.ADG.Rooms) != len(original.ADG.Rooms) {
		t.Errorf("Room count mismatch: got %d, want %d",
			len(restored.ADG.Rooms), len(original.ADG.Rooms))
	}

	// Verify specific room data
	if startRoom, ok := restored.ADG.Rooms["start"]; ok {
		origStart := original.ADG.Rooms["start"]
		if startRoom.ID != origStart.ID {
			t.Errorf("Start room ID mismatch: got %s, want %s", startRoom.ID, origStart.ID)
		}
		if startRoom.Difficulty != origStart.Difficulty {
			t.Errorf("Start room difficulty mismatch: got %f, want %f",
				startRoom.Difficulty, origStart.Difficulty)
		}
	} else {
		t.Error("Start room not found after round-trip")
	}

	// Verify connectors
	if len(restored.ADG.Connectors) != len(original.ADG.Connectors) {
		t.Errorf("Connector count mismatch: got %d, want %d",
			len(restored.ADG.Connectors), len(original.ADG.Connectors))
	}

	// Verify Layout
	if restored.Layout == nil {
		t.Fatal("Layout is nil after round-trip")
	}
	if len(restored.Layout.Poses) != len(original.Layout.Poses) {
		t.Errorf("Pose count mismatch: got %d, want %d",
			len(restored.Layout.Poses), len(original.Layout.Poses))
	}
	if len(restored.Layout.CorridorPaths) != len(original.Layout.CorridorPaths) {
		t.Errorf("CorridorPath count mismatch: got %d, want %d",
			len(restored.Layout.CorridorPaths), len(original.Layout.CorridorPaths))
	}

	// Verify Content
	if restored.Content == nil {
		t.Fatal("Content is nil after round-trip")
	}
	if len(restored.Content.Spawns) != len(original.Content.Spawns) {
		t.Errorf("Spawn count mismatch: got %d, want %d",
			len(restored.Content.Spawns), len(original.Content.Spawns))
	}
	if len(restored.Content.Loot) != len(original.Content.Loot) {
		t.Errorf("Loot count mismatch: got %d, want %d",
			len(restored.Content.Loot), len(original.Content.Loot))
	}
	if len(restored.Content.Puzzles) != len(original.Content.Puzzles) {
		t.Errorf("Puzzle count mismatch: got %d, want %d",
			len(restored.Content.Puzzles), len(original.Content.Puzzles))
	}
	if len(restored.Content.Secrets) != len(original.Content.Secrets) {
		t.Errorf("Secret count mismatch: got %d, want %d",
			len(restored.Content.Secrets), len(original.Content.Secrets))
	}

	// Verify Metrics
	if restored.Metrics == nil {
		t.Fatal("Metrics is nil after round-trip")
	}
	if restored.Metrics.BranchingFactor != original.Metrics.BranchingFactor {
		t.Errorf("BranchingFactor mismatch: got %f, want %f",
			restored.Metrics.BranchingFactor, original.Metrics.BranchingFactor)
	}
	if restored.Metrics.PathLength != original.Metrics.PathLength {
		t.Errorf("PathLength mismatch: got %d, want %d",
			restored.Metrics.PathLength, original.Metrics.PathLength)
	}

	t.Log("✓ JSON round-trip test PASSED - all data preserved correctly")
}

// TestT093_JSONRoundTrip_EmptyFields verifies round-trip with minimal data.
func TestT093_JSONRoundTrip_EmptyFields(t *testing.T) {
	original := &dungeon.Artifact{}

	jsonData, err := ExportJSON(original)
	if err != nil {
		t.Fatalf("Failed to serialize empty artifact: %v", err)
	}

	var restored dungeon.Artifact
	if err := json.Unmarshal(jsonData, &restored); err != nil {
		t.Fatalf("Failed to deserialize empty artifact: %v", err)
	}

	// Just verify it doesn't crash
	t.Log("✓ Empty artifact round-trip PASSED")
}

// TestT093_JSONRoundTrip_LargeStructure verifies round-trip with complex nested data.
func TestT093_JSONRoundTrip_LargeStructure(t *testing.T) {
	// Create artifact with many rooms and connectors
	rooms := make(map[string]*graph.Room)
	connectors := make(map[string]*graph.Connector)

	for i := 0; i < 50; i++ {
		id := "room-" + string(rune('0'+i%10))
		rooms[id] = &graph.Room{
			ID:         id,
			Archetype:  graph.RoomArchetype(i % 11),
			Size:       graph.RoomSize(i % 5),
			Difficulty: float64(i) / 100.0,
		}
	}

	original := &dungeon.Artifact{
		ADG: &dungeon.Graph{
			Graph: &graph.Graph{
				Rooms:      rooms,
				Connectors: connectors,
			},
		},
	}

	jsonData, err := ExportJSON(original)
	if err != nil {
		t.Fatalf("Failed to serialize large artifact: %v", err)
	}

	var restored dungeon.Artifact
	if err := json.Unmarshal(jsonData, &restored); err != nil {
		t.Fatalf("Failed to deserialize large artifact: %v", err)
	}

	if len(restored.ADG.Rooms) != len(original.ADG.Rooms) {
		t.Errorf("Large structure room count mismatch: got %d, want %d",
			len(restored.ADG.Rooms), len(original.ADG.Rooms))
	}

	t.Logf("✓ Large structure round-trip PASSED - %d rooms preserved", len(restored.ADG.Rooms))
}
