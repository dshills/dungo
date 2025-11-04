package graph

import (
	"fmt"
)

// Graph represents the complete Abstract Dungeon Graph container.
type Graph struct {
	Rooms      map[string]*Room
	Connectors map[string]*Connector
	Adjacency  map[string][]string // Adjacency list for pathfinding
	Seed       uint64
	Metadata   map[string]interface{}
}

// NewGraph creates a new empty graph with the given seed.
func NewGraph(seed uint64) *Graph {
	return &Graph{
		Rooms:      make(map[string]*Room),
		Connectors: make(map[string]*Connector),
		Adjacency:  make(map[string][]string),
		Seed:       seed,
		Metadata:   make(map[string]interface{}),
	}
}

// AddRoom adds a room to the graph after validation and updates indices.
func (g *Graph) AddRoom(room *Room) error {
	if room == nil {
		return fmt.Errorf("cannot add nil room")
	}

	// Validate the room
	if err := room.Validate(); err != nil {
		return fmt.Errorf("room validation failed: %w", err)
	}

	// Check for duplicate ID
	if _, exists := g.Rooms[room.ID]; exists {
		return fmt.Errorf("room with ID %s already exists", room.ID)
	}

	// Add room to map and initialize adjacency list
	g.Rooms[room.ID] = room
	if g.Adjacency[room.ID] == nil {
		g.Adjacency[room.ID] = []string{}
	}

	return nil
}

// AddConnector adds a connector to the graph after validation and updates adjacency.
func (g *Graph) AddConnector(conn *Connector) error {
	if conn == nil {
		return fmt.Errorf("cannot add nil connector")
	}

	// Validate the connector
	if err := conn.Validate(); err != nil {
		return fmt.Errorf("connector validation failed: %w", err)
	}

	// Check that From and To rooms exist
	if _, exists := g.Rooms[conn.From]; !exists {
		return fmt.Errorf("connector %s: From room %s does not exist", conn.ID, conn.From)
	}
	if _, exists := g.Rooms[conn.To]; !exists {
		return fmt.Errorf("connector %s: To room %s does not exist", conn.ID, conn.To)
	}

	// Check for duplicate connector ID
	if _, exists := g.Connectors[conn.ID]; exists {
		return fmt.Errorf("connector with ID %s already exists", conn.ID)
	}

	// Add connector
	g.Connectors[conn.ID] = conn

	// Update adjacency list
	g.Adjacency[conn.From] = append(g.Adjacency[conn.From], conn.To)
	if conn.Bidirectional {
		g.Adjacency[conn.To] = append(g.Adjacency[conn.To], conn.From)
	}

	return nil
}

// RemoveRoom removes a room and all its connected edges from the graph.
func (g *Graph) RemoveRoom(id string) error {
	if _, exists := g.Rooms[id]; !exists {
		return fmt.Errorf("room %s does not exist", id)
	}

	// Remove all connectors involving this room
	connectorsToRemove := []string{}
	for connID, conn := range g.Connectors {
		if conn.From == id || conn.To == id {
			connectorsToRemove = append(connectorsToRemove, connID)
		}
	}

	// Remove connectors and update adjacency
	for _, connID := range connectorsToRemove {
		conn := g.Connectors[connID]
		delete(g.Connectors, connID)

		// Remove from adjacency lists
		g.removeFromAdjacency(conn.From, conn.To)
		if conn.Bidirectional {
			g.removeFromAdjacency(conn.To, conn.From)
		}
	}

	// Remove room and its adjacency list
	delete(g.Rooms, id)
	delete(g.Adjacency, id)

	return nil
}

// removeFromAdjacency removes 'to' from the adjacency list of 'from'.
func (g *Graph) removeFromAdjacency(from, to string) {
	adj, exists := g.Adjacency[from]
	if !exists {
		return
	}

	// Filter out the 'to' room
	newAdj := []string{}
	for _, neighbor := range adj {
		if neighbor != to {
			newAdj = append(newAdj, neighbor)
		}
	}
	g.Adjacency[from] = newAdj
}

// GetPath finds the shortest path between two rooms using BFS.
// Returns a slice of room IDs representing the path from 'from' to 'to',
// including both endpoints. Returns an error if no path exists.
func (g *Graph) GetPath(from, to string) ([]string, error) {
	// Check that both rooms exist
	if _, exists := g.Rooms[from]; !exists {
		return nil, fmt.Errorf("room %s does not exist", from)
	}
	if _, exists := g.Rooms[to]; !exists {
		return nil, fmt.Errorf("room %s does not exist", to)
	}

	// Special case: from == to
	if from == to {
		return []string{from}, nil
	}

	// BFS to find shortest path
	queue := []string{from}
	visited := make(map[string]bool)
	parent := make(map[string]string)
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check neighbors
		for _, neighbor := range g.Adjacency[current] {
			if visited[neighbor] {
				continue
			}

			visited[neighbor] = true
			parent[neighbor] = current
			queue = append(queue, neighbor)

			// Found the target
			if neighbor == to {
				// Reconstruct path
				path := []string{}
				for node := to; node != ""; node = parent[node] {
					path = append([]string{node}, path...)
					if node == from {
						break
					}
				}
				return path, nil
			}
		}
	}

	// No path found
	return nil, fmt.Errorf("no path exists from %s to %s", from, to)
}

// IsConnected checks if the graph is a single connected component.
// Returns true if all rooms are reachable from any starting room.
// NOTE: This checks STRONG connectivity (respects edge direction).
// For dungeons with one-way doors, this may return false even if
// the graph is weakly connected.
func (g *Graph) IsConnected() bool {
	if len(g.Rooms) == 0 {
		return true
	}

	// Pick any room as starting point
	var startID string
	for id := range g.Rooms {
		startID = id
		break
	}

	// Get all reachable rooms from start
	reachable := g.GetReachable(startID)

	// Check if all rooms are reachable
	return len(reachable) == len(g.Rooms)
}

// IsWeaklyConnected checks if the graph is weakly connected.
// Returns true if all rooms are reachable when treating all edges as bidirectional.
// This is more appropriate for dungeons with one-way passages.
func (g *Graph) IsWeaklyConnected() bool {
	if len(g.Rooms) == 0 {
		return true
	}

	// Pick any room as starting point
	var startID string
	for id := range g.Rooms {
		startID = id
		break
	}

	// Build undirected adjacency (treat all edges as bidirectional)
	undirectedAdj := make(map[string][]string)
	for from, neighbors := range g.Adjacency {
		for _, to := range neighbors {
			// Add forward edge
			undirectedAdj[from] = append(undirectedAdj[from], to)
			// Add reverse edge
			undirectedAdj[to] = append(undirectedAdj[to], from)
		}
	}

	// BFS with undirected graph
	visited := make(map[string]bool)
	queue := []string{startID}
	visited[startID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range undirectedAdj[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	// Check if all rooms were visited
	return len(visited) == len(g.Rooms)
}

// GetReachable returns all rooms reachable from the given room using BFS.
func (g *Graph) GetReachable(from string) map[string]bool {
	reachable := make(map[string]bool)

	// Check if starting room exists
	if _, exists := g.Rooms[from]; !exists {
		return reachable
	}

	// BFS to find all reachable rooms
	queue := []string{from}
	reachable[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range g.Adjacency[current] {
			if !reachable[neighbor] {
				reachable[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	return reachable
}

// GetCycles detects all cycles in the graph and returns them as a list of paths.
// Each cycle is represented as a slice of room IDs forming the cycle.
func (g *Graph) GetCycles() [][]string {
	cycles := [][]string{}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	parent := make(map[string]string)

	// DFS helper function to detect cycles
	var dfs func(string) []string
	dfs = func(node string) []string {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range g.Adjacency[node] {
			// Skip back edges to immediate parent in undirected graphs
			if parent[node] == neighbor {
				continue
			}

			if !visited[neighbor] {
				parent[neighbor] = node
				if cycle := dfs(neighbor); cycle != nil {
					return cycle
				}
			} else if recStack[neighbor] {
				// Found a cycle - reconstruct it
				cycle := []string{neighbor}
				for curr := node; curr != neighbor; curr = parent[curr] {
					cycle = append([]string{curr}, cycle...)
				}
				cycle = append(cycle, neighbor) // Complete the cycle
				return cycle
			}
		}

		recStack[node] = false
		return nil
	}

	// Try DFS from each unvisited node
	for roomID := range g.Rooms {
		if !visited[roomID] {
			if cycle := dfs(roomID); cycle != nil {
				cycles = append(cycles, cycle)
				// Reset for finding more cycles
				visited = make(map[string]bool)
				recStack = make(map[string]bool)
				parent = make(map[string]string)
			}
		}
	}

	return cycles
}
