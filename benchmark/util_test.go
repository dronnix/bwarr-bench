package benchmark

import (
	"math"
	"testing"
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
	// Create a simple comparison with one run
	comparison := Comparison{
		Name:           "Test",
		BWArrBenchFunc: BenchBWArrInsert,
		BTreeBenchFunc: BenchBTreeInsert,
		Runs: []Run{
			{
				Params: Params{
					ElementsToApply: 100,
					InitValues:      GenerateRandomDataset(100, Seed, math.MaxInt64),
				},
			},
		},
		MeasureAllocs: true,
	}

	// Execute
	comparison.Execute()

	// Verify results are populated
	run := comparison.Runs[0]

	if run.BwarrResult.ExecTimePerOp == 0 {
		t.Error("BwarrResult.ExecTimePerOp is zero")
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
