package rng_test

import (
	"crypto/sha256"
	"fmt"

	"github.com/dshills/dungo/pkg/rng"
)

// ExampleNewRNG demonstrates creating a deterministic RNG for a pipeline stage.
func ExampleNewRNG() {
	// Master seed for the entire generation
	masterSeed := uint64(123456789)

	// Each pipeline stage gets its own RNG
	configHash := sha256.Sum256([]byte("dungeon_config_v1"))

	// Create RNGs for different stages
	graphRNG := rng.NewRNG(masterSeed, "graph_synthesis", configHash[:])
	embedRNG := rng.NewRNG(masterSeed, "embedding", configHash[:])

	// Each stage produces independent but deterministic sequences
	fmt.Printf("Graph stage seed: %d\n", graphRNG.Seed())
	fmt.Printf("Embed stage seed: %d\n", embedRNG.Seed())
	fmt.Printf("Graph first value: %d\n", graphRNG.Intn(100))
	fmt.Printf("Embed first value: %d\n", embedRNG.Intn(100))

	// Same inputs produce same results
	graphRNG2 := rng.NewRNG(masterSeed, "graph_synthesis", configHash[:])
	fmt.Printf("Graph repeated: %d\n", graphRNG2.Intn(100))

	// Output:
	// Graph stage seed: 10126480545457960121
	// Embed stage seed: 11758735888959734649
	// Graph first value: 11
	// Embed first value: 74
	// Graph repeated: 11
}

// ExampleRNG_Shuffle demonstrates deterministic shuffling.
func ExampleRNG_Shuffle() {
	masterSeed := uint64(42)
	configHash := sha256.Sum256([]byte("config"))
	rng := rng.NewRNG(masterSeed, "content_placement", configHash[:])

	// Shuffle room order deterministically
	rooms := []string{"Start", "Treasure", "Boss", "Hub", "Secret"}
	rng.Shuffle(len(rooms), func(i, j int) {
		rooms[i], rooms[j] = rooms[j], rooms[i]
	})

	fmt.Printf("Shuffled rooms: %v\n", rooms)

	// Output:
	// Shuffled rooms: [Boss Hub Treasure Start Secret]
}

// ExampleRNG_WeightedChoice demonstrates weighted random selection.
func ExampleRNG_WeightedChoice() {
	masterSeed := uint64(999)
	configHash := sha256.Sum256([]byte("config"))
	rng := rng.NewRNG(masterSeed, "loot_generation", configHash[:])

	// Loot rarity weights: [common, uncommon, rare, legendary]
	weights := []float64{50.0, 30.0, 15.0, 5.0}

	// Generate 10 items
	rarities := []string{"common", "uncommon", "rare", "legendary"}
	for i := 0; i < 10; i++ {
		choice := rng.WeightedChoice(weights)
		fmt.Printf("Item %d: %s\n", i+1, rarities[choice])
	}

	// Output:
	// Item 1: common
	// Item 2: rare
	// Item 3: common
	// Item 4: uncommon
	// Item 5: common
	// Item 6: uncommon
	// Item 7: common
	// Item 8: common
	// Item 9: common
	// Item 10: common
}

// ExampleRNG_Float64Range demonstrates generating difficulty values.
func ExampleRNG_Float64Range() {
	masterSeed := uint64(777)
	configHash := sha256.Sum256([]byte("config"))
	rng := rng.NewRNG(masterSeed, "difficulty_scaling", configHash[:])

	// Generate difficulty values for 5 rooms
	for i := 0; i < 5; i++ {
		difficulty := rng.Float64Range(0.3, 0.8)
		fmt.Printf("Room %d difficulty: %.2f\n", i+1, difficulty)
	}

	// Output:
	// Room 1 difficulty: 0.74
	// Room 2 difficulty: 0.73
	// Room 3 difficulty: 0.43
	// Room 4 difficulty: 0.42
	// Room 5 difficulty: 0.56
}
