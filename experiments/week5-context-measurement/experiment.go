package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	contextpkg "github.com/rcliao/teeny-orb/internal/context"
)

// Week5Experiment implements the context measurement experiment for Phase 2
type Week5Experiment struct {
	analyzer    contextpkg.ContextAnalyzer
	optimizer   contextpkg.ContextOptimizer
	tokenCounter contextpkg.TokenCounter
	results     *ExperimentResults
}

// ExperimentResults tracks the results of Week 5 context measurement experiments
type ExperimentResults struct {
	ExperimentName   string                    `json:"experiment_name"`
	StartTime        time.Time                 `json:"start_time"`
	EndTime          time.Time                 `json:"end_time"`
	Duration         time.Duration             `json:"duration"`
	ProjectsAnalyzed int                       `json:"projects_analyzed"`
	Measurements     []ProjectMeasurement      `json:"measurements"`
	Summary          *MeasurementSummary       `json:"summary"`
	Conclusions      []string                  `json:"conclusions"`
}

// ProjectMeasurement represents measurements for a single project
type ProjectMeasurement struct {
	ProjectPath      string                    `json:"project_path"`
	ProjectName      string                    `json:"project_name"`
	TotalFiles       int                       `json:"total_files"`
	TotalTokens      int                       `json:"total_tokens"`
	Languages        map[string]int            `json:"languages"`
	FileTypeBreakdown map[string]FileTypeStats `json:"file_type_breakdown"`
	DependencyMetrics *DependencyMetrics       `json:"dependency_metrics"`
	AnalysisTime     time.Duration             `json:"analysis_time"`
	TaskScenarios    []TaskScenario            `json:"task_scenarios"`
}

// FileTypeStats provides statistics for a specific file type
type FileTypeStats struct {
	Count       int     `json:"count"`
	TotalTokens int     `json:"total_tokens"`
	AvgTokens   float64 `json:"avg_tokens"`
	Percentage  float64 `json:"percentage"`
}

// DependencyMetrics provides dependency graph analysis
type DependencyMetrics struct {
	TotalNodes       int     `json:"total_nodes"`
	TotalEdges       int     `json:"total_edges"`
	AvgDependencies  float64 `json:"avg_dependencies"`
	MaxDependencies  int     `json:"max_dependencies"`
	ConnectedComponents int  `json:"connected_components"`
}

// TaskScenario represents different coding task scenarios for context testing
type TaskScenario struct {
	TaskType         contextpkg.TaskType `json:"task_type"`
	Description      string              `json:"description"`
	TokenBudget      int                 `json:"token_budget"`
	SelectedFiles    int                 `json:"selected_files"`
	SelectedTokens   int                 `json:"selected_tokens"`
	ReductionRatio   float64             `json:"reduction_ratio"`
	SelectionTime    time.Duration       `json:"selection_time"`
	RelevanceScore   float64             `json:"relevance_score"`
}

// MeasurementSummary provides aggregate insights across all projects
type MeasurementSummary struct {
	TotalProjects        int                 `json:"total_projects"`
	TotalFiles           int                 `json:"total_files"`
	TotalTokens          int                 `json:"total_tokens"`
	AvgTokensPerProject  float64             `json:"avg_tokens_per_project"`
	AvgTokensPerFile     float64             `json:"avg_tokens_per_file"`
	LanguageDistribution map[string]float64  `json:"language_distribution"`
	FileTypeDistribution map[string]float64  `json:"file_type_distribution"`
	ContextOptimization  *OptimizationStats  `json:"context_optimization"`
}

// OptimizationStats tracks context optimization effectiveness
type OptimizationStats struct {
	AvgReductionRatio    float64 `json:"avg_reduction_ratio"`
	AvgSelectionTime     float64 `json:"avg_selection_time_ms"`
	AvgRelevanceScore    float64 `json:"avg_relevance_score"`
	BestCase            float64 `json:"best_case_reduction"`
	WorstCase           float64 `json:"worst_case_reduction"`
	TokenBudgetEfficiency float64 `json:"token_budget_efficiency"`
}

// NewWeek5Experiment creates a new Week 5 context measurement experiment
func NewWeek5Experiment() *Week5Experiment {
	tokenCounter := contextpkg.NewSimpleTokenCounter()
	analyzer := contextpkg.NewDefaultAnalyzer(tokenCounter, nil)
	
	// For now, we'll create a basic optimizer
	// In a real implementation, you'd inject proper dependencies
	var optimizer contextpkg.ContextOptimizer
	
	return &Week5Experiment{
		analyzer:     analyzer,
		optimizer:    optimizer,
		tokenCounter: tokenCounter,
		results: &ExperimentResults{
			ExperimentName: "Week 5: Context Measurement Foundation",
			StartTime:      time.Now(),
			Measurements:   []ProjectMeasurement{},
		},
	}
}

// RunExperiment executes the complete Week 5 experiment
func (e *Week5Experiment) RunExperiment(ctx context.Context, projectPaths []string) error {
	log.Println("Starting Week 5 Context Measurement Experiment")
	
	for i, projectPath := range projectPaths {
		log.Printf("Analyzing project %d/%d: %s", i+1, len(projectPaths), projectPath)
		
		measurement, err := e.measureProject(ctx, projectPath)
		if err != nil {
			log.Printf("Error analyzing project %s: %v", projectPath, err)
			continue
		}
		
		e.results.Measurements = append(e.results.Measurements, *measurement)
		e.results.ProjectsAnalyzed++
	}
	
	// Generate summary and conclusions
	e.generateSummary()
	e.generateConclusions()
	
	e.results.EndTime = time.Now()
	e.results.Duration = e.results.EndTime.Sub(e.results.StartTime)
	
	log.Printf("Experiment completed. Analyzed %d projects in %v", 
		e.results.ProjectsAnalyzed, e.results.Duration)
	
	return nil
}

// measureProject performs comprehensive measurement of a single project
func (e *Week5Experiment) measureProject(ctx context.Context, projectPath string) (*ProjectMeasurement, error) {
	startTime := time.Now()
	
	// Analyze project context
	projectContext, err := e.analyzer.AnalyzeProject(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}
	
	measurement := &ProjectMeasurement{
		ProjectPath:      projectPath,
		ProjectName:      filepath.Base(projectPath),
		TotalFiles:       projectContext.TotalFiles,
		TotalTokens:      projectContext.TotalTokens,
		Languages:        projectContext.Languages,
		FileTypeBreakdown: e.calculateFileTypeBreakdown(projectContext),
		DependencyMetrics: e.calculateDependencyMetrics(projectContext),
		AnalysisTime:     time.Since(startTime),
		TaskScenarios:    []TaskScenario{},
	}
	
	// Test different task scenarios to measure context optimization
	scenarios := e.generateTaskScenarios()
	for _, scenario := range scenarios {
		if e.optimizer != nil {
			taskResult, err := e.measureTaskScenario(ctx, projectContext, scenario)
			if err == nil {
				measurement.TaskScenarios = append(measurement.TaskScenarios, *taskResult)
			}
		}
	}
	
	return measurement, nil
}

// calculateFileTypeBreakdown analyzes file type distribution
func (e *Week5Experiment) calculateFileTypeBreakdown(projectCtx *contextpkg.ProjectContext) map[string]FileTypeStats {
	breakdown := make(map[string]FileTypeStats)
	
	for _, file := range projectCtx.Files {
		stats, exists := breakdown[file.FileType]
		if !exists {
			stats = FileTypeStats{}
		}
		
		stats.Count++
		stats.TotalTokens += file.TokenCount
		breakdown[file.FileType] = stats
	}
	
	// Calculate averages and percentages
	for fileType, stats := range breakdown {
		if stats.Count > 0 {
			stats.AvgTokens = float64(stats.TotalTokens) / float64(stats.Count)
		}
		if projectCtx.TotalTokens > 0 {
			stats.Percentage = float64(stats.TotalTokens) / float64(projectCtx.TotalTokens) * 100
		}
		breakdown[fileType] = stats
	}
	
	return breakdown
}

// calculateDependencyMetrics analyzes dependency graph metrics
func (e *Week5Experiment) calculateDependencyMetrics(projectCtx *contextpkg.ProjectContext) *DependencyMetrics {
	if projectCtx.DependencyGraph == nil {
		return &DependencyMetrics{}
	}
	
	graph := projectCtx.DependencyGraph
	totalNodes := len(graph.Nodes)
	totalEdges := len(graph.Edges)
	
	// Calculate average dependencies per node
	avgDependencies := 0.0
	maxDependencies := 0
	
	if totalNodes > 0 {
		totalDeps := 0
		for _, node := range graph.Nodes {
			deps := len(node.Dependencies)
			totalDeps += deps
			if deps > maxDependencies {
				maxDependencies = deps
			}
		}
		avgDependencies = float64(totalDeps) / float64(totalNodes)
	}
	
	return &DependencyMetrics{
		TotalNodes:          totalNodes,
		TotalEdges:          totalEdges,
		AvgDependencies:     avgDependencies,
		MaxDependencies:     maxDependencies,
		ConnectedComponents: 1, // Simplified for now
	}
}

// generateTaskScenarios creates different task scenarios for testing
func (e *Week5Experiment) generateTaskScenarios() []contextpkg.Task {
	return []contextpkg.Task{
		{
			Type:        contextpkg.TaskTypeFeature,
			Description: "Add new REST API endpoint",
			Priority:    contextpkg.PriorityMedium,
			Scope:       contextpkg.ScopeModule,
		},
		{
			Type:        contextpkg.TaskTypeDebug,
			Description: "Fix memory leak in session management",
			Priority:    contextpkg.PriorityHigh,
			Scope:       contextpkg.ScopeFile,
		},
		{
			Type:        contextpkg.TaskTypeRefactor,
			Description: "Refactor error handling patterns",
			Priority:    contextpkg.PriorityMedium,
			Scope:       contextpkg.ScopeProject,
		},
		{
			Type:        contextpkg.TaskTypeTest,
			Description: "Add integration tests for MCP server",
			Priority:    contextpkg.PriorityLow,
			Scope:       contextpkg.ScopeModule,
		},
	}
}

// measureTaskScenario measures context optimization for a specific task
func (e *Week5Experiment) measureTaskScenario(ctx context.Context, projectCtx *contextpkg.ProjectContext, task contextpkg.Task) (*TaskScenario, error) {
	// Test with medium budget (8000 tokens)
	// In future experiments, we can test with different budgets: [2000, 4000, 8000, 16000]
	tokenBudget := 8000
	
	startTime := time.Now()
	selectedContext, err := e.optimizer.OptimizeForTokenBudget(ctx, projectCtx, tokenBudget, &task)
	if err != nil {
		return nil, err
	}
	selectionTime := time.Since(startTime)
	
	reductionRatio := 1.0
	if projectCtx.TotalTokens > 0 {
		reductionRatio = 1.0 - (float64(selectedContext.TotalTokens) / float64(projectCtx.TotalTokens))
	}
	
	return &TaskScenario{
		TaskType:       task.Type,
		Description:    task.Description,
		TokenBudget:    tokenBudget,
		SelectedFiles:  selectedContext.TotalFiles,
		SelectedTokens: selectedContext.TotalTokens,
		ReductionRatio: reductionRatio,
		SelectionTime:  selectionTime,
		RelevanceScore: selectedContext.SelectionScore,
	}, nil
}

// generateSummary creates aggregate summary statistics
func (e *Week5Experiment) generateSummary() {
	summary := &MeasurementSummary{
		TotalProjects:        e.results.ProjectsAnalyzed,
		LanguageDistribution: make(map[string]float64),
		FileTypeDistribution: make(map[string]float64),
	}
	
	totalFiles := 0
	totalTokens := 0
	langCounts := make(map[string]int)
	typeCounts := make(map[string]int)
	
	// Aggregate statistics
	for _, measurement := range e.results.Measurements {
		totalFiles += measurement.TotalFiles
		totalTokens += measurement.TotalTokens
		
		// Aggregate language counts
		for lang, count := range measurement.Languages {
			langCounts[lang] += count
		}
		
		// Aggregate file type counts
		for fileType, stats := range measurement.FileTypeBreakdown {
			typeCounts[fileType] += stats.Count
		}
	}
	
	summary.TotalFiles = totalFiles
	summary.TotalTokens = totalTokens
	
	if summary.TotalProjects > 0 {
		summary.AvgTokensPerProject = float64(totalTokens) / float64(summary.TotalProjects)
	}
	if totalFiles > 0 {
		summary.AvgTokensPerFile = float64(totalTokens) / float64(totalFiles)
	}
	
	// Calculate distributions
	for lang, count := range langCounts {
		if totalFiles > 0 {
			summary.LanguageDistribution[lang] = float64(count) / float64(totalFiles) * 100
		}
	}
	
	for fileType, count := range typeCounts {
		if totalFiles > 0 {
			summary.FileTypeDistribution[fileType] = float64(count) / float64(totalFiles) * 100
		}
	}
	
	// Calculate optimization statistics
	summary.ContextOptimization = e.calculateOptimizationStats()
	
	e.results.Summary = summary
}

// calculateOptimizationStats aggregates context optimization performance
func (e *Week5Experiment) calculateOptimizationStats() *OptimizationStats {
	stats := &OptimizationStats{}
	
	totalScenarios := 0
	totalReduction := 0.0
	totalSelectionTime := time.Duration(0)
	totalRelevanceScore := 0.0
	bestReduction := 0.0
	worstReduction := 1.0
	
	for _, measurement := range e.results.Measurements {
		for _, scenario := range measurement.TaskScenarios {
			totalScenarios++
			totalReduction += scenario.ReductionRatio
			totalSelectionTime += scenario.SelectionTime
			totalRelevanceScore += scenario.RelevanceScore
			
			if scenario.ReductionRatio > bestReduction {
				bestReduction = scenario.ReductionRatio
			}
			if scenario.ReductionRatio < worstReduction {
				worstReduction = scenario.ReductionRatio
			}
		}
	}
	
	if totalScenarios > 0 {
		stats.AvgReductionRatio = totalReduction / float64(totalScenarios)
		stats.AvgSelectionTime = float64(totalSelectionTime.Nanoseconds()) / float64(totalScenarios) / 1e6 // Convert to ms
		stats.AvgRelevanceScore = totalRelevanceScore / float64(totalScenarios)
		stats.BestCase = bestReduction
		stats.WorstCase = worstReduction
		stats.TokenBudgetEfficiency = stats.AvgReductionRatio * stats.AvgRelevanceScore
	}
	
	return stats
}

// generateConclusions creates experiment conclusions based on results
func (e *Week5Experiment) generateConclusions() {
	conclusions := []string{}
	
	if e.results.Summary == nil {
		e.results.Conclusions = conclusions
		return
	}
	
	summary := e.results.Summary
	
	// Token distribution insights
	if summary.AvgTokensPerFile > 500 {
		conclusions = append(conclusions, 
			fmt.Sprintf("Large average file size (%.0f tokens) suggests significant context optimization potential", 
				summary.AvgTokensPerFile))
	}
	
	// Optimization effectiveness
	if summary.ContextOptimization != nil {
		opt := summary.ContextOptimization
		if opt.AvgReductionRatio > 0.7 {
			conclusions = append(conclusions, 
				fmt.Sprintf("Context optimization achieved %.1f%% average token reduction - exceeding 70%% target", 
					opt.AvgReductionRatio*100))
		}
		
		if opt.AvgSelectionTime < 100 {
			conclusions = append(conclusions, 
				fmt.Sprintf("Fast context selection (%.1f ms average) meets <100ms performance target", 
					opt.AvgSelectionTime))
		}
	}
	
	// Language-specific insights
	for lang, percentage := range summary.LanguageDistribution {
		if percentage > 50 {
			conclusions = append(conclusions, 
				fmt.Sprintf("%s dominates codebase (%.1f%%) - language-specific optimization recommended", 
					lang, percentage))
		}
	}
	
	e.results.Conclusions = conclusions
}

// SaveResults saves experiment results to a JSON file
func (e *Week5Experiment) SaveResults(outputPath string) error {
	data, err := json.MarshalIndent(e.results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}
	
	return os.WriteFile(outputPath, data, 0644)
}

// PrintSummary prints a summary of experiment results
func (e *Week5Experiment) PrintSummary() {
	fmt.Println("\n=== Week 5 Context Measurement Experiment Results ===")
	fmt.Printf("Duration: %v\n", e.results.Duration)
	fmt.Printf("Projects Analyzed: %d\n", e.results.ProjectsAnalyzed)
	
	if e.results.Summary != nil {
		s := e.results.Summary
		fmt.Printf("Total Files: %d\n", s.TotalFiles)
		fmt.Printf("Total Tokens: %d\n", s.TotalTokens)
		fmt.Printf("Average Tokens per Project: %.0f\n", s.AvgTokensPerProject)
		fmt.Printf("Average Tokens per File: %.0f\n", s.AvgTokensPerFile)
		
		if s.ContextOptimization != nil {
			opt := s.ContextOptimization
			fmt.Printf("\nContext Optimization Results:\n")
			fmt.Printf("  Average Token Reduction: %.1f%%\n", opt.AvgReductionRatio*100)
			fmt.Printf("  Average Selection Time: %.1f ms\n", opt.AvgSelectionTime)
			fmt.Printf("  Average Relevance Score: %.3f\n", opt.AvgRelevanceScore)
			fmt.Printf("  Best Case Reduction: %.1f%%\n", opt.BestCase*100)
			fmt.Printf("  Worst Case Reduction: %.1f%%\n", opt.WorstCase*100)
		}
	}
	
	if len(e.results.Conclusions) > 0 {
		fmt.Println("\nKey Conclusions:")
		for i, conclusion := range e.results.Conclusions {
			fmt.Printf("  %d. %s\n", i+1, conclusion)
		}
	}
}

// Main experiment execution
func main() {
	ctx := context.Background()
	
	// Test projects (you can add more project paths here)
	projectPaths := []string{
		"../../", // Test on teeny-orb itself
	}
	
	experiment := NewWeek5Experiment()
	
	if err := experiment.RunExperiment(ctx, projectPaths); err != nil {
		log.Fatalf("Experiment failed: %v", err)
	}
	
	// Print summary
	experiment.PrintSummary()
	
	// Save results
	outputPath := "week5_context_measurement_results.json"
	if err := experiment.SaveResults(outputPath); err != nil {
		log.Printf("Failed to save results: %v", err)
	} else {
		fmt.Printf("\nResults saved to: %s\n", outputPath)
	}
}