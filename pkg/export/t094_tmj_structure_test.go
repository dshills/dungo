package export

import (
	"encoding/json"
	"testing"

	"github.com/dshills/dungo/pkg/carving"
)

// T094: Unit test for TMJ export structure validation
func TestT094_TMJExportStructure(t *testing.T) {
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
		{
			name: "empty tile map",
			setup: func() *carving.TileMap {
				return carving.NewTileMap(8, 8, 16, 16)
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

			if tmjMap.Orientation != "orthogonal" {
				t.Errorf("Orientation = %v, want 'orthogonal'", tmjMap.Orientation)
			}

			// Validate tilesets
			if len(tmjMap.Tilesets) == 0 {
				t.Error("Expected at least one tileset")
			}

			if len(tmjMap.Tilesets) > 0 {
				tileset := tmjMap.Tilesets[0]
				if tileset.FirstGID != 1 {
					t.Errorf("FirstGID = %v, want 1", tileset.FirstGID)
				}
				if tileset.TileWidth != 16 {
					t.Errorf("TileWidth = %v, want 16", tileset.TileWidth)
				}
				if tileset.TileHeight != 16 {
					t.Errorf("TileHeight = %v, want 16", tileset.TileHeight)
				}
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

			// Verify layer counts
			expectedTileLayers := 0
			expectedObjectLayers := 0
			for _, layer := range tm.Layers {
				if layer.Type == "tilelayer" {
					expectedTileLayers++
				} else if layer.Type == "objectgroup" {
					expectedObjectLayers++
				}
			}

			if tileLayerCount != expectedTileLayers {
				t.Errorf("Tile layer count = %v, want %v", tileLayerCount, expectedTileLayers)
			}
			if objectLayerCount != expectedObjectLayers {
				t.Errorf("Object layer count = %v, want %v", objectLayerCount, expectedObjectLayers)
			}
		})
	}
}

// T094: Test JSON serialization round-trip
func TestT094_TMJSerializationRoundTrip(t *testing.T) {
	// Create a test TMJ map
	tmjMap := NewTMJMap(16, 16, 16, 16)
	tmjMap.AddTileset("test_tiles", "test.png", 16, 16, 64, 8)

	floorData := make([]uint32, 16*16)
	for i := range floorData {
		floorData[i] = 1
	}
	tmjMap.AddTileLayer("floor", floorData)

	entityLayer := tmjMap.AddObjectLayer("entities")
	entityLayer.AddObject(TMJObject{
		Name:    "test_object",
		Type:    "test",
		X:       64.0,
		Y:       64.0,
		Width:   16.0,
		Height:  16.0,
		Visible: true,
	}, tmjMap)

	// Serialize
	data, err := MarshalTMJ(tmjMap)
	if err != nil {
		t.Fatalf("MarshalTMJ() error = %v", err)
	}

	// Deserialize
	var decoded TMJMap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify key fields
	if decoded.Width != tmjMap.Width {
		t.Errorf("Width = %v, want %v", decoded.Width, tmjMap.Width)
	}
	if decoded.Height != tmjMap.Height {
		t.Errorf("Height = %v, want %v", decoded.Height, tmjMap.Height)
	}
	if len(decoded.Layers) != len(tmjMap.Layers) {
		t.Errorf("Layer count = %v, want %v", len(decoded.Layers), len(tmjMap.Layers))
	}
	if len(decoded.Tilesets) != len(tmjMap.Tilesets) {
		t.Errorf("Tileset count = %v, want %v", len(decoded.Tilesets), len(tmjMap.Tilesets))
	}
}

// T094: Test compression
func TestT094_TMJCompression(t *testing.T) {
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
