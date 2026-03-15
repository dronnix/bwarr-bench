package benchmark

import (
	"math/rand"
	"runtime"
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
)

// Comparison consists of multiple benchmark runs comparing two implementations.
type Comparison struct {
	Name           string
	MeasureAllocs  bool // Whether to measure allocations
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

// Params contains parameters for running a single benchmark.
type Params struct {
	ElementsToApply int     // Number of elements to apply inside the benchmark
	InitValues      []int64 // Pre-generated values to populate the data structures
}

// Result holds the performance metrics from a benchmark run.
type Result struct {
	ExecTimePerOp   time.Duration // Time per operation
	AllocsPerOp     uint64        // Number of allocations per operation
	AllocBytesPerOp uint64        // Bytes allocated per operation
}

// Func is a benchmark function signature that accepts testing.B and benchmark parameters.
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
			ExecTimePerOp:   time.Duration(bwarrResult.NsPerOp()),
			AllocsPerOp:     uint64(bwarrResult.AllocsPerOp()),       //nolint:gosec // AllocsPerOp always returns non-negative value
			AllocBytesPerOp: uint64(bwarrResult.AllocedBytesPerOp()), //nolint:gosec // AllocedBytesPerOp always returns non-negative value
		}

		// Run btree benchmark
		btreeResult := testing.Benchmark(func(b *testing.B) { //nolint:thelper // This is a benchmark runner, not a helper
			c.BTreeBenchFunc(b, run.Params)
		})

		// Populate btree result
		run.BTreeResult = Result{
			ExecTimePerOp:   time.Duration(btreeResult.NsPerOp()),
			AllocsPerOp:     uint64(btreeResult.AllocsPerOp()),       //nolint:gosec // AllocsPerOp always returns non-negative value
			AllocBytesPerOp: uint64(btreeResult.AllocedBytesPerOp()), //nolint:gosec // AllocedBytesPerOp always returns non-negative value
		}
	}
}

// BenchBWArrInsert benchmarks bwarr insert operations with a pre-generated dataset.
func BenchBWArrInsert(b *testing.B, params Params) {
	b.Helper()
	values := params.InitValues

	// Enable memory allocation reporting
	b.ReportAllocs()

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		// Stop timer during bwa creation (setup, not measured)
		b.StopTimer()
		bwa := bwarr.New(func(a, b int64) int {
			return int(a - b)
		}, 0)
		b.StartTimer()

		// Measured operation: Insert all values into fresh tree
		for _, v := range values {
			bwa.Insert(v)
		}
	}
}

// BenchBTreeInsert benchmarks btree insert operations with a pre-generated dataset.
func BenchBTreeInsert(b *testing.B, params Params) {
	b.Helper()
	values := params.InitValues

	// Enable memory allocation reporting
	b.ReportAllocs()

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

// BenchBWArrGet benchmarks BWArr Get operations on a pre-populated data structure.
func BenchBWArrGet(b *testing.B, params Params) {
	b.Helper()

	bwa := bwarr.NewFromSlice(func(a, b int64) int {
		return int(a - b)
	}, params.InitValues)

	toFind := params.InitValues[:params.ElementsToApply] // TODO: use a better selection strategy (shuffle?)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		for _, v := range toFind {
			r, ok := bwa.Get(v)
			if !ok || r != v { // Use return values to avoid compiler optimizations
				b.Fatalf("Expected to find %d, got %d (found: %v)", v, r, ok)
			}
		}
	}
}

// BenchBTreeGet benchmarks BTree Get operations on a pre-populated data structure.
func BenchBTreeGet(b *testing.B, params Params) {
	b.Helper()

	tree := btree.NewOrderedG[int64](BTreeDegree)
	for _, v := range params.InitValues {
		tree.ReplaceOrInsert(v)
	}

	toFind := params.InitValues[:params.ElementsToApply] // TODO: use a better selection strategy (shuffle?)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		for _, v := range toFind {
			r, ok := tree.Get(v)
			if !ok || r != v { // Use return values to avoid compiler optimizations
				b.Fatalf("Expected to find %d, got %d (found: %v)", v, r, ok)
			}
		}
	}
}

// BenchBWArrOrderedIterate benchmarks iterating through all values in sorted order using BWArr.
func BenchBWArrOrderedIterate(b *testing.B, params Params) {
	b.Helper()

	bwa := bwarr.NewFromSlice(func(a, b int64) int {
		return int(a - b)
	}, params.InitValues)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	s := int64(0)
	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		bwa.Ascend(func(item int64) bool {
			s += item // Use item to avoid compiler optimizations
			return true
		})
	}
}

// BenchBTreeOrderedIterate benchmarks iterating through all values in sorted order using BTree.
func BenchBTreeOrderedIterate(b *testing.B, params Params) {
	b.Helper()

	tree := btree.NewOrderedG[int64](BTreeDegree)
	for _, v := range params.InitValues {
		tree.ReplaceOrInsert(v)
	}

	// Reset timer to exclude any setup time
	b.ResetTimer()

	s := int64(0)
	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		tree.Ascend(func(item int64) bool {
			s += item // Use item to avoid compiler optimizations
			return true
		})
	}
}

// BenchBWArrUnorderedIterate benchmarks iterating through all values without ordering using BWArr.
func BenchBWArrUnorderedIterate(b *testing.B, params Params) {
	b.Helper()

	bwa := bwarr.NewFromSlice(func(a, b int64) int {
		return int(a - b)
	}, params.InitValues)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	s := int64(0)
	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		bwa.UnorderedWalk(func(item int64) bool {
			s += item // Use item to avoid compiler optimizations
			return true
		})
	}
}

// BenchBTreeDelete benchmarks deleting all values from a pre-populated BTree.
func BenchBTreeDelete(b *testing.B, params Params) {
	b.Helper()

	tree := btree.NewOrderedG[int64](BTreeDegree)
	for _, v := range params.InitValues {
		tree.ReplaceOrInsert(v)
	}

	toDel := params.InitValues[:params.ElementsToApply] // TODO: use a better selection strategy (shuffle?)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		for _, v := range toDel {
			tree.Delete(v)
		}
	}
}

// BenchBWArrDelete benchmarks deleting all values from a pre-populated BWArr.
func BenchBWArrDelete(b *testing.B, params Params) {
	b.Helper()

	bwa := bwarr.NewFromSlice(func(a, b int64) int {
		return int(a - b)
	}, params.InitValues)

	toDel := params.InitValues[:params.ElementsToApply] // TODO: use a better selection strategy (shuffle?)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		for _, v := range toDel {
			bwa.Delete(v)
		}
	}
}

// MemoryMeasureFunc measures heap bytes for a given dataset.
type MemoryMeasureFunc func(values []int64) uint64

// MemoryComparison measures heap memory footprint of both data structures at various sizes.
type MemoryComparison struct {
	Name         string
	BWArrMemFunc MemoryMeasureFunc
	BTreeMemFunc MemoryMeasureFunc
	Runs         []MemoryRun
}

// MemoryRun holds parameters and results for a single memory footprint measurement.
type MemoryRun struct {
	DatasetSize int
	InitValues  []int64

	BWArrHeapBytes uint64
	BTreeHeapBytes uint64
}

// Execute runs the memory measurement functions for each dataset size.
func (mc *MemoryComparison) Execute() {
	for i := range mc.Runs {
		run := &mc.Runs[i]
		run.BWArrHeapBytes = mc.BWArrMemFunc(run.InitValues)
		run.BTreeHeapBytes = mc.BTreeMemFunc(run.InitValues)
	}
}

// DeleteFraction is the fraction of elements to delete in delete-memory benchmarks.
const DeleteFraction = 0.75

// MeasureBWArrInsertHeap returns the live heap bytes used by a BWArr populated with values.
func MeasureBWArrInsertHeap(values []int64) uint64 {
	runtime.GC()
	time.Sleep(100 * time.Millisecond) //nolint:mnd // Give GC time to stabilize
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	bwa := bwarr.NewWithOptions(func(a, b int64) int {
		return int(a - b)
	}, 0, bwarr.Options{
		ElementsKeepAllocated: 0,
		DeleteUnusedSegments:  true,
	})
	for _, v := range values {
		bwa.Insert(v)
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond) //nolint:mnd // Give GC time to stabilize
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	runtime.KeepAlive(bwa)

	if after.HeapAlloc < before.HeapAlloc {
		return 0
	}
	return after.HeapAlloc - before.HeapAlloc
}

// MeasureBTreeInsertHeap returns the live heap bytes used by a BTree populated with values.
func MeasureBTreeInsertHeap(values []int64) uint64 {
	runtime.GC()
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	tree := btree.NewOrderedG[int64](BTreeDegree)
	for _, v := range values {
		tree.ReplaceOrInsert(v)
	}

	runtime.GC()
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	runtime.KeepAlive(tree)

	if after.HeapAlloc < before.HeapAlloc {
		return 0
	}
	return after.HeapAlloc - before.HeapAlloc
}

// MeasureBWArrDeleteFreed returns the heap bytes freed after deleting 75% of elements from a BWArr.
func MeasureBWArrDeleteFreed(values []int64) uint64 {
	bwa := bwarr.NewWithOptions(func(a, b int64) int {
		return int(a - b)
	}, 0, bwarr.Options{
		ElementsKeepAllocated: 0,
		DeleteUnusedSegments:  true,
	})
	for _, v := range values {
		bwa.Insert(v)
	}

	toDel := values[:int(float64(len(values))*DeleteFraction)]

	runtime.GC()
	time.Sleep(100 * time.Millisecond) //nolint:mnd // Give GC time to stabilize
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for _, v := range toDel {
		bwa.Delete(v)
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond) //nolint:mnd // Give GC time to stabilize
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	runtime.KeepAlive(bwa)

	if before.HeapAlloc < after.HeapAlloc {
		return 0
	}
	return before.HeapAlloc - after.HeapAlloc
}

// MeasureBTreeDeleteFreed returns the heap bytes freed after deleting 75% of elements from a BTree.
func MeasureBTreeDeleteFreed(values []int64) uint64 {
	tree := btree.NewOrderedG[int64](BTreeDegree)
	for _, v := range values {
		tree.ReplaceOrInsert(v)
	}

	toDel := values[:int(float64(len(values))*DeleteFraction)]

	runtime.GC()
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	for _, v := range toDel {
		tree.Delete(v)
	}

	runtime.GC()
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	runtime.KeepAlive(tree)

	if before.HeapAlloc < after.HeapAlloc {
		return 0
	}
	return before.HeapAlloc - after.HeapAlloc
}

// GenerateRandomDataset creates a reproducible slice of random int64 values.
// Values are in range [0, maxValue).
func GenerateRandomDataset(count int, seed, maxValue int64) []int64 {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // Using math/rand for reproducible benchmark data
	values := make([]int64, count)
	for i := range count {
		values[i] = rng.Int63n(maxValue)
	}
	return values
}
