package content

import (
	"fmt"
	"sort"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// placeRequiredKeys finds all key-locked connectors and places keys before locks.
// This ensures that keys are always reachable before their corresponding locks.
//
// Algorithm:
//  1. Find all connectors with gate.type == "key"
//  2. For each key, find the path from Start to the locked connector
//  3. Place the key in a room on that path, before the lock
//  4. Mark the loot as Required=true
func placeRequiredKeys(g *graph.Graph, content *Content, rng *rng.RNG) error {
	// Find start room
	startRoom := findStartRoom(g)
	if startRoom == "" {
		return fmt.Errorf("no start room found in graph")
	}

	lootID := 0
	keysSeen := make(map[string]bool) // Track which keys we've placed

	// Sort connector IDs for deterministic iteration
	connIDs := make([]string, 0, len(g.Connectors))
	for id := range g.Connectors {
		connIDs = append(connIDs, id)
	}
	sort.Strings(connIDs)

	// Find all key-locked connectors
	for _, connID := range connIDs {
		conn := g.Connectors[connID]
		if conn.Gate == nil || conn.Gate.Type != "key" {
			continue
		}

		keyName := conn.Gate.Value

		// Skip if we've already placed this key
		if keysSeen[keyName] {
			continue
		}

		// Find rooms on path from start to the locked door
		// Place key in one of these rooms (before the lock)
		candidateRooms := findRoomsBeforeLock(g, startRoom, conn.From, conn.To)

		if len(candidateRooms) == 0 {
			// If no path found, place in a random accessible room
			// This handles edge cases in graph structure
			candidateRooms = findAccessibleRooms(g, startRoom, conn.From)
		}

		if len(candidateRooms) == 0 {
			return fmt.Errorf("no suitable room found for key %s", keyName)
		}

		// Select a room from candidates (prefer high-reward rooms for keys)
		roomID := selectKeyPlacementRoom(g, candidateRooms, rng)

		// Place the key
		key := Loot{
			ID:       fmt.Sprintf("loot_%d", lootID),
			RoomID:   roomID,
			Position: Point{X: 0, Y: 0}, // Placeholder - needs layout
			ItemType: fmt.Sprintf("key_%s", keyName),
			Value:    1,
			Required: true,
		}

		if err := key.Validate(); err != nil {
			return fmt.Errorf("invalid key loot: %w", err)
		}

		content.Loot = append(content.Loot, key)
		keysSeen[keyName] = true
		lootID++
	}

	return nil
}

// distributeLoot places treasure loot based on room.Reward values.
// Higher reward rooms get more valuable loot.
//
// Algorithm:
//  1. Calculate total reward budget from lootBudgetBase
//  2. For each room, allocate loot proportional to room.Reward
//  3. Place loot items in eligible rooms
//  4. Skip rooms that shouldn't have loot (Start, corridors, etc.)
func distributeLoot(g *graph.Graph, content *Content, budgetBase int, rng *rng.RNG) error {
	// Sort room IDs for deterministic iteration
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)

	// Calculate total reward weight
	totalReward := 0.0
	eligibleRooms := make([]*graph.Room, 0)

	for _, roomID := range roomIDs {
		room := g.Rooms[roomID]
		if shouldSkipLootPlacement(room) {
			continue
		}
		totalReward += room.Reward
		eligibleRooms = append(eligibleRooms, room)
	}

	if totalReward == 0.0 {
		// No rooms with rewards, skip loot distribution
		return nil
	}

	// Distribute budget proportionally
	lootID := len(content.Loot) // Continue from keys

	for _, room := range eligibleRooms {
		// Calculate loot value for this room
		roomBudget := int((room.Reward / totalReward) * float64(budgetBase))

		if roomBudget == 0 && room.Reward > 0.0 {
			roomBudget = 10 // Minimum loot value
		}

		if roomBudget == 0 {
			continue
		}

		// Determine number of loot items (1-3 based on room size)
		itemCount := 1
		switch room.Size {
		case graph.SizeXS:
			itemCount = 1
		case graph.SizeS:
			itemCount = rng.IntRange(1, 2)
		case graph.SizeM:
			itemCount = rng.IntRange(1, 2)
		case graph.SizeL:
			itemCount = rng.IntRange(2, 3)
		case graph.SizeXL:
			itemCount = rng.IntRange(2, 4)
		}

		// Distribute budget across items
		for i := 0; i < itemCount; i++ {
			itemValue := roomBudget / itemCount

			// Select loot type based on value
			lootType := selectLootType(itemValue, rng)

			loot := Loot{
				ID:       fmt.Sprintf("loot_%d", lootID),
				RoomID:   room.ID,
				Position: Point{X: 0, Y: 0}, // Placeholder - needs layout
				ItemType: lootType,
				Value:    itemValue,
				Required: false,
			}

			if err := loot.Validate(); err != nil {
				return fmt.Errorf("invalid loot: %w", err)
			}

			content.Loot = append(content.Loot, loot)
			lootID++
		}
	}

	return nil
}

// findStartRoom returns the ID of the start room.
func findStartRoom(g *graph.Graph) string {
	for id, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			return id
		}
	}
	return ""
}

// findRoomsBeforeLock finds rooms on the path from start to the lock.
// Returns rooms that are accessible before reaching the locked connector.
func findRoomsBeforeLock(g *graph.Graph, start, lockFrom, lockTo string) []string {
	// Get path from start to the room before the lock
	path, err := g.GetPath(start, lockFrom)
	if err != nil {
		return []string{}
	}

	// Filter out unsuitable rooms (Start, corridors, etc.)
	candidates := make([]string, 0)
	for _, roomID := range path {
		room := g.Rooms[roomID]
		if room == nil {
			continue
		}

		// Skip start and unsuitable room types
		if room.Archetype == graph.ArchetypeStart ||
			room.Archetype == graph.ArchetypeCorridor {
			continue
		}

		candidates = append(candidates, roomID)
	}

	return candidates
}

// findAccessibleRooms finds all rooms reachable from start without going through lockFrom.
func findAccessibleRooms(g *graph.Graph, start, excludeRoom string) []string {
	reachable := make(map[string]bool)
	queue := []string{start}
	reachable[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == excludeRoom {
			continue // Don't traverse past the locked room
		}

		for _, neighbor := range g.Adjacency[current] {
			if !reachable[neighbor] {
				reachable[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	// Convert to slice and filter
	rooms := make([]string, 0)
	for roomID := range reachable {
		room := g.Rooms[roomID]
		if room == nil || room.Archetype == graph.ArchetypeStart ||
			room.Archetype == graph.ArchetypeCorridor {
			continue
		}
		rooms = append(rooms, roomID)
	}

	return rooms
}

// selectKeyPlacementRoom chooses the best room for key placement.
// Prefers rooms with higher reward values (treasure rooms, boss rooms, etc.).
func selectKeyPlacementRoom(g *graph.Graph, candidates []string, rng *rng.RNG) string {
	if len(candidates) == 0 {
		return ""
	}

	if len(candidates) == 1 {
		return candidates[0]
	}

	// Build weights based on room reward
	weights := make([]float64, len(candidates))
	for i, roomID := range candidates {
		room := g.Rooms[roomID]
		if room == nil {
			weights[i] = 0.1
			continue
		}

		// Weight by reward + base weight
		weights[i] = room.Reward + 0.1

		// Bonus weight for treasure rooms
		if room.Archetype == graph.ArchetypeTreasure {
			weights[i] *= 2.0
		}
	}

	index := rng.WeightedChoice(weights)
	if index < 0 || index >= len(candidates) {
		return candidates[0]
	}

	return candidates[index]
}

// shouldSkipLootPlacement determines if a room should not have loot.
func shouldSkipLootPlacement(room *graph.Room) bool {
	switch room.Archetype {
	case graph.ArchetypeStart:
		return true // Start room typically empty
	case graph.ArchetypeCorridor:
		return true // Corridors don't have loot
	default:
		return false
	}
}

// lootTypes maps value ranges to item types.
var lootTypes = []struct {
	name     string
	minValue int
	maxValue int
}{
	{"gold_small", 0, 50},
	{"gold_medium", 40, 150},
	{"gold_large", 100, 300},
	{"potion_health", 50, 100},
	{"potion_mana", 50, 100},
	{"gem", 150, 400},
	{"artifact", 300, 1000},
}

// selectLootType chooses a loot type appropriate for the given value.
func selectLootType(value int, rng *rng.RNG) string {
	eligible := make([]string, 0)
	weights := make([]float64, 0)

	for _, lt := range lootTypes {
		if value >= lt.minValue && value <= lt.maxValue {
			eligible = append(eligible, lt.name)

			// Weight items toward center of their range
			center := (lt.minValue + lt.maxValue) / 2
			dist := float64(value - center)
			if dist < 0 {
				dist = -dist
			}
			rangeSize := float64(lt.maxValue - lt.minValue)
			weight := 1.0 - (dist / rangeSize)
			if weight < 0.1 {
				weight = 0.1
			}
			weights = append(weights, weight)
		}
	}

	if len(eligible) == 0 {
		return "gold_small" // Default fallback
	}

	index := rng.WeightedChoice(weights)
	if index < 0 || index >= len(eligible) {
		return eligible[0]
	}

	return eligible[index]
}
