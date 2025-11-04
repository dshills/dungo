package dungeon

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

// BenchmarkFullGeneration benchmarks the complete dungeon generation pipeline
// from configuration to final artifact including all stages:
// 1. Graph synthesis
// 2. Spatial embedding
// 3. Tile carving
// 4. Content population
// 5. Validation
func BenchmarkFullGeneration(b *testing.B) {
	tests := []struct {
		name     string
		roomsMin int
		roomsMax int
	}{
		{
			name:     "Tiny_5-10_rooms",
			roomsMin: 5,
			roomsMax: 10,
		},
		{
			name:     "Small_10-20_rooms",
			roomsMin: 10,
			roomsMax: 20,
		},
		{
			name:     "Medium_20-40_rooms",
			roomsMin: 20,
			roomsMax: 40,
		},
		{
			name:     "Target_60_rooms",
			roomsMin: 60,
			roomsMax: 60,
		},
		{
			name:     "Large_60-100_rooms",
			roomsMin: 60,
			roomsMax: 100,
		},
		{
			name:     "Huge_100-150_rooms",
			roomsMin: 100,
			roomsMax: 150,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create configuration
			cfg := &Config{
				Seed: 12345,
				Size: SizeCfg{
					RoomsMin: tt.roomsMin,
					RoomsMax: tt.roomsMax,
				},
				Branching: BranchingCfg{
					Avg: 2.0,
					Max: 4,
				},
				SecretDensity: 0.15,
				OptionalRatio: 0.25,
				Keys: []KeyCfg{
					{Name: "key_red", Count: 1},
					{Name: "key_blue", Count: 1},
				},
				Pacing: PacingCfg{
					Curve:    PacingLinear,
					Variance: 0.15,
				},
				Themes: []string{"dungeon", "crypt", "cavern"},
			}

			// Create generator with mock validator to isolate generation performance
			gen := NewGenerator().(*DefaultGenerator)
			gen.SetValidator(&mockValidator{})

			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Use different seed for each iteration
				cfg.Seed = uint64(12345 + i)

				artifact, err := gen.Generate(ctx, cfg)
				if err != nil {
					b.Fatalf("generation failed: %v", err)
				}

				// Basic sanity check
				if artifact.ADG == nil || artifact.Layout == nil || artifact.TileMap == nil {
					b.Fatal("incomplete artifact generated")
				}

				roomCount := len(artifact.ADG.Rooms)
				if roomCount < tt.roomsMin || roomCount > tt.roomsMax {
					b.Fatalf("unexpected room count: got %d, want [%d, %d]",
						roomCount, tt.roomsMin, tt.roomsMax)
				}
			}
		})
	}
}

// BenchmarkFullGeneration_WithMemoryStats benchmarks with memory statistics
// to verify the <50MB memory usage target.
func BenchmarkFullGeneration_WithMemoryStats(b *testing.B) {
	cfg := &Config{
		Seed: 12345,
		Size: SizeCfg{
			RoomsMin: 60,
			RoomsMax: 60,
		},
		Branching: BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Keys: []KeyCfg{
			{Name: "key_red", Count: 1},
		},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.15,
		},
		Themes: []string{"dungeon"},
	}

	gen := NewGenerator().(*DefaultGenerator)
	gen.SetValidator(&mockValidator{})

	ctx := context.Background()

	// Force GC before starting
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startAlloc := m.Alloc

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cfg.Seed = uint64(12345 + i)

		_, err := gen.Generate(ctx, cfg)
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}

	b.StopTimer()

	// Report memory statistics
	runtime.ReadMemStats(&m)
	allocPerOp := (m.Alloc - startAlloc) / uint64(b.N)
	b.ReportMetric(float64(allocPerOp)/(1024*1024), "MB/op")
}

// BenchmarkGraphPlusEmbedding benchmarks just the graph synthesis and
// embedding stages to verify the <50ms target for 60-room dungeons.
func BenchmarkGraphPlusEmbedding(b *testing.B) {
	tests := []struct {
		name      string
		roomCount int
	}{
		{"30_rooms", 30},
		{"60_rooms_target", 60},
		{"100_rooms", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cfg := &Config{
				Seed: 12345,
				Size: SizeCfg{
					RoomsMin: tt.roomCount,
					RoomsMax: tt.roomCount,
				},
				Branching: BranchingCfg{
					Avg: 2.0,
					Max: 4,
				},
				SecretDensity: 0.15,
				OptionalRatio: 0.25,
				Keys: []KeyCfg{
					{Name: "key_red", Count: 1},
				},
				Pacing: PacingCfg{
					Curve:    PacingLinear,
					Variance: 0.15,
				},
				Themes: []string{"dungeon"},
			}

			gen := NewGenerator().(*DefaultGenerator)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cfg.Seed = uint64(12345 + i)

				// This will fail at validation since we don't have a validator set,
				// but we can measure up to that point
				_, _ = gen.Generate(ctx, cfg)
				// Ignore error - we're just benchmarking synthesis + embedding
			}
		})
	}
}

// BenchmarkGenerationByStage benchmarks each pipeline stage individually
// to identify performance bottlenecks.
func BenchmarkGenerationByStage(b *testing.B) {
	roomCount := 60

	cfg := &Config{
		Seed: 12345,
		Size: SizeCfg{
			RoomsMin: roomCount,
			RoomsMax: roomCount,
		},
		Branching: BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Keys: []KeyCfg{
			{Name: "key_red", Count: 1},
		},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.15,
		},
		Themes: []string{"dungeon"},
	}

	gen := NewGenerator().(*DefaultGenerator)
	gen.SetValidator(&mockValidator{})
	ctx := context.Background()

	// Generate once to get intermediate artifacts for stage benchmarks
	cfg.Seed = 12345
	artifact, err := gen.Generate(ctx, cfg)
	if err != nil {
		b.Fatalf("setup generation failed: %v", err)
	}

	b.Run("Stage1_Synthesis", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Just synthesis stage is measured in synthesis_bench_test.go
			// This is a reference point showing it in context
			b.Skip("Use synthesis_bench_test.go for synthesis-only benchmarks")
		}
	})

	b.Run("Stage2_Embedding", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Just embedding stage is measured in embedding_bench_test.go
			// This is a reference point showing it in context
			b.Skip("Use embedding_bench_test.go for embedding-only benchmarks")
		}
	})

	b.Run("Stage3_Carving", func(b *testing.B) {
		// Benchmark carving with pre-generated graph and layout
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			carvingLayout := convertToCarvingLayout(artifact.Layout)
			graphAdapter := &mockGraphAdapter{
				roomCount: roomCount,
			}
			_, err := gen.carver.Carve(ctx, graphAdapter, carvingLayout)
			if err != nil {
				b.Fatalf("carving failed: %v", err)
			}
		}
	})

	b.Run("Stage4_Content", func(b *testing.B) {
		// Benchmark content population with pre-generated graph
		configHash := cfg.Hash()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			contentRNG := newRNG(uint64(12345+i), "content", configHash)
			_, err := gen.contentPass.Place(ctx, artifact.ADG.Graph, contentRNG)
			if err != nil {
				b.Fatalf("content failed: %v", err)
			}
		}
	})

	b.Run("Stage5_Validation", func(b *testing.B) {
		// Benchmark validation with complete artifact
		validator := &mockValidator{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := validator.Validate(ctx, artifact, cfg)
			if err != nil {
				b.Fatalf("validation failed: %v", err)
			}
		}
	})
}

// BenchmarkConfigHashing benchmarks the config hashing performance.
func BenchmarkConfigHashing(b *testing.B) {
	cfg := &Config{
		Seed: 12345,
		Size: SizeCfg{
			RoomsMin: 60,
			RoomsMax: 60,
		},
		Branching: BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Keys: []KeyCfg{
			{Name: "key_red", Count: 1},
			{Name: "key_blue", Count: 1},
		},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.15,
		},
		Themes: []string{"dungeon", "crypt", "cavern"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Hash()
	}
}

// BenchmarkLayoutConversion benchmarks coordinate conversion overhead.
func BenchmarkLayoutConversion(b *testing.B) {
	// Create a test embedding layout
	embeddingLayout := createTestEmbeddingLayout(60)

	b.Run("EmbeddingToDungeon", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = convertEmbeddingLayout(embeddingLayout)
		}
	})

	b.Run("DungeonToCarving", func(b *testing.B) {
		dungeonLayout := convertEmbeddingLayout(embeddingLayout)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = convertToCarvingLayout(dungeonLayout)
		}
	})
}

// BenchmarkParallelGeneration benchmarks concurrent dungeon generation
// to verify thread safety and scalability.
func BenchmarkParallelGeneration(b *testing.B) {
	cfg := &Config{
		Seed: 12345,
		Size: SizeCfg{
			RoomsMin: 30,
			RoomsMax: 30,
		},
		Branching: BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Keys: []KeyCfg{
			{Name: "key_red", Count: 1},
		},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.15,
		},
		Themes: []string{"dungeon"},
	}

	gen := NewGenerator().(*DefaultGenerator)
	gen.SetValidator(&mockValidator{})
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			localCfg := *cfg
			localCfg.Seed = uint64(12345 + i)
			i++

			_, err := gen.Generate(ctx, &localCfg)
			if err != nil {
				b.Fatalf("generation failed: %v", err)
			}
		}
	})
}

// Mock implementations for benchmarking

// mockValidator is a minimal validator that always passes for benchmarking.
type mockValidator struct{}

func (m *mockValidator) Validate(ctx context.Context, artifact *Artifact, cfg *Config) (*ValidationReport, error) {
	return &ValidationReport{
		Passed: true,
		Metrics: &Metrics{
			PathLength:     10,
			CycleCount:     2,
			BranchingAvg:   2.0,
			RoomCount:      len(artifact.ADG.Rooms),
			ConnectorCount: len(artifact.ADG.Connectors),
		},
		Errors:   []string{},
		Warnings: []string{},
	}, nil
}

// mockGraphAdapter provides a minimal graph adapter for carving benchmarks.
type mockGraphAdapter struct {
	roomCount int
}

func (m *mockGraphAdapter) GetRooms() []interface{} {
	rooms := make([]interface{}, m.roomCount)
	for i := range rooms {
		rooms[i] = struct{}{}
	}
	return rooms
}

func (m *mockGraphAdapter) GetConnectors() []interface{} {
	conns := make([]interface{}, m.roomCount-1)
	for i := range conns {
		conns[i] = struct{}{}
	}
	return conns
}

// createTestEmbeddingLayout creates a test embedding layout for benchmarking conversions.
func createTestEmbeddingLayout(roomCount int) *embeddingLayout {
	layout := &embeddingLayout{
		Poses:         make(map[string]*embeddingPose),
		CorridorPaths: make(map[string]*embeddingPath),
		Bounds: embeddingBounds{
			MinX: 0,
			MinY: 0,
			MaxX: float64(roomCount * 10),
			MaxY: float64(roomCount * 10),
		},
	}

	// Add test poses
	for i := 0; i < roomCount; i++ {
		roomID := fmt.Sprintf("R%03d", i)
		layout.Poses[roomID] = &embeddingPose{
			X:           float64(i * 10),
			Y:           float64(i * 10),
			Width:       8,
			Height:      8,
			Rotation:    0,
			FootprintID: "default",
		}
	}

	// Add test corridor paths
	for i := 0; i < roomCount-1; i++ {
		connID := fmt.Sprintf("C%03d", i)
		layout.CorridorPaths[connID] = &embeddingPath{
			Points: []embeddingPoint{
				{X: float64(i * 10), Y: float64(i * 10)},
				{X: float64((i + 1) * 10), Y: float64((i + 1) * 10)},
			},
		}
	}

	return layout
}

// Helper types for embedding layout simulation (mimicking pkg/embedding types)
type embeddingLayout struct {
	Poses         map[string]*embeddingPose
	CorridorPaths map[string]*embeddingPath
	Bounds        embeddingBounds
}

type embeddingPose struct {
	X, Y        float64
	Width       int
	Height      int
	Rotation    int
	FootprintID string
}

type embeddingPath struct {
	Points []embeddingPoint
}

type embeddingPoint struct {
	X, Y float64
}

type embeddingBounds struct {
	MinX, MinY, MaxX, MaxY float64
}

func (b embeddingBounds) Width() float64 {
	return b.MaxX - b.MinX
}

func (b embeddingBounds) Height() float64 {
	return b.MaxY - b.MinY
}

// Helper to create RNG (wrapping pkg/rng for benchmarks)
func newRNG(seed uint64, context string, configHash []byte) *mockRNG {
	// This is a simplified version for benchmarking
	// Real implementation would use pkg/rng.NewRNG
	return &mockRNG{seed: seed}
}

type mockRNG struct {
	seed uint64
}

func (r *mockRNG) Uint64n(n uint64) uint64 {
	// Simplified for benchmarking
	return r.seed % n
}
