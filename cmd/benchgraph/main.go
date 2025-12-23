package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"sort"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/dronnix/bwarr-bench/benchmark"
)

const (
	// Dataset sizes for benchmarking
	size1024 = 1024
	size2048 = 2048
	size4096 = 4096
	size8192 = 8192

	// Graph styling constants
	lineWidth       = 2
	pointRadius     = 4
	colorMaxValue   = 255
	graphWidthInch  = 8
	graphHeightInch = 6
)

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

func main() {
	flag.Parse()

	log.Println("Running benchmarks...")

	// Create standard Runs for all benchmark sizes
	standardRuns := []benchmark.Run{
		{
			Params: benchmark.Params{
				N:          size1024,
				InitValues: benchmark.GenerateDataset(size1024, benchmark.Seed),
			},
		},
		{
			Params: benchmark.Params{
				N:          size2048,
				InitValues: benchmark.GenerateDataset(size2048, benchmark.Seed),
			},
		},
		{
			Params: benchmark.Params{
				N:          size4096,
				InitValues: benchmark.GenerateDataset(size4096, benchmark.Seed),
			},
		},
		{
			Params: benchmark.Params{
				N:          size8192,
				InitValues: benchmark.GenerateDataset(size8192, benchmark.Seed),
			},
		},
	}

	// Create Comparisons for different operations
	comparisons := []benchmark.Comparison{
		{
			Name:           "Insert Performance",
			BWArrBenchFunc: benchmark.BenchBWArrInsert,
			BTreeBenchFunc: benchmark.BenchBtreeInsert,
			Runs:           standardRuns,
		},
		{
			Name:           "Get Performance",
			BWArrBenchFunc: benchmark.BenchBWArrGet,
			BTreeBenchFunc: benchmark.BenchBTreeGet,
			Runs:           standardRuns,
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

		// Generate allocations graph
		allocsPath := baseName + "_allocs.png"
		err = generateAllocsGraph(comp, allocsPath)
		if err != nil {
			log.Fatalf("Error generating allocations graph for %s: %v", comp.Name, err)
		}
		log.Printf("Generated graph: %s", allocsPath)
		graphCount++
	}

	log.Printf("Done! Generated %d graphs", graphCount)
}

// generateTimeGraph creates a PNG graph comparing benchmark time results
func generateTimeGraph(comparison benchmark.Comparison, outputPath string) error {
	// Create new plot
	p := plot.New()

	p.Title.Text = "Benchmark Comparison: BWArr vs BTree Insert Performance"
	p.X.Label.Text = "Dataset Size (N)"
	p.Y.Label.Text = "Time (microseconds)"

	// Prepare data points for each implementation
	bwarrPoints := make(plotter.XYs, 0, len(comparison.Runs))
	btreePoints := make(plotter.XYs, 0, len(comparison.Runs))

	for _, run := range comparison.Runs {
		// Convert time.Duration to microseconds
		bwarrMicros := float64(run.BwarrResult.ExecTimePerOp.Microseconds())
		btreeMicros := float64(run.BTreeResult.ExecTimePerOp.Microseconds())

		bwarrPoints = append(bwarrPoints, plotter.XY{X: float64(run.N), Y: bwarrMicros})
		btreePoints = append(btreePoints, plotter.XY{X: float64(run.N), Y: btreeMicros})
	}

	// Sort points by X (size) for proper line drawing
	sort.Slice(bwarrPoints, func(i, j int) bool {
		return bwarrPoints[i].X < bwarrPoints[j].X
	})
	sort.Slice(btreePoints, func(i, j int) bool {
		return btreePoints[i].X < btreePoints[j].X
	})

	// Create line and points for bwarr
	bwarrLine, bwarrPts, err := plotter.NewLinePoints(bwarrPoints)
	if err != nil {
		return fmt.Errorf("creating bwarr line: %w", err)
	}
	bwarrLine.Color = color.RGBA{R: 0, G: 0, B: colorMaxValue, A: colorMaxValue} // Blue
	bwarrLine.Width = vg.Points(lineWidth)
	bwarrPts.Shape = draw.CircleGlyph{}
	bwarrPts.Color = color.RGBA{R: 0, G: 0, B: colorMaxValue, A: colorMaxValue}
	bwarrPts.Radius = vg.Points(pointRadius)

	// Create line and points for btree
	btreeLine, btreePts, err := plotter.NewLinePoints(btreePoints)
	if err != nil {
		return fmt.Errorf("creating btree line: %w", err)
	}
	btreeLine.Color = color.RGBA{R: colorMaxValue, G: 0, B: 0, A: colorMaxValue} // Red
	btreeLine.Width = vg.Points(lineWidth)
	btreePts.Shape = draw.BoxGlyph{}
	btreePts.Color = color.RGBA{R: colorMaxValue, G: 0, B: 0, A: colorMaxValue}
	btreePts.Radius = vg.Points(pointRadius)

	// Add to plot
	p.Add(bwarrLine, bwarrPts, btreeLine, btreePts)
	p.Legend.Add("bwarr", bwarrLine, bwarrPts)
	p.Legend.Add("btree", btreeLine, btreePts)
	p.Legend.Top = true
	p.Legend.Left = true

	// Add grid
	p.Add(plotter.NewGrid())

	// Save as PNG
	err = p.Save(graphWidthInch*vg.Inch, graphHeightInch*vg.Inch, outputPath)
	if err != nil {
		return fmt.Errorf("saving plot: %w", err)
	}

	return nil
}

// generateAllocsGraph creates a PNG graph comparing benchmark allocations results
func generateAllocsGraph(comparison benchmark.Comparison, outputPath string) error {
	// Create new plot
	p := plot.New()

	p.Title.Text = "Benchmark Comparison: bwarr vs btree Allocations"
	p.X.Label.Text = "Dataset Size (N)"
	p.Y.Label.Text = "Allocations per Operation"

	// Prepare data points for each implementation
	bwarrPoints := make(plotter.XYs, 0, len(comparison.Runs))
	btreePoints := make(plotter.XYs, 0, len(comparison.Runs))

	for _, run := range comparison.Runs {
		bwarrAllocs := float64(run.BwarrResult.AllocsPerOp)
		btreeAllocs := float64(run.BTreeResult.AllocsPerOp)

		bwarrPoints = append(bwarrPoints, plotter.XY{X: float64(run.N), Y: bwarrAllocs})
		btreePoints = append(btreePoints, plotter.XY{X: float64(run.N), Y: btreeAllocs})
	}

	// Sort points by X (size) for proper line drawing
	sort.Slice(bwarrPoints, func(i, j int) bool {
		return bwarrPoints[i].X < bwarrPoints[j].X
	})
	sort.Slice(btreePoints, func(i, j int) bool {
		return btreePoints[i].X < btreePoints[j].X
	})

	// Create line and points for bwarr
	bwarrLine, bwarrPts, err := plotter.NewLinePoints(bwarrPoints)
	if err != nil {
		return fmt.Errorf("creating bwarr line: %w", err)
	}
	bwarrLine.Color = color.RGBA{R: 0, G: 0, B: colorMaxValue, A: colorMaxValue} // Blue
	bwarrLine.Width = vg.Points(lineWidth)
	bwarrPts.Shape = draw.CircleGlyph{}
	bwarrPts.Color = color.RGBA{R: 0, G: 0, B: colorMaxValue, A: colorMaxValue}
	bwarrPts.Radius = vg.Points(pointRadius)

	// Create line and points for btree
	btreeLine, btreePts, err := plotter.NewLinePoints(btreePoints)
	if err != nil {
		return fmt.Errorf("creating btree line: %w", err)
	}
	btreeLine.Color = color.RGBA{R: colorMaxValue, G: 0, B: 0, A: colorMaxValue} // Red
	btreeLine.Width = vg.Points(lineWidth)
	btreePts.Shape = draw.BoxGlyph{}
	btreePts.Color = color.RGBA{R: colorMaxValue, G: 0, B: 0, A: colorMaxValue}
	btreePts.Radius = vg.Points(pointRadius)

	// Add to plot
	p.Add(bwarrLine, bwarrPts, btreeLine, btreePts)
	p.Legend.Add("bwarr", bwarrLine, bwarrPts)
	p.Legend.Add("btree", btreeLine, btreePts)
	p.Legend.Top = true
	p.Legend.Left = true

	// Add grid
	p.Add(plotter.NewGrid())

	// Save as PNG
	err = p.Save(graphWidthInch*vg.Inch, graphHeightInch*vg.Inch, outputPath)
	if err != nil {
		return fmt.Errorf("saving plot: %w", err)
	}

	return nil
}
