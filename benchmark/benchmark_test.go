package benchmark

import (
	"math"
	"testing"
)

// Benchmark functions for bwarr with different dataset sizes

func BenchmarkBWArr_Insert_1024(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: 1024, InitValues: GenerateRandomDataset(1024, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_2048(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: 2048, InitValues: GenerateRandomDataset(2048, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_4096(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: 4096, InitValues: GenerateRandomDataset(4096, Seed, math.MaxInt64)})
}

func BenchmarkBWArr_Insert_8192(b *testing.B) {
	BenchBWArrInsert(b, Params{ElementsToApply: 8192, InitValues: GenerateRandomDataset(8192, Seed, math.MaxInt64)})
}

// Benchmark functions for btree with different dataset sizes

func BenchmarkBTree_Insert_1024(b *testing.B) {
	BenchBtreeInsert(b, Params{ElementsToApply: 1024, InitValues: GenerateRandomDataset(1024, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_2048(b *testing.B) {
	BenchBtreeInsert(b, Params{ElementsToApply: 2048, InitValues: GenerateRandomDataset(2048, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_4096(b *testing.B) {
	BenchBtreeInsert(b, Params{ElementsToApply: 4096, InitValues: GenerateRandomDataset(4096, Seed, math.MaxInt64)})
}

func BenchmarkBTree_Insert_8192(b *testing.B) {
	BenchBtreeInsert(b, Params{ElementsToApply: 8192, InitValues: GenerateRandomDataset(8192, Seed, math.MaxInt64)})
}
