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
		Name:   "Insert unique values",
		Series: []benchmark.Series{{Name: seriesBwarr}, {Name: "btree (key+value)"}},
		Runs: []benchmark.Run{{
			Params: benchmark.Params{ElementsToApply: 100_000},
			Results: []benchmark.Result{
				{Samples: []testing.BenchmarkResult{sample(1, 100), sample(1, 200)}},
				{Samples: []testing.BenchmarkResult{sample(2, 300)}},
			},
		}},
	}}

	var buf bytes.Buffer
	err := formatRawResults(&buf, []string{"Benchmarkother/bwarr/100K-1\t1\t5 ns/op"}, comps)
	if err != nil {
		t.Fatalf("formatRawResults: %v", err)
	}
	out := buf.String()

	// Kept lines come right after the header, before the new results
	if strings.Index(out, "Benchmarkother/") > strings.Index(out, "Benchmarkinsert_unique_values/") {
		t.Errorf("kept lines should precede new results:\n%s", out)
	}

	for _, want := range []string{
		"goos: ", "goarch: ", "btree-degree: 32\n",
		"Benchmarkinsert_unique_values/bwarr/100K-",
		"Benchmarkinsert_unique_values/btree_keyvalue/100K-",
		"ns/op", "B/op", "allocs/op",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	// One line per sample (2 bwarr + 1 btree) plus the kept line
	lines := 0
	for l := range strings.SplitSeq(out, "\n") {
		if strings.HasPrefix(l, "Benchmark") {
			lines++
		}
	}
	if lines != 4 {
		t.Errorf("got %d benchmark lines, want 4:\n%s", lines, out)
	}

	// bwarr lines must come before btree lines (deterministic order)
	if strings.Index(out, "/bwarr/") > strings.Index(out, "/btree_keyvalue/") {
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
		Name:   "Synthetic",
		Series: []benchmark.Series{{Name: seriesBwarr}, {Name: seriesBtree}, {Name: "third"}},
		Runs: []benchmark.Run{
			// Deliberately out of order to check sorting by size
			{Params: benchmark.Params{ElementsToApply: 1_000_000}, Results: []benchmark.Result{res(10, 2), res(20, 1), res(15, 0)}},
			{Params: benchmark.Params{ElementsToApply: 100_000}, Results: []benchmark.Result{res(0.8, 0.1), res(0.15, 0.05), res(0.5, 0)}},
			{Params: benchmark.Params{ElementsToApply: 500_000}, Results: []benchmark.Result{res(5, 0.5), res(8, 3), res(6, 0)}},
		},
	}
}

func TestNewSeries_SortedWithSpread(t *testing.T) {
	s := newSeries(syntheticComparison().Runs, 0, timeMetric())

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
	if !s.hasErrors() {
		t.Error("time series with spread should report hasErrors() == true")
	}

	// Third series has zero spread, so no error bars should be drawn for it
	if newSeries(syntheticComparison().Runs, 2, timeMetric()).hasErrors() {
		t.Error("series without spread should report hasErrors() == false")
	}
	// Allocation metric never has error bars
	if newSeries(syntheticComparison().Runs, 0, allocsMetric()).hasErrors() {
		t.Error("allocs metric should report hasErrors() == false")
	}
}

func TestGenerateGraph_WritesPNG(t *testing.T) {
	for _, m := range []struct {
		name string
		m    metric
	}{{"time", timeMetric()}, {"allocs", allocsMetric()}, {"bytes", bytesMetric()}} {
		t.Run(m.name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "synthetic.png")

			comp := syntheticComparison()
			err := generateGraph(&comp, m.m, out)
			if err != nil {
				t.Fatalf("generateGraph: %v", err)
			}

			info, err := os.Stat(out)
			if err != nil {
				t.Fatalf("output file missing: %v", err)
			}
			if info.Size() == 0 {
				t.Error("output file is empty")
			}
		})
	}
}

func TestFileBase(t *testing.T) {
	if got := fileBase(&benchmark.Comparison{Name: "Get all values", FileName: "custom"}); got != "custom" {
		t.Errorf("explicit FileName: got %q, want %q", got, "custom")
	}
	if got := fileBase(&benchmark.Comparison{Name: "Get all values"}); got != "get_all_values" {
		t.Errorf("derived FileName: got %q, want %q", got, "get_all_values")
	}
}

func TestBuildComparisons(t *testing.T) {
	comps := buildComparisons()
	seen := map[string]bool{}
	for i := range comps {
		c := &comps[i]
		if len(c.Series) < 2 {
			t.Errorf("%s: needs at least 2 series, has %d", c.Name, len(c.Series))
		}
		// Metadata only: no datasets are generated until attachRuns
		if len(c.Runs) != 0 {
			t.Errorf("%s: buildComparisons should not attach runs, got %d", c.Name, len(c.Runs))
		}
		base := fileBase(c)
		if seen[base] {
			t.Errorf("duplicate output file base %q", base)
		}
		seen[base] = true
	}
}

func TestAttachRuns_SharesDatasets(t *testing.T) {
	comps := []benchmark.Comparison{{FileName: fileInsert}, {FileName: fileGet}} // names are not used here
	attachRuns(comps)

	sizes := standardSizes()
	for i := range comps {
		c := &comps[i]
		if len(c.Runs) != len(sizes) {
			t.Fatalf("%s: %d runs, want %d", c.Name, len(c.Runs), len(sizes))
		}
		for j, r := range c.Runs {
			if r.ElementsToApply != sizes[j] || len(r.InitValues) != sizes[j] {
				t.Errorf("%s: run %d has %d elements / %d values, want %d", c.Name, j, r.ElementsToApply, len(r.InitValues), sizes[j])
			}
		}
	}

	// The same size shares one backing array across comparisons
	for j := range sizes {
		a, b := comps[0].Runs[j].InitValues, comps[1].Runs[j].InitValues
		if &a[0] != &b[0] {
			t.Errorf("size %d: datasets are not shared between comparisons", sizes[j])
		}
	}
}

func TestWriteRawResults_RefusesDifferentEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "benchmarks.txt")
	foreign := "goos: plan9\ngoarch: mips\ngoversion: go0.1\nbtree-degree: 32\n" +
		"Benchmark" + fileGet + "/bwarr/100K-1\t1\t5 ns/op\n"
	err := os.WriteFile(path, []byte(foreign), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	insert := benchmark.Comparison{
		Name: "Insert", FileName: fileInsert,
		Series: []benchmark.Series{{Name: "bwarr"}},
		Runs: []benchmark.Run{{
			Params:  benchmark.Params{ElementsToApply: 100_000},
			Results: []benchmark.Result{{Samples: []testing.BenchmarkResult{{N: 1, T: 100}}}},
		}},
	}

	// Kept lines from another environment: refuse
	err = writeRawResults(path, []benchmark.Comparison{insert})
	if err == nil || !strings.Contains(err.Error(), "different environment") {
		t.Fatalf("expected a different-environment error, got %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != foreign {
		t.Error("the existing file must be left untouched when the merge is refused")
	}

	// Nothing to keep (the only comparison in the file is being rerun): overwrite is fine
	get := insert
	get.Name, get.FileName = "Lookup", fileGet
	err = writeRawResults(path, []benchmark.Comparison{get})
	if err != nil {
		t.Fatalf("rerun of the only stored comparison should overwrite: %v", err)
	}
	data, _ = os.ReadFile(path)
	if strings.Contains(string(data), "plan9") {
		t.Error("old header should be replaced when no lines are kept")
	}
}

// File names of comparisons used in the selection tests.
const (
	fileInsert  = "insert_unique_values"
	fileReplace = "replaceorinsert_unique_collection"
	fileGet     = "get_all_values_by_key"
	fileDelete  = "delete_all_values"
)

func TestSelectComparisons(t *testing.T) {
	all := []benchmark.Comparison{
		{Name: "Insert (duplicates allowed)", FileName: fileInsert},
		{Name: "ReplaceOrInsert (unique collection)"},
		{Name: "Lookup values by key", FileName: fileGet},
		{Name: "Delete all values"},
	}
	names := func(cs []benchmark.Comparison) []string {
		out := make([]string, 0, len(cs))
		for i := range cs {
			out = append(out, fileBase(&cs[i]))
		}
		return out
	}

	tests := []struct {
		pattern string
		want    []string
	}{
		{"", []string{fileInsert, fileReplace, fileGet, fileDelete}},
		{"insert", []string{fileInsert, fileReplace}}, // matches both, case-insensitive
		{"^insert", []string{fileInsert}},
		{"Lookup", []string{fileGet}},                 // title match
		{"get|delete", []string{fileGet, fileDelete}}, // alternation
		{"REPLACEORINSERT", []string{fileReplace}},
	}
	for _, tt := range tests {
		got, err := selectComparisons(all, tt.pattern)
		if err != nil {
			t.Errorf("pattern %q: unexpected error: %v", tt.pattern, err)
			continue
		}
		if strings.Join(names(got), ",") != strings.Join(tt.want, ",") {
			t.Errorf("pattern %q: got %v, want %v", tt.pattern, names(got), tt.want)
		}
	}

	_, err := selectComparisons(all, "nothing_matches")
	if err == nil {
		t.Error("expected an error when nothing matches")
	}
	_, err = selectComparisons(all, "(")
	if err == nil {
		t.Error("expected an error for an invalid regexp")
	}
}

func TestWriteRawResults_KeepsOtherComparisons(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results", "benchmarks.txt")
	sample := testing.BenchmarkResult{N: 1, T: 100}
	insert := benchmark.Comparison{
		Name: "Insert", FileName: fileInsert,
		Series: []benchmark.Series{{Name: "bwarr"}},
		Runs: []benchmark.Run{{
			Params:  benchmark.Params{ElementsToApply: 100_000},
			Results: []benchmark.Result{{Samples: []testing.BenchmarkResult{sample}}},
		}},
	}
	get := insert
	get.Name, get.FileName = "Get", fileGet

	// First run: both comparisons (file does not exist yet)
	err := writeRawResults(path, []benchmark.Comparison{insert, get})
	if err != nil {
		t.Fatalf("first write: %v", err)
	}

	// Second run: only insert, with a distinguishable sample
	insert.Runs[0].Results[0].Samples = []testing.BenchmarkResult{{N: 1, T: 999}}
	err = writeRawResults(path, []benchmark.Comparison{insert})
	if err != nil {
		t.Fatalf("second write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading results: %v", err)
	}
	out := string(data)

	if strings.Count(out, "goos: ") != 1 {
		t.Errorf("header should be written exactly once:\n%s", out)
	}
	if !strings.Contains(out, "Benchmark"+fileGet+"/bwarr/100K") {
		t.Errorf("results of the comparison not in the second run were lost:\n%s", out)
	}
	if strings.Count(out, "Benchmark"+fileInsert+"/") != 1 || !strings.Contains(out, "999") {
		t.Errorf("insert results should be replaced by the second run, once:\n%s", out)
	}
}

func TestPrintComparisons(t *testing.T) {
	var buf bytes.Buffer
	err := printComparisons(&buf, buildComparisons())
	if err != nil {
		t.Fatalf("printComparisons: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != len(buildComparisons()) {
		t.Fatalf("got %d lines, want %d", len(lines), len(buildComparisons()))
	}
	if !strings.HasPrefix(lines[0], fileInsert+"\t") {
		t.Errorf("first line = %q, want it to start with the insert file name", lines[0])
	}
}
