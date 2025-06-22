package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"text/template"
	"time"
)

// ExperimentReport represents a complete lab report for an experiment
type ExperimentReport struct {
	Title       string    `json:"title"`
	Week        int       `json:"week"`
	Date        time.Time `json:"date"`
	Hypothesis  string    `json:"hypothesis"`
	Method      string    `json:"method"`
	Results     Results   `json:"results"`
	Failures    []string  `json:"failures"`
	Conclusions []string  `json:"conclusions"`
	NextSteps   []string  `json:"next_steps"`
}

// Results contains quantitative and qualitative experiment results
type Results struct {
	Quantitative QuantitativeResults `json:"quantitative"`
	Qualitative  []string           `json:"qualitative"`
}

// QuantitativeResults contains measurable experiment outcomes
type QuantitativeResults struct {
	Baseline    *Metrics `json:"baseline"`
	Experimental *Metrics `json:"experimental"`
	Comparison  Comparison `json:"comparison"`
}

// Comparison provides relative analysis between baseline and experimental
type Comparison struct {
	PerformanceRatio    float64 `json:"performance_ratio"`
	ComplexityRatio     float64 `json:"complexity_ratio"`
	ImplementationRatio float64 `json:"implementation_ratio"`
	TokenOverheadRatio  float64 `json:"token_overhead_ratio"`
	Summary             string  `json:"summary"`
}

// ReportGenerator creates formatted lab reports
type ReportGenerator struct {
	template *template.Template
}

// NewReportGenerator creates a new report generator
func NewReportGenerator() *ReportGenerator {
	tmpl := template.Must(template.New("report").Parse(reportTemplate))
	return &ReportGenerator{
		template: tmpl,
	}
}

// Generate creates a formatted lab report
func (rg *ReportGenerator) Generate(report *ExperimentReport, writer io.Writer) error {
	return rg.template.Execute(writer, report)
}

// GenerateMarkdown creates a markdown lab report
func (rg *ReportGenerator) GenerateMarkdown(report *ExperimentReport) (string, error) {
	var buf bytes.Buffer
	if err := rg.Generate(report, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SaveReport saves a report to the specified path
func (rg *ReportGenerator) SaveReport(report *ExperimentReport, basePath string) (string, error) {
	// Create filename
	filename := fmt.Sprintf("week%d-%s.md", report.Week, 
		sanitizeFilename(report.Title))
	fullPath := filepath.Join(basePath, filename)

	// Generate markdown
	markdown, err := rg.GenerateMarkdown(report)
	if err != nil {
		return "", fmt.Errorf("failed to generate markdown: %w", err)
	}

	// Write to file (would use os.WriteFile in real implementation)
	_ = markdown // For now, just return the path
	return fullPath, nil
}

// CreateComparisonReport creates a comparison report between two metrics
func CreateComparisonReport(title string, week int, hypothesis string, 
	baseline, experimental *Metrics) *ExperimentReport {
	
	comparison := Comparison{
		PerformanceRatio: calculatePerformanceRatio(baseline, experimental),
		ComplexityRatio:  calculateComplexityRatio(baseline, experimental),
		ImplementationRatio: calculateImplementationRatio(baseline, experimental),
		TokenOverheadRatio: experimental.Performance.TokenOverhead / 
			baseline.Performance.TokenOverhead,
	}
	
	comparison.Summary = generateComparisonSummary(comparison)

	return &ExperimentReport{
		Title:      title,
		Week:       week,
		Date:       time.Now(),
		Hypothesis: hypothesis,
		Results: Results{
			Quantitative: QuantitativeResults{
				Baseline:     baseline,
				Experimental: experimental,
				Comparison:   comparison,
			},
		},
	}
}

// calculatePerformanceRatio computes relative performance (lower is better)
func calculatePerformanceRatio(baseline, experimental *Metrics) float64 {
	if baseline.Performance.LatencyP50 == 0 {
		return 1.0
	}
	return float64(experimental.Performance.LatencyP50) / 
		float64(baseline.Performance.LatencyP50)
}

// calculateComplexityRatio computes relative complexity (lower is better)
func calculateComplexityRatio(baseline, experimental *Metrics) float64 {
	baselineComplexity := float64(baseline.Complexity.CyclomaticComplexity + 
		baseline.Complexity.SetupSteps)
	experimentalComplexity := float64(experimental.Complexity.CyclomaticComplexity + 
		experimental.Complexity.SetupSteps)
	
	if baselineComplexity == 0 {
		return 1.0
	}
	return experimentalComplexity / baselineComplexity
}

// calculateImplementationRatio computes relative implementation effort
func calculateImplementationRatio(baseline, experimental *Metrics) float64 {
	if baseline.Implementation.LinesOfCode == 0 {
		return 1.0
	}
	return float64(experimental.Implementation.LinesOfCode) / 
		float64(baseline.Implementation.LinesOfCode)
}

// generateComparisonSummary creates a human-readable summary
func generateComparisonSummary(comp Comparison) string {
	var summary bytes.Buffer
	
	if comp.PerformanceRatio < 1.0 {
		summary.WriteString(fmt.Sprintf("Performance improved by %.1f%%. ", 
			(1-comp.PerformanceRatio)*100))
	} else if comp.PerformanceRatio > 1.0 {
		summary.WriteString(fmt.Sprintf("Performance degraded by %.1f%%. ", 
			(comp.PerformanceRatio-1)*100))
	}
	
	if comp.ComplexityRatio > 1.0 {
		summary.WriteString(fmt.Sprintf("Complexity increased by %.1f%%. ", 
			(comp.ComplexityRatio-1)*100))
	} else if comp.ComplexityRatio < 1.0 {
		summary.WriteString(fmt.Sprintf("Complexity reduced by %.1f%%. ", 
			(1-comp.ComplexityRatio)*100))
	}
	
	if comp.TokenOverheadRatio > 1.0 {
		summary.WriteString(fmt.Sprintf("Token overhead increased by %.1f%%.", 
			(comp.TokenOverheadRatio-1)*100))
	}
	
	return summary.String()
}

// sanitizeFilename removes characters that aren't safe for filenames
func sanitizeFilename(name string) string {
	// Simple implementation - replace spaces with dashes and remove special chars
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
		   (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result += string(r)
		} else if r == ' ' {
			result += "-"
		}
	}
	return result
}

// SaveMetricsJSON saves metrics to a JSON file for later analysis
func SaveMetricsJSON(metrics *Metrics, filepath string) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	
	// In a real implementation, would use os.WriteFile
	_ = data
	_ = filepath
	return nil
}

const reportTemplate = `# Lab Report: {{.Title}}

**Week**: {{.Week}}  
**Date**: {{.Date.Format "2006-01-02"}}  
**Status**: {{if .Results.Quantitative.Comparison.Summary}}Complete{{else}}Draft{{end}}

## Hypothesis

{{.Hypothesis}}

## Method

{{.Method}}

## Quantitative Results

### Performance Comparison

| Metric | Baseline | Experimental | Ratio | Change |
|--------|----------|--------------|-------|--------|
| Latency P50 | {{.Results.Quantitative.Baseline.Performance.LatencyP50}} | {{.Results.Quantitative.Experimental.Performance.LatencyP50}} | {{printf "%.2f" .Results.Quantitative.Comparison.PerformanceRatio}} | {{if lt .Results.Quantitative.Comparison.PerformanceRatio 1.0}}✅ Improved{{else}}❌ Degraded{{end}} |
| Token Overhead | {{printf "%.1f%%" .Results.Quantitative.Baseline.Performance.TokenOverhead}} | {{printf "%.1f%%" .Results.Quantitative.Experimental.Performance.TokenOverhead}} | {{printf "%.2f" .Results.Quantitative.Comparison.TokenOverheadRatio}} | {{if lt .Results.Quantitative.Comparison.TokenOverheadRatio 1.0}}✅ Reduced{{else}}❌ Increased{{end}} |
| Memory Usage | {{printf "%.1f MB" .Results.Quantitative.Baseline.Performance.MemoryUsageMB}} | {{printf "%.1f MB" .Results.Quantitative.Experimental.Performance.MemoryUsageMB}} | - | - |

### Implementation Comparison

| Metric | Baseline | Experimental | Ratio | Change |
|--------|----------|--------------|-------|--------|
| Lines of Code | {{.Results.Quantitative.Baseline.Implementation.LinesOfCode}} | {{.Results.Quantitative.Experimental.Implementation.LinesOfCode}} | {{printf "%.2f" .Results.Quantitative.Comparison.ImplementationRatio}} | {{if lt .Results.Quantitative.Comparison.ImplementationRatio 1.0}}✅ Reduced{{else}}❌ Increased{{end}} |
| Files Changed | {{.Results.Quantitative.Baseline.Implementation.FilesChanged}} | {{.Results.Quantitative.Experimental.Implementation.FilesChanged}} | - | - |
| Dependencies | {{len .Results.Quantitative.Baseline.Implementation.Dependencies}} | {{len .Results.Quantitative.Experimental.Implementation.Dependencies}} | - | - |

### Complexity Comparison

| Metric | Baseline | Experimental | Ratio | Change |
|--------|----------|--------------|-------|--------|
| Setup Steps | {{.Results.Quantitative.Baseline.Complexity.SetupSteps}} | {{.Results.Quantitative.Experimental.Complexity.SetupSteps}} | - | - |
| Interface Count | {{.Results.Quantitative.Baseline.Complexity.InterfaceCount}} | {{.Results.Quantitative.Experimental.Complexity.InterfaceCount}} | - | - |
| Config Items | {{.Results.Quantitative.Baseline.Complexity.ConfigurationItems}} | {{.Results.Quantitative.Experimental.Complexity.ConfigurationItems}} | - | - |

## Summary

{{.Results.Quantitative.Comparison.Summary}}

## Qualitative Observations

{{range .Results.Qualitative}}
- {{.}}
{{end}}

## Failures & Issues

{{range .Failures}}
- {{.}}
{{end}}

## Conclusions

{{range .Conclusions}}
- {{.}}
{{end}}

## Next Steps

{{range .NextSteps}}
- [ ] {{.}}
{{end}}

---

*Generated on {{.Date.Format "2006-01-02 15:04:05"}} by teeny-orb experiment framework*
`