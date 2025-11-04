package carving

import (
	"fmt"
)

// CorridorRouter handles routing corridor paths onto tile layers.
type CorridorRouter struct {
	width  int
	height int
}

// NewCorridorRouter creates a new corridor router for the given tile map dimensions.
func NewCorridorRouter(width, height int) *CorridorRouter {
	return &CorridorRouter{
		width:  width,
		height: height,
	}
}

// RouteCorridor routes a corridor path onto the tile data.
// The path is drawn as a series of connected line segments.
func (r *CorridorRouter) RouteCorridor(path Path, tileData []uint32) error {
	if len(path.Points) < 2 {
		return fmt.Errorf("corridor path must have at least 2 points, got %d", len(path.Points))
	}

	// Draw lines between consecutive points
	for i := 0; i < len(path.Points)-1; i++ {
		p1 := path.Points[i]
		p2 := path.Points[i+1]

		if err := DrawLine(tileData, p1.X, p1.Y, p2.X, p2.Y, r.width, r.height, uint32(TileFloor)); err != nil {
			return fmt.Errorf("drawing line segment %d: %w", i, err)
		}
	}

	return nil
}

// RouteWideCorridor routes a corridor with the specified width.
// This creates a corridor that is wider than a single tile.
func (r *CorridorRouter) RouteWideCorridor(path Path, corridorWidth int, tileData []uint32) error {
	if len(path.Points) < 2 {
		return fmt.Errorf("corridor path must have at least 2 points, got %d", len(path.Points))
	}

	if corridorWidth < 1 {
		corridorWidth = 1
	}

	// For each segment, draw multiple parallel lines
	for i := 0; i < len(path.Points)-1; i++ {
		p1 := path.Points[i]
		p2 := path.Points[i+1]

		// Calculate perpendicular offset
		dx := p2.X - p1.X
		dy := p2.Y - p1.Y

		// Determine if horizontal or vertical segment
		if abs(dx) > abs(dy) {
			// Horizontal segment - offset vertically
			offset := corridorWidth / 2
			for w := -offset; w <= offset; w++ {
				if err := DrawLine(tileData, p1.X, p1.Y+w, p2.X, p2.Y+w, r.width, r.height, uint32(TileFloor)); err != nil {
					return fmt.Errorf("drawing wide line segment %d (offset %d): %w", i, w, err)
				}
			}
		} else {
			// Vertical segment - offset horizontally
			offset := corridorWidth / 2
			for w := -offset; w <= offset; w++ {
				if err := DrawLine(tileData, p1.X+w, p1.Y, p2.X+w, p2.Y, r.width, r.height, uint32(TileFloor)); err != nil {
					return fmt.Errorf("drawing wide line segment %d (offset %d): %w", i, w, err)
				}
			}
		}
	}

	return nil
}

// RouteSmoothCorridor routes a corridor with rounded corners.
// This uses a simple approach of adding extra tiles at corners.
func (r *CorridorRouter) RouteSmoothCorridor(path Path, tileData []uint32) error {
	if len(path.Points) < 2 {
		return fmt.Errorf("corridor path must have at least 2 points, got %d", len(path.Points))
	}

	// First, draw the basic corridor
	if err := r.RouteCorridor(path, tileData); err != nil {
		return err
	}

	// Then smooth corners by filling in diagonal neighbors at corner points
	for i := 1; i < len(path.Points)-1; i++ {
		prev := path.Points[i-1]
		curr := path.Points[i]
		next := path.Points[i+1]

		// Check if this is a corner (direction change)
		dx1 := curr.X - prev.X
		dy1 := curr.Y - prev.Y
		dx2 := next.X - curr.X
		dy2 := next.Y - curr.Y

		// If direction changed, add corner smoothing
		if (dx1 != 0 && dy2 != 0) || (dy1 != 0 && dx2 != 0) {
			// Add diagonal tile
			if dx1 != 0 && dy2 != 0 {
				// Horizontal to vertical
				offsetX := 0
				if dx1 > 0 {
					offsetX = 1
				} else if dx1 < 0 {
					offsetX = -1
				}
				offsetY := 0
				if dy2 > 0 {
					offsetY = 1
				} else if dy2 < 0 {
					offsetY = -1
				}
				_ = SetTile(tileData, curr.X+offsetX, curr.Y+offsetY, r.width, r.height, uint32(TileFloor))
			} else if dy1 != 0 && dx2 != 0 {
				// Vertical to horizontal
				offsetX := 0
				if dx2 > 0 {
					offsetX = 1
				} else if dx2 < 0 {
					offsetX = -1
				}
				offsetY := 0
				if dy1 > 0 {
					offsetY = 1
				} else if dy1 < 0 {
					offsetY = -1
				}
				_ = SetTile(tileData, curr.X+offsetX, curr.Y+offsetY, r.width, r.height, uint32(TileFloor))
			}
		}
	}

	return nil
}

// RouteNaturalCorridor routes a corridor with a more organic, cave-like appearance.
// This adds random variation to the corridor width.
func (r *CorridorRouter) RouteNaturalCorridor(path Path, baseWidth int, tileData []uint32, rng func(int) int) error {
	if len(path.Points) < 2 {
		return fmt.Errorf("corridor path must have at least 2 points, got %d", len(path.Points))
	}

	if baseWidth < 1 {
		baseWidth = 1
	}

	// For each segment, vary the width slightly
	for i := 0; i < len(path.Points)-1; i++ {
		// Determine width for this segment (baseWidth +/- 1)
		variation := 0
		if rng != nil && baseWidth > 1 {
			variation = rng(3) - 1 // -1, 0, or 1
		}
		segmentWidth := baseWidth + variation
		if segmentWidth < 1 {
			segmentWidth = 1
		}

		// Create a path with just this segment
		segment := Path{
			Points: []Point{path.Points[i], path.Points[i+1]},
		}

		if err := r.RouteWideCorridor(segment, segmentWidth, tileData); err != nil {
			return fmt.Errorf("drawing natural corridor segment %d: %w", i, err)
		}
	}

	return nil
}

// SimplifyPath reduces the number of points in a path by removing collinear points.
// This helps optimize corridor rendering.
func SimplifyPath(path Path) Path {
	if len(path.Points) <= 2 {
		return path
	}

	simplified := []Point{path.Points[0]}

	for i := 1; i < len(path.Points)-1; i++ {
		prev := path.Points[i-1]
		curr := path.Points[i]
		next := path.Points[i+1]

		// Check if current point is collinear with prev and next
		dx1 := curr.X - prev.X
		dy1 := curr.Y - prev.Y
		dx2 := next.X - curr.X
		dy2 := next.Y - curr.Y

		// If not collinear, keep the point
		if !isCollinear(dx1, dy1, dx2, dy2) {
			simplified = append(simplified, curr)
		}
	}

	// Always keep the last point
	simplified = append(simplified, path.Points[len(path.Points)-1])

	return Path{Points: simplified}
}

// isCollinear checks if two direction vectors are collinear.
func isCollinear(dx1, dy1, dx2, dy2 int) bool {
	// Two vectors are collinear if their cross product is zero
	// or if they point in the same direction
	return dx1*dy2 == dy1*dx2
}

// PathLength calculates the total Manhattan distance of a path.
func PathLength(path Path) int {
	if len(path.Points) < 2 {
		return 0
	}

	length := 0
	for i := 0; i < len(path.Points)-1; i++ {
		p1 := path.Points[i]
		p2 := path.Points[i+1]
		length += abs(p2.X-p1.X) + abs(p2.Y-p1.Y)
	}

	return length
}

// SplitPathAtMidpoint splits a path into two paths at the midpoint.
// Useful for placing features in the middle of corridors.
func SplitPathAtMidpoint(path Path) (Path, Path, Point) {
	if len(path.Points) < 2 {
		return path, Path{}, Point{}
	}

	totalLength := PathLength(path)
	halfLength := totalLength / 2

	// Find the point at halfLength
	currentLength := 0
	for i := 0; i < len(path.Points)-1; i++ {
		p1 := path.Points[i]
		p2 := path.Points[i+1]
		segmentLength := abs(p2.X-p1.X) + abs(p2.Y-p1.Y)

		if currentLength+segmentLength >= halfLength {
			// Midpoint is in this segment
			remaining := halfLength - currentLength

			// Calculate midpoint position
			midX := p1.X
			midY := p1.Y

			if p2.X > p1.X {
				midX += remaining
			} else if p2.X < p1.X {
				midX -= remaining
			} else if p2.Y > p1.Y {
				midY += remaining
			} else if p2.Y < p1.Y {
				midY -= remaining
			}

			midpoint := Point{X: midX, Y: midY}

			// Create first half
			firstHalf := make([]Point, i+2)
			copy(firstHalf, path.Points[:i+1])
			firstHalf[i+1] = midpoint

			// Create second half
			secondHalf := make([]Point, len(path.Points)-i)
			secondHalf[0] = midpoint
			copy(secondHalf[1:], path.Points[i+1:])

			return Path{Points: firstHalf}, Path{Points: secondHalf}, midpoint
		}

		currentLength += segmentLength
	}

	// Fallback: split at actual midpoint index
	midIndex := len(path.Points) / 2
	return Path{Points: path.Points[:midIndex+1]},
		Path{Points: path.Points[midIndex:]},
		path.Points[midIndex]
}
