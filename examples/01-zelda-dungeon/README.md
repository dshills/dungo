# Example 1: Classic Zelda-Style Dungeon

This example generates a dungeon inspired by classic Zelda games, featuring:

## Features

🗝️ **Key/Lock Progression**
- 3 small keys for basic locked doors
- 1 big key for the treasure room
- 1 boss key for the final boss chamber

🎮 **Classic Structure**
- Linear progression with some exploration
- S-curve difficulty (starts easy, gets hard, manageable boss)
- 20-25 interconnected rooms
- Moderate branching (avg 2.0 connections per room)

🎨 **Atmospheric Themes**
- Crypt aesthetic with arcane elements
- Suitable for fantasy dungeon crawlers

🏆 **Replayability**
- 15% secret rooms with bonus loot
- 20% optional treasure rooms off the critical path
- Deterministic generation: same seed = same dungeon

## What This Showcases

✅ **Key/Lock System**: The dungeon enforces a key-before-lock constraint, ensuring all keys are reachable before their corresponding locks. This prevents softlocks and guarantees solvability.

✅ **Pacing Curves**: Uses an S-curve for difficulty progression, which feels natural—easy intro, challenging middle, manageable boss.

✅ **Determinism**: The same seed always generates the identical dungeon, perfect for challenge runs or testing.

## How to Run

**From the example directory:**
```bash
cd examples/01-zelda-dungeon
go run main.go
```

**From project root:**
```bash
make run-zelda
```

**Using the binary:**
```bash
make examples
./bin/zelda-dungeon -config examples/01-zelda-dungeon/config.yaml
```

This will generate three files:
- `zelda_dungeon_<seed>.json` - Full artifact data
- `zelda_dungeon_<seed>.tmj` - Tiled Map Editor format
- `zelda_dungeon_<seed>.svg` - Visual graph representation

## Try It Out

**Experiment with different seeds:**
```yaml
# Edit config.yaml
seed: 12345  # Your lucky number!
```

**Make it harder:**
```yaml
pacing:
  curve: "EXPONENTIAL"  # Brutal difficulty spike!
  variance: 0.2
```

**More exploration:**
```yaml
branching:
  avg: 2.5
  max: 4
secretDensity: 0.25  # 25% secret rooms!
```

## Use Cases

- **Roguelike games** needing structured progression
- **Puzzle dungeons** with gated advancement
- **Tutorial levels** with controlled difficulty
- **Challenge modes** with fixed seeds for leaderboards
