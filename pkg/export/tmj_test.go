package export_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/export"
)

// TMJMap represents the structure of a Tiled TMJ (JSON map) file.
// This is a subset of the full Tiled format focusing on key fields we need to validate.
type TMJMap struct {
	Version         string       `json:"version"`
	TiledVersion    string       `json:"tiledversion"`
	Type            string       `json:"type"`
	Orientation     string       `json:"orientation"`
	Width           int          `json:"width"`
	Height          int          `json:"height"`
	TileWidth       int          `json:"tilewidth"`
	TileHeight      int          `json:"tileheight"`
	Infinite        bool         `json:"infinite"`
	Layers          []TMJLayer   `json:"layers"`
	Tilesets        []TMJTileset `json:"tilesets"`
	NextObjectID    int          `json:"nextobjectid,omitempty"`
	NextLayerID     int          `json:"nextlayerid,omitempty"`
	RenderOrder     string       `json:"renderorder,omitempty"`
	BackgroundColor string       `json:"backgroundcolor,omitempty"`
}

// TMJLayer represents a layer in the TMJ format.
type TMJLayer struct {
	ID       int         `json:"id"`
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "tilelayer" or "objectgroup"
	Visible  bool        `json:"visible"`
	Opacity  float64     `json:"opacity"`
	X        int         `json:"x"`
	Y        int         `json:"y"`
	Width    int         `json:"width,omitempty"`
	Height   int         `json:"height,omitempty"`
	Data     []uint32    `json:"data,omitempty"`     // For tile layers
	Encoding string      `json:"encoding,omitempty"` // "csv" or omitted for JSON
	Objects  []TMJObject `json:"objects,omitempty"`  // For object layers
}

// TMJObject represents an object in an object layer.
type TMJObject struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	X          float64       `json:"x"`
	Y          float64       `json:"y"`
	Width      float64       `json:"width"`
	Height     float64       `json:"height"`
	Rotation   float64       `json:"rotation"`
	GID        uint32        `json:"gid,omitempty"`
	Visible    bool          `json:"visible"`
	Properties []TMJProperty `json:"properties,omitempty"`
}

// TMJProperty represents a custom property in Tiled.
type TMJProperty struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// TMJTileset represents a tileset reference in the TMJ format.
type TMJTileset struct {
	FirstGID   int    `json:"firstgid"`
	Source     string `json:"source,omitempty"` // External tileset file
	Name       string `json:"name,omitempty"`   // Embedded tileset
	TileWidth  int    `json:"tilewidth,omitempty"`
	TileHeight int    `json:"tileheight,omitempty"`
	TileCount  int    `json:"tilecount,omitempty"`
	Columns    int    `json:"columns,omitempty"`
}

// TestTMJStructureValidation validates that TMJ export produces valid Tiled map structure.
// This is the RED PHASE - test will fail until TMJ export is implemented.
func TestTMJStructureValidation(t *testing.T) {
	// Generate a small test dungeon
	cfg := &dungeon.Config{
		Seed: 55555,
		Size: dungeon.SizeCfg{
			RoomsMin: 4,
			RoomsMax: 6,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 1.4,
			Max: 2,
		},
		SecretDensity: 0.15,
		OptionalRatio: 0.25,
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.1,
		},
		Themes: []string{"dungeon"},
	}

	gen := dungeon.NewGenerator()
	artifact, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping TMJ test: generation failed: %v", err)
	}

	// Export to TMJ (this will fail in RED phase)
	tmj, err := export.ExportTMJ(artifact, false)
	if err != nil {
		t.Fatalf("ExportTMJ failed (expected in RED phase): %v", err)
	}

	// Serialize to JSON for validation
	tmjData, err := json.Marshal(tmj)
	if err != nil {
		t.Fatalf("Failed to marshal TMJ to JSON: %v", err)
	}

	// Parse back to verify
	var tmjParsed TMJMap
	if err := json.Unmarshal(tmjData, &tmjParsed); err != nil {
		t.Fatalf("Failed to parse TMJ JSON: %v", err)
	}

	// Validate TMJ structure
	t.Run("MapMetadata", func(t *testing.T) {
		if tmjParsed.Type != "map" {
			t.Errorf("Type should be 'map', got %q", tmjParsed.Type)
		}
		if tmjParsed.Orientation != "orthogonal" {
			t.Errorf("Orientation should be 'orthogonal', got %q", tmjParsed.Orientation)
		}
		if tmjParsed.Width <= 0 {
			t.Errorf("Width should be positive, got %d", tmjParsed.Width)
		}
		if tmjParsed.Height <= 0 {
			t.Errorf("Height should be positive, got %d", tmjParsed.Height)
		}
		if tmjParsed.TileWidth <= 0 {
			t.Errorf("TileWidth should be positive, got %d", tmjParsed.TileWidth)
		}
		if tmjParsed.TileHeight <= 0 {
			t.Errorf("TileHeight should be positive, got %d", tmjParsed.TileHeight)
		}
		if tmjParsed.Infinite {
			t.Error("Infinite should be false for dungeon maps")
		}
	})

	t.Run("Tilesets", func(t *testing.T) {
		if len(tmjParsed.Tilesets) == 0 {
			t.Error("TMJ should include at least one tileset")
		}
		for i, tileset := range tmjParsed.Tilesets {
			if tileset.FirstGID <= 0 {
				t.Errorf("Tileset %d: FirstGID should be positive, got %d", i, tileset.FirstGID)
			}
			// Either Source (external) or Name (embedded) should be present
			if tileset.Source == "" && tileset.Name == "" {
				t.Errorf("Tileset %d: should have either Source or Name", i)
			}
		}
	})

	t.Run("Layers", func(t *testing.T) {
		if len(tmjParsed.Layers) == 0 {
			t.Fatal("TMJ should have at least one layer")
		}

		// Expected layers for a dungeon
		expectedLayers := map[string]string{
			"floor":    "tilelayer",
			"walls":    "tilelayer",
			"doors":    "tilelayer",
			"entities": "objectgroup",
		}

		foundLayers := make(map[string]bool)
		for _, layer := range tmjParsed.Layers {
			// Validate common fields
			if layer.Name == "" {
				t.Error("Layer should have a name")
			}
			if layer.Type != "tilelayer" && layer.Type != "objectgroup" {
				t.Errorf("Layer %q: invalid type %q", layer.Name, layer.Type)
			}

			// Check if this is an expected layer
			if expectedType, ok := expectedLayers[layer.Name]; ok {
				foundLayers[layer.Name] = true
				if layer.Type != expectedType {
					t.Errorf("Layer %q: expected type %q, got %q",
						layer.Name, expectedType, layer.Type)
				}
			}

			// Validate tile layer specific fields
			if layer.Type == "tilelayer" {
				if layer.Width <= 0 || layer.Height <= 0 {
					t.Errorf("Tile layer %q: dimensions should be positive (w=%d, h=%d)",
						layer.Name, layer.Width, layer.Height)
				}
				expectedDataSize := layer.Width * layer.Height
				if len(layer.Data) != expectedDataSize {
					t.Errorf("Tile layer %q: data size mismatch (got %d, want %d)",
						layer.Name, len(layer.Data), expectedDataSize)
				}
			}

			// Validate object layer specific fields
			if layer.Type == "objectgroup" {
				// Objects array should exist (may be empty)
				if layer.Objects == nil {
					t.Errorf("Object layer %q: Objects should not be nil", layer.Name)
				}
			}
		}

		// Check that we found critical layers
		if !foundLayers["floor"] {
			t.Error("Missing required 'floor' layer")
		}
		if !foundLayers["walls"] {
			t.Error("Missing required 'walls' layer")
		}
	})

	t.Run("EntityObjects", func(t *testing.T) {
		// Find entity layer
		var entityLayer *TMJLayer
		for i := range tmjParsed.Layers {
			if tmjParsed.Layers[i].Type == "objectgroup" {
				entityLayer = &tmjParsed.Layers[i]
				break
			}
		}

		if entityLayer == nil {
			t.Fatal("No object layer found")
		}

		// Validate objects in entity layer
		for _, obj := range entityLayer.Objects {
			if obj.ID <= 0 {
				t.Errorf("Object %q: ID should be positive, got %d", obj.Name, obj.ID)
			}
			if obj.Type == "" {
				t.Errorf("Object %q: Type should not be empty", obj.Name)
			}
			// Objects should have valid positions (can be 0, but not negative)
			if obj.X < 0 || obj.Y < 0 {
				t.Errorf("Object %q: Position should be non-negative (x=%f, y=%f)",
					obj.Name, obj.X, obj.Y)
			}
		}
	})

	t.Log("TMJ structure validation passed")
}

// TestTiledCompatibility verifies that TMJ output can be parsed by standard JSON parser
// and has valid Tiled structure.
// This is an integration test for Tiled compatibility (RED PHASE).
func TestTiledCompatibility(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 77777,
		Size: dungeon.SizeCfg{
			RoomsMin: 3,
			RoomsMax: 5,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 1.3,
			Max: 2,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.1,
		},
		Themes: []string{"cave"},
	}

	gen := dungeon.NewGenerator()
	artifact, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping Tiled compatibility test: generation failed: %v", err)
	}

	// Export to TMJ
	tmjMap, err := export.ExportTMJ(artifact, false)
	if err != nil {
		t.Fatalf("ExportTMJ failed: %v", err)
	}

	// Serialize to JSON
	tmjData, err := json.Marshal(tmjMap)
	if err != nil {
		t.Fatalf("Failed to marshal TMJ: %v", err)
	}

	// Test 1: Valid JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal(tmjData, &jsonCheck); err != nil {
		t.Fatalf("TMJ is not valid JSON: %v", err)
	}

	// Test 2: Has required Tiled fields
	requiredFields := []string{"type", "version", "orientation", "width", "height",
		"tilewidth", "tileheight", "layers"}
	for _, field := range requiredFields {
		if _, ok := jsonCheck[field]; !ok {
			t.Errorf("Missing required Tiled field: %q", field)
		}
	}

	// Test 3: Parse as TMJMap structure
	var tmj TMJMap
	if err := json.Unmarshal(tmjData, &tmj); err != nil {
		t.Fatalf("Failed to parse as TMJMap: %v", err)
	}

	// Test 4: Verify version compatibility
	// Tiled uses semantic versioning, we should support at least version 1.0
	if tmj.Version == "" {
		t.Error("Version field is empty")
	}

	// Test 5: Verify data integrity
	// All tile layers should have data that matches dimensions
	for _, layer := range tmj.Layers {
		if layer.Type == "tilelayer" {
			expectedSize := layer.Width * layer.Height
			if len(layer.Data) != expectedSize {
				t.Errorf("Layer %q: data integrity check failed (size=%d, expected=%d)",
					layer.Name, len(layer.Data), expectedSize)
			}
		}
	}

	// Test 6: Verify object layer structure
	hasObjectLayer := false
	for _, layer := range tmj.Layers {
		if layer.Type == "objectgroup" {
			hasObjectLayer = true
			// Object layer should have Objects field
			if layer.Objects == nil {
				t.Errorf("Object layer %q: Objects field is nil", layer.Name)
			}
		}
	}
	if !hasObjectLayer {
		t.Error("TMJ should have at least one object layer for entities")
	}

	t.Log("Tiled compatibility test passed")
}

// TestTMJLayerStructure validates specific layer requirements.
func TestTMJLayerStructure(t *testing.T) {
	cfg := &dungeon.Config{
		Seed: 88888,
		Size: dungeon.SizeCfg{
			RoomsMin: 5,
			RoomsMax: 7,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 1.5,
			Max: 3,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.15,
		},
		Themes: []string{"fortress"},
	}

	gen := dungeon.NewGenerator()
	artifact, err := gen.Generate(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping layer structure test: generation failed: %v", err)
	}

	tmjMap, err := export.ExportTMJ(artifact, false)
	if err != nil {
		t.Fatalf("ExportTMJ failed: %v", err)
	}

	// Serialize and parse for validation
	tmjData, err := json.Marshal(tmjMap)
	if err != nil {
		t.Fatalf("Failed to marshal TMJ: %v", err)
	}

	var tmj TMJMap
	if err := json.Unmarshal(tmjData, &tmj); err != nil {
		t.Fatalf("Failed to parse TMJ: %v", err)
	}

	// Check layer ordering: floor < walls < doors < entities
	// This ensures proper rendering order in Tiled
	layerOrder := make(map[string]int)
	for i, layer := range tmj.Layers {
		layerOrder[layer.Name] = i
	}

	if floorIdx, ok := layerOrder["floor"]; ok {
		if wallIdx, ok := layerOrder["walls"]; ok {
			if floorIdx >= wallIdx {
				t.Error("Floor layer should come before walls layer")
			}
		}
		if doorIdx, ok := layerOrder["doors"]; ok {
			if floorIdx >= doorIdx {
				t.Error("Floor layer should come before doors layer")
			}
		}
	}

	// Check that tile layers have consistent dimensions
	var mapWidth, mapHeight int
	for _, layer := range tmj.Layers {
		if layer.Type == "tilelayer" {
			if mapWidth == 0 {
				mapWidth = layer.Width
				mapHeight = layer.Height
			} else {
				if layer.Width != mapWidth || layer.Height != mapHeight {
					t.Errorf("Layer %q: dimensions mismatch (w=%d, h=%d, expected w=%d, h=%d)",
						layer.Name, layer.Width, layer.Height, mapWidth, mapHeight)
				}
			}
		}
	}

	t.Log("TMJ layer structure test passed")
}
