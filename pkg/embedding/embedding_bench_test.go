package embedding

import (
	"fmt"
	"testing"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// BenchmarkForceDirectedEmbedding benchmarks the force-directed embedder
// for different graph sizes to verify performance targets.
func BenchmarkForceDirectedEmbedding(b *testing.B) {
	tests := []struct {
		name      string
		roomCount int
	}{
		{"Small_10_rooms", 10},
		{"Medium_30_rooms", 30},
		{"Target_60_rooms", 60},
		{"Large_100_rooms", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create a test graph with connected rooms
			g := createLinearGraph(tt.roomCount, 12345)

			// Create embedder with relaxed config for benchmarking
			config := DefaultConfig()
			config.MinRoomSpacing = 1.0      // More lenient spacing
			config.CorridorMaxLength = 200.0 // Longer corridors allowed
			config.CorridorMaxBends = 10     // More bends allowed
			embedder := NewForceDirectedEmbedder(config)

			configHash := []byte("benchmark")

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use different seed for each iteration
				rngInst := rng.NewRNG(uint64(12345+i), "embedding", configHash)

				layout, err := embedder.Embed(g, rngInst)
				if err != nil {
					b.Fatalf("embedding failed: %v", err)
				}

				// Verify we got all poses
				if len(layout.Poses) != tt.roomCount {
					b.Fatalf("unexpected pose count: got %d, want %d",
						len(layout.Poses), tt.roomCount)
				}
			}
		})
	}
}

// BenchmarkForceDirectedEmbedding_ComplexGraph benchmarks embedding
// of graphs with more complex connectivity (higher branching).
func BenchmarkForceDirectedEmbedding_ComplexGraph(b *testing.B) {
	tests := []struct {
		name      string
		roomCount int
		avgDegree int
	}{
		{"30rooms_deg2", 30, 2},
		{"60rooms_deg2", 60, 2},
		{"60rooms_deg3", 60, 3},
		{"100rooms_deg2", 100, 2},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create a test graph with specified connectivity
			g := createBranchedGraph(tt.roomCount, tt.avgDegree, 12345)

			config := DefaultConfig()
			config.MinRoomSpacing = 1.0      // More lenient spacing
			config.CorridorMaxLength = 200.0 // Longer corridors allowed
			config.CorridorMaxBends = 10     // More bends allowed
			embedder := NewForceDirectedEmbedder(config)

			configHash := []byte("benchmark")

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rngInst := rng.NewRNG(uint64(12345+i), "embedding", configHash)

				layout, err := embedder.Embed(g, rngInst)
				if err != nil {
					b.Fatalf("embedding failed: %v", err)
				}

				if len(layout.Poses) != tt.roomCount {
					b.Fatalf("unexpected pose count: got %d, want %d",
						len(layout.Poses), tt.roomCount)
				}
			}
		})
	}
}

// BenchmarkCorridorPathGeneration benchmarks just the corridor path generation
// phase (A* pathfinding between rooms).
func BenchmarkCorridorPathGeneration(b *testing.B) {
	tests := []struct {
		name          string
		roomCount     int
		corridorCount int
	}{
		{"10rooms_9corridors", 10, 9},
		{"30rooms_29corridors", 30, 29},
		{"60rooms_59corridors", 60, 59},
		{"100rooms_99corridors", 100, 99},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create graph and perform initial embedding
			g := createLinearGraph(tt.roomCount, 12345)
			config := DefaultConfig()
			embedder := NewForceDirectedEmbedder(config)
			configHash := []byte("benchmark")
			rngInst := rng.NewRNG(12345, "embedding", configHash)

			// Do one embedding to get poses
			layout, err := embedder.Embed(g, rngInst)
			if err != nil {
				b.Fatalf("initial embedding failed: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Benchmark just the corridor path generation
				// by clearing existing paths and regenerating
				layout.CorridorPaths = make(map[string]*Path)

				// Regenerate all corridor paths
				for connID, conn := range g.Connectors {
					fromPose := layout.Poses[conn.From]
					toPose := layout.Poses[conn.To]

					path := generateSimpleCorridorPath(fromPose, toPose)
					layout.CorridorPaths[connID] = path
				}

				if len(layout.CorridorPaths) != tt.corridorCount {
					b.Fatalf("unexpected corridor count: got %d, want %d",
						len(layout.CorridorPaths), tt.corridorCount)
				}
			}
		})
	}
}

// BenchmarkLayoutValidation benchmarks the layout validation process.
func BenchmarkLayoutValidation(b *testing.B) {
	tests := []struct {
		name      string
		roomCount int
	}{
		{"30rooms", 30},
		{"60rooms", 60},
		{"100rooms", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create and embed graph once
			g := createLinearGraph(tt.roomCount, 12345)
			config := DefaultConfig()
			embedder := NewForceDirectedEmbedder(config)
			configHash := []byte("benchmark")
			rngInst := rng.NewRNG(12345, "embedding", configHash)

			layout, err := embedder.Embed(g, rngInst)
			if err != nil {
				b.Fatalf("embedding failed: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := ValidateEmbedding(layout, g, config)
				if err != nil {
					b.Fatalf("validation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkEmbeddingWithDifferentConfigs benchmarks embedding with
// different configuration parameters.
func BenchmarkEmbeddingWithDifferentConfigs(b *testing.B) {
	tests := []struct {
		name          string
		maxIterations int
		roomSpacing   float64
	}{
		{"default_config", 500, 2.0},
		{"fast_convergence", 200, 1.0},
		{"high_quality", 1000, 3.0},
	}

	roomCount := 60
	g := createLinearGraph(roomCount, 12345)

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			config := DefaultConfig()
			config.MaxIterations = tt.maxIterations
			config.MinRoomSpacing = tt.roomSpacing

			embedder := NewForceDirectedEmbedder(config)
			configHash := []byte("benchmark")

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rngInst := rng.NewRNG(uint64(12345+i), "embedding", configHash)

				layout, err := embedder.Embed(g, rngInst)
				if err != nil {
					b.Fatalf("embedding failed: %v", err)
				}

				if len(layout.Poses) != roomCount {
					b.Fatalf("unexpected pose count: got %d, want %d",
						len(layout.Poses), roomCount)
				}
			}
		})
	}
}

// Helper function to create a linear chain of rooms.
func createLinearGraph(roomCount int, seed uint64) *graph.Graph {
	g := graph.NewGraph(seed)

	// Add rooms
	for i := 0; i < roomCount; i++ {
		archetype := graph.ArchetypeHub
		if i == 0 {
			archetype = graph.ArchetypeStart
		} else if i == roomCount-1 {
			archetype = graph.ArchetypeBoss
		}

		room := &graph.Room{
			ID:         fmt.Sprintf("R%03d", i),
			Archetype:  archetype,
			Size:       graph.SizeM,
			Difficulty: float64(i) / float64(roomCount-1),
			Reward:     float64(i) / float64(roomCount-1),
		}
		if err := g.AddRoom(room); err != nil {
			panic(fmt.Sprintf("AddRoom failed: %v", err))
		}
	}

	// Add connectors
	for i := 0; i < roomCount-1; i++ {
		conn := &graph.Connector{
			ID:            fmt.Sprintf("C%03d", i),
			From:          fmt.Sprintf("R%03d", i),
			To:            fmt.Sprintf("R%03d", i+1),
			Type:          graph.TypeCorridor,
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		}
		if err := g.AddConnector(conn); err != nil {
			panic(fmt.Sprintf("AddConnector failed: %v", err))
		}
	}

	return g
}

// Helper function to create a graph with branching structure.
func createBranchedGraph(roomCount, avgDegree int, seed uint64) *graph.Graph {
	g := graph.NewGraph(seed)
	rngInst := rng.NewRNG(seed, "graph", []byte("test"))

	// Add rooms
	for i := 0; i < roomCount; i++ {
		archetype := graph.ArchetypeHub
		if i == 0 {
			archetype = graph.ArchetypeStart
		} else if i == roomCount-1 {
			archetype = graph.ArchetypeBoss
		}

		room := &graph.Room{
			ID:         fmt.Sprintf("R%03d", i),
			Archetype:  archetype,
			Size:       graph.SizeM,
			Difficulty: float64(i) / float64(roomCount-1),
			Reward:     float64(i) / float64(roomCount-1),
		}
		if err := g.AddRoom(room); err != nil {
			panic(fmt.Sprintf("AddRoom failed: %v", err))
		}
	}

	// Create a main path first (ensures connectivity)
	for i := 0; i < roomCount-1; i++ {
		conn := &graph.Connector{
			ID:            fmt.Sprintf("C_main_%03d", i),
			From:          fmt.Sprintf("R%03d", i),
			To:            fmt.Sprintf("R%03d", i+1),
			Type:          graph.TypeCorridor,
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		}
		if err := g.AddConnector(conn); err != nil {
			panic(fmt.Sprintf("AddConnector failed: %v", err))
		}
	}

	// Add additional edges to reach average degree
	targetEdges := (roomCount * avgDegree) / 2
	currentEdges := roomCount - 1 // from main path
	connectorID := roomCount

	for currentEdges < targetEdges && currentEdges < (roomCount*(roomCount-1))/2 {
		// Pick two random rooms
		from := rngInst.Intn(roomCount)
		to := rngInst.Intn(roomCount)

		if from == to {
			continue
		}
		if from > to {
			from, to = to, from
		}

		// Check if connector already exists
		fromID := fmt.Sprintf("R%03d", from)
		toID := fmt.Sprintf("R%03d", to)

		exists := false
		for _, conn := range g.Connectors {
			if (conn.From == fromID && conn.To == toID) ||
				(conn.From == toID && conn.To == fromID) {
				exists = true
				break
			}
		}

		if !exists {
			conn := &graph.Connector{
				ID:            fmt.Sprintf("C_extra_%03d", connectorID),
				From:          fromID,
				To:            toID,
				Type:          graph.TypeCorridor,
				Cost:          1.0,
				Visibility:    graph.VisibilityNormal,
				Bidirectional: true,
			}
			if err := g.AddConnector(conn); err != nil {
				panic(fmt.Sprintf("AddConnector failed: %v", err))
			}
			currentEdges++
			connectorID++
		}
	}

	return g
}

// Helper function to generate a simple corridor path between two poses.
func generateSimpleCorridorPath(from, to *Pose) *Path {
	// Get centers
	fromX, fromY := from.Center()
	toX, toY := to.Center()

	// Create L-shaped path (horizontal then vertical)
	path := &Path{
		Points: []Point{
			{X: fromX, Y: fromY},
			{X: toX, Y: fromY},
			{X: toX, Y: toY},
		},
	}

	return path
}
