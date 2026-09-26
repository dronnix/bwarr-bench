package benchmark

import (
	"math"
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

// newBWArrFromSlice builds a BWArr pre-populated with the given values.
// It replaces bwarr.NewFromSlice, which was removed in bwarr v1.2.0.
func newBWArrFromSlice(values []int64) *bwarr.BWArr[int64] {
	bwa := bwarr.New(func(a, b int64) int {
		return int(a - b)
	}, len(values))
	for _, v := range values {
		bwa.Insert(v)
	}
	return bwa
}

// Every benchmark calls runtime.GC() while the timer is stopped, right before timing
// (re)starts, so that:
//   - garbage left by the previous iteration (its data structure) or by the framework's
//     earlier probe runs is not collected inside the timed section;
//   - no GC cycle started by the setup allocations is still marking in the background
//     when the timed section begins.
//
// GC cycles triggered by allocations inside the timed section are left in place: they
// are a real cost of the operation under test.

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

// Result holds the performance metrics aggregated over all repetitions of a benchmark run.
type Result struct {
	ExecTimePerOp   time.Duration             // Mean time per operation over all repetitions
	ExecTimeMin     time.Duration             // Fastest repetition
	ExecTimeMax     time.Duration             // Slowest repetition
	ExecTimeStdDev  time.Duration             // Sample standard deviation of time per operation
	AllocsPerOp     uint64                    // Mean number of allocations per operation
	AllocBytesPerOp uint64                    // Mean bytes allocated per operation
	Samples         []testing.BenchmarkResult // Raw result of every repetition, in benchstat-compatible form
}

// Func is a benchmark function signature that accepts testing.B and benchmark parameters.
type Func func(b *testing.B, params Params)

// Execute runs every Run of the comparison `count` times and aggregates the results.
// Within each repetition the bwarr and btree benchmarks are interleaved (bwarr, btree,
// bwarr, btree, ...) so that slow drift of the machine (thermal state, background
// load) affects both implementations equally instead of only the one that runs last.
// A count below 1 is treated as 1.
func (c *Comparison) Execute(count int) {
	if count < 1 {
		count = 1
	}
	for i := range c.Runs {
		run := &c.Runs[i]
		bwarrSamples := make([]testing.BenchmarkResult, 0, count)
		btreeSamples := make([]testing.BenchmarkResult, 0, count)

		for range count {
			bwarrSamples = append(bwarrSamples, testing.Benchmark(func(b *testing.B) { //nolint:thelper // This is a benchmark runner, not a helper
				c.BWArrBenchFunc(b, run.Params)
			}))
			btreeSamples = append(btreeSamples, testing.Benchmark(func(b *testing.B) { //nolint:thelper // This is a benchmark runner, not a helper
				c.BTreeBenchFunc(b, run.Params)
			}))
		}

		run.BwarrResult = Aggregate(bwarrSamples)
		run.BTreeResult = Aggregate(btreeSamples)
	}
}

// Aggregate computes mean, min, max and sample standard deviation of ns/op over
// the given repetitions. Allocation metrics are averaged. Samples are kept so
// callers can write raw results for tools like benchstat.
func Aggregate(samples []testing.BenchmarkResult) Result {
	n := len(samples)
	if n == 0 {
		return Result{}
	}

	var (
		sumNs, sumAllocs, sumBytes float64
		minNs, maxNs               = math.Inf(1), math.Inf(-1)
	)
	for _, s := range samples {
		ns := float64(s.NsPerOp())
		sumNs += ns
		minNs = math.Min(minNs, ns)
		maxNs = math.Max(maxNs, ns)
		sumAllocs += float64(s.AllocsPerOp())
		sumBytes += float64(s.AllocedBytesPerOp())
	}
	mean := sumNs / float64(n)

	var stddev float64
	if n > 1 {
		var sq float64
		for _, s := range samples {
			d := float64(s.NsPerOp()) - mean
			sq += d * d
		}
		stddev = math.Sqrt(sq / float64(n-1))
	}

	return Result{
		ExecTimePerOp:   time.Duration(mean),
		ExecTimeMin:     time.Duration(minNs),
		ExecTimeMax:     time.Duration(maxNs),
		ExecTimeStdDev:  time.Duration(stddev),
		AllocsPerOp:     uint64(math.Round(sumAllocs / float64(n))), //nolint:gosec // AllocsPerOp is never negative
		AllocBytesPerOp: uint64(math.Round(sumBytes / float64(n))),  //nolint:gosec // AllocedBytesPerOp is never negative
		Samples:         samples,
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
		runtime.GC()
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
		runtime.GC()
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

	bwa := newBWArrFromSlice(params.InitValues)

	toFind := params.InitValues[:params.ElementsToApply] // TODO: use a better selection strategy (shuffle?)

	// Collect setup garbage and finish any running GC cycle before the timer starts
	runtime.GC()
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

	// Collect setup garbage and finish any running GC cycle before the timer starts
	runtime.GC()
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

	bwa := newBWArrFromSlice(params.InitValues)

	// Collect setup garbage and finish any running GC cycle before the timer starts
	runtime.GC()
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

	// Collect setup garbage and finish any running GC cycle before the timer starts
	runtime.GC()
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

	bwa := newBWArrFromSlice(params.InitValues)

	// Collect setup garbage and finish any running GC cycle before the timer starts
	runtime.GC()
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

	toDel := params.InitValues[:params.ElementsToApply] // TODO: use a better selection strategy (shuffle?)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		// Stop timer during tree population (setup, not measured).
		// A fresh tree per iteration so every iteration deletes real elements.
		b.StopTimer()
		tree := btree.NewOrderedG[int64](BTreeDegree)
		for _, v := range params.InitValues {
			tree.ReplaceOrInsert(v)
		}
		runtime.GC()
		b.StartTimer()

		// Measured operation: delete all values
		for _, v := range toDel {
			tree.Delete(v)
		}
	}
}

// BenchBWArrDelete benchmarks deleting all values from a pre-populated BWArr.
func BenchBWArrDelete(b *testing.B, params Params) {
	b.Helper()

	toDel := params.InitValues[:params.ElementsToApply] // TODO: use a better selection strategy (shuffle?)

	// Reset timer to exclude any setup time
	b.ResetTimer()

	// Run b.N iterations (controlled by testing.B framework)
	for range b.N {
		// Stop timer during bwa population (setup, not measured).
		// A fresh bwa per iteration so every iteration deletes real elements.
		b.StopTimer()
		bwa := newBWArrFromSlice(params.InitValues)
		runtime.GC()
		b.StartTimer()

		// Measured operation: delete all values
		for _, v := range toDel {
			bwa.Delete(v)
		}
	}
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
