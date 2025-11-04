package dungeon

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config specifies all dungeon generation parameters.
// It supports YAML parsing and includes comprehensive validation.
type Config struct {
	// Seed is the master seed for deterministic generation.
	// Use 0 to auto-generate from current time.
	Seed uint64 `yaml:"seed" json:"seed"`

	// Size specifies room count constraints.
	Size SizeCfg `yaml:"size" json:"size"`

	// Branching controls connectivity parameters.
	Branching BranchingCfg `yaml:"branching" json:"branching"`

	// Pacing defines the difficulty curve.
	Pacing PacingCfg `yaml:"pacing" json:"pacing"`

	// Themes lists biome/theme names to use.
	Themes []string `yaml:"themes" json:"themes"`

	// Keys defines key/lock configurations.
	Keys []KeyCfg `yaml:"keys,omitempty" json:"keys,omitempty"`

	// Constraints lists hard and soft constraints.
	Constraints []Constraint `yaml:"constraints,omitempty" json:"constraints,omitempty"`

	// AllowDisconnected permits teleport motifs if true.
	AllowDisconnected bool `yaml:"allowDisconnected" json:"allowDisconnected"`

	// SecretDensity is the target ratio of secret rooms (0.0-0.3).
	SecretDensity float64 `yaml:"secretDensity" json:"secretDensity"`

	// OptionalRatio is the target ratio of optional rooms (0.1-0.4).
	OptionalRatio float64 `yaml:"optionalRatio" json:"optionalRatio"`
}

// SizeCfg specifies room count constraints.
type SizeCfg struct {
	// RoomsMin is the minimum number of rooms (10-300).
	RoomsMin int `yaml:"roomsMin" json:"roomsMin"`

	// RoomsMax is the maximum number of rooms (10-300).
	RoomsMax int `yaml:"roomsMax" json:"roomsMax"`
}

// BranchingCfg controls connectivity parameters.
type BranchingCfg struct {
	// Avg is the target average connections per room (1.5-3.0).
	Avg float64 `yaml:"avg" json:"avg"`

	// Max is the maximum connections for any single room (2-5).
	Max int `yaml:"max" json:"max"`
}

// PacingCfg defines the difficulty curve configuration.
type PacingCfg struct {
	// Curve specifies the curve type.
	Curve PacingCurve `yaml:"curve" json:"curve"`

	// Variance is the allowed deviation from curve (0.0-0.3).
	Variance float64 `yaml:"variance" json:"variance"`

	// CustomPoints are optional custom pacing points for CUSTOM curve.
	// Each point is [progress, difficulty] where both are 0.0-1.0.
	CustomPoints [][2]float64 `yaml:"customPoints,omitempty" json:"customPoints,omitempty"`
}

// PacingCurve defines valid pacing curve types.
type PacingCurve string

const (
	// PacingLinear represents a linear difficulty increase.
	PacingLinear PacingCurve = "LINEAR"

	// PacingSCurve represents an S-curve difficulty progression.
	PacingSCurve PacingCurve = "S_CURVE"

	// PacingExponential represents exponential difficulty growth.
	PacingExponential PacingCurve = "EXPONENTIAL"

	// PacingCustom allows user-defined difficulty points.
	PacingCustom PacingCurve = "CUSTOM"
)

// ValidPacingCurves lists all valid curve types.
var ValidPacingCurves = []PacingCurve{
	PacingLinear,
	PacingSCurve,
	PacingExponential,
	PacingCustom,
}

// KeyCfg defines a key/lock configuration.
type KeyCfg struct {
	// Name is the key identifier (e.g., "silver", "gold").
	Name string `yaml:"name" json:"name"`

	// Count is the number of this key type (1-5).
	Count int `yaml:"count" json:"count"`
}

// Constraint represents a rule that must be satisfied or optimized.
// The actual constraint system is defined in pkg/graph but Config needs
// to reference it for YAML parsing.
type Constraint struct {
	// Kind categorizes the constraint.
	Kind string `yaml:"kind" json:"kind"`

	// Severity determines enforcement level ("hard" or "soft").
	Severity string `yaml:"severity" json:"severity"`

	// Expr is the DSL expression defining the constraint.
	Expr string `yaml:"expr" json:"expr"`

	// Priority determines order for constraint solving (higher = earlier).
	Priority int `yaml:"priority,omitempty" json:"priority,omitempty"`
}

// LoadConfig reads and validates a YAML configuration file.
// Returns a validated Config or an error if parsing or validation fails.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Auto-generate seed if not provided
	if cfg.Seed == 0 {
		cfg.Seed = generateSeed()
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadConfigFromBytes parses YAML configuration from a byte slice.
// Useful for testing and programmatic config generation.
func LoadConfigFromBytes(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Auto-generate seed if not provided
	if cfg.Seed == 0 {
		cfg.Seed = generateSeed()
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate checks all configuration constraints.
// Returns an error describing the first validation failure, or nil if valid.
func (c *Config) Validate() error {
	// Validate Size
	if err := c.Size.Validate(); err != nil {
		return fmt.Errorf("size: %w", err)
	}

	// Validate Branching
	if err := c.Branching.Validate(); err != nil {
		return fmt.Errorf("branching: %w", err)
	}

	// Validate Pacing
	if err := c.Pacing.Validate(); err != nil {
		return fmt.Errorf("pacing: %w", err)
	}

	// Validate Themes
	if len(c.Themes) == 0 {
		return errors.New("at least one theme must be specified")
	}

	// Validate Keys
	for i, key := range c.Keys {
		if err := key.Validate(); err != nil {
			return fmt.Errorf("key[%d]: %w", i, err)
		}
	}

	// Validate SecretDensity
	if c.SecretDensity < 0.0 || c.SecretDensity > 0.3 {
		return fmt.Errorf("secretDensity must be in range [0.0, 0.3], got %f", c.SecretDensity)
	}

	// Validate OptionalRatio
	if c.OptionalRatio < 0.1 || c.OptionalRatio > 0.4 {
		return fmt.Errorf("optionalRatio must be in range [0.1, 0.4], got %f", c.OptionalRatio)
	}

	// Validate Constraints
	for i, constraint := range c.Constraints {
		if err := constraint.Validate(); err != nil {
			return fmt.Errorf("constraint[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate checks SizeCfg constraints.
func (s *SizeCfg) Validate() error {
	if s.RoomsMin < 10 {
		return fmt.Errorf("roomsMin must be at least 10, got %d", s.RoomsMin)
	}
	if s.RoomsMax > 300 {
		return fmt.Errorf("roomsMax must be at most 300, got %d", s.RoomsMax)
	}
	if s.RoomsMin > s.RoomsMax {
		return fmt.Errorf("roomsMin (%d) must be <= roomsMax (%d)", s.RoomsMin, s.RoomsMax)
	}
	return nil
}

// Validate checks BranchingCfg constraints.
func (b *BranchingCfg) Validate() error {
	if b.Avg < 1.5 || b.Avg > 3.0 {
		return fmt.Errorf("avg must be in range [1.5, 3.0], got %f", b.Avg)
	}
	if b.Max < 2 || b.Max > 5 {
		return fmt.Errorf("max must be in range [2, 5], got %d", b.Max)
	}
	return nil
}

// Validate checks PacingCfg constraints.
func (p *PacingCfg) Validate() error {
	// Validate curve type
	valid := false
	for _, validCurve := range ValidPacingCurves {
		if p.Curve == validCurve {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid curve type %q, must be one of: LINEAR, S_CURVE, EXPONENTIAL, CUSTOM", p.Curve)
	}

	// Validate variance
	if p.Variance < 0.0 || p.Variance > 0.3 {
		return fmt.Errorf("variance must be in range [0.0, 0.3], got %f", p.Variance)
	}

	// Validate custom points if CUSTOM curve
	if p.Curve == PacingCustom {
		if len(p.CustomPoints) < 2 {
			return errors.New("CUSTOM curve requires at least 2 custom points")
		}
		for i, point := range p.CustomPoints {
			if point[0] < 0.0 || point[0] > 1.0 {
				return fmt.Errorf("customPoints[%d]: progress must be in [0.0, 1.0], got %f", i, point[0])
			}
			if point[1] < 0.0 || point[1] > 1.0 {
				return fmt.Errorf("customPoints[%d]: difficulty must be in [0.0, 1.0], got %f", i, point[1])
			}
			// Ensure points are sorted by progress
			if i > 0 && point[0] <= p.CustomPoints[i-1][0] {
				return fmt.Errorf("customPoints[%d]: points must be sorted by progress", i)
			}
		}
	}

	return nil
}

// Validate checks KeyCfg constraints.
func (k *KeyCfg) Validate() error {
	if k.Name == "" {
		return errors.New("name must not be empty")
	}
	if k.Count < 1 || k.Count > 5 {
		return fmt.Errorf("count must be in range [1, 5], got %d", k.Count)
	}
	return nil
}

// Validate checks Constraint constraints.
func (c *Constraint) Validate() error {
	if c.Kind == "" {
		return errors.New("kind must not be empty")
	}
	if c.Severity != "hard" && c.Severity != "soft" {
		return fmt.Errorf("severity must be 'hard' or 'soft', got %q", c.Severity)
	}
	if c.Expr == "" {
		return errors.New("expr must not be empty")
	}
	return nil
}

// ToYAML serializes the config to YAML bytes.
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// Hash computes a deterministic hash of the configuration.
// Used for deriving per-stage RNG seeds.
func (c *Config) Hash() []byte {
	// For deterministic hashing, we serialize to YAML and hash that
	data, err := c.ToYAML()
	if err != nil {
		// Fallback: just hash the seed if YAML fails
		h := sha256.New()
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], c.Seed)
		h.Write(buf[:])
		return h.Sum(nil)
	}

	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

// generateSeed creates a seed from the current time.
// Uses nanosecond precision for better uniqueness.
func generateSeed() uint64 {
	now := time.Now().UnixNano()
	if now < 0 {
		now = -now
	}
	seed := uint64(now)
	// Ensure non-zero (though extremely unlikely with time-based seed)
	if seed == 0 {
		seed = 1
	}
	return seed
}
