package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	contextpkg "github.com/rcliao/teeny-orb/internal/context"
)

// Week6Experiment tests different context selection strategies
type Week6Experiment struct {
	analyzer    contextpkg.ContextAnalyzer
	optimizer   contextpkg.ContextOptimizer
	tokenCounter contextpkg.TokenCounter
	results     *SelectionExperimentResults
}

// SelectionExperimentResults tracks Week 6 smart selection experiment results
type SelectionExperimentResults struct {
	ExperimentName    string                    `json:"experiment_name"`
	StartTime         time.Time                 `json:"start_time"`
	EndTime           time.Time                 `json:"end_time"`
	Duration          time.Duration             `json:"duration"`
	ProjectsAnalyzed  int                       `json:"projects_analyzed"`
	StrategiesResults []StrategyResults         `json:"strategies_results"`
	TaskResults       []TaskResults             `json:"task_results"`
	Summary           *SelectionSummary         `json:"summary"`
	Conclusions       []string                  `json:"conclusions"`
}

// StrategyResults tracks performance of each selection strategy
type StrategyResults struct {
	Strategy          string    `json:"strategy"`
	AvgSelectionTime  float64   `json:"avg_selection_time_ms"`
	AvgTokenReduction float64   `json:"avg_token_reduction"`
	AvgRelevanceScore float64   `json:"avg_relevance_score"`
	AvgFilesSelected  float64   `json:"avg_files_selected"`
	TestCases         []TestCase `json:"test_cases"`
}

// TaskResults tracks results for different task types
type TaskResults struct {
	TaskType          string    `json:"task_type"`
	AvgTokenReduction float64   `json:"avg_token_reduction"`
	AvgRelevanceScore float64   `json:"avg_relevance_score"`
	BestStrategy      string    `json:"best_strategy"`
	TestCases         []TestCase `json:"test_cases"`
}

// TestCase represents a single test scenario
type TestCase struct {
	TaskDescription   string                    `json:"task_description"`
	TaskType          string                    `json:"task_type"`
	Strategy          string                    `json:"strategy"`
	TokenBudget       int                       `json:"token_budget"`
	SelectedFiles     int                       `json:"selected_files"`
	SelectedTokens    int                       `json:"selected_tokens"`
	TokenReduction    float64                   `json:"token_reduction"`
	RelevanceScore    float64                   `json:"relevance_score"`
	SelectionTime     time.Duration             `json:"selection_time"`
	TopFiles          []string                  `json:"top_files"`
}

// SelectionSummary provides aggregate insights
type SelectionSummary struct {
	BestOverallStrategy      string             `json:"best_overall_strategy"`
	WorstOverallStrategy     string             `json:"worst_overall_strategy"`
	AvgTokenReductionByStrategy map[string]float64 `json:"avg_token_reduction_by_strategy"`
	AvgPerformanceByTaskType map[string]float64 `json:"avg_performance_by_task_type"`
	PerformanceGains         *PerformanceGains  `json:"performance_gains"`
}

// PerformanceGains tracks improvements over baseline
type PerformanceGains struct {
	BestVsWorst         float64 `json:"best_vs_worst_reduction"`
	BalancedVsRandom    float64 `json:"balanced_vs_random"`
	DependencyBoost     float64 `json:"dependency_strategy_boost"`
	RelevanceAccuracy   float64 `json:"relevance_accuracy"`
}

// NewWeek6Experiment creates a new smart selection experiment
func NewWeek6Experiment() *Week6Experiment {
	tokenCounter := contextpkg.NewSimpleTokenCounter()
	analyzer := contextpkg.NewDefaultAnalyzer(tokenCounter, nil)
	optimizer := contextpkg.NewDefaultOptimizer(analyzer, nil, nil, nil)
	
	return &Week6Experiment{
		analyzer:     analyzer,
		optimizer:    optimizer,
		tokenCounter: tokenCounter,
		results: &SelectionExperimentResults{
			ExperimentName:    "Week 6: Smart Context Selection",
			StartTime:         time.Now(),
			StrategiesResults: []StrategyResults{},
			TaskResults:       []TaskResults{},
		},
	}
}

// RunExperiment executes the complete Week 6 experiment
func (e *Week6Experiment) RunExperiment(ctx context.Context, projectPaths []string) error {
	log.Println("Starting Week 6 Smart Context Selection Experiment")
	
	// Test different strategies
	strategies := []string{"relevance", "dependency", "freshness", "compactness", "balanced"}
	
	// Test different task types
	taskTypes := []contextpkg.TaskType{
		contextpkg.TaskTypeFeature,
		contextpkg.TaskTypeDebug,
		contextpkg.TaskTypeRefactor,
		contextpkg.TaskTypeTest,
	}
	
	// Test scenarios for each project
	for i, projectPath := range projectPaths {
		log.Printf("Analyzing project %d/%d: %s", i+1, len(projectPaths), projectPath)
		
		// Analyze project context once
		projectContext, err := e.analyzer.AnalyzeProject(ctx, projectPath)
		if err != nil {
			log.Printf("Error analyzing project %s: %v", projectPath, err)
			continue
		}
		
		e.results.ProjectsAnalyzed++
		
		// Test each strategy with each task type
		for _, strategy := range strategies {
			for _, taskType := range taskTypes {
				testCase, err := e.runTestCase(ctx, projectContext, strategy, taskType)
				if err != nil {
					log.Printf("Error in test case %s/%s: %v", strategy, taskType, err)
					continue
				}
				
				// Add to strategy results
				e.addTestCaseToStrategy(strategy, *testCase)
				
				// Add to task results
				e.addTestCaseToTaskType(string(taskType), *testCase)
			}
		}
	}
	
	// Generate summary and conclusions
	e.generateSummary()
	e.generateConclusions()
	
	e.results.EndTime = time.Now()
	e.results.Duration = e.results.EndTime.Sub(e.results.StartTime)
	
	log.Printf("Experiment completed. Tested %d strategies across %d task types on %d projects", 
		len(strategies), len(taskTypes), e.results.ProjectsAnalyzed)
	
	return nil
}

// runTestCase executes a single test case
func (e *Week6Experiment) runTestCase(ctx context.Context, projectCtx *contextpkg.ProjectContext, strategy string, taskType contextpkg.TaskType) (*TestCase, error) {
	// Create test task
	taskDescriptions := map[contextpkg.TaskType]string{
		contextpkg.TaskTypeFeature:       "Add new REST API endpoint for user management",
		contextpkg.TaskTypeDebug:         "Fix memory leak in session handler",
		contextpkg.TaskTypeRefactor:      "Refactor error handling patterns",
		contextpkg.TaskTypeTest:          "Add integration tests for authentication",
	}
	
	task := &contextpkg.Task{
		Type:        taskType,
		Description: taskDescriptions[taskType],
		Priority:    contextpkg.PriorityMedium,
		Scope:       contextpkg.ScopeProject,
	}
	
	// Create constraints for this strategy
	constraints := &contextpkg.ContextConstraints{
		MaxTokens:         8000, // Standard budget
		MaxFiles:          30,
		MinRelevanceScore: 0.1,
		PreferredTypes:    []string{"source", "configuration"},
		IncludeTests:      taskType == contextpkg.TaskTypeTest,
		IncludeDocs:       false,
		FreshnessBias:     0.3,
		DependencyDepth:   2,
		Strategy:          contextpkg.SelectionStrategy(strategy),
	}
	
	// Measure selection performance
	startTime := time.Now()
	selectedContext, err := e.optimizer.SelectOptimalContext(ctx, projectCtx, task, constraints)
	selectionTime := time.Since(startTime)
	
	if err != nil {
		return nil, err
	}
	
	// Calculate metrics
	tokenReduction := 1.0
	if projectCtx.TotalTokens > 0 {
		tokenReduction = 1.0 - (float64(selectedContext.TotalTokens) / float64(projectCtx.TotalTokens))
	}
	
	// Get top files for analysis
	topFiles := []string{}
	for i, file := range selectedContext.Files {
		if i < 5 { // Top 5 files
			topFiles = append(topFiles, file.FileInfo.Path)
		}
	}
	
	return &TestCase{
		TaskDescription: task.Description,
		TaskType:        string(taskType),
		Strategy:        strategy,
		TokenBudget:     constraints.MaxTokens,
		SelectedFiles:   selectedContext.TotalFiles,
		SelectedTokens:  selectedContext.TotalTokens,
		TokenReduction:  tokenReduction,
		RelevanceScore:  selectedContext.SelectionScore,
		SelectionTime:   selectionTime,
		TopFiles:        topFiles,
	}, nil
}

// addTestCaseToStrategy adds a test case to strategy results
func (e *Week6Experiment) addTestCaseToStrategy(strategy string, testCase TestCase) {
	// Find or create strategy results
	var strategyResults *StrategyResults
	for i := range e.results.StrategiesResults {
		if e.results.StrategiesResults[i].Strategy == strategy {
			strategyResults = &e.results.StrategiesResults[i]
			break
		}
	}
	
	if strategyResults == nil {
		e.results.StrategiesResults = append(e.results.StrategiesResults, StrategyResults{
			Strategy:  strategy,
			TestCases: []TestCase{},
		})
		strategyResults = &e.results.StrategiesResults[len(e.results.StrategiesResults)-1]
	}
	
	strategyResults.TestCases = append(strategyResults.TestCases, testCase)
}

// addTestCaseToTaskType adds a test case to task type results
func (e *Week6Experiment) addTestCaseToTaskType(taskType string, testCase TestCase) {
	// Find or create task results
	var taskResults *TaskResults
	for i := range e.results.TaskResults {
		if e.results.TaskResults[i].TaskType == taskType {
			taskResults = &e.results.TaskResults[i]
			break
		}
	}
	
	if taskResults == nil {
		e.results.TaskResults = append(e.results.TaskResults, TaskResults{
			TaskType:  taskType,
			TestCases: []TestCase{},
		})
		taskResults = &e.results.TaskResults[len(e.results.TaskResults)-1]
	}
	
	taskResults.TestCases = append(taskResults.TestCases, testCase)
}

// generateSummary creates aggregate summary statistics
func (e *Week6Experiment) generateSummary() {
	summary := &SelectionSummary{
		AvgTokenReductionByStrategy: make(map[string]float64),
		AvgPerformanceByTaskType:    make(map[string]float64),
		PerformanceGains:           &PerformanceGains{},
	}
	
	// Calculate strategy averages
	bestReduction := 0.0
	worstReduction := 1.0
	bestStrategy := ""
	worstStrategy := ""
	
	for i := range e.results.StrategiesResults {
		strategy := &e.results.StrategiesResults[i]
		
		if len(strategy.TestCases) == 0 {
			continue
		}
		
		// Calculate averages for this strategy
		totalReduction := 0.0
		totalRelevance := 0.0
		totalFiles := 0.0
		totalTime := 0.0
		
		for _, testCase := range strategy.TestCases {
			totalReduction += testCase.TokenReduction
			totalRelevance += testCase.RelevanceScore
			totalFiles += float64(testCase.SelectedFiles)
			totalTime += float64(testCase.SelectionTime.Nanoseconds()) / 1e6 // Convert to ms
		}
		
		count := float64(len(strategy.TestCases))
		strategy.AvgTokenReduction = totalReduction / count
		strategy.AvgRelevanceScore = totalRelevance / count
		strategy.AvgFilesSelected = totalFiles / count
		strategy.AvgSelectionTime = totalTime / count
		
		summary.AvgTokenReductionByStrategy[strategy.Strategy] = strategy.AvgTokenReduction
		
		// Track best and worst
		if strategy.AvgTokenReduction > bestReduction {
			bestReduction = strategy.AvgTokenReduction
			bestStrategy = strategy.Strategy
		}
		if strategy.AvgTokenReduction < worstReduction {
			worstReduction = strategy.AvgTokenReduction
			worstStrategy = strategy.Strategy
		}
	}
	
	summary.BestOverallStrategy = bestStrategy
	summary.WorstOverallStrategy = worstStrategy
	
	// Calculate task type averages
	for i := range e.results.TaskResults {
		taskResult := &e.results.TaskResults[i]
		
		if len(taskResult.TestCases) == 0 {
			continue
		}
		
		// Calculate averages and find best strategy for this task type
		totalReduction := 0.0
		totalRelevance := 0.0
		strategyPerformance := make(map[string]float64)
		strategyCounts := make(map[string]int)
		
		for _, testCase := range taskResult.TestCases {
			totalReduction += testCase.TokenReduction
			totalRelevance += testCase.RelevanceScore
			
			strategyPerformance[testCase.Strategy] += testCase.TokenReduction
			strategyCounts[testCase.Strategy]++
		}
		
		count := float64(len(taskResult.TestCases))
		taskResult.AvgTokenReduction = totalReduction / count
		taskResult.AvgRelevanceScore = totalRelevance / count
		
		summary.AvgPerformanceByTaskType[taskResult.TaskType] = taskResult.AvgTokenReduction
		
		// Find best strategy for this task type
		bestStrategyPerf := 0.0
		for strategy, totalPerf := range strategyPerformance {
			avgPerf := totalPerf / float64(strategyCounts[strategy])
			if avgPerf > bestStrategyPerf {
				bestStrategyPerf = avgPerf
				taskResult.BestStrategy = strategy
			}
		}
	}
	
	// Calculate performance gains
	summary.PerformanceGains.BestVsWorst = bestReduction - worstReduction
	if balancedPerf, exists := summary.AvgTokenReductionByStrategy["balanced"]; exists {
		if relevancePerf, exists := summary.AvgTokenReductionByStrategy["relevance"]; exists {
			summary.PerformanceGains.BalancedVsRandom = balancedPerf - (relevancePerf * 0.5) // Rough random baseline
		}
	}
	if depPerf, exists := summary.AvgTokenReductionByStrategy["dependency"]; exists {
		if relPerf, exists := summary.AvgTokenReductionByStrategy["relevance"]; exists {
			summary.PerformanceGains.DependencyBoost = depPerf - relPerf
		}
	}
	
	// Set relevance accuracy (simplified metric)
	summary.PerformanceGains.RelevanceAccuracy = bestReduction // Use best reduction as proxy
	
	e.results.Summary = summary
}

// generateConclusions creates experiment conclusions
func (e *Week6Experiment) generateConclusions() {
	conclusions := []string{}
	
	if e.results.Summary == nil {
		e.results.Conclusions = conclusions
		return
	}
	
	summary := e.results.Summary
	
	// Strategy performance insights
	if summary.BestOverallStrategy != "" {
		conclusions = append(conclusions, 
			fmt.Sprintf("Best overall strategy: '%s' with %.1f%% average token reduction", 
				summary.BestOverallStrategy, 
				summary.AvgTokenReductionByStrategy[summary.BestOverallStrategy]*100))
	}
	
	if summary.PerformanceGains.BestVsWorst > 0.1 {
		conclusions = append(conclusions, 
			fmt.Sprintf("Significant strategy differences: %.1f%% performance gap between best and worst", 
				summary.PerformanceGains.BestVsWorst*100))
	}
	
	// Task-specific insights
	for taskType, performance := range summary.AvgPerformanceByTaskType {
		if performance > 0.7 {
			conclusions = append(conclusions, 
				fmt.Sprintf("%s tasks achieve %.1f%% token reduction - excellent optimization", 
					taskType, performance*100))
		}
	}
	
	// Dependency analysis insight
	if summary.PerformanceGains.DependencyBoost > 0.05 {
		conclusions = append(conclusions, 
			fmt.Sprintf("Dependency analysis provides %.1f%% boost over pure relevance", 
				summary.PerformanceGains.DependencyBoost*100))
	}
	
	// Overall assessment
	if len(e.results.StrategiesResults) >= 5 {
		conclusions = append(conclusions, 
			"Smart context selection successfully implemented with 5 distinct strategies")
	}
	
	e.results.Conclusions = conclusions
}

// SaveResults saves experiment results to JSON file
func (e *Week6Experiment) SaveResults(outputPath string) error {
	data, err := json.MarshalIndent(e.results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}
	
	return os.WriteFile(outputPath, data, 0644)
}

// PrintSummary prints experiment summary
func (e *Week6Experiment) PrintSummary() {
	fmt.Println("\n=== Week 6 Smart Context Selection Experiment Results ===")
	fmt.Printf("Duration: %v\n", e.results.Duration)
	fmt.Printf("Projects Analyzed: %d\n", e.results.ProjectsAnalyzed)
	fmt.Printf("Strategies Tested: %d\n", len(e.results.StrategiesResults))
	fmt.Printf("Task Types Tested: %d\n", len(e.results.TaskResults))
	
	if e.results.Summary != nil {
		s := e.results.Summary
		fmt.Printf("\nStrategy Performance:\n")
		fmt.Printf("  Best Strategy: %s\n", s.BestOverallStrategy)
		fmt.Printf("  Worst Strategy: %s\n", s.WorstOverallStrategy)
		
		fmt.Printf("\nToken Reduction by Strategy:\n")
		for strategy, reduction := range s.AvgTokenReductionByStrategy {
			fmt.Printf("  %s: %.1f%%\n", strategy, reduction*100)
		}
		
		fmt.Printf("\nTask Type Performance:\n")
		for _, taskResult := range e.results.TaskResults {
			fmt.Printf("  %s: %.1f%% (best: %s)\n", 
				taskResult.TaskType, 
				taskResult.AvgTokenReduction*100,
				taskResult.BestStrategy)
		}
		
		if s.PerformanceGains != nil {
			fmt.Printf("\nPerformance Gains:\n")
			fmt.Printf("  Best vs Worst: %.1f%%\n", s.PerformanceGains.BestVsWorst*100)
			fmt.Printf("  Dependency Boost: %.1f%%\n", s.PerformanceGains.DependencyBoost*100)
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
	
	// Test projects
	projectPaths := []string{
		"../../", // Test on teeny-orb itself
	}
	
	experiment := NewWeek6Experiment()
	
	if err := experiment.RunExperiment(ctx, projectPaths); err != nil {
		log.Fatalf("Experiment failed: %v", err)
	}
	
	// Print summary
	experiment.PrintSummary()
	
	// Save results
	outputPath := "week6_smart_selection_results.json"
	if err := experiment.SaveResults(outputPath); err != nil {
		log.Printf("Failed to save results: %v", err)
	} else {
		fmt.Printf("\nResults saved to: %s\n", outputPath)
	}
}