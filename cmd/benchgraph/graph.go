package main

import (
	"fmt"
	"image/color"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/dronnix/bwarr-bench/benchmark"
)

const (
	// Graph styling constants
	lineWidth       = 2
	pointRadius     = 4
	colorMaxValue   = 255
	graphWidthInch  = 8
	graphHeightInch = 6

	// Graph labels
	xAxisLabel = "Dataset Size (thousands of elements)"

	// Unit conversions
	thousandDivisor = 1000.0
	bytesToKB       = 1024.0
)

// generateTimeGraph creates a PNG graph comparing benchmark time results
func generateTimeGraph(comparison benchmark.Comparison, outputPath string) error {
	// Create new plot
	p := plot.New()

	p.Title.Text = "Benchmark Comparison: BWArr vs BTree - " + comparison.Name
	p.X.Label.Text = xAxisLabel
	p.Y.Label.Text = "Time (milliseconds)"

	// Prepare data points for each implementation
	bwarrPoints := make(plotter.XYs, 0, len(comparison.Runs))
	btreePoints := make(plotter.XYs, 0, len(comparison.Runs))

	for _, run := range comparison.Runs {
		// Convert time.Duration to milliseconds
		bwarrMillis := float64(run.BwarrResult.ExecTimePerOp.Milliseconds())
		btreeMillis := float64(run.BTreeResult.ExecTimePerOp.Milliseconds())
		// Convert ElementsToApply to thousands for X-axis
		xValue := float64(run.ElementsToApply) / thousandDivisor

		bwarrPoints = append(bwarrPoints, plotter.XY{X: xValue, Y: bwarrMillis})
		btreePoints = append(btreePoints, plotter.XY{X: xValue, Y: btreeMillis})
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
//
//nolint:dupl // Intentionally similar to generateBytesGraph, different metric extraction
func generateAllocsGraph(comparison benchmark.Comparison, outputPath string) error {
	// Create new plot
	p := plot.New()

	p.Title.Text = "Benchmark Comparison: BWArr vs BTree - " + comparison.Name + " (Allocations)"
	p.X.Label.Text = xAxisLabel
	p.Y.Label.Text = "Allocations per Operation"

	// Prepare data points for each implementation
	bwarrPoints := make(plotter.XYs, 0, len(comparison.Runs))
	btreePoints := make(plotter.XYs, 0, len(comparison.Runs))

	for _, run := range comparison.Runs {
		bwarrAllocs := float64(run.BwarrResult.AllocsPerOp)
		btreeAllocs := float64(run.BTreeResult.AllocsPerOp)
		// Convert ElementsToApply to thousands for X-axis
		xValue := float64(run.ElementsToApply) / thousandDivisor

		bwarrPoints = append(bwarrPoints, plotter.XY{X: xValue, Y: bwarrAllocs})
		btreePoints = append(btreePoints, plotter.XY{X: xValue, Y: btreeAllocs})
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

// generateBytesGraph creates a PNG graph comparing benchmark allocated bytes results
//
//nolint:dupl // Intentionally similar to generateAllocsGraph, different metric extraction
func generateBytesGraph(comparison benchmark.Comparison, outputPath string) error {
	// Create new plot
	p := plot.New()

	p.Title.Text = "Benchmark Comparison: BWArr vs BTree - " + comparison.Name + " (Bytes)"
	p.X.Label.Text = xAxisLabel
	p.Y.Label.Text = "Allocated KB per Operation"

	// Prepare data points for each implementation
	bwarrPoints := make(plotter.XYs, 0, len(comparison.Runs))
	btreePoints := make(plotter.XYs, 0, len(comparison.Runs))

	for _, run := range comparison.Runs {
		// Convert bytes to kilobytes
		bwarrKB := float64(run.BwarrResult.AllocBytesPerOp) / bytesToKB
		btreeKB := float64(run.BTreeResult.AllocBytesPerOp) / bytesToKB
		// Convert ElementsToApply to thousands for X-axis
		xValue := float64(run.ElementsToApply) / thousandDivisor

		bwarrPoints = append(bwarrPoints, plotter.XY{X: xValue, Y: bwarrKB})
		btreePoints = append(btreePoints, plotter.XY{X: xValue, Y: btreeKB})
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
