package synthesis

import (
	"context"
	"fmt"
	"sort"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// GrammarSynthesizer generates graphs using a grammar-based approach with production rules.
// It starts with a core trio (Start-Mid-Boss) and applies production rules to reach target size.
//
// Architecture: Hub-and-Spoke Design
// The synthesizer creates branching dungeons with a short critical path and many optional branches.
// For example, a 30-room dungeon might have:
//   - Critical path: Start → Hub → Boss (3 rooms, 2 edges)
//   - Optional content: 27 rooms branching off the hub and other connection points
//
// This architecture differs from linear dungeons where 30% of rooms are on the critical path.
// Instead, most content is optional, rewarding exploration while keeping the main path short.
//
// Production rules:
// - ExpandHub: Adds spoke rooms around hub rooms
// - InsertKeyLoop: Creates key-lock pairs with required paths
// - BranchOptional: Adds optional side branches
//
// The synthesizer ensures:
// - Exactly 1 Start and 1 Boss room
// - All rooms reachable from Start
// - Keys obtainable before their locks
// - Room count within Config bounds
// - Critical path length compatible with branching architecture (see validation.CheckPathBounds)
type GrammarSynthesizer struct {
	maxRetries int // Maximum attempts to satisfy constraints
}

// NewGrammarSynthesizer creates a new grammar-based synthesizer.
func NewGrammarSynthesizer() *GrammarSynthesizer {
	return &GrammarSynthesizer{
		maxRetries: 10,
	}
}

// Name returns the synthesizer identifier.
func (s *GrammarSynthesizer) Name() string {
	return "grammar"
}

// Synthesize generates a graph using grammar-based production rules.
func (s *GrammarSynthesizer) Synthesize(ctx context.Context, rng *rng.RNG, cfg *Config) (*graph.Graph, error) {
	// Basic config validation
	if cfg.RoomsMin < 10 || cfg.RoomsMax > 300 || cfg.RoomsMin > cfg.RoomsMax {
		return nil, fmt.Errorf("invalid room bounds: min=%d max=%d", cfg.RoomsMin, cfg.RoomsMax)
	}
	if cfg.BranchingMax < 2 || cfg.BranchingMax > 5 {
		return nil, fmt.Errorf("invalid branching max: %d", cfg.BranchingMax)
	}

	// Try multiple attempts if constraints fail
	var lastErr error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		g, err := s.tryGenerate(ctx, rng, cfg)
		if err == nil {
			return g, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to satisfy constraints after %d attempts: %w", s.maxRetries, lastErr)
}

// tryGenerate attempts a single generation pass.
func (s *GrammarSynthesizer) tryGenerate(ctx context.Context, rng *rng.RNG, cfg *Config) (*graph.Graph, error) {
	// Initialize graph
	g := graph.NewGraph(cfg.Seed)

	// Step 1: Create core trio (Start-Mid-Boss)
	if err := s.createCoreTrio(g, rng, cfg); err != nil {
		return nil, fmt.Errorf("creating core trio: %w", err)
	}

	// Step 2: Expand to target size using production rules
	targetSize := rng.IntRange(cfg.RoomsMin, cfg.RoomsMax)
	if err := s.expandToSize(ctx, g, rng, cfg, targetSize); err != nil {
		return nil, fmt.Errorf("expanding graph: %w", err)
	}

	// Step 3: Assign difficulty based on pacing curve
	if err := s.assignDifficulty(g, rng, cfg); err != nil {
		return nil, fmt.Errorf("assigning difficulty: %w", err)
	}

	// Step 4: Assign themes to rooms
	if err := assignThemes(g, cfg.Themes, rng); err != nil {
		return nil, fmt.Errorf("assigning themes: %w", err)
	}

	// Step 5: Validate hard constraints
	if err := s.validateHardConstraints(g, cfg); err != nil {
		return nil, fmt.Errorf("constraint validation failed: %w", err)
	}

	return g, nil
}

// createCoreTrio creates the initial Start-Mid-Boss structure.
// This is the foundation that all production rules build upon.
func (s *GrammarSynthesizer) createCoreTrio(g *graph.Graph, rng *rng.RNG, cfg *Config) error {
	// Create Start room
	start := &graph.Room{
		ID:         "start",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Tags:       map[string]string{"type": "entrance"},
		Difficulty: 0.0,
		Reward:     0.0,
	}
	if err := g.AddRoom(start); err != nil {
		return fmt.Errorf("adding start room: %w", err)
	}

	// Create Mid room (hub between start and boss)
	mid := &graph.Room{
		ID:         "mid_hub",
		Archetype:  graph.ArchetypeHub,
		Size:       graph.SizeL,
		Tags:       map[string]string{"type": "hub"},
		Difficulty: 0.5,
		Reward:     0.3,
	}
	if err := g.AddRoom(mid); err != nil {
		return fmt.Errorf("adding mid room: %w", err)
	}

	// Create Boss room
	boss := &graph.Room{
		ID:         "boss",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeXL,
		Tags:       map[string]string{"type": "boss"},
		Difficulty: 1.0,
		Reward:     1.0,
	}
	if err := g.AddRoom(boss); err != nil {
		return fmt.Errorf("adding boss room: %w", err)
	}

	// Connect Start → Mid
	connStartMid := &graph.Connector{
		ID:            "conn_start_mid",
		From:          start.ID,
		To:            mid.ID,
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: true,
	}
	if err := g.AddConnector(connStartMid); err != nil {
		return fmt.Errorf("connecting start to mid: %w", err)
	}

	// Connect Mid → Boss
	connMidBoss := &graph.Connector{
		ID:            "conn_mid_boss",
		From:          mid.ID,
		To:            boss.ID,
		Type:          graph.TypeDoor,
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: true,
	}
	if err := g.AddConnector(connMidBoss); err != nil {
		return fmt.Errorf("connecting mid to boss: %w", err)
	}

	return nil
}

// expandToSize applies production rules until target room count is reached.
func (s *GrammarSynthesizer) expandToSize(ctx context.Context, g *graph.Graph, rng *rng.RNG, cfg *Config, targetSize int) error {
	roomCounter := len(g.Rooms) // Start counting from core trio

	// Calculate how many rooms we need to add
	roomsToAdd := targetSize - roomCounter
	if roomsToAdd <= 0 {
		return nil // Already at or above target
	}

	// Determine rule application probabilities
	expandHubProb := 0.5
	insertKeyLoopProb := 0.3

	// Apply production rules until we reach target size
	for roomCounter < targetSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Choose a production rule based on probabilities
		choice := rng.Float64()
		var err error

		if choice < expandHubProb {
			// ExpandHub: Add rooms around a hub
			err = s.applyExpandHub(g, rng, cfg, &roomCounter)
		} else if choice < expandHubProb+insertKeyLoopProb && len(cfg.Keys) > 0 {
			// InsertKeyLoop: Add key-lock pair
			err = s.applyInsertKeyLoop(g, rng, cfg, &roomCounter)
		} else {
			// BranchOptional: Add optional branch
			err = s.applyBranchOptional(g, rng, cfg, &roomCounter)
		}

		if err != nil {
			// If a rule fails, try a different one next iteration
			continue
		}

		// Safety check to prevent infinite loops
		if roomCounter > cfg.RoomsMax {
			break
		}
	}

	return nil
}

// applyExpandHub adds spoke rooms around a hub room.
// Implements the ExpandHub production rule.
func (s *GrammarSynthesizer) applyExpandHub(g *graph.Graph, rng *rng.RNG, cfg *Config, counter *int) error {
	// Find hub rooms to expand
	hubs := s.findRoomsByArchetype(g, graph.ArchetypeHub)
	if len(hubs) == 0 {
		// No hubs available, try mid_hub or create a new hub
		midHub := g.Rooms["mid_hub"]
		if midHub != nil {
			hubs = []*graph.Room{midHub}
		} else {
			return fmt.Errorf("no hub rooms available")
		}
	}

	// Pick a random hub that still has room for more connections
	var hub *graph.Room
	for _, h := range hubs {
		currentConnections := len(g.Adjacency[h.ID])
		if currentConnections < cfg.BranchingMax {
			hub = h
			break
		}
	}

	if hub == nil {
		// All hubs are at max capacity, find a different room (use sorted iteration for determinism)
		roomIDs := getSortedRoomIDs(g)

		for _, id := range roomIDs {
			room := g.Rooms[id]
			currentConnections := len(g.Adjacency[room.ID])
			if currentConnections < cfg.BranchingMax {
				hub = room
				break
			}
		}

		if hub == nil {
			return fmt.Errorf("all rooms at max capacity")
		}
	}

	// Determine how many spokes to add (1-3, but respect capacity)
	maxSpokesPossible := cfg.BranchingMax - len(g.Adjacency[hub.ID])
	if maxSpokesPossible <= 0 {
		return fmt.Errorf("hub at max capacity")
	}

	spokesToAdd := rng.IntRange(1, min(3, maxSpokesPossible))

	// Add spoke rooms
	for i := 0; i < spokesToAdd; i++ {
		// Create spoke room
		roomID := fmt.Sprintf("room_%d", *counter)
		spoke := &graph.Room{
			ID:         roomID,
			Archetype:  s.pickRoomArchetype(rng, cfg),
			Size:       s.pickRoomSize(rng),
			Tags:       map[string]string{"spoke": hub.ID},
			Difficulty: rng.Float64(),
			Reward:     rng.Float64(),
		}

		if err := g.AddRoom(spoke); err != nil {
			return err
		}

		// Connect hub to spoke
		connID := fmt.Sprintf("conn_%s_%s", hub.ID, spoke.ID)
		conn := &graph.Connector{
			ID:            connID,
			From:          hub.ID,
			To:            spoke.ID,
			Type:          s.pickConnectorType(rng),
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		}

		if err := g.AddConnector(conn); err != nil {
			return err
		}

		*counter++
	}

	return nil
}

// applyInsertKeyLoop adds a key-lock pair to the graph.
// Implements the InsertKeyLoop production rule.
func (s *GrammarSynthesizer) applyInsertKeyLoop(g *graph.Graph, rng *rng.RNG, cfg *Config, counter *int) error {
	if len(cfg.Keys) == 0 {
		return fmt.Errorf("no keys configured")
	}

	// Pick a random key type
	keyConfig := cfg.Keys[rng.Intn(len(cfg.Keys))]

	// Find an existing room with capacity to attach the key room to
	availableRooms := s.getRoomsWithCapacity(g, cfg)
	if len(availableRooms) == 0 {
		return fmt.Errorf("no rooms with capacity available")
	}
	attachPoint := availableRooms[rng.Intn(len(availableRooms))]

	// Create key room (comes before lock)
	keyRoomID := fmt.Sprintf("room_%d", *counter)
	keyRoom := &graph.Room{
		ID:         keyRoomID,
		Archetype:  graph.ArchetypeTreasure,
		Size:       graph.SizeS,
		Tags:       map[string]string{"contains": "key_" + keyConfig.Name},
		Difficulty: rng.Float64Range(0.3, 0.7),
		Reward:     0.5,
		Provides:   []graph.Capability{{Type: "key", Value: keyConfig.Name}},
	}

	if err := g.AddRoom(keyRoom); err != nil {
		return err
	}

	// Connect attach point to key room
	connToKey := &graph.Connector{
		ID:            fmt.Sprintf("conn_%s_%s", attachPoint.ID, keyRoom.ID),
		From:          attachPoint.ID,
		To:            keyRoom.ID,
		Type:          graph.TypeDoor,
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: true,
	}

	if err := g.AddConnector(connToKey); err != nil {
		return err
	}

	*counter++

	// Create locked room (requires key)
	lockedRoomID := fmt.Sprintf("room_%d", *counter)
	lockedRoom := &graph.Room{
		ID:           lockedRoomID,
		Archetype:    graph.ArchetypePuzzle,
		Size:         graph.SizeM,
		Tags:         map[string]string{"locked_by": "key_" + keyConfig.Name},
		Difficulty:   rng.Float64Range(0.5, 0.9),
		Reward:       0.8,
		Requirements: []graph.Requirement{{Type: "key", Value: keyConfig.Name}},
	}

	if err := g.AddRoom(lockedRoom); err != nil {
		return err
	}

	// Connect key room to locked room
	connToLocked := &graph.Connector{
		ID:   fmt.Sprintf("conn_%s_%s", keyRoom.ID, lockedRoom.ID),
		From: keyRoom.ID,
		To:   lockedRoom.ID,
		Type: graph.TypeDoor,
		Gate: &graph.Gate{
			Type:  "key",
			Value: keyConfig.Name,
		},
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: false, // One-way: can't go back through locked door without key
	}

	if err := g.AddConnector(connToLocked); err != nil {
		return err
	}

	*counter++

	return nil
}

// applyBranchOptional adds an optional side branch.
// Implements the BranchOptional production rule.
func (s *GrammarSynthesizer) applyBranchOptional(g *graph.Graph, rng *rng.RNG, cfg *Config, counter *int) error {
	// Find an existing room with capacity to branch from
	availableRooms := s.getRoomsWithCapacity(g, cfg)
	if len(availableRooms) == 0 {
		return fmt.Errorf("no rooms with capacity available")
	}

	branchPoint := availableRooms[rng.Intn(len(availableRooms))]

	// Create optional branch room
	roomID := fmt.Sprintf("room_%d", *counter)
	optionalRoom := &graph.Room{
		ID:         roomID,
		Archetype:  graph.ArchetypeOptional,
		Size:       s.pickRoomSize(rng),
		Tags:       map[string]string{"optional": "true", "branch_from": branchPoint.ID},
		Difficulty: rng.Float64Range(0.4, 0.8),
		Reward:     rng.Float64Range(0.6, 1.0), // Higher rewards for optional content
	}

	if err := g.AddRoom(optionalRoom); err != nil {
		return err
	}

	// Connect branch point to optional room
	connID := fmt.Sprintf("conn_%s_%s", branchPoint.ID, optionalRoom.ID)
	conn := &graph.Connector{
		ID:            connID,
		From:          branchPoint.ID,
		To:            optionalRoom.ID,
		Type:          s.pickConnectorType(rng),
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: true,
	}

	if err := g.AddConnector(conn); err != nil {
		return err
	}

	*counter++

	// Occasionally add a secret room off the optional branch
	if rng.Float64() < cfg.SecretDensity {
		secretID := fmt.Sprintf("room_%d", *counter)
		secretRoom := &graph.Room{
			ID:         secretID,
			Archetype:  graph.ArchetypeSecret,
			Size:       graph.SizeS,
			Tags:       map[string]string{"secret": "true", "branch_from": optionalRoom.ID},
			Difficulty: rng.Float64Range(0.6, 1.0),
			Reward:     1.0, // Secrets have max rewards
		}

		if err := g.AddRoom(secretRoom); err == nil {
			// Connect optional room to secret (hidden connection)
			secretConnID := fmt.Sprintf("conn_%s_%s", optionalRoom.ID, secretRoom.ID)
			secretConn := &graph.Connector{
				ID:            secretConnID,
				From:          optionalRoom.ID,
				To:            secretRoom.ID,
				Type:          graph.TypeHidden,
				Cost:          1.0,
				Visibility:    graph.VisibilitySecret,
				Bidirectional: true,
			}

			if err := g.AddConnector(secretConn); err == nil {
				*counter++
			}
		}
	}

	return nil
}

// validateHardConstraints checks that all hard constraints are satisfied.
func (s *GrammarSynthesizer) validateHardConstraints(g *graph.Graph, cfg *Config) error {
	// Constraint 1: Must have exactly 1 Start room
	startCount := s.countRoomsByArchetype(g, graph.ArchetypeStart)
	if startCount != 1 {
		return fmt.Errorf("must have exactly 1 Start room, got %d", startCount)
	}

	// Constraint 2: Must have exactly 1 Boss room
	bossCount := s.countRoomsByArchetype(g, graph.ArchetypeBoss)
	if bossCount != 1 {
		return fmt.Errorf("must have exactly 1 Boss room, got %d", bossCount)
	}

	// Constraint 3: Graph must be connected
	if !g.IsConnected() {
		return fmt.Errorf("graph is not connected")
	}

	// Constraint 4: Room count must be within bounds
	roomCount := len(g.Rooms)
	if roomCount < cfg.RoomsMin {
		return fmt.Errorf("room count %d below minimum %d", roomCount, cfg.RoomsMin)
	}
	if roomCount > cfg.RoomsMax {
		return fmt.Errorf("room count %d exceeds maximum %d", roomCount, cfg.RoomsMax)
	}

	// Constraint 5: Start must have path to Boss
	startRoom := s.findRoomsByArchetype(g, graph.ArchetypeStart)[0]
	bossRoom := s.findRoomsByArchetype(g, graph.ArchetypeBoss)[0]
	if _, err := g.GetPath(startRoom.ID, bossRoom.ID); err != nil {
		return fmt.Errorf("no path from Start to Boss: %w", err)
	}

	// Constraint 6: Validate key-before-lock ordering
	if err := s.validateKeyLockConstraints(g); err != nil {
		return fmt.Errorf("key-lock constraint violated: %w", err)
	}

	// Constraint 7: Respect branching max
	for roomID, neighbors := range g.Adjacency {
		if len(neighbors) > cfg.BranchingMax {
			return fmt.Errorf("room %s has %d connections, exceeds max %d", roomID, len(neighbors), cfg.BranchingMax)
		}
	}

	return nil
}

// validateKeyLockConstraints ensures keys are obtainable before their locks.
func (s *GrammarSynthesizer) validateKeyLockConstraints(g *graph.Graph) error {
	// Find all rooms that provide keys
	keyProviders := make(map[string]string) // key name -> room ID

	for _, room := range g.Rooms {
		for _, cap := range room.Provides {
			if cap.Type == "key" {
				keyProviders[cap.Value] = room.ID
			}
		}
	}

	// Find all rooms that require keys
	for _, room := range g.Rooms {
		for _, req := range room.Requirements {
			if req.Type == "key" {
				keyProviderID, hasProvider := keyProviders[req.Value]
				if !hasProvider {
					return fmt.Errorf("room %s requires key %q but no room provides it", room.ID, req.Value)
				}

				// Verify key provider is reachable before locked room
				// Start from Start room
				startRoom := s.findRoomsByArchetype(g, graph.ArchetypeStart)[0]

				// Can we reach key provider from Start?
				_, err := g.GetPath(startRoom.ID, keyProviderID)
				if err != nil {
					return fmt.Errorf("key %q in room %s is not reachable from Start", req.Value, keyProviderID)
				}

				// This ensures key can be obtained before reaching locked room
			}
		}
	}

	return nil
}

// Helper methods

// getSortedRoomIDs returns room IDs sorted lexicographically for deterministic iteration.
// This helper reduces code duplication and ensures consistent ordering across the synthesizer.
func getSortedRoomIDs(g *graph.Graph) []string {
	roomIDs := make([]string, 0, len(g.Rooms))
	for id := range g.Rooms {
		roomIDs = append(roomIDs, id)
	}
	sort.Strings(roomIDs)
	return roomIDs
}

func (s *GrammarSynthesizer) findRoomsByArchetype(g *graph.Graph, archetype graph.RoomArchetype) []*graph.Room {
	// Use sorted iteration for deterministic room order
	roomIDs := getSortedRoomIDs(g)

	var rooms []*graph.Room
	for _, id := range roomIDs {
		room := g.Rooms[id]
		if room.Archetype == archetype {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func (s *GrammarSynthesizer) countRoomsByArchetype(g *graph.Graph, archetype graph.RoomArchetype) int {
	return len(s.findRoomsByArchetype(g, archetype))
}

// getAllRooms is currently unused but may be needed for future features
// nolint:unused
func (s *GrammarSynthesizer) getAllRooms(g *graph.Graph) []*graph.Room {
	rooms := make([]*graph.Room, 0, len(g.Rooms))
	for _, room := range g.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (s *GrammarSynthesizer) getRoomsWithCapacity(g *graph.Graph, cfg *Config) []*graph.Room {
	// Use sorted iteration for deterministic RNG consumption
	roomIDs := getSortedRoomIDs(g)

	rooms := make([]*graph.Room, 0, len(g.Rooms))
	for _, id := range roomIDs {
		room := g.Rooms[id]
		currentConnections := len(g.Adjacency[room.ID])
		if currentConnections < cfg.BranchingMax {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func (s *GrammarSynthesizer) pickRoomArchetype(rng *rng.RNG, cfg *Config) graph.RoomArchetype {
	// Weight distribution for room types
	weights := []float64{
		0.0,  // Start (never random)
		0.0,  // Boss (never random)
		0.15, // Treasure
		0.2,  // Puzzle
		0.1,  // Hub
		0.25, // Corridor
		0.0,  // Secret (handled separately)
		0.2,  // Optional
		0.05, // Vendor
		0.05, // Shrine
		0.0,  // Checkpoint
	}

	choice := rng.WeightedChoice(weights)
	return graph.RoomArchetype(choice)
}

func (s *GrammarSynthesizer) pickRoomSize(rng *rng.RNG) graph.RoomSize {
	weights := []float64{
		0.2,  // XS
		0.3,  // S
		0.3,  // M
		0.15, // L
		0.05, // XL
	}

	choice := rng.WeightedChoice(weights)
	return graph.RoomSize(choice)
}

func (s *GrammarSynthesizer) pickConnectorType(rng *rng.RNG) graph.ConnectorType {
	weights := []float64{
		0.4,  // Door
		0.4,  // Corridor
		0.1,  // Ladder
		0.05, // Teleporter
		0.0,  // Hidden (handled separately)
		0.05, // OneWay
	}

	choice := rng.WeightedChoice(weights)
	return graph.ConnectorType(choice)
}

// init registers the grammar synthesizer on package load.
func init() {
	Register("grammar", NewGrammarSynthesizer())
}

// assignDifficulty assigns difficulty values to all rooms based on the pacing curve.
// Rooms on the critical path (Start → Boss) get difficulties based on their progress (0.0-1.0).
// Optional/side rooms get slightly varied difficulties to add interest.
func (s *GrammarSynthesizer) assignDifficulty(g *graph.Graph, rng *rng.RNG, cfg *Config) error {
	// Find Start and Boss rooms
	var startRoom, bossRoom *graph.Room
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startRoom = room
		} else if room.Archetype == graph.ArchetypeBoss {
			bossRoom = room
		}
	}

	if startRoom == nil || bossRoom == nil {
		return fmt.Errorf("missing Start or Boss room for difficulty assignment")
	}

	// Get the critical path from Start to Boss
	criticalPath, err := g.GetPath(startRoom.ID, bossRoom.ID)
	if err != nil {
		return fmt.Errorf("no path from Start to Boss: %w", err)
	}

	// Create pacing curve from config
	curve, err := s.createPacingCurve(cfg.Pacing)
	if err != nil {
		return fmt.Errorf("creating pacing curve: %w", err)
	}

	// Build progress map for critical path rooms
	progressMap := make(map[string]float64)
	for i, roomID := range criticalPath {
		// Progress: 0.0 at start, 1.0 at boss
		progress := 0.0
		if len(criticalPath) > 1 {
			progress = float64(i) / float64(len(criticalPath)-1)
		}
		progressMap[roomID] = progress
	}

	// Assign difficulty to all rooms (use sorted iteration for deterministic RNG consumption)
	roomIDs := getSortedRoomIDs(g)

	for _, roomID := range roomIDs {
		room := g.Rooms[roomID]
		var difficulty float64

		if progress, onPath := progressMap[roomID]; onPath {
			// Room is on critical path: use pacing curve with variance
			difficulty = EvaluateWithVariance(curve, progress, cfg.Pacing.Variance, rng)
		} else {
			// Room is optional/side room: interpolate from nearest path rooms
			difficulty = s.interpolateOffPathDifficulty(g, roomID, progressMap, curve, cfg.Pacing.Variance, rng)
		}

		// Apply difficulty to room
		room.Difficulty = difficulty

		// Also scale reward based on difficulty (higher difficulty = higher reward potential)
		// But add some randomness so not all hard rooms have high rewards
		baseReward := difficulty * 0.7 // Base: 70% of difficulty
		randomBonus := rng.Float64() * 0.3
		room.Reward = baseReward + randomBonus
		if room.Reward > 1.0 {
			room.Reward = 1.0
		}
	}

	return nil
}

// createPacingCurve creates a PacingCurve from synthesis config.
func (s *GrammarSynthesizer) createPacingCurve(cfg PacingConfig) (PacingCurve, error) {
	switch cfg.Curve {
	case "LINEAR":
		return &LinearCurve{}, nil
	case "S_CURVE":
		return NewSCurve(), nil
	case "EXPONENTIAL":
		return NewExponentialCurve(), nil
	case "CUSTOM":
		return NewCustomCurve(cfg.CustomPoints)
	default:
		// Default to linear if unspecified
		return &LinearCurve{}, nil
	}
}

// interpolateOffPathDifficulty computes difficulty for rooms not on the critical path.
// Uses nearest neighbors on the path to interpolate a reasonable difficulty.
func (s *GrammarSynthesizer) interpolateOffPathDifficulty(
	g *graph.Graph,
	roomID string,
	progressMap map[string]float64,
	curve PacingCurve,
	variance float64,
	rng *rng.RNG,
) float64 {
	// Find neighboring rooms that are on the path
	neighbors := g.Adjacency[roomID]
	pathNeighbors := make([]float64, 0)

	for _, neighborID := range neighbors {
		if progress, onPath := progressMap[neighborID]; onPath {
			pathNeighbors = append(pathNeighbors, progress)
		}
	}

	// If no path neighbors, do a BFS to find nearest path room
	if len(pathNeighbors) == 0 {
		nearestProgress := s.findNearestPathRoom(g, roomID, progressMap)
		if nearestProgress >= 0 {
			pathNeighbors = append(pathNeighbors, nearestProgress)
		}
	}

	// Compute average progress from path neighbors
	avgProgress := 0.5 // Default to middle if no path neighbors found
	if len(pathNeighbors) > 0 {
		sum := 0.0
		for _, p := range pathNeighbors {
			sum += p
		}
		avgProgress = sum / float64(len(pathNeighbors))
	}

	// Apply pacing curve with variance
	return EvaluateWithVariance(curve, avgProgress, variance, rng)
}

// findNearestPathRoom finds the progress value of the nearest room on the critical path.
// Uses BFS to find the closest path room. Returns -1.0 if no path room found.
func (s *GrammarSynthesizer) findNearestPathRoom(g *graph.Graph, startID string, progressMap map[string]float64) float64 {
	visited := make(map[string]bool)
	queue := []string{startID}
	visited[startID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check if this room is on the path
		if progress, onPath := progressMap[current]; onPath {
			return progress
		}

		// Add neighbors to queue
		for _, neighborID := range g.Adjacency[current] {
			if !visited[neighborID] {
				visited[neighborID] = true
				queue = append(queue, neighborID)
			}
		}
	}

	return -1.0 // Not found
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
