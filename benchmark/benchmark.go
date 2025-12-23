package benchmark

import "math/rand"

// generateRandomInt64s generates a slice of random int64 values using the given seed.
// This function is kept for use in unit tests.
func generateRandomInt64s(count int, seed int64) []int64 {
	rng := rand.New(rand.NewSource(seed))
	values := make([]int64, count)
	for i := 0; i < count; i++ {
		values[i] = rng.Int63()
	}
	return values
}
