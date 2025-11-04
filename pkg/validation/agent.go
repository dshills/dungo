package validation

import (
	"fmt"

	"github.com/dshills/dungo/pkg/graph"
)

// Agent represents a simulated player exploring the dungeon.
// It tracks discovered rooms, acquired capabilities, and visited rooms
// to verify that critical objectives (boss, keys) are findable.
type Agent struct {
	currentRoom  string
	discovered   map[string]bool            // Rooms player knows about
	visited      map[string]bool            // Rooms player has entered
	capabilities map[string]map[string]bool // Type -> Value -> has it
	path         []string                   // History of movement
}

// NewAgent creates a new agent starting at the given room.
func NewAgent(startRoomID string) *Agent {
	return &Agent{
		currentRoom:  startRoomID,
		discovered:   map[string]bool{startRoomID: true},
		visited:      map[string]bool{startRoomID: true},
		capabilities: make(map[string]map[string]bool),
		path:         []string{startRoomID},
	}
}

// CurrentRoom returns the room the agent is currently in.
func (a *Agent) CurrentRoom() string {
	return a.currentRoom
}

// Path returns the sequence of rooms visited.
func (a *Agent) Path() []string {
	return append([]string(nil), a.path...)
}

// HasCapability checks if the agent has acquired a specific capability.
func (a *Agent) HasCapability(capType, value string) bool {
	if typeMap, exists := a.capabilities[capType]; exists {
		return typeMap[value]
	}
	return false
}

// AddCapability grants a capability to the agent.
func (a *Agent) AddCapability(capType, value string) {
	if _, exists := a.capabilities[capType]; !exists {
		a.capabilities[capType] = make(map[string]bool)
	}
	a.capabilities[capType][value] = true
}

// CanTraverse checks if the agent can traverse a connector given current capabilities.
// Returns true if all requirements are met (no gate, or gate satisfied).
func (a *Agent) CanTraverse(conn *graph.Connector) bool {
	if conn.Gate == nil {
		return true
	}

	// Check if agent has the required capability
	return a.HasCapability(conn.Gate.Type, conn.Gate.Value)
}

// Move attempts to move the agent through a connector to the target room.
// Returns error if the move is not valid (can't traverse, invalid connector).
func (a *Agent) Move(g *graph.Graph, conn *graph.Connector, targetRoom *graph.Room) error {
	// Verify connector connects current room to target
	if conn.From != a.currentRoom && conn.To != a.currentRoom {
		return fmt.Errorf("connector %s does not connect to current room %s", conn.ID, a.currentRoom)
	}

	// Check direction
	var destID string
	if conn.From == a.currentRoom {
		destID = conn.To
	} else if conn.To == a.currentRoom && conn.Bidirectional {
		destID = conn.From
	} else {
		return fmt.Errorf("cannot traverse one-way connector %s in reverse", conn.ID)
	}

	if destID != targetRoom.ID {
		return fmt.Errorf("connector leads to %s, not target %s", destID, targetRoom.ID)
	}

	// Check if agent can traverse (gate requirements)
	if !a.CanTraverse(conn) {
		return fmt.Errorf("cannot traverse connector %s: missing capability %s=%s",
			conn.ID, conn.Gate.Type, conn.Gate.Value)
	}

	// Move to new room
	a.currentRoom = targetRoom.ID
	a.visited[targetRoom.ID] = true
	a.discovered[targetRoom.ID] = true
	a.path = append(a.path, targetRoom.ID)

	// Acquire capabilities provided by the room
	for _, cap := range targetRoom.Provides {
		a.AddCapability(cap.Type, cap.Value)
	}

	return nil
}

// ExploreResult contains the result of an exploration attempt.
type ExploreResult struct {
	Found      bool     // Whether the target was found
	Path       []string // Path taken to target (if found)
	Reachable  bool     // Whether target is theoretically reachable (not blocked by unsatisfiable gates)
	PathLength int      // Length of path (if found)
}

// FindPath attempts to find a path from the current room to a target room archetype.
// Uses breadth-first search with capability tracking to find the shortest valid path.
// Returns ExploreResult indicating if target was found and the path taken.
func (a *Agent) FindPath(g *graph.Graph, targetArchetype graph.RoomArchetype) ExploreResult {
	// Find target room
	var targetRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == targetArchetype {
			targetRoom = room
			break
		}
	}

	if targetRoom == nil {
		return ExploreResult{Found: false, Reachable: false}
	}

	// If already at target, we're done
	if a.currentRoom == targetRoom.ID {
		return ExploreResult{
			Found:      true,
			Path:       []string{a.currentRoom},
			Reachable:  true,
			PathLength: 0,
		}
	}

	// BFS with capability state tracking
	type searchState struct {
		roomID       string
		capabilities map[string]map[string]bool
		path         []string
	}

	// Initialize queue with current state
	startState := searchState{
		roomID:       a.currentRoom,
		capabilities: copyCapabilities(a.capabilities),
		path:         []string{a.currentRoom},
	}

	queue := []searchState{startState}
	visited := make(map[string]bool)
	visited[a.currentRoom] = true

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		// Get current room
		currentRoom := g.Rooms[state.roomID]
		if currentRoom == nil {
			continue
		}

		// Collect capabilities from current room
		for _, cap := range currentRoom.Provides {
			if _, exists := state.capabilities[cap.Type]; !exists {
				state.capabilities[cap.Type] = make(map[string]bool)
			}
			state.capabilities[cap.Type][cap.Value] = true
		}

		// Try all connectors from this room
		for _, conn := range g.Connectors {
			var nextRoomID string

			// Check if connector is from current room
			if conn.From == state.roomID {
				nextRoomID = conn.To
			} else if conn.To == state.roomID && conn.Bidirectional {
				nextRoomID = conn.From
			} else {
				continue // Not connected to current room
			}

			// Check gate requirements
			if conn.Gate != nil {
				capMap, exists := state.capabilities[conn.Gate.Type]
				if !exists || !capMap[conn.Gate.Value] {
					continue // Can't traverse this connector yet
				}
			}

			// Skip if already visited in this search
			if visited[nextRoomID] {
				continue
			}

			visited[nextRoomID] = true

			// Create new state for next room
			nextPath := append(append([]string(nil), state.path...), nextRoomID)
			nextState := searchState{
				roomID:       nextRoomID,
				capabilities: copyCapabilities(state.capabilities),
				path:         nextPath,
			}

			// Check if we reached target
			if nextRoomID == targetRoom.ID {
				return ExploreResult{
					Found:      true,
					Path:       nextPath,
					Reachable:  true,
					PathLength: len(nextPath) - 1,
				}
			}

			// Add to queue for further exploration
			queue = append(queue, nextState)
		}
	}

	// Target not found
	return ExploreResult{
		Found:      false,
		Reachable:  false, // Could not find any path
		PathLength: -1,
	}
}

// VerifyBossFindable checks if the boss room is reachable from the start.
// Returns true if a valid path exists, false otherwise.
func VerifyBossFindable(g *graph.Graph) (bool, []string, error) {
	// Find start room
	var startRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startRoom = room
			break
		}
	}

	if startRoom == nil {
		return false, nil, fmt.Errorf("no start room found in graph")
	}

	// Create agent at start
	agent := NewAgent(startRoom.ID)

	// Try to find path to boss
	result := agent.FindPath(g, graph.ArchetypeBoss)

	return result.Found, result.Path, nil
}

// VerifyKeyFindable checks if a specific key is reachable from the start.
// Returns true if the key room can be reached, false otherwise.
func VerifyKeyFindable(g *graph.Graph, keyType, keyValue string) (bool, []string, error) {
	// Find start room
	var startRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startRoom = room
			break
		}
	}

	if startRoom == nil {
		return false, nil, fmt.Errorf("no start room found in graph")
	}

	// Find room that provides the key
	var keyRoom *graph.Room
	for _, room := range g.Rooms {
		for _, cap := range room.Provides {
			if cap.Type == keyType && cap.Value == keyValue {
				keyRoom = room
				break
			}
		}
		if keyRoom != nil {
			break
		}
	}

	if keyRoom == nil {
		return false, nil, fmt.Errorf("no room provides %s=%s", keyType, keyValue)
	}

	// BFS to find path to key room
	// Note: Could use Agent-based pathfinding here in future for more complex scenarios
	type searchState struct {
		roomID       string
		capabilities map[string]map[string]bool
		path         []string
	}

	startState := searchState{
		roomID:       startRoom.ID,
		capabilities: make(map[string]map[string]bool),
		path:         []string{startRoom.ID},
	}

	queue := []searchState{startState}
	visited := make(map[string]bool)
	visited[startRoom.ID] = true

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		// Get current room
		currentRoom := g.Rooms[state.roomID]
		if currentRoom == nil {
			continue
		}

		// Collect capabilities from current room
		for _, cap := range currentRoom.Provides {
			if _, exists := state.capabilities[cap.Type]; !exists {
				state.capabilities[cap.Type] = make(map[string]bool)
			}
			state.capabilities[cap.Type][cap.Value] = true
		}

		// Check if this is the key room
		if state.roomID == keyRoom.ID {
			return true, state.path, nil
		}

		// Try all connectors from this room
		for _, conn := range g.Connectors {
			var nextRoomID string

			if conn.From == state.roomID {
				nextRoomID = conn.To
			} else if conn.To == state.roomID && conn.Bidirectional {
				nextRoomID = conn.From
			} else {
				continue
			}

			// Check gate requirements
			if conn.Gate != nil {
				capMap, exists := state.capabilities[conn.Gate.Type]
				if !exists || !capMap[conn.Gate.Value] {
					continue
				}
			}

			if visited[nextRoomID] {
				continue
			}

			visited[nextRoomID] = true

			nextPath := append(append([]string(nil), state.path...), nextRoomID)
			nextState := searchState{
				roomID:       nextRoomID,
				capabilities: copyCapabilities(state.capabilities),
				path:         nextPath,
			}

			queue = append(queue, nextState)
		}
	}

	return false, nil, nil
}

// copyCapabilities creates a deep copy of the capabilities map.
func copyCapabilities(caps map[string]map[string]bool) map[string]map[string]bool {
	result := make(map[string]map[string]bool)
	for capType, valueMap := range caps {
		result[capType] = make(map[string]bool)
		for value, has := range valueMap {
			result[capType][value] = has
		}
	}
	return result
}

// SimulateFullExploration simulates a full dungeon exploration.
// The agent explores all reachable rooms, collecting capabilities along the way.
// Returns statistics about the exploration.
type ExplorationStats struct {
	RoomsVisited     int
	RoomsReachable   int
	BossReached      bool
	KeysCollected    int
	SecretsFound     int
	PathToCompletion []string
}

// SimulateExploration performs a simulated exploration of the dungeon.
// Uses greedy BFS to explore as much as possible with capability collection.
func SimulateExploration(g *graph.Graph) (*ExplorationStats, error) {
	// Find start room
	var startRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startRoom = room
			break
		}
	}

	if startRoom == nil {
		return nil, fmt.Errorf("no start room found")
	}

	agent := NewAgent(startRoom.ID)
	stats := &ExplorationStats{}

	// BFS to explore all reachable rooms
	type searchState struct {
		roomID       string
		capabilities map[string]map[string]bool
	}

	queue := []searchState{{
		roomID:       startRoom.ID,
		capabilities: make(map[string]map[string]bool),
	}}

	visited := make(map[string]bool)
	visited[startRoom.ID] = true
	reachable := make(map[string]bool)
	reachable[startRoom.ID] = true

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		currentRoom := g.Rooms[state.roomID]
		if currentRoom == nil {
			continue
		}

		// Collect capabilities
		for _, cap := range currentRoom.Provides {
			if _, exists := state.capabilities[cap.Type]; !exists {
				state.capabilities[cap.Type] = make(map[string]bool)
			}
			state.capabilities[cap.Type][cap.Value] = true
			stats.KeysCollected++
		}

		// Count room types
		if currentRoom.Archetype == graph.ArchetypeBoss {
			stats.BossReached = true
		}
		if currentRoom.Archetype == graph.ArchetypeSecret {
			stats.SecretsFound++
		}

		// Try all connectors
		for _, conn := range g.Connectors {
			var nextRoomID string

			if conn.From == state.roomID {
				nextRoomID = conn.To
			} else if conn.To == state.roomID && conn.Bidirectional {
				nextRoomID = conn.From
			} else {
				continue
			}

			// Check gate
			if conn.Gate != nil {
				capMap, exists := state.capabilities[conn.Gate.Type]
				if !exists || !capMap[conn.Gate.Value] {
					continue
				}
			}

			reachable[nextRoomID] = true

			if visited[nextRoomID] {
				continue
			}

			visited[nextRoomID] = true

			queue = append(queue, searchState{
				roomID:       nextRoomID,
				capabilities: copyCapabilities(state.capabilities),
			})
		}
	}

	stats.RoomsVisited = len(visited)
	stats.RoomsReachable = len(reachable)

	// Find path to boss if reached
	if stats.BossReached {
		result := agent.FindPath(g, graph.ArchetypeBoss)
		if result.Found {
			stats.PathToCompletion = result.Path
		}
	}

	return stats, nil
}
