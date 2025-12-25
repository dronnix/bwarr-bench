# Makefile for bwarr-bench

.PHONY: all test bench bench-quick build run fmt lint help

all: test

test:
	go test -v ./...

bench:
	go test -bench=. -benchmem -benchtime=10s ./benchmark

bench-quick:
	go test -bench=. -benchmem -benchtime=1s ./benchmark

build:
	go build -o bin/benchgraph ./cmd/benchgraph

run: build
	./bin/benchgraph

fmt:
	go fmt ./...

lint:
	golangci-lint run -c .golangci.yml