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

// standardSizes drives the table-driven benchmarks for the newer cases below.
func standardSizes() []sizeCase {
	return []sizeCase{
		{"100K", size100K}, {"250K", size250K}, {"500K", size500K},
		{"1M", size1M}, {"2M", size2M}, {"4M", size4M},
	}
}

// benchValues runs f as a sub-benchmark per standard size with a unique int64 dataset.
func benchValues(b *testing.B, f Func) {
	b.Helper()
	for _, s := range standardSizes() {
		b.Run(s.name, func(b *testing.B) {
			f(b, Params{ElementsToApply: s.n, InitValues: GenerateRandomDataset(s.n, Seed, math.MaxInt64)})
		})
	}
}

// benchDataset runs f as a sub-benchmark per standard size with a dataset from gen.
func benchDataset(b *testing.B, f Func, gen func(n int) []int64) {
	b.Helper()
	for _, s := range standardSizes() {
		b.Run(s.name, func(b *testing.B) {
			f(b, Params{ElementsToApply: s.n, InitValues: gen(s.n)})
		})
	}
}

// Insert benchmarks on sorted input: keys arrive in increasing or decreasing order
// (case 3). Same operations as case 1: bwarr Insert vs btree ReplaceOrInsert.

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

// ReplaceOrInsert benchmarks (unique collection). btree has only ReplaceOrInsert, so its
// side is the same function as BenchmarkBTree_Insert_*; it is listed here under a matching
// name so `go test -bench=ReplaceOrInsert` runs both implementations.

func BenchmarkBWArr_ReplaceOrInsert(b *testing.B) { benchValues(b, BenchBWArrReplaceOrInsert) }
func BenchmarkBTree_ReplaceOrInsert(b *testing.B) { benchValues(b, BenchBTreeInsert) }

// Insert benchmarks - BWArr

func BenchmarkBWArr_Insert_100K(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_250K(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_500K(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_1M(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_2M(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_4M(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Insert benchmarks - BTree

func BenchmarkBTree_Insert_100K(b *testing.B) {
	BenchBTreeInsert(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_250K(b *testing.B) {
	BenchBTreeInsert(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_500K(b *testing.B) {
	BenchBTreeInsert(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_1M(b *testing.B) {
	BenchBTreeInsert(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_2M(b *testing.B) {
	BenchBTreeInsert(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_4M(b *testing.B) {
	BenchBTreeInsert(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Get benchmarks - BWArr

func BenchmarkBWArr_Get_100K(b *testing.B) {
	BenchBWArrGet(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Get_250K(b *testing.B) {
	BenchBWArrGet(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Get_500K(b *testing.B) {
	BenchBWArrGet(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Get_1M(b *testing.B) {
	BenchBWArrGet(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Get_2M(b *testing.B) {
	BenchBWArrGet(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Get_4M(b *testing.B) {
	BenchBWArrGet(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Get benchmarks - BTree

func BenchmarkBTree_Get_100K(b *testing.B) {
	BenchBTreeGet(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Get_250K(b *testing.B) {
	BenchBTreeGet(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Get_500K(b *testing.B) {
	BenchBTreeGet(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Get_1M(b *testing.B) {
	BenchBTreeGet(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Get_2M(b *testing.B) {
	BenchBTreeGet(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Get_4M(b *testing.B) {
	BenchBTreeGet(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Ordered iteration benchmarks - BWArr

func BenchmarkBWArr_OrderedIterate_100K(b *testing.B) {
	BenchBWArrOrderedIterate(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_OrderedIterate_250K(b *testing.B) {
	BenchBWArrOrderedIterate(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_OrderedIterate_500K(b *testing.B) {
	BenchBWArrOrderedIterate(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_OrderedIterate_1M(b *testing.B) {
	BenchBWArrOrderedIterate(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_OrderedIterate_2M(b *testing.B) {
	BenchBWArrOrderedIterate(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_OrderedIterate_4M(b *testing.B) {
	BenchBWArrOrderedIterate(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Ordered iteration benchmarks - BTree

func BenchmarkBTree_OrderedIterate_100K(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_OrderedIterate_250K(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_OrderedIterate_500K(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_OrderedIterate_1M(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_OrderedIterate_2M(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_OrderedIterate_4M(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Unordered iteration benchmarks - BWArr

func BenchmarkBWArr_UnorderedIterate_100K(b *testing.B) {
	BenchBWArrUnorderedIterate(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_UnorderedIterate_250K(b *testing.B) {
	BenchBWArrUnorderedIterate(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_UnorderedIterate_500K(b *testing.B) {
	BenchBWArrUnorderedIterate(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_UnorderedIterate_1M(b *testing.B) {
	BenchBWArrUnorderedIterate(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_UnorderedIterate_2M(b *testing.B) {
	BenchBWArrUnorderedIterate(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_UnorderedIterate_4M(b *testing.B) {
	BenchBWArrUnorderedIterate(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Unordered iteration benchmarks - BTree (uses ordered since BTree doesn't have unordered)

func BenchmarkBTree_UnorderedIterate_100K(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_UnorderedIterate_250K(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_UnorderedIterate_500K(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_UnorderedIterate_1M(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_UnorderedIterate_2M(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_UnorderedIterate_4M(b *testing.B) {
	BenchBTreeOrderedIterate(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Delete benchmarks - BWArr

func BenchmarkBWArr_Delete_100K(b *testing.B) {
	BenchBWArrDelete(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Delete_250K(b *testing.B) {
	BenchBWArrDelete(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Delete_500K(b *testing.B) {
	BenchBWArrDelete(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Delete_1M(b *testing.B) {
	BenchBWArrDelete(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Delete_2M(b *testing.B) {
	BenchBWArrDelete(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Delete_4M(b *testing.B) {
	BenchBWArrDelete(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}

// Delete benchmarks - BTree

func BenchmarkBTree_Delete_100K(b *testing.B) {
	BenchBTreeDelete(b, Params{ElementsToApply: size100K, InitValues: GenerateRandomDataset(size100K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Delete_250K(b *testing.B) {
	BenchBTreeDelete(b, Params{ElementsToApply: size250K, InitValues: GenerateRandomDataset(size250K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Delete_500K(b *testing.B) {
	BenchBTreeDelete(b, Params{ElementsToApply: size500K, InitValues: GenerateRandomDataset(size500K, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Delete_1M(b *testing.B) {
	BenchBTreeDelete(b, Params{ElementsToApply: size1M, InitValues: GenerateRandomDataset(size1M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Delete_2M(b *testing.B) {
	BenchBTreeDelete(b, Params{ElementsToApply: size2M, InitValues: GenerateRandomDataset(size2M, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Delete_4M(b *testing.B) {
	BenchBTreeDelete(b, Params{ElementsToApply: size4M, InitValues: GenerateRandomDataset(size4M, Seed, math.MaxInt64)})
}
