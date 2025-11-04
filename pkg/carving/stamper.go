package carving

import (
	"fmt"
)

// Stamper handles stamping room footprints onto tile layers.
type Stamper struct {
	width  int
	height int
}

// NewStamper creates a new stamper for the given tile map dimensions.
func NewStamper(width, height int) *Stamper {
	return &Stamper{
		width:  width,
		height: height,
	}
}

// StampRoom stamps a room's footprint onto the tile data at the given pose.
func (s *Stamper) StampRoom(room Room, pose Pose, tileData []uint32) error {
	if room == nil {
		return fmt.Errorf("room cannot be nil")
	}

	// Get room dimensions based on size
	roomWidth, roomHeight := s.getRoomDimensions(room.GetSize())

	// Calculate stamp position based on pose
	// The pose gives the center position, so we offset by half the room size
	startX := pose.X - roomWidth/2
	startY := pose.Y - roomHeight/2

	// Apply rotation if needed (0, 90, 180, 270 degrees)
	rotatedWidth, rotatedHeight := roomWidth, roomHeight
	if pose.Rotation == 90 || pose.Rotation == 270 {
		rotatedWidth, rotatedHeight = roomHeight, roomWidth
	}

	// Stamp the room shape
	switch room.GetSize() {
	case SizeXS, SizeS, SizeM, SizeL, SizeXL:
		// For now, all rooms are rectangular
		// Future enhancement: support different shapes based on FootprintID
		if err := s.stampRectangle(startX, startY, rotatedWidth, rotatedHeight, tileData); err != nil {
			return fmt.Errorf("stamping rectangle for room %s: %w", room.GetID(), err)
		}
	default:
		return fmt.Errorf("unsupported room size: %v", room.GetSize())
	}

	return nil
}

// getRoomDimensions returns the tile dimensions for a given room size.
// Mapping:
// XS: 3x3 tiles
// S: 5x5 tiles
// M: 7x7 tiles
// L: 10x10 tiles
// XL: 15x15 tiles
func (s *Stamper) getRoomDimensions(size RoomSize) (width, height int) {
	switch size {
	case SizeXS:
		return 3, 3
	case SizeS:
		return 5, 5
	case SizeM:
		return 7, 7
	case SizeL:
		return 10, 10
	case SizeXL:
		return 15, 15
	default:
		return 5, 5 // Default to small
	}
}

// stampRectangle stamps a rectangular room at the given position.
func (s *Stamper) stampRectangle(x, y, w, h int, tileData []uint32) error {
	return FillRect(tileData, x, y, w, h, s.width, s.height, uint32(TileFloor))
}

// stampOval stamps an oval-shaped room at the given position.
// Uses a simple ellipse rasterization algorithm.
func (s *Stamper) stampOval(x, y, w, h int, tileData []uint32) error {
	centerX := x + w/2
	centerY := y + h/2
	radiusX := w / 2
	radiusY := h / 2

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			px := x + dx - centerX
			py := y + dy - centerY

			// Check if point is inside ellipse: (px/rx)^2 + (py/ry)^2 <= 1
			if radiusX > 0 && radiusY > 0 {
				normalizedX := float64(px*px) / float64(radiusX*radiusX)
				normalizedY := float64(py*py) / float64(radiusY*radiusY)

				if normalizedX+normalizedY <= 1.0 {
					if err := SetTile(tileData, x+dx, y+dy, s.width, s.height, uint32(TileFloor)); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// stampCross stamps a cross-shaped room at the given position.
// Useful for hub rooms or intersections.
func (s *Stamper) stampCross(x, y, w, h int, tileData []uint32) error {
	// Horizontal bar
	barWidth := w
	barHeight := h / 3
	if barHeight < 1 {
		barHeight = 1
	}

	yOffset := (h - barHeight) / 2
	if err := FillRect(tileData, x, y+yOffset, barWidth, barHeight, s.width, s.height, uint32(TileFloor)); err != nil {
		return err
	}

	// Vertical bar
	barWidth = w / 3
	if barWidth < 1 {
		barWidth = 1
	}
	barHeight = h

	xOffset := (w - barWidth) / 2
	if err := FillRect(tileData, x+xOffset, y, barWidth, barHeight, s.width, s.height, uint32(TileFloor)); err != nil {
		return err
	}

	return nil
}

// stampLShape stamps an L-shaped room at the given position.
// Useful for corner rooms or irregular spaces.
func (s *Stamper) stampLShape(x, y, w, h int, tileData []uint32) error {
	// Horizontal section (bottom)
	horizontalHeight := h / 2
	if err := FillRect(tileData, x, y+h-horizontalHeight, w, horizontalHeight, s.width, s.height, uint32(TileFloor)); err != nil {
		return err
	}

	// Vertical section (left)
	verticalWidth := w / 2
	if err := FillRect(tileData, x, y, verticalWidth, h-horizontalHeight, s.width, s.height, uint32(TileFloor)); err != nil {
		return err
	}

	return nil
}

// StampShape stamps a room using a specific shape identifier.
// This allows for more complex room shapes beyond the default rectangular footprint.
func (s *Stamper) StampShape(shape string, x, y, w, h int, tileData []uint32) error {
	switch shape {
	case "rect", "rectangle":
		return s.stampRectangle(x, y, w, h, tileData)
	case "oval", "ellipse":
		return s.stampOval(x, y, w, h, tileData)
	case "cross", "+":
		return s.stampCross(x, y, w, h, tileData)
	case "L", "l-shape":
		return s.stampLShape(x, y, w, h, tileData)
	default:
		return fmt.Errorf("unsupported room shape: %s", shape)
	}
}
