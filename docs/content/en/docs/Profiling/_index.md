---
title: Profiling azqr
description: This document describes how to profile azqr to identify performance bottlenecks and memory usage patterns for optimization.
weight: 8
---

## Overview

azqr includes built-in profiling capabilities that allow you to capture:
- **CPU Profile**: Identifies where your program spends most of its CPU time
- **Memory Profile**: Shows heap allocations and memory usage patterns
- **Execution Trace**: Provides detailed information about goroutine scheduling, garbage collection, and system calls

**Note**: Profiling features are only available in debug builds to keep production binaries lightweight and secure.

## Building Debug Version

To use profiling features, you must build azqr with the debug build tag:

```bash
# Build debug version with profiling support
make debug

# This creates: bin/linux_arm64/azqr-debug (or appropriate OS/arch)
```

The debug build includes:
- CPU profiling (`--cpu-profile`)
- Memory profiling (`--mem-profile`) 
- Execution trace profiling (`--trace-profile`)
- Debug logging and additional instrumentation

## Profiling Commands

### Basic Profiling

To enable profiling, use the new flags with the debug version of `azqr scan`:

```bash
# CPU profiling
azqr-debug scan --subscription-id "your-sub-id" --cpu-profile cpu.prof

# Memory profiling  
azqr-debug scan --subscription-id "your-sub-id" --mem-profile mem.prof

# Execution trace profiling
azqr-debug scan --subscription-id "your-sub-id" --trace-profile trace.prof

# Combined profiling (recommended for comprehensive analysis)
azqr-debug scan --subscription-id "your-sub-id" \
  --cpu-profile cpu.prof \
  --mem-profile mem.prof \
  --trace-profile trace.prof
```

**Note**: Replace `azqr-debug` with the actual path to your debug binary, e.g., `./bin/linux_arm64/azqr-debug`

## Analyzing Profile Data

### Prerequisites

Install Go profiling tools:
```bash
go install github.com/google/pprof@latest
```

**Optional: Install Graphviz for visual call graphs**
```bash
# Ubuntu/Debian
sudo apt-get install graphviz

# macOS
brew install graphviz

# Windows (using chocolatey)
choco install graphviz

# Or download from: https://graphviz.org/download/
```

> **Note**: Graphviz is only required for generating visual call graphs (PNG, SVG) and web-based graph views. The core profiling analysis works without it.

### CPU Profile Analysis

```bash
# Interactive analysis (no Graphviz required)
go tool pprof cpu.prof

# Web interface (recommended - no Graphviz required)
go tool pprof -http=:8080 cpu.prof

# Generate call graph (requires Graphviz)
go tool pprof -png cpu.prof > cpu_profile.png
go tool pprof -svg cpu.prof > cpu_profile.svg

# Show top 10 functions consuming CPU (no Graphviz required)
go tool pprof -top cpu.prof

# Text-based call graph (no Graphviz required)
go tool pprof -text cpu.prof
```

#### Key CPU Metrics to Look For:
- Functions with high **cumulative** time (including called functions)
- Functions with high **flat** time (excluding called functions)
- Hot spots in Azure SDK calls
- Expensive operations in scanners

### Memory Profile Analysis

```bash
# Interactive memory analysis (no Graphviz required)
go tool pprof mem.prof

# Web interface for memory (no Graphviz required)
go tool pprof -http=:8081 mem.prof

# Show memory allocations (no Graphviz required)
go tool pprof -alloc_space mem.prof

# Show objects in memory (no Graphviz required)
go tool pprof -inuse_objects mem.prof

# Generate memory allocation graph (requires Graphviz)
go tool pprof -png -alloc_space mem.prof > mem_alloc.png
```

#### Key Memory Metrics:
- **alloc_space**: Total memory allocated during execution
- **alloc_objects**: Total number of objects allocated
- **inuse_space**: Memory currently in use
- **inuse_objects**: Number of objects currently in use

### Execution Trace Analysis

```bash
# View trace in web browser (no Graphviz required)
go tool trace trace.prof
```

#### Trace Analysis Views:
- **Goroutine analysis**: Shows goroutine lifecycle and blocking
- **Network blocking profile**: Network I/O bottlenecks
- **Synchronization blocking profile**: Mutex and channel contention
- **Syscall blocking profile**: System call performance
- **Scheduler latency profile**: Goroutine scheduling delays

## Working Without Graphviz

If you don't have Graphviz installed, you can still perform comprehensive profiling analysis:

### Alternative Analysis Commands

```bash
# Text-based top functions
go tool pprof -text -cum cpu.prof | head -20

# Text-based call graph
go tool pprof -text cpu.prof

# Interactive command-line interface
go tool pprof cpu.prof
# Then use commands: top, list, web (if available), quit

# Web interface (works without Graphviz for most features)
go tool pprof -http=:8080 cpu.prof
# Note: Some graph views may not work, but tables and flamegraphs will
```

## Performance Optimization Strategies

### Based on Profiling Results

#### 1. CPU Optimization
- **High JSON Processing**: Consider streaming parsers for large responses
- **Expensive Reflection**: Cache reflection operations
- **HTTP Client Overhead**: Implement connection pooling
- **Azure SDK Calls**: Batch API calls where possible

#### 2. Memory Optimization
- **Large Object Allocations**: Use object pooling for frequently allocated objects
- **String Concatenations**: Use `strings.Builder` or pre-allocated buffers
- **Slice Growth**: Pre-allocate slices with known capacity
- **Memory Leaks**: Ensure proper cleanup of resources and references

#### 3. Concurrency Optimization
- **Goroutine Pools**: Limit concurrent operations to prevent resource exhaustion
- **Channel Buffering**: Use appropriately sized buffered channels
- **Context Timeouts**: Implement proper timeout handling
- **Rate Limiting**: Respect Azure API rate limits

### Red Flags
- Memory usage growing continuously (memory leaks)
- High GC pressure (> 50% CPU time in GC)
- Excessive goroutine creation (> 10,000 goroutines)
- Long-running HTTP requests without proper timeouts

## Contributing Performance Improvements

When submitting performance optimizations:

1. Include before/after profiling data
2. Provide benchmark results
3. Document the optimization approach
4. Ensure no functionality regressions

Run `make test` to ensure all tests pass after performance optimizations.