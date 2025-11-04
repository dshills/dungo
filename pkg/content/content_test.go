package content

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// TestDefaultContentPass_Place tests the complete content placement pipeline.
func TestDefaultContentPass_Place(t *testing.T) {
	tests := []struct {
		name         string
		setupGraph   func() *graph.Graph
		wantErr      bool
		checkContent func(*testing.T, *Content)
	}{
		{
			name: "simple linear dungeon",
			setupGraph: func() *graph.Graph {
				g := graph.NewGraph(12345)

				// Create simple path: Start -> Room1 -> Room2 -> Boss
				start := &graph.Room{
					ID: "start", Archetype: graph.ArchetypeStart,
					Size: graph.SizeM, Difficulty: 0.0, Reward: 0.0,
				}
				room1 := &graph.Room{
					ID: "room1", Archetype: graph.ArchetypeOptional,
					Size: graph.SizeM, Difficulty: 0.5, Reward: 0.6,
				}
				room2 := &graph.Room{
					ID: "room2", Archetype: graph.ArchetypeTreasure,
					Size: graph.SizeL, Difficulty: 0.3, Reward: 0.9,
				}
				boss := &graph.Room{
					ID: "boss", Archetype: graph.ArchetypeBoss,
					Size: graph.SizeXL, Difficulty: 1.0, Reward: 1.0,
				}

				_ = g.AddRoom(start)
				_ = g.AddRoom(room1)
				_ = g.AddRoom(room2)
				_ = g.AddRoom(boss)

				_ = g.AddConnector(&graph.Connector{
					ID: "c1", From: "start", To: "room1",
					Type: graph.TypeDoor, Cost: 1.0,
					Bidirectional: true, Visibility: graph.VisibilityNormal,
				})
				_ = g.AddConnector(&graph.Connector{
					ID: "c2", From: "room1", To: "room2",
					Type: graph.TypeDoor, Cost: 1.0,
					Bidirectional: true, Visibility: graph.VisibilityNormal,
				})
				_ = g.AddConnector(&graph.Connector{
					ID: "c3", From: "room2", To: "boss",
					Type: graph.TypeDoor, Cost: 1.0,
					Bidirectional: true, Visibility: graph.VisibilityNormal,
				})

				return g
			},
			wantErr: false,
			checkContent: func(t *testing.T, c *Content) {
				// Should have enemy spawns (not in Start, Treasure)
				if len(c.Spawns) == 0 {
					t.Error("expected enemy spawns, got none")
				}

				// Should have loot (from Treasure and Boss rooms)
				if len(c.Loot) == 0 {
					t.Error("expected loot, got none")
				}

				// No spawns in start room
				for _, spawn := range c.Spawns {
					if spawn.RoomID == "start" {
						t.Error("found spawn in start room")
					}
				}
			},
		},
		{
			name: "dungeon with key-locked door",
			setupGraph: func() *graph.Graph {
				g := graph.NewGraph(54321)

				start := &graph.Room{
					ID: "start", Archetype: graph.ArchetypeStart,
					Size: graph.SizeM, Difficulty: 0.0, Reward: 0.0,
				}
				keyRoom := &graph.Room{
					ID: "key_room", Archetype: graph.ArchetypeTreasure,
					Size: graph.SizeM, Difficulty: 0.2, Reward: 0.8,
				}
				lockedRoom := &graph.Room{
					ID: "locked_room", Archetype: graph.ArchetypeOptional,
					Size: graph.SizeM, Difficulty: 0.7, Reward: 0.5,
				}

				_ = g.AddRoom(start)
				_ = g.AddRoom(keyRoom)
				_ = g.AddRoom(lockedRoom)

				_ = g.AddConnector(&graph.Connector{
					ID: "c1", From: "start", To: "key_room",
					Type: graph.TypeDoor, Cost: 1.0,
					Bidirectional: true, Visibility: graph.VisibilityNormal,
				})
				_ = g.AddConnector(&graph.Connector{
					ID: "c2", From: "key_room", To: "locked_room",
					Type: graph.TypeDoor, Cost: 1.0,
					Gate:          &graph.Gate{Type: "key", Value: "silver"},
					Bidirectional: true, Visibility: graph.VisibilityNormal,
				})

				return g
			},
			wantErr: false,
			checkContent: func(t *testing.T, c *Content) {
				// Should have a key placed before the lock
				foundKey := false
				for _, loot := range c.Loot {
					if loot.Required && loot.ItemType == "key_silver" {
						foundKey = true
						// Key should be in key_room or start (before lock)
						if loot.RoomID == "locked_room" {
							t.Error("key placed in locked room (after lock)")
						}
					}
				}
				if !foundKey {
					t.Error("required key not placed")
				}
			},
		},
		{
			name: "dungeon with puzzle room",
			setupGraph: func() *graph.Graph {
				g := graph.NewGraph(99999)

				start := &graph.Room{
					ID: "start", Archetype: graph.ArchetypeStart,
					Size: graph.SizeM, Difficulty: 0.0, Reward: 0.0,
				}
				puzzle := &graph.Room{
					ID: "puzzle", Archetype: graph.ArchetypePuzzle,
					Size: graph.SizeM, Difficulty: 0.6, Reward: 0.4,
				}

				_ = g.AddRoom(start)
				_ = g.AddRoom(puzzle)

				_ = g.AddConnector(&graph.Connector{
					ID: "c1", From: "start", To: "puzzle",
					Type: graph.TypeDoor, Cost: 1.0,
					Bidirectional: true, Visibility: graph.VisibilityNormal,
				})

				return g
			},
			wantErr: false,
			checkContent: func(t *testing.T, c *Content) {
				// Should have a puzzle instance
				if len(c.Puzzles) != 1 {
					t.Errorf("expected 1 puzzle, got %d", len(c.Puzzles))
				}
				if len(c.Puzzles) > 0 && c.Puzzles[0].RoomID != "puzzle" {
					t.Error("puzzle not placed in puzzle room")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setupGraph()
			r := rng.NewRNG(12345, "content_test", []byte("test"))

			pass := NewDefaultContentPass()
			content, err := pass.Place(context.Background(), g, r)

			if (err != nil) != tt.wantErr {
				t.Errorf("Place() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkContent != nil {
				tt.checkContent(t, content)
			}
		})
	}
}

// TestContent_Validate tests content validation.
func TestContent_Validate(t *testing.T) {
	g := graph.NewGraph(12345)
	start := &graph.Room{
		ID: "start", Archetype: graph.ArchetypeStart,
		Size: graph.SizeM, Difficulty: 0.0, Reward: 0.0,
	}
	_ = g.AddRoom(start)

	tests := []struct {
		name    string
		content *Content
		wantErr bool
	}{
		{
			name: "valid content",
			content: &Content{
				Spawns: []Spawn{
					{ID: "s1", RoomID: "start", EnemyType: "rat", Count: 1, Position: Point{0, 0}},
				},
				Loot: []Loot{
					{ID: "l1", RoomID: "start", ItemType: "gold", Value: 10, Position: Point{0, 0}},
				},
			},
			wantErr: false,
		},
		{
			name: "spawn with non-existent room",
			content: &Content{
				Spawns: []Spawn{
					{ID: "s1", RoomID: "missing", EnemyType: "rat", Count: 1, Position: Point{0, 0}},
				},
			},
			wantErr: true,
		},
		{
			name: "loot with non-existent room",
			content: &Content{
				Loot: []Loot{
					{ID: "l1", RoomID: "missing", ItemType: "gold", Value: 10, Position: Point{0, 0}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.content.Validate(g)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSpawnValidate tests spawn validation.
func TestSpawn_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spawn   Spawn
		wantErr bool
	}{
		{
			name: "valid spawn",
			spawn: Spawn{
				ID: "s1", RoomID: "room1", EnemyType: "rat",
				Count: 5, Position: Point{10, 20},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			spawn: Spawn{
				ID: "", RoomID: "room1", EnemyType: "rat", Count: 1,
			},
			wantErr: true,
		},
		{
			name: "empty RoomID",
			spawn: Spawn{
				ID: "s1", RoomID: "", EnemyType: "rat", Count: 1,
			},
			wantErr: true,
		},
		{
			name: "empty EnemyType",
			spawn: Spawn{
				ID: "s1", RoomID: "room1", EnemyType: "", Count: 1,
			},
			wantErr: true,
		},
		{
			name: "zero count",
			spawn: Spawn{
				ID: "s1", RoomID: "room1", EnemyType: "rat", Count: 0,
			},
			wantErr: true,
		},
		{
			name: "negative count",
			spawn: Spawn{
				ID: "s1", RoomID: "room1", EnemyType: "rat", Count: -5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spawn.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLoot_Validate tests loot validation.
func TestLoot_Validate(t *testing.T) {
	tests := []struct {
		name    string
		loot    Loot
		wantErr bool
	}{
		{
			name: "valid loot",
			loot: Loot{
				ID: "l1", RoomID: "room1", ItemType: "gold",
				Value: 100, Position: Point{5, 5},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			loot: Loot{
				ID: "", RoomID: "room1", ItemType: "gold", Value: 100,
			},
			wantErr: true,
		},
		{
			name: "empty RoomID",
			loot: Loot{
				ID: "l1", RoomID: "", ItemType: "gold", Value: 100,
			},
			wantErr: true,
		},
		{
			name: "empty ItemType",
			loot: Loot{
				ID: "l1", RoomID: "room1", ItemType: "", Value: 100,
			},
			wantErr: true,
		},
		{
			name: "negative value",
			loot: Loot{
				ID: "l1", RoomID: "room1", ItemType: "gold", Value: -10,
			},
			wantErr: true,
		},
		{
			name: "zero value is valid",
			loot: Loot{
				ID: "l1", RoomID: "room1", ItemType: "key", Value: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.loot.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPuzzleInstance_Validate tests puzzle validation.
func TestPuzzleInstance_Validate(t *testing.T) {
	tests := []struct {
		name    string
		puzzle  PuzzleInstance
		wantErr bool
	}{
		{
			name: "valid puzzle",
			puzzle: PuzzleInstance{
				ID: "p1", RoomID: "room1", Type: "lever",
				Difficulty: 0.5,
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			puzzle: PuzzleInstance{
				ID: "", RoomID: "room1", Type: "lever", Difficulty: 0.5,
			},
			wantErr: true,
		},
		{
			name: "empty RoomID",
			puzzle: PuzzleInstance{
				ID: "p1", RoomID: "", Type: "lever", Difficulty: 0.5,
			},
			wantErr: true,
		},
		{
			name: "empty Type",
			puzzle: PuzzleInstance{
				ID: "p1", RoomID: "room1", Type: "", Difficulty: 0.5,
			},
			wantErr: true,
		},
		{
			name: "difficulty too low",
			puzzle: PuzzleInstance{
				ID: "p1", RoomID: "room1", Type: "lever", Difficulty: -0.1,
			},
			wantErr: true,
		},
		{
			name: "difficulty too high",
			puzzle: PuzzleInstance{
				ID: "p1", RoomID: "room1", Type: "lever", Difficulty: 1.1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.puzzle.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSecretInstance_Validate tests secret validation.
func TestSecretInstance_Validate(t *testing.T) {
	tests := []struct {
		name    string
		secret  SecretInstance
		wantErr bool
	}{
		{
			name: "valid secret",
			secret: SecretInstance{
				ID: "sec1", RoomID: "room1", Type: "hidden_door",
				Position: Point{3, 7},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			secret: SecretInstance{
				ID: "", RoomID: "room1", Type: "hidden_door",
			},
			wantErr: true,
		},
		{
			name: "empty RoomID",
			secret: SecretInstance{
				ID: "sec1", RoomID: "", Type: "hidden_door",
			},
			wantErr: true,
		},
		{
			name: "empty Type",
			secret: SecretInstance{
				ID: "sec1", RoomID: "room1", Type: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.secret.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSelectEnemyType tests enemy selection logic.
func TestSelectEnemyType(t *testing.T) {
	r := rng.NewRNG(42, "test", []byte("test"))

	tests := []struct {
		name       string
		difficulty float64
		wantValid  bool
	}{
		{"low difficulty", 0.1, true},
		{"medium difficulty", 0.5, true},
		{"high difficulty", 0.9, true},
		{"zero difficulty", 0.0, true},
		{"max difficulty", 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enemy := selectEnemyType(tt.difficulty, r)
			if enemy == "" {
				t.Error("selectEnemyType returned empty string")
			}
		})
	}
}

// TestSelectLootType tests loot type selection logic.
func TestSelectLootType(t *testing.T) {
	r := rng.NewRNG(42, "test", []byte("test"))

	tests := []struct {
		name  string
		value int
	}{
		{"low value", 10},
		{"medium value", 100},
		{"high value", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loot := selectLootType(tt.value, r)
			if loot == "" {
				t.Error("selectLootType returned empty string")
			}
		})
	}
}

// TestDeterminism verifies that content placement is deterministic.
func TestDeterminism(t *testing.T) {
	setupGraph := func() *graph.Graph {
		g := graph.NewGraph(12345)

		start := &graph.Room{
			ID: "start", Archetype: graph.ArchetypeStart,
			Size: graph.SizeM, Difficulty: 0.0, Reward: 0.0,
		}
		room1 := &graph.Room{
			ID: "room1", Archetype: graph.ArchetypeOptional,
			Size: graph.SizeM, Difficulty: 0.5, Reward: 0.6,
		}
		boss := &graph.Room{
			ID: "boss", Archetype: graph.ArchetypeBoss,
			Size: graph.SizeXL, Difficulty: 1.0, Reward: 1.0,
		}

		_ = g.AddRoom(start)
		_ = g.AddRoom(room1)
		_ = g.AddRoom(boss)

		_ = g.AddConnector(&graph.Connector{
			ID: "c1", From: "start", To: "room1",
			Type: graph.TypeDoor, Cost: 1.0,
			Bidirectional: true, Visibility: graph.VisibilityNormal,
		})
		_ = g.AddConnector(&graph.Connector{
			ID: "c2", From: "room1", To: "boss",
			Type: graph.TypeDoor, Cost: 1.0,
			Bidirectional: true, Visibility: graph.VisibilityNormal,
		})

		return g
	}

	// Generate content twice with same seed
	g1 := setupGraph()
	r1 := rng.NewRNG(99999, "content_determinism", []byte("test"))
	pass1 := NewDefaultContentPass()
	content1, err1 := pass1.Place(context.Background(), g1, r1)

	if err1 != nil {
		t.Fatalf("first Place() failed: %v", err1)
	}

	g2 := setupGraph()
	r2 := rng.NewRNG(99999, "content_determinism", []byte("test"))
	pass2 := NewDefaultContentPass()
	content2, err2 := pass2.Place(context.Background(), g2, r2)

	if err2 != nil {
		t.Fatalf("second Place() failed: %v", err2)
	}

	// Compare results
	if len(content1.Spawns) != len(content2.Spawns) {
		t.Errorf("spawn count differs: %d vs %d", len(content1.Spawns), len(content2.Spawns))
	}

	if len(content1.Loot) != len(content2.Loot) {
		t.Errorf("loot count differs: %d vs %d", len(content1.Loot), len(content2.Loot))
	}

	// Check that spawn enemy types match
	for i := 0; i < len(content1.Spawns) && i < len(content2.Spawns); i++ {
		if content1.Spawns[i].EnemyType != content2.Spawns[i].EnemyType {
			t.Errorf("spawn %d enemy type differs: %s vs %s",
				i, content1.Spawns[i].EnemyType, content2.Spawns[i].EnemyType)
		}
	}
}

// TestCapacityLimits verifies that capacity limits are respected.
func TestCapacityLimits(t *testing.T) {
	g := graph.NewGraph(12345)

	start := &graph.Room{
		ID: "start", Archetype: graph.ArchetypeStart,
		Size: graph.SizeM, Difficulty: 0.0, Reward: 0.0,
	}
	highDifficulty := &graph.Room{
		ID: "hard", Archetype: graph.ArchetypeOptional,
		Size: graph.SizeXL, Difficulty: 1.0, Reward: 0.5,
	}

	_ = g.AddRoom(start)
	_ = g.AddRoom(highDifficulty)

	_ = g.AddConnector(&graph.Connector{
		ID: "c1", From: "start", To: "hard",
		Type: graph.TypeDoor, Cost: 1.0,
		Bidirectional: true, Visibility: graph.VisibilityNormal,
	})

	r := rng.NewRNG(12345, "capacity_test", []byte("test"))
	maxEnemies := 5

	pass := NewDefaultContentPass().WithMaxEnemiesPerRoom(maxEnemies)
	content, err := pass.Place(context.Background(), g, r)

	if err != nil {
		t.Fatalf("Place() failed: %v", err)
	}

	// Check that no room exceeds capacity
	for _, spawn := range content.Spawns {
		if spawn.Count > maxEnemies {
			t.Errorf("spawn %s exceeds capacity: %d > %d",
				spawn.ID, spawn.Count, maxEnemies)
		}
	}
}
