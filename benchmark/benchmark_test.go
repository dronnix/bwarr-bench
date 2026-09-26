package benchmark

import (
	"math"
	"testing"
)

const (
	size100K = 100_000
	size250K = 250_000
	size500K = 500_000
	size1M   = 1_000_000
	size2M   = 2_000_000
	size4M   = 4_000_000
)

// sizeCase is one dataset size for the table-driven benchmarks below.
type sizeCase struct {
	name string
	n    int
}

// standardSizes drives every benchmark in this file.
func standardSizes() []sizeCase {
	return []sizeCase{
		{"100K", size100K}, {"250K", size250K}, {"500K", size500K},
		{"1M", size1M}, {"2M", size2M}, {"4M", size4M},
	}
}

// randomDataset is the default dataset: unique random int64 values over the full range.
func randomDataset(n int) []int64 {
	return GenerateRandomDataset(n, Seed, math.MaxInt64)
}

// benchDataset runs f as a sub-benchmark per standard size with a dataset from gen.
// Sub-benchmark names are the sizes, so `-bench='Insert/1M$'` selects one point.
func benchDataset(b *testing.B, f Func, gen func(n int) []int64) {
	b.Helper()
	for _, s := range standardSizes() {
		b.Run(s.name, func(b *testing.B) {
			f(b, Params{ElementsToApply: s.n, InitValues: gen(s.n)})
		})
	}
}

// benchValues runs f per standard size with the default random dataset.
func benchValues(b *testing.B, f Func) {
	b.Helper()
	benchDataset(b, f, randomDataset)
}

// Case 1: Insert, duplicates allowed. btree has no multiset Insert and is measured
// with ReplaceOrInsert, its only insert.

func BenchmarkBWArr_Insert(b *testing.B) { benchValues(b, BenchBWArrInsert) }
func BenchmarkBTree_Insert(b *testing.B) { benchValues(b, BenchBTreeInsert) }

// Case 2: ReplaceOrInsert, unique collection. The btree side is the same function as
// in case 1; it is listed under a matching name so `-bench=ReplaceOrInsert` runs both.

func BenchmarkBWArr_ReplaceOrInsert(b *testing.B) { benchValues(b, BenchBWArrReplaceOrInsert) }
func BenchmarkBTree_ReplaceOrInsert(b *testing.B) { benchValues(b, BenchBTreeInsert) }

// Case 3: Insert on sorted input, keys arriving in increasing or decreasing order.
// Same operations as case 1.

func BenchmarkBWArr_InsertIncreasing(b *testing.B) {
	benchDataset(b, BenchBWArrInsert, GenerateIncreasingDataset)
}

func BenchmarkBTree_InsertIncreasing(b *testing.B) {
	benchDataset(b, BenchBTreeInsert, GenerateIncreasingDataset)
}

func BenchmarkBWArr_InsertDecreasing(b *testing.B) {
	benchDataset(b, BenchBWArrInsert, GenerateDecreasingDataset)
}

func BenchmarkBTree_InsertDecreasing(b *testing.B) {
	benchDataset(b, BenchBTreeInsert, GenerateDecreasingDataset)
}

// Get: look up every value by key in a pre-populated structure.

func BenchmarkBWArr_Get(b *testing.B) { benchValues(b, BenchBWArrGet) }
func BenchmarkBTree_Get(b *testing.B) { benchValues(b, BenchBTreeGet) }

// Ordered iteration over all values.

func BenchmarkBWArr_OrderedIterate(b *testing.B) { benchValues(b, BenchBWArrOrderedIterate) }
func BenchmarkBTree_OrderedIterate(b *testing.B) { benchValues(b, BenchBTreeOrderedIterate) }

// Unordered iteration over all values. btree has no unordered walk, so its side is
// Ascend, listed under a matching name so `-bench=UnorderedIterate` runs both.

func BenchmarkBWArr_UnorderedIterate(b *testing.B) { benchValues(b, BenchBWArrUnorderedIterate) }
func BenchmarkBTree_UnorderedIterate(b *testing.B) { benchValues(b, BenchBTreeOrderedIterate) }

// Delete all values from a pre-populated structure.

func BenchmarkBWArr_Delete(b *testing.B) { benchValues(b, BenchBWArrDelete) }
func BenchmarkBTree_Delete(b *testing.B) { benchValues(b, BenchBTreeDelete) }
