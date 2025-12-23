# bwarr-bench

Benchmarks for Black-White Array implementation ([dronnix/bwarr](https://github.com/dronnix/bwarr)) against [google/btree](https://github.com/google/btree).

This project uses Go's standard `testing.Benchmark` framework to compare the `Insert()` operation performance between bwarr and btree data structures with identical operations and reproducible results.

## Features

- **Standard Go benchmarking**: Uses `go test -bench` workflow
- **Identical operations guarantee**: Both data structures operate on the same pre-generated random int64 values
- **Reproducible benchmarks**: Fixed random seed (42) for consistent datasets
- **Multiple sizes**: 1024, 2048, 4096, 8192 operations per benchmark
- **Memory profiling**: Built-in allocation tracking with `-benchmem`
- **Statistical analysis**: Compatible with `benchstat` for comparing results

## Installation

```bash
git clone https://github.com/dronnix/bwarr-bench.git
cd bwarr-bench
```

Install benchstat for statistical comparison (optional):
```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## Quick Start

Run all benchmarks:
```bash
make bench
```

Example output:
```
BenchmarkBWArr_Insert_1024-10    16884    70441 ns/op   116.30 MB/s    36992 B/op     92 allocs/op
BenchmarkBTree_Insert_1024-10     6946   179284 ns/op    45.69 MB/s    67872 B/op   2321 allocs/op
```

## Usage

### Running Benchmarks

**All benchmarks (recommended settings, 10s each):**
```bash
make bench
```

**Quick benchmarks (faster, for development):**
```bash
make bench-quick
```

**Only bwarr benchmarks:**
```bash
make bench-bwarr
```

**Only btree benchmarks:**
```bash
make bench-btree
```

**Specific operation size:**
```bash
make bench-size SIZE=1024
# Or directly:
go test -bench=Insert_1024 -benchmem ./benchmark
```

**Longer benchmarks for accuracy:**
```bash
go test -bench=. -benchmem -benchtime=30s ./benchmark
```

**More iterations:**
```bash
go test -bench=. -benchmem -count=10 ./benchmark
```

### Understanding Output

Example benchmark output:
```
BenchmarkBWArr_Insert_1024-10    16884    70441 ns/op   116.30 MB/s    36992 B/op     92 allocs/op
```

Columns explained:
- `BenchmarkBWArr_Insert_1024-10`: Benchmark name and GOMAXPROCS
- `16884`: Number of iterations run (determined by testing framework)
- `70441 ns/op`: Average time per operation (nanoseconds)
- `116.30 MB/s`: Throughput in megabytes per second
- `36992 B/op`: Average bytes allocated per operation
- `92 allocs/op`: Average number of allocations per operation

### Generating Comparison Graphs

Generate PNG graphs comparing bwarr vs btree performance:
```bash
make bench-graph
```

This runs benchmarks programmatically for 10 seconds each and creates **two graphs**:

**1. Time comparison** (`benchmark_comparison.png`):
- X-axis: Dataset size (1024, 2048, 4096, 8192)
- Y-axis: Time in milliseconds
- Blue line with circles: bwarr performance
- Red line with squares: btree performance

**2. Allocations comparison** (`benchmark_comparison_allocs.png`):
- X-axis: Dataset size (1024, 2048, 4096, 8192)
- Y-axis: Allocations per operation
- Blue line with circles: bwarr allocations
- Red line with squares: btree allocations

The graph generation tool uses Go's `testing.Benchmark()` API internally, ensuring accurate and consistent results. The output PNGs are suitable for documentation and presentations.

**Custom benchmark duration:**
```bash
# Run benchmarks for 30 seconds each
./bin/benchgraph -test.benchtime=30s -output benchmark_comparison.png

# Or build and run directly
make build-benchgraph
./bin/benchgraph -test.benchtime=5s -output quick_benchmark.png
```

### Comparing Benchmarks

Save baseline results:
```bash
make bench > bench_old.txt
```

Make code changes, then compare:
```bash
make bench > bench_new.txt
benchstat bench_old.txt bench_new.txt
```

Example benchstat output:
```
name                  old time/op    new time/op    delta
BWArr_Insert_1024-10  70.4µs ± 2%    63.2µs ± 3%  -10.23%  (p=0.000 n=10+10)
BTree_Insert_1024-10   179µs ± 1%     176µs ± 2%   -1.68%  (p=0.001 n=10+10)

name                  old alloc/op   new alloc/op   delta
BWArr_Insert_1024-10  37.0kB ± 0%    35.2kB ± 0%   -4.86%  (p=0.000 n=10+10)
BTree_Insert_1024-10  67.9kB ± 0%    67.9kB ± 0%    0.00%  (p=1.000 n=10+10)

name                  old allocs/op  new allocs/op  delta
BWArr_Insert_1024-10  92.0 ± 0%      88.0 ± 0%      -4.35%  (p=0.000 n=10+10)
BTree_Insert_1024-10  2.32k ± 0%     2.32k ± 0%     0.00%  (p=1.000 n=10+10)
```

### Advanced Profiling

**CPU profiling:**
```bash
go test -bench=. -cpuprofile=cpu.prof ./benchmark
go tool pprof cpu.prof
```

**Memory profiling:**
```bash
go test -bench=. -memprofile=mem.prof ./benchmark
go tool pprof mem.prof
```

## Graphing Benchmark Results

### Built-in Graph Generator (Recommended)

The project includes a built-in graph generator that uses pure Go libraries:

```bash
make bench-graph
```

This creates two professional comparison graphs:
- `benchmark_comparison.png` - Time performance comparison
- `benchmark_comparison_allocs.png` - Memory allocations comparison

The tool:
- Runs benchmarks programmatically using `testing.Benchmark()` with 10s benchtime
- Uses gonum/plot for graph generation (pure Go, no external dependencies)
- Generates publication-ready PNG output for both time and allocations metrics
- Works on all platforms without requiring Python or external tools
- Supports custom benchmark duration via `-test.benchtime` flag

### Alternative: Using benchstat + CSV Export

Export benchstat results to CSV:
```bash
benchstat -format=csv bench_new.txt > results.csv
```

Import into Excel, Google Sheets, or use for plotting.

### Alternative: Using Python/matplotlib

Example Python script (`plot_benchmarks.py`):
```python
import re
import matplotlib.pyplot as plt

def parse_bench_output(filename):
    results = {}
    with open(filename) as f:
        for line in f:
            match = re.match(r'Benchmark(\w+)_Insert_(\d+)-\d+\s+\d+\s+(\d+)\s+ns/op', line)
            if match:
                ds, size, time = match.groups()
                results[(ds, int(size))] = int(time)
    return results

results = parse_bench_output('bench_new.txt')

sizes = [1024, 2048, 4096, 8192]
bwarr_times = [results.get(('BWArr', size), 0) for size in sizes]
btree_times = [results.get(('BTree', size), 0) for size in sizes]

plt.plot(sizes, bwarr_times, 'o-', label='bwarr')
plt.plot(sizes, btree_times, 's-', label='btree')
plt.xlabel('Number of Operations')
plt.ylabel('Time (ns)')
plt.title('Insert Performance Comparison')
plt.legend()
plt.grid(True)
plt.savefig('benchmark_comparison.png')
```

Run:
```bash
python plot_benchmarks.py
```

### Alternative: Using benchgraph

Install benchgraph:
```bash
go install github.com/bojanz/benchgraph@latest
```

Generate graph:
```bash
go test -bench=. -benchmem ./benchmark | tee results.txt
benchgraph results.txt output.png
```

## Benchmark Configuration

### Dataset Sizes
Default sizes: 1024, 2048, 4096, 8192

To add more sizes, edit `benchmark/datasets.go`:
```go
var Dataset16384 []int64

func init() {
    // Add to existing init()
    Dataset16384 = generateDataset(16384, BenchmarkSeed)
}
```

Then add benchmark function to `benchmark/benchmark_test.go`:
```go
func BenchmarkBWArr_Insert_16384(b *testing.B) {
    benchmarkBWarrWithDataset(b, Dataset16384)
}

func BenchmarkBTree_Insert_16384(b *testing.B) {
    benchmarkBTreeWithDataset(b, Dataset16384)
}
```

### Random Seed
Default seed: 42 (defined in `benchmark/datasets.go`)

To change, modify `BenchmarkSeed` constant.

### BTree Degree
Default: 2 (defined in `benchmark/datasets.go`)

To change, modify `BTreeDegree` constant.

## Development

### Running Tests
```bash
make test
# Or:
go test -v ./...
```

### Code Formatting
```bash
make fmt
# Or:
go fmt ./...
```

### Dependency Management
```bash
go mod tidy
```

## Implementation Details

### Identical Operations Guarantee

Both data structures operate on the exact same sequence of random int64 values, pre-generated once at package initialization in `benchmark/datasets.go`. This ensures a fair comparison.

**Pre-generation approach:**
- Package-level variables (`Dataset1024`, `Dataset2048`, etc.) initialized in `init()` function
- All benchmark functions reference the same dataset variables
- Fixed seed (42) ensures reproducibility across runs

### Measurement Methodology

- **Standard Go benchmarking**: Uses `testing.B` framework
- **Dataset pre-generation**: All random values generated once before any measurements
- **Tree creation excluded**: Data structure initialization is outside the timed section using `b.StopTimer()`/`b.StartTimer()`
- **Automatic iteration**: The testing framework determines optimal iteration count for statistical significance
- **Memory tracking**: `-benchmem` flag reports allocations per operation
- **Throughput reporting**: `b.SetBytes()` enables MB/s calculation

### Benchmark Function Structure

Each benchmark function:
1. Receives a pre-generated dataset (e.g., `Dataset1024`)
2. Calls `b.ReportAllocs()` to track memory
3. Calls `b.SetBytes()` to report throughput
4. Runs `b.N` iterations (determined by framework)
5. For each iteration:
   - Stops timer
   - Creates fresh tree (setup)
   - Starts timer
   - Inserts all N values (measured operation)

### API Differences

- **bwarr**: Uses `bwarr.New(cmp, capacity)` with comparison function + `Insert(value)`
- **btree**: Uses `btree.NewOrderedG[int64](degree)` + `ReplaceOrInsert(value)` (no `Insert()` method)
- **btree degree**: Fixed at 2 per benchmark configuration

## Makefile Targets

```
make test          - Run unit tests
make bench         - Run all benchmarks (10s each)
make bench-quick   - Run quick benchmarks (1s each)
make bench-bwarr   - Run only bwarr benchmarks
make bench-btree   - Run only btree benchmarks
make bench-size    - Run benchmarks for specific size (use SIZE=N)
make fmt           - Format Go source code
make help          - Show this help message
```

## Troubleshooting

**Benchmarks too fast/unstable:**
```bash
# Increase benchtime
go test -bench=. -benchmem -benchtime=30s ./benchmark
```

**Want more statistical samples:**
```bash
# Run 10 times and use benchstat
go test -bench=. -benchmem -count=10 ./benchmark > results.txt
benchstat results.txt
```

**benchstat not found:**
```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
