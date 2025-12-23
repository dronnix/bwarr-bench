package benchmark

import (
	"testing"

	"github.com/dronnix/bwarr"
	"github.com/google/btree"
)

func TestGenerateRandomInt64s(t *testing.T) {
	t.Run("same seed produces same values", func(t *testing.T) {
		values1 := generateRandomInt64s(100, 42)
		values2 := generateRandomInt64s(100, 42)

		if len(values1) != len(values2) {
			t.Errorf("expected same length, got %d and %d", len(values1), len(values2))
		}

		for i := range values1 {
			if values1[i] != values2[i] {
				t.Errorf("values differ at index %d: %d != %d", i, values1[i], values2[i])
			}
		}
	})

	t.Run("different seeds produce different values", func(t *testing.T) {
		values1 := generateRandomInt64s(100, 42)
		values2 := generateRandomInt64s(100, 99)

		allSame := true
		for i := range values1 {
			if values1[i] != values2[i] {
				allSame = false
				break
			}
		}

		if allSame {
			t.Error("expected different values with different seeds, but all values were the same")
		}
	})

	t.Run("generates correct count", func(t *testing.T) {
		counts := []int{10, 100, 1000}
		for _, count := range counts {
			values := generateRandomInt64s(count, 42)
			if len(values) != count {
				t.Errorf("expected %d values, got %d", count, len(values))
			}
		}
	})
}

func TestDatasetInitialization(t *testing.T) {
	// Verify all datasets are initialized and non-empty
	datasets := []struct {
		name string
		data []int64
		size int
	}{
		{"Dataset1024", Dataset1024, 1024},
		{"Dataset2048", Dataset2048, 2048},
		{"Dataset4096", Dataset4096, 4096},
		{"Dataset8192", Dataset8192, 8192},
	}

	for _, ds := range datasets {
		if len(ds.data) != ds.size {
			t.Errorf("%s: expected length %d, got %d", ds.name, ds.size, len(ds.data))
		}

		// Verify dataset contains non-zero values
		hasNonZero := false
		for _, v := range ds.data {
			if v != 0 {
				hasNonZero = true
				break
			}
		}
		if !hasNonZero {
			t.Errorf("%s: dataset contains only zero values", ds.name)
		}
	}
}

// Benchmark functions for bwarr with different dataset sizes

func BenchmarkBWArr_Insert_1024(b *testing.B) {
	benchmarkBWarrWithDataset(b, Dataset1024)
}

func BenchmarkBWArr_Insert_2048(b *testing.B) {
	benchmarkBWarrWithDataset(b, Dataset2048)
}

func BenchmarkBWArr_Insert_4096(b *testing.B) {
	benchmarkBWarrWithDataset(b, Dataset4096)
}

func BenchmarkBWArr_Insert_8192(b *testing.B) {
	benchmarkBWarrWithDataset(b, Dataset8192)
}

// Benchmark functions for btree with different dataset sizes

func BenchmarkBTree_Insert_1024(b *testing.B) {
	benchmarkBTreeWithDataset(b, Dataset1024)
}

func BenchmarkBTree_Insert_2048(b *testing.B) {
	benchmarkBTreeWithDataset(b, Dataset2048)
}

func BenchmarkBTree_Insert_4096(b *testing.B) {
	benchmarkBTreeWithDataset(b, Dataset4096)
}

func BenchmarkBTree_Insert_8192(b *testing.B) {
	benchmarkBTreeWithDataset(b, Dataset8192)
}

// benchmarkBWarrWithDataset benchmarks bwarr insert operations with a pre-generated dataset.
func benchmarkBWarrWithDataset(b *testing.B, values []int64) {
	// Enable memory allocation reporting
	b.ReportAllocs()

	// Report the number of bytes processed per operation (for throughput calculation)
	b.SetBytes(int64(len(values) * 8)) // 8 bytes per int64

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for i := 0; i < b.N; i++ {
		// Stop timer during tree creation (setup, not measured)
		b.StopTimer()
		tree := bwarr.New(func(a, b int64) int {
			return int(a - b)
		}, BWarrCapacity)
		b.StartTimer()

		// Measured operation: Insert all values into fresh tree
		for _, v := range values {
			tree.Insert(v)
		}
	}
}

// benchmarkBTreeWithDataset benchmarks btree insert operations with a pre-generated dataset.
func benchmarkBTreeWithDataset(b *testing.B, values []int64) {
	// Enable memory allocation reporting
	b.ReportAllocs()

	// Report the number of bytes processed per operation (for throughput calculation)
	b.SetBytes(int64(len(values) * 8)) // 8 bytes per int64

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for i := 0; i < b.N; i++ {
		// Stop timer during tree creation (setup, not measured)
		b.StopTimer()
		tree := btree.NewOrderedG[int64](BTreeDegree)
		b.StartTimer()

		// Measured operation: Insert all values into fresh tree
		for _, v := range values {
			tree.ReplaceOrInsert(v)
		}
	}
}
