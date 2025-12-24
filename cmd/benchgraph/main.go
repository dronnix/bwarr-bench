package main

import (
	"flag"
	"log"
	"math"
	"strings"

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

func main() {
	flag.Parse()

	log.Println("Running benchmarks...")

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
			BTreeBenchFunc: benchmark.BenchBtreeInsert,
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
	}

	// Execute all comparisons
	log.Println("Executing benchmarks...")
	for i := range comparisons {
		log.Printf("Executing %s...", comparisons[i].Name)
		comparisons[i].Execute()
	}

	// Generate graphs for all comparisons
	log.Println("Generating graphs...")
	graphCount := 0
	for _, comp := range comparisons {
		// Generate time graph
		baseName := sanitizeFilename(comp.Name)
		timePath := baseName + ".png"
		err := generateTimeGraph(comp, timePath)
		if err != nil {
			log.Fatalf("Error generating time graph for %s: %v", comp.Name, err)
		}
		log.Printf("Generated graph: %s", timePath)
		graphCount++

		// Generate allocations graph only if MeasureAllocs is enabled
		if comp.MeasureAllocs {
			allocsPath := baseName + "_allocs.png"
			err = generateAllocsGraph(comp, allocsPath)
			if err != nil {
				log.Fatalf("Error generating allocations graph for %s: %v", comp.Name, err)
			}
			log.Printf("Generated graph: %s", allocsPath)
			graphCount++

			// Generate allocated bytes graph
			bytesPath := baseName + "_bytes.png"
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
