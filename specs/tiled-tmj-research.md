# Tiled TMJ (JSON) Format Research

## Executive Summary

This document provides comprehensive research on the Tiled TMJ (JSON) format for tile map export, with specific focus on Go implementation approaches. TMJ is the JSON-based map format used by the Tiled Map Editor, supporting all features including multiple layers, objects, tilesets, and advanced properties.

---

## 1. Overview of TMJ Format Structure

### 1.1 Format Background

The **TMJ format** is Tiled's native JSON export format, providing a more accessible alternative to the XML-based TMX format. It supports all Tiled features and can be opened directly in the Tiled editor. The format uses the `.tmj` extension for maps and `.tsj` for tilesets.

**Official Specification:** https://doc.mapeditor.org/en/stable/reference/json-map-format/

### 1.2 Root Map Structure

The root JSON object represents the entire map with the following core properties:

```json
{
  "version": "1.10",
  "tiledversion": "1.10.2",
  "width": 32,
  "height": 24,
  "tilewidth": 16,
  "tileheight": 16,
  "orientation": "orthogonal",
  "renderorder": "right-down",
  "infinite": false,
  "nextlayerid": 10,
  "nextobjectid": 50,
  "backgroundcolor": "#000000",
  "class": "dungeon",
  "parallaxoriginx": 0,
  "parallaxoriginy": 0,
  "compressionlevel": -1,
  "properties": [],
  "layers": [],
  "tilesets": []
}
```

**Key Properties:**
- `width`, `height`: Map dimensions in tiles
- `tilewidth`, `tileheight`: Individual tile size in pixels
- `orientation`: `"orthogonal"`, `"isometric"`, `"staggered"`, or `"hexagonal"`
- `renderorder`: Rendering direction (`"right-down"`, `"right-up"`, `"left-down"`, `"left-up"`)
- `infinite`: Boolean indicating infinite map support
- `nextlayerid`, `nextobjectid`: ID counters for new elements
- `version`: TMJ format version (string since 1.6)
- `tiledversion`: Version of Tiled that created the map

### 1.3 Layer Types

Tiled supports four primary layer types, each serving different purposes:

#### Tile Layer (tilelayer)

The most common layer type for grid-based tiles:

```json
{
  "id": 1,
  "name": "floor",
  "type": "tilelayer",
  "visible": true,
  "opacity": 1.0,
  "x": 0,
  "y": 0,
  "width": 32,
  "height": 24,
  "parallaxx": 1.0,
  "parallaxy": 1.0,
  "offsetx": 0,
  "offsety": 0,
  "class": "",
  "properties": [],
  "encoding": "csv",
  "compression": "",
  "data": [1, 2, 1, 2, 3, 1, 3, 1, ...]
}
```

**Data Encoding Options:**
1. **CSV (default)**: Plain array of tile GIDs
2. **Base64**: Base64-encoded binary data
3. **Base64 + Compression**: Base64-encoded with zlib, gzip, or zstd compression

**Tile GIDs (Global IDs):**
- Bits 0-28: Tile ID
- Bit 29: Horizontal flip flag
- Bit 30: Vertical flip flag
- Bit 31: Diagonal flip flag
- Value 0: Empty tile

**Infinite Maps:**
For infinite maps, tile data is organized into chunks:

```json
{
  "chunks": [
    {
      "x": 0,
      "y": 0,
      "width": 16,
      "height": 16,
      "data": [...]
    }
  ]
}
```

#### Object Layer (objectgroup)

Contains interactive elements, collision shapes, and entity spawn points:

```json
{
  "id": 2,
  "name": "entities",
  "type": "objectgroup",
  "visible": true,
  "opacity": 1.0,
  "x": 0,
  "y": 0,
  "offsetx": 0,
  "offsety": 0,
  "parallaxx": 1.0,
  "parallaxy": 1.0,
  "class": "",
  "draworder": "topdown",
  "properties": [],
  "objects": [
    {
      "id": 1,
      "name": "player_spawn",
      "type": "spawn_point",
      "class": "PlayerSpawn",
      "x": 128,
      "y": 96,
      "width": 16,
      "height": 16,
      "rotation": 0,
      "visible": true,
      "properties": [
        {
          "name": "facing",
          "type": "string",
          "value": "south"
        }
      ]
    }
  ]
}
```

**Object Shape Types:**
1. **Rectangle** (default): Has `x`, `y`, `width`, `height`
2. **Ellipse**: Adds `"ellipse": true`
3. **Point**: Adds `"point": true`, width/height = 0
4. **Polygon**: Adds `"polygon": [{"x": 0, "y": 0}, ...]` (closed shape)
5. **Polyline**: Adds `"polyline": [{"x": 0, "y": 0}, ...]` (open shape)
6. **Tile Object**: Has `"gid"` referencing a tile
7. **Text**: Contains `"text"` object with formatting

**Text Objects:**
```json
{
  "id": 10,
  "name": "sign",
  "x": 200,
  "y": 100,
  "width": 150,
  "height": 50,
  "text": {
    "text": "Welcome to the dungeon!",
    "fontfamily": "sans-serif",
    "pixelsize": 16,
    "wrap": true,
    "color": "#ffffff",
    "bold": false,
    "italic": false,
    "underline": false,
    "strikeout": false,
    "kerning": true,
    "halign": "center",
    "valign": "top"
  }
}
```

#### Image Layer (imagelayer)

Displays static background or foreground images:

```json
{
  "id": 3,
  "name": "background",
  "type": "imagelayer",
  "visible": true,
  "opacity": 0.5,
  "offsetx": 0,
  "offsety": 0,
  "parallaxx": 0.5,
  "parallaxy": 0.5,
  "image": "backgrounds/cave.png",
  "transparentcolor": "#ff00ff",
  "repeatx": false,
  "repeaty": false,
  "properties": []
}
```

#### Group Layer (group)

Hierarchical organization with nested layers:

```json
{
  "id": 4,
  "name": "environment",
  "type": "group",
  "visible": true,
  "opacity": 1.0,
  "offsetx": 0,
  "offsety": 0,
  "parallaxx": 1.0,
  "parallaxy": 1.0,
  "properties": [],
  "layers": [
    {
      "id": 5,
      "name": "floor",
      "type": "tilelayer",
      ...
    },
    {
      "id": 6,
      "name": "walls",
      "type": "tilelayer",
      ...
    }
  ]
}
```

### 1.4 Tileset Structure

Tilesets define the available tile graphics and their properties:

```json
{
  "firstgid": 1,
  "source": "tilesets/dungeon.tsj",
  "name": "dungeon_tiles",
  "class": "",
  "tilewidth": 16,
  "tileheight": 16,
  "spacing": 0,
  "margin": 0,
  "tilecount": 256,
  "columns": 16,
  "image": "tilesets/dungeon.png",
  "imagewidth": 256,
  "imageheight": 256,
  "transparentcolor": "#ff00ff",
  "tileoffset": {
    "x": 0,
    "y": 0
  },
  "grid": {
    "orientation": "orthogonal",
    "width": 16,
    "height": 16
  },
  "properties": [],
  "terrains": [],
  "wangsets": [],
  "tiles": []
}
```

**Tileset Types:**
1. **Embedded**: All data in the map file
2. **External**: Reference to a `.tsj` file via `"source"` property

**Individual Tile Definitions:**
```json
{
  "tiles": [
    {
      "id": 0,
      "type": "floor",
      "class": "stone_floor",
      "probability": 1.0,
      "properties": [
        {
          "name": "walkable",
          "type": "bool",
          "value": true
        }
      ],
      "terrain": [0, 0, 0, 0],
      "animation": [
        {
          "tileid": 0,
          "duration": 100
        },
        {
          "tileid": 1,
          "duration": 100
        }
      ],
      "objectgroup": {
        "draworder": "index",
        "objects": []
      }
    }
  ]
}
```

**Image Collection Tilesets:**
Instead of a single image, each tile can have its own image:

```json
{
  "tiles": [
    {
      "id": 0,
      "image": "objects/chest.png",
      "imagewidth": 32,
      "imageheight": 32
    }
  ]
}
```

### 1.5 Property System

Custom properties provide metadata for maps, layers, objects, and tiles:

```json
{
  "properties": [
    {
      "name": "difficulty",
      "type": "int",
      "value": 5
    },
    {
      "name": "theme",
      "type": "string",
      "value": "cave"
    },
    {
      "name": "fogOfWar",
      "type": "bool",
      "value": true
    },
    {
      "name": "ambientLight",
      "type": "color",
      "value": "#3366aa"
    },
    {
      "name": "backgroundMusic",
      "type": "file",
      "value": "audio/dungeon.ogg"
    },
    {
      "name": "spawnRate",
      "type": "float",
      "value": 2.5
    },
    {
      "name": "linkedDoor",
      "type": "object",
      "value": 42
    },
    {
      "name": "entityType",
      "type": "class",
      "propertytype": "Entity",
      "value": {
        "health": 100,
        "speed": 3.5
      }
    }
  ]
}
```

**Property Types:**
- `string`: Text values
- `int`: Integer numbers
- `float`: Floating-point numbers
- `bool`: Boolean true/false
- `color`: Hex color (#RRGGBB or #AARRGGBB)
- `file`: File path reference
- `object`: Reference to object ID
- `class`: Custom type with nested properties (since 1.8)

### 1.6 Coordinate Systems

**Tile Coordinates:**
- X increases left to right
- Y increases top to bottom
- Origin (0,0) is top-left corner

**Pixel Coordinates:**
- Same axes as tile coordinates
- Objects use pixel coordinates
- Layer offsets in pixels

**Polygon/Polyline Points:**
- Relative to parent object's position
- First point is often (0, 0)

**Parallax Scrolling:**
- `parallaxx`, `parallaxy`: Scroll factors (default 1.0)
- Values < 1.0: Slower scrolling (background effect)
- Values > 1.0: Faster scrolling (foreground effect)

---

## 2. Existing Go Libraries for Tiled

### 2.1 Survey of Available Libraries

After extensive research, the following Go libraries were identified:

#### github.com/lafriks/go-tiled

**Status:** Most feature-complete and actively maintained

**Features:**
- Parse TMX (XML) format maps and tilesets
- Render maps to images
- Support for embedded filesystems
- Orthogonal finite maps

**Limitations:**
- No JSON/TMJ format support mentioned
- Import/render only, no export functionality
- Limited to orthogonal maps

**API Example:**
```go
import "github.com/lafriks/go-tiled"
import "github.com/lafriks/go-tiled/render"

gameMap, err := tiled.LoadFile("maps/dungeon.tmx")
if err != nil {
    return err
}

renderer, err := render.NewRenderer(gameMap)
if err != nil {
    return err
}

err = renderer.RenderLayer(0)
img := renderer.Result // *image.RGBA
```

**GitHub:** https://github.com/lafriks/go-tiled

**License:** MIT

#### github.com/salviati/go-tmx

**Features:**
- Basic TMX parsing
- Simple API

**Limitations:**
- TMX format only
- Limited documentation
- Appears less actively maintained

#### github.com/fardog/tmx

**Features:**
- Pure Go TMX parser
- No external dependencies
- Well-tested

**Limitations:**
- TMX format only
- No rendering capabilities
- Import-focused

#### Other Libraries

- **github.com/go-stuff/tiled**: TMX unmarshalling to structs
- **github.com/divVerent/tmx**: Another TMX parser
- **github.com/yanndr/tmx**: TMX format parser
- **azul3d.org/engine/tmx**: Part of larger game engine

### 2.2 TMJ-Specific Support

**Critical Finding:** None of the surveyed Go libraries explicitly mention TMJ/JSON format support. All focus on the older TMX (XML) format.

**Implications:**
- Custom TMJ export implementation likely needed
- Opportunity to create first-class TMJ support in Go
- Can leverage existing TMX library structures as reference

---

## 3. Implementation Approach

### 3.1 Library vs Custom Encoding

Given the lack of TMJ export support in existing Go libraries, we recommend a **custom encoding approach** using Go's standard library.

**Rationale:**
1. **No existing TMJ export libraries**: All libraries focus on import/parsing
2. **Simple format**: TMJ is straightforward JSON, well-suited to Go structs
3. **Standard library sufficiency**: `encoding/json` handles all requirements
4. **Full control**: Can optimize for specific use case (dungeon generation)
5. **Maintainability**: No external dependencies for core functionality

### 3.2 Recommended Architecture

**Three-Layer Approach:**

1. **Domain Layer**: Internal map representation
   - Optimized for game logic
   - Flexible structure for procedural generation
   - No Tiled-specific concerns

2. **Export Layer**: TMJ-specific structures
   - Go structs matching TMJ specification
   - JSON tags for proper serialization
   - Validation logic

3. **Translation Layer**: Domain → TMJ conversion
   - Maps internal structures to TMJ format
   - Handles GID calculation and layer ordering
   - Applies export configuration

**Directory Structure:**
```
dungo/
├── internal/
│   ├── tilemap/
│   │   ├── map.go           # Domain map structures
│   │   ├── layer.go         # Domain layer structures
│   │   ├── tile.go          # Domain tile structures
│   │   └── object.go        # Domain object structures
│   └── export/
│       ├── tmj/
│       │   ├── format.go    # TMJ struct definitions
│       │   ├── encoder.go   # JSON encoding logic
│       │   └── converter.go # Domain → TMJ conversion
│       └── exporter.go      # Export orchestration
```

### 3.3 Integration with Existing Libraries

For **import** functionality (reading existing TMJ files), consider:

1. **github.com/lafriks/go-tiled**: Use for reference, adapt structs for TMJ
2. **Custom JSON unmarshaling**: Implement complementary import logic
3. **Round-trip testing**: Ensure export→import consistency

**Hybrid Approach:**
```go
// Use go-tiled for TMX import if needed
import "github.com/lafriks/go-tiled"

// Custom TMJ export
import "github.com/dshills/dungo/internal/export/tmj"

// Convert between formats
func ConvertTMXToTMJ(tmxPath, tmjPath string) error {
    // Load with go-tiled
    tiledMap, err := tiled.LoadFile(tmxPath)
    if err != nil {
        return err
    }

    // Convert to internal format
    internalMap := convertTiledToInternal(tiledMap)

    // Export as TMJ
    return tmj.Export(internalMap, tmjPath)
}
```

### 3.4 Performance Considerations

**Encoding Strategies:**

1. **Small Maps (< 50x50)**: Use uncompressed CSV encoding
   - Fast encoding/decoding
   - Human-readable debugging
   - Minimal size difference

2. **Large Maps (> 100x100)**: Use base64 + gzip compression
   - Significant size reduction
   - Faster network transfer
   - Standard library support

3. **Very Large/Infinite Maps**: Use chunk-based encoding
   - Only serialize active chunks
   - Lazy loading support
   - Memory efficient

**Optimization Tips:**
- Pre-allocate slices for tile data
- Use `json.Encoder` for streaming output
- Pool buffers for compression
- Validate data before encoding to catch errors early

---

## 4. Code Example: Basic TMJ Structure

### 4.1 Core Type Definitions

```go
package tmj

import (
    "encoding/json"
    "io"
)

// Map represents the root TMJ map structure
type Map struct {
    Version        string      `json:"version"`
    TiledVersion   string      `json:"tiledversion"`
    Type           string      `json:"type"`
    Width          int         `json:"width"`
    Height         int         `json:"height"`
    TileWidth      int         `json:"tilewidth"`
    TileHeight     int         `json:"tileheight"`
    Orientation    string      `json:"orientation"`
    RenderOrder    string      `json:"renderorder"`
    Infinite       bool        `json:"infinite"`
    NextLayerID    int         `json:"nextlayerid"`
    NextObjectID   int         `json:"nextobjectid"`
    BackgroundColor *string    `json:"backgroundcolor,omitempty"`
    Class          string      `json:"class,omitempty"`
    CompressionLevel int       `json:"compressionlevel"`
    Layers         []Layer     `json:"layers"`
    Tilesets       []Tileset   `json:"tilesets"`
    Properties     []Property  `json:"properties,omitempty"`
}

// Layer represents any layer type (tile, object, image, group)
// Uses interface{} for Data to support multiple formats
type Layer struct {
    ID          int         `json:"id"`
    Name        string      `json:"name"`
    Type        string      `json:"type"` // "tilelayer", "objectgroup", "imagelayer", "group"
    Visible     bool        `json:"visible"`
    Opacity     float64     `json:"opacity"`
    X           int         `json:"x"`
    Y           int         `json:"y"`
    Width       int         `json:"width,omitempty"`
    Height      int         `json:"height,omitempty"`
    OffsetX     int         `json:"offsetx,omitempty"`
    OffsetY     int         `json:"offsety,omitempty"`
    ParallaxX   float64     `json:"parallaxx,omitempty"`
    ParallaxY   float64     `json:"parallaxy,omitempty"`
    Class       string      `json:"class,omitempty"`
    Properties  []Property  `json:"properties,omitempty"`

    // Tile layer specific
    Data        []uint32    `json:"data,omitempty"`
    Encoding    string      `json:"encoding,omitempty"`
    Compression string      `json:"compression,omitempty"`
    Chunks      []Chunk     `json:"chunks,omitempty"`

    // Object layer specific
    DrawOrder   string      `json:"draworder,omitempty"`
    Objects     []Object    `json:"objects,omitempty"`

    // Image layer specific
    Image       string      `json:"image,omitempty"`
    TransparentColor *string `json:"transparentcolor,omitempty"`
    RepeatX     bool        `json:"repeatx,omitempty"`
    RepeatY     bool        `json:"repeaty,omitempty"`

    // Group layer specific
    Layers      []Layer     `json:"layers,omitempty"`
}

// Chunk represents a tile data chunk for infinite maps
type Chunk struct {
    X      int      `json:"x"`
    Y      int      `json:"y"`
    Width  int      `json:"width"`
    Height int      `json:"height"`
    Data   []uint32 `json:"data"`
}

// Object represents an entity or collision shape
type Object struct {
    ID         int         `json:"id"`
    Name       string      `json:"name"`
    Type       string      `json:"type,omitempty"`
    Class      string      `json:"class,omitempty"`
    X          float64     `json:"x"`
    Y          float64     `json:"y"`
    Width      float64     `json:"width"`
    Height     float64     `json:"height"`
    Rotation   float64     `json:"rotation"`
    GID        uint32      `json:"gid,omitempty"`
    Visible    bool        `json:"visible"`
    Properties []Property  `json:"properties,omitempty"`

    // Shape types
    Ellipse   bool        `json:"ellipse,omitempty"`
    Point     bool        `json:"point,omitempty"`
    Polygon   []Point     `json:"polygon,omitempty"`
    Polyline  []Point     `json:"polyline,omitempty"`
    Text      *Text       `json:"text,omitempty"`
}

// Point represents a coordinate in a polygon or polyline
type Point struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

// Text represents text object formatting
type Text struct {
    Text       string `json:"text"`
    FontFamily string `json:"fontfamily,omitempty"`
    PixelSize  int    `json:"pixelsize,omitempty"`
    Wrap       bool   `json:"wrap,omitempty"`
    Color      string `json:"color,omitempty"`
    Bold       bool   `json:"bold,omitempty"`
    Italic     bool   `json:"italic,omitempty"`
    Underline  bool   `json:"underline,omitempty"`
    Strikeout  bool   `json:"strikeout,omitempty"`
    Kerning    bool   `json:"kerning,omitempty"`
    HAlign     string `json:"halign,omitempty"`
    VAlign     string `json:"valign,omitempty"`
}

// Tileset references a collection of tiles
type Tileset struct {
    FirstGID    uint32       `json:"firstgid"`
    Source      string       `json:"source,omitempty"` // For external tilesets
    Name        string       `json:"name,omitempty"`
    Class       string       `json:"class,omitempty"`
    TileWidth   int          `json:"tilewidth,omitempty"`
    TileHeight  int          `json:"tileheight,omitempty"`
    Spacing     int          `json:"spacing,omitempty"`
    Margin      int          `json:"margin,omitempty"`
    TileCount   int          `json:"tilecount,omitempty"`
    Columns     int          `json:"columns,omitempty"`
    Image       string       `json:"image,omitempty"`
    ImageWidth  int          `json:"imagewidth,omitempty"`
    ImageHeight int          `json:"imageheight,omitempty"`
    TransparentColor *string `json:"transparentcolor,omitempty"`
    TileOffset  *TileOffset  `json:"tileoffset,omitempty"`
    Grid        *Grid        `json:"grid,omitempty"`
    Properties  []Property   `json:"properties,omitempty"`
    Terrains    []Terrain    `json:"terrains,omitempty"`
    WangSets    []WangSet    `json:"wangsets,omitempty"`
    Tiles       []Tile       `json:"tiles,omitempty"`
}

// TileOffset represents tileset tile rendering offset
type TileOffset struct {
    X int `json:"x"`
    Y int `json:"y"`
}

// Grid defines tileset grid structure
type Grid struct {
    Orientation string `json:"orientation"`
    Width       int    `json:"width"`
    Height      int    `json:"height"`
}

// Tile defines an individual tile with properties
type Tile struct {
    ID          int          `json:"id"`
    Type        string       `json:"type,omitempty"`
    Class       string       `json:"class,omitempty"`
    Image       string       `json:"image,omitempty"`
    ImageWidth  int          `json:"imagewidth,omitempty"`
    ImageHeight int          `json:"imageheight,omitempty"`
    Probability float64      `json:"probability,omitempty"`
    Properties  []Property   `json:"properties,omitempty"`
    Terrain     []int        `json:"terrain,omitempty"`
    Animation   []Frame      `json:"animation,omitempty"`
    ObjectGroup *ObjectGroup `json:"objectgroup,omitempty"`
}

// Frame represents an animation frame
type Frame struct {
    TileID   int `json:"tileid"`
    Duration int `json:"duration"` // milliseconds
}

// ObjectGroup for tile collision shapes
type ObjectGroup struct {
    DrawOrder string   `json:"draworder"`
    Objects   []Object `json:"objects"`
}

// Terrain defines terrain type
type Terrain struct {
    Name string `json:"name"`
    Tile int    `json:"tile"`
}

// WangSet defines Wang tile patterns
type WangSet struct {
    Name        string      `json:"name"`
    Class       string      `json:"class,omitempty"`
    Tile        int         `json:"tile"`
    Colors      []WangColor `json:"colors"`
    WangTiles   []WangTile  `json:"wangtiles"`
}

// WangColor defines a Wang tile color
type WangColor struct {
    Color       string  `json:"color"`
    Name        string  `json:"name"`
    Probability float64 `json:"probability"`
    Tile        int     `json:"tile"`
}

// WangTile defines Wang tile pattern
type WangTile struct {
    TileID int   `json:"tileid"`
    WangID []int `json:"wangid"` // 8 color indices
}

// Property represents a custom property
type Property struct {
    Name         string      `json:"name"`
    Type         string      `json:"type"`
    Value        interface{} `json:"value"`
    PropertyType string      `json:"propertytype,omitempty"` // For class types
}
```

### 4.2 Encoding Functions

```go
// Encode writes a Map to JSON format
func (m *Map) Encode(w io.Writer) error {
    encoder := json.NewEncoder(w)
    encoder.SetIndent("", "  ") // Pretty print
    return encoder.Encode(m)
}

// EncodeCompact writes a Map to compact JSON
func (m *Map) EncodeCompact(w io.Writer) error {
    encoder := json.NewEncoder(w)
    return encoder.Encode(m)
}

// Marshal returns the JSON encoding of Map
func (m *Map) Marshal() ([]byte, error) {
    return json.MarshalIndent(m, "", "  ")
}

// MarshalCompact returns compact JSON encoding
func (m *Map) MarshalCompact() ([]byte, error) {
    return json.Marshal(m)
}
```

### 4.3 Helper Functions

```go
// NewMap creates a basic orthogonal map
func NewMap(width, height, tileWidth, tileHeight int) *Map {
    return &Map{
        Version:      "1.10",
        TiledVersion: "1.10.2",
        Type:         "map",
        Width:        width,
        Height:       height,
        TileWidth:    tileWidth,
        TileHeight:   tileHeight,
        Orientation:  "orthogonal",
        RenderOrder:  "right-down",
        Infinite:     false,
        NextLayerID:  1,
        NextObjectID: 1,
        CompressionLevel: -1,
        Layers:       []Layer{},
        Tilesets:     []Tileset{},
        Properties:   []Property{},
    }
}

// AddTileLayer adds a tile layer to the map
func (m *Map) AddTileLayer(name string, data []uint32) *Layer {
    layer := Layer{
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

// AddObjectLayer adds an object layer to the map
func (m *Map) AddObjectLayer(name string) *Layer {
    layer := Layer{
        ID:        m.NextLayerID,
        Name:      name,
        Type:      "objectgroup",
        Visible:   true,
        Opacity:   1.0,
        DrawOrder: "topdown",
        Objects:   []Object{},
    }
    m.NextLayerID++
    m.Layers = append(m.Layers, layer)
    return &m.Layers[len(m.Layers)-1]
}

// AddObject adds an object to an object layer
func (l *Layer) AddObject(obj Object, m *Map) {
    if l.Type != "objectgroup" {
        return
    }
    obj.ID = m.NextObjectID
    m.NextObjectID++
    l.Objects = append(l.Objects, obj)
}

// AddTileset adds a tileset reference
func (m *Map) AddTileset(name, imagePath string, tileWidth, tileHeight int) *Tileset {
    firstGID := uint32(1)
    if len(m.Tilesets) > 0 {
        last := m.Tilesets[len(m.Tilesets)-1]
        firstGID = last.FirstGID + uint32(last.TileCount)
    }

    tileset := Tileset{
        FirstGID:   firstGID,
        Name:       name,
        TileWidth:  tileWidth,
        TileHeight: tileHeight,
        Image:      imagePath,
    }
    m.Tilesets = append(m.Tilesets, tileset)
    return &m.Tilesets[len(m.Tilesets)-1]
}

// CalculateGID converts tileset local ID to global ID
func CalculateGID(tilesetFirstGID uint32, localTileID int, flipH, flipV, flipD bool) uint32 {
    gid := tilesetFirstGID + uint32(localTileID)

    const (
        FlippedHorizontallyFlag = 0x80000000
        FlippedVerticallyFlag   = 0x40000000
        FlippedDiagonallyFlag   = 0x20000000
    )

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

// ParseGID extracts tile ID and flip flags from GID
func ParseGID(gid uint32) (tileID uint32, flipH, flipV, flipD bool) {
    const (
        FlippedHorizontallyFlag = 0x80000000
        FlippedVerticallyFlag   = 0x40000000
        FlippedDiagonallyFlag   = 0x20000000
        TileIDMask              = 0x1FFFFFFF
    )

    flipH = (gid & FlippedHorizontallyFlag) != 0
    flipV = (gid & FlippedVerticallyFlag) != 0
    flipD = (gid & FlippedDiagonallyFlag) != 0
    tileID = gid & TileIDMask

    return
}
```

### 4.4 Usage Example

```go
package main

import (
    "os"
    "github.com/dshills/dungo/internal/export/tmj"
)

func main() {
    // Create a 32x24 map with 16x16 tiles
    gameMap := tmj.NewMap(32, 24, 16, 16)
    gameMap.Class = "dungeon"
    gameMap.BackgroundColor = ptr("#1a1a2e")

    // Add tileset
    tileset := gameMap.AddTileset(
        "dungeon_tiles",
        "tilesets/dungeon.png",
        16, 16,
    )
    tileset.TileCount = 256
    tileset.Columns = 16
    tileset.ImageWidth = 256
    tileset.ImageHeight = 256

    // Create floor layer
    floorData := make([]uint32, 32*24)
    for i := range floorData {
        floorData[i] = tmj.CalculateGID(1, 0, false, false, false) // Tile 0
    }
    floorLayer := gameMap.AddTileLayer("floor", floorData)
    floorLayer.Class = "ground"

    // Create wall layer
    wallData := make([]uint32, 32*24)
    // Add walls around perimeter
    for y := 0; y < 24; y++ {
        for x := 0; x < 32; x++ {
            if x == 0 || x == 31 || y == 0 || y == 23 {
                wallData[y*32+x] = tmj.CalculateGID(1, 1, false, false, false)
            }
        }
    }
    wallLayer := gameMap.AddTileLayer("walls", wallData)
    wallLayer.Class = "solid"

    // Create entity layer
    entityLayer := gameMap.AddObjectLayer("entities")

    // Add player spawn point
    playerSpawn := tmj.Object{
        Name:    "player_spawn",
        Class:   "spawn_point",
        X:       256.0,
        Y:       192.0,
        Width:   16.0,
        Height:  16.0,
        Visible: true,
        Properties: []tmj.Property{
            {
                Name:  "facing",
                Type:  "string",
                Value: "south",
            },
        },
    }
    entityLayer.AddObject(playerSpawn, gameMap)

    // Add enemy spawn
    enemySpawn := tmj.Object{
        Name:    "enemy_spawn",
        Class:   "spawn_point",
        X:       400.0,
        Y:       300.0,
        Width:   16.0,
        Height:  16.0,
        Visible: true,
        Properties: []tmj.Property{
            {
                Name:  "enemy_type",
                Type:  "string",
                Value: "goblin",
            },
            {
                Name:  "difficulty",
                Type:  "int",
                Value: 3,
            },
        },
    }
    entityLayer.AddObject(enemySpawn, gameMap)

    // Add trigger zone (polygon)
    triggerZone := tmj.Object{
        Name:    "entry_trigger",
        Class:   "trigger",
        X:       240.0,
        Y:       180.0,
        Width:   0.0,
        Height:  0.0,
        Visible: true,
        Polygon: []tmj.Point{
            {X: 0, Y: 0},
            {X: 32, Y: 0},
            {X: 32, Y: 32},
            {X: 0, Y: 32},
        },
        Properties: []tmj.Property{
            {
                Name:  "trigger_event",
                Type:  "string",
                Value: "show_tutorial",
            },
        },
    }
    entityLayer.AddObject(triggerZone, gameMap)

    // Add map-level properties
    gameMap.Properties = append(gameMap.Properties,
        tmj.Property{
            Name:  "difficulty",
            Type:  "int",
            Value: 5,
        },
        tmj.Property{
            Name:  "theme",
            Type:  "string",
            Value: "cave",
        },
        tmj.Property{
            Name:  "fogOfWar",
            Type:  "bool",
            Value: true,
        },
    )

    // Write to file
    file, err := os.Create("output/dungeon.tmj")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    if err := gameMap.Encode(file); err != nil {
        panic(err)
    }
}

func ptr(s string) *string {
    return &s
}
```

### 4.5 Advanced: Compression Support

```go
package tmj

import (
    "bytes"
    "compress/gzip"
    "compress/zlib"
    "encoding/base64"
)

// EncodeDataCompressed encodes tile data with compression
func EncodeDataCompressed(data []uint32, compression string) (string, error) {
    // Convert uint32 slice to byte slice (little-endian)
    buf := new(bytes.Buffer)
    for _, gid := range data {
        buf.WriteByte(byte(gid))
        buf.WriteByte(byte(gid >> 8))
        buf.WriteByte(byte(gid >> 16))
        buf.WriteByte(byte(gid >> 24))
    }

    var compressed bytes.Buffer
    var err error

    switch compression {
    case "gzip":
        w := gzip.NewWriter(&compressed)
        _, err = w.Write(buf.Bytes())
        if err != nil {
            return "", err
        }
        err = w.Close()

    case "zlib":
        w := zlib.NewWriter(&compressed)
        _, err = w.Write(buf.Bytes())
        if err != nil {
            return "", err
        }
        err = w.Close()

    default:
        compressed = *buf
    }

    if err != nil {
        return "", err
    }

    // Base64 encode
    return base64.StdEncoding.EncodeToString(compressed.Bytes()), nil
}

// SetCompressedData sets layer data with compression
func (l *Layer) SetCompressedData(data []uint32, compression string) error {
    encoded, err := EncodeDataCompressed(data, compression)
    if err != nil {
        return err
    }

    l.Encoding = "base64"
    l.Compression = compression
    // Store as string in custom field, or modify Layer struct to support string data
    // For now, this is a conceptual example

    return nil
}
```

---

## 5. Best Practices for Tile Map Organization

### 5.1 Layer Organization

**Recommended Layer Structure for Dungeon Generation:**

1. **floor** (tilelayer)
   - Base terrain tiles
   - Z-index: 0
   - Always visible
   - Class: "ground"

2. **floor_decor** (tilelayer)
   - Carpet, cracks, stains
   - Z-index: 1
   - Can use blend modes
   - Class: "decor"

3. **walls** (tilelayer)
   - Wall tiles
   - Z-index: 2
   - Collision handled separately
   - Class: "solid"

4. **wall_decor** (tilelayer)
   - Banners, sconces, paintings
   - Z-index: 3
   - Class: "decor"

5. **doors** (tilelayer)
   - Door tiles (closed state)
   - Z-index: 4
   - Animated tiles for opening
   - Class: "interactive"

6. **furniture** (tilelayer)
   - Tables, chairs, chests
   - Z-index: 5
   - Some may need object layer counterparts
   - Class: "decor" or "interactive"

7. **triggers** (objectgroup)
   - Invisible trigger zones
   - No visual representation
   - Polygon/rectangle shapes
   - Class: "trigger"
   - Visible: false in final build

8. **hazards** (objectgroup)
   - Traps, lava, spikes
   - Can reference tiles with GID
   - Include damage properties
   - Class: "hazard"

9. **entities** (objectgroup)
   - Player spawn
   - Enemy spawns
   - NPC locations
   - Item spawns
   - Class: "spawn_point"

10. **collision** (objectgroup)
    - Collision shapes
    - Separate from visual walls
    - Can be simplified shapes
    - Class: "collision"
    - Visible: false

11. **lighting** (objectgroup)
    - Light sources
    - Point objects for dynamic lights
    - Properties: radius, color, intensity
    - Class: "light"

12. **overlay** (tilelayer)
    - Fog of war
    - Shadow overlay
    - Z-index: 100
    - Opacity: 0.5-0.8
    - ParallaxX/Y: 1.0

**Layer Naming Conventions:**
- Use lowercase with underscores
- Prefix special layers: `debug_`, `editor_`
- Group related layers: `floor_`, `wall_`, `decor_`
- Suffix state variants: `_open`, `_closed`, `_broken`

### 5.2 Tileset Organization

**Single Tileset vs Multiple Tilesets:**

**Single Tileset (Recommended for most cases):**
- Simpler GID calculation
- Better performance (single texture)
- Easier to manage
- Use 256x256 or 512x512 atlas

**Multiple Tilesets (Use when):**
- Different tile sizes needed
- Mixing themed content
- External tileset reuse across maps
- Image collection for unique objects

**Tileset Structure:**
```
dungeon.png (256x256, 16x16 tiles = 256 tiles)
├── Tiles 0-63:   Floor variants
├── Tiles 64-127: Wall variants
├── Tiles 128-159: Door variants
├── Tiles 160-191: Furniture
├── Tiles 192-223: Decorations
└── Tiles 224-255: Special/effects
```

**Tile Property Best Practices:**
```json
{
  "tiles": [
    {
      "id": 0,
      "class": "stone_floor",
      "properties": [
        {
          "name": "walkable",
          "type": "bool",
          "value": true
        },
        {
          "name": "movement_cost",
          "type": "float",
          "value": 1.0
        },
        {
          "name": "sound",
          "type": "string",
          "value": "footstep_stone"
        }
      ]
    },
    {
      "id": 65,
      "class": "stone_wall",
      "properties": [
        {
          "name": "walkable",
          "type": "bool",
          "value": false
        },
        {
          "name": "blocks_vision",
          "type": "bool",
          "value": true
        },
        {
          "name": "destructible",
          "type": "bool",
          "value": false
        }
      ]
    }
  ]
}
```

### 5.3 Object Layer Best Practices

**Object Naming:**
- Descriptive, unique names for important objects
- Type/Class for categorization
- Use IDs for reference between objects

**Object Properties:**
```go
// Example: Door object
door := tmj.Object{
    Name:  "door_main_entrance",
    Class: "door",
    Properties: []tmj.Property{
        {Name: "locked", Type: "bool", Value: true},
        {Name: "key_id", Type: "string", Value: "silver_key"},
        {Name: "linked_door", Type: "object", Value: 42}, // Object ID
        {Name: "open_sound", Type: "file", Value: "audio/door_open.ogg"},
    },
}

// Example: Enemy spawn
enemy := tmj.Object{
    Name:  "goblin_spawn_1",
    Class: "enemy_spawn",
    Properties: []tmj.Property{
        {Name: "enemy_type", Type: "string", Value: "goblin_warrior"},
        {Name: "level", Type: "int", Value: 5},
        {Name: "patrol_path", Type: "string", Value: "path_hallway_1"},
        {Name: "aggro_radius", Type: "float", Value: 128.0},
        {Name: "respawn_time", Type: "int", Value: 300}, // seconds
    },
}

// Example: Trigger zone
trigger := tmj.Object{
    Name:    "enter_boss_room",
    Class:   "trigger",
    Polygon: []tmj.Point{{X: 0, Y: 0}, {X: 64, Y: 0}, {X: 64, Y: 32}, {X: 0, Y: 32}},
    Properties: []tmj.Property{
        {Name: "trigger_once", Type: "bool", Value: true},
        {Name: "event_id", Type: "string", Value: "boss_intro"},
        {Name: "required_item", Type: "string", Value: "boss_key"},
    },
}
```

### 5.4 Custom Properties

**Define Property Types:**
Create a consistent schema across your game:

```json
{
  "propertytypes": [
    {
      "id": 1,
      "name": "Enemy",
      "type": "class",
      "members": [
        {
          "name": "health",
          "type": "int",
          "value": 100
        },
        {
          "name": "damage",
          "type": "int",
          "value": 10
        },
        {
          "name": "speed",
          "type": "float",
          "value": 2.0
        },
        {
          "name": "ai_type",
          "type": "string",
          "value": "melee"
        }
      ]
    }
  ]
}
```

**Property Naming:**
- Use snake_case for consistency
- Prefix booleans with `is_`, `has_`, `can_`, or use question form
- Use units in name: `speed_tiles_per_second`, `radius_pixels`
- Group related properties: `door_key_id`, `door_locked`, `door_hidden`

### 5.5 Coordinate System Considerations

**Pixel vs Tile Coordinates:**
- Tile layers: Index-based (0 to width-1, 0 to height-1)
- Objects: Pixel-based (0.0 to width*tilewidth, 0.0 to height*tileheight)
- Origin: Top-left (0, 0)

**Object Positioning:**
```go
// Center an object in a tile
func CenterInTile(tileX, tileY, tileWidth, tileHeight, objWidth, objHeight int) (float64, float64) {
    pixelX := float64(tileX*tileWidth + (tileWidth-objWidth)/2)
    pixelY := float64(tileY*tileHeight + (tileHeight-objHeight)/2)
    return pixelX, pixelY
}

// Tile to pixel
func TileToPixel(tileX, tileY, tileWidth, tileHeight int) (float64, float64) {
    return float64(tileX * tileWidth), float64(tileY * tileHeight)
}

// Pixel to tile
func PixelToTile(pixelX, pixelY float64, tileWidth, tileHeight int) (int, int) {
    return int(pixelX) / tileWidth, int(pixelY) / tileHeight
}
```

### 5.6 Performance Optimization

**Layer Optimization:**
1. Merge static decoration layers before export
2. Use appropriate compression for large maps
3. Limit objects per layer (consider spatial partitioning)
4. Use chunks for infinite maps

**Data Encoding:**
```go
// Choose encoding based on map size
func ChooseEncoding(width, height int) (encoding, compression string) {
    tileCount := width * height

    if tileCount < 2500 { // 50x50
        return "csv", ""
    } else if tileCount < 10000 { // 100x100
        return "base64", "gzip"
    } else {
        return "base64", "zstd"
    }
}
```

**Tileset Optimization:**
- Power-of-two dimensions (256, 512, 1024)
- Consistent tile size across tilesets
- Minimize transparent tiles
- Use external tilesets for reusability

### 5.7 Version Control

**TMJ in Git:**
```gitignore
# Don't ignore TMJ files
# *.tmj

# Ignore backups
*.tmj~
*.tmj.bak

# Ignore autosave
*_autosave.tmj
```

**Merge Conflict Resolution:**
- Pretty-print JSON for better diffs
- Use consistent property ordering
- Consider separate files for large maps
- Document map schema in repository

### 5.8 Documentation

**Map Properties for Documentation:**
```go
gameMap.Properties = append(gameMap.Properties,
    tmj.Property{
        Name:  "description",
        Type:  "string",
        Value: "Main entrance to the dungeon. Tutorial area.",
    },
    tmj.Property{
        Name:  "author",
        Type:  "string",
        Value: "MapGenerator v1.0",
    },
    tmj.Property{
        Name:  "version",
        Type:  "int",
        Value: 1,
    },
    tmj.Property{
        Name:  "created",
        Type:  "string",
        Value: time.Now().Format(time.RFC3339),
    },
)
```

**Layer Documentation:**
```go
layer.Properties = append(layer.Properties,
    tmj.Property{
        Name:  "note",
        Type:  "string",
        Value: "This layer contains all enemy spawn points for the first area",
    },
)
```

### 5.9 Testing and Validation

**Validation Checklist:**
- [ ] All GIDs reference valid tilesets
- [ ] No negative coordinates
- [ ] Object IDs are unique
- [ ] Layer IDs are unique
- [ ] Required properties present
- [ ] Tileset images exist
- [ ] No overlapping collision objects (if unintended)
- [ ] Trigger zones have proper polygon closure
- [ ] Map size matches tile data length

**Test TMJ Export:**
```go
func ValidateMap(m *tmj.Map) []error {
    var errors []error

    // Validate dimensions
    if m.Width <= 0 || m.Height <= 0 {
        errors = append(errors, fmt.Errorf("invalid map dimensions: %dx%d", m.Width, m.Height))
    }

    // Validate tile layers
    for i, layer := range m.Layers {
        if layer.Type == "tilelayer" {
            expectedLen := m.Width * m.Height
            if len(layer.Data) != expectedLen {
                errors = append(errors, fmt.Errorf("layer %d (%s): data length %d doesn't match dimensions %d",
                    i, layer.Name, len(layer.Data), expectedLen))
            }

            // Check GIDs
            for j, gid := range layer.Data {
                if gid > 0 {
                    tileID, _, _, _ := tmj.ParseGID(gid)
                    if !m.IsValidGID(tileID) {
                        errors = append(errors, fmt.Errorf("layer %d (%s): invalid GID %d at index %d",
                            i, layer.Name, gid, j))
                    }
                }
            }
        }
    }

    // Validate object uniqueness
    objectIDs := make(map[int]bool)
    for _, layer := range m.Layers {
        if layer.Type == "objectgroup" {
            for _, obj := range layer.Objects {
                if objectIDs[obj.ID] {
                    errors = append(errors, fmt.Errorf("duplicate object ID: %d", obj.ID))
                }
                objectIDs[obj.ID] = true
            }
        }
    }

    return errors
}

// IsValidGID checks if a GID references an existing tile
func (m *Map) IsValidGID(gid uint32) bool {
    for _, tileset := range m.Tilesets {
        if gid >= tileset.FirstGID && gid < tileset.FirstGID+uint32(tileset.TileCount) {
            return true
        }
    }
    return false
}
```

---

## 6. Recommended Next Steps

### 6.1 Implementation Roadmap

**Phase 1: Core TMJ Export**
1. Implement basic TMJ struct definitions
2. Create map builder with fluent API
3. Add layer management (tile, object)
4. Implement tileset handling
5. Add JSON encoding

**Phase 2: Advanced Features**
1. Property system with type safety
2. GID calculation utilities
3. Compression support (gzip, zlib)
4. Object shape helpers
5. Validation framework

**Phase 3: Integration**
1. Connect to dungeon generation
2. Create export templates
3. Add configuration system
4. Implement batch export
5. Error handling and logging

**Phase 4: Enhancement**
1. TMJ import (round-trip)
2. Map manipulation utilities
3. Layer merging/splitting
4. Tileset optimization
5. Documentation generation

### 6.2 Testing Strategy

**Unit Tests:**
- JSON encoding/decoding
- GID calculation
- Coordinate conversion
- Validation rules

**Integration Tests:**
- Export complete map
- Open in Tiled editor
- Verify all layers visible
- Check object properties
- Test with game engine

**Test Maps:**
Create reference maps in Tiled:
- minimal.tmj: Smallest valid map
- single-layer.tmj: One tile layer
- multi-layer.tmj: All layer types
- objects.tmj: All object shapes
- properties.tmj: All property types
- compressed.tmj: Compression test
- infinite.tmj: Chunk-based map

### 6.3 Tools and Resources

**Development Tools:**
- **Tiled Editor**: https://www.mapeditor.org/ (Qt-based, cross-platform)
- **JSON Formatter**: https://jsonformatter.org/ (validate output)
- **Go JSON to Struct**: https://mholt.github.io/json-to-go/ (generate structs)

**Documentation:**
- **TMJ Specification**: https://doc.mapeditor.org/en/stable/reference/json-map-format/
- **Tiled Manual**: https://doc.mapeditor.org/en/stable/manual/
- **Tiled Forum**: https://discourse.mapeditor.org/

**Example Assets:**
- **OpenGameArt**: https://opengameart.org/ (free tilesets)
- **Kenney.nl**: https://kenney.nl/assets (CC0 assets)
- **itch.io**: https://itch.io/game-assets/tag-tileset (various licenses)

---

## 7. Conclusion

### 7.1 Summary

The Tiled TMJ format provides a robust, well-documented JSON structure for tile maps with excellent editor support. While no existing Go libraries offer TMJ export functionality, the format is straightforward to implement using Go's standard library.

**Key Findings:**
1. TMJ is a comprehensive JSON format supporting all Tiled features
2. No existing Go libraries provide TMJ export (only TMX import)
3. Custom implementation using `encoding/json` is recommended
4. Format supports multiple layer types needed for dungeon generation
5. Objects and properties enable flexible entity system
6. Tiled editor provides excellent validation and testing

### 7.2 Recommended Approach

**For the Dungo project:**

1. **Implement custom TMJ export** using the struct definitions provided
2. **Layer structure**: floor, walls, doors, decor, triggers, hazards, entities
3. **Single tileset** approach for simplicity (256x256 atlas)
4. **CSV encoding** for development, gzip compression for production
5. **Object layers** for entities, triggers, and collision
6. **Property system** for game-specific metadata
7. **Validation** before export to catch errors early

**Benefits:**
- Full control over export format
- No external dependencies
- Optimized for dungeon generation use case
- Compatible with Tiled editor for manual editing
- Supports future enhancements

### 7.3 Example Integration

```go
// Example: Dungeon generator integration
func (d *Dungeon) ExportToTMJ(filepath string) error {
    // Create TMJ map
    tmjMap := tmj.NewMap(d.Width, d.Height, 16, 16)

    // Add tileset
    tileset := tmjMap.AddTileset("dungeon", "assets/dungeon.png", 16, 16)
    tileset.TileCount = 256
    tileset.Columns = 16

    // Export each layer
    tmjMap.AddTileLayer("floor", d.FloorLayer.ToGIDs(tileset.FirstGID))
    tmjMap.AddTileLayer("walls", d.WallLayer.ToGIDs(tileset.FirstGID))
    tmjMap.AddTileLayer("doors", d.DoorLayer.ToGIDs(tileset.FirstGID))
    tmjMap.AddTileLayer("decor", d.DecorLayer.ToGIDs(tileset.FirstGID))

    // Export entities
    entityLayer := tmjMap.AddObjectLayer("entities")
    for _, entity := range d.Entities {
        entityLayer.AddObject(entity.ToTMJObject(), tmjMap)
    }

    // Export triggers
    triggerLayer := tmjMap.AddObjectLayer("triggers")
    triggerLayer.Visible = false
    for _, trigger := range d.Triggers {
        triggerLayer.AddObject(trigger.ToTMJObject(), tmjMap)
    }

    // Validate
    if errors := tmj.ValidateMap(tmjMap); len(errors) > 0 {
        return fmt.Errorf("validation failed: %v", errors)
    }

    // Write to file
    file, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer file.Close()

    return tmjMap.Encode(file)
}
```

This research provides a complete foundation for implementing TMJ export in Go, specifically tailored for the Dungo dungeon generation project's requirements.

---

**Document Version:** 1.0
**Date:** 2025-11-04
**Author:** Research compiled for Dungo project
**References:** Tiled 1.11.0 documentation, Go 1.25.3 standard library
