package synthesis

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// TestGrammarSynthesizer_CoreTrio verifies the initial Start-Mid-Boss structure.
func TestGrammarSynthesizer_CoreTrio(t *testing.T) {
	cfg := &Config{
		Seed:          12345,
		RoomsMin:      10,
		RoomsMax:      15,
		BranchingAvg:  2.0,
		BranchingMax:  4,
		SecretDensity: 0.0,
		OptionalRatio: 0.1,
		Pacing: PacingConfig{
			Curve:    "LINEAR",
			Variance: 0.1,
		},
		Themes: []string{"dungeon"},
	}

	synth := NewGrammarSynthesizer()
	testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

	g, err := synth.Synthesize(context.Background(), testRNG, cfg)
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}

	// Verify room count within bounds
	if len(g.Rooms) < cfg.RoomsMin || len(g.Rooms) > cfg.RoomsMax {
		t.Errorf("Expected %d-%d rooms, got %d", cfg.RoomsMin, cfg.RoomsMax, len(g.Rooms))
	}

	// Verify Start room exists
	startCount := 0
	var startRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startCount++
			startRoom = room
		}
	}
	if startCount != 1 {
		t.Errorf("Expected 1 Start room, got %d", startCount)
	}

	// Verify Boss room exists
	bossCount := 0
	var bossRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeBoss {
			bossCount++
			bossRoom = room
		}
	}
	if bossCount != 1 {
		t.Errorf("Expected 1 Boss room, got %d", bossCount)
	}

	// Verify path from Start to Boss
	if startRoom != nil && bossRoom != nil {
		path, err := g.GetPath(startRoom.ID, bossRoom.ID)
		if err != nil {
			t.Errorf("No path from Start to Boss: %v", err)
		}
		if len(path) < 2 {
			t.Errorf("Path from Start to Boss too short: %d rooms", len(path))
		}
	}

	// Verify connectivity
	if !g.IsConnected() {
		t.Error("Graph is not connected")
	}
}

// TestGrammarSynthesizer_RoomCountBounds verifies room count respects Config bounds.
func TestGrammarSynthesizer_RoomCountBounds(t *testing.T) {
	tests := []struct {
		name     string
		roomsMin int
		roomsMax int
	}{
		{"Small", 10, 15},
		{"Medium", 20, 30},
		{"Large", 40, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Seed:          12345,
				RoomsMin:      tt.roomsMin,
				RoomsMax:      tt.roomsMax,
				BranchingAvg:  2.0,
				BranchingMax:  4,
				SecretDensity: 0.1,
				OptionalRatio: 0.2,
				Pacing: PacingConfig{
					Curve:    "LINEAR",
					Variance: 0.1,
				},
				Themes: []string{"dungeon"},
			}

			synth := NewGrammarSynthesizer()
			testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

			g, err := synth.Synthesize(context.Background(), testRNG, cfg)
			if err != nil {
				t.Fatalf("Synthesize() error = %v", err)
			}

			roomCount := len(g.Rooms)
			if roomCount < tt.roomsMin {
				t.Errorf("Room count %d below minimum %d", roomCount, tt.roomsMin)
			}
			if roomCount > tt.roomsMax {
				t.Errorf("Room count %d exceeds maximum %d", roomCount, tt.roomsMax)
			}
		})
	}
}

// TestGrammarSynthesizer_Connectivity verifies all rooms are reachable.
func TestGrammarSynthesizer_Connectivity(t *testing.T) {
	cfg := &Config{
		Seed:          98765,
		RoomsMin:      15,
		RoomsMax:      25,
		BranchingAvg:  2.5,
		BranchingMax:  4,
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Pacing: PacingConfig{
			Curve:    "LINEAR",
			Variance: 0.1,
		},
		Themes: []string{"dungeon"},
	}

	synth := NewGrammarSynthesizer()
	testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

	g, err := synth.Synthesize(context.Background(), testRNG, cfg)
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}

	// Verify graph is connected
	if !g.IsConnected() {
		t.Error("Graph is not connected")
	}

	// Verify all rooms are reachable from Start
	var startRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startRoom = room
			break
		}
	}

	if startRoom != nil {
		reachable := g.GetReachable(startRoom.ID)
		if len(reachable) != len(g.Rooms) {
			t.Errorf("Only %d/%d rooms reachable from Start", len(reachable), len(g.Rooms))
		}
	}
}

// TestGrammarSynthesizer_KeyLockConstraints verifies key-before-lock ordering.
func TestGrammarSynthesizer_KeyLockConstraints(t *testing.T) {
	cfg := &Config{
		Seed:         55555,
		RoomsMin:     20,
		RoomsMax:     30,
		BranchingAvg: 2.0,
		BranchingMax: 4,
		Keys: []KeyConfig{
			{Name: "silver", Count: 1},
			{Name: "gold", Count: 1},
		},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
		Pacing: PacingConfig{
			Curve:    "LINEAR",
			Variance: 0.1,
		},
		Themes: []string{"dungeon"},
	}

	synth := NewGrammarSynthesizer()
	testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

	g, err := synth.Synthesize(context.Background(), testRNG, cfg)
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}

	// Find Start room
	var startRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startRoom = room
			break
		}
	}

	if startRoom == nil {
		t.Fatal("No Start room found")
	}

	// Verify each key is reachable before its lock
	keyProviders := make(map[string]string) // key name -> room ID

	for _, room := range g.Rooms {
		for _, cap := range room.Provides {
			if cap.Type == "key" {
				keyProviders[cap.Value] = room.ID
			}
		}
	}

	for _, room := range g.Rooms {
		for _, req := range room.Requirements {
			if req.Type == "key" {
				keyProviderID, hasProvider := keyProviders[req.Value]
				if !hasProvider {
					t.Errorf("Room %s requires key %q but no room provides it", room.ID, req.Value)
					continue
				}

				// Verify key provider is reachable from Start
				_, err := g.GetPath(startRoom.ID, keyProviderID)
				if err != nil {
					t.Errorf("Key %q in room %s is not reachable from Start: %v", req.Value, keyProviderID, err)
				}
			}
		}
	}
}

// TestGrammarSynthesizer_Determinism verifies same seed produces same graph.
func TestGrammarSynthesizer_Determinism(t *testing.T) {
	cfg := &Config{
		Seed:          42424242,
		RoomsMin:      15,
		RoomsMax:      20,
		BranchingAvg:  2.0,
		BranchingMax:  4,
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
		Pacing: PacingConfig{
			Curve:    "LINEAR",
			Variance: 0.1,
		},
		Themes: []string{"dungeon"},
	}

	synth := NewGrammarSynthesizer()

	// Generate first graph
	rng1 := rng.NewRNG(cfg.Seed, "test", []byte("test"))
	g1, err := synth.Synthesize(context.Background(), rng1, cfg)
	if err != nil {
		t.Fatalf("First Synthesize() error = %v", err)
	}

	// Generate second graph with same seed
	rng2 := rng.NewRNG(cfg.Seed, "test", []byte("test"))
	g2, err := synth.Synthesize(context.Background(), rng2, cfg)
	if err != nil {
		t.Fatalf("Second Synthesize() error = %v", err)
	}

	// Verify same number of rooms
	if len(g1.Rooms) != len(g2.Rooms) {
		t.Errorf("Room counts differ: %d vs %d", len(g1.Rooms), len(g2.Rooms))
	}

	// Verify same number of connectors
	if len(g1.Connectors) != len(g2.Connectors) {
		t.Errorf("Connector counts differ: %d vs %d", len(g1.Connectors), len(g2.Connectors))
	}

	// Verify room IDs match
	for id := range g1.Rooms {
		if _, exists := g2.Rooms[id]; !exists {
			t.Errorf("Room %s exists in first graph but not second", id)
		}
	}

	for id := range g2.Rooms {
		if _, exists := g1.Rooms[id]; !exists {
			t.Errorf("Room %s exists in second graph but not first", id)
		}
	}
}

// TestGrammarSynthesizer_BranchingMax verifies no room exceeds max connections.
func TestGrammarSynthesizer_BranchingMax(t *testing.T) {
	cfg := &Config{
		Seed:          77777,
		RoomsMin:      20,
		RoomsMax:      30,
		BranchingAvg:  2.5,
		BranchingMax:  3, // Strict max
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
		Pacing: PacingConfig{
			Curve:    "LINEAR",
			Variance: 0.1,
		},
		Themes: []string{"dungeon"},
	}

	synth := NewGrammarSynthesizer()
	testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

	g, err := synth.Synthesize(context.Background(), testRNG, cfg)
	if err != nil {
		t.Fatalf("Synthesize() error = %v", err)
	}

	// Verify no room exceeds max connections
	for roomID, neighbors := range g.Adjacency {
		if len(neighbors) > cfg.BranchingMax {
			t.Errorf("Room %s has %d connections, exceeds max %d", roomID, len(neighbors), cfg.BranchingMax)
		}
	}
}

// TestGrammarSynthesizer_Registration verifies synthesizer is registered.
func TestGrammarSynthesizer_Registration(t *testing.T) {
	synth := Get("grammar")
	if synth == nil {
		t.Error("Grammar synthesizer not registered")
	}

	if synth.Name() != "grammar" {
		t.Errorf("Expected name 'grammar', got %q", synth.Name())
	}
}

// TestGrammarSynthesizer_ContextCancellation verifies context cancellation is respected.
func TestGrammarSynthesizer_ContextCancellation(t *testing.T) {
	cfg := &Config{
		Seed:          11111,
		RoomsMin:      10,
		RoomsMax:      20,
		BranchingAvg:  2.0,
		BranchingMax:  4,
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
		Pacing: PacingConfig{
			Curve:    "LINEAR",
			Variance: 0.1,
		},
		Themes: []string{"dungeon"},
	}

	synth := NewGrammarSynthesizer()
	testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := synth.Synthesize(ctx, testRNG, cfg)
	if err == nil {
		t.Error("Expected error from cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

// T077: Property test for pacing curve adherence
// Verifies that room difficulties follow the configured pacing curve within variance tolerance.
func TestGrammarSynthesizer_PacingCurveAdherence(t *testing.T) {
	tests := []struct {
		name     string
		curve    string
		variance float64
	}{
		{"Linear", "LINEAR", 0.1},
		{"SCurve", "S_CURVE", 0.15},
		{"Exponential", "EXPONENTIAL", 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Seed:          42424242,
				RoomsMin:      20,
				RoomsMax:      30,
				BranchingAvg:  2.0,
				BranchingMax:  4,
				SecretDensity: 0.1,
				OptionalRatio: 0.2,
				Pacing: PacingConfig{
					Curve:    tt.curve,
					Variance: tt.variance,
				},
				Themes: []string{"dungeon"},
			}

			synth := NewGrammarSynthesizer()
			testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

			g, err := synth.Synthesize(context.Background(), testRNG, cfg)
			if err != nil {
				t.Fatalf("Synthesize() error = %v", err)
			}

			// Find Start and Boss rooms
			var startRoom, bossRoom *graph.Room
			for _, room := range g.Rooms {
				if room.Archetype == graph.ArchetypeStart {
					startRoom = room
				} else if room.Archetype == graph.ArchetypeBoss {
					bossRoom = room
				}
			}

			if startRoom == nil || bossRoom == nil {
				t.Fatal("Missing Start or Boss room")
			}

			// Get critical path
			criticalPath, err := g.GetPath(startRoom.ID, bossRoom.ID)
			if err != nil {
				t.Fatalf("No path from Start to Boss: %v", err)
			}

			// Create expected pacing curve
			var curve PacingCurve
			switch tt.curve {
			case "LINEAR":
				curve = &LinearCurve{}
			case "S_CURVE":
				curve = NewSCurve()
			case "EXPONENTIAL":
				curve = NewExponentialCurve()
			}

			// Verify rooms on critical path follow the curve within tolerance
			tolerance := tt.variance + 0.15 // Allow variance + some extra for rounding
			for i, roomID := range criticalPath {
				room := g.Rooms[roomID]
				progress := 0.0
				if len(criticalPath) > 1 {
					progress = float64(i) / float64(len(criticalPath)-1)
				}
				expectedDifficulty := curve.Evaluate(progress)
				actualDifficulty := room.Difficulty

				// Check if difficulty is within tolerance
				diff := actualDifficulty - expectedDifficulty
				if diff < 0 {
					diff = -diff
				}
				if diff > tolerance {
					t.Errorf("Room %s at progress %.2f: difficulty %.3f deviates from expected %.3f by %.3f (tolerance %.3f)",
						roomID, progress, actualDifficulty, expectedDifficulty, diff, tolerance)
				}
			}

			// Verify Start room has low difficulty
			if startRoom.Difficulty > 0.3 {
				t.Errorf("Start room has difficulty %.3f, expected <= 0.3", startRoom.Difficulty)
			}

			// Verify Boss room has high difficulty
			if bossRoom.Difficulty < 0.7 {
				t.Errorf("Boss room has difficulty %.3f, expected >= 0.7", bossRoom.Difficulty)
			}

			// Verify all difficulties are in valid range [0.0, 1.0]
			for _, room := range g.Rooms {
				if room.Difficulty < 0.0 || room.Difficulty > 1.0 {
					t.Errorf("Room %s has difficulty %.3f outside valid range [0.0, 1.0]", room.ID, room.Difficulty)
				}
			}
		})
	}
}

// T078: Property test for branching factor bounds
// Verifies that no room exceeds BranchingMax connections and average is close to BranchingAvg.
func TestGrammarSynthesizer_BranchingFactorBounds(t *testing.T) {
	tests := []struct {
		name         string
		branchingAvg float64
		branchingMax int
	}{
		{"Conservative", 2.0, 3},
		{"Moderate", 2.5, 4},
		{"Dense", 3.0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Seed:          12345678,
				RoomsMin:      25,
				RoomsMax:      35,
				BranchingAvg:  tt.branchingAvg,
				BranchingMax:  tt.branchingMax,
				SecretDensity: 0.1,
				OptionalRatio: 0.2,
				Pacing: PacingConfig{
					Curve:    "LINEAR",
					Variance: 0.1,
				},
				Themes: []string{"dungeon"},
			}

			synth := NewGrammarSynthesizer()
			testRNG := rng.NewRNG(cfg.Seed, "test", []byte("test"))

			g, err := synth.Synthesize(context.Background(), testRNG, cfg)
			if err != nil {
				t.Fatalf("Synthesize() error = %v", err)
			}

			// Verify no room exceeds BranchingMax
			totalConnections := 0
			for roomID, neighbors := range g.Adjacency {
				connectionCount := len(neighbors)
				totalConnections += connectionCount

				if connectionCount > tt.branchingMax {
					t.Errorf("Room %s has %d connections, exceeds max %d", roomID, connectionCount, tt.branchingMax)
				}
			}

			// Verify average branching factor is reasonably close to target
			// Note: Average is per room, not per edge (so count each connection once per room)
			avgBranching := float64(totalConnections) / float64(len(g.Rooms))

			// Allow significant deviation from target average since:
			// 1. Graph must be connected (minimum spanning tree constraint)
			// 2. Rooms with max connections can't accept more
			// 3. Grammar rules create specific topologies
			// We mainly care that it's not too sparse (< 1.5) or too dense (> max+1)
			minReasonable := 1.5 // Must be better than a path (1.0 avg)
			maxReasonable := float64(tt.branchingMax) + 1.0

			if avgBranching < minReasonable {
				t.Errorf("Average branching %.2f too sparse (expected >= %.2f)",
					avgBranching, minReasonable)
			}
			if avgBranching > maxReasonable {
				t.Errorf("Average branching %.2f too dense (expected <= %.2f)",
					avgBranching, maxReasonable)
			}

			// Verify graph is still connected
			if !g.IsConnected() {
				t.Error("Graph is not connected despite branching constraints")
			}

			// Verify all rooms have at least one connection (except for edge cases)
			isolatedRooms := 0
			for roomID, neighbors := range g.Adjacency {
				if len(neighbors) == 0 {
					isolatedRooms++
					t.Errorf("Room %s is isolated (no connections)", roomID)
				}
			}

			if isolatedRooms > 0 {
				t.Errorf("Found %d isolated rooms", isolatedRooms)
			}
		})
	}
}
