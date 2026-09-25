# BWArr benchmarks

Performance benchmarks comparing [BWArr](https://github.com/dronnix/bwarr) against [Google BTree](https://github.com/google/btree).

See [benchmark.md](benchmark.md) for detailed results and graphs.

Feel free to add more data structures and benchmark them!

## Quick Start

```bash
# Run standard benchmarks
make bench

# Generate performance graphs (5 repetitions per point by default)
make run

# More repetitions and a longer measured window per repetition
./bin/benchgraph -count=10 -test.benchtime=3s
```

Graphs are saved to `images/` directory. Each time graph shows the mean over all
repetitions with error bars for the min..max spread. Raw per-repetition results are
written to `results/benchmarks.txt` in Go benchmark format, so they can be compared
with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
benchstat results/benchmarks.txt
```

## Benchmark Operations

- **Insert** - Insert unique values into empty data structure
- **Get** - Look up values by key
- **Iterate** - Traverse all values (ordered/unordered)
- **Delete** - Remove all values

Feel free to add more operations!

## Make Targets

```
make test         - Run tests
make bench        - Run all benchmarks (10s each)
make bench-quick  - Quick benchmarks (1s each)
make run          - Generate graphs in images/
make fmt          - Format code
make lint         - Run linter
```

## Configuration

- **Dataset sizes**: 100K, 250K, 500K, 1M, 2M, 4M elements
- **BTree degree**: 32
- **Repetitions per point**: 5 (`-count`), bwarr and btree interleaved within each repetition
- **Measured window per repetition**: 1s (`-test.benchtime`)

## License

MIT License
