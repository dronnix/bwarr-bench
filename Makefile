# Makefile for bwarr-bench

.PHONY: all test bench bench-quick bench-bwarr bench-btree bench-size fmt help

all: test

test:
	go test -v ./...

bench:
	go test -bench=. -benchmem -benchtime=10s ./benchmark

bench-quick:
	go test -bench=. -benchmem -benchtime=1s ./benchmark

bench-bwarr:
	go test -bench=BenchmarkBWArr -benchmem -benchtime=10s ./benchmark

bench-btree:
	go test -bench=BenchmarkBTree -benchmem -benchtime=10s ./benchmark

bench-size:
	go test -bench=Insert_$(SIZE) -benchmem -benchtime=10s ./benchmark

fmt:
	go fmt ./...

help:
	@echo "Available targets:"
	@echo "  make test          - Run unit tests"
	@echo "  make bench         - Run all benchmarks (10s each)"
	@echo "  make bench-quick   - Run quick benchmarks (1s each)"
	@echo "  make bench-bwarr   - Run only bwarr benchmarks"
	@echo "  make bench-btree   - Run only btree benchmarks"
	@echo "  make bench-size    - Run benchmarks for specific size (use SIZE=N)"
	@echo "  make fmt           - Format Go source code"
	@echo "  make help          - Show this help message"
