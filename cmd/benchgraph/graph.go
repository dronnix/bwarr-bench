package main

import (
	"fmt"
	"image/color"
	"sort"
	"time"

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

// durationToMillis converts a duration to milliseconds as a float.
// Duration.Milliseconds() truncates to an integer, which turns sub-millisecond
// results (e.g. iteration over 100K elements) into 0 on the graph.
func durationToMillis(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / float64(time.Millisecond)
}

// timeSeries holds mean time points with their min..max spread, for plotting
// a line with error bars. It implements plotter.XYer and plotter.YErrorer.
type timeSeries struct {
	plotter.XYs

	errs plotter.YErrors
}

func (s timeSeries) YError(i int) (low, high float64) { return s.errs[i].Low, s.errs[i].High }

// newTimeSeries converts results to milliseconds and sorts them by X (size).
func newTimeSeries(runs []benchmark.Run, pick func(benchmark.Run) benchmark.Result) timeSeries {
	idx := make([]int, len(runs))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return runs[idx[a]].ElementsToApply < runs[idx[b]].ElementsToApply })

	s := timeSeries{XYs: make(plotter.XYs, 0, len(runs)), errs: make(plotter.YErrors, 0, len(runs))}
	for _, i := range idx {
		res := pick(runs[i])
		mean := durationToMillis(res.ExecTimePerOp)
		s.XYs = append(s.XYs, plotter.XY{X: float64(runs[i].ElementsToApply) / thousandDivisor, Y: mean})
		// YErrors are distances from the point: Low goes down, High goes up.
		s.errs = append(s.errs, struct{ Low, High float64 }{
			Low:  mean - durationToMillis(res.ExecTimeMin),
			High: durationToMillis(res.ExecTimeMax) - mean,
		})
	}
	return s
}

// generateTimeGraph creates a PNG graph comparing benchmark time results.
// Points are the mean over all repetitions; error bars span the min..max repetition.
func generateTimeGraph(comparison benchmark.Comparison, outputPath string) error {
	// Create new plot
	p := plot.New()

	p.Title.Text = "Benchmark Comparison: BWArr vs BTree - " + comparison.Name
	p.X.Label.Text = xAxisLabel
	p.Y.Label.Text = "Time (milliseconds, mean; bars = min..max over repetitions)"

	// Prepare data points for each implementation, sorted by X (size)
	bwarrSeries := newTimeSeries(comparison.Runs, func(r benchmark.Run) benchmark.Result { return r.BwarrResult })
	btreeSeries := newTimeSeries(comparison.Runs, func(r benchmark.Run) benchmark.Result { return r.BTreeResult })
	bwarrPoints := bwarrSeries.XYs
	btreePoints := btreeSeries.XYs

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

	// Error bars showing the min..max spread over repetitions
	bwarrErr, err := plotter.NewYErrorBars(bwarrSeries)
	if err != nil {
		return fmt.Errorf("creating bwarr error bars: %w", err)
	}
	bwarrErr.Color = bwarrLine.Color
	btreeErr, err := plotter.NewYErrorBars(btreeSeries)
	if err != nil {
		return fmt.Errorf("creating btree error bars: %w", err)
	}
	btreeErr.Color = btreeLine.Color

	// Add to plot
	p.Add(bwarrLine, bwarrPts, bwarrErr, btreeLine, btreePts, btreeErr)
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

	for i := range comparison.Runs {
		run := &comparison.Runs[i]
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

	for i := range comparison.Runs {
		run := &comparison.Runs[i]
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
