package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/carving"
)

// T096: Integration test for Tiled editor compatibility
func TestT096_TiledEditorCompatibility(t *testing.T) {
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
	tmjFilePath := filepath.Join(tmpDir, "test_dungeon.tmj")

	if err := os.WriteFile(tmjFilePath, data, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(tmjFilePath); os.IsNotExist(err) {
		t.Fatalf("TMJ file was not created: %v", err)
	}

	// Read back and validate JSON structure
	readData, err := os.ReadFile(tmjFilePath)
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
	t.Logf("Successfully created Tiled-compatible TMJ file: %s", tmjFilePath)
	t.Logf("File size: %d bytes", len(readData))
	t.Logf("Layers: %d, Objects: %d", len(decoded.Layers), len(objectIDs))
}
