package framework

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// BenchmarkConfig defines configuration for benchmark runs
type BenchmarkConfig struct {
	Name            string        `json:"name"`
	Iterations      int           `json:"iterations"`
	Warmup          int           `json:"warmup"`
	Timeout         time.Duration `json:"timeout"`
	ParallelWorkers int           `json:"parallel_workers"`
}

// BenchmarkResult contains the results of a benchmark run
type BenchmarkResult struct {
	Config      BenchmarkConfig `json:"config"`
	Latencies   []time.Duration `json:"latencies"`
	Errors      []string        `json:"errors"`
	MemoryStats MemoryStats     `json:"memory_stats"`
	Duration    time.Duration   `json:"total_duration"`
	Timestamp   time.Time       `json:"timestamp"`
}

// MemoryStats tracks memory usage during benchmarks
type MemoryStats struct {
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
}

// Operation represents a single operation to benchmark
type Operation func(ctx context.Context) error

// Benchmark runs performance benchmarks for operations
type Benchmark struct {
	config BenchmarkConfig
	op     Operation
}

// NewBenchmark creates a new benchmark
func NewBenchmark(config BenchmarkConfig, op Operation) *Benchmark {
	return &Benchmark{
		config: config,
		op:     op,
	}
}

// Run executes the benchmark and returns results
func (b *Benchmark) Run(ctx context.Context) (*BenchmarkResult, error) {
	result := &BenchmarkResult{
		Config:    b.config,
		Latencies: make([]time.Duration, 0, b.config.Iterations),
		Errors:    make([]string, 0),
		Timestamp: time.Now(),
	}

	// Create context with timeout
	if b.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.config.Timeout)
		defer cancel()
	}

	startTime := time.Now()
	defer func() {
		result.Duration = time.Since(startTime)
	}()

	// Warmup runs
	for i := 0; i < b.config.Warmup; i++ {
		if err := b.op(ctx); err != nil {
			return nil, fmt.Errorf("warmup failed: %w", err)
		}
	}

	// Capture initial memory stats
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Run actual benchmark
	if b.config.ParallelWorkers > 1 {
		result = b.runParallel(ctx, result)
	} else {
		result = b.runSequential(ctx, result)
	}

	// Capture final memory stats
	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	result.MemoryStats = MemoryStats{
		AllocMB:      float64(memAfter.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(memAfter.TotalAlloc) / 1024 / 1024,
		SysMB:        float64(memAfter.Sys) / 1024 / 1024,
		NumGC:        memAfter.NumGC - memBefore.NumGC,
	}

	return result, nil
}

// runSequential runs benchmark operations sequentially
func (b *Benchmark) runSequential(ctx context.Context, result *BenchmarkResult) *BenchmarkResult {
	for i := 0; i < b.config.Iterations; i++ {
		select {
		case <-ctx.Done():
			result.Errors = append(result.Errors, "benchmark cancelled")
			return result
		default:
		}

		start := time.Now()
		if err := b.op(ctx); err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Latencies = append(result.Latencies, time.Since(start))
		}
	}
	return result
}

// runParallel runs benchmark operations in parallel
func (b *Benchmark) runParallel(ctx context.Context, result *BenchmarkResult) *BenchmarkResult {
	type benchResult struct {
		latency time.Duration
		err     error
	}

	results := make(chan benchResult, b.config.Iterations)
	sem := make(chan struct{}, b.config.ParallelWorkers)

	// Start workers
	for i := 0; i < b.config.Iterations; i++ {
		go func() {
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			start := time.Now()
			err := b.op(ctx)
			results <- benchResult{
				latency: time.Since(start),
				err:     err,
			}
		}()
	}

	// Collect results
	for i := 0; i < b.config.Iterations; i++ {
		select {
		case <-ctx.Done():
			result.Errors = append(result.Errors, "benchmark cancelled")
			return result
		case res := <-results:
			if res.err != nil {
				result.Errors = append(result.Errors, res.err.Error())
			} else {
				result.Latencies = append(result.Latencies, res.latency)
			}
		}
	}

	return result
}

// DefaultBenchmarkConfig returns a sensible default benchmark configuration
func DefaultBenchmarkConfig(name string) BenchmarkConfig {
	return BenchmarkConfig{
		Name:            name,
		Iterations:      1000,
		Warmup:          10,
		Timeout:         5 * time.Minute,
		ParallelWorkers: 1,
	}
}

// QuickBenchmarkConfig returns a configuration for quick testing
func QuickBenchmarkConfig(name string) BenchmarkConfig {
	return BenchmarkConfig{
		Name:            name,
		Iterations:      10,
		Warmup:          2,
		Timeout:         30 * time.Second,
		ParallelWorkers: 1,
	}
}