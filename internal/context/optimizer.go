package context

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ContextOptimizer provides intelligent context selection and optimization
type ContextOptimizer interface {
	// SelectOptimalContext selects the best context for a given task
	SelectOptimalContext(ctx context.Context, project *ProjectContext, task *Task, constraints *ContextConstraints) (*SelectedContext, error)
	
	// ApplyCompressionStrategy applies context compression techniques
	ApplyCompressionStrategy(ctx context.Context, selection *SelectedContext, strategy CompressionStrategy) (*CompressedContext, error)
	
	// CacheContextSelection caches context selection for reuse
	CacheContextSelection(key string, selection *SelectedContext) error
	
	// GetCachedSelection retrieves cached context selection
	GetCachedSelection(key string) (*SelectedContext, bool)
	
	// OptimizeForTokenBudget optimizes context to fit within token budget
	OptimizeForTokenBudget(ctx context.Context, project *ProjectContext, tokenBudget int, task *Task) (*SelectedContext, error)
}

// Task represents a coding task with context requirements
type Task struct {
	Type        TaskType  `json:"type"`
	Description string    `json:"description"`
	Priority    Priority  `json:"priority"`
	Scope       TaskScope `json:"scope"`
	Keywords    []string  `json:"keywords"`
	Files       []string  `json:"files"` // Explicitly mentioned files
	CreatedAt   time.Time `json:"created_at"`
}

// Priority represents task priority levels
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityCritical Priority = "critical"
)

// TaskScope represents the scope of a task
type TaskScope string

const (
	ScopeFile      TaskScope = "file"      // Single file modification
	ScopeModule    TaskScope = "module"    // Module/package level
	ScopeProject   TaskScope = "project"   // Project-wide changes
	ScopeSystem    TaskScope = "system"    // Cross-project dependencies
)

// ContextConstraints defines limits and preferences for context selection
type ContextConstraints struct {
	MaxTokens        int                    `json:"max_tokens"`
	MaxFiles         int                    `json:"max_files"`
	MinRelevanceScore float64              `json:"min_relevance_score"`
	PreferredTypes   []string              `json:"preferred_types"`
	ExcludedPatterns []string              `json:"excluded_patterns"`
	IncludeTests     bool                   `json:"include_tests"`
	IncludeDocs      bool                   `json:"include_docs"`
	FreshnessBias    float64               `json:"freshness_bias"` // 0-1, prefer recently modified files
	DependencyDepth  int                   `json:"dependency_depth"` // How deep to follow dependencies
	Strategy         SelectionStrategy     `json:"strategy"`
}

// SelectionStrategy defines different context selection strategies
type SelectionStrategy string

const (
	StrategyRelevance   SelectionStrategy = "relevance"   // Prioritize by relevance score
	StrategyDependency  SelectionStrategy = "dependency"  // Follow dependency graph
	StrategyFreshness   SelectionStrategy = "freshness"   // Prefer recently modified
	StrategyCompactness SelectionStrategy = "compactness" // Maximize information density
	StrategyBalanced    SelectionStrategy = "balanced"    // Balanced approach
)

// SelectedContext represents optimally selected context for a task
type SelectedContext struct {
	Task             *Task                  `json:"task"`
	Files            []ContextFile          `json:"files"`
	TotalTokens      int                    `json:"total_tokens"`
	TotalFiles       int                    `json:"total_files"`
	SelectionScore   float64                `json:"selection_score"`
	Strategy         SelectionStrategy      `json:"strategy"`
	Constraints      *ContextConstraints    `json:"constraints"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	SelectionTime    time.Duration          `json:"selection_time"`
}

// ContextFile represents a file selected for context with additional metadata
type ContextFile struct {
	FileInfo       *FileInfo `json:"file_info"`
	RelevanceScore float64   `json:"relevance_score"`
	InclusionReason string   `json:"inclusion_reason"`
	Priority       int       `json:"priority"`
	Content        string    `json:"content,omitempty"` // Actual file content if loaded
}

// CompressionStrategy defines different compression approaches
type CompressionStrategy string

const (
	CompressionNone     CompressionStrategy = "none"
	CompressionSummary  CompressionStrategy = "summary"  // Summarize large files
	CompressionSnippet  CompressionStrategy = "snippet"  // Extract relevant snippets
	CompressionMinify   CompressionStrategy = "minify"   // Remove whitespace/comments
	CompressionSemantic CompressionStrategy = "semantic" // Semantic compression
)

// CompressedContext represents context after compression
type CompressedContext struct {
	Original          *SelectedContext    `json:"original"`
	CompressedFiles   []CompressedFile    `json:"compressed_files"`
	CompressionRatio  float64             `json:"compression_ratio"`
	TokenReduction    int                 `json:"token_reduction"`
	Strategy          CompressionStrategy `json:"strategy"`
	QualityScore      float64             `json:"quality_score"`
	CompressionTime   time.Duration       `json:"compression_time"`
}

// CompressedFile represents a compressed file
type CompressedFile struct {
	OriginalPath     string `json:"original_path"`
	CompressedContent string `json:"compressed_content"`
	OriginalTokens   int    `json:"original_tokens"`
	CompressedTokens int    `json:"compressed_tokens"`
	CompressionRatio float64 `json:"compression_ratio"`
	Method           string `json:"method"`
}

// DefaultOptimizer implements the ContextOptimizer interface
type DefaultOptimizer struct {
	analyzer    ContextAnalyzer
	cache       ContextCache
	compressor  ContextCompressor
	config      *OptimizerConfig
}

// OptimizerConfig contains configuration for the context optimizer
type OptimizerConfig struct {
	EnableCaching        bool    `json:"enable_caching"`
	CacheExpiryMinutes   int     `json:"cache_expiry_minutes"`
	DefaultTokenBudget   int     `json:"default_token_budget"`
	MaxSelectionTime     time.Duration `json:"max_selection_time"`
	EnableProfiling      bool    `json:"enable_profiling"`
	DefaultStrategy      SelectionStrategy `json:"default_strategy"`
}

// ContextCache provides caching capabilities for context selections
type ContextCache interface {
	Set(key string, value *SelectedContext, expiry time.Duration) error
	Get(key string) (*SelectedContext, bool)
	Delete(key string) error
	Clear() error
}


// NewDefaultOptimizer creates a new default context optimizer
func NewDefaultOptimizer(analyzer ContextAnalyzer, cache ContextCache, compressor ContextCompressor, config *OptimizerConfig) *DefaultOptimizer {
	if config == nil {
		config = &OptimizerConfig{
			EnableCaching:       true,
			CacheExpiryMinutes:  30,
			DefaultTokenBudget:  8000, // Conservative default
			MaxSelectionTime:    5 * time.Second,
			EnableProfiling:     false,
			DefaultStrategy:     StrategyBalanced,
		}
	}
	
	return &DefaultOptimizer{
		analyzer:   analyzer,
		cache:      cache,
		compressor: compressor,
		config:     config,
	}
}

// SelectOptimalContext selects the best context for a given task
func (o *DefaultOptimizer) SelectOptimalContext(ctx context.Context, project *ProjectContext, task *Task, constraints *ContextConstraints) (*SelectedContext, error) {
	startTime := time.Now()
	
	// Apply default constraints if none provided
	if constraints == nil {
		constraints = o.getDefaultConstraints()
	}
	
	// Check cache first
	if o.config.EnableCaching {
		cacheKey := o.generateCacheKey(project, task, constraints)
		if cached, found := o.GetCachedSelection(cacheKey); found {
			return cached, nil
		}
	}
	
	// Select files based on strategy
	selectedFiles, err := o.selectFilesByStrategy(project, task, constraints)
	if err != nil {
		return nil, fmt.Errorf("failed to select files: %w", err)
	}
	
	// Calculate scores and metadata
	selectionScore := o.calculateSelectionScore(selectedFiles, task)
	totalTokens := o.calculateTotalTokens(selectedFiles)
	
	selection := &SelectedContext{
		Task:            task,
		Files:           selectedFiles,
		TotalTokens:     totalTokens,
		TotalFiles:      len(selectedFiles),
		SelectionScore:  selectionScore,
		Strategy:        constraints.Strategy,
		Constraints:     constraints,
		Metadata:        make(map[string]interface{}),
		CreatedAt:       time.Now(),
		SelectionTime:   time.Since(startTime),
	}
	
	// Cache the selection
	if o.config.EnableCaching {
		cacheKey := o.generateCacheKey(project, task, constraints)
		o.CacheContextSelection(cacheKey, selection)
	}
	
	return selection, nil
}

// OptimizeForTokenBudget optimizes context to fit within token budget
func (o *DefaultOptimizer) OptimizeForTokenBudget(ctx context.Context, project *ProjectContext, tokenBudget int, task *Task) (*SelectedContext, error) {
	constraints := &ContextConstraints{
		MaxTokens:         tokenBudget,
		MaxFiles:          100, // Reasonable default
		MinRelevanceScore: 0.1,
		Strategy:          o.config.DefaultStrategy,
		IncludeTests:      false, // Exclude tests to save tokens
		IncludeDocs:       false, // Exclude docs to save tokens
		FreshnessBias:     0.3,
		DependencyDepth:   2,
	}
	
	// First attempt with default constraints
	selection, err := o.SelectOptimalContext(ctx, project, task, constraints)
	if err != nil {
		return nil, err
	}
	
	// If over budget, progressively tighten constraints
	if selection.TotalTokens > tokenBudget {
		// Try increasing relevance threshold
		constraints.MinRelevanceScore = 0.3
		selection, err = o.SelectOptimalContext(ctx, project, task, constraints)
		if err != nil {
			return nil, err
		}
		
		// If still over budget, reduce dependency depth
		if selection.TotalTokens > tokenBudget {
			constraints.DependencyDepth = 1
			selection, err = o.SelectOptimalContext(ctx, project, task, constraints)
			if err != nil {
				return nil, err
			}
		}
		
		// If still over budget, apply compression
		if selection.TotalTokens > tokenBudget && o.compressor != nil {
			compressed, err := o.ApplyCompressionStrategy(ctx, selection, CompressionSnippet)
			if err != nil {
				return nil, err
			}
			
			// Convert compressed context back to selected context
			selection = o.convertCompressedToSelected(compressed)
		}
	}
	
	return selection, nil
}

// Placeholder implementations
func (o *DefaultOptimizer) ApplyCompressionStrategy(ctx context.Context, selection *SelectedContext, strategy CompressionStrategy) (*CompressedContext, error) {
	if o.compressor == nil {
		return nil, fmt.Errorf("no compressor configured")
	}
	return o.compressor.Compress(ctx, selection, strategy)
}

func (o *DefaultOptimizer) CacheContextSelection(key string, selection *SelectedContext) error {
	if o.cache == nil || !o.config.EnableCaching {
		return nil
	}
	expiry := time.Duration(o.config.CacheExpiryMinutes) * time.Minute
	return o.cache.Set(key, selection, expiry)
}

func (o *DefaultOptimizer) GetCachedSelection(key string) (*SelectedContext, bool) {
	if o.cache == nil || !o.config.EnableCaching {
		return nil, false
	}
	return o.cache.Get(key)
}

// Helper methods
func (o *DefaultOptimizer) getDefaultConstraints() *ContextConstraints {
	return &ContextConstraints{
		MaxTokens:         o.config.DefaultTokenBudget,
		MaxFiles:          50,
		MinRelevanceScore: 0.2,
		PreferredTypes:    []string{"source", "configuration"},
		IncludeTests:      true,
		IncludeDocs:       true,
		FreshnessBias:     0.2,
		DependencyDepth:   3,
		Strategy:          o.config.DefaultStrategy,
	}
}

func (o *DefaultOptimizer) selectFilesByStrategy(project *ProjectContext, task *Task, constraints *ContextConstraints) ([]ContextFile, error) {
	switch constraints.Strategy {
	case StrategyRelevance:
		return o.selectByRelevance(project, task, constraints)
	case StrategyDependency:
		return o.selectByDependency(project, task, constraints)
	case StrategyFreshness:
		return o.selectByFreshness(project, task, constraints)
	case StrategyCompactness:
		return o.selectByCompactness(project, task, constraints)
	case StrategyBalanced:
		return o.selectByBalanced(project, task, constraints)
	default:
		return o.selectByBalanced(project, task, constraints)
	}
}

// selectByRelevance prioritizes files by semantic relevance to the task
func (o *DefaultOptimizer) selectByRelevance(project *ProjectContext, task *Task, constraints *ContextConstraints) ([]ContextFile, error) {
	contextFiles := []ContextFile{}
	
	// Score all files and filter by minimum threshold
	for _, file := range project.Files {
		if o.shouldIncludeFile(&file, task, constraints) {
			score := o.analyzer.ScoreFileRelevance(&file, task.Type, task.Description)
			if score >= constraints.MinRelevanceScore {
				contextFiles = append(contextFiles, ContextFile{
					FileInfo:        &file,
					RelevanceScore:  score,
					InclusionReason: "relevance_score",
					Priority:        1,
				})
			}
		}
	}
	
	// Sort by relevance score (highest first)
	sort.Slice(contextFiles, func(i, j int) bool {
		return contextFiles[i].RelevanceScore > contextFiles[j].RelevanceScore
	})
	
	return o.applyTokenBudget(contextFiles, constraints), nil
}

// selectByDependency prioritizes files based on dependency relationships
func (o *DefaultOptimizer) selectByDependency(project *ProjectContext, task *Task, constraints *ContextConstraints) ([]ContextFile, error) {
	contextFiles := []ContextFile{}
	
	// Score files by dependency centrality and relevance
	for _, file := range project.Files {
		if o.shouldIncludeFile(&file, task, constraints) {
			baseScore := o.analyzer.ScoreFileRelevance(&file, task.Type, task.Description)
			
			// Boost score based on dependency centrality
			var centralityBoost float64 = 0.0
			if project.DependencyGraph != nil {
				centralityBoost = o.calculateDependencyCentrality(project.DependencyGraph, file.Path)
			}
			
			// Combine relevance and centrality (70% relevance, 30% centrality)
			finalScore := baseScore*0.7 + centralityBoost*0.3
			
			if finalScore >= constraints.MinRelevanceScore {
				contextFiles = append(contextFiles, ContextFile{
					FileInfo:        &file,
					RelevanceScore:  finalScore,
					InclusionReason: "dependency_centrality",
					Priority:        1,
				})
			}
		}
	}
	
	// Sort by combined score
	sort.Slice(contextFiles, func(i, j int) bool {
		return contextFiles[i].RelevanceScore > contextFiles[j].RelevanceScore
	})
	
	return o.applyTokenBudget(contextFiles, constraints), nil
}

// selectByFreshness prioritizes recently modified files
func (o *DefaultOptimizer) selectByFreshness(project *ProjectContext, task *Task, constraints *ContextConstraints) ([]ContextFile, error) {
	contextFiles := []ContextFile{}
	
	for _, file := range project.Files {
		if o.shouldIncludeFile(&file, task, constraints) {
			baseScore := o.analyzer.ScoreFileRelevance(&file, task.Type, task.Description)
			
			// Apply freshness bias
			freshnessScore := o.calculateFreshnessScore(file.LastModified)
			finalScore := baseScore*(1-constraints.FreshnessBias) + freshnessScore*constraints.FreshnessBias
			
			if finalScore >= constraints.MinRelevanceScore {
				contextFiles = append(contextFiles, ContextFile{
					FileInfo:        &file,
					RelevanceScore:  finalScore,
					InclusionReason: "freshness_bias",
					Priority:        1,
				})
			}
		}
	}
	
	// Sort by combined score
	sort.Slice(contextFiles, func(i, j int) bool {
		return contextFiles[i].RelevanceScore > contextFiles[j].RelevanceScore
	})
	
	return o.applyTokenBudget(contextFiles, constraints), nil
}

// selectByCompactness prioritizes information density (tokens per relevance)
func (o *DefaultOptimizer) selectByCompactness(project *ProjectContext, task *Task, constraints *ContextConstraints) ([]ContextFile, error) {
	contextFiles := []ContextFile{}
	
	for _, file := range project.Files {
		if o.shouldIncludeFile(&file, task, constraints) {
			relevanceScore := o.analyzer.ScoreFileRelevance(&file, task.Type, task.Description)
			
			if relevanceScore >= constraints.MinRelevanceScore {
				// Calculate compactness: relevance per token
				var compactness float64
				if file.TokenCount > 0 {
					compactness = relevanceScore / float64(file.TokenCount) * 1000 // Scale up for readability
				}
				
				contextFiles = append(contextFiles, ContextFile{
					FileInfo:        &file,
					RelevanceScore:  compactness,
					InclusionReason: "information_density",
					Priority:        1,
				})
			}
		}
	}
	
	// Sort by compactness (highest first)
	sort.Slice(contextFiles, func(i, j int) bool {
		return contextFiles[i].RelevanceScore > contextFiles[j].RelevanceScore
	})
	
	return o.applyTokenBudget(contextFiles, constraints), nil
}

// selectByBalanced uses a balanced approach combining multiple factors
func (o *DefaultOptimizer) selectByBalanced(project *ProjectContext, task *Task, constraints *ContextConstraints) ([]ContextFile, error) {
	contextFiles := []ContextFile{}
	
	for _, file := range project.Files {
		if o.shouldIncludeFile(&file, task, constraints) {
			// Base relevance score
			relevanceScore := o.analyzer.ScoreFileRelevance(&file, task.Type, task.Description)
			
			// Dependency centrality boost
			var centralityBoost float64 = 0.0
			if project.DependencyGraph != nil {
				centralityBoost = o.calculateDependencyCentrality(project.DependencyGraph, file.Path)
			}
			
			// Freshness boost
			freshnessScore := o.calculateFreshnessScore(file.LastModified)
			
			// Size penalty for very large files
			var sizePenalty float64 = 1.0
			if file.TokenCount > 2000 {
				sizePenalty = 2000.0 / float64(file.TokenCount)
			}
			
			// Balanced combination:
			// 50% relevance, 20% centrality, 15% freshness, 15% size efficiency
			balancedScore := relevanceScore*0.5 + 
				centralityBoost*0.2 + 
				freshnessScore*constraints.FreshnessBias*0.15 +
				sizePenalty*0.15
			
			if balancedScore >= constraints.MinRelevanceScore {
				contextFiles = append(contextFiles, ContextFile{
					FileInfo:        &file,
					RelevanceScore:  balancedScore,
					InclusionReason: "balanced_strategy",
					Priority:        1,
				})
			}
		}
	}
	
	// Sort by balanced score
	sort.Slice(contextFiles, func(i, j int) bool {
		return contextFiles[i].RelevanceScore > contextFiles[j].RelevanceScore
	})
	
	return o.applyTokenBudget(contextFiles, constraints), nil
}

// shouldIncludeFile checks if a file should be considered based on constraints
func (o *DefaultOptimizer) shouldIncludeFile(file *FileInfo, task *Task, constraints *ContextConstraints) bool {
	// Check file type preferences
	if len(constraints.PreferredTypes) > 0 {
		found := false
		for _, preferredType := range constraints.PreferredTypes {
			if file.FileType == preferredType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check excluded patterns
	for _, pattern := range constraints.ExcludedPatterns {
		if strings.Contains(file.Path, pattern) {
			return false
		}
	}
	
	// Check test file inclusion
	if !constraints.IncludeTests && file.FileType == "test" {
		return false
	}
	
	// Check documentation inclusion
	if !constraints.IncludeDocs && file.FileType == "documentation" {
		return false
	}
	
	return true
}

// applyTokenBudget applies token budget constraints to file selection
func (o *DefaultOptimizer) applyTokenBudget(contextFiles []ContextFile, constraints *ContextConstraints) []ContextFile {
	selectedFiles := []ContextFile{}
	totalTokens := 0
	
	for _, file := range contextFiles {
		if totalTokens+file.FileInfo.TokenCount <= constraints.MaxTokens && 
		   len(selectedFiles) < constraints.MaxFiles {
			selectedFiles = append(selectedFiles, file)
			totalTokens += file.FileInfo.TokenCount
		} else {
			break
		}
	}
	
	return selectedFiles
}

// calculateDependencyCentrality calculates dependency centrality for a file
func (o *DefaultOptimizer) calculateDependencyCentrality(graph *DependencyGraph, filePath string) float64 {
	// Use relative path for lookup
	relPath := filePath
	if strings.HasPrefix(filePath, "/") {
		// Strip absolute path if needed
		parts := strings.Split(filePath, "/")
		if len(parts) > 2 {
			relPath = strings.Join(parts[len(parts)-2:], "/")
		}
	}
	
	node, exists := graph.Nodes[relPath]
	if !exists {
		return 0.0
	}
	
	totalNodes := len(graph.Nodes)
	if totalNodes <= 1 {
		return 0.5
	}
	
	// Calculate centrality based on incoming and outgoing connections
	inDegree := float64(len(node.Dependents))
	outDegree := float64(len(node.Dependencies))
	
	// Files that many others depend on are more central
	centrality := (inDegree*2 + outDegree) / float64(3*(totalNodes-1))
	
	return min(1.0, centrality)
}

// calculateFreshnessScore calculates freshness score based on modification time
func (o *DefaultOptimizer) calculateFreshnessScore(lastModified time.Time) float64 {
	age := time.Since(lastModified)
	
	// Files modified within 24 hours get full score
	if age < 24*time.Hour {
		return 1.0
	}
	
	// Exponential decay with 1 week half-life
	halfLife := 7 * 24 * time.Hour
	return math.Exp(-0.693 * float64(age) / float64(halfLife))
}

// Note: min function is defined in dependency.go

func (o *DefaultOptimizer) calculateSelectionScore(files []ContextFile, task *Task) float64 {
	if len(files) == 0 {
		return 0.0
	}
	
	totalScore := 0.0
	for _, file := range files {
		totalScore += file.RelevanceScore
	}
	
	return totalScore / float64(len(files))
}

func (o *DefaultOptimizer) calculateTotalTokens(files []ContextFile) int {
	total := 0
	for _, file := range files {
		total += file.FileInfo.TokenCount
	}
	return total
}

func (o *DefaultOptimizer) generateCacheKey(project *ProjectContext, task *Task, constraints *ContextConstraints) string {
	return fmt.Sprintf("ctx_%s_%s_%s_%d", 
		project.RootPath, 
		string(task.Type), 
		task.Description, 
		constraints.MaxTokens)
}

func (o *DefaultOptimizer) convertCompressedToSelected(compressed *CompressedContext) *SelectedContext {
	// Convert compressed context back to selected context format
	// This is a simplified implementation
	return compressed.Original
}