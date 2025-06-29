package context

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// SimpleTokenCounter provides basic token counting functionality
// This is a simplified implementation for Phase 2 experiments
// In production, you would integrate with tiktoken-go or similar
type SimpleTokenCounter struct {
	// Approximate tokens per word ratio for different languages
	languageMultipliers map[string]float64
}

// NewSimpleTokenCounter creates a new simple token counter
func NewSimpleTokenCounter() *SimpleTokenCounter {
	return &SimpleTokenCounter{
		languageMultipliers: map[string]float64{
			"go":         1.3,  // Go code tends to be more verbose
			"javascript": 1.2,  // JavaScript has moderate token density
			"python":     1.1,  // Python is relatively concise
			"java":       1.4,  // Java is quite verbose
			"c++":        1.3,  // C++ moderate verbosity
			"rust":       1.2,  // Rust moderate verbosity
			"markdown":   0.8,  // Documentation is more natural language
			"yaml":       0.9,  // Config files are structured
			"json":       1.0,  // JSON is structured data
			"unknown":    1.0,  // Default multiplier
		},
	}
}

// CountTokens estimates token count for text content
func (tc *SimpleTokenCounter) CountTokens(content string) (int, error) {
	if content == "" {
		return 0, nil
	}
	
	// Basic token estimation algorithm:
	// 1. Split into words
	// 2. Count punctuation separately
	// 3. Apply language-specific multipliers if available
	
	words := tc.countWords(content)
	punctuation := tc.countPunctuation(content)
	symbols := tc.countSymbols(content)
	
	// Base token count (words + punctuation + symbols)
	baseTokens := words + punctuation + symbols
	
	// Apply a general multiplier for subword tokenization
	// Most modern tokenizers split words into subwords
	estimatedTokens := float64(baseTokens) * 1.2
	
	return int(estimatedTokens), nil
}

// CountFile estimates token count for a file
func (tc *SimpleTokenCounter) CountFile(filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	return tc.CountTokens(string(content))
}

// CountTokensWithLanguage estimates tokens with language-specific adjustments
func (tc *SimpleTokenCounter) CountTokensWithLanguage(content string, language string) (int, error) {
	baseTokens, err := tc.CountTokens(content)
	if err != nil {
		return 0, err
	}
	
	multiplier, exists := tc.languageMultipliers[language]
	if !exists {
		multiplier = tc.languageMultipliers["unknown"]
	}
	
	return int(float64(baseTokens) * multiplier), nil
}

// countWords counts words in the content
func (tc *SimpleTokenCounter) countWords(content string) int {
	if content == "" {
		return 0
	}
	
	words := 0
	inWord := false
	
	for _, r := range content {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if !inWord {
				words++
				inWord = true
			}
		} else {
			inWord = false
		}
	}
	
	return words
}

// countPunctuation counts punctuation marks
func (tc *SimpleTokenCounter) countPunctuation(content string) int {
	count := 0
	for _, r := range content {
		if unicode.IsPunct(r) {
			count++
		}
	}
	return count
}

// countSymbols counts programming symbols and operators
func (tc *SimpleTokenCounter) countSymbols(content string) int {
	symbolChars := "{}[]()+-*/=<>!&|^~%#@$"
	count := 0
	
	for _, r := range content {
		if strings.ContainsRune(symbolChars, r) {
			count++
		}
	}
	
	return count
}

// GetEstimationAccuracy provides accuracy estimates for different content types
func (tc *SimpleTokenCounter) GetEstimationAccuracy() map[string]float64 {
	return map[string]float64{
		"source_code":    0.75, // 75% accuracy for source code
		"documentation":  0.85, // 85% accuracy for natural language
		"configuration":  0.80, // 80% accuracy for config files
		"mixed_content":  0.70, // 70% accuracy for mixed content
	}
}

// GetTokenStatistics provides detailed token statistics for content
func (tc *SimpleTokenCounter) GetTokenStatistics(content string) (*TokenStatistics, error) {
	words := tc.countWords(content)
	punctuation := tc.countPunctuation(content)
	symbols := tc.countSymbols(content)
	lines := len(strings.Split(content, "\n"))
	characters := len(content)
	
	totalTokens, err := tc.CountTokens(content)
	if err != nil {
		return nil, err
	}
	
	stats := &TokenStatistics{
		TotalTokens:     totalTokens,
		Words:           words,
		Punctuation:     punctuation,
		Symbols:         symbols,
		Lines:           lines,
		Characters:      characters,
		TokensPerLine:   0,
		TokensPerWord:   0,
		CharactersPerToken: 0,
	}
	
	if lines > 0 {
		stats.TokensPerLine = float64(totalTokens) / float64(lines)
	}
	if words > 0 {
		stats.TokensPerWord = float64(totalTokens) / float64(words)
	}
	if totalTokens > 0 {
		stats.CharactersPerToken = float64(characters) / float64(totalTokens)
	}
	
	return stats, nil
}

// TokenStatistics provides detailed breakdown of token analysis
type TokenStatistics struct {
	TotalTokens        int     `json:"total_tokens"`
	Words              int     `json:"words"`
	Punctuation        int     `json:"punctuation"`
	Symbols            int     `json:"symbols"`
	Lines              int     `json:"lines"`
	Characters         int     `json:"characters"`
	TokensPerLine      float64 `json:"tokens_per_line"`
	TokensPerWord      float64 `json:"tokens_per_word"`
	CharactersPerToken float64 `json:"characters_per_token"`
}

// Ensure SimpleTokenCounter implements TokenCounter interface
var _ TokenCounter = (*SimpleTokenCounter)(nil)