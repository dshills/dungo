package graph

import "fmt"

// ConnectorType defines the connection mechanism between rooms.
type ConnectorType int

const (
	TypeDoor ConnectorType = iota
	TypeCorridor
	TypeLadder
	TypeTeleporter
	TypeHidden
	TypeOneWay
)

// String returns the string representation of a ConnectorType.
func (c ConnectorType) String() string {
	switch c {
	case TypeDoor:
		return "Door"
	case TypeCorridor:
		return "Corridor"
	case TypeLadder:
		return "Ladder"
	case TypeTeleporter:
		return "Teleporter"
	case TypeHidden:
		return "Hidden"
	case TypeOneWay:
		return "OneWay"
	default:
		return fmt.Sprintf("Unknown(%d)", c)
	}
}

// VisibilityType defines how a connector is discovered.
type VisibilityType int

const (
	VisibilityNormal VisibilityType = iota
	VisibilitySecret
	VisibilityIllusory
)

// String returns the string representation of a VisibilityType.
func (v VisibilityType) String() string {
	switch v {
	case VisibilityNormal:
		return "Normal"
	case VisibilitySecret:
		return "Secret"
	case VisibilityIllusory:
		return "Illusory"
	default:
		return fmt.Sprintf("Unknown(%d)", v)
	}
}

// Gate represents an optional gating requirement for a connector.
type Gate struct {
	Type  string `json:"type"`  // "key", "puzzle", "ability"
	Value string `json:"value"` // Specific gate (e.g., "silver_key", "runes_3")
}

// Connector represents an edge in the Abstract Dungeon Graph.
type Connector struct {
	ID            string         `json:"id"`
	From          string         `json:"from"` // Room ID
	To            string         `json:"to"`   // Room ID
	Type          ConnectorType  `json:"type"`
	Gate          *Gate          `json:"gate,omitempty"`
	Cost          float64        `json:"cost"`       // Pathfinding weight (1.0 = normal)
	Visibility    VisibilityType `json:"visibility"` // Discovery mechanism
	Bidirectional bool           `json:"bidirectional"`
}

// Validate checks if the connector data is valid.
// Note: This does not validate that From and To room IDs exist in the graph,
// as that validation is done by the Graph when adding the connector.
func (c *Connector) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}

	if c.From == "" {
		return fmt.Errorf("connector %s: From room ID cannot be empty", c.ID)
	}

	if c.To == "" {
		return fmt.Errorf("connector %s: To room ID cannot be empty", c.ID)
	}

	if c.From == c.To {
		return fmt.Errorf("connector %s: From and To must be different (no self-loops), got %s", c.ID, c.From)
	}

	if c.Cost <= 0.0 {
		return fmt.Errorf("connector %s: Cost must be > 0.0, got %f", c.ID, c.Cost)
	}

	return nil
}

// String returns a human-readable representation of the Connector.
func (c *Connector) String() string {
	direction := "↔"
	if !c.Bidirectional {
		direction = "→"
	}
	gateInfo := ""
	if c.Gate != nil {
		gateInfo = fmt.Sprintf(" [Gate: %s=%s]", c.Gate.Type, c.Gate.Value)
	}
	return fmt.Sprintf("Connector[%s: %s %s %s (%s, Cost=%.2f)%s]",
		c.ID, c.From, direction, c.To, c.Type, c.Cost, gateInfo)
}
