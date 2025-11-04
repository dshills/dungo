# Fix for Non-Deterministic Room Count Generation

## Problem
Same seed produces different room counts (25 vs 28 vs 30) due to non-deterministic map iteration order in Go.

## Root Cause
Go maps iterate in random order by design. The codebase has multiple locations where `g.Rooms` (type `map[string]*Room`) is iterated without sorting keys first. This causes:
1. RNG calls in different orders
2. Different random values consumed
3. Different graph generation outcomes

## Files Requiring Fixes

### 1. pkg/synthesis/grammar.go
**Lines to fix:** 265, 556-565, 754, 586-589, 602-604, 610-614
- `applyExpandHub()` - fallback room selection
- `validateKeyLockConstraints()` - key provider and requirement loops
- `assignDifficulty()` - room difficulty assignment
- `findRoomsByArchetype()` - room finding helper
- `getAllRooms()` - unused but needs fix for consistency
- `getRoomsWithCapacity()` - capacity checking

### 2. pkg/synthesis/themes.go
**Lines to fix:** 41, 60, 111
- `assignSingleTheme()` - theme application
- `assignMultiTheme()` - seed selection (needs sort before shuffle)
- `assignMultiTheme()` - remaining room assignment

### 3. pkg/embedding/force_directed.go
**Line to fix:** 109
- `initializePositions()` - initial random positioning

### 4. pkg/content/loot.go
**Line to fix:** 183
- `findStartRoom()` - start room search

### 5. pkg/embedding/layout.go
**Lines to fix:** 266, 272
- `Validate()` - validation loops (cosmetic, but good practice)

## Solution Pattern
For every `for _, room := range g.Rooms` or `for roomID := range g.Rooms`:

```go
// Sort room IDs to ensure deterministic iteration order
roomIDs := make([]string, 0, len(g.Rooms))
for id := range g.Rooms {
    roomIDs = append(roomIDs, id)
}
sort.Strings(roomIDs)

// Then iterate over sorted IDs
for _, roomID := range roomIDs {
    room := g.Rooms[roomID]
    // ... use room
}
```

## Testing
After applying fixes:
- `go test -count=50 -run TestGrammarSynthesizer_Determinism ./pkg/synthesis` - PASS
- `go test -count=50 -run TestGolden_Determinism ./pkg/dungeon` - PASS
- `go test -count=50 -run TestGolden_Determinism ./test/integration` - PASS

## Impact
- **Performance:** Negligible - sorting small lists of room IDs
- **Memory:** Minimal - temporary slice for sorted IDs
- **Behavior:** Identical output for same seed (deterministic)
- **Compatibility:** None - internal implementation detail

## Notes
- Always add `"sort"` to imports when applying these fixes
- The linter may remove unused imports - ensure sort is actually used
- This pattern should be applied to ANY map iteration where order matters for RNG determinism
- Consider adding a lint rule to detect unsorted map iterations in RNG-dependent code
