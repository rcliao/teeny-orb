package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TaskType represents different types of coding tasks for context optimization
type TaskType string

const (
	TaskTypeGeneral    TaskType = "general"
	TaskTypeDebug      TaskType = "debug"
	TaskTypeRefactor   TaskType = "refactor"
	TaskTypeFeature    TaskType = "feature"
	TaskTypeTest       TaskType = "test"
	TaskTypeDocumentation TaskType = "documentation"
)

// FileInfo represents analyzed file information
type FileInfo struct {
	Path         string            `json:"path"`
	Size         int64             `json:"size"`
	TokenCount   int               `json:"token_count"`
	LastModified time.Time         `json:"last_modified"`
	FileType     string            `json:"file_type"`
	Language     string            `json:"language"`
	RelevanceScore float64         `json:"relevance_score"`
	Dependencies []string          `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ProjectContext represents the analyzed context of a project
type ProjectContext struct {
	RootPath      string                `json:"root_path"`
	TotalFiles    int                   `json:"total_files"`
	TotalTokens   int                   `json:"total_tokens"`
	Files         []FileInfo            `json:"files"`
	DependencyGraph *DependencyGraph    `json:"dependency_graph"`
	Languages     map[string]int        `json:"languages"`
	Analysis      *ContextAnalysis      `json:"analysis"`
	CreatedAt     time.Time             `json:"created_at"`
}

// DependencyGraph represents file dependencies within a project
type DependencyGraph struct {
	Nodes map[string]*DependencyNode `json:"nodes"`
	Edges []DependencyEdge           `json:"edges"`
}

// DependencyNode represents a file in the dependency graph
type DependencyNode struct {
	Path         string   `json:"path"`
	Imports      []string `json:"imports"`
	Exports      []string `json:"exports"`
	Dependencies []string `json:"dependencies"`
	Dependents   []string `json:"dependents"`
}

// DependencyEdge represents a dependency relationship
type DependencyEdge struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Type       string `json:"type"` // "import", "reference", etc.
	Strength   float64 `json:"strength"`
}

// ContextAnalysis provides insights about project context
type ContextAnalysis struct {
	CoreFiles        []string          `json:"core_files"`
	EntryPoints      []string          `json:"entry_points"`
	TestFiles        []string          `json:"test_files"`
	ConfigFiles      []string          `json:"config_files"`
	LanguageStats    map[string]int    `json:"language_stats"`
	ComplexityMetrics map[string]float64 `json:"complexity_metrics"`
	Recommendations  []string          `json:"recommendations"`
}

// ContextAnalyzer provides project context analysis capabilities
type ContextAnalyzer interface {
	// AnalyzeProject performs comprehensive project analysis
	AnalyzeProject(ctx context.Context, rootPath string) (*ProjectContext, error)
	
	// ScoreFileRelevance calculates relevance score for a file given a task type
	ScoreFileRelevance(file *FileInfo, taskType TaskType, taskDescription string) float64
	
	// BuildDependencyGraph constructs dependency relationships between files
	BuildDependencyGraph(ctx context.Context, files []FileInfo) (*DependencyGraph, error)
	
	// CountTokens estimates token count for file content
	CountTokens(content string) (int, error)
	
	// GetFileInfo analyzes a single file
	GetFileInfo(ctx context.Context, filePath string) (*FileInfo, error)
	
	// FilterFilesByType filters files by type or pattern
	FilterFilesByType(files []FileInfo, fileTypes []string) []FileInfo
	
	// SortFilesByRelevance sorts files by relevance score
	SortFilesByRelevance(files []FileInfo) []FileInfo
}

// DefaultAnalyzer implements the ContextAnalyzer interface
type DefaultAnalyzer struct {
	tokenCounter TokenCounter
	config       *AnalyzerConfig
}

// AnalyzerConfig contains configuration for the context analyzer
type AnalyzerConfig struct {
	MaxFileSize       int64             `json:"max_file_size"`
	IgnorePatterns    []string          `json:"ignore_patterns"`
	SupportedLanguages map[string][]string `json:"supported_languages"`
	TokenCountCache   bool              `json:"token_count_cache"`
	EnableProfiling   bool              `json:"enable_profiling"`
}

// TokenCounter provides token counting capabilities
type TokenCounter interface {
	CountTokens(content string) (int, error)
	CountFile(filePath string) (int, error)
}

// NewDefaultAnalyzer creates a new default context analyzer
func NewDefaultAnalyzer(tokenCounter TokenCounter, config *AnalyzerConfig) *DefaultAnalyzer {
	if config == nil {
		config = &AnalyzerConfig{
			MaxFileSize: 1024 * 1024, // 1MB
			IgnorePatterns: []string{
				".git/*", "node_modules/*", "vendor/*", "*.log",
				"*.tmp", "*.cache", "build/*", "dist/*",
			},
			SupportedLanguages: map[string][]string{
				"go":         {".go"},
				"javascript": {".js", ".ts", ".jsx", ".tsx"},
				"python":     {".py"},
				"java":       {".java"},
				"c++":        {".cpp", ".cc", ".cxx", ".c"},
				"rust":       {".rs"},
				"markdown":   {".md", ".mdx"},
				"yaml":       {".yml", ".yaml"},
				"json":       {".json"},
			},
			TokenCountCache: true,
			EnableProfiling: false,
		}
	}
	
	return &DefaultAnalyzer{
		tokenCounter: tokenCounter,
		config:       config,
	}
}

// AnalyzeProject performs comprehensive project analysis
func (a *DefaultAnalyzer) AnalyzeProject(ctx context.Context, rootPath string) (*ProjectContext, error) {
	startTime := time.Now()
	
	projectCtx := &ProjectContext{
		RootPath:    rootPath,
		Files:       []FileInfo{},
		Languages:   make(map[string]int),
		CreatedAt:   startTime,
	}
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories and ignored files
		if info.IsDir() || a.shouldIgnoreFile(path) {
			return nil
		}
		
		// Skip files that are too large
		if info.Size() > a.config.MaxFileSize {
			return nil
		}
		
		fileInfo, err := a.GetFileInfo(ctx, path)
		if err != nil {
			// Log error but continue processing
			return nil
		}
		
		projectCtx.Files = append(projectCtx.Files, *fileInfo)
		projectCtx.TotalFiles++
		projectCtx.TotalTokens += fileInfo.TokenCount
		
		// Update language statistics
		if fileInfo.Language != "" {
			projectCtx.Languages[fileInfo.Language]++
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk project directory: %w", err)
	}
	
	// Build dependency graph
	dependencyGraph, err := a.BuildDependencyGraph(ctx, projectCtx.Files)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}
	projectCtx.DependencyGraph = dependencyGraph
	
	// Perform analysis
	analysis := a.analyzeProjectStructure(projectCtx)
	projectCtx.Analysis = analysis
	
	return projectCtx, nil
}

// GetFileInfo analyzes a single file
func (a *DefaultAnalyzer) GetFileInfo(ctx context.Context, filePath string) (*FileInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	tokenCount := 0
	if a.tokenCounter != nil {
		tokenCount, _ = a.tokenCounter.CountTokens(string(content))
	}
	
	fileInfo := &FileInfo{
		Path:         filePath,
		Size:         stat.Size(),
		TokenCount:   tokenCount,
		LastModified: stat.ModTime(),
		FileType:     a.getFileType(filePath),
		Language:     a.detectLanguage(filePath),
		Metadata:     make(map[string]interface{}),
	}
	
	return fileInfo, nil
}

// shouldIgnoreFile checks if a file should be ignored based on patterns
func (a *DefaultAnalyzer) shouldIgnoreFile(path string) bool {
	for _, pattern := range a.config.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if strings.Contains(path, strings.TrimSuffix(pattern, "/*")) {
			return true
		}
	}
	return false
}

// getFileType determines the file type based on extension
func (a *DefaultAnalyzer) getFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return "source"
	case ".md", ".txt", ".rst":
		return "documentation"
	case ".yml", ".yaml", ".json", ".toml":
		return "configuration"
	case ".test.go", "_test.go":
		return "test"
	case ".sh", ".bat", ".ps1":
		return "script"
	default:
		if strings.Contains(filePath, "test") {
			return "test"
		}
		return "unknown"
	}
}

// detectLanguage detects programming language based on file extension
func (a *DefaultAnalyzer) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	for language, extensions := range a.config.SupportedLanguages {
		for _, supportedExt := range extensions {
			if ext == supportedExt {
				return language
			}
		}
	}
	
	return "unknown"
}

// analyzeProjectStructure performs high-level project analysis
func (a *DefaultAnalyzer) analyzeProjectStructure(projectCtx *ProjectContext) *ContextAnalysis {
	analysis := &ContextAnalysis{
		CoreFiles:         []string{},
		EntryPoints:      []string{},
		TestFiles:        []string{},
		ConfigFiles:      []string{},
		LanguageStats:    make(map[string]int),
		ComplexityMetrics: make(map[string]float64),
		Recommendations:  []string{},
	}
	
	for _, file := range projectCtx.Files {
		switch file.FileType {
		case "test":
			analysis.TestFiles = append(analysis.TestFiles, file.Path)
		case "configuration":
			analysis.ConfigFiles = append(analysis.ConfigFiles, file.Path)
		case "source":
			if strings.Contains(file.Path, "main.go") || strings.Contains(file.Path, "cmd/") {
				analysis.EntryPoints = append(analysis.EntryPoints, file.Path)
			}
		}
		
		analysis.LanguageStats[file.Language]++
	}
	
	// Calculate complexity metrics
	analysis.ComplexityMetrics["total_files"] = float64(projectCtx.TotalFiles)
	analysis.ComplexityMetrics["total_tokens"] = float64(projectCtx.TotalTokens)
	analysis.ComplexityMetrics["avg_tokens_per_file"] = float64(projectCtx.TotalTokens) / float64(projectCtx.TotalFiles)
	
	// Generate recommendations
	if projectCtx.TotalTokens > 100000 {
		analysis.Recommendations = append(analysis.Recommendations, "Large codebase detected - context optimization recommended")
	}
	
	return analysis
}

// Placeholder implementations for interface compliance
func (a *DefaultAnalyzer) ScoreFileRelevance(file *FileInfo, taskType TaskType, taskDescription string) float64 {
	// TODO: Implement sophisticated relevance scoring
	return 0.5
}

func (a *DefaultAnalyzer) BuildDependencyGraph(ctx context.Context, files []FileInfo) (*DependencyGraph, error) {
	// TODO: Implement dependency analysis
	return &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: []DependencyEdge{},
	}, nil
}

func (a *DefaultAnalyzer) CountTokens(content string) (int, error) {
	if a.tokenCounter != nil {
		return a.tokenCounter.CountTokens(content)
	}
	return len(strings.Fields(content)), nil // Rough approximation
}

func (a *DefaultAnalyzer) FilterFilesByType(files []FileInfo, fileTypes []string) []FileInfo {
	filtered := []FileInfo{}
	for _, file := range files {
		for _, fileType := range fileTypes {
			if file.FileType == fileType {
				filtered = append(filtered, file)
				break
			}
		}
	}
	return filtered
}

func (a *DefaultAnalyzer) SortFilesByRelevance(files []FileInfo) []FileInfo {
	// TODO: Implement proper sorting by relevance score
	return files
}