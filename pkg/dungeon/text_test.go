package dungeon_test

import (
	"strings"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

func TestArtifact_RenderText(t *testing.T) {
	// Create a minimal artifact for testing
	g := graph.NewGraph(12345)

	start := &graph.Room{
		ID:         "R001",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Difficulty: 0.0,
		Reward:     0.0,
	}
	boss := &graph.Room{
		ID:         "R002",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeL,
		Difficulty: 1.0,
		Reward:     0.8,
	}

	g.AddRoom(start)
	g.AddRoom(boss)
	g.AddConnector(&graph.Connector{
		ID:            "C001",
		From:          "R001",
		To:            "R002",
		Type:          graph.TypeCorridor,
		Bidirectional: true,
		Cost:          1.0,
	})

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
		Metrics: &dungeon.Metrics{
			BranchingFactor: 1.0,
			PathLength:      2,
			CycleCount:      0,
		},
	}

	// Test rendering
	text := artifact.RenderText()

	// Verify output contains expected elements
	if !strings.Contains(text, "DUNGEON GENERATOR") {
		t.Error("Missing header")
	}
	if !strings.Contains(text, "Rooms: 2") {
		t.Error("Missing room count")
	}
	if !strings.Contains(text, "Connections: 1") {
		t.Error("Missing connection count")
	}
	if !strings.Contains(text, "R001") {
		t.Error("Missing start room ID")
	}
	if !strings.Contains(text, "R002") {
		t.Error("Missing boss room ID")
	}
	if !strings.Contains(text, "Start") {
		t.Error("Missing Start archetype")
	}
	if !strings.Contains(text, "Boss") {
		t.Error("Missing Boss archetype")
	}
}

func TestArtifact_RenderTextSimple(t *testing.T) {
	g := graph.NewGraph(12345)

	start := &graph.Room{
		ID:        "R001",
		Archetype: graph.ArchetypeStart,
		Size:      graph.SizeM,
	}
	treasure := &graph.Room{
		ID:        "R002",
		Archetype: graph.ArchetypeTreasure,
		Size:      graph.SizeS,
	}

	g.AddRoom(start)
	g.AddRoom(treasure)
	g.AddConnector(&graph.Connector{
		ID:            "C001",
		From:          "R001",
		To:            "R002",
		Type:          graph.TypeDoor,
		Bidirectional: true,
		Cost:          1.0,
	})

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}

	// Test simple rendering
	text := artifact.RenderTextSimple()

	if !strings.Contains(text, "2 rooms") {
		t.Error("Missing room count")
	}
	if !strings.Contains(text, "Path from Start") {
		t.Error("Missing path header")
	}
	if !strings.Contains(text, "R001") {
		t.Error("Missing start room in path")
	}
}

func TestArtifact_RenderText_NilArtifact(t *testing.T) {
	var artifact *dungeon.Artifact
	text := artifact.RenderText()

	if !strings.Contains(text, "No dungeon data") {
		t.Error("Should handle nil artifact gracefully")
	}
}

func TestArtifact_RenderText_WithKeyLock(t *testing.T) {
	g := graph.NewGraph(99999)

	start := &graph.Room{
		ID:        "R001",
		Archetype: graph.ArchetypeStart,
		Size:      graph.SizeM,
	}
	keyRoom := &graph.Room{
		ID:        "R002",
		Archetype: graph.ArchetypeTreasure,
		Size:      graph.SizeS,
		Provides: []graph.Capability{
			{Type: "key", Value: "silver"},
		},
	}
	lockedRoom := &graph.Room{
		ID:        "R003",
		Archetype: graph.ArchetypeBoss,
		Size:      graph.SizeXL,
	}

	g.AddRoom(start)
	g.AddRoom(keyRoom)
	g.AddRoom(lockedRoom)

	g.AddConnector(&graph.Connector{
		ID:            "C001",
		From:          "R001",
		To:            "R002",
		Type:          graph.TypeCorridor,
		Bidirectional: true,
		Cost:          1.0,
	})

	g.AddConnector(&graph.Connector{
		ID:   "C002",
		From: "R002",
		To:   "R003",
		Type: graph.TypeDoor,
		Gate: &graph.Gate{
			Type:  "key",
			Value: "silver",
		},
		Bidirectional: true,
		Cost:          1.0,
	})

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}

	text := artifact.RenderText()

	// Verify key-lock information is shown
	if !strings.Contains(text, "Provides: key:silver") {
		t.Error("Missing key provision information")
	}
	if !strings.Contains(text, "🔒 Requires key:silver") {
		t.Error("Missing lock requirement information")
	}
}
