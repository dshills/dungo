# Quick Start Guide

Three fun examples to get started with dungo!

## 🎮 Try Them Out

### 1. Zelda-Style Dungeon (Recommended First!)

Linear progression with keys and locks:

```bash
cd examples/01-zelda-dungeon
go run main.go
```

**You'll get:**
- 20-25 rooms with structured progression
- 3 types of keys (small, big, boss)
- S-curve difficulty
- Perfect for learning the basics

### 2. Metroidvania Maze

Large, interconnected exploration:

```bash
cd examples/02-metroidvania-maze
go run main.go
```

**You'll get:**
- 60-80 rooms (huge!)
- Multiple paths everywhere
- High secret density
- Perfect for open-world dungeons

### 3. Dark Souls Challenge

Brutal difficulty spike:

```bash
cd examples/03-darksouls-challenge
go run main.go
```

**You'll get:**
- Custom difficulty curve
- Visual difficulty bars
- 85-90% peak difficulty
- Perfect for challenge modes

## 📂 Output Files

Each example generates:
- `.json` - Full dungeon data
- `.tmj` - Tiled Map Editor format
- `.svg` - Visual graph (open in browser!)

Files are saved in the current working directory with absolute paths shown.

## 🔧 Running Binaries

After building (`make examples`), run binaries from project root:

```bash
./bin/zelda-dungeon -config examples/01-zelda-dungeon/config.yaml
./bin/metroidvania-maze -config examples/02-metroidvania-maze/config.yaml
./bin/darksouls-challenge -config examples/03-darksouls-challenge/config.yaml
```

## 🎨 Customize

Edit `config.yaml` in any example:

**Change seed:**
```yaml
seed: 12345  # Reproducible dungeon
```

**Adjust size:**
```yaml
size:
  roomsMin: 30
  roomsMax: 40
```

**More secrets:**
```yaml
secretDensity: 0.3
```

## 🔧 What's Next?

- Edit configs to experiment
- Check individual READMEs for details
- See main `examples/README.md` for feature comparison
- Read project docs for API usage

## 🎯 Which Example for Your Game?

| Your Game Type | Use This Example |
|----------------|------------------|
| Roguelike with progression | 01-zelda-dungeon |
| Metroidvania exploration | 02-metroidvania-maze |
| Challenge/permadeath mode | 03-darksouls-challenge |
| Just learning the library | 01-zelda-dungeon |

Happy dungeon generating! 🎮
