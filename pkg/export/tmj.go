package export

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dshills/dungo/pkg/carving"
	"github.com/dshills/dungo/pkg/dungeon"
)

// TMJ Format Types
// Based on Tiled Map Editor JSON specification (TMJ 1.10)
// Reference: https://doc.mapeditor.org/en/stable/reference/json-map-format/

// TMJMap represents the root TMJ map structure.
type TMJMap struct {
	Type             string        `json:"type"`
	Version          string        `json:"version"`
	TiledVersion     string        `json:"tiledversion"`
	Width            int           `json:"width"`
	Height           int           `json:"height"`
	TileWidth        int           `json:"tilewidth"`
	TileHeight       int           `json:"tileheight"`
	Orientation      string        `json:"orientation"`
	RenderOrder      string        `json:"renderorder"`
	Infinite         bool          `json:"infinite"`
	NextLayerID      int           `json:"nextlayerid"`
	NextObjectID     int           `json:"nextobjectid"`
	BackgroundColor  *string       `json:"backgroundcolor,omitempty"`
	Class            string        `json:"class,omitempty"`
	CompressionLevel int           `json:"compressionlevel"`
	Layers           []TMJLayer    `json:"layers"`
	Tilesets         []TMJTileset  `json:"tilesets"`
	Properties       []TMJProperty `json:"properties,omitempty"`
}

// TMJLayer represents any layer type (tile, object, image, group).
type TMJLayer struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"` // "tilelayer" or "objectgroup"
	Visible    bool          `json:"visible"`
	Opacity    float64       `json:"opacity"`
	X          int           `json:"x"`
	Y          int           `json:"y"`
	Width      int           `json:"width,omitempty"`
	Height     int           `json:"height,omitempty"`
	OffsetX    int           `json:"offsetx,omitempty"`
	OffsetY    int           `json:"offsety,omitempty"`
	ParallaxX  float64       `json:"parallaxx,omitempty"`
	ParallaxY  float64       `json:"parallaxy,omitempty"`
	Class      string        `json:"class,omitempty"`
	Properties []TMJProperty `json:"properties,omitempty"`

	// Tile layer specific
	Data        interface{} `json:"data,omitempty"`        // []uint32 or string (base64)
	Encoding    string      `json:"encoding,omitempty"`    // "csv" or "base64"
	Compression string      `json:"compression,omitempty"` // "" or "gzip"

	// Object layer specific
	DrawOrder string      `json:"draworder,omitempty"`
	Objects   []TMJObject `json:"objects,omitempty"`
}

// TMJObject represents an entity or collision shape.
type TMJObject struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type,omitempty"`
	Class      string        `json:"class,omitempty"`
	X          float64       `json:"x"`
	Y          float64       `json:"y"`
	Width      float64       `json:"width"`
	Height     float64       `json:"height"`
	Rotation   float64       `json:"rotation"`
	GID        uint32        `json:"gid,omitempty"`
	Visible    bool          `json:"visible"`
	Properties []TMJProperty `json:"properties,omitempty"`

	// Shape types
	Ellipse  bool       `json:"ellipse,omitempty"`
	Point    bool       `json:"point,omitempty"`
	Polygon  []TMJPoint `json:"polygon,omitempty"`
	Polyline []TMJPoint `json:"polyline,omitempty"`
}

// TMJPoint represents a coordinate in a polygon or polyline.
type TMJPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// TMJTileset references a collection of tiles.
type TMJTileset struct {
	FirstGID         uint32        `json:"firstgid"`
	Source           string        `json:"source,omitempty"` // For external tilesets
	Name             string        `json:"name,omitempty"`
	Class            string        `json:"class,omitempty"`
	TileWidth        int           `json:"tilewidth,omitempty"`
	TileHeight       int           `json:"tileheight,omitempty"`
	Spacing          int           `json:"spacing,omitempty"`
	Margin           int           `json:"margin,omitempty"`
	TileCount        int           `json:"tilecount,omitempty"`
	Columns          int           `json:"columns,omitempty"`
	Image            string        `json:"image,omitempty"`
	ImageWidth       int           `json:"imagewidth,omitempty"`
	ImageHeight      int           `json:"imageheight,omitempty"`
	TransparentColor *string       `json:"transparentcolor,omitempty"`
	Properties       []TMJProperty `json:"properties,omitempty"`
}

// TMJProperty represents a custom property.
type TMJProperty struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// TMJ GID Flags
const (
	FlippedHorizontallyFlag = 0x80000000
	FlippedVerticallyFlag   = 0x40000000
	FlippedDiagonallyFlag   = 0x20000000
	TileIDMask              = 0x1FFFFFFF
)

// Builder Functions

// NewTMJMap creates a new TMJ map with default settings.
func NewTMJMap(width, height, tileWidth, tileHeight int) *TMJMap {
	return &TMJMap{
		Type:             "map",
		Version:          "1.10",
		TiledVersion:     "1.10.2",
		Width:            width,
		Height:           height,
		TileWidth:        tileWidth,
		TileHeight:       tileHeight,
		Orientation:      "orthogonal",
		RenderOrder:      "right-down",
		Infinite:         false,
		NextLayerID:      1,
		NextObjectID:     1,
		CompressionLevel: -1,
		Layers:           []TMJLayer{},
		Tilesets:         []TMJTileset{},
		Properties:       []TMJProperty{},
	}
}

// AddTileLayer adds a tile layer to the map.
func (m *TMJMap) AddTileLayer(name string, data []uint32) *TMJLayer {
	layer := TMJLayer{
		ID:       m.NextLayerID,
		Name:     name,
		Type:     "tilelayer",
		Visible:  true,
		Opacity:  1.0,
		X:        0,
		Y:        0,
		Width:    m.Width,
		Height:   m.Height,
		Data:     data,
		Encoding: "csv",
	}
	m.NextLayerID++
	m.Layers = append(m.Layers, layer)
	return &m.Layers[len(m.Layers)-1]
}

// AddObjectLayer adds an object layer to the map.
func (m *TMJMap) AddObjectLayer(name string) *TMJLayer {
	layer := TMJLayer{
		ID:        m.NextLayerID,
		Name:      name,
		Type:      "objectgroup",
		Visible:   true,
		Opacity:   1.0,
		DrawOrder: "topdown",
		Objects:   []TMJObject{},
	}
	m.NextLayerID++
	m.Layers = append(m.Layers, layer)
	return &m.Layers[len(m.Layers)-1]
}

// AddObject adds an object to an object layer.
func (l *TMJLayer) AddObject(obj TMJObject, m *TMJMap) {
	if l.Type != "objectgroup" {
		return
	}
	obj.ID = m.NextObjectID
	m.NextObjectID++
	l.Objects = append(l.Objects, obj)
}

// AddTileset adds a tileset reference to the map.
func (m *TMJMap) AddTileset(name, imagePath string, tileWidth, tileHeight, tileCount, columns int) *TMJTileset {
	firstGID := uint32(1)
	if len(m.Tilesets) > 0 {
		last := m.Tilesets[len(m.Tilesets)-1]
		firstGID = last.FirstGID + uint32(last.TileCount)
	}

	imageWidth := columns * tileWidth
	imageHeight := (tileCount / columns) * tileHeight
	if tileCount%columns != 0 {
		imageHeight += tileHeight
	}

	tileset := TMJTileset{
		FirstGID:    firstGID,
		Name:        name,
		TileWidth:   tileWidth,
		TileHeight:  tileHeight,
		TileCount:   tileCount,
		Columns:     columns,
		Image:       imagePath,
		ImageWidth:  imageWidth,
		ImageHeight: imageHeight,
		Spacing:     0,
		Margin:      0,
	}
	m.Tilesets = append(m.Tilesets, tileset)
	return &m.Tilesets[len(m.Tilesets)-1]
}

// Compression Support

// CompressLayerData compresses tile data with gzip and encodes as base64.
func (l *TMJLayer) CompressLayerData() error {
	if l.Type != "tilelayer" {
		return fmt.Errorf("cannot compress non-tile layer")
	}

	data, ok := l.Data.([]uint32)
	if !ok {
		return fmt.Errorf("layer data is not []uint32")
	}

	// Convert uint32 slice to byte slice (little-endian)
	buf := new(bytes.Buffer)
	for _, gid := range data {
		buf.WriteByte(byte(gid))
		buf.WriteByte(byte(gid >> 8))
		buf.WriteByte(byte(gid >> 16))
		buf.WriteByte(byte(gid >> 24))
	}

	// Compress with gzip
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	if _, err := gzipWriter.Write(buf.Bytes()); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(compressed.Bytes())

	// Update layer
	l.Data = encoded
	l.Encoding = "base64"
	l.Compression = "gzip"

	return nil
}

// GID Utility Functions

// CalculateGID converts tileset local ID to global ID with flip flags.
func CalculateGID(tilesetFirstGID uint32, localTileID int, flipH, flipV, flipD bool) uint32 {
	gid := tilesetFirstGID + uint32(localTileID)

	if flipH {
		gid |= FlippedHorizontallyFlag
	}
	if flipV {
		gid |= FlippedVerticallyFlag
	}
	if flipD {
		gid |= FlippedDiagonallyFlag
	}

	return gid
}

// ParseGID extracts tile ID and flip flags from GID.
func ParseGID(gid uint32) (tileID uint32, flipH, flipV, flipD bool) {
	flipH = (gid & FlippedHorizontallyFlag) != 0
	flipV = (gid & FlippedVerticallyFlag) != 0
	flipD = (gid & FlippedDiagonallyFlag) != 0
	tileID = gid & TileIDMask
	return
}

// Export Functions

// ExportTMJ converts a dungeon artifact to TMJ format.
func ExportTMJ(artifact *dungeon.Artifact, compress bool) (*TMJMap, error) {
	if artifact.TileMap == nil {
		return nil, fmt.Errorf("artifact has no tile map")
	}

	tm := artifact.TileMap

	// Create TMJ map
	tmjMap := NewTMJMap(tm.Width, tm.Height, tm.TileWidth, tm.TileHeight)
	tmjMap.Class = "dungeon"

	// Add default tileset
	tmjMap.AddTileset("dungeon_tiles", "tilesets/dungeon.png", 16, 16, 256, 16)

	// Export tile layers (floor, walls, doors, decor)
	layerNames := []string{"floor", "walls", "doors", "decor"}
	for _, name := range layerNames {
		if layer, exists := tm.Layers[name]; exists && layer.Type == "tilelayer" {
			tmjLayer := tmjMap.AddTileLayer(name, layer.Data)
			tmjLayer.Class = name

			// Apply compression if requested
			if compress {
				if err := tmjLayer.CompressLayerData(); err != nil {
					return nil, fmt.Errorf("failed to compress layer %s: %w", name, err)
				}
			}
		}
	}

	// Export object layers (entities, triggers)
	objectLayerNames := []string{"entities", "triggers", "hazards"}
	for _, name := range objectLayerNames {
		if layer, exists := tm.Layers[name]; exists && layer.Type == "objectgroup" {
			tmjLayer := tmjMap.AddObjectLayer(name)
			tmjLayer.Class = name

			// Convert objects
			for _, obj := range layer.Objects {
				tmjObj := TMJObject{
					Name:     obj.Name,
					Type:     obj.Type,
					Class:    obj.Type,
					X:        obj.X,
					Y:        obj.Y,
					Width:    obj.Width,
					Height:   obj.Height,
					Rotation: obj.Rotation,
					GID:      obj.GID,
					Visible:  obj.Visible,
				}

				// Convert properties
				if obj.Properties != nil {
					tmjObj.Properties = make([]TMJProperty, 0, len(obj.Properties))
					for key, value := range obj.Properties {
						prop := TMJProperty{
							Name:  key,
							Value: value,
						}
						// Infer type from value
						switch value.(type) {
						case bool:
							prop.Type = "bool"
						case int, int32, int64:
							prop.Type = "int"
						case float32, float64:
							prop.Type = "float"
						default:
							prop.Type = "string"
						}
						tmjObj.Properties = append(tmjObj.Properties, prop)
					}
				}

				tmjLayer.AddObject(tmjObj, tmjMap)
			}
		}
	}

	// Add metadata properties
	tmjMap.Properties = append(tmjMap.Properties,
		TMJProperty{Name: "generator", Type: "string", Value: "dungo"},
	)

	return tmjMap, nil
}

// ExportTMJFromCarving converts a carving.TileMap to TMJ format (helper for testing).
func ExportTMJFromCarving(tm *carving.TileMap, compress bool) (*TMJMap, error) {
	return ConvertTileMapToTMJ(tm, compress)
}

// MarshalTMJ serializes a TMJ map to JSON with indentation.
func MarshalTMJ(tmjMap *TMJMap) ([]byte, error) {
	return json.MarshalIndent(tmjMap, "", "  ")
}

// MarshalTMJCompact serializes a TMJ map to compact JSON.
func MarshalTMJCompact(tmjMap *TMJMap) ([]byte, error) {
	return json.Marshal(tmjMap)
}

// SaveTMJToFile exports a TMJ map to a file.
func SaveTMJToFile(tmjMap *TMJMap, filepath string) error {
	data, err := MarshalTMJ(tmjMap)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// EncodeTMJ writes a TMJ map to a writer with indentation.
func EncodeTMJ(tmjMap *TMJMap, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tmjMap)
}

// Convenience Functions

// ExportArtifactToTMJ exports an artifact to TMJ format with options.
func ExportArtifactToTMJ(artifact *dungeon.Artifact, compress bool) ([]byte, error) {
	tmjMap, err := ExportTMJ(artifact, compress)
	if err != nil {
		return nil, err
	}
	return MarshalTMJ(tmjMap)
}

// SaveArtifactToTMJFile exports an artifact directly to a TMJ file.
func SaveArtifactToTMJFile(artifact *dungeon.Artifact, filepath string, compress bool) error {
	tmjMap, err := ExportTMJ(artifact, compress)
	if err != nil {
		return err
	}
	return SaveTMJToFile(tmjMap, filepath)
}

// ConvertTileMapToTMJ converts a carving.TileMap to TMJ format (used for testing).
func ConvertTileMapToTMJ(tm *carving.TileMap, compress bool) (*TMJMap, error) {
	if tm == nil {
		return nil, fmt.Errorf("tile map is nil")
	}

	// Create TMJ map
	tmjMap := NewTMJMap(tm.Width, tm.Height, tm.TileWidth, tm.TileHeight)
	tmjMap.Class = "dungeon"

	// Add default tileset
	tmjMap.AddTileset("dungeon_tiles", "tilesets/dungeon.png", 16, 16, 256, 16)

	// Export all tile layers
	for name, layer := range tm.Layers {
		if layer.Type == "tilelayer" {
			tmjLayer := tmjMap.AddTileLayer(name, layer.Data)
			tmjLayer.Class = name

			if compress {
				if err := tmjLayer.CompressLayerData(); err != nil {
					return nil, fmt.Errorf("failed to compress layer %s: %w", name, err)
				}
			}
		} else if layer.Type == "objectgroup" {
			tmjLayer := tmjMap.AddObjectLayer(name)
			tmjLayer.Class = name

			for _, obj := range layer.Objects {
				tmjObj := TMJObject{
					Name:     obj.Name,
					Type:     obj.Type,
					Class:    obj.Type,
					X:        obj.X,
					Y:        obj.Y,
					Width:    obj.Width,
					Height:   obj.Height,
					Rotation: obj.Rotation,
					GID:      obj.GID,
					Visible:  obj.Visible,
				}

				if obj.Properties != nil {
					tmjObj.Properties = make([]TMJProperty, 0, len(obj.Properties))
					for key, value := range obj.Properties {
						prop := TMJProperty{Name: key, Value: value}
						switch value.(type) {
						case bool:
							prop.Type = "bool"
						case int, int32, int64:
							prop.Type = "int"
						case float32, float64:
							prop.Type = "float"
						default:
							prop.Type = "string"
						}
						tmjObj.Properties = append(tmjObj.Properties, prop)
					}
				}

				tmjLayer.AddObject(tmjObj, tmjMap)
			}
		}
	}

	return tmjMap, nil
}
