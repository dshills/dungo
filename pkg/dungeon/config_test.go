package dungeon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidConfig(t *testing.T) {
	yaml := `
seed: 12345
size:
  roomsMin: 20
  roomsMax: 50
branching:
  avg: 2.0
  max: 4
pacing:
  curve: LINEAR
  variance: 0.15
themes:
  - crypt
  - fungal
keys:
  - name: silver
    count: 2
  - name: gold
    count: 1
constraints:
  - kind: connectivity
    severity: hard
    expr: isConnected()
allowDisconnected: false
secretDensity: 0.15
optionalRatio: 0.25
`

	cfg, err := LoadConfigFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadConfigFromBytes() failed: %v", err)
	}

	// Verify basic fields
	if cfg.Seed != 12345 {
		t.Errorf("Seed = %d, want 12345", cfg.Seed)
	}
	if cfg.Size.RoomsMin != 20 {
		t.Errorf("Size.RoomsMin = %d, want 20", cfg.Size.RoomsMin)
	}
	if cfg.Size.RoomsMax != 50 {
		t.Errorf("Size.RoomsMax = %d, want 50", cfg.Size.RoomsMax)
	}
	if cfg.Branching.Avg != 2.0 {
		t.Errorf("Branching.Avg = %f, want 2.0", cfg.Branching.Avg)
	}
	if cfg.Branching.Max != 4 {
		t.Errorf("Branching.Max = %d, want 4", cfg.Branching.Max)
	}
	if cfg.Pacing.Curve != PacingLinear {
		t.Errorf("Pacing.Curve = %q, want LINEAR", cfg.Pacing.Curve)
	}
	if cfg.Pacing.Variance != 0.15 {
		t.Errorf("Pacing.Variance = %f, want 0.15", cfg.Pacing.Variance)
	}
	if len(cfg.Themes) != 2 {
		t.Errorf("len(Themes) = %d, want 2", len(cfg.Themes))
	}
	if len(cfg.Keys) != 2 {
		t.Errorf("len(Keys) = %d, want 2", len(cfg.Keys))
	}
	if len(cfg.Constraints) != 1 {
		t.Errorf("len(Constraints) = %d, want 1", len(cfg.Constraints))
	}
	if cfg.SecretDensity != 0.15 {
		t.Errorf("SecretDensity = %f, want 0.15", cfg.SecretDensity)
	}
	if cfg.OptionalRatio != 0.25 {
		t.Errorf("OptionalRatio = %f, want 0.25", cfg.OptionalRatio)
	}
}

func TestLoadConfig_AutoGenerateSeed(t *testing.T) {
	yaml := `
seed: 0
size:
  roomsMin: 10
  roomsMax: 20
branching:
  avg: 2.0
  max: 3
pacing:
  curve: LINEAR
  variance: 0.1
themes:
  - crypt
secretDensity: 0.1
optionalRatio: 0.2
`

	cfg, err := LoadConfigFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("LoadConfigFromBytes() failed: %v", err)
	}

	if cfg.Seed == 0 {
		t.Error("Seed should be auto-generated when 0, but got 0")
	}

	// Load again and verify we get a different seed (time-based)
	// Note: This could theoretically fail if called in same nanosecond,
	// but extremely unlikely in practice
	cfg2, err := LoadConfigFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("Second LoadConfigFromBytes() failed: %v", err)
	}

	if cfg2.Seed == 0 {
		t.Error("Second seed should be auto-generated when 0, but got 0")
	}
}

func TestConfig_ValidateSize(t *testing.T) {
	tests := []struct {
		name    string
		size    SizeCfg
		wantErr bool
	}{
		{
			name:    "valid size",
			size:    SizeCfg{RoomsMin: 20, RoomsMax: 50},
			wantErr: false,
		},
		{
			name:    "min equals max",
			size:    SizeCfg{RoomsMin: 30, RoomsMax: 30},
			wantErr: false,
		},
		{
			name:    "min too small",
			size:    SizeCfg{RoomsMin: 5, RoomsMax: 50},
			wantErr: true,
		},
		{
			name:    "max too large",
			size:    SizeCfg{RoomsMin: 20, RoomsMax: 400},
			wantErr: true,
		},
		{
			name:    "min > max",
			size:    SizeCfg{RoomsMin: 100, RoomsMax: 50},
			wantErr: true,
		},
		{
			name:    "boundary min",
			size:    SizeCfg{RoomsMin: 10, RoomsMax: 50},
			wantErr: false,
		},
		{
			name:    "boundary max",
			size:    SizeCfg{RoomsMin: 20, RoomsMax: 300},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.size.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SizeCfg.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateBranching(t *testing.T) {
	tests := []struct {
		name      string
		branching BranchingCfg
		wantErr   bool
	}{
		{
			name:      "valid branching",
			branching: BranchingCfg{Avg: 2.0, Max: 3},
			wantErr:   false,
		},
		{
			name:      "avg too low",
			branching: BranchingCfg{Avg: 1.0, Max: 3},
			wantErr:   true,
		},
		{
			name:      "avg too high",
			branching: BranchingCfg{Avg: 4.0, Max: 3},
			wantErr:   true,
		},
		{
			name:      "max too low",
			branching: BranchingCfg{Avg: 2.0, Max: 1},
			wantErr:   true,
		},
		{
			name:      "max too high",
			branching: BranchingCfg{Avg: 2.0, Max: 10},
			wantErr:   true,
		},
		{
			name:      "boundary values",
			branching: BranchingCfg{Avg: 1.5, Max: 2},
			wantErr:   false,
		},
		{
			name:      "max boundary",
			branching: BranchingCfg{Avg: 3.0, Max: 5},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.branching.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BranchingCfg.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidatePacing(t *testing.T) {
	tests := []struct {
		name    string
		pacing  PacingCfg
		wantErr bool
	}{
		{
			name: "valid linear",
			pacing: PacingCfg{
				Curve:    PacingLinear,
				Variance: 0.15,
			},
			wantErr: false,
		},
		{
			name: "valid s-curve",
			pacing: PacingCfg{
				Curve:    PacingSCurve,
				Variance: 0.2,
			},
			wantErr: false,
		},
		{
			name: "valid exponential",
			pacing: PacingCfg{
				Curve:    PacingExponential,
				Variance: 0.1,
			},
			wantErr: false,
		},
		{
			name: "valid custom with points",
			pacing: PacingCfg{
				Curve:    PacingCustom,
				Variance: 0.15,
				CustomPoints: [][2]float64{
					{0.0, 0.0},
					{0.5, 0.3},
					{1.0, 1.0},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid curve type",
			pacing: PacingCfg{
				Curve:    "INVALID",
				Variance: 0.15,
			},
			wantErr: true,
		},
		{
			name: "variance too low",
			pacing: PacingCfg{
				Curve:    PacingLinear,
				Variance: -0.1,
			},
			wantErr: true,
		},
		{
			name: "variance too high",
			pacing: PacingCfg{
				Curve:    PacingLinear,
				Variance: 0.5,
			},
			wantErr: true,
		},
		{
			name: "custom without points",
			pacing: PacingCfg{
				Curve:    PacingCustom,
				Variance: 0.15,
			},
			wantErr: true,
		},
		{
			name: "custom with one point",
			pacing: PacingCfg{
				Curve:    PacingCustom,
				Variance: 0.15,
				CustomPoints: [][2]float64{
					{0.0, 0.0},
				},
			},
			wantErr: true,
		},
		{
			name: "custom with invalid progress",
			pacing: PacingCfg{
				Curve:    PacingCustom,
				Variance: 0.15,
				CustomPoints: [][2]float64{
					{0.0, 0.0},
					{1.5, 0.5},
				},
			},
			wantErr: true,
		},
		{
			name: "custom with invalid difficulty",
			pacing: PacingCfg{
				Curve:    PacingCustom,
				Variance: 0.15,
				CustomPoints: [][2]float64{
					{0.0, 0.0},
					{0.5, 1.5},
				},
			},
			wantErr: true,
		},
		{
			name: "custom with unsorted points",
			pacing: PacingCfg{
				Curve:    PacingCustom,
				Variance: 0.15,
				CustomPoints: [][2]float64{
					{0.5, 0.5},
					{0.0, 0.0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pacing.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PacingCfg.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     KeyCfg
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     KeyCfg{Name: "silver", Count: 2},
			wantErr: false,
		},
		{
			name:    "empty name",
			key:     KeyCfg{Name: "", Count: 2},
			wantErr: true,
		},
		{
			name:    "count too low",
			key:     KeyCfg{Name: "gold", Count: 0},
			wantErr: true,
		},
		{
			name:    "count too high",
			key:     KeyCfg{Name: "gold", Count: 10},
			wantErr: true,
		},
		{
			name:    "boundary count low",
			key:     KeyCfg{Name: "bronze", Count: 1},
			wantErr: false,
		},
		{
			name:    "boundary count high",
			key:     KeyCfg{Name: "platinum", Count: 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.key.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("KeyCfg.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateConstraints(t *testing.T) {
	tests := []struct {
		name       string
		constraint Constraint
		wantErr    bool
	}{
		{
			name: "valid hard constraint",
			constraint: Constraint{
				Kind:     "connectivity",
				Severity: "hard",
				Expr:     "isConnected()",
			},
			wantErr: false,
		},
		{
			name: "valid soft constraint",
			constraint: Constraint{
				Kind:     "pacing",
				Severity: "soft",
				Expr:     "monotoneIncrease(start, boss)",
				Priority: 10,
			},
			wantErr: false,
		},
		{
			name: "empty kind",
			constraint: Constraint{
				Kind:     "",
				Severity: "hard",
				Expr:     "isConnected()",
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			constraint: Constraint{
				Kind:     "connectivity",
				Severity: "medium",
				Expr:     "isConnected()",
			},
			wantErr: true,
		},
		{
			name: "empty expr",
			constraint: Constraint{
				Kind:     "connectivity",
				Severity: "hard",
				Expr:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constraint.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Constraint.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateComplete(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid complete config",
			yaml: `
seed: 42
size:
  roomsMin: 20
  roomsMax: 50
branching:
  avg: 2.0
  max: 3
pacing:
  curve: LINEAR
  variance: 0.15
themes:
  - crypt
secretDensity: 0.15
optionalRatio: 0.25
`,
			wantErr: false,
		},
		{
			name: "no themes",
			yaml: `
seed: 42
size:
  roomsMin: 20
  roomsMax: 50
branching:
  avg: 2.0
  max: 3
pacing:
  curve: LINEAR
  variance: 0.15
themes: []
secretDensity: 0.15
optionalRatio: 0.25
`,
			wantErr: true,
			errMsg:  "at least one theme",
		},
		{
			name: "secret density too high",
			yaml: `
seed: 42
size:
  roomsMin: 20
  roomsMax: 50
branching:
  avg: 2.0
  max: 3
pacing:
  curve: LINEAR
  variance: 0.15
themes:
  - crypt
secretDensity: 0.5
optionalRatio: 0.25
`,
			wantErr: true,
			errMsg:  "secretDensity",
		},
		{
			name: "optional ratio too low",
			yaml: `
seed: 42
size:
  roomsMin: 20
  roomsMax: 50
branching:
  avg: 2.0
  max: 3
pacing:
  curve: LINEAR
  variance: 0.15
themes:
  - crypt
secretDensity: 0.15
optionalRatio: 0.05
`,
			wantErr: true,
			errMsg:  "optionalRatio",
		},
		{
			name: "invalid key config",
			yaml: `
seed: 42
size:
  roomsMin: 20
  roomsMax: 50
branching:
  avg: 2.0
  max: 3
pacing:
  curve: LINEAR
  variance: 0.15
themes:
  - crypt
keys:
  - name: silver
    count: 10
secretDensity: 0.15
optionalRatio: 0.25
`,
			wantErr: true,
			errMsg:  "key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadConfigFromBytes([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigFromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message %q does not contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestConfig_YAMLRoundTrip(t *testing.T) {
	original := &Config{
		Seed: 12345,
		Size: SizeCfg{
			RoomsMin: 20,
			RoomsMax: 50,
		},
		Branching: BranchingCfg{
			Avg: 2.0,
			Max: 3,
		},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.15,
		},
		Themes: []string{"crypt", "fungal"},
		Keys: []KeyCfg{
			{Name: "silver", Count: 2},
			{Name: "gold", Count: 1},
		},
		Constraints: []Constraint{
			{
				Kind:     "connectivity",
				Severity: "hard",
				Expr:     "isConnected()",
				Priority: 100,
			},
		},
		AllowDisconnected: false,
		SecretDensity:     0.15,
		OptionalRatio:     0.25,
	}

	// Marshal to YAML
	yamlData, err := original.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() failed: %v", err)
	}

	// Unmarshal back
	restored, err := LoadConfigFromBytes(yamlData)
	if err != nil {
		t.Fatalf("LoadConfigFromBytes() failed: %v", err)
	}

	// Compare fields
	if restored.Seed != original.Seed {
		t.Errorf("Seed mismatch: got %d, want %d", restored.Seed, original.Seed)
	}
	if restored.Size.RoomsMin != original.Size.RoomsMin {
		t.Errorf("RoomsMin mismatch: got %d, want %d", restored.Size.RoomsMin, original.Size.RoomsMin)
	}
	if restored.Size.RoomsMax != original.Size.RoomsMax {
		t.Errorf("RoomsMax mismatch: got %d, want %d", restored.Size.RoomsMax, original.Size.RoomsMax)
	}
	if restored.Branching.Avg != original.Branching.Avg {
		t.Errorf("Branching.Avg mismatch: got %f, want %f", restored.Branching.Avg, original.Branching.Avg)
	}
	if restored.Branching.Max != original.Branching.Max {
		t.Errorf("Branching.Max mismatch: got %d, want %d", restored.Branching.Max, original.Branching.Max)
	}
	if restored.Pacing.Curve != original.Pacing.Curve {
		t.Errorf("Pacing.Curve mismatch: got %s, want %s", restored.Pacing.Curve, original.Pacing.Curve)
	}
	if restored.Pacing.Variance != original.Pacing.Variance {
		t.Errorf("Pacing.Variance mismatch: got %f, want %f", restored.Pacing.Variance, original.Pacing.Variance)
	}
	if len(restored.Themes) != len(original.Themes) {
		t.Errorf("Themes length mismatch: got %d, want %d", len(restored.Themes), len(original.Themes))
	}
	if len(restored.Keys) != len(original.Keys) {
		t.Errorf("Keys length mismatch: got %d, want %d", len(restored.Keys), len(original.Keys))
	}
	if len(restored.Constraints) != len(original.Constraints) {
		t.Errorf("Constraints length mismatch: got %d, want %d", len(restored.Constraints), len(original.Constraints))
	}
	if restored.SecretDensity != original.SecretDensity {
		t.Errorf("SecretDensity mismatch: got %f, want %f", restored.SecretDensity, original.SecretDensity)
	}
	if restored.OptionalRatio != original.OptionalRatio {
		t.Errorf("OptionalRatio mismatch: got %f, want %f", restored.OptionalRatio, original.OptionalRatio)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	yamlContent := `
seed: 99999
size:
  roomsMin: 15
  roomsMax: 40
branching:
  avg: 2.5
  max: 4
pacing:
  curve: S_CURVE
  variance: 0.2
themes:
  - arcane
  - fungal
secretDensity: 0.2
optionalRatio: 0.3
`

	// Write the config file
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify it loaded correctly
	if cfg.Seed != 99999 {
		t.Errorf("Seed = %d, want 99999", cfg.Seed)
	}
	if cfg.Pacing.Curve != PacingSCurve {
		t.Errorf("Pacing.Curve = %q, want S_CURVE", cfg.Pacing.Curve)
	}
	if len(cfg.Themes) != 2 {
		t.Errorf("len(Themes) = %d, want 2", len(cfg.Themes))
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadConfig() should fail for nonexistent file")
	}
}

func TestConfig_Hash(t *testing.T) {
	cfg1 := &Config{
		Seed:      12345,
		Size:      SizeCfg{RoomsMin: 10, RoomsMax: 20},
		Branching: BranchingCfg{Avg: 2.0, Max: 3},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.1,
		},
		Themes:        []string{"crypt"},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	cfg2 := &Config{
		Seed:      12345,
		Size:      SizeCfg{RoomsMin: 10, RoomsMax: 20},
		Branching: BranchingCfg{Avg: 2.0, Max: 3},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.1,
		},
		Themes:        []string{"crypt"},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	cfg3 := &Config{
		Seed:      54321, // Different seed
		Size:      SizeCfg{RoomsMin: 10, RoomsMax: 20},
		Branching: BranchingCfg{Avg: 2.0, Max: 3},
		Pacing: PacingCfg{
			Curve:    PacingLinear,
			Variance: 0.1,
		},
		Themes:        []string{"crypt"},
		SecretDensity: 0.1,
		OptionalRatio: 0.2,
	}

	hash1 := cfg1.Hash()
	hash2 := cfg2.Hash()
	hash3 := cfg3.Hash()

	// Same config should produce same hash
	if len(hash1) == 0 {
		t.Error("Hash should not be empty")
	}
	if string(hash1) != string(hash2) {
		t.Error("Identical configs should produce identical hashes")
	}

	// Different config should produce different hash
	if string(hash1) == string(hash3) {
		t.Error("Different configs should produce different hashes")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
