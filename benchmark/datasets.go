package benchmark

import "math/rand"

const (
	// BenchmarkSeed is the fixed seed used for generating reproducible datasets.
	BenchmarkSeed = 42

	// BTreeDegree is the degree parameter for btree (per user requirements).
	BTreeDegree = 2

	// BWarrCapacity is the initial capacity hint for bwarr (0 = auto).
	BWarrCapacity = 0
)

// Pre-generated datasets for benchmarking.
// These are initialized once at package load time to ensure
// identical operations across all benchmark runs.
var (
	Dataset1024 []int64
	Dataset2048 []int64
	Dataset4096 []int64
	Dataset8192 []int64
)

// init generates all datasets once at package initialization.
func init() {
	Dataset1024 = generateDataset(1024, BenchmarkSeed)
	Dataset2048 = generateDataset(2048, BenchmarkSeed)
	Dataset4096 = generateDataset(4096, BenchmarkSeed)
	Dataset8192 = generateDataset(8192, BenchmarkSeed)
}

// generateDataset creates a reproducible slice of random int64 values.
func generateDataset(count int, seed int64) []int64 {
	rng := rand.New(rand.NewSource(seed))
	values := make([]int64, count)
	for i := 0; i < count; i++ {
		values[i] = rng.Int63()
	}
	return values
}
