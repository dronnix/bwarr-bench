package benchmark

import (
	"math"
	"testing"
	"time"
)

func TestGenerateRandomDataset_Reproducibility(t *testing.T) {
	const count = 1000
	const seed = 42
	const maxValue = 10000

	// Generate dataset twice with same seed
	dataset1 := GenerateRandomDataset(count, seed, maxValue)
	dataset2 := GenerateRandomDataset(count, seed, maxValue)

	// Should be identical
	if len(dataset1) != len(dataset2) {
		t.Fatalf("Datasets have different lengths: %d vs %d", len(dataset1), len(dataset2))
	}

	for i := range dataset1 {
		if dataset1[i] != dataset2[i] {
			t.Errorf("Datasets differ at index %d: %d vs %d", i, dataset1[i], dataset2[i])
		}
	}
}

func TestGenerateRandomDataset_Count(t *testing.T) {
	tests := []struct {
		count int
	}{
		{0},
		{1},
		{100},
		{10000},
	}

	for _, tt := range tests {
		dataset := GenerateRandomDataset(tt.count, Seed, math.MaxInt64)
		if len(dataset) != tt.count {
			t.Errorf("GenerateRandomDataset(%d) returned %d elements", tt.count, len(dataset))
		}
	}
}

func TestGenerateRandomDataset_Range(t *testing.T) {
	const count = 1000
	const maxValue = 5000

	dataset := GenerateRandomDataset(count, Seed, maxValue)

	for i, v := range dataset {
		if v < 0 || v >= maxValue {
			t.Errorf("Value at index %d is out of range [0, %d): got %d", i, maxValue, v)
		}
	}
}

func TestGenerateRandomDataset_DifferentSeeds(t *testing.T) {
	const count = 1000
	const maxValue = 10000

	dataset1 := GenerateRandomDataset(count, 42, maxValue)
	dataset2 := GenerateRandomDataset(count, 43, maxValue)

	// Should be different (with high probability)
	identical := true
	for i := range dataset1 {
		if dataset1[i] != dataset2[i] {
			identical = false
			break
		}
	}

	if identical {
		t.Error("Datasets with different seeds are identical (highly unlikely)")
	}
}

func TestComparisonExecute(t *testing.T) {
	// Create a simple comparison with one run. 10K elements keeps each testing.Benchmark
	// run close to its 1s budget: with much smaller datasets the per-iteration
	// StopTimer/StartTimer calls (each a stop-the-world ReadMemStats) dominate wall time.
	const elements = 10_000
	comparison := Comparison{
		Name:           "Test",
		BWArrBenchFunc: BenchBWArrInsert,
		BTreeBenchFunc: BenchBTreeInsert,
		Runs: []Run{
			{
				Params: Params{
					ElementsToApply: elements,
					InitValues:      GenerateRandomDataset(elements, Seed, math.MaxInt64),
				},
			},
		},
		MeasureAllocs: true,
	}

	// Execute with 2 repetitions so the spread statistics are exercised
	comparison.Execute(2)

	// Verify results are populated
	run := comparison.Runs[0]

	if run.BwarrResult.ExecTimePerOp == 0 {
		t.Error("BwarrResult.ExecTimePerOp is zero")
	}

	if len(run.BwarrResult.Samples) != 2 || len(run.BTreeResult.Samples) != 2 {
		t.Errorf("expected 2 samples per implementation, got %d and %d",
			len(run.BwarrResult.Samples), len(run.BTreeResult.Samples))
	}

	if run.BwarrResult.ExecTimeMin > run.BwarrResult.ExecTimePerOp || run.BwarrResult.ExecTimePerOp > run.BwarrResult.ExecTimeMax {
		t.Errorf("expected min <= mean <= max, got %v <= %v <= %v",
			run.BwarrResult.ExecTimeMin, run.BwarrResult.ExecTimePerOp, run.BwarrResult.ExecTimeMax)
	}

	if run.BTreeResult.ExecTimePerOp == 0 {
		t.Error("BTreeResult.ExecTimePerOp is zero")
	}

	// Since we're measuring allocations, these should be non-zero
	if run.BwarrResult.AllocsPerOp == 0 {
		t.Error("BwarrResult.AllocsPerOp is zero")
	}

	if run.BTreeResult.AllocsPerOp == 0 {
		t.Error("BTreeResult.AllocsPerOp is zero")
	}

	if run.BwarrResult.AllocBytesPerOp == 0 {
		t.Error("BwarrResult.AllocBytesPerOp is zero")
	}

	if run.BTreeResult.AllocBytesPerOp == 0 {
		t.Error("BTreeResult.AllocBytesPerOp is zero")
	}
}

func TestAggregate(t *testing.T) {
	sample := func(ns int64, allocs, bytes uint64) testing.BenchmarkResult {
		return testing.BenchmarkResult{N: 1, T: time.Duration(ns), MemAllocs: allocs, MemBytes: bytes}
	}

	t.Run("empty", func(t *testing.T) {
		got := Aggregate(nil)
		if got.ExecTimePerOp != 0 || got.ExecTimeMin != 0 || got.ExecTimeMax != 0 || got.ExecTimeStdDev != 0 ||
			got.AllocsPerOp != 0 || got.AllocBytesPerOp != 0 || len(got.Samples) != 0 {
			t.Errorf("Aggregate(nil) = %+v, want zero Result", got)
		}
	})

	t.Run("single sample has zero spread", func(t *testing.T) {
		got := Aggregate([]testing.BenchmarkResult{sample(100, 3, 64)})
		if got.ExecTimePerOp != 100 || got.ExecTimeMin != 100 || got.ExecTimeMax != 100 || got.ExecTimeStdDev != 0 {
			t.Errorf("unexpected stats: %+v", got)
		}
		if got.AllocsPerOp != 3 || got.AllocBytesPerOp != 64 {
			t.Errorf("unexpected alloc stats: %+v", got)
		}
	})

	t.Run("mean min max stddev", func(t *testing.T) {
		// ns/op: 100, 200, 300 -> mean 200, sample stddev 100
		got := Aggregate([]testing.BenchmarkResult{
			sample(100, 1, 10),
			sample(200, 2, 20),
			sample(300, 3, 30),
		})
		if got.ExecTimePerOp != 200 {
			t.Errorf("mean = %v, want 200ns", got.ExecTimePerOp)
		}
		if got.ExecTimeMin != 100 || got.ExecTimeMax != 300 {
			t.Errorf("min/max = %v/%v, want 100ns/300ns", got.ExecTimeMin, got.ExecTimeMax)
		}
		if got.ExecTimeStdDev != 100 {
			t.Errorf("stddev = %v, want 100ns", got.ExecTimeStdDev)
		}
		if got.AllocsPerOp != 2 || got.AllocBytesPerOp != 20 {
			t.Errorf("alloc means = %d/%d, want 2/20", got.AllocsPerOp, got.AllocBytesPerOp)
		}
		if len(got.Samples) != 3 {
			t.Errorf("len(Samples) = %d, want 3", len(got.Samples))
		}
	})
}
