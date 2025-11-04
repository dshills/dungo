package synthesis

import (
	"context"
	"fmt"

	"github.com/dshills/dungo/pkg/graph"
	"github.com/dshills/dungo/pkg/rng"
)

// TemplateSynthesizer generates graphs by stitching together pre-defined templates.
// This is a simpler alternative to GrammarSynthesizer that trades flexibility for
// predictability. Templates are small graph fragments (3-10 rooms) that are connected
// to form the complete dungeon.
//
// Approach:
//  1. Select a Start template (contains Start room + 2-3 adjacent rooms)
//  2. Select a Boss template (contains Boss room + approach rooms)
//  3. Select and connect Mid templates (corridors, hubs, branches)
//  4. Validate connectivity and constraints
//
// Templates ensure certain spatial patterns and room arrangements that can be
// difficult to achieve with grammar rules alone. This synthesizer is ideal for
// dungeons with specific architectural styles (e.g., symmetrical temples, linear
// gauntlets, hub-and-spoke fortresses).
type TemplateSynthesizer struct {
	templates  map[string][]GraphTemplate // Templates by category
	maxRetries int                        // Maximum generation attempts
}

// GraphTemplate defines a reusable dungeon fragment.
// Templates contain rooms and connectors that can be instantiated and connected.
type GraphTemplate struct {
	Name        string               // Template identifier
	Category    string               // "start", "mid", "boss", "branch"
	Rooms       []*RoomTemplate      // Room definitions
	Connectors  []*ConnectorTemplate // Connection definitions
	Attachments []string             // Room IDs that can connect to other templates
}

// RoomTemplate defines a room within a template.
type RoomTemplate struct {
	LocalID   string              // ID within template (will be prefixed)
	Archetype graph.RoomArchetype // Room type
	Size      graph.RoomSize      // Room size
}

// ConnectorTemplate defines a connection within a template.
type ConnectorTemplate struct {
	LocalFromID   string               // From room (local ID)
	LocalToID     string               // To room (local ID)
	Type          graph.ConnectorType  // Connector type
	Bidirectional bool                 // Two-way connection
	Visibility    graph.VisibilityType // Discovery method
}

// NewTemplateSynthesizer creates a template-based synthesizer with default templates.
func NewTemplateSynthesizer() *TemplateSynthesizer {
	s := &TemplateSynthesizer{
		templates:  make(map[string][]GraphTemplate),
		maxRetries: 10,
	}

	// Register default templates
	s.registerDefaultTemplates()

	return s
}

// Name returns the synthesizer identifier.
func (s *TemplateSynthesizer) Name() string {
	return "template"
}

// Synthesize generates a graph by stitching templates together.
func (s *TemplateSynthesizer) Synthesize(ctx context.Context, rng *rng.RNG, cfg *Config) (*graph.Graph, error) {
	// Basic config validation
	if cfg.RoomsMin < 10 || cfg.RoomsMax > 300 || cfg.RoomsMin > cfg.RoomsMax {
		return nil, fmt.Errorf("invalid room bounds: min=%d max=%d", cfg.RoomsMin, cfg.RoomsMax)
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
func (s *TemplateSynthesizer) tryGenerate(ctx context.Context, rng *rng.RNG, cfg *Config) (*graph.Graph, error) {
	g := graph.NewGraph(cfg.Seed)

	// Track room count and instance counter
	roomCount := 0
	instanceID := 0

	// Step 1: Instantiate Start template
	startTemplate := s.selectTemplate("start", rng)
	startPrefix := fmt.Sprintf("s%d_", instanceID)
	startAttachments, err := s.instantiateTemplate(g, startTemplate, startPrefix, rng)
	if err != nil {
		return nil, fmt.Errorf("instantiating start template: %w", err)
	}
	roomCount += len(startTemplate.Rooms)
	instanceID++

	// Step 2: Instantiate Boss template
	bossTemplate := s.selectTemplate("boss", rng)
	bossPrefix := fmt.Sprintf("b%d_", instanceID)
	_, err = s.instantiateTemplate(g, bossTemplate, bossPrefix, rng)
	if err != nil {
		return nil, fmt.Errorf("instantiating boss template: %w", err)
	}
	roomCount += len(bossTemplate.Rooms)
	instanceID++

	// Step 3: Connect Start to Boss via Mid templates
	targetSize := rng.IntRange(cfg.RoomsMin, cfg.RoomsMax)
	lastAttachment := startAttachments[rng.Intn(len(startAttachments))]

	for roomCount < targetSize {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Select template type based on remaining rooms needed
		var category string
		remainingRooms := targetSize - roomCount
		if remainingRooms < 15 {
			// Nearly done, use small templates or branch
			if rng.Float64() < 0.6 {
				category = "mid"
			} else {
				category = "branch"
			}
		} else {
			// More rooms needed, prefer mid templates
			category = "mid"
		}

		midTemplate := s.selectTemplate(category, rng)
		midPrefix := fmt.Sprintf("m%d_", instanceID)
		_, err := s.instantiateTemplate(g, midTemplate, midPrefix, rng)
		if err != nil {
			// Skip this template if it fails
			instanceID++
			continue
		}

		// Connect to previous attachment point
		firstMidRoom := midPrefix + midTemplate.Attachments[0]
		connID := fmt.Sprintf("conn_%s_%s", lastAttachment, firstMidRoom)
		conn := &graph.Connector{
			ID:            connID,
			From:          lastAttachment,
			To:            firstMidRoom,
			Type:          graph.TypeCorridor,
			Cost:          1.0,
			Visibility:    graph.VisibilityNormal,
			Bidirectional: true,
		}
		if err := g.AddConnector(conn); err != nil {
			// Connection failed, try again
			instanceID++
			continue
		}

		roomCount += len(midTemplate.Rooms)
		lastAttachment = midPrefix + midTemplate.Attachments[len(midTemplate.Attachments)-1]
		instanceID++

		// Safety check
		if roomCount > cfg.RoomsMax {
			break
		}
	}

	// Step 4: Connect final attachment to Boss entrance
	bossEntrance := bossPrefix + bossTemplate.Attachments[0]
	finalConn := &graph.Connector{
		ID:            fmt.Sprintf("conn_%s_%s", lastAttachment, bossEntrance),
		From:          lastAttachment,
		To:            bossEntrance,
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Visibility:    graph.VisibilityNormal,
		Bidirectional: true,
	}
	if err := g.AddConnector(finalConn); err != nil {
		return nil, fmt.Errorf("connecting to boss: %w", err)
	}

	// Step 5: Assign difficulty based on pacing
	if err := assignDifficultyTemplate(g, rng, cfg); err != nil {
		return nil, fmt.Errorf("assigning difficulty: %w", err)
	}

	// Step 6: Assign themes
	if err := assignThemes(g, cfg.Themes, rng); err != nil {
		return nil, fmt.Errorf("assigning themes: %w", err)
	}

	// Step 7: Validate
	if err := validateTemplateGraph(g, cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return g, nil
}

// instantiateTemplate creates rooms and connectors from a template.
// Returns the list of attachment point room IDs (prefixed).
func (s *TemplateSynthesizer) instantiateTemplate(g *graph.Graph, tmpl GraphTemplate, prefix string, rng *rng.RNG) ([]string, error) {
	// Create rooms
	for _, roomTmpl := range tmpl.Rooms {
		room := &graph.Room{
			ID:         prefix + roomTmpl.LocalID,
			Archetype:  roomTmpl.Archetype,
			Size:       roomTmpl.Size,
			Tags:       map[string]string{"template": tmpl.Name},
			Difficulty: rng.Float64(),
			Reward:     rng.Float64(),
		}
		if err := g.AddRoom(room); err != nil {
			return nil, fmt.Errorf("adding room %s: %w", room.ID, err)
		}
	}

	// Create connectors
	for _, connTmpl := range tmpl.Connectors {
		conn := &graph.Connector{
			ID:            prefix + "conn_" + connTmpl.LocalFromID + "_" + connTmpl.LocalToID,
			From:          prefix + connTmpl.LocalFromID,
			To:            prefix + connTmpl.LocalToID,
			Type:          connTmpl.Type,
			Cost:          1.0,
			Visibility:    connTmpl.Visibility,
			Bidirectional: connTmpl.Bidirectional,
		}
		if err := g.AddConnector(conn); err != nil {
			return nil, fmt.Errorf("adding connector %s: %w", conn.ID, err)
		}
	}

	// Collect attachment points
	attachments := make([]string, len(tmpl.Attachments))
	for i, localID := range tmpl.Attachments {
		attachments[i] = prefix + localID
	}

	return attachments, nil
}

// selectTemplate picks a random template from the given category.
func (s *TemplateSynthesizer) selectTemplate(category string, rng *rng.RNG) GraphTemplate {
	templates := s.templates[category]
	if len(templates) == 0 {
		// Fallback to simple template
		return createFallbackTemplate(category)
	}
	return templates[rng.Intn(len(templates))]
}

// registerDefaultTemplates adds built-in template patterns.
func (s *TemplateSynthesizer) registerDefaultTemplates() {
	// Start template: Start room with 2 adjacent rooms
	s.templates["start"] = []GraphTemplate{
		{
			Name:     "linear_start",
			Category: "start",
			Rooms: []*RoomTemplate{
				{LocalID: "start", Archetype: graph.ArchetypeStart, Size: graph.SizeM},
				{LocalID: "r1", Archetype: graph.ArchetypeCorridor, Size: graph.SizeS},
				{LocalID: "r2", Archetype: graph.ArchetypeTreasure, Size: graph.SizeM},
			},
			Connectors: []*ConnectorTemplate{
				{LocalFromID: "start", LocalToID: "r1", Type: graph.TypeCorridor, Bidirectional: true, Visibility: graph.VisibilityNormal},
				{LocalFromID: "r1", LocalToID: "r2", Type: graph.TypeDoor, Bidirectional: true, Visibility: graph.VisibilityNormal},
			},
			Attachments: []string{"r2"}, // r2 can connect forward
		},
	}

	// Boss template: Boss room with approach corridor
	s.templates["boss"] = []GraphTemplate{
		{
			Name:     "guarded_boss",
			Category: "boss",
			Rooms: []*RoomTemplate{
				{LocalID: "approach", Archetype: graph.ArchetypeCorridor, Size: graph.SizeM},
				{LocalID: "boss", Archetype: graph.ArchetypeBoss, Size: graph.SizeXL},
			},
			Connectors: []*ConnectorTemplate{
				{LocalFromID: "approach", LocalToID: "boss", Type: graph.TypeDoor, Bidirectional: true, Visibility: graph.VisibilityNormal},
			},
			Attachments: []string{"approach"}, // approach is the entry point
		},
	}

	// Mid templates: Various connecting sections
	s.templates["mid"] = []GraphTemplate{
		{
			Name:     "corridor_chain",
			Category: "mid",
			Rooms: []*RoomTemplate{
				{LocalID: "c1", Archetype: graph.ArchetypeCorridor, Size: graph.SizeS},
				{LocalID: "c2", Archetype: graph.ArchetypeCorridor, Size: graph.SizeS},
				{LocalID: "c3", Archetype: graph.ArchetypeCorridor, Size: graph.SizeM},
			},
			Connectors: []*ConnectorTemplate{
				{LocalFromID: "c1", LocalToID: "c2", Type: graph.TypeCorridor, Bidirectional: true, Visibility: graph.VisibilityNormal},
				{LocalFromID: "c2", LocalToID: "c3", Type: graph.TypeCorridor, Bidirectional: true, Visibility: graph.VisibilityNormal},
			},
			Attachments: []string{"c1", "c3"},
		},
		{
			Name:     "hub_room",
			Category: "mid",
			Rooms: []*RoomTemplate{
				{LocalID: "hub", Archetype: graph.ArchetypeHub, Size: graph.SizeL},
			},
			Connectors:  []*ConnectorTemplate{},
			Attachments: []string{"hub"},
		},
	}

	// Branch templates: Optional side content
	s.templates["branch"] = []GraphTemplate{
		{
			Name:     "treasure_branch",
			Category: "branch",
			Rooms: []*RoomTemplate{
				{LocalID: "fork", Archetype: graph.ArchetypeCorridor, Size: graph.SizeS},
				{LocalID: "treasure", Archetype: graph.ArchetypeTreasure, Size: graph.SizeM},
			},
			Connectors: []*ConnectorTemplate{
				{LocalFromID: "fork", LocalToID: "treasure", Type: graph.TypeDoor, Bidirectional: true, Visibility: graph.VisibilityNormal},
			},
			Attachments: []string{"fork"},
		},
	}
}

// createFallbackTemplate generates a simple template when category is empty.
func createFallbackTemplate(category string) GraphTemplate {
	switch category {
	case "start":
		return GraphTemplate{
			Name:     "simple_start",
			Category: "start",
			Rooms: []*RoomTemplate{
				{LocalID: "start", Archetype: graph.ArchetypeStart, Size: graph.SizeM},
			},
			Connectors:  []*ConnectorTemplate{},
			Attachments: []string{"start"},
		}
	case "boss":
		return GraphTemplate{
			Name:     "simple_boss",
			Category: "boss",
			Rooms: []*RoomTemplate{
				{LocalID: "boss", Archetype: graph.ArchetypeBoss, Size: graph.SizeXL},
			},
			Connectors:  []*ConnectorTemplate{},
			Attachments: []string{"boss"},
		}
	default:
		return GraphTemplate{
			Name:     "simple_room",
			Category: category,
			Rooms: []*RoomTemplate{
				{LocalID: "r1", Archetype: graph.ArchetypeCorridor, Size: graph.SizeM},
			},
			Connectors:  []*ConnectorTemplate{},
			Attachments: []string{"r1"},
		}
	}
}

// assignDifficultyTemplate assigns difficulty to all rooms based on pacing curve.
func assignDifficultyTemplate(g *graph.Graph, rng *rng.RNG, cfg *Config) error {
	// Find Start and Boss
	var startID, bossID string
	for id, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			startID = id
		} else if room.Archetype == graph.ArchetypeBoss {
			bossID = id
		}
	}

	if startID == "" || bossID == "" {
		return fmt.Errorf("missing Start or Boss room")
	}

	// Get critical path
	path, err := g.GetPath(startID, bossID)
	if err != nil {
		return fmt.Errorf("no path from Start to Boss: %w", err)
	}

	// Create pacing curve
	curve, err := createPacingCurveFromConfig(cfg.Pacing)
	if err != nil {
		return err
	}

	// Assign difficulty along path
	progressMap := make(map[string]float64)
	for i, roomID := range path {
		progress := 0.0
		if len(path) > 1 {
			progress = float64(i) / float64(len(path)-1)
		}
		progressMap[roomID] = progress

		room := g.Rooms[roomID]
		room.Difficulty = EvaluateWithVariance(curve, progress, cfg.Pacing.Variance, rng)
		room.Reward = room.Difficulty * 0.8 // Scale reward with difficulty
	}

	// Assign difficulty to off-path rooms
	for id, room := range g.Rooms {
		if _, onPath := progressMap[id]; !onPath {
			// Use average difficulty of neighbors
			avgDifficulty := 0.5
			neighborCount := 0
			for _, neighborID := range g.Adjacency[id] {
				if neighborProgress, exists := progressMap[neighborID]; exists {
					avgDifficulty += EvaluateWithVariance(curve, neighborProgress, cfg.Pacing.Variance, rng)
					neighborCount++
				}
			}
			if neighborCount > 0 {
				avgDifficulty /= float64(neighborCount + 1)
			}
			room.Difficulty = avgDifficulty
			room.Reward = avgDifficulty * 0.8
		}
	}

	return nil
}

// createPacingCurveFromConfig creates a curve from config.
func createPacingCurveFromConfig(cfg PacingConfig) (PacingCurve, error) {
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
		return &LinearCurve{}, nil
	}
}

// validateTemplateGraph checks hard constraints.
func validateTemplateGraph(g *graph.Graph, cfg *Config) error {
	// Check room count
	if len(g.Rooms) < cfg.RoomsMin || len(g.Rooms) > cfg.RoomsMax {
		return fmt.Errorf("room count %d outside bounds [%d, %d]", len(g.Rooms), cfg.RoomsMin, cfg.RoomsMax)
	}

	// Check connectivity
	if !g.IsConnected() {
		return fmt.Errorf("graph is not connected")
	}

	// Check for Start and Boss
	hasStart := false
	hasBoss := false
	for _, room := range g.Rooms {
		if room.Archetype == graph.ArchetypeStart {
			hasStart = true
		}
		if room.Archetype == graph.ArchetypeBoss {
			hasBoss = true
		}
	}
	if !hasStart {
		return fmt.Errorf("missing Start room")
	}
	if !hasBoss {
		return fmt.Errorf("missing Boss room")
	}

	return nil
}

// init registers the template synthesizer.
func init() {
	Register("template", NewTemplateSynthesizer())
}
