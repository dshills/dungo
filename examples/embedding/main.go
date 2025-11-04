package main

import (
	"fmt"
	"log"

	"github.com/dshills/dungo/pkg/embedding"
	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

func main() {
	// Create a simple dungeon graph
	g := graph.NewGraph(12345)

	// Add rooms
	rooms := []*graph.Room{
		{
			ID:         "R001",
			Archetype:  graph.ArchetypeStart,
			Size:       graph.SizeM,
			Difficulty: 0.0,
			Reward:     0.0,
		},
		{
			ID:         "R002",
			Archetype:  graph.ArchetypeHub,
			Size:       graph.SizeM,
			Difficulty: 0.3,
			Reward:     0.3,
		},
		{
			ID:         "R003",
			Archetype:  graph.ArchetypeTreasure,
			Size:       graph.SizeS,
			Difficulty: 0.4,
			Reward:     0.8,
		},
		{
			ID:         "R004",
			Archetype:  graph.ArchetypePuzzle,
			Size:       graph.SizeM,
			Difficulty: 0.6,
			Reward:     0.5,
		},
		{
			ID:         "R005",
			Archetype:  graph.ArchetypeBoss,
			Size:       graph.SizeXL,
			Difficulty: 1.0,
			Reward:     1.0,
		},
	}

	for _, room := range rooms {
		if err := g.AddRoom(room); err != nil {
			log.Fatalf("Failed to add room: %v", err)
		}
	}

	// Add connectors (creating a simple path with one branch)
	connectors := []*graph.Connector{
		{
			ID:            "C001",
			From:          "R001",
			To:            "R002",
			Type:          graph.TypeCorridor,
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		},
		{
			ID:            "C002",
			From:          "R002",
			To:            "R003",
			Type:          graph.TypeDoor,
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		},
		{
			ID:            "C003",
			From:          "R002",
			To:            "R004",
			Type:          graph.TypeCorridor,
			Cost:          1.5,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		},
		{
			ID:            "C004",
			From:          "R004",
			To:            "R005",
			Type:          graph.TypeCorridor,
			Cost:          2.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		},
	}

	for _, conn := range connectors {
		if err := g.AddConnector(conn); err != nil {
			log.Fatalf("Failed to add connector: %v", err)
		}
	}

	fmt.Printf("Created graph with %d rooms and %d connectors\n", len(g.Rooms), len(g.Connectors))
	fmt.Printf("Graph is connected: %v\n\n", g.IsConnected())

	// Create embedder with relaxed spacing
	config := embedding.DefaultConfig()
	config.MinRoomSpacing = 1.0 // Reduce minimum spacing for this example
	embedder, err := embedding.Get("force_directed", config)
	if err != nil {
		log.Fatalf("Failed to get embedder: %v", err)
	}

	// Create RNG for embedding stage
	configHash := []byte("example_config")
	rngInstance := rng.NewRNG(12345, "embedding", configHash)

	// Perform spatial embedding
	fmt.Println("Performing spatial embedding...")
	layout, err := embedder.Embed(g, rngInstance)
	if err != nil {
		log.Fatalf("Failed to embed graph: %v", err)
	}

	fmt.Printf("Embedding complete! Algorithm: %s, Seed: %d\n\n", layout.Algorithm, layout.Seed)

	// Display results
	fmt.Println("Room Positions:")
	fmt.Println("===============")
	for roomID, pose := range layout.Poses {
		room := g.Rooms[roomID]
		fmt.Printf("%s (%s, %s):\n", roomID, room.Archetype, room.Size)
		fmt.Printf("  Position: (%.1f, %.1f)\n", pose.X, pose.Y)
		fmt.Printf("  Dimensions: %dx%d\n", pose.Width, pose.Height)
		cx, cy := pose.Center()
		fmt.Printf("  Center: (%.1f, %.1f)\n\n", cx, cy)
	}

	fmt.Println("Corridor Paths:")
	fmt.Println("===============")
	for connID, path := range layout.CorridorPaths {
		conn := g.Connectors[connID]
		fmt.Printf("%s: %s → %s (%s)\n", connID, conn.From, conn.To, conn.Type)
		fmt.Printf("  Length: %.1f\n", path.Length())
		fmt.Printf("  Bends: %d\n", path.BendCount())
		fmt.Printf("  Points: %d\n\n", len(path.Points))
	}

	fmt.Println("Layout Bounds:")
	fmt.Println("==============")
	fmt.Printf("Min: (%.1f, %.1f)\n", layout.Bounds.MinX, layout.Bounds.MinY)
	fmt.Printf("Max: (%.1f, %.1f)\n", layout.Bounds.MaxX, layout.Bounds.MaxY)
	fmt.Printf("Size: %.1f x %.1f\n\n", layout.Bounds.Width(), layout.Bounds.Height())

	// Validate the embedding
	if err := embedding.ValidateEmbedding(layout, g, config); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("Validation: PASSED")
	fmt.Println("\nEmbedding stage completed successfully!")
}
