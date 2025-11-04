package embedding

import (
	"fmt"
	"testing"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// TestPoseValidation tests Pose validation rules.
func TestPoseValidation(t *testing.T) {
	tests := []struct {
		name    string
		pose    Pose
		wantErr bool
	}{
		{
			name: "valid pose",
			pose: Pose{
				X: 10, Y: 20, Width: 5, Height: 8, Rotation: 0,
			},
			wantErr: false,
		},
		{
			name: "valid rotation 90",
			pose: Pose{
				X: 0, Y: 0, Width: 3, Height: 3, Rotation: 90,
			},
			wantErr: false,
		},
		{
			name: "valid rotation 180",
			pose: Pose{
				X: 0, Y: 0, Width: 3, Height: 3, Rotation: 180,
			},
			wantErr: false,
		},
		{
			name: "valid rotation 270",
			pose: Pose{
				X: 0, Y: 0, Width: 3, Height: 3, Rotation: 270,
			},
			wantErr: false,
		},
		{
			name: "zero width",
			pose: Pose{
				X: 0, Y: 0, Width: 0, Height: 5, Rotation: 0,
			},
			wantErr: true,
		},
		{
			name: "negative width",
			pose: Pose{
				X: 0, Y: 0, Width: -5, Height: 5, Rotation: 0,
			},
			wantErr: true,
		},
		{
			name: "zero height",
			pose: Pose{
				X: 0, Y: 0, Width: 5, Height: 0, Rotation: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid rotation 45",
			pose: Pose{
				X: 0, Y: 0, Width: 5, Height: 5, Rotation: 45,
			},
			wantErr: true,
		},
		{
			name: "invalid rotation 360",
			pose: Pose{
				X: 0, Y: 0, Width: 5, Height: 5, Rotation: 360,
			},
			wantErr: true,
		},
		{
			name: "negative rotation",
			pose: Pose{
				X: 0, Y: 0, Width: 5, Height: 5, Rotation: -90,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pose.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPoseBounds tests bounding box calculations.
func TestPoseBounds(t *testing.T) {
	pose := Pose{X: 10, Y: 20, Width: 5, Height: 8, Rotation: 0}
	minX, minY, maxX, maxY := pose.Bounds()

	if minX != 10 || minY != 20 || maxX != 15 || maxY != 28 {
		t.Errorf("Bounds() = (%f, %f, %f, %f), want (10, 20, 15, 28)",
			minX, minY, maxX, maxY)
	}
}

// TestPoseCenter tests center point calculation.
func TestPoseCenter(t *testing.T) {
	pose := Pose{X: 10, Y: 20, Width: 6, Height: 8, Rotation: 0}
	cx, cy := pose.Center()

	if cx != 13 || cy != 24 {
		t.Errorf("Center() = (%f, %f), want (13, 24)", cx, cy)
	}
}

// TestPoseOverlaps tests overlap detection.
func TestPoseOverlaps(t *testing.T) {
	tests := []struct {
		name  string
		pose1 Pose
		pose2 Pose
		want  bool
	}{
		{
			name:  "no overlap - separated horizontally",
			pose1: Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0},
			pose2: Pose{X: 10, Y: 0, Width: 5, Height: 5, Rotation: 0},
			want:  false,
		},
		{
			name:  "no overlap - separated vertically",
			pose1: Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0},
			pose2: Pose{X: 0, Y: 10, Width: 5, Height: 5, Rotation: 0},
			want:  false,
		},
		{
			name:  "no overlap - touching edges",
			pose1: Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0},
			pose2: Pose{X: 5, Y: 0, Width: 5, Height: 5, Rotation: 0},
			want:  false,
		},
		{
			name:  "overlap - partial",
			pose1: Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0},
			pose2: Pose{X: 3, Y: 3, Width: 5, Height: 5, Rotation: 0},
			want:  true,
		},
		{
			name:  "overlap - contained",
			pose1: Pose{X: 0, Y: 0, Width: 10, Height: 10, Rotation: 0},
			pose2: Pose{X: 2, Y: 2, Width: 3, Height: 3, Rotation: 0},
			want:  true,
		},
		{
			name:  "overlap - identical",
			pose1: Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0},
			pose2: Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pose1.Overlaps(&tt.pose2)
			if got != tt.want {
				t.Errorf("Overlaps() = %v, want %v", got, tt.want)
			}

			// Test symmetry
			gotReverse := tt.pose2.Overlaps(&tt.pose1)
			if gotReverse != tt.want {
				t.Errorf("Overlaps() (reversed) = %v, want %v", gotReverse, tt.want)
			}
		})
	}
}

// TestPathLength tests path length calculation.
func TestPathLength(t *testing.T) {
	tests := []struct {
		name string
		path Path
		want float64
	}{
		{
			name: "straight horizontal",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 10, Y: 0}},
			},
			want: 10,
		},
		{
			name: "straight vertical",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 0, Y: 10}},
			},
			want: 10,
		},
		{
			name: "L-shaped",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 8}},
			},
			want: 13, // 5 + 8
		},
		{
			name: "zigzag",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 3, Y: 0}, {X: 3, Y: 4}, {X: 6, Y: 4}},
			},
			want: 10, // 3 + 4 + 3
		},
		{
			name: "single point",
			path: Path{
				Points: []Point{{X: 0, Y: 0}},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.path.Length()
			if got != tt.want {
				t.Errorf("Length() = %f, want %f", got, tt.want)
			}
		})
	}
}

// TestPathBendCount tests bend counting.
func TestPathBendCount(t *testing.T) {
	tests := []struct {
		name string
		path Path
		want int
	}{
		{
			name: "straight line",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 10, Y: 0}},
			},
			want: 0,
		},
		{
			name: "single bend",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 8}},
			},
			want: 1,
		},
		{
			name: "two bends",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 8}, {X: 10, Y: 8}},
			},
			want: 2,
		},
		{
			name: "zigzag three bends",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 3, Y: 0}, {X: 3, Y: 4}, {X: 6, Y: 4}, {X: 6, Y: 8}},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.path.BendCount()
			if got != tt.want {
				t.Errorf("BendCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestPathValidation tests path validation.
func TestPathValidation(t *testing.T) {
	tests := []struct {
		name    string
		path    Path
		wantErr bool
	}{
		{
			name: "valid path",
			path: Path{
				Points: []Point{{X: 0, Y: 0}, {X: 10, Y: 0}},
			},
			wantErr: false,
		},
		{
			name: "valid with door positions",
			path: Path{
				Points:        []Point{{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 8}},
				DoorPositions: []int{0, 1},
			},
			wantErr: false,
		},
		{
			name: "empty path",
			path: Path{
				Points: []Point{},
			},
			wantErr: true,
		},
		{
			name: "single point",
			path: Path{
				Points: []Point{{X: 0, Y: 0}},
			},
			wantErr: true,
		},
		{
			name: "invalid door position negative",
			path: Path{
				Points:        []Point{{X: 0, Y: 0}, {X: 10, Y: 0}},
				DoorPositions: []int{-1},
			},
			wantErr: true,
		},
		{
			name: "invalid door position out of range",
			path: Path{
				Points:        []Point{{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 8}},
				DoorPositions: []int{2}, // Max is 1 (len(points)-2)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.path.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLayoutAddPose tests adding poses to a layout.
func TestLayoutAddPose(t *testing.T) {
	layout := NewLayout()

	// Add valid pose
	pose := &Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0}
	err := layout.AddPose("room1", pose)
	if err != nil {
		t.Errorf("AddPose() unexpected error: %v", err)
	}

	// Check pose was added
	if layout.Poses["room1"] != pose {
		t.Errorf("Pose not added correctly")
	}

	// Try to add nil pose
	err = layout.AddPose("room2", nil)
	if err == nil {
		t.Errorf("AddPose() with nil should error")
	}

	// Try to add invalid pose
	invalidPose := &Pose{X: 0, Y: 0, Width: -5, Height: 5, Rotation: 0}
	err = layout.AddPose("room3", invalidPose)
	if err == nil {
		t.Errorf("AddPose() with invalid pose should error")
	}
}

// TestLayoutComputeBounds tests bounding box computation.
func TestLayoutComputeBounds(t *testing.T) {
	layout := NewLayout()

	// Add some poses
	layout.Poses["room1"] = &Pose{X: 0, Y: 0, Width: 5, Height: 5, Rotation: 0}
	layout.Poses["room2"] = &Pose{X: 10, Y: 10, Width: 8, Height: 6, Rotation: 0}
	layout.Poses["room3"] = &Pose{X: -5, Y: -3, Width: 3, Height: 3, Rotation: 0}

	// Add a corridor path
	layout.CorridorPaths["conn1"] = &Path{
		Points: []Point{{X: 20, Y: 20}, {X: 25, Y: 25}},
	}

	layout.ComputeBounds()

	// Expected bounds: min(-5, -3) to max(25, 25)
	if layout.Bounds.MinX != -5 || layout.Bounds.MinY != -3 {
		t.Errorf("Bounds min = (%f, %f), want (-5, -3)",
			layout.Bounds.MinX, layout.Bounds.MinY)
	}
	if layout.Bounds.MaxX != 25 || layout.Bounds.MaxY != 25 {
		t.Errorf("Bounds max = (%f, %f), want (25, 25)",
			layout.Bounds.MaxX, layout.Bounds.MaxY)
	}
}

// TestConfigValidation tests config validation.
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "default config valid",
			config:  *DefaultConfig(),
			wantErr: false,
		},
		{
			name: "zero max iterations",
			config: Config{
				MaxIterations:     0,
				CorridorMaxLength: 50,
				CorridorMaxBends:  4,
			},
			wantErr: true,
		},
		{
			name: "negative corridor length",
			config: Config{
				MaxIterations:     100,
				CorridorMaxLength: -10,
				CorridorMaxBends:  4,
			},
			wantErr: true,
		},
		{
			name: "negative corridor bends",
			config: Config{
				MaxIterations:     100,
				CorridorMaxLength: 50,
				CorridorMaxBends:  -1,
			},
			wantErr: true,
		},
		{
			name: "invalid damping factor high",
			config: Config{
				MaxIterations:     100,
				CorridorMaxLength: 50,
				DampingFactor:     1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid damping factor negative",
			config: Config{
				MaxIterations:     100,
				CorridorMaxLength: 50,
				DampingFactor:     -0.1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEmbedderRegistry tests the embedder registration system.
func TestEmbedderRegistry(t *testing.T) {
	// Test that force_directed is registered
	embedder, err := Get("force_directed", nil)
	if err != nil {
		t.Errorf("Get(force_directed) unexpected error: %v", err)
	}
	if embedder == nil {
		t.Errorf("Get(force_directed) returned nil")
	}
	if embedder.Name() != "force_directed" {
		t.Errorf("embedder.Name() = %s, want force_directed", embedder.Name())
	}

	// Test getting non-existent embedder
	_, err = Get("nonexistent", nil)
	if err == nil {
		t.Errorf("Get(nonexistent) should return error")
	}

	// Test List
	names := List()
	found := false
	for _, name := range names {
		if name == "force_directed" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("List() should include force_directed, got %v", names)
	}
}

// TestSizeToGridDimensions tests size to dimension conversion.
func TestSizeToGridDimensions(t *testing.T) {
	tests := []struct {
		size       graph.RoomSize
		wantWidth  int
		wantHeight int
	}{
		{graph.SizeXS, 3, 3},
		{graph.SizeS, 5, 5},
		{graph.SizeM, 8, 8},
		{graph.SizeL, 12, 12},
		{graph.SizeXL, 16, 16},
	}

	for _, tt := range tests {
		t.Run(tt.size.String(), func(t *testing.T) {
			w, h := SizeToGridDimensions(tt.size)
			if w != tt.wantWidth || h != tt.wantHeight {
				t.Errorf("SizeToGridDimensions(%v) = (%d, %d), want (%d, %d)",
					tt.size, w, h, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

// TestForceDirectedEmbedSimpleGraph tests embedding a simple graph.
func TestForceDirectedEmbedSimpleGraph(t *testing.T) {
	// Create a simple graph with 3 rooms
	g := graph.NewGraph(12345)

	room1 := &graph.Room{
		ID:         "R001",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Difficulty: 0.0,
		Reward:     0.0,
	}
	room2 := &graph.Room{
		ID:         "R002",
		Archetype:  graph.ArchetypeHub,
		Size:       graph.SizeM,
		Difficulty: 0.5,
		Reward:     0.5,
	}
	room3 := &graph.Room{
		ID:         "R003",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeL,
		Difficulty: 1.0,
		Reward:     1.0,
	}

	if err := g.AddRoom(room1); err != nil {
		t.Fatalf("AddRoom failed: %v", err)
	}
	if err := g.AddRoom(room2); err != nil {
		t.Fatalf("AddRoom failed: %v", err)
	}
	if err := g.AddRoom(room3); err != nil {
		t.Fatalf("AddRoom failed: %v", err)
	}

	conn1 := &graph.Connector{
		ID:            "C001",
		From:          "R001",
		To:            "R002",
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: true,
	}
	conn2 := &graph.Connector{
		ID:            "C002",
		From:          "R002",
		To:            "R003",
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: true,
	}

	if err := g.AddConnector(conn1); err != nil {
		t.Fatalf("AddConnector failed: %v", err)
	}
	if err := g.AddConnector(conn2); err != nil {
		t.Fatalf("AddConnector failed: %v", err)
	}

	// Create RNG
	configHash := []byte("test_config")
	rngInstance := rng.NewRNG(12345, "embedding", configHash)

	// Create embedder and embed
	config := DefaultConfig()
	embedder := NewForceDirectedEmbedder(config)

	layout, err := embedder.Embed(g, rngInstance)
	if err != nil {
		t.Fatalf("Embed() failed: %v", err)
	}

	// Validate results
	if len(layout.Poses) != 3 {
		t.Errorf("Expected 3 poses, got %d", len(layout.Poses))
	}

	if len(layout.CorridorPaths) != 2 {
		t.Errorf("Expected 2 corridor paths, got %d", len(layout.CorridorPaths))
	}

	// Check no overlaps
	for id1, pose1 := range layout.Poses {
		for id2, pose2 := range layout.Poses {
			if id1 >= id2 {
				continue
			}
			if pose1.Overlaps(pose2) {
				t.Errorf("Rooms %s and %s overlap", id1, id2)
			}
		}
	}

	// Validate against config
	if err := ValidateEmbedding(layout, g, config); err != nil {
		t.Errorf("ValidateEmbedding() failed: %v", err)
	}
}

// TestForceDirectedEmbedDeterminism tests that embedding is deterministic.
func TestForceDirectedEmbedDeterminism(t *testing.T) {
	// Create a graph
	g := graph.NewGraph(12345)

	for i := 0; i < 5; i++ {
		room := &graph.Room{
			ID:         fmt.Sprintf("R%03d", i),
			Archetype:  graph.ArchetypeHub,
			Size:       graph.SizeM,
			Difficulty: float64(i) / 4.0,
			Reward:     float64(i) / 4.0,
		}
		if err := g.AddRoom(room); err != nil {
			t.Fatalf("AddRoom failed: %v", err)
		}
	}

	// Create a path through all rooms
	for i := 0; i < 4; i++ {
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
			t.Fatalf("AddConnector failed: %v", err)
		}
	}

	configHash := []byte("test_config")
	config := DefaultConfig()
	embedder := NewForceDirectedEmbedder(config)

	// Embed twice with same seed
	rng1 := rng.NewRNG(12345, "embedding", configHash)
	layout1, err := embedder.Embed(g, rng1)
	if err != nil {
		t.Fatalf("Embed() failed: %v", err)
	}

	rng2 := rng.NewRNG(12345, "embedding", configHash)
	layout2, err := embedder.Embed(g, rng2)
	if err != nil {
		t.Fatalf("Embed() failed: %v", err)
	}

	// Compare layouts
	for roomID, pose1 := range layout1.Poses {
		pose2 := layout2.Poses[roomID]
		if pose1.X != pose2.X || pose1.Y != pose2.Y {
			t.Errorf("Room %s has different positions: (%f, %f) vs (%f, %f)",
				roomID, pose1.X, pose1.Y, pose2.X, pose2.Y)
		}
	}
}
