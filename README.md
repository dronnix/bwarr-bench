# BWArr benchamrks

Performance benchmarks comparing [BWArr](https://github.com/dronnix/bwarr) against [Google BTree](https://github.com/google/btree).

See [benchmark.md](benchmark.md) for detailed results and graphs.

Feel free to add more data structures and benchmark them!

## Quick Start

```bash
# Run standard benchmarks
make bench

# Generate performance graphs
make run
```

Graphs are saved to `images/` directory.

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

## License

MIT License
