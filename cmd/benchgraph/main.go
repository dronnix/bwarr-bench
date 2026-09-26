package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/dronnix/bwarr-bench/benchmark"
)

const (
	// Dataset sizes for benchmarking
	size100K = 100_000
	size250K = 250_000
	size500K = 500_000
	size1M   = 1_000_000
	size2M   = 2_000_000
	size4M   = 4_000_000
)

const (
	// defaultCount is the number of repetitions per data point. Each repetition is a
	// full testing.Benchmark run (about -test.benchtime of measured work), so the
	// total run time grows linearly with this value.
	defaultCount = 5

	// defaultResultsPath is where raw per-repetition results are written in Go
	// benchmark format, so they can be committed and analysed with benchstat.
	defaultResultsPath = "results/benchmarks.txt"
)

// standardSizes returns the dataset sizes every comparison is run at.
func standardSizes() []int {
	return []int{size100K, size250K, size500K, size1M, size2M, size4M}
}

func main() {
	// Register testing flags (e.g. -test.benchtime) so testing.Benchmark honours them.
	testing.Init()
	count := flag.Int("count", defaultCount, "number of repetitions per data point")
	resultsPath := flag.String("results", defaultResultsPath, "file for raw results in Go benchmark format (empty to skip)")
	benchPattern := flag.String("bench", "",
		"run only comparisons whose file name or title matches this regexp (case-insensitive, like go test -bench); empty runs all")
	list := flag.Bool("list", false, "print the available comparisons and exit")
	flag.Parse()

	all := buildComparisons()

	if *list {
		printComparisons(os.Stdout, all)
		return
	}

	comparisons, err := selectComparisons(all, *benchPattern)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Datasets are generated only now, only for the selected comparisons, and shared
	// between them: every comparison uses the same seed and sizes, so the values are
	// identical and the benchmarks only read them.
	attachRuns(comparisons)

	log.Printf("Running %d of %d comparisons (%d repetitions per point)...", len(comparisons), len(all), *count)

	// Execute the selected comparisons
	for i := range comparisons {
		log.Printf("Executing %s...", comparisons[i].Name)
		comparisons[i].Execute(*count)
	}

	// Write raw results before drawing, so a failure in plotting does not lose the data.
	// Lines of comparisons that were not run this time are kept.
	if *resultsPath != "" {
		err := writeRawResults(*resultsPath, comparisons)
		if err != nil {
			log.Fatalf("Error writing raw results: %v", err)
		}
		log.Printf("Wrote raw results: %s", *resultsPath)
	}

	// Create images directory
	imagesDir := "images"
	err = os.MkdirAll(imagesDir, 0755) //nolint:mnd // Standard directory permissions
	if err != nil {
		log.Fatalf("Error creating images directory: %v", err)
	}

	// Generate graphs for all comparisons
	log.Println("Generating graphs...")
	graphCount := 0
	for _, comp := range comparisons {
		// Generate time graph
		baseName := sanitizeFilename(comp.Name)
		timePath := filepath.Join(imagesDir, baseName+".png")
		err := generateTimeGraph(comp, timePath)
		if err != nil {
			log.Fatalf("Error generating time graph for %s: %v", comp.Name, err)
		}
		log.Printf("Generated graph: %s", timePath)
		graphCount++

		// Generate allocations graph only if MeasureAllocs is enabled
		if comp.MeasureAllocs {
			allocsPath := filepath.Join(imagesDir, baseName+"_allocs.png")
			err = generateAllocsGraph(comp, allocsPath)
			if err != nil {
				log.Fatalf("Error generating allocations graph for %s: %v", comp.Name, err)
			}
			log.Printf("Generated graph: %s", allocsPath)
			graphCount++

			// Generate allocated bytes graph
			bytesPath := filepath.Join(imagesDir, baseName+"_bytes.png")
			err = generateBytesGraph(comp, bytesPath)
			if err != nil {
				log.Fatalf("Error generating bytes graph for %s: %v", comp.Name, err)
			}
			log.Printf("Generated graph: %s", bytesPath)
			graphCount++
		}
	}

	log.Printf("Done! Generated %d graphs", graphCount)
}

// attachRuns gives every comparison one Run per standard size. The dataset for each
// size is generated once and shared by all comparisons (they only read it), so the
// cost is one copy of each size instead of one per comparison, and nothing is
// allocated for comparisons that were filtered out.
func attachRuns(comparisons []benchmark.Comparison) {
	sizes := standardSizes()
	datasets := make(map[int][]int64, len(sizes))
	for _, n := range sizes {
		datasets[n] = benchmark.GenerateRandomDataset(n, benchmark.Seed, math.MaxInt64)
	}
	for i := range comparisons {
		runs := make([]benchmark.Run, 0, len(sizes))
		for _, n := range sizes {
			runs = append(runs, benchmark.Run{Params: benchmark.Params{
				ElementsToApply: n,
				InitValues:      datasets[n],
			}})
		}
		comparisons[i].Runs = runs
	}
}

// buildComparisons defines every graph the tool produces. It returns metadata only;
// datasets are attached later by attachRuns, after -list / -bench filtering.
func buildComparisons() []benchmark.Comparison {
	return []benchmark.Comparison{
		{
			Name:           "Insert unique values",
			BWArrBenchFunc: benchmark.BenchBWArrInsert,
			BTreeBenchFunc: benchmark.BenchBTreeInsert,
			MeasureAllocs:  true,
		},
		{
			Name:           "Get all values by key",
			BWArrBenchFunc: benchmark.BenchBWArrGet,
			BTreeBenchFunc: benchmark.BenchBTreeGet,
			MeasureAllocs:  false,
		},
		{
			Name:           "Ordered iteration over all values",
			BWArrBenchFunc: benchmark.BenchBWArrOrderedIterate,
			BTreeBenchFunc: benchmark.BenchBTreeOrderedIterate,
			MeasureAllocs:  false,
		},
		{
			Name:           "Unordered iteration over all values",
			BWArrBenchFunc: benchmark.BenchBWArrUnorderedIterate,
			BTreeBenchFunc: benchmark.BenchBTreeOrderedIterate,
			MeasureAllocs:  false,
		},
		{
			Name:           "Delete all values",
			BWArrBenchFunc: benchmark.BenchBWArrDelete,
			BTreeBenchFunc: benchmark.BenchBTreeDelete,
			MeasureAllocs:  false,
		},
	}
}

// printComparisons lists every comparison as "<file name>\t<title>", one per line.
func printComparisons(w io.Writer, comparisons []benchmark.Comparison) {
	for i := range comparisons {
		fmt.Fprintf(w, "%s\t%s\n", sanitizeFilename(comparisons[i].Name), comparisons[i].Name)
	}
}

// selectComparisons returns the comparisons whose file name or title matches pattern,
// a case-insensitive regular expression in the spirit of `go test -bench`. An empty
// pattern selects everything. It is an error if nothing matches.
func selectComparisons(all []benchmark.Comparison, pattern string) ([]benchmark.Comparison, error) {
	if pattern == "" {
		return all, nil
	}
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid -bench pattern %q: %w", pattern, err)
	}

	var selected []benchmark.Comparison
	for i := range all {
		c := &all[i]
		if re.MatchString(sanitizeFilename(c.Name)) || re.MatchString(c.Name) {
			selected = append(selected, *c)
		}
	}
	if len(selected) == 0 {
		names := make([]string, 0, len(all))
		for i := range all {
			names = append(names, sanitizeFilename(all[i].Name))
		}
		return nil, fmt.Errorf("no comparison matches -bench %q; available: %s", pattern, strings.Join(names, ", "))
	}
	return selected, nil
}

// sanitizeFilename converts a comparison name to a valid filename
// Example: "Insert Performance" → "insert_performance"
func sanitizeFilename(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")
	// Remove special characters (keep only alphanumeric and underscores)
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// writeRawResults writes every repetition of every run in the standard Go benchmark
// output format, which benchstat and similar tools can read. One line per repetition:
//
//	Benchmarkinsert_unique_values/bwarr/100K-18   16   70000344 ns/op   8528439 B/op   127 allocs/op
//
// If the file already exists, result lines of comparisons that are not in this run are
// kept, so running a subset with -bench does not discard the other comparisons' data.
// Kept lines are only valid under the same header (OS, architecture, Go version, btree
// degree); if the existing file was produced in a different environment the merge is
// refused, because benchstat would attribute the old lines to the new environment.
func writeRawResults(path string, comparisons []benchmark.Comparison) error {
	err := os.MkdirAll(filepath.Dir(path), 0755) //nolint:mnd // Standard directory permissions
	if err != nil {
		return fmt.Errorf("creating results directory: %w", err)
	}

	kept, oldHeader, err := readOtherResults(path, comparisons)
	if err != nil {
		return err
	}
	if len(kept) > 0 && !slices.Equal(oldHeader, resultsHeader()) {
		return fmt.Errorf("%s holds results from a different environment (%s) than this one (%s); "+
			"delete the file or run all comparisons", path, strings.Join(oldHeader, ", "), strings.Join(resultsHeader(), ", "))
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating results file: %w", err)
	}

	err = formatRawResults(f, kept, comparisons)
	if err != nil {
		_ = f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("closing results file: %w", err)
	}
	return nil
}

// resultsHeader returns the configuration lines written at the top of the results file.
// benchstat treats "key: value" lines as configuration of the results that follow.
func resultsHeader() []string {
	return []string{
		"goos: " + runtime.GOOS,
		"goarch: " + runtime.GOARCH,
		"goversion: " + runtime.Version(),
		fmt.Sprintf("btree-degree: %d", benchmark.BTreeDegree),
	}
}

// readOtherResults returns the result lines of an existing results file that do not
// belong to any of the given comparisons, together with the file's header lines.
// A missing file yields no lines and no error.
func readOtherResults(path string, comparisons []benchmark.Comparison) (kept, header []string, err error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("opening existing results file: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only file

	prefixes := make([]string, 0, len(comparisons))
	for i := range comparisons {
		prefixes = append(prefixes, "Benchmark"+sanitizeFilename(comparisons[i].Name)+"/")
	}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "Benchmark"):
			if !belongsTo(line, prefixes) {
				kept = append(kept, line)
			}
		case strings.Contains(line, ": "):
			header = append(header, line)
		}
	}
	err = sc.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("reading existing results file: %w", err)
	}
	return kept, header, nil
}

// belongsTo reports whether a result line starts with one of the comparison prefixes.
func belongsTo(line string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}

// formatRawResults is the io.Writer-based core of writeRawResults, split out for testing.
// keptLines are written verbatim after the header, before the new results.
func formatRawResults(w io.Writer, keptLines []string, comparisons []benchmark.Comparison) error {
	// Header lines in "key: value" form are treated as configuration by benchstat.
	_, err := fmt.Fprintln(w, strings.Join(resultsHeader(), "\n"))
	if err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	for _, line := range keptLines {
		_, err = fmt.Fprintln(w, line)
		if err != nil {
			return fmt.Errorf("writing kept line: %w", err)
		}
	}

	procs := runtime.GOMAXPROCS(0)
	for _, comp := range comparisons {
		name := "Benchmark" + sanitizeFilename(comp.Name)
		for i := range comp.Runs {
			run := &comp.Runs[i]
			size := fmt.Sprintf("%dK", run.ElementsToApply/1000) //nolint:mnd // Sizes are always whole thousands
			// Fixed order so the file is deterministic and diff-friendly
			impls := []struct {
				name string
				res  benchmark.Result
			}{{"bwarr", run.BwarrResult}, {"btree", run.BTreeResult}}
			for _, impl := range impls {
				for _, s := range impl.res.Samples {
					_, err = fmt.Fprintf(w, "%s/%s/%s-%d\t%s\t%s\n", name, impl.name, size, procs, s.String(), s.MemString())
					if err != nil {
						return fmt.Errorf("writing result line: %w", err)
					}
				}
			}
		}
	}
	return nil
}
