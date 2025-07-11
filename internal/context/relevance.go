package context

import (
	"math"
	"path/filepath"
	"strings"
	"time"
)

// RelevanceScorer provides intelligent file relevance scoring
type RelevanceScorer interface {
	// ScoreFile calculates relevance score for a file given a task
	ScoreFile(file *FileInfo, task *Task) float64
	
	// ScoreFiles scores multiple files and returns sorted results
	ScoreFiles(files []FileInfo, task *Task) []ScoredFile
	
	// GetScoringFactors returns the individual scoring factors for debugging
	GetScoringFactors(file *FileInfo, task *Task) ScoringFactors
}

// ScoredFile represents a file with its relevance score
type ScoredFile struct {
	File   *FileInfo
	Score  float64
	Factors ScoringFactors
}

// ScoringFactors breaks down the components of a relevance score
type ScoringFactors struct {
	KeywordMatch      float64 `json:"keyword_match"`
	PathRelevance     float64 `json:"path_relevance"`
	FileTypeScore     float64 `json:"file_type_score"`
	RecencyScore      float64 `json:"recency_score"`
	SizeScore         float64 `json:"size_score"`
	DependencyScore   float64 `json:"dependency_score"`
	TaskTypeScore     float64 `json:"task_type_score"`
	LanguageScore     float64 `json:"language_score"`
}

// SemanticRelevanceScorer implements intelligent relevance scoring
type SemanticRelevanceScorer struct {
	config *RelevanceScorerConfig
}

// RelevanceScorerConfig configures the relevance scoring behavior
type RelevanceScorerConfig struct {
	// Weight factors for different scoring components (must sum to 1.0)
	Weights struct {
		KeywordMatch    float64
		PathRelevance   float64
		FileType        float64
		Recency         float64
		Size            float64
		Dependency      float64
		TaskType        float64
		Language        float64
	}
	
	// Recency decay parameters
	RecencyHalfLife time.Duration // How fast recency score decays
	
	// Size preferences
	OptimalFileSize int64 // Optimal file size in tokens
	SizePenalty     float64 // Penalty multiplier for oversized files
	
	// Task-specific boosts
	TaskTypeBoosts map[TaskType]map[string]float64 // file type boosts per task
	
	// Keyword matching
	StopWords       []string // Words to ignore in keyword matching
	StemWords       bool     // Whether to use word stemming
}

// NewSemanticRelevanceScorer creates a new relevance scorer with default config
func NewSemanticRelevanceScorer(config *RelevanceScorerConfig) *SemanticRelevanceScorer {
	if config == nil {
		config = getDefaultRelevanceScorerConfig()
	}
	return &SemanticRelevanceScorer{config: config}
}

// ScoreFile calculates the relevance score for a single file
func (s *SemanticRelevanceScorer) ScoreFile(file *FileInfo, task *Task) float64 {
	factors := s.GetScoringFactors(file, task)
	
	// Weighted sum of all factors
	score := factors.KeywordMatch * s.config.Weights.KeywordMatch +
		factors.PathRelevance * s.config.Weights.PathRelevance +
		factors.FileTypeScore * s.config.Weights.FileType +
		factors.RecencyScore * s.config.Weights.Recency +
		factors.SizeScore * s.config.Weights.Size +
		factors.DependencyScore * s.config.Weights.Dependency +
		factors.TaskTypeScore * s.config.Weights.TaskType +
		factors.LanguageScore * s.config.Weights.Language
	
	// Ensure score is between 0 and 1
	return math.Max(0, math.Min(1, score))
}

// ScoreFiles scores multiple files and returns sorted results
func (s *SemanticRelevanceScorer) ScoreFiles(files []FileInfo, task *Task) []ScoredFile {
	scored := make([]ScoredFile, len(files))
	
	for i, file := range files {
		scored[i] = ScoredFile{
			File:    &files[i],
			Score:   s.ScoreFile(&file, task),
			Factors: s.GetScoringFactors(&file, task),
		}
	}
	
	// Sort by score descending (highest relevance first)
	// Using simple bubble sort for now - can optimize later
	for i := 0; i < len(scored)-1; i++ {
		for j := 0; j < len(scored)-i-1; j++ {
			if scored[j].Score < scored[j+1].Score {
				scored[j], scored[j+1] = scored[j+1], scored[j]
			}
		}
	}
	
	return scored
}

// GetScoringFactors returns detailed scoring breakdown
func (s *SemanticRelevanceScorer) GetScoringFactors(file *FileInfo, task *Task) ScoringFactors {
	return ScoringFactors{
		KeywordMatch:    s.calculateKeywordMatch(file, task),
		PathRelevance:   s.calculatePathRelevance(file, task),
		FileTypeScore:   s.calculateFileTypeScore(file, task),
		RecencyScore:    s.calculateRecencyScore(file),
		SizeScore:       s.calculateSizeScore(file),
		DependencyScore: s.calculateDependencyScore(file, task),
		TaskTypeScore:   s.calculateTaskTypeScore(file, task),
		LanguageScore:   s.calculateLanguageScore(file, task),
	}
}

// calculateKeywordMatch scores based on keyword presence in file path/name
func (s *SemanticRelevanceScorer) calculateKeywordMatch(file *FileInfo, task *Task) float64 {
	if len(task.Keywords) == 0 && task.Description == "" {
		return 0.5 // Neutral score if no keywords
	}
	
	// Extract keywords from task description if not explicitly provided
	keywords := task.Keywords
	if len(keywords) == 0 {
		keywords = s.extractKeywords(task.Description)
	}
	
	// Also check explicitly mentioned files
	for _, mentionedFile := range task.Files {
		if strings.Contains(file.Path, mentionedFile) {
			return 1.0 // Perfect match for explicitly mentioned files
		}
	}
	
	// Score based on keyword matches in file path and name
	fileName := strings.ToLower(filepath.Base(file.Path))
	filePath := strings.ToLower(file.Path)
	
	matchCount := 0
	for _, keyword := range keywords {
		keyword = strings.ToLower(keyword)
		if strings.Contains(fileName, keyword) {
			matchCount += 2 // Double weight for filename matches
		}
		if strings.Contains(filePath, keyword) {
			matchCount += 1
		}
	}
	
	// Normalize by number of keywords
	if len(keywords) > 0 {
		return math.Min(1.0, float64(matchCount)/(float64(len(keywords))*2))
	}
	
	return 0.5
}

// calculatePathRelevance scores based on path structure
func (s *SemanticRelevanceScorer) calculatePathRelevance(file *FileInfo, task *Task) float64 {
	path := strings.ToLower(file.Path)
	
	// Core paths get higher scores
	corePathScores := map[string]float64{
		"/internal/": 0.8,
		"/pkg/":      0.7,
		"/cmd/":      0.9,
		"/api/":      0.8,
		"/core/":     0.9,
		"/lib/":      0.7,
		"/src/":      0.8,
	}
	
	// Test and doc paths get lower scores for non-test/doc tasks
	if task.Type != TaskTypeTest && strings.Contains(path, "/test") {
		return 0.2
	}
	if task.Type != TaskTypeDocumentation && strings.Contains(path, "/doc") {
		return 0.3
	}
	
	// Check for core path patterns
	for pattern, score := range corePathScores {
		if strings.Contains(path, pattern) {
			return score
		}
	}
	
	// Vendor/external dependencies get low scores
	if strings.Contains(path, "/vendor/") || strings.Contains(path, "/node_modules/") {
		return 0.1
	}
	
	return 0.5 // Default neutral score
}

// calculateFileTypeScore scores based on file type relevance to task
func (s *SemanticRelevanceScorer) calculateFileTypeScore(file *FileInfo, task *Task) float64 {
	// Task-specific file type preferences
	taskTypePreferences := map[TaskType]map[string]float64{
		TaskTypeFeature: {
			"source":        0.9,
			"test":          0.3,
			"configuration": 0.5,
			"documentation": 0.2,
		},
		TaskTypeDebug: {
			"source":        1.0,
			"test":          0.7,
			"configuration": 0.4,
			"documentation": 0.1,
		},
		TaskTypeRefactor: {
			"source":        1.0,
			"test":          0.8,
			"configuration": 0.3,
			"documentation": 0.2,
		},
		TaskTypeTest: {
			"source":        0.8,
			"test":          1.0,
			"configuration": 0.3,
			"documentation": 0.2,
		},
		TaskTypeDocumentation: {
			"source":        0.5,
			"test":          0.2,
			"configuration": 0.4,
			"documentation": 1.0,
		},
	}
	
	if prefs, exists := taskTypePreferences[task.Type]; exists {
		if score, exists := prefs[file.FileType]; exists {
			return score
		}
	}
	
	return 0.5 // Default neutral score
}

// calculateRecencyScore scores based on file modification time
func (s *SemanticRelevanceScorer) calculateRecencyScore(file *FileInfo) float64 {
	age := time.Since(file.LastModified)
	halfLife := s.config.RecencyHalfLife
	
	// Exponential decay based on half-life
	return math.Exp(-0.693 * float64(age) / float64(halfLife))
}

// calculateSizeScore scores based on file size
func (s *SemanticRelevanceScorer) calculateSizeScore(file *FileInfo) float64 {
	optimalSize := float64(s.config.OptimalFileSize)
	actualSize := float64(file.TokenCount)
	
	if actualSize <= optimalSize {
		// Linear increase up to optimal size
		return actualSize / optimalSize
	}
	
	// Penalty for oversized files
	oversize := actualSize - optimalSize
	penaltyFactor := 1.0 - (oversize/optimalSize)*s.config.SizePenalty
	return math.Max(0.3, penaltyFactor) // Minimum score of 0.3
}

// calculateDependencyScore scores based on dependency relationships
func (s *SemanticRelevanceScorer) calculateDependencyScore(file *FileInfo, task *Task) float64 {
	// For now, return neutral score - will be enhanced when dependency graph is implemented
	// This will consider:
	// - How many other relevant files depend on this file
	// - How many files this file depends on
	// - Centrality in the dependency graph
	return 0.5
}

// calculateTaskTypeScore provides task-specific scoring
func (s *SemanticRelevanceScorer) calculateTaskTypeScore(file *FileInfo, task *Task) float64 {
	// Check for task-specific boosts
	if boosts, exists := s.config.TaskTypeBoosts[task.Type]; exists {
		if boost, exists := boosts[file.FileType]; exists {
			return boost
		}
	}
	
	// Special patterns for different task types
	switch task.Type {
	case TaskTypeDebug:
		// Debugging often involves logs, error handling
		if strings.Contains(strings.ToLower(file.Path), "error") ||
			strings.Contains(strings.ToLower(file.Path), "log") {
			return 0.8
		}
	case TaskTypeTest:
		// Test tasks focus on test files
		if strings.Contains(file.Path, "_test") || strings.Contains(file.Path, "test_") {
			return 1.0
		}
	case TaskTypeRefactor:
		// Refactoring needs to see interfaces and abstractions
		if strings.Contains(strings.ToLower(file.Path), "interface") ||
			strings.Contains(strings.ToLower(file.Path), "abstract") {
			return 0.8
		}
	}
	
	return 0.5
}

// calculateLanguageScore scores based on language relevance
func (s *SemanticRelevanceScorer) calculateLanguageScore(file *FileInfo, task *Task) float64 {
	// Certain languages are more relevant for certain tasks
	languageRelevance := map[string]map[TaskType]float64{
		"go": {
			TaskTypeFeature:       0.9,
			TaskTypeDebug:         0.9,
			TaskTypeRefactor:      0.9,
			TaskTypeTest:          0.9,
			TaskTypeDocumentation: 0.6,
		},
		"markdown": {
			TaskTypeFeature:       0.3,
			TaskTypeDebug:         0.2,
			TaskTypeRefactor:      0.2,
			TaskTypeTest:          0.3,
			TaskTypeDocumentation: 1.0,
		},
		"yaml": {
			TaskTypeFeature:       0.5,
			TaskTypeDebug:         0.4,
			TaskTypeRefactor:      0.3,
			TaskTypeTest:          0.4,
			TaskTypeDocumentation: 0.6,
		},
	}
	
	if langScores, exists := languageRelevance[file.Language]; exists {
		if score, exists := langScores[task.Type]; exists {
			return score
		}
	}
	
	return 0.5
}

// extractKeywords extracts keywords from task description
func (s *SemanticRelevanceScorer) extractKeywords(description string) []string {
	// Simple keyword extraction - can be enhanced with NLP
	words := strings.Fields(strings.ToLower(description))
	keywords := []string{}
	
	stopWords := make(map[string]bool)
	for _, word := range s.config.StopWords {
		stopWords[word] = true
	}
	
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'")
		
		// Skip stop words and very short words
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// getDefaultRelevanceScorerConfig returns default configuration
func getDefaultRelevanceScorerConfig() *RelevanceScorerConfig {
	config := &RelevanceScorerConfig{
		RecencyHalfLife: 7 * 24 * time.Hour, // 1 week half-life
		OptimalFileSize: 500,                 // 500 tokens is optimal
		SizePenalty:     0.5,
		TaskTypeBoosts:  make(map[TaskType]map[string]float64),
		StopWords: []string{
			"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for",
			"of", "with", "by", "from", "as", "is", "was", "are", "were", "been",
			"be", "have", "has", "had", "do", "does", "did", "will", "would",
			"should", "could", "may", "might", "must", "can", "this", "that",
			"these", "those", "i", "you", "he", "she", "it", "we", "they",
		},
		StemWords: false, // Disabled by default for simplicity
	}
	
	// Set default weights (must sum to 1.0)
	config.Weights.KeywordMatch = 0.25
	config.Weights.PathRelevance = 0.15
	config.Weights.FileType = 0.20
	config.Weights.Recency = 0.10
	config.Weights.Size = 0.05
	config.Weights.Dependency = 0.10
	config.Weights.TaskType = 0.10
	config.Weights.Language = 0.05
	
	return config
}