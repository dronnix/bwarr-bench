package benchmark

import (
	"cmp"
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
//
// cmp.Compare is used as the comparator instead of `int(a - b)` because
// subtraction overflows for values of opposite sign and truncates on 32-bit targets.
func newBWArrFromSlice(values []int64) *bwarr.BWArr[int64] {
	bwa := bwarr.New(cmp.Compare[int64], len(values))
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

// Series is one line on a graph: a named implementation of the operation under test.
type Series struct {
	Name string // Legend label, e.g. "bwarr" or "btree (key+value)"
	Func Func   // Benchmark function producing the measurements for this series
}

// Comparison consists of multiple benchmark runs comparing several implementations
// (Series) of the same operation across dataset sizes (Runs).
type Comparison struct {
	Name          string   // Human-readable title used on graphs
	FileName      string   // Base name for output files; derived from Name when empty
	MeasureAllocs bool     // Whether to measure allocations
	Series        []Series // Implementations under comparison, in legend order
	// Dataset generates the InitValues for a run of n elements. nil means the default
	// shared dataset of unique random values.
	Dataset func(n int) []int64
	Runs    []Run
}

// Run represents a single benchmark run with specific parameters and one result per Series.
type Run struct {
	Params

	Results []Result // Index-aligned with Comparison.Series
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
// Within each repetition all series are run in order (s1, s2, ..., s1, s2, ...) so that
// slow drift of the machine (thermal state, background load) affects every
// implementation equally instead of only the one that runs last.
// A count below 1 is treated as 1.
func (c *Comparison) Execute(count int) {
	if count < 1 {
		count = 1
	}
	for i := range c.Runs {
		run := &c.Runs[i]
		samples := make([][]testing.BenchmarkResult, len(c.Series))
		for s := range samples {
			samples[s] = make([]testing.BenchmarkResult, 0, count)
		}

		for range count {
			for s, series := range c.Series {
				samples[s] = append(samples[s], testing.Benchmark(func(b *testing.B) { //nolint:thelper // This is a benchmark runner, not a helper
					series.Func(b, run.Params)
				}))
			}
		}

		run.Results = make([]Result, len(c.Series))
		for s := range c.Series {
			run.Results[s] = Aggregate(samples[s])
		}
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

// ---------------------------------------------------------------------------
// Case 1: Insert, duplicates allowed.
// bwarr has a true multiset Insert. btree has no such operation, so it is measured
// with ReplaceOrInsert, the closest call. On a dataset of unique values both do the
// same amount of useful work.
// ---------------------------------------------------------------------------

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
		bwa := bwarr.New(cmp.Compare[int64], 0)
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

// ---------------------------------------------------------------------------
// Case 2: ReplaceOrInsert, the collection stays unique (set semantics).
// Both libraries offer this operation, so this is the like-for-like comparison and
// the number a user migrating from btree to bwarr will observe.
// ---------------------------------------------------------------------------

// BenchBWArrReplaceOrInsert benchmarks bwarr ReplaceOrInsert (search + insert) on a
// pre-generated dataset of unique values, so every call ends in an insert.
func BenchBWArrReplaceOrInsert(b *testing.B, params Params) {
	b.Helper()
	values := params.InitValues

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		b.StopTimer()
		bwa := bwarr.New(cmp.Compare[int64], 0)
		runtime.GC()
		b.StartTimer()

		for _, v := range values {
			bwa.ReplaceOrInsert(v)
		}
	}
}

// ---------------------------------------------------------------------------
// Read / iterate / delete benchmarks (unique int64 values).
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Dataset generators.
// ---------------------------------------------------------------------------

// GenerateIncreasingDataset returns 0, 1, ..., count-1: keys arriving in sorted order,
// as timestamps or auto-increment IDs do. Every value lands after all existing ones.
func GenerateIncreasingDataset(count int) []int64 {
	values := make([]int64, count)
	for i := range count {
		values[i] = int64(i)
	}
	return values
}

// GenerateDecreasingDataset returns count-1, ..., 1, 0: the reverse of
// GenerateIncreasingDataset. Every value lands before all existing ones.
func GenerateDecreasingDataset(count int) []int64 {
	values := make([]int64, count)
	for i := range count {
		values[i] = int64(count - 1 - i)
	}
	return values
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

// ---------------------------------------------------------------------------
// Mixed workload: Insert, Get and Delete interleaved 1:1:1 in a seeded random order
// on a structure pre-populated with n keys. Each kind gets n/3 operations, so the
// structure stays at about n elements and the total is n operations, comparable
// with the single-operation graphs. The insert side follows case 1: bwarr blind
// Insert vs btree ReplaceOrInsert.
// ---------------------------------------------------------------------------

// mixedOpKinds is the number of operation kinds in the mixed workload.
const mixedOpKinds = 3

// GenerateMixedDataset returns n + n/3 random keys for the mixed workload. The first n
// are identical to the shared random dataset (same seed, same sequence) and
// pre-populate the structure; the last n/3 are the new keys inserted by the workload.
func GenerateMixedDataset(n int) []int64 {
	return GenerateRandomDataset(n+n/mixedOpKinds, Seed, math.MaxInt64)
}

type mixedOp uint8

const (
	mixedInsert mixedOp = iota
	mixedGet
	mixedDelete
)

// mixedWorkload holds the pre-computed operation order and key streams of one run.
// The dataset is already in random order, so the streams are plain index ranges:
// deletes take the first third of the base keys, gets the second third (never
// deleted, so every Get hits), inserts the extra keys after the base.
type mixedWorkload struct {
	base    []int64   // Keys the structure is pre-populated with
	inserts []int64   // New keys, one per insert
	gets    []int64   // Existing keys that are never deleted, one per get
	deletes []int64   // Existing keys, one per delete
	ops     []mixedOp // Operation order: exactly len(inserts) of each kind, seeded shuffle
}

// newMixedWorkload builds the workload for params. Both series are given the same
// params and the shuffle is seeded, so they run the identical sequence.
func newMixedWorkload(b *testing.B, params Params) mixedWorkload {
	b.Helper()
	n := params.ElementsToApply
	per := n / mixedOpKinds
	if len(params.InitValues) < n+per {
		b.Fatalf("mixed workload needs %d init values (n + n/%d), got %d", n+per, mixedOpKinds, len(params.InitValues))
	}
	d := params.InitValues

	ops := make([]mixedOp, 0, per*mixedOpKinds)
	for _, kind := range []mixedOp{mixedInsert, mixedGet, mixedDelete} {
		for range per {
			ops = append(ops, kind)
		}
	}
	rng := rand.New(rand.NewSource(Seed)) //nolint:gosec // Reproducible operation order, not security
	rng.Shuffle(len(ops), func(i, j int) { ops[i], ops[j] = ops[j], ops[i] })

	return mixedWorkload{
		base:    d[:n],
		deletes: d[:per],
		gets:    d[per : 2*per],
		inserts: d[n : n+per],
		ops:     ops,
	}
}

// BenchBWArrMixed benchmarks the mixed workload on bwarr: Insert, Get, Delete.
func BenchBWArrMixed(b *testing.B, params Params) {
	b.Helper()
	w := newMixedWorkload(b, params)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		// Fresh pre-populated structure per iteration, outside the timer.
		b.StopTimer()
		bwa := newBWArrFromSlice(w.base)
		runtime.GC()
		b.StartTimer()

		var ii, gi, di int
		for _, op := range w.ops {
			switch op {
			case mixedInsert:
				bwa.Insert(w.inserts[ii])
				ii++
			case mixedGet:
				k := w.gets[gi]
				r, ok := bwa.Get(k)
				if !ok || r != k { // Use return values to avoid compiler optimizations
					b.Fatalf("Expected to find %d, got %d (found: %v)", k, r, ok)
				}
				gi++
			case mixedDelete:
				bwa.Delete(w.deletes[di])
				di++
			}
		}
	}
}

// BenchBTreeMixed benchmarks the mixed workload on btree: ReplaceOrInsert, Get, Delete.
func BenchBTreeMixed(b *testing.B, params Params) {
	b.Helper()
	w := newMixedWorkload(b, params)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		// Fresh pre-populated structure per iteration, outside the timer.
		b.StopTimer()
		tree := btree.NewOrderedG[int64](BTreeDegree)
		for _, v := range w.base {
			tree.ReplaceOrInsert(v)
		}
		runtime.GC()
		b.StartTimer()

		var ii, gi, di int
		for _, op := range w.ops {
			switch op {
			case mixedInsert:
				tree.ReplaceOrInsert(w.inserts[ii])
				ii++
			case mixedGet:
				k := w.gets[gi]
				r, ok := tree.Get(k)
				if !ok || r != k { // Use return values to avoid compiler optimizations
					b.Fatalf("Expected to find %d, got %d (found: %v)", k, r, ok)
				}
				gi++
			case mixedDelete:
				tree.Delete(w.deletes[di])
				di++
			}
		}
	}
}
