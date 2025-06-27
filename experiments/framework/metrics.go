package framework

import (
	"encoding/json"
	"time"
)

// Metrics collects performance and complexity measurements for experiments
type Metrics struct {
	Implementation ImplementationMetrics `json:"implementation"`
	Performance    PerformanceMetrics    `json:"performance"`
	Complexity     ComplexityMetrics     `json:"complexity"`
	Timestamp      time.Time            `json:"timestamp"`
}

// ImplementationMetrics tracks implementation effort and characteristics
type ImplementationMetrics struct {
	LinesOfCode       int               `json:"lines_of_code"`
	FilesChanged      int               `json:"files_changed"`
	Dependencies      []string          `json:"dependencies"`
	ImplementationTime time.Duration     `json:"implementation_time"`
	TestCoverage      float64           `json:"test_coverage"`
	ErrorHandling     ErrorHandlingInfo `json:"error_handling"`
}

// PerformanceMetrics tracks runtime performance characteristics
type PerformanceMetrics struct {
	LatencyP50       time.Duration `json:"latency_p50"`
	LatencyP95       time.Duration `json:"latency_p95"`
	LatencyP99       time.Duration `json:"latency_p99"`
	TokenOverhead    float64       `json:"token_overhead"`
	MemoryUsageMB    float64       `json:"memory_usage_mb"`
	CPUUsagePercent  float64       `json:"cpu_usage_percent"`
	OperationsPerSec int           `json:"operations_per_sec"`
	ErrorRate        float64       `json:"error_rate"`
}

// ComplexityMetrics tracks code and conceptual complexity
type ComplexityMetrics struct {
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	CognitiveComplexity  int     `json:"cognitive_complexity"`
	InterfaceCount       int     `json:"interface_count"`
	ConfigurationItems   int     `json:"configuration_items"`
	SetupSteps          int     `json:"setup_steps"`
	MaintenanceScore    float64 `json:"maintenance_score"`
}

// ErrorHandlingInfo tracks error handling characteristics
type ErrorHandlingInfo struct {
	ErrorTypes     []string `json:"error_types"`
	RecoveryMethods []string `json:"recovery_methods"`
	GracefulDegradation bool  `json:"graceful_degradation"`
}

// MetricsCollector provides methods to collect and store metrics
type MetricsCollector struct {
	startTime time.Time
	metrics   *Metrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
		metrics: &Metrics{
			Timestamp: time.Now(),
		},
	}
}

// RecordImplementation records implementation-related metrics
func (mc *MetricsCollector) RecordImplementation(linesOfCode, filesChanged int, dependencies []string, coverage float64) {
	mc.metrics.Implementation = ImplementationMetrics{
		LinesOfCode:        linesOfCode,
		FilesChanged:       filesChanged,
		Dependencies:       dependencies,
		ImplementationTime: time.Since(mc.startTime),
		TestCoverage:       coverage,
	}
}

// RecordPerformance records performance-related metrics
func (mc *MetricsCollector) RecordPerformance(latencies []time.Duration, tokenOverhead, memoryMB, cpuPercent float64, opsPerSec int, errorRate float64) {
	if len(latencies) == 0 {
		return
	}

	// Calculate percentiles (simple implementation)
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)
	
	mc.metrics.Performance = PerformanceMetrics{
		LatencyP50:       percentile(sortedLatencies, 0.5),
		LatencyP95:       percentile(sortedLatencies, 0.95),
		LatencyP99:       percentile(sortedLatencies, 0.99),
		TokenOverhead:    tokenOverhead,
		MemoryUsageMB:    memoryMB,
		CPUUsagePercent:  cpuPercent,
		OperationsPerSec: opsPerSec,
		ErrorRate:        errorRate,
	}
}

// RecordComplexity records complexity-related metrics
func (mc *MetricsCollector) RecordComplexity(cyclomatic, cognitive, interfaces, configItems, setupSteps int, maintenanceScore float64) {
	mc.metrics.Complexity = ComplexityMetrics{
		CyclomaticComplexity: cyclomatic,
		CognitiveComplexity:  cognitive,
		InterfaceCount:       interfaces,
		ConfigurationItems:   configItems,
		SetupSteps:          setupSteps,
		MaintenanceScore:    maintenanceScore,
	}
}

// GetMetrics returns the collected metrics
func (mc *MetricsCollector) GetMetrics() *Metrics {
	return mc.metrics
}

// ToJSON serializes metrics to JSON
func (mc *MetricsCollector) ToJSON() ([]byte, error) {
	return json.MarshalIndent(mc.metrics, "", "  ")
}

// percentile calculates the nth percentile of a sorted slice
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}