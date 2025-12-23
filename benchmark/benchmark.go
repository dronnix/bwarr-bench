package benchmark

import (
	"math/rand"
	"testing"
	"time"

	"github.com/dronnix/bwarr"
	"github.com/google/btree"
)

const (
	// Seed is the fixed seed used for generating reproducible datasets.
	Seed = 42

	// BTreeDegree is the degree parameter for btree (per user requirements).
	BTreeDegree = 32

	// BWArrCapacity is the initial capacity hint for bwarr (0 = auto).
	BWArrCapacity = 0
)

// GenerateDataset creates a reproducible slice of random int64 values.
func GenerateDataset(count int, seed int64) []int64 {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // Using math/rand for reproducible benchmark data
	values := make([]int64, count)
	for i := range count {
		values[i] = rng.Int63()
	}
	return values
}

// Comparison consists of multiple benchmark runs comparing two implementations.
type Comparison struct {
	Name           string
	BWArrBenchFunc Func
	BTreeBenchFunc Func
	Runs           []Run
}

// Run represents a single benchmark run with specific parameters and results for both datastructures.
type Run struct {
	Params

	BwarrResult Result
	BTreeResult Result
}

type Params struct {
	N          int     // Number of elements to apply inside the benchmark
	InitValues []int64 // Pre-generated values to populate the data structures
}

type Result struct {
	ExecTimePerOp time.Duration
	AllocsPerOp   uint64
}

type Func func(b *testing.B, params Params)

// Execute runs all benchmark runs and populates the result fields for each.
func (c *Comparison) Execute() {
	for i := range c.Runs {
		run := &c.Runs[i]

		// Run bwarr benchmark
		bwarrResult := testing.Benchmark(func(b *testing.B) { //nolint:thelper // This is a benchmark runner, not a helper
			c.BWArrBenchFunc(b, run.Params)
		})

		// Populate bwarr result
		run.BwarrResult = Result{
			ExecTimePerOp: time.Duration(bwarrResult.NsPerOp()),
			AllocsPerOp:   uint64(bwarrResult.AllocsPerOp()), //nolint:gosec // AllocsPerOp always returns non-negative value
		}

		// Run btree benchmark
		btreeResult := testing.Benchmark(func(b *testing.B) { //nolint:thelper // This is a benchmark runner, not a helper
			c.BTreeBenchFunc(b, run.Params)
		})

		// Populate btree result
		run.BTreeResult = Result{
			ExecTimePerOp: time.Duration(btreeResult.NsPerOp()),
			AllocsPerOp:   uint64(btreeResult.AllocsPerOp()), //nolint:gosec // AllocsPerOp always returns non-negative value
		}
	}
}

// BenchBWArrInsert benchmarks bwarr insert operations with a pre-generated dataset.
func BenchBWArrInsert(b *testing.B, params Params) {
	b.Helper()
	values := params.InitValues

	// Enable memory allocation reporting
	b.ReportAllocs()

	// Report the number of bytes processed per operation (for throughput calculation)
	b.SetBytes(int64(len(values) * 8)) //nolint:mnd // 8 is the size of int64 in bytes

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		// Stop timer during tree creation (setup, not measured)
		b.StopTimer()
		tree := bwarr.New(func(a, b int64) int {
			return int(a - b)
		}, BWArrCapacity)
		b.StartTimer()

		// Measured operation: Insert all values into fresh tree
		for _, v := range values {
			tree.Insert(v)
		}
	}
}

// BenchBtreeInsert benchmarks btree insert operations with a pre-generated dataset.
func BenchBtreeInsert(b *testing.B, params Params) {
	b.Helper()
	values := params.InitValues

	// Enable memory allocation reporting
	b.ReportAllocs()

	// Report the number of bytes processed per operation (for throughput calculation)
	b.SetBytes(int64(len(values) * 8)) //nolint:mnd // 8 is the size of int64 in bytes

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
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
