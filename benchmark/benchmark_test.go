package benchmark

import (
	"testing"
)

// Benchmark functions for bwarr with different dataset sizes

func BenchmarkBWArr_Insert_1024(b *testing.B) {
	BenchBWArrInsert(b, Params{N: 1024, InitValues: GenerateDataset(1024, Seed)})
}

func BenchmarkBWArr_Insert_2048(b *testing.B) {
	BenchBWArrInsert(b, Params{N: 2048, InitValues: GenerateDataset(2048, Seed)})
}

func BenchmarkBWArr_Insert_4096(b *testing.B) {
	BenchBWArrInsert(b, Params{N: 4096, InitValues: GenerateDataset(4096, Seed)})
}

func BenchmarkBWArr_Insert_8192(b *testing.B) {
	BenchBWArrInsert(b, Params{N: 8192, InitValues: GenerateDataset(8192, Seed)})
}

// Benchmark functions for btree with different dataset sizes

func BenchmarkBTree_Insert_1024(b *testing.B) {
	BenchBtreeInsert(b, Params{N: 1024, InitValues: GenerateDataset(1024, Seed)})
}

func BenchmarkBTree_Insert_2048(b *testing.B) {
	BenchBtreeInsert(b, Params{N: 2048, InitValues: GenerateDataset(2048, Seed)})
}

func BenchmarkBTree_Insert_4096(b *testing.B) {
	BenchBtreeInsert(b, Params{N: 4096, InitValues: GenerateDataset(4096, Seed)})
}

func BenchmarkBTree_Insert_8192(b *testing.B) {
	BenchBtreeInsert(b, Params{N: 8192, InitValues: GenerateDataset(8192, Seed)})
}
