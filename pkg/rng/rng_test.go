package rng

import (
	"crypto/sha256"
	"encoding/binary"
	"testing"
)

// TestNewRNG_Determinism verifies that the same inputs always produce the same RNG.
func TestNewRNG_Determinism(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test_stage"
	configHash := sha256.Sum256([]byte("test_config"))

	// Create two RNGs with identical inputs
	rng1 := NewRNG(masterSeed, stageName, configHash[:])
	rng2 := NewRNG(masterSeed, stageName, configHash[:])

	// Verify they have the same derived seed
	if rng1.Seed() != rng2.Seed() {
		t.Errorf("Same inputs produced different seeds: %d vs %d", rng1.Seed(), rng2.Seed())
	}

	// Verify they produce the same sequence
	for i := 0; i < 100; i++ {
		v1 := rng1.Uint64()
		v2 := rng2.Uint64()
		if v1 != v2 {
			t.Errorf("Iteration %d: Same RNGs produced different values: %d vs %d", i, v1, v2)
		}
	}
}

// TestNewRNG_SequenceDeterminism verifies the entire sequence is reproducible.
func TestNewRNG_SequenceDeterminism(t *testing.T) {
	masterSeed := uint64(987654321)
	stageName := "graph_synthesis"
	configHash := sha256.Sum256([]byte("config_v1"))

	// Generate first sequence
	rng1 := NewRNG(masterSeed, stageName, configHash[:])
	sequence1 := make([]uint64, 50)
	for i := range sequence1 {
		sequence1[i] = rng1.Uint64()
	}

	// Generate second sequence with same inputs
	rng2 := NewRNG(masterSeed, stageName, configHash[:])
	sequence2 := make([]uint64, 50)
	for i := range sequence2 {
		sequence2[i] = rng2.Uint64()
	}

	// Verify sequences match exactly
	for i := range sequence1 {
		if sequence1[i] != sequence2[i] {
			t.Errorf("Position %d: sequences differ: %d vs %d", i, sequence1[i], sequence2[i])
		}
	}
}

// TestNewRNG_DifferentStages verifies different stage names produce different sequences.
func TestNewRNG_DifferentStages(t *testing.T) {
	masterSeed := uint64(123456789)
	configHash := sha256.Sum256([]byte("same_config"))

	rng1 := NewRNG(masterSeed, "graph_synthesis", configHash[:])
	rng2 := NewRNG(masterSeed, "embedding", configHash[:])
	rng3 := NewRNG(masterSeed, "carving", configHash[:])

	// Verify different derived seeds
	if rng1.Seed() == rng2.Seed() {
		t.Error("Different stages produced identical seeds")
	}
	if rng1.Seed() == rng3.Seed() {
		t.Error("Different stages produced identical seeds")
	}
	if rng2.Seed() == rng3.Seed() {
		t.Error("Different stages produced identical seeds")
	}

	// Verify stage names are preserved
	if rng1.StageName() != "graph_synthesis" {
		t.Errorf("Stage name not preserved: got %s", rng1.StageName())
	}

	// Generate sequences and verify they differ
	v1 := rng1.Uint64()
	v2 := rng2.Uint64()
	v3 := rng3.Uint64()

	if v1 == v2 && v2 == v3 {
		t.Error("Different stages produced identical first values (extremely unlikely)")
	}
}

// TestNewRNG_DifferentConfigs verifies different config hashes produce different sequences.
func TestNewRNG_DifferentConfigs(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test_stage"

	config1Hash := sha256.Sum256([]byte("config_v1"))
	config2Hash := sha256.Sum256([]byte("config_v2"))
	config3Hash := sha256.Sum256([]byte("config_v3"))

	rng1 := NewRNG(masterSeed, stageName, config1Hash[:])
	rng2 := NewRNG(masterSeed, stageName, config2Hash[:])
	rng3 := NewRNG(masterSeed, stageName, config3Hash[:])

	// Verify different derived seeds
	if rng1.Seed() == rng2.Seed() {
		t.Error("Different configs produced identical seeds")
	}
	if rng1.Seed() == rng3.Seed() {
		t.Error("Different configs produced identical seeds")
	}
	if rng2.Seed() == rng3.Seed() {
		t.Error("Different configs produced identical seeds")
	}

	// Verify they produce different sequences
	v1 := rng1.Uint64()
	v2 := rng2.Uint64()
	v3 := rng3.Uint64()

	if v1 == v2 && v2 == v3 {
		t.Error("Different configs produced identical first values (extremely unlikely)")
	}
}

// TestNewRNG_DifferentMasterSeeds verifies different master seeds produce different sequences.
func TestNewRNG_DifferentMasterSeeds(t *testing.T) {
	stageName := "test_stage"
	configHash := sha256.Sum256([]byte("same_config"))

	rng1 := NewRNG(uint64(111), stageName, configHash[:])
	rng2 := NewRNG(uint64(222), stageName, configHash[:])
	rng3 := NewRNG(uint64(333), stageName, configHash[:])

	// Verify different derived seeds
	if rng1.Seed() == rng2.Seed() {
		t.Error("Different master seeds produced identical seeds")
	}
	if rng1.Seed() == rng3.Seed() {
		t.Error("Different master seeds produced identical seeds")
	}
	if rng2.Seed() == rng3.Seed() {
		t.Error("Different master seeds produced identical seeds")
	}
}

// TestRNG_Intn verifies Intn produces values in correct range and is deterministic.
func TestRNG_Intn(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))

	rng := NewRNG(masterSeed, stageName, configHash[:])

	// Test range bounds
	for i := 0; i < 100; i++ {
		v := rng.Intn(10)
		if v < 0 || v >= 10 {
			t.Errorf("Intn(10) produced out-of-range value: %d", v)
		}
	}

	// Test determinism
	rng1 := NewRNG(masterSeed, stageName, configHash[:])
	rng2 := NewRNG(masterSeed, stageName, configHash[:])

	for i := 0; i < 50; i++ {
		v1 := rng1.Intn(100)
		v2 := rng2.Intn(100)
		if v1 != v2 {
			t.Errorf("Iteration %d: Intn not deterministic: %d vs %d", i, v1, v2)
		}
	}
}

// TestRNG_IntnPanic verifies Intn panics on invalid input.
func TestRNG_IntnPanic(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))
	rng := NewRNG(masterSeed, stageName, configHash[:])

	defer func() {
		if r := recover(); r == nil {
			t.Error("Intn(0) did not panic")
		}
	}()

	rng.Intn(0)
}

// TestRNG_Float64 verifies Float64 produces values in [0, 1) and is deterministic.
func TestRNG_Float64(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))

	rng := NewRNG(masterSeed, stageName, configHash[:])

	// Test range bounds
	for i := 0; i < 100; i++ {
		v := rng.Float64()
		if v < 0.0 || v >= 1.0 {
			t.Errorf("Float64() produced out-of-range value: %f", v)
		}
	}

	// Test determinism
	rng1 := NewRNG(masterSeed, stageName, configHash[:])
	rng2 := NewRNG(masterSeed, stageName, configHash[:])

	for i := 0; i < 50; i++ {
		v1 := rng1.Float64()
		v2 := rng2.Float64()
		if v1 != v2 {
			t.Errorf("Iteration %d: Float64 not deterministic: %f vs %f", i, v1, v2)
		}
	}
}

// TestRNG_Shuffle verifies Shuffle produces deterministic permutations.
func TestRNG_Shuffle(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))

	// Create first shuffled sequence
	rng1 := NewRNG(masterSeed, stageName, configHash[:])
	slice1 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	rng1.Shuffle(len(slice1), func(i, j int) {
		slice1[i], slice1[j] = slice1[j], slice1[i]
	})

	// Create second shuffled sequence with same seed
	rng2 := NewRNG(masterSeed, stageName, configHash[:])
	slice2 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	rng2.Shuffle(len(slice2), func(i, j int) {
		slice2[i], slice2[j] = slice2[j], slice2[i]
	})

	// Verify identical shuffles
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			t.Errorf("Position %d: Shuffle not deterministic: %d vs %d", i, slice1[i], slice2[i])
		}
	}

	// Verify shuffle actually changed the order (extremely likely)
	allSame := true
	for i := range slice1 {
		if slice1[i] != i {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("Shuffle did not change order (extremely unlikely)")
	}
}

// TestRNG_IntRange verifies IntRange produces values in correct range.
func TestRNG_IntRange(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))

	rng := NewRNG(masterSeed, stageName, configHash[:])

	// Test various ranges
	for i := 0; i < 100; i++ {
		v := rng.IntRange(5, 10)
		if v < 5 || v > 10 {
			t.Errorf("IntRange(5, 10) produced out-of-range value: %d", v)
		}
	}

	// Test single value range
	for i := 0; i < 10; i++ {
		v := rng.IntRange(7, 7)
		if v != 7 {
			t.Errorf("IntRange(7, 7) produced wrong value: %d", v)
		}
	}
}

// TestRNG_IntRangePanic verifies IntRange panics on invalid input.
func TestRNG_IntRangePanic(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))
	rng := NewRNG(masterSeed, stageName, configHash[:])

	defer func() {
		if r := recover(); r == nil {
			t.Error("IntRange(10, 5) did not panic")
		}
	}()

	rng.IntRange(10, 5)
}

// TestRNG_Float64Range verifies Float64Range produces values in correct range.
func TestRNG_Float64Range(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))

	rng := NewRNG(masterSeed, stageName, configHash[:])

	// Test range bounds
	for i := 0; i < 100; i++ {
		v := rng.Float64Range(5.0, 10.0)
		if v < 5.0 || v >= 10.0 {
			t.Errorf("Float64Range(5.0, 10.0) produced out-of-range value: %f", v)
		}
	}
}

// TestRNG_Float64RangePanic verifies Float64Range panics on invalid input.
func TestRNG_Float64RangePanic(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))
	rng := NewRNG(masterSeed, stageName, configHash[:])

	defer func() {
		if r := recover(); r == nil {
			t.Error("Float64Range(10.0, 5.0) did not panic")
		}
	}()

	rng.Float64Range(10.0, 5.0)
}

// TestRNG_Bool verifies Bool produces deterministic boolean values.
func TestRNG_Bool(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))

	// Test determinism
	rng1 := NewRNG(masterSeed, stageName, configHash[:])
	rng2 := NewRNG(masterSeed, stageName, configHash[:])

	for i := 0; i < 50; i++ {
		v1 := rng1.Bool()
		v2 := rng2.Bool()
		if v1 != v2 {
			t.Errorf("Iteration %d: Bool not deterministic: %v vs %v", i, v1, v2)
		}
	}

	// Verify we get both true and false (extremely likely in 100 samples)
	rng3 := NewRNG(masterSeed, stageName, configHash[:])
	trueCount := 0
	falseCount := 0
	for i := 0; i < 100; i++ {
		if rng3.Bool() {
			trueCount++
		} else {
			falseCount++
		}
	}

	if trueCount == 0 || falseCount == 0 {
		t.Error("Bool() produced only one value across 100 samples (extremely unlikely)")
	}
}

// TestRNG_WeightedChoice verifies weighted random selection.
func TestRNG_WeightedChoice(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))

	tests := []struct {
		name    string
		weights []float64
		want    int // -1 for "should return -1"
	}{
		{"empty weights", []float64{}, -1},
		{"all zero weights", []float64{0, 0, 0}, -1},
		{"single weight", []float64{1.0}, 0},
		{"equal weights", []float64{1.0, 1.0, 1.0}, -2}, // -2 means "valid index"
		{"skewed weights", []float64{0.0, 10.0, 0.0}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := NewRNG(masterSeed, stageName, configHash[:])
			got := rng.WeightedChoice(tt.weights)

			if tt.want == -1 {
				if got != -1 {
					t.Errorf("WeightedChoice() = %d, want -1", got)
				}
			} else if tt.want >= 0 {
				if got != tt.want {
					t.Errorf("WeightedChoice() = %d, want %d", got, tt.want)
				}
			} else {
				// Valid index check
				if got < 0 || got >= len(tt.weights) {
					t.Errorf("WeightedChoice() = %d, want valid index [0, %d)", got, len(tt.weights))
				}
			}
		})
	}

	// Test determinism
	weights := []float64{1.0, 2.0, 3.0}
	rng1 := NewRNG(masterSeed, stageName, configHash[:])
	rng2 := NewRNG(masterSeed, stageName, configHash[:])

	for i := 0; i < 50; i++ {
		v1 := rng1.WeightedChoice(weights)
		v2 := rng2.WeightedChoice(weights)
		if v1 != v2 {
			t.Errorf("Iteration %d: WeightedChoice not deterministic: %d vs %d", i, v1, v2)
		}
	}
}

// TestRNG_WeightedChoicePanic verifies negative weights cause panic.
func TestRNG_WeightedChoicePanic(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test"
	configHash := sha256.Sum256([]byte("config"))
	rng := NewRNG(masterSeed, stageName, configHash[:])

	defer func() {
		if r := recover(); r == nil {
			t.Error("WeightedChoice with negative weights did not panic")
		}
	}()

	rng.WeightedChoice([]float64{1.0, -1.0, 2.0})
}

// TestSubSeedDerivationFormula verifies the exact derivation formula.
func TestSubSeedDerivationFormula(t *testing.T) {
	masterSeed := uint64(123456789)
	stageName := "test_stage"
	configHash := []byte{1, 2, 3, 4, 5}

	// Manually compute expected derived seed
	h := sha256.New()
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], masterSeed)
	h.Write(buf[:])
	h.Write([]byte(stageName))
	h.Write(configHash)
	hash := h.Sum(nil)
	expected := binary.BigEndian.Uint64(hash[:8])

	// Create RNG and verify it matches
	rng := NewRNG(masterSeed, stageName, configHash)
	if rng.Seed() != expected {
		t.Errorf("Derived seed mismatch: got %d, want %d", rng.Seed(), expected)
	}
}

// BenchmarkNewRNG measures RNG creation performance.
func BenchmarkNewRNG(b *testing.B) {
	masterSeed := uint64(123456789)
	stageName := "benchmark_stage"
	configHash := sha256.Sum256([]byte("benchmark_config"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRNG(masterSeed, stageName, configHash[:])
	}
}

// BenchmarkRNG_Uint64 measures Uint64 performance.
func BenchmarkRNG_Uint64(b *testing.B) {
	masterSeed := uint64(123456789)
	stageName := "benchmark"
	configHash := sha256.Sum256([]byte("config"))
	rng := NewRNG(masterSeed, stageName, configHash[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rng.Uint64()
	}
}

// BenchmarkRNG_Intn measures Intn performance.
func BenchmarkRNG_Intn(b *testing.B) {
	masterSeed := uint64(123456789)
	stageName := "benchmark"
	configHash := sha256.Sum256([]byte("config"))
	rng := NewRNG(masterSeed, stageName, configHash[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rng.Intn(100)
	}
}

// BenchmarkRNG_Float64 measures Float64 performance.
func BenchmarkRNG_Float64(b *testing.B) {
	masterSeed := uint64(123456789)
	stageName := "benchmark"
	configHash := sha256.Sum256([]byte("config"))
	rng := NewRNG(masterSeed, stageName, configHash[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rng.Float64()
	}
}
