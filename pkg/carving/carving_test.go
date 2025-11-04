package carving

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/graph"
)

// TestCarverRegistry tests the carver registry functionality.
func TestCarverRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		registry := NewCarverRegistry()
		carver := NewDefaultCarver(16, 16)

		err := registry.Register("default", carver)
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		retrieved, err := registry.Get("default")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if retrieved != carver {
			t.Error("Retrieved carver does not match registered carver")
		}
	})

	t.Run("Register duplicate", func(t *testing.T) {
		registry := NewCarverRegistry()
		carver := NewDefaultCarver(16, 16)

		_ = registry.Register("default", carver)
		err := registry.Register("default", carver)
		if err == nil {
			t.Error("Expected error when registering duplicate, got nil")
		}
	})

	t.Run("Get non-existent", func(t *testing.T) {
		registry := NewCarverRegistry()
		_, err := registry.Get("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent carver, got nil")
		}
	})

	t.Run("List carvers", func(t *testing.T) {
		registry := NewCarverRegistry()
		_ = registry.Register("carver1", NewDefaultCarver(16, 16))
		_ = registry.Register("carver2", NewDefaultCarver(32, 32))

		names := registry.List()
		if len(names) != 2 {
			t.Errorf("List() returned %d carvers, want 2", len(names))
		}
	})
}

// TestTileMapHelpers tests tile map helper functions.
func TestTileMapHelpers(t *testing.T) {
	t.Run("NewTileMap", func(t *testing.T) {
		tm := NewTileMap(100, 100, 16, 16)
		if tm.Width != 100 || tm.Height != 100 {
			t.Errorf("NewTileMap() dimensions = %dx%d, want 100x100", tm.Width, tm.Height)
		}
		if tm.TileWidth != 16 || tm.TileHeight != 16 {
			t.Errorf("NewTileMap() tile size = %dx%d, want 16x16", tm.TileWidth, tm.TileHeight)
		}
	})

	t.Run("AddLayer", func(t *testing.T) {
		tm := NewTileMap(100, 100, 16, 16)
		layer := AddLayer(tm, "floor", "tilelayer")

		if layer == nil {
			t.Fatal("AddLayer() returned nil")
		}
		if layer.Name != "floor" {
			t.Errorf("AddLayer() name = %s, want floor", layer.Name)
		}
		if len(layer.Data) != 10000 {
			t.Errorf("AddLayer() data length = %d, want 10000", len(layer.Data))
		}
	})

	t.Run("GetTile and SetTile", func(t *testing.T) {
		data := make([]uint32, 100)
		width, height := 10, 10

		err := SetTile(data, 5, 5, width, height, 42)
		if err != nil {
			t.Fatalf("SetTile() error = %v", err)
		}

		value := GetTile(data, 5, 5, width, height)
		if value != 42 {
			t.Errorf("GetTile() = %d, want 42", value)
		}
	})

	t.Run("FillRect", func(t *testing.T) {
		data := make([]uint32, 100)
		width, height := 10, 10

		err := FillRect(data, 2, 2, 3, 3, width, height, 5)
		if err != nil {
			t.Fatalf("FillRect() error = %v", err)
		}

		// Check that interior is filled
		for y := 2; y < 5; y++ {
			for x := 2; x < 5; x++ {
				if GetTile(data, x, y, width, height) != 5 {
					t.Errorf("FillRect() tile at (%d, %d) = %d, want 5", x, y, GetTile(data, x, y, width, height))
				}
			}
		}
	})

	t.Run("DrawLine", func(t *testing.T) {
		data := make([]uint32, 100)
		width, height := 10, 10

		err := DrawLine(data, 0, 0, 5, 5, width, height, 1)
		if err != nil {
			t.Fatalf("DrawLine() error = %v", err)
		}

		// Check that diagonal is drawn
		nonZeroCount := 0
		for i := 0; i < len(data); i++ {
			if data[i] != 0 {
				nonZeroCount++
			}
		}

		if nonZeroCount < 5 {
			t.Errorf("DrawLine() drew %d tiles, want at least 5", nonZeroCount)
		}
	})
}

// TestStamper tests room stamping functionality.
func TestStamper(t *testing.T) {
	t.Run("StampRoom XS", func(t *testing.T) {
		stamper := NewStamper(50, 50)
		data := make([]uint32, 2500)
		room := &graph.Room{
			ID:   "room1",
			Size: graph.SizeXS,
		}
		adapter := &RoomAdapter{room: room}
		pose := Pose{X: 25, Y: 25, Rotation: 0}

		err := stamper.StampRoom(adapter, pose, data)
		if err != nil {
			t.Fatalf("StampRoom() error = %v", err)
		}

		// Count floor tiles (should be 3x3 = 9)
		floorCount := 0
		for _, tile := range data {
			if tile == uint32(TileFloor) {
				floorCount++
			}
		}

		if floorCount != 9 {
			t.Errorf("StampRoom(XS) created %d floor tiles, want 9", floorCount)
		}
	})

	t.Run("StampRoom sizes", func(t *testing.T) {
		tests := []struct {
			size     graph.RoomSize
			expected int // Expected floor tile count
		}{
			{graph.SizeXS, 9},   // 3x3
			{graph.SizeS, 25},   // 5x5
			{graph.SizeM, 49},   // 7x7
			{graph.SizeL, 100},  // 10x10
			{graph.SizeXL, 225}, // 15x15
		}

		for _, tt := range tests {
			t.Run(tt.size.String(), func(t *testing.T) {
				stamper := NewStamper(100, 100)
				data := make([]uint32, 10000)
				room := &graph.Room{
					ID:   "test",
					Size: tt.size,
				}
				adapter := &RoomAdapter{room: room}
				pose := Pose{X: 50, Y: 50, Rotation: 0}

				err := stamper.StampRoom(adapter, pose, data)
				if err != nil {
					t.Fatalf("StampRoom(%s) error = %v", tt.size, err)
				}

				floorCount := 0
				for _, tile := range data {
					if tile == uint32(TileFloor) {
						floorCount++
					}
				}

				if floorCount != tt.expected {
					t.Errorf("StampRoom(%s) created %d floor tiles, want %d", tt.size, floorCount, tt.expected)
				}
			})
		}
	})

	t.Run("StampShape rectangle", func(t *testing.T) {
		stamper := NewStamper(50, 50)
		data := make([]uint32, 2500)

		err := stamper.StampShape("rect", 10, 10, 5, 5, data)
		if err != nil {
			t.Fatalf("StampShape(rect) error = %v", err)
		}

		floorCount := 0
		for _, tile := range data {
			if tile == uint32(TileFloor) {
				floorCount++
			}
		}

		if floorCount != 25 {
			t.Errorf("StampShape(rect) created %d floor tiles, want 25", floorCount)
		}
	})
}

// TestCorridorRouter tests corridor routing functionality.
func TestCorridorRouter(t *testing.T) {
	t.Run("RouteCorridor simple", func(t *testing.T) {
		router := NewCorridorRouter(50, 50)
		data := make([]uint32, 2500)

		path := Path{
			Points: []Point{
				{X: 10, Y: 10},
				{X: 20, Y: 10},
				{X: 20, Y: 20},
			},
		}

		err := router.RouteCorridor(path, data)
		if err != nil {
			t.Fatalf("RouteCorridor() error = %v", err)
		}

		// Check that corridor was drawn
		floorCount := 0
		for _, tile := range data {
			if tile == uint32(TileFloor) {
				floorCount++
			}
		}

		if floorCount < 20 {
			t.Errorf("RouteCorridor() created %d floor tiles, want at least 20", floorCount)
		}
	})

	t.Run("RouteCorridor too few points", func(t *testing.T) {
		router := NewCorridorRouter(50, 50)
		data := make([]uint32, 2500)

		path := Path{
			Points: []Point{{X: 10, Y: 10}},
		}

		err := router.RouteCorridor(path, data)
		if err == nil {
			t.Error("RouteCorridor() with 1 point should error, got nil")
		}
	})

	t.Run("RouteWideCorridor", func(t *testing.T) {
		router := NewCorridorRouter(50, 50)
		data := make([]uint32, 2500)

		path := Path{
			Points: []Point{
				{X: 10, Y: 10},
				{X: 30, Y: 10},
			},
		}

		err := router.RouteWideCorridor(path, 3, data)
		if err != nil {
			t.Fatalf("RouteWideCorridor() error = %v", err)
		}

		// Wide corridor should have more tiles than single-width
		floorCount := 0
		for _, tile := range data {
			if tile == uint32(TileFloor) {
				floorCount++
			}
		}

		if floorCount < 40 {
			t.Errorf("RouteWideCorridor() created %d floor tiles, want at least 40", floorCount)
		}
	})
}

// TestDefaultCarver tests the complete carving pipeline.
func TestDefaultCarver(t *testing.T) {
	t.Run("Carve simple dungeon", func(t *testing.T) {
		// Create a simple graph with 2 rooms and 1 corridor
		rooms := map[string]*graph.Room{
			"room1": {
				ID:   "room1",
				Size: graph.SizeM,
			},
			"room2": {
				ID:   "room2",
				Size: graph.SizeM,
			},
		}
		connectors := map[string]*graph.Connector{
			"conn1": {
				ID:            "conn1",
				From:          "room1",
				To:            "room2",
				Type:          graph.TypeCorridor,
				Bidirectional: true,
				Cost:          1.0,
			},
		}
		g := NewGraphAdapter(rooms, connectors)

		layout := &Layout{
			Poses: map[string]Pose{
				"room1": {X: 20, Y: 20, Rotation: 0},
				"room2": {X: 60, Y: 60, Rotation: 0},
			},
			CorridorPaths: map[string]Path{
				"conn1": {
					Points: []Point{
						{X: 20, Y: 20},
						{X: 40, Y: 20},
						{X: 40, Y: 60},
						{X: 60, Y: 60},
					},
				},
			},
			Bounds: Rect{Width: 100, Height: 100},
		}

		carver := NewDefaultCarver(16, 16)
		tm, err := carver.Carve(context.Background(), g, layout)
		if err != nil {
			t.Fatalf("Carve() error = %v", err)
		}

		if tm == nil {
			t.Fatal("Carve() returned nil TileMap")
		}

		if tm.Width != 100 || tm.Height != 100 {
			t.Errorf("Carve() TileMap dimensions = %dx%d, want 100x100", tm.Width, tm.Height)
		}

		// Check that layers were created
		if _, ok := tm.Layers["floor"]; !ok {
			t.Error("Carve() did not create floor layer")
		}
		if _, ok := tm.Layers["walls"]; !ok {
			t.Error("Carve() did not create walls layer")
		}
		if _, ok := tm.Layers["doors"]; !ok {
			t.Error("Carve() did not create doors layer")
		}

		// Check that floor has tiles
		floorLayer := tm.Layers["floor"]
		floorCount := 0
		for _, tile := range floorLayer.Data {
			if tile != 0 {
				floorCount++
			}
		}

		if floorCount == 0 {
			t.Error("Carve() floor layer has no tiles")
		}

		// Check that walls were generated
		wallLayer := tm.Layers["walls"]
		wallCount := 0
		for _, tile := range wallLayer.Data {
			if tile != 0 {
				wallCount++
			}
		}

		if wallCount == 0 {
			t.Error("Carve() wall layer has no tiles")
		}

		// Check that doors were placed
		doorLayer := tm.Layers["doors"]
		if len(doorLayer.Objects) == 0 {
			t.Error("Carve() did not place any doors")
		}
	})

	t.Run("Carve with nil inputs", func(t *testing.T) {
		carver := NewDefaultCarver(16, 16)

		_, err := carver.Carve(context.Background(), nil, nil)
		if err == nil {
			t.Error("Carve() with nil graph should error, got nil")
		}

		g := NewGraphAdapter(make(map[string]*graph.Room), make(map[string]*graph.Connector))
		_, err = carver.Carve(context.Background(), g, nil)
		if err == nil {
			t.Error("Carve() with nil layout should error, got nil")
		}
	})
}

// TestPathUtilities tests path manipulation utilities.
func TestPathUtilities(t *testing.T) {
	t.Run("SimplifyPath", func(t *testing.T) {
		path := Path{
			Points: []Point{
				{X: 0, Y: 0},
				{X: 5, Y: 0},
				{X: 10, Y: 0}, // Collinear, should be removed
				{X: 10, Y: 5},
				{X: 10, Y: 10}, // Collinear, should be removed
				{X: 15, Y: 10},
			},
		}

		simplified := SimplifyPath(path)
		if len(simplified.Points) >= len(path.Points) {
			t.Errorf("SimplifyPath() did not reduce points: %d -> %d", len(path.Points), len(simplified.Points))
		}
	})

	t.Run("PathLength", func(t *testing.T) {
		path := Path{
			Points: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
			},
		}

		length := PathLength(path)
		if length != 20 {
			t.Errorf("PathLength() = %d, want 20", length)
		}
	})

	t.Run("SplitPathAtMidpoint", func(t *testing.T) {
		path := Path{
			Points: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
			},
		}

		first, second, mid := SplitPathAtMidpoint(path)
		if len(first.Points) == 0 || len(second.Points) == 0 {
			t.Error("SplitPathAtMidpoint() produced empty path")
		}
		if mid.X == 0 && mid.Y == 0 {
			t.Error("SplitPathAtMidpoint() returned zero midpoint")
		}
	})
}

// TestTileType tests the TileType enumeration.
func TestTileType(t *testing.T) {
	tests := []struct {
		tileType TileType
		want     string
	}{
		{TileEmpty, "Empty"},
		{TileFloor, "Floor"},
		{TileWall, "Wall"},
		{TileDoor, "Door"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.tileType.String(); got != tt.want {
				t.Errorf("TileType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCountNeighbors tests neighbor counting functionality.
func TestCountNeighbors(t *testing.T) {
	t.Run("CountNeighbors cardinal", func(t *testing.T) {
		data := make([]uint32, 25)
		width, height := 5, 5

		// Set up a cross pattern
		_ = SetTile(data, 2, 2, width, height, 1)
		_ = SetTile(data, 1, 2, width, height, 1)
		_ = SetTile(data, 3, 2, width, height, 1)
		_ = SetTile(data, 2, 1, width, height, 1)
		_ = SetTile(data, 2, 3, width, height, 1)

		count := CountNeighbors(data, 2, 2, width, height, 1, false)
		if count != 4 {
			t.Errorf("CountNeighbors(cardinal) = %d, want 4", count)
		}
	})

	t.Run("CountNeighbors diagonal", func(t *testing.T) {
		data := make([]uint32, 25)
		width, height := 5, 5

		// Set up all 8 neighbors
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				_ = SetTile(data, 2+dx, 2+dy, width, height, 1)
			}
		}

		count := CountNeighbors(data, 2, 2, width, height, 1, true)
		if count != 8 {
			t.Errorf("CountNeighbors(diagonal) = %d, want 8", count)
		}
	})
}

// TestFloodFill tests flood fill functionality.
func TestFloodFill(t *testing.T) {
	t.Run("FloodFill simple", func(t *testing.T) {
		data := make([]uint32, 100)
		width, height := 10, 10

		// Create a rectangular region
		_ = FillRect(data, 2, 2, 5, 5, width, height, 1)

		// Fill the interior with a different value
		err := FloodFill(data, 3, 3, width, height, 2)
		if err != nil {
			t.Fatalf("FloodFill() error = %v", err)
		}

		// Check that interior was filled
		centerValue := GetTile(data, 3, 3, width, height)
		if centerValue != 2 {
			t.Errorf("FloodFill() center tile = %d, want 2", centerValue)
		}
	})

	t.Run("FloodFill bounds", func(t *testing.T) {
		data := make([]uint32, 100)
		width, height := 10, 10

		err := FloodFill(data, -1, -1, width, height, 1)
		if err == nil {
			t.Error("FloodFill() with out-of-bounds start should error, got nil")
		}
	})
}
