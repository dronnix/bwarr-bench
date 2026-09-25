package main

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dronnix/bwarr-bench/benchmark"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Insert Performance", "insert_performance"},
		{"Get all values by key", "get_all_values_by_key"},
		{"Delete All Values", "delete_all_values"},
		{"Simple", "simple"},
		{"WITH-DASHES", "withdashes"},
		{"With Spaces", "with_spaces"},
		{"Multiple   Spaces", "multiple___spaces"},
		{"CamelCase", "camelcase"},
		{"with123numbers", "with123numbers"},
		{"123start", "123start"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"only spaces", "   ", "___"},
		{"only special chars", "!@#$%^&*()", ""},
		{"underscores preserved", "test_file_name", "test_file_name"},
		{"mixed case with specials", "Test-File!Name", "testfilename"},
		{"unicode characters", "Test™File®", "testfile"},
		{"numbers only", "12345", "12345"},
		{"underscores only", "___", "___"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDurationToMillis(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected float64
	}{
		{0, 0},
		{time.Millisecond, 1},
		{812039 * time.Nanosecond, 0.812039}, // sub-millisecond must not become 0
		{1500 * time.Microsecond, 1.5},
		{2 * time.Second, 2000},
	}

	for _, tt := range tests {
		got := durationToMillis(tt.d)
		if math.Abs(got-tt.expected) > 1e-9 {
			t.Errorf("durationToMillis(%v) = %v, want %v", tt.d, got, tt.expected)
		}
	}
}

func TestFormatRawResults(t *testing.T) {
	sample := func(n int, ns int64) testing.BenchmarkResult {
		return testing.BenchmarkResult{N: n, T: time.Duration(ns) * time.Duration(n), MemAllocs: 7, MemBytes: 1024}
	}
	comps := []benchmark.Comparison{{
		Name: "Insert unique values",
		Runs: []benchmark.Run{{
			Params:      benchmark.Params{ElementsToApply: 100_000},
			BwarrResult: benchmark.Result{Samples: []testing.BenchmarkResult{sample(1, 100), sample(1, 200)}},
			BTreeResult: benchmark.Result{Samples: []testing.BenchmarkResult{sample(2, 300)}},
		}},
	}}

	var buf bytes.Buffer
	err := formatRawResults(&buf, comps)
	if err != nil {
		t.Fatalf("formatRawResults: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"goos: ", "goarch: ", "btree-degree: 32\n",
		"Benchmarkinsert_unique_values/bwarr/100K-",
		"Benchmarkinsert_unique_values/btree/100K-",
		"ns/op", "B/op", "allocs/op",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	// One line per sample: 2 bwarr + 1 btree
	lines := 0
	for l := range strings.SplitSeq(out, "\n") {
		if strings.HasPrefix(l, "Benchmark") {
			lines++
		}
	}
	if lines != 3 {
		t.Errorf("got %d benchmark lines, want 3:\n%s", lines, out)
	}

	// bwarr lines must come before btree lines (deterministic order)
	if strings.Index(out, "/bwarr/") > strings.Index(out, "/btree/") {
		t.Errorf("bwarr results should precede btree results:\n%s", out)
	}
}

// syntheticComparison builds a comparison with fake results (mean and min..max spread)
// for graph tests, so no real benchmarks need to run.
func syntheticComparison() benchmark.Comparison {
	res := func(meanMs, spreadMs float64) benchmark.Result {
		mean := time.Duration(meanMs * float64(time.Millisecond))
		spread := time.Duration(spreadMs * float64(time.Millisecond))
		return benchmark.Result{ExecTimePerOp: mean, ExecTimeMin: mean - spread, ExecTimeMax: mean + spread}
	}
	return benchmark.Comparison{
		Name: "Synthetic",
		Runs: []benchmark.Run{
			// Deliberately out of order to check sorting by size
			{Params: benchmark.Params{ElementsToApply: 1_000_000}, BwarrResult: res(10, 2), BTreeResult: res(20, 1)},
			{Params: benchmark.Params{ElementsToApply: 100_000}, BwarrResult: res(0.8, 0.1), BTreeResult: res(0.15, 0.05)},
			{Params: benchmark.Params{ElementsToApply: 500_000}, BwarrResult: res(5, 0.5), BTreeResult: res(8, 3)},
		},
	}
}

func TestNewTimeSeries_SortedWithSpread(t *testing.T) {
	s := newTimeSeries(syntheticComparison().Runs, func(r benchmark.Run) benchmark.Result { return r.BwarrResult })

	if s.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", s.Len())
	}
	for i := 1; i < s.Len(); i++ {
		if s.XYs[i-1].X >= s.XYs[i].X {
			t.Errorf("points not sorted by X: %v", s.XYs)
		}
	}
	// 100K point: mean 0.8ms, min 0.7, max 0.9 -> error distances 0.1 and 0.1
	x, y := s.XY(0)
	low, high := s.YError(0)
	if x != 100 || math.Abs(y-0.8) > 1e-9 || math.Abs(low-0.1) > 1e-9 || math.Abs(high-0.1) > 1e-9 {
		t.Errorf("point 0 = (x=%v, y=%v, -%v, +%v), want (100, 0.8, 0.1, 0.1)", x, y, low, high)
	}
}

func TestGenerateTimeGraph_WritesPNG(t *testing.T) {
	out := filepath.Join(t.TempDir(), "synthetic.png")

	err := generateTimeGraph(syntheticComparison(), out)
	if err != nil {
		t.Fatalf("generateTimeGraph: %v", err)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}
