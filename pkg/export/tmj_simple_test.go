package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/carving"
	"github.com/dshills/dungo/pkg/dungeon"
)

// T094: Unit test for TMJ export structure validation
func TestTMJ_T094_ExportStructure(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *carving.TileMap
		wantErr bool
	}{
		{
			name: "basic tile map with floor and walls",
			setup: func() *carving.TileMap {
				tm := carving.NewTileMap(32, 24, 16, 16)
				floorLayer := carving.AddLayer(tm, "floor", "tilelayer")
				wallsLayer := carving.AddLayer(tm, "walls", "tilelayer")

				// Fill floor
				for i := range floorLayer.Data {
					floorLayer.Data[i] = 1
				}

				// Add walls around perimeter
				for y := 0; y < tm.Height; y++ {
					for x := 0; x < tm.Width; x++ {
						if x == 0 || x == tm.Width-1 || y == 0 || y == tm.Height-1 {
							wallsLayer.Data[y*tm.Width+x] = 2
						}
					}
				}

				return tm
			},
			wantErr: false,
		},
		{
			name: "tile map with all layer types",
			setup: func() *carving.TileMap {
				tm := carving.NewTileMap(16, 16, 16, 16)

				// Tile layers
				floor := carving.AddLayer(tm, "floor", "tilelayer")
				walls := carving.AddLayer(tm, "walls", "tilelayer")
				doors := carving.AddLayer(tm, "doors", "tilelayer")
				decor := carving.AddLayer(tm, "decor", "tilelayer")

				// Fill with test data
				for i := range floor.Data {
					floor.Data[i] = 1
					walls.Data[i] = 0
					doors.Data[i] = 0
					decor.Data[i] = 0
				}

				// Object layers
				entities := carving.AddLayer(tm, "entities", "objectgroup")
				triggers := carving.AddLayer(tm, "triggers", "objectgroup")

				// Add test objects
				entities.Objects = append(entities.Objects, carving.Object{
					ID:      1,
					Name:    "player_spawn",
					Type:    "spawn",
					X:       128.0,
					Y:       128.0,
					Width:   16.0,
					Height:  16.0,
					Visible: true,
					Properties: map[string]interface{}{
						"facing": "south",
					},
				})

				triggers.Objects = append(triggers.Objects, carving.Object{
					ID:      2,
					Name:    "entry_trigger",
					Type:    "trigger",
					X:       64.0,
					Y:       64.0,
					Width:   32.0,
					Height:  32.0,
					Visible: false,
					Properties: map[string]interface{}{
						"event": "show_tutorial",
					},
				})

				return tm
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := tt.setup()

			// Convert to TMJ
			tmjMap, err := ConvertTileMapToTMJ(tm, false)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ConvertTileMapToTMJ() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			// Validate structure
			if tmjMap.Type != "map" {
				t.Errorf("Type = %v, want 'map'", tmjMap.Type)
			}

			if tmjMap.Version != "1.10" {
				t.Errorf("Version = %v, want '1.10'", tmjMap.Version)
			}

			if tmjMap.Width != tm.Width {
				t.Errorf("Width = %v, want %v", tmjMap.Width, tm.Width)
			}

			if tmjMap.Height != tm.Height {
				t.Errorf("Height = %v, want %v", tmjMap.Height, tm.Height)
			}

			// Validate tilesets
			if len(tmjMap.Tilesets) == 0 {
				t.Error("Expected at least one tileset")
			}

			// Validate layers match input
			tileLayerCount := 0
			objectLayerCount := 0

			for _, layer := range tmjMap.Layers {
				if layer.Type == "tilelayer" {
					tileLayerCount++

					// Validate tile layer dimensions
					if layer.Width != tm.Width {
						t.Errorf("Layer %s: Width = %v, want %v", layer.Name, layer.Width, tm.Width)
					}
					if layer.Height != tm.Height {
						t.Errorf("Layer %s: Height = %v, want %v", layer.Name, layer.Height, tm.Height)
					}

					// Validate data
					if data, ok := layer.Data.([]uint32); ok {
						expectedLen := tm.Width * tm.Height
						if len(data) != expectedLen {
							t.Errorf("Layer %s: data length = %v, want %v", layer.Name, len(data), expectedLen)
						}
					}
				} else if layer.Type == "objectgroup" {
					objectLayerCount++

					// Validate objects have IDs
					for _, obj := range layer.Objects {
						if obj.ID <= 0 {
							t.Errorf("Layer %s: object %s has invalid ID %d", layer.Name, obj.Name, obj.ID)
						}
					}
				}
			}
		})
	}
}

// T096: Integration test for Tiled editor compatibility
func TestTMJ_T096_TiledEditorCompatibility(t *testing.T) {
	// Create a realistic dungeon tile map
	tm := carving.NewTileMap(32, 24, 16, 16)

	// Floor layer
	floor := carving.AddLayer(tm, "floor", "tilelayer")
	for i := range floor.Data {
		floor.Data[i] = 1 // Stone floor
	}

	// Walls layer
	walls := carving.AddLayer(tm, "walls", "tilelayer")
	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			idx := y*tm.Width + x
			if x == 0 || x == tm.Width-1 || y == 0 || y == tm.Height-1 {
				walls.Data[idx] = 2 // Wall tile
			}
		}
	}

	// Doors layer
	doors := carving.AddLayer(tm, "doors", "tilelayer")
	doors.Data[12*tm.Width+16] = 3 // Door at middle

	// Decor layer
	_ = carving.AddLayer(tm, "decor", "tilelayer")

	// Entities layer
	entities := carving.AddLayer(tm, "entities", "objectgroup")
	entities.Objects = append(entities.Objects,
		carving.Object{
			Name:    "player_spawn",
			Type:    "spawn_point",
			X:       256.0,
			Y:       192.0,
			Width:   16.0,
			Height:  16.0,
			Visible: true,
			Properties: map[string]interface{}{
				"facing": "south",
			},
		},
		carving.Object{
			Name:    "enemy_spawn",
			Type:    "spawn_point",
			X:       400.0,
			Y:       192.0,
			Width:   16.0,
			Height:  16.0,
			Visible: true,
			Properties: map[string]interface{}{
				"enemy_type": "goblin",
				"level":      5,
			},
		},
	)

	// Triggers layer
	triggers := carving.AddLayer(tm, "triggers", "objectgroup")
	triggers.Objects = append(triggers.Objects,
		carving.Object{
			Name:    "entry_trigger",
			Type:    "trigger",
			X:       240.0,
			Y:       180.0,
			Width:   32.0,
			Height:  32.0,
			Visible: false,
			Properties: map[string]interface{}{
				"event": "show_tutorial",
			},
		},
	)

	// Convert to TMJ
	tmjMap, err := ConvertTileMapToTMJ(tm, false)
	if err != nil {
		t.Fatalf("ConvertTileMapToTMJ() error = %v", err)
	}

	// Add metadata
	tmjMap.Properties = append(tmjMap.Properties,
		TMJProperty{Name: "difficulty", Type: "int", Value: 5},
		TMJProperty{Name: "theme", Type: "string", Value: "crypt"},
		TMJProperty{Name: "fogOfWar", Type: "bool", Value: true},
	)

	// Serialize to JSON
	data, err := MarshalTMJ(tmjMap)
	if err != nil {
		t.Fatalf("MarshalTMJ() error = %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "test_dungeon.tmj")

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Fatalf("TMJ file was not created: %v", err)
	}

	// Read back and validate JSON structure
	readData, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var decoded TMJMap
	if err := json.Unmarshal(readData, &decoded); err != nil {
		t.Fatalf("Failed to parse exported TMJ: %v", err)
	}

	// Validate Tiled-specific requirements

	// 1. Map must have type "map"
	if decoded.Type != "map" {
		t.Errorf("Type = %v, want 'map'", decoded.Type)
	}

	// 2. Must have version string
	if decoded.Version == "" {
		t.Error("Version is empty")
	}

	// 3. Must have tiledversion
	if decoded.TiledVersion == "" {
		t.Error("TiledVersion is empty")
	}

	// 4. Must have at least one tileset
	if len(decoded.Tilesets) == 0 {
		t.Error("No tilesets defined")
	}

	// 5. Tileset must have firstgid starting at 1
	if len(decoded.Tilesets) > 0 && decoded.Tilesets[0].FirstGID != 1 {
		t.Errorf("First tileset FirstGID = %v, want 1", decoded.Tilesets[0].FirstGID)
	}

	// 6. Layers must have unique IDs
	layerIDs := make(map[int]bool)
	for _, layer := range decoded.Layers {
		if layerIDs[layer.ID] {
			t.Errorf("Duplicate layer ID: %d", layer.ID)
		}
		layerIDs[layer.ID] = true
	}

	// 7. Objects must have unique IDs
	objectIDs := make(map[int]bool)
	for _, layer := range decoded.Layers {
		if layer.Type == "objectgroup" {
			for _, obj := range layer.Objects {
				if objectIDs[obj.ID] {
					t.Errorf("Duplicate object ID: %d", obj.ID)
				}
				objectIDs[obj.ID] = true
			}
		}
	}

	// 8. Tile layers must have correct data length
	for _, layer := range decoded.Layers {
		if layer.Type == "tilelayer" {
			if data, ok := layer.Data.([]interface{}); ok {
				expectedLen := decoded.Width * decoded.Height
				if len(data) != expectedLen {
					t.Errorf("Layer %s: data length = %v, want %v", layer.Name, len(data), expectedLen)
				}
			}
		}
	}

	// 9. Object properties must have name, type, and value
	for _, layer := range decoded.Layers {
		if layer.Type == "objectgroup" {
			for _, obj := range layer.Objects {
				for _, prop := range obj.Properties {
					if prop.Name == "" {
						t.Errorf("Object %s has property with empty name", obj.Name)
					}
					if prop.Type == "" {
						t.Errorf("Object %s property %s has empty type", obj.Name, prop.Name)
					}
				}
			}
		}
	}

	// Log success
	t.Logf("Successfully created Tiled-compatible TMJ file: %s", filepath)
	t.Logf("File size: %d bytes", len(readData))
	t.Logf("Layers: %d, Objects: %d", len(decoded.Layers), len(objectIDs))
}

// Test compression functionality
func TestTMJ_Compression(t *testing.T) {
	tm := carving.NewTileMap(32, 32, 16, 16)
	floor := carving.AddLayer(tm, "floor", "tilelayer")

	// Fill with pattern
	for i := range floor.Data {
		floor.Data[i] = uint32((i % 10) + 1)
	}

	// Without compression
	tmjUncompressed, err := ConvertTileMapToTMJ(tm, false)
	if err != nil {
		t.Fatalf("ConvertTileMapToTMJ(compress=false) error = %v", err)
	}

	uncompressedData, err := MarshalTMJ(tmjUncompressed)
	if err != nil {
		t.Fatalf("MarshalTMJ(uncompressed) error = %v", err)
	}

	// With compression
	tmjCompressed, err := ConvertTileMapToTMJ(tm, true)
	if err != nil {
		t.Fatalf("ConvertTileMapToTMJ(compress=true) error = %v", err)
	}

	compressedData, err := MarshalTMJ(tmjCompressed)
	if err != nil {
		t.Fatalf("MarshalTMJ(compressed) error = %v", err)
	}

	t.Logf("Uncompressed size: %d bytes", len(uncompressedData))
	t.Logf("Compressed size: %d bytes", len(compressedData))

	// Verify compression settings
	for _, layer := range tmjCompressed.Layers {
		if layer.Type == "tilelayer" {
			if layer.Encoding != "base64" {
				t.Errorf("Compressed layer encoding = %v, want 'base64'", layer.Encoding)
			}
			if layer.Compression != "gzip" {
				t.Errorf("Compressed layer compression = %v, want 'gzip'", layer.Compression)
			}

			// Data should be string, not []uint32
			if _, ok := layer.Data.(string); !ok {
				t.Errorf("Compressed layer data type = %T, want string", layer.Data)
			}
		}
	}
}

// Test artifact export
func TestTMJ_ArtifactExport(t *testing.T) {
	// Create a simple artifact with tile map
	tm := &dungeon.TileMap{
		Width:      16,
		Height:     16,
		TileWidth:  16,
		TileHeight: 16,
		Layers:     make(map[string]*dungeon.Layer),
	}
	artifact := &dungeon.Artifact{
		TileMap: tm,
	}

	// Add layers
	floorLayer := &dungeon.Layer{
		ID:      0,
		Name:    "floor",
		Type:    "tilelayer",
		Visible: true,
		Opacity: 1.0,
		Data:    make([]uint32, 16*16),
	}
	for i := range floorLayer.Data {
		floorLayer.Data[i] = 1
	}
	tm.Layers["floor"] = floorLayer

	wallsLayer := &dungeon.Layer{
		ID:      1,
		Name:    "walls",
		Type:    "tilelayer",
		Visible: true,
		Opacity: 1.0,
		Data:    make([]uint32, 16*16),
	}
	tm.Layers["walls"] = wallsLayer

	entitiesLayer := &dungeon.Layer{
		ID:      2,
		Name:    "entities",
		Type:    "objectgroup",
		Visible: true,
		Opacity: 1.0,
		Objects: []dungeon.Object{},
	}
	tm.Layers["entities"] = entitiesLayer

	// Test without compression
	tmjMap, err := ExportTMJ(artifact, false)
	if err != nil {
		t.Fatalf("ExportTMJ() error = %v", err)
	}

	data, err := MarshalTMJ(tmjMap)
	if err != nil {
		t.Fatalf("MarshalTMJ() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Exported data is empty")
	}

	// Verify it's valid JSON
	var decoded TMJMap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Exported data is not valid JSON: %v", err)
	}

	// Verify generator property
	foundGenerator := false
	for _, prop := range decoded.Properties {
		if prop.Name == "generator" && prop.Value == "dungo" {
			foundGenerator = true
			break
		}
	}
	if !foundGenerator {
		t.Error("Generator property not found in exported TMJ")
	}
}
