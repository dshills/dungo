package carving

import (
	"fmt"
)

// NewTileMap creates a new TileMap with the specified dimensions.
func NewTileMap(width, height, tileWidth, tileHeight int) *TileMap {
	return &TileMap{
		Width:      width,
		Height:     height,
		TileWidth:  tileWidth,
		TileHeight: tileHeight,
		Layers:     make(map[string]*Layer),
	}
}

// AddLayer adds a new layer to the tile map.
func AddLayer(tm *TileMap, name, layerType string) *Layer {
	layerID := len(tm.Layers)
	layer := &Layer{
		ID:      layerID,
		Name:    name,
		Type:    layerType,
		Visible: true,
		Opacity: 1.0,
	}

	if layerType == "tilelayer" {
		layer.Data = make([]uint32, tm.Width*tm.Height)
	} else if layerType == "objectgroup" {
		layer.Objects = []Object{}
	}

	tm.Layers[name] = layer
	return layer
}

// GetTile retrieves the tile value at the given position.
// Returns 0 if the position is out of bounds.
func GetTile(data []uint32, x, y, width, height int) uint32 {
	if x < 0 || x >= width || y < 0 || y >= height {
		return 0
	}
	idx := y*width + x
	if idx < 0 || idx >= len(data) {
		return 0
	}
	return data[idx]
}

// SetTile sets the tile value at the given position.
// Returns an error if the position is out of bounds.
func SetTile(data []uint32, x, y, width, height int, value uint32) error {
	if x < 0 || x >= width || y < 0 || y >= height {
		return fmt.Errorf("position (%d, %d) out of bounds [0, %d) x [0, %d)", x, y, width, height)
	}
	idx := y*width + x
	if idx < 0 || idx >= len(data) {
		return fmt.Errorf("index %d out of data range [0, %d)", idx, len(data))
	}
	data[idx] = value
	return nil
}

// FillRect fills a rectangular region with the specified tile value.
func FillRect(data []uint32, x, y, w, h, width, height int, value uint32) error {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			tx := x + dx
			ty := y + dy
			if err := SetTile(data, tx, ty, width, height, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// DrawRect draws a rectangular outline with the specified tile value.
func DrawRect(data []uint32, x, y, w, h, width, height int, value uint32) error {
	// Top and bottom edges
	for dx := 0; dx < w; dx++ {
		if err := SetTile(data, x+dx, y, width, height, value); err != nil {
			return err
		}
		if err := SetTile(data, x+dx, y+h-1, width, height, value); err != nil {
			return err
		}
	}

	// Left and right edges
	for dy := 0; dy < h; dy++ {
		if err := SetTile(data, x, y+dy, width, height, value); err != nil {
			return err
		}
		if err := SetTile(data, x+w-1, y+dy, width, height, value); err != nil {
			return err
		}
	}

	return nil
}

// DrawLine draws a line between two points using Bresenham's algorithm.
func DrawLine(data []uint32, x0, y0, x1, y1, width, height int, value uint32) error {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)

	sx := -1
	if x0 < x1 {
		sx = 1
	}

	sy := -1
	if y0 < y1 {
		sy = 1
	}

	err := dx - dy

	for {
		if setErr := SetTile(data, x0, y0, width, height, value); setErr != nil {
			return setErr
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}

	return nil
}

// CountNeighbors counts how many neighbors of a given tile match the target value.
func CountNeighbors(data []uint32, x, y, width, height int, target uint32, includeDiagonal bool) int {
	count := 0
	deltas := []struct{ dx, dy int }{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Cardinal directions
	}

	if includeDiagonal {
		deltas = append(deltas, []struct{ dx, dy int }{
			{-1, -1}, {-1, 1}, {1, -1}, {1, 1}, // Diagonals
		}...)
	}

	for _, d := range deltas {
		nx, ny := x+d.dx, y+d.dy
		if GetTile(data, nx, ny, width, height) == target {
			count++
		}
	}

	return count
}

// FloodFill performs a flood fill starting at (x, y) with the given value.
// Only fills tiles that match the original tile value at the start position.
func FloodFill(data []uint32, x, y, width, height int, value uint32) error {
	if x < 0 || x >= width || y < 0 || y >= height {
		return fmt.Errorf("start position (%d, %d) out of bounds", x, y)
	}

	targetValue := GetTile(data, x, y, width, height)
	if targetValue == value {
		return nil // Already the target value
	}

	// BFS flood fill
	type point struct{ x, y int }
	queue := []point{{x, y}}
	visited := make(map[point]bool)

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		if visited[p] {
			continue
		}
		visited[p] = true

		if GetTile(data, p.x, p.y, width, height) != targetValue {
			continue
		}

		if err := SetTile(data, p.x, p.y, width, height, value); err != nil {
			return err
		}

		// Add neighbors to queue
		neighbors := []point{
			{p.x - 1, p.y}, {p.x + 1, p.y},
			{p.x, p.y - 1}, {p.x, p.y + 1},
		}

		for _, n := range neighbors {
			if n.x >= 0 && n.x < width && n.y >= 0 && n.y < height && !visited[n] {
				queue = append(queue, n)
			}
		}
	}

	return nil
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
