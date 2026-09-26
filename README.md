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

# Only some comparisons: -bench is a case-insensitive regexp matched against the
# file name and the title of each comparison (like `go test -bench`). Default: all.
./bin/benchgraph -list                 # show available names
./bin/benchgraph -bench=get            # one comparison
./bin/benchgraph -bench='insert|delete'  # several
make run ARGS="-bench=^insert"
```

A partial run regenerates only the matching graphs and replaces only their lines in
`results/benchmarks.txt`; results of the other comparisons are kept. The kept lines
must come from the same environment (OS, architecture, Go version, btree degree); if
the file was produced elsewhere the run stops with an error instead of mixing results.

Graphs are saved to `images/` directory. Each time graph shows the mean over all
repetitions with error bars for the min..max spread. Raw per-repetition results are
written to `results/benchmarks.txt` in Go benchmark format, so they can be compared
with [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
benchstat results/benchmarks.txt
```

## Benchmark Operations

- **Insert (duplicates allowed)** - bwarr `Insert` vs btree `ReplaceOrInsert` (btree doesn't support a plain `Insert`)
- **ReplaceOrInsert (unique collection)** - set semantics on both sides; the like-for-like number
- **Insert increasing / decreasing sequence** - case 1 operations on keys that arrive already sorted (timestamps, IDs)
- **Get** - Look up values by key
- **Iterate** - Traverse all values (ordered/unordered)
- **Delete** - Remove all values
- **Mixed workload** - Insert, Get and Delete interleaved 1:1:1 in random order on a structure of steady size N

Feel free to add more operations!

## Make Targets

```
make test         - Run tests
make bench        - Run all benchmarks (10s each)
make bench-quick  - Quick benchmarks (1s each)
make run          - Generate graphs in images/ (ARGS="-bench=..." for a subset)
make fmt          - Format code
make lint         - Run linter
```

## Configuration

- **Dataset sizes**: 100K, 250K, 500K, 1M, 2M, 4M elements
- **BTree degree**: 32
- **Repetitions per point**: 5 (`-count`), all series interleaved within each repetition
- **Measured window per repetition**: 1s (`-test.benchtime`)
- **GC**: a full GC runs before each timed section starts; GC work caused by the timed operation itself is measured

## License

MIT License
