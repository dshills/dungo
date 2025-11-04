package validation

import (
	"context"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/graph"
)

// Helper to create a basic connected graph for testing
func createTestGraph() *graph.Graph {
	g := graph.NewGraph(12345)

	// Create rooms
	start := &graph.Room{
		ID:         "start",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Difficulty: 0.0,
		Reward:     0.1,
	}

	mid1 := &graph.Room{
		ID:         "mid1",
		Archetype:  graph.ArchetypeHub,
		Size:       graph.SizeM,
		Difficulty: 0.3,
		Reward:     0.2,
	}

	mid2 := &graph.Room{
		ID:         "mid2",
		Archetype:  graph.ArchetypeTreasure,
		Size:       graph.SizeS,
		Difficulty: 0.5,
		Reward:     0.5,
	}

	boss := &graph.Room{
		ID:         "boss",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeXL,
		Difficulty: 1.0,
		Reward:     1.0,
	}

	if err := g.AddRoom(start); err != nil {
		panic(err)
	}
	if err := g.AddRoom(mid1); err != nil {
		panic(err)
	}
	if err := g.AddRoom(mid2); err != nil {
		panic(err)
	}
	if err := g.AddRoom(boss); err != nil {
		panic(err)
	}

	// Create connections
	if err := g.AddConnector(&graph.Connector{
		ID:            "c1",
		From:          "start",
		To:            "mid1",
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Bidirectional: true,
	}); err != nil {
		panic(err)
	}

	if err := g.AddConnector(&graph.Connector{
		ID:            "c2",
		From:          "mid1",
		To:            "mid2",
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Bidirectional: true,
	}); err != nil {
		panic(err)
	}

	if err := g.AddConnector(&graph.Connector{
		ID:            "c3",
		From:          "mid2",
		To:            "boss",
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Bidirectional: true,
	}); err != nil {
		panic(err)
	}

	return g
}

// Helper to create a disconnected graph
func createDisconnectedGraph() *graph.Graph {
	g := graph.NewGraph(12345)

	// Component 1
	start := &graph.Room{
		ID:         "start",
		Archetype:  graph.ArchetypeStart,
		Size:       graph.SizeM,
		Difficulty: 0.0,
		Reward:     0.1,
	}

	mid1 := &graph.Room{
		ID:         "mid1",
		Archetype:  graph.ArchetypeHub,
		Size:       graph.SizeM,
		Difficulty: 0.3,
		Reward:     0.2,
	}

	// Component 2 (disconnected)
	boss := &graph.Room{
		ID:         "boss",
		Archetype:  graph.ArchetypeBoss,
		Size:       graph.SizeXL,
		Difficulty: 1.0,
		Reward:     1.0,
	}

	if err := g.AddRoom(start); err != nil {
		panic(err)
	}
	if err := g.AddRoom(mid1); err != nil {
		panic(err)
	}
	if err := g.AddRoom(boss); err != nil {
		panic(err)
	}

	// Only connect start and mid1, leaving boss disconnected
	if err := g.AddConnector(&graph.Connector{
		ID:            "c1",
		From:          "start",
		To:            "mid1",
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Bidirectional: true,
	}); err != nil {
		panic(err)
	}

	return g
}

// Helper to create test config
func createTestConfig() *dungeon.Config {
	return &dungeon.Config{
		Seed: 12345,
		Size: dungeon.SizeCfg{
			RoomsMin: 10,
			RoomsMax: 50,
		},
		Branching: dungeon.BranchingCfg{
			Avg: 2.0,
			Max: 4,
		},
		Pacing: dungeon.PacingCfg{
			Curve:    dungeon.PacingLinear,
			Variance: 0.2,
		},
		Themes: []string{"crypt"},
	}
}

func TestCheckConnectivity_Connected(t *testing.T) {
	g := createTestGraph()

	// Debug: Check adjacency
	t.Logf("Adjacency list: %+v", g.Adjacency)
	t.Logf("Rooms: %d, Connectors: %d", len(g.Rooms), len(g.Connectors))

	result := CheckConnectivity(g)

	if !result.Satisfied {
		t.Errorf("Expected connectivity check to pass, got: %s", result.Details)
	}

	if result.Score != 1.0 {
		t.Errorf("Expected score 1.0, got: %f", result.Score)
	}
}

func TestCheckConnectivity_Disconnected(t *testing.T) {
	g := createDisconnectedGraph()

	result := CheckConnectivity(g)

	if result.Satisfied {
		t.Errorf("Expected connectivity check to fail for disconnected graph")
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score 0.0, got: %f", result.Score)
	}
}

func TestCheckKeyReachability_NoKeys(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()

	result := CheckKeyReachability(g, cfg)

	if !result.Satisfied {
		t.Errorf("Expected key reachability to pass when no keys configured, got: %s", result.Details)
	}
}

func TestCheckKeyReachability_ValidKeyPlacement(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()

	// Add a key to mid1
	g.Rooms["mid1"].Provides = []graph.Capability{
		{Type: "key", Value: "silver"},
	}

	// Add a lock to boss requiring silver key
	g.Rooms["boss"].Requirements = []graph.Requirement{
		{Type: "key", Value: "silver"},
	}

	// Configure the key
	cfg.Keys = []dungeon.KeyCfg{
		{Name: "silver", Count: 1},
	}

	result := CheckKeyReachability(g, cfg)

	if !result.Satisfied {
		t.Errorf("Expected key reachability to pass, got: %s", result.Details)
	}
}

func TestCheckKeyReachability_CircularDependency(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()

	// Add a key to mid1 that requires itself
	g.Rooms["mid1"].Requirements = []graph.Requirement{
		{Type: "key", Value: "silver"},
	}
	g.Rooms["mid1"].Provides = []graph.Capability{
		{Type: "key", Value: "silver"},
	}

	cfg.Keys = []dungeon.KeyCfg{
		{Name: "silver", Count: 1},
	}

	result := CheckKeyReachability(g, cfg)

	if result.Satisfied {
		t.Errorf("Expected key reachability to fail for circular dependency")
	}
}

func TestCheckPathBounds_ValidPath(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()

	result := CheckPathBounds(g, cfg)

	if !result.Satisfied {
		t.Errorf("Expected path bounds check to pass, got: %s", result.Details)
	}
}

func TestCheckPathBounds_NoPath(t *testing.T) {
	g := createDisconnectedGraph()
	cfg := createTestConfig()

	result := CheckPathBounds(g, cfg)

	if result.Satisfied {
		t.Errorf("Expected path bounds check to fail when no path exists")
	}
}

func TestCalculateBranchingFactor(t *testing.T) {
	g := createTestGraph()

	// Graph has 4 rooms and 3 connectors (edges)
	// Average degree = (2 * 3) / 4 = 1.5
	expected := 1.5

	actual := CalculateBranchingFactor(g)

	if actual != expected {
		t.Errorf("Expected branching factor %f, got %f", expected, actual)
	}
}

func TestCalculatePathLength(t *testing.T) {
	g := createTestGraph()

	// Path: start -> mid1 -> mid2 -> boss = 3 edges
	expected := 3

	actual := CalculatePathLength(g)

	if actual != expected {
		t.Errorf("Expected path length %d, got %d", expected, actual)
	}
}

func TestCountCycles_NoCycles(t *testing.T) {
	g := createTestGraph()

	// This is a linear graph with no cycles
	expected := 0

	actual := CountCycles(g)

	if actual != expected {
		t.Errorf("Expected %d cycles, got %d", expected, actual)
	}
}

func TestCountCycles_WithCycle(t *testing.T) {
	g := createTestGraph()

	// Add a connection that creates a cycle
	if err := g.AddConnector(&graph.Connector{
		ID:            "c4",
		From:          "mid1",
		To:            "boss",
		Type:          graph.TypeCorridor,
		Cost:          1.0,
		Bidirectional: true,
	}); err != nil {
		t.Fatalf("Failed to add connector: %v", err)
	}

	// Now there should be at least one cycle
	actual := CountCycles(g)

	if actual == 0 {
		t.Errorf("Expected at least one cycle, got %d", actual)
	}
}

func TestCalculatePacingDeviation_Linear(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()
	cfg.Pacing.Curve = dungeon.PacingLinear

	deviation := CalculatePacingDeviation(g, cfg)

	// Deviation should be low for a well-paced dungeon
	// Values: 0.0, 0.3, 0.5, 1.0 vs expected 0.0, 0.33, 0.67, 1.0
	if deviation > 0.5 {
		t.Errorf("Expected low deviation, got %f", deviation)
	}
}

func TestCalculatePacingDeviation_SCurve(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()
	cfg.Pacing.Curve = dungeon.PacingSCurve

	deviation := CalculatePacingDeviation(g, cfg)

	// S-curve should have different expected values
	if deviation < 0.0 || deviation > 1.0 {
		t.Errorf("Deviation out of range: %f", deviation)
	}
}

func TestCalculatePacingDeviation_Exponential(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()
	cfg.Pacing.Curve = dungeon.PacingExponential

	deviation := CalculatePacingDeviation(g, cfg)

	if deviation < 0.0 || deviation > 1.0 {
		t.Errorf("Deviation out of range: %f", deviation)
	}
}

func TestGetDegreeDistribution(t *testing.T) {
	g := createTestGraph()

	dist := GetDegreeDistribution(g)

	// start and boss have degree 1
	// mid1 and mid2 have degree 2
	if dist[1] != 2 {
		t.Errorf("Expected 2 rooms with degree 1, got %d", dist[1])
	}
	if dist[2] != 2 {
		t.Errorf("Expected 2 rooms with degree 2, got %d", dist[2])
	}
}

func TestValidator_ValidDungeon(t *testing.T) {
	g := createTestGraph()
	cfg := createTestConfig()

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
		Layout: &dungeon.Layout{
			Poses: map[string]dungeon.Pose{
				// Poses now use center coordinates with proper spacing
				// start/mid1: SizeM(8), mid2: SizeS(5), boss: SizeXL(16)
				"start": {X: 4, Y: 8, Rotation: 0},  // SizeM(8): center at 4 (corners: 0-8)
				"mid1":  {X: 14, Y: 8, Rotation: 0}, // SizeM(8): center at 14 (corners: 10-18)
				"mid2":  {X: 24, Y: 8, Rotation: 0}, // SizeS(5): center at 24 (corners: 22-27)
				"boss":  {X: 38, Y: 8, Rotation: 0}, // SizeXL(16): center at 38 (corners: 30-46)
			},
			CorridorPaths: map[string]dungeon.Path{},
			Bounds:        dungeon.Rect{X: 0, Y: 0, Width: 50, Height: 18},
		},
	}

	validator := NewValidator()
	report, err := validator.Validate(context.Background(), artifact, cfg)

	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}

	if !report.Passed {
		t.Errorf("Expected validation to pass, but got failures: %v", report.Errors)
	}

	if report.Metrics == nil {
		t.Errorf("Expected metrics to be calculated")
	}

	if len(report.HardConstraintResults) == 0 {
		t.Errorf("Expected hard constraint results")
	}
}

func TestValidator_DisconnectedGraph(t *testing.T) {
	g := createDisconnectedGraph()
	cfg := createTestConfig()

	artifact := &dungeon.Artifact{
		ADG: &dungeon.Graph{Graph: g},
	}

	validator := NewValidator()
	report, err := validator.Validate(context.Background(), artifact, cfg)

	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}

	if report.Passed {
		t.Errorf("Expected validation to fail for disconnected graph")
	}

	if len(report.Errors) == 0 {
		t.Errorf("Expected error messages for disconnected graph")
	}
}

func TestFindStartRoom(t *testing.T) {
	g := createTestGraph()

	startID := FindStartRoom(g)

	if startID != "start" {
		t.Errorf("Expected to find start room, got: %s", startID)
	}
}

func TestFindBossRoom(t *testing.T) {
	g := createTestGraph()

	bossID := FindBossRoom(g)

	if bossID != "boss" {
		t.Errorf("Expected to find boss room, got: %s", bossID)
	}
}

func TestFindKeyRooms(t *testing.T) {
	g := createTestGraph()

	// Add a key to mid1
	g.Rooms["mid1"].Provides = []graph.Capability{
		{Type: "key", Value: "silver"},
	}

	keyRooms := FindKeyRooms(g)

	if len(keyRooms["silver"]) != 1 {
		t.Errorf("Expected 1 room with silver key, got %d", len(keyRooms["silver"]))
	}

	if keyRooms["silver"][0] != "mid1" {
		t.Errorf("Expected mid1 to have silver key, got %s", keyRooms["silver"][0])
	}
}

func TestFindLockedRooms(t *testing.T) {
	g := createTestGraph()

	// Add a lock to boss
	g.Rooms["boss"].Requirements = []graph.Requirement{
		{Type: "key", Value: "silver"},
	}

	lockedRooms := FindLockedRooms(g)

	if len(lockedRooms["silver"]) != 1 {
		t.Errorf("Expected 1 room requiring silver key, got %d", len(lockedRooms["silver"]))
	}

	if lockedRooms["silver"][0] != "boss" {
		t.Errorf("Expected boss to require silver key, got %s", lockedRooms["silver"][0])
	}
}

func TestReportSummary(t *testing.T) {
	report := NewValidationReport()
	report.Passed = true
	report.Metrics = &dungeon.Metrics{
		BranchingFactor: 2.0,
		PathLength:      10,
		CycleCount:      2,
		PacingDeviation: 0.15,
	}

	report.HardConstraintResults = append(report.HardConstraintResults,
		NewHardConstraintResult("Connectivity", "test", true, "All connected"))

	summary := Summary(report)

	if summary == "" {
		t.Errorf("Expected non-empty summary")
	}

	if !contains(summary, "PASSED") {
		t.Errorf("Expected summary to contain PASSED status")
	}

	if !contains(summary, "Branching Factor") {
		t.Errorf("Expected summary to contain metrics")
	}
}

func TestGetFailedConstraints(t *testing.T) {
	report := NewValidationReport()
	report.HardConstraintResults = []dungeon.ConstraintResult{
		NewHardConstraintResult("Test1", "expr1", true, "passed"),
		NewHardConstraintResult("Test2", "expr2", false, "failed"),
		NewHardConstraintResult("Test3", "expr3", true, "passed"),
	}

	failed := GetFailedConstraints(report)

	if len(failed) != 1 {
		t.Errorf("Expected 1 failed constraint, got %d", len(failed))
	}

	if failed[0].Constraint.Kind != "Test2" {
		t.Errorf("Expected Test2 to be the failed constraint")
	}
}

func TestGetLowScoringConstraints(t *testing.T) {
	report := NewValidationReport()
	report.SoftConstraintResults = []dungeon.ConstraintResult{
		NewSoftConstraintResult("Test1", "expr1", 0.9, "good"),
		NewSoftConstraintResult("Test2", "expr2", 0.6, "okay"),
		NewSoftConstraintResult("Test3", "expr3", 0.3, "poor"),
	}

	lowScoring := GetLowScoringConstraints(report, 0.7)

	if len(lowScoring) != 2 {
		t.Errorf("Expected 2 low-scoring constraints, got %d", len(lowScoring))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
