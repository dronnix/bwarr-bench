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

// seriesStyle is the colour and point glyph of one line on a graph.
type seriesStyle struct {
	color color.RGBA
	glyph draw.GlyphDrawer
}

// seriesStyles returns the styles assigned to Comparison.Series by index, so the first
// series (conventionally bwarr) is always blue circles and the second (btree) red boxes.
func seriesStyles() []seriesStyle {
	//nolint:mnd // RGB colour components
	return []seriesStyle{
		{color.RGBA{R: 0, G: 0, B: colorMaxValue, A: colorMaxValue}, draw.CircleGlyph{}}, // blue
		{color.RGBA{R: colorMaxValue, G: 0, B: 0, A: colorMaxValue}, draw.BoxGlyph{}},    // red
		{color.RGBA{R: 0, G: 150, B: 0, A: colorMaxValue}, draw.TriangleGlyph{}},         // green
		{color.RGBA{R: 148, G: 0, B: 211, A: colorMaxValue}, draw.PyramidGlyph{}},        // purple
		{color.RGBA{R: colorMaxValue, G: 140, B: 0, A: colorMaxValue}, draw.PlusGlyph{}}, // orange
		{color.RGBA{R: 0, G: 139, B: 139, A: colorMaxValue}, draw.CrossGlyph{}},          // teal
	}
}

// metric selects which Result field is plotted and how it is labelled.
// extract returns the plotted value and the downward/upward error-bar distances
// (both zero when the metric has no meaningful spread).
type metric struct {
	yLabel      string
	titleSuffix string
	extract     func(benchmark.Result) (value, errLow, errHigh float64)
}

// timeMetric plots mean time per operation with min..max error bars.
func timeMetric() metric {
	return metric{
		yLabel: "Time (milliseconds, mean; bars = min..max over repetitions)",
		extract: func(r benchmark.Result) (float64, float64, float64) {
			mean := durationToMillis(r.ExecTimePerOp)
			return mean, mean - durationToMillis(r.ExecTimeMin), durationToMillis(r.ExecTimeMax) - mean
		},
	}
}

// allocsMetric plots allocations per operation.
func allocsMetric() metric {
	return metric{
		yLabel:      "Allocations per Operation",
		titleSuffix: " (Allocations)",
		extract: func(r benchmark.Result) (float64, float64, float64) {
			return float64(r.AllocsPerOp), 0, 0
		},
	}
}

// bytesMetric plots allocated kilobytes per operation.
func bytesMetric() metric {
	return metric{
		yLabel:      "Allocated KB per Operation",
		titleSuffix: " (Bytes)",
		extract: func(r benchmark.Result) (float64, float64, float64) {
			return float64(r.AllocBytesPerOp) / bytesToKB, 0, 0
		},
	}
}

// durationToMillis converts a duration to milliseconds as a float.
// Duration.Milliseconds() truncates to an integer, which turns sub-millisecond
// results (e.g. iteration over 100K elements) into 0 on the graph.
func durationToMillis(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / float64(time.Millisecond)
}

// series holds the points of one line with their error-bar distances, sorted by X.
// It implements plotter.XYer and plotter.YErrorer.
type series struct {
	plotter.XYs

	errs plotter.YErrors
}

func (s series) YError(i int) (low, high float64) { return s.errs[i].Low, s.errs[i].High }

// hasErrors reports whether any point has a non-zero spread, i.e. error bars are worth drawing.
func (s series) hasErrors() bool {
	for _, e := range s.errs {
		if e.Low != 0 || e.High != 0 {
			return true
		}
	}
	return false
}

// newSeries extracts metric m for series index idx from every run, sorted by dataset size.
func newSeries(runs []benchmark.Run, idx int, m metric) series {
	order := make([]int, len(runs))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool { return runs[order[a]].ElementsToApply < runs[order[b]].ElementsToApply })

	s := series{XYs: make(plotter.XYs, 0, len(runs)), errs: make(plotter.YErrors, 0, len(runs))}
	for _, i := range order {
		run := &runs[i]
		value, errLow, errHigh := m.extract(run.Results[idx])
		s.XYs = append(s.XYs, plotter.XY{X: float64(run.ElementsToApply) / thousandDivisor, Y: value})
		// YErrors are distances from the point: Low goes down, High goes up.
		s.errs = append(s.errs, struct{ Low, High float64 }{Low: errLow, High: errHigh})
	}
	return s
}

// generateGraph creates a PNG graph with one line per Comparison.Series for the given metric.
func generateGraph(comparison *benchmark.Comparison, m metric, outputPath string) error {
	p := plot.New()
	styles := seriesStyles()

	p.Title.Text = "Benchmark Comparison: " + comparison.Name + m.titleSuffix
	p.X.Label.Text = xAxisLabel
	p.Y.Label.Text = m.yLabel

	for idx, ser := range comparison.Series {
		style := styles[idx%len(styles)]
		data := newSeries(comparison.Runs, idx, m)

		line, pts, err := plotter.NewLinePoints(data.XYs)
		if err != nil {
			return fmt.Errorf("creating line for %q: %w", ser.Name, err)
		}
		line.Color = style.color
		line.Width = vg.Points(lineWidth)
		pts.Shape = style.glyph
		pts.Color = style.color
		pts.Radius = vg.Points(pointRadius)
		p.Add(line, pts)

		if data.hasErrors() {
			bars, err := plotter.NewYErrorBars(data)
			if err != nil {
				return fmt.Errorf("creating error bars for %q: %w", ser.Name, err)
			}
			bars.Color = style.color
			p.Add(bars)
		}

		p.Legend.Add(ser.Name, line, pts)
	}
	p.Legend.Top = true
	p.Legend.Left = true

	p.Add(plotter.NewGrid())

	err := p.Save(graphWidthInch*vg.Inch, graphHeightInch*vg.Inch, outputPath)
	if err != nil {
		return fmt.Errorf("saving plot: %w", err)
	}
	return nil
}
