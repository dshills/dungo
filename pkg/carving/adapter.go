package carving

import (
	"github.com/dshills/dungo/pkg/graph"
)

// GraphAdapter adapts graph.Graph to the carving.Graph interface.
type GraphAdapter struct {
	rooms      map[string]*graph.Room
	connectors map[string]*graph.Connector
}

// NewGraphAdapter creates a new graph adapter.
func NewGraphAdapter(rooms map[string]*graph.Room, connectors map[string]*graph.Connector) *GraphAdapter {
	return &GraphAdapter{
		rooms:      rooms,
		connectors: connectors,
	}
}

// GetRoom retrieves a room by ID.
func (g *GraphAdapter) GetRoom(id string) Room {
	room := g.rooms[id]
	if room == nil {
		return nil
	}
	return &RoomAdapter{room: room}
}

// GetConnector retrieves a connector by ID.
func (g *GraphAdapter) GetConnector(id string) Connector {
	conn := g.connectors[id]
	if conn == nil {
		return nil
	}
	return &ConnectorAdapter{conn: conn}
}

// GetRoomIDs returns all room IDs.
func (g *GraphAdapter) GetRoomIDs() []string {
	ids := make([]string, 0, len(g.rooms))
	for id := range g.rooms {
		ids = append(ids, id)
	}
	return ids
}

// GetConnectorIDs returns all connector IDs.
func (g *GraphAdapter) GetConnectorIDs() []string {
	ids := make([]string, 0, len(g.connectors))
	for id := range g.connectors {
		ids = append(ids, id)
	}
	return ids
}

// RoomAdapter adapts graph.Room to the carving.Room interface.
type RoomAdapter struct {
	room *graph.Room
}

// GetID returns the room ID.
func (r *RoomAdapter) GetID() string {
	return r.room.ID
}

// GetSize returns the room size.
func (r *RoomAdapter) GetSize() RoomSize {
	return RoomSize(r.room.Size)
}

// ConnectorAdapter adapts graph.Connector to the carving.Connector interface.
type ConnectorAdapter struct {
	conn *graph.Connector
}

// GetID returns the connector ID.
func (c *ConnectorAdapter) GetID() string {
	return c.conn.ID
}

// GetFrom returns the from room ID.
func (c *ConnectorAdapter) GetFrom() string {
	return c.conn.From
}

// GetTo returns the to room ID.
func (c *ConnectorAdapter) GetTo() string {
	return c.conn.To
}

// GetType returns the connector type.
func (c *ConnectorAdapter) GetType() ConnectorType {
	return ConnectorType(c.conn.Type)
}

// GetGate returns the gate information.
func (c *ConnectorAdapter) GetGate() *Gate {
	if c.conn.Gate == nil {
		return nil
	}
	return &Gate{
		Type:  c.conn.Gate.Type,
		Value: c.conn.Gate.Value,
	}
}
