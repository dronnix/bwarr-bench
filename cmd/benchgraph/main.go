package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
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

func main() { //nolint:funlen
	// Register testing flags (e.g. -test.benchtime) so testing.Benchmark honours them.
	testing.Init()
	count := flag.Int("count", defaultCount, "number of repetitions per data point")
	resultsPath := flag.String("results", defaultResultsPath, "file for raw results in Go benchmark format (empty to skip)")
	flag.Parse()

	log.Printf("Running benchmarks (%d repetitions per point)...", *count)

	// Helper function to create standard runs
	createStandardRuns := func() []benchmark.Run {
		return []benchmark.Run{
			{
				Params: benchmark.Params{
					ElementsToApply: size100K,
					InitValues:      benchmark.GenerateRandomDataset(size100K, benchmark.Seed, math.MaxInt64),
				},
			},
			{
				Params: benchmark.Params{
					ElementsToApply: size250K,
					InitValues:      benchmark.GenerateRandomDataset(size250K, benchmark.Seed, math.MaxInt64),
				},
			},
			{
				Params: benchmark.Params{
					ElementsToApply: size500K,
					InitValues:      benchmark.GenerateRandomDataset(size500K, benchmark.Seed, math.MaxInt64),
				},
			},
			{
				Params: benchmark.Params{
					ElementsToApply: size1M,
					InitValues:      benchmark.GenerateRandomDataset(size1M, benchmark.Seed, math.MaxInt64),
				},
			},
			{
				Params: benchmark.Params{
					ElementsToApply: size2M,
					InitValues:      benchmark.GenerateRandomDataset(size2M, benchmark.Seed, math.MaxInt64),
				},
			},
			{
				Params: benchmark.Params{
					ElementsToApply: size4M,
					InitValues:      benchmark.GenerateRandomDataset(size4M, benchmark.Seed, math.MaxInt64),
				},
			},
		}
	}

	// Create Comparisons for different operations (each gets its own Runs slice)
	comparisons := []benchmark.Comparison{
		{
			Name:           "Insert unique values",
			BWArrBenchFunc: benchmark.BenchBWArrInsert,
			BTreeBenchFunc: benchmark.BenchBTreeInsert,
			Runs:           createStandardRuns(),
			MeasureAllocs:  true,
		},
		{
			Name:           "Get all values by key",
			BWArrBenchFunc: benchmark.BenchBWArrGet,
			BTreeBenchFunc: benchmark.BenchBTreeGet,
			Runs:           createStandardRuns(),
			MeasureAllocs:  false,
		},
		{
			Name:           "Ordered iteration over all values",
			BWArrBenchFunc: benchmark.BenchBWArrOrderedIterate,
			BTreeBenchFunc: benchmark.BenchBTreeOrderedIterate,
			Runs:           createStandardRuns(),
			MeasureAllocs:  false,
		},
		{
			Name:           "Unordered iteration over all values",
			BWArrBenchFunc: benchmark.BenchBWArrUnorderedIterate,
			BTreeBenchFunc: benchmark.BenchBTreeOrderedIterate,
			Runs:           createStandardRuns(),
			MeasureAllocs:  false,
		},
		{
			Name:           "Delete all values",
			BWArrBenchFunc: benchmark.BenchBWArrDelete,
			BTreeBenchFunc: benchmark.BenchBTreeDelete,
			Runs:           createStandardRuns(),
			MeasureAllocs:  false,
		},
	}

	// Execute all comparisons
	log.Println("Executing benchmarks...")
	for i := range comparisons {
		log.Printf("Executing %s...", comparisons[i].Name)
		comparisons[i].Execute(*count)
	}

	// Write raw results before drawing, so a failure in plotting does not lose the data
	if *resultsPath != "" {
		err := writeRawResults(*resultsPath, comparisons)
		if err != nil {
			log.Fatalf("Error writing raw results: %v", err)
		}
		log.Printf("Wrote raw results: %s", *resultsPath)
	}

	// Create images directory
	imagesDir := "images"
	err := os.MkdirAll(imagesDir, 0755) //nolint:mnd // Standard directory permissions
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
func writeRawResults(path string, comparisons []benchmark.Comparison) error {
	err := os.MkdirAll(filepath.Dir(path), 0755) //nolint:mnd // Standard directory permissions
	if err != nil {
		return fmt.Errorf("creating results directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating results file: %w", err)
	}

	err = formatRawResults(f, comparisons)
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

// formatRawResults is the io.Writer-based core of writeRawResults, split out for testing.
func formatRawResults(w io.Writer, comparisons []benchmark.Comparison) error {
	// Header lines in "key: value" form are treated as configuration by benchstat.
	_, err := fmt.Fprintf(w, "goos: %s\ngoarch: %s\ngoversion: %s\nbtree-degree: %d\n",
		runtime.GOOS, runtime.GOARCH, runtime.Version(), benchmark.BTreeDegree)
	if err != nil {
		return fmt.Errorf("writing header: %w", err)
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
