# Example 3: Dark Souls Challenge Dungeon

This example generates a brutally difficult dungeon inspired by Dark Souls, featuring:

## Features

⚔️ **Custom Difficulty Curve**
- Easy intro (20% difficulty)
- **BRUTAL mid-game (85-90% difficulty!)**
- Slightly easier boss (75% - you've earned it)
- High variance for unpredictability

💀 **Punishing Design**
- Most content is mandatory (85% critical path)
- Limited keys force careful exploration
- Long critical path (15+ rooms)
- Peak difficulty before the boss

🏰 **Dark Atmosphere**
- Pure crypt theme
- Oppressive and claustrophobic
- Hidden secrets offer crucial advantages

🎯 **Strategic Depth**
- 20% secret rooms with powerful rewards
- Careful routing required to survive
- Finding secrets is essential, not optional

## What This Showcases

✅ **Custom Pacing Curves**: The CUSTOM curve type lets you define exact difficulty at specific progress points. This creates a unique mid-game challenge spike!

✅ **Fine-Tuned Constraints**: Multiple soft constraints shape the dungeon structure to match the design vision.

✅ **High Variance**: 0.25 variance means actual room difficulty can vary ±25%, creating unpredictable encounters.

✅ **Strategic Design**: By making most content mandatory and secrets valuable, the generator creates meaningful choices.

## How to Run

**From the example directory:**
```bash
cd examples/03-darksouls-challenge
go run main.go
```

**From project root:**
```bash
make run-darksouls
```

**Using the binary:**
```bash
make examples
./bin/darksouls-challenge -config examples/03-darksouls-challenge/config.yaml
```

The output includes visual difficulty bars showing the custom curve:

```
📈 Difficulty Curve:
    0% progress: ░░░░········ 20%
   20% progress: ░░░░░░░░░░·· 50%
   40% progress: ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓· 85%
        └─ ☠️  BRUTAL SECTION
   60% progress: ███████████████████ 90%
        └─ ☠️  BRUTAL SECTION
   80% progress: ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓·· 80%
  100% progress: ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓·· 75%
```

## Try It Out

**Different challenge levels:**
```bash
# Change the seed for different layouts
sed -i '' 's/seed: 666/seed: 777/' config.yaml
go run main.go
```

**Extreme mode:**
```yaml
pacing:
  curve: "EXPONENTIAL"  # Just... why?
  variance: 0.3
secretDensity: 0.1     # Fewer safe havens
```

**More merciful:**
```yaml
customPoints:
  - [0.0, 0.3]
  - [0.3, 0.6]
  - [0.6, 0.7]
  - [1.0, 0.8]  # Boss is hardest
secretDensity: 0.3
```

**Speedrun mode:**
```yaml
size:
  roomsMin: 20
  roomsMax: 25
optionalRatio: 0.1  # Minimal side content
```

## Custom Curve Explained

Each point in the `customPoints` array is `[progress, difficulty]`:

- **Progress**: 0.0 = start room, 1.0 = boss room
- **Difficulty**: 0.0 = trivial, 1.0 = maximum

The generator interpolates between points and uses `variance` to add randomness.

**Example interpretation:**
```yaml
customPoints:
  - [0.0, 0.2]   # Start: Easy (tutorial vibes)
  - [0.4, 0.85]  # 40% in: VERY HARD (first major challenge)
  - [1.0, 0.75]  # Boss: Hard but fair
```

## Use Cases

- **Challenge runs** with specific difficulty profiles
- **Boss rush modes** with custom ramping
- **Tutorial → Hard → Boss** structured campaigns
- **Roguelike permadeath** where difficulty matters
- **Testing player skill** with known challenge curves
- **Competitive leaderboards** with deterministic seeds

## The Philosophy

This example shows how dungo isn't just about random generation—it's about **controlled procedural design**. By specifying exact difficulty curves, you maintain creative control while still getting the benefits of procedural variation.

Perfect for when you want "designed chaos" rather than pure randomness.

---

**"Prepare to die."** 💀
