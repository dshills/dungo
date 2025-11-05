# Example 2: Metroidvania-Style Labyrinth

This example generates a large, interconnected dungeon inspired by Metroidvania games, featuring:

## Features

🗺️ **Massive Exploration**
- 60-80 rooms to explore
- Multiple paths to every major objective
- High interconnectivity (avg 2.8 connections/room)
- Abundant loops and shortcuts

🔑 **Multi-Key Gating**
- Red, blue, and yellow keys for different areas
- Color-coded progression system
- Multiple keys of each color for flexibility

🎨 **Organic Atmosphere**
- Fungal caves with arcane mysteries
- Non-linear exploration
- Backtracking-friendly design

🏆 **Exploration Rewards**
- 25% secret rooms (highest of all examples!)
- 35% optional content off critical path
- Rewards curiosity and thorough exploration

## What This Showcases

✅ **Scale**: Demonstrates the library can handle large, complex dungeons (60-80 rooms) efficiently.

✅ **Topology**: High branching factor creates natural loops and alternate routes, preventing the "hallway simulator" problem.

✅ **Multiple Viable Paths**: Unlike linear dungeons, this has many ways to reach the boss, encouraging different playthroughs.

✅ **Performance**: Even with 80 rooms and high connectivity, generation completes in <100ms.

## How to Run

**From the example directory:**
```bash
cd examples/02-metroidvania-maze
go run main.go
```

**From project root:**
```bash
make run-metroidvania
```

**Using the binary:**
```bash
make examples
./bin/metroidvania-maze -config examples/02-metroidvania-maze/config.yaml
```

Output includes:
- Full artifact JSON
- Tiled-compatible TMJ file
- SVG visualization (may be complex—zoom in!)

## Key Differences from Example 1

| Aspect | Zelda Dungeon | Metroidvania Maze |
|--------|---------------|-------------------|
| Size | 20-25 rooms | 60-80 rooms |
| Branching | 2.0 avg | 2.8 avg (40% more) |
| Pacing | S-curve | Linear |
| Secrets | 15% | 25% |
| Optional | 20% | 35% |
| Style | Guided progression | Open exploration |

## Try It Out

**Make it HUGE:**
```yaml
size:
  roomsMin: 100
  roomsMax: 150
branching:
  avg: 3.0
  max: 5
```

**Maximum interconnectivity:**
```yaml
branching:
  avg: 2.9
  max: 5
constraints:
  - kind: "topology"
    severity: "soft"
    expr: "cycleDensity > 0.6"  # Tons of loops!
```

**Simpler exploration:**
```yaml
size:
  roomsMin: 40
  roomsMax: 50
branching:
  avg: 2.2  # Less overwhelming
secretDensity: 0.15
```

## Use Cases

- **Metroidvania games** needing exploration-heavy maps
- **Open-world dungeons** with non-linear progression
- **Roguelikes** wanting variety beyond hallway crawls
- **Procedural RPGs** with backtracking mechanics
- **Speedrun challenges** with route optimization potential
