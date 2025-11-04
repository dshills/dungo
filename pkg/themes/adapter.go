package themes

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dshills/dungo/pkg/rng"
)

// Loader provides cached loading of theme packs from a base directory.
// This adapter provides the interface needed by content placement logic.
type Loader struct {
	baseDir string
	cache   map[string]*ThemePack
	mu      sync.RWMutex
}

// NewLoader creates a theme pack loader for the given base directory.
func NewLoader(baseDir string) *Loader {
	return &Loader{
		baseDir: baseDir,
		cache:   make(map[string]*ThemePack),
	}
}

// Load loads a theme pack by name from baseDir/<name>/theme.yml.
// Results are cached for subsequent loads.
func (l *Loader) Load(name string) (*ThemePack, error) {
	// Validate name to prevent path traversal attacks
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return nil, fmt.Errorf("invalid theme name: %s", name)
	}

	// Check cache with read lock
	l.mu.RLock()
	if pack, ok := l.cache[name]; ok {
		l.mu.RUnlock()
		return pack, nil
	}
	l.mu.RUnlock()

	// Load from disk using secure path joining
	themePath := filepath.Join(l.baseDir, name)
	pack, err := LoadThemeFromDirectory(themePath)
	if err != nil {
		return nil, err
	}

	// Cache with write lock
	l.mu.Lock()
	l.cache[name] = pack
	l.mu.Unlock()

	return pack, nil
}

// SelectEncounterFromTheme selects an enemy type from theme encounter tables.
// Returns empty string if theme is nil or no encounters match.
func SelectEncounterFromTheme(theme *ThemePack, difficulty float64, rng *rng.RNG) string {
	if theme == nil {
		return ""
	}

	// Get encounter tables for difficulty
	tables := theme.GetEncountersForDifficulty(difficulty)
	if len(tables) == 0 {
		return ""
	}

	// Convert to standard rand.Rand for SelectWeightedEntry
	stdRand := rand.New(rand.NewSource(int64(rng.Uint64())))

	// Collect all entries from matching tables
	var allEntries []WeightedEntry
	for _, table := range tables {
		allEntries = append(allEntries, table.Entries...)
	}

	// Select from combined entries
	entry := SelectWeightedEntry(allEntries, stdRand)
	if entry == nil {
		return ""
	}

	return entry.Type
}

// SelectLootFromTheme selects an item type from theme loot tables.
// roomType is the room archetype (e.g., "treasure", "boss").
// Returns empty string if theme is nil or no loot table matches.
func SelectLootFromTheme(theme *ThemePack, roomType string, rng *rng.RNG) string {
	if theme == nil {
		return ""
	}

	// Get loot table for room type
	table := theme.GetLootTableForRoomType(roomType)
	if table == nil {
		return ""
	}

	// Convert to standard rand.Rand for SelectWeightedEntry
	stdRand := rand.New(rand.NewSource(int64(rng.Uint64())))

	// Select from loot entries
	entry := SelectWeightedEntry(table.Entries, stdRand)
	if entry == nil {
		return ""
	}

	return entry.Type
}
