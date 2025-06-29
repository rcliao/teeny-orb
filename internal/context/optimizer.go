package context

import (
	"context"
	"fmt"
	"sort"
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

// ContextCompressor provides context compression capabilities
type ContextCompressor interface {
	Compress(ctx context.Context, selection *SelectedContext, strategy CompressionStrategy) (*CompressedContext, error)
	EstimateCompression(selection *SelectedContext, strategy CompressionStrategy) (float64, error)
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
	// TODO: Implement sophisticated file selection logic
	// For now, return a basic selection
	contextFiles := []ContextFile{}
	
	// Score all files
	for _, file := range project.Files {
		score := o.analyzer.ScoreFileRelevance(&file, task.Type, task.Description)
		if score >= constraints.MinRelevanceScore {
			contextFiles = append(contextFiles, ContextFile{
				FileInfo:        &file,
				RelevanceScore:  score,
				InclusionReason: "relevance_threshold",
				Priority:        1,
			})
		}
	}
	
	// Sort by relevance score
	sort.Slice(contextFiles, func(i, j int) bool {
		return contextFiles[i].RelevanceScore > contextFiles[j].RelevanceScore
	})
	
	// Apply token budget constraint
	totalTokens := 0
	selectedFiles := []ContextFile{}
	for _, file := range contextFiles {
		if totalTokens+file.FileInfo.TokenCount <= constraints.MaxTokens && len(selectedFiles) < constraints.MaxFiles {
			selectedFiles = append(selectedFiles, file)
			totalTokens += file.FileInfo.TokenCount
		}
	}
	
	return selectedFiles, nil
}

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