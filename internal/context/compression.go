package context

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ContextCompressor provides various context compression strategies
type ContextCompressor interface {
	// Compress applies compression to a selected context
	Compress(ctx context.Context, selection *SelectedContext, strategy CompressionStrategy) (*CompressedContext, error)
	
	// EstimateCompression estimates compression ratio without actually compressing
	EstimateCompression(selection *SelectedContext, strategy CompressionStrategy) (float64, error)
	
	// GetCompressionStrategies returns available compression strategies
	GetCompressionStrategies() []CompressionStrategy
}

// DefaultContextCompressor implements various compression techniques
type DefaultContextCompressor struct {
	tokenCounter TokenCounter
	config       *CompressionConfig
}

// CompressionConfig configures compression behavior
type CompressionConfig struct {
	MaxCompressionRatio  float64           `json:"max_compression_ratio"`
	PreserveImports      bool              `json:"preserve_imports"`
	PreserveFunctionSigs bool              `json:"preserve_function_signatures"`
	PreserveComments     bool              `json:"preserve_comments"`
	MinFunctionLines     int               `json:"min_function_lines"`
	SnippetContextLines  int               `json:"snippet_context_lines"`
	SummaryMaxTokens     int               `json:"summary_max_tokens"`
	EnableSemanticCompr  bool              `json:"enable_semantic_compression"`
	LanguageRules        map[string]*LanguageCompressionRules `json:"language_rules"`
}

// LanguageCompressionRules defines compression rules for specific languages
type LanguageCompressionRules struct {
	ImportPatterns      []string `json:"import_patterns"`
	FunctionPatterns    []string `json:"function_patterns"`
	ClassPatterns       []string `json:"class_patterns"`
	CommentPatterns     []string `json:"comment_patterns"`
	PreservePatterns    []string `json:"preserve_patterns"`
	RemovablePatterns   []string `json:"removable_patterns"`
}

// CompressionAnalysis provides detailed compression analysis
type CompressionAnalysis struct {
	OriginalTokens    int                    `json:"original_tokens"`
	CompressedTokens  int                    `json:"compressed_tokens"`
	CompressionRatio  float64                `json:"compression_ratio"`
	TokenSavings      int                    `json:"token_savings"`
	QualityEstimate   float64                `json:"quality_estimate"`
	TechniquesApplied []string               `json:"techniques_applied"`
	FileAnalysis      []FileCompressionInfo  `json:"file_analysis"`
}

// FileCompressionInfo tracks compression for individual files
type FileCompressionInfo struct {
	FilePath          string   `json:"file_path"`
	OriginalTokens    int      `json:"original_tokens"`
	CompressedTokens  int      `json:"compressed_tokens"`
	CompressionRatio  float64  `json:"compression_ratio"`
	TechniquesUsed    []string `json:"techniques_used"`
	QualityImpact     float64  `json:"quality_impact"`
}

// NewDefaultContextCompressor creates a new context compressor
func NewDefaultContextCompressor(tokenCounter TokenCounter, config *CompressionConfig) *DefaultContextCompressor {
	if config == nil {
		config = &CompressionConfig{
			MaxCompressionRatio:  0.8,
			PreserveImports:      true,
			PreserveFunctionSigs: true,
			PreserveComments:     false,
			MinFunctionLines:     3,
			SnippetContextLines:  2,
			SummaryMaxTokens:     200,
			EnableSemanticCompr:  true,
			LanguageRules:        getDefaultLanguageRules(),
		}
	}

	return &DefaultContextCompressor{
		tokenCounter: tokenCounter,
		config:       config,
	}
}

// Compress applies compression to a selected context
func (c *DefaultContextCompressor) Compress(ctx context.Context, selection *SelectedContext, strategy CompressionStrategy) (*CompressedContext, error) {
	startTime := time.Now()
	
	compressed := &CompressedContext{
		Original:         selection,
		CompressedFiles:  []CompressedFile{},
		CompressionRatio: 1.0,
		TokenReduction:   0,
		Strategy:         strategy,
		QualityScore:     1.0,
		CompressionTime:  0,
	}

	totalOriginalTokens := 0
	totalCompressedTokens := 0

	for _, contextFile := range selection.Files {
		// Read file content if not already loaded
		content := contextFile.Content
		if content == "" {
			// In a real implementation, you would load the content here
			// For now, we'll simulate with the file info
			content = fmt.Sprintf("// File: %s\n// Tokens: %d\n// Type: %s\n", 
				contextFile.FileInfo.Path, 
				contextFile.FileInfo.TokenCount,
				contextFile.FileInfo.FileType)
		}

		originalTokens := contextFile.FileInfo.TokenCount
		if originalTokens == 0 {
			if c.tokenCounter != nil {
				originalTokens, _ = c.tokenCounter.CountTokens(content)
			}
		}

		compressedContent, compressedTokens, _, err := c.compressFileContent(content, contextFile.FileInfo, strategy)
		if err != nil {
			// If compression fails, use original content
			compressedContent = content
			compressedTokens = originalTokens
		}

		compressedFile := CompressedFile{
			OriginalPath:     contextFile.FileInfo.Path,
			CompressedContent: compressedContent,
			OriginalTokens:   originalTokens,
			CompressedTokens: compressedTokens,
			CompressionRatio: 1.0,
			Method:           string(strategy),
		}

		if originalTokens > 0 {
			compressedFile.CompressionRatio = float64(compressedTokens) / float64(originalTokens)
		}

		compressed.CompressedFiles = append(compressed.CompressedFiles, compressedFile)
		
		totalOriginalTokens += originalTokens
		totalCompressedTokens += compressedTokens
	}

	// Calculate overall metrics
	if totalOriginalTokens > 0 {
		compressed.CompressionRatio = float64(totalCompressedTokens) / float64(totalOriginalTokens)
		compressed.TokenReduction = totalOriginalTokens - totalCompressedTokens
	}

	// Estimate quality impact
	compressed.QualityScore = c.estimateQualityImpact(strategy, compressed.CompressionRatio)
	compressed.CompressionTime = time.Since(startTime)

	return compressed, nil
}

// EstimateCompression estimates compression ratio without actually compressing
func (c *DefaultContextCompressor) EstimateCompression(selection *SelectedContext, strategy CompressionStrategy) (float64, error) {
	switch strategy {
	case CompressionNone:
		return 1.0, nil
	case CompressionSummary:
		return 0.3, nil // Summaries typically achieve 70% reduction
	case CompressionSnippet:
		return 0.4, nil // Snippets achieve ~60% reduction
	case CompressionMinify:
		return 0.8, nil // Minification achieves ~20% reduction
	case CompressionSemantic:
		return 0.5, nil // Semantic compression achieves ~50% reduction
	default:
		return 0.7, nil // Conservative estimate
	}
}

// GetCompressionStrategies returns available compression strategies
func (c *DefaultContextCompressor) GetCompressionStrategies() []CompressionStrategy {
	return []CompressionStrategy{
		CompressionNone,
		CompressionSummary,
		CompressionSnippet,
		CompressionMinify,
		CompressionSemantic,
	}
}

// compressFileContent compresses content of a single file
func (c *DefaultContextCompressor) compressFileContent(content string, fileInfo *FileInfo, strategy CompressionStrategy) (string, int, []string, error) {
	switch strategy {
	case CompressionNone:
		return content, fileInfo.TokenCount, []string{"none"}, nil
		
	case CompressionSummary:
		return c.createSummary(content, fileInfo)
		
	case CompressionSnippet:
		return c.extractSnippets(content, fileInfo)
		
	case CompressionMinify:
		return c.minifyContent(content, fileInfo)
		
	case CompressionSemantic:
		return c.semanticCompression(content, fileInfo)
		
	default:
		return content, fileInfo.TokenCount, []string{"unknown"}, fmt.Errorf("unknown compression strategy: %s", strategy)
	}
}

// createSummary creates a summary of the file content
func (c *DefaultContextCompressor) createSummary(content string, fileInfo *FileInfo) (string, int, []string, error) {
	var summary strings.Builder
	techniques := []string{"summary"}
	
	// File header with metadata
	summary.WriteString(fmt.Sprintf("// SUMMARY of %s (%s, %d tokens)\n", 
		fileInfo.Path, fileInfo.Language, fileInfo.TokenCount))
	
	// Extract key elements based on language
	switch fileInfo.Language {
	case "go":
		summary.WriteString(c.summarizeGoFile(content))
	case "javascript", "typescript":
		summary.WriteString(c.summarizeJSFile(content))
	case "python":
		summary.WriteString(c.summarizePythonFile(content))
	default:
		summary.WriteString(c.summarizeGenericFile(content))
	}
	
	summaryContent := summary.String()
	tokens := 0
	if c.tokenCounter != nil {
		tokens, _ = c.tokenCounter.CountTokens(summaryContent)
	}
	
	return summaryContent, tokens, techniques, nil
}

// extractSnippets extracts relevant code snippets
func (c *DefaultContextCompressor) extractSnippets(content string, fileInfo *FileInfo) (string, int, []string, error) {
	var result strings.Builder
	techniques := []string{"snippets"}
	
	result.WriteString(fmt.Sprintf("// SNIPPETS from %s\n", fileInfo.Path))
	
	lines := strings.Split(content, "\n")
	
	// Extract imports/includes
	if c.config.PreserveImports {
		for _, line := range lines {
			if c.isImportLine(line, fileInfo.Language) {
				result.WriteString(line + "\n")
			}
		}
		result.WriteString("\n")
	}
	
	// Extract function signatures and key structures
	
	for i, line := range lines {
		if c.isFunctionStart(line, fileInfo.Language) {
			
			// Include function signature and a few lines
			for j := i; j < len(lines) && j < i+c.config.MinFunctionLines+1; j++ {
				result.WriteString(lines[j] + "\n")
			}
			result.WriteString("    // ... function body truncated ...\n")
			
			// Find and include function end
			for j := i + c.config.MinFunctionLines + 1; j < len(lines); j++ {
				if c.isFunctionEnd(lines[j], fileInfo.Language) {
					result.WriteString(lines[j] + "\n")
					break
				}
			}
			result.WriteString("\n")
		}
	}
	
	snippetsContent := result.String()
	tokens := 0
	if c.tokenCounter != nil {
		tokens, _ = c.tokenCounter.CountTokens(snippetsContent)
	}
	
	return snippetsContent, tokens, techniques, nil
}

// minifyContent removes unnecessary whitespace and comments
func (c *DefaultContextCompressor) minifyContent(content string, fileInfo *FileInfo) (string, int, []string, error) {
	techniques := []string{"minify"}
	
	// Remove comments unless configured to preserve them
	if !c.config.PreserveComments {
		content = c.removeComments(content, fileInfo.Language)
		techniques = append(techniques, "remove_comments")
	}
	
	// Remove excessive whitespace
	content = c.removeExcessiveWhitespace(content)
	techniques = append(techniques, "remove_whitespace")
	
	// Remove empty lines
	lines := strings.Split(content, "\n")
	nonEmptyLines := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	content = strings.Join(nonEmptyLines, "\n")
	techniques = append(techniques, "remove_empty_lines")
	
	tokens := 0
	if c.tokenCounter != nil {
		tokens, _ = c.tokenCounter.CountTokens(content)
	}
	
	return content, tokens, techniques, nil
}

// semanticCompression applies semantic understanding for compression
func (c *DefaultContextCompressor) semanticCompression(content string, fileInfo *FileInfo) (string, int, []string, error) {
	var result strings.Builder
	techniques := []string{"semantic"}
	
	result.WriteString(fmt.Sprintf("// SEMANTIC COMPRESSION of %s\n", fileInfo.Path))
	
	// Extract package/module declaration
	if packageLine := c.extractPackageDeclaration(content, fileInfo.Language); packageLine != "" {
		result.WriteString(packageLine + "\n")
	}
	
	// Extract imports with grouping
	imports := c.extractAndGroupImports(content, fileInfo.Language)
	if len(imports) > 0 {
		result.WriteString("// Imports:\n")
		for _, imp := range imports {
			result.WriteString(imp + "\n")
		}
		result.WriteString("\n")
	}
	
	// Extract type definitions and interfaces
	types := c.extractTypeDefinitions(content, fileInfo.Language)
	if len(types) > 0 {
		result.WriteString("// Type Definitions:\n")
		for _, typedef := range types {
			result.WriteString(typedef + "\n")
		}
		result.WriteString("\n")
	}
	
	// Extract function signatures with condensed bodies
	functions := c.extractFunctionSignatures(content, fileInfo.Language)
	if len(functions) > 0 {
		result.WriteString("// Functions:\n")
		for _, function := range functions {
			result.WriteString(function + "\n")
		}
	}
	
	semanticContent := result.String()
	tokens := 0
	if c.tokenCounter != nil {
		tokens, _ = c.tokenCounter.CountTokens(semanticContent)
	}
	
	return semanticContent, tokens, techniques, nil
}

// Language-specific summary methods
func (c *DefaultContextCompressor) summarizeGoFile(content string) string {
	var summary strings.Builder
	
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Include package declaration
		if strings.HasPrefix(trimmed, "package ") {
			summary.WriteString(line + "\n")
		}
		
		// Include imports
		if strings.HasPrefix(trimmed, "import ") || (strings.HasPrefix(trimmed, "\"") && strings.Contains(line, "import")) {
			summary.WriteString(line + "\n")
		}
		
		// Include type definitions
		if strings.HasPrefix(trimmed, "type ") {
			summary.WriteString(line + "\n")
		}
		
		// Include function signatures
		if strings.HasPrefix(trimmed, "func ") {
			summary.WriteString(line + " { /* ... */ }\n")
		}
	}
	
	return summary.String()
}

func (c *DefaultContextCompressor) summarizeJSFile(content string) string {
	// Similar to Go but with JS syntax patterns
	var summary strings.Builder
	
	summary.WriteString("// JavaScript/TypeScript summary\n")
	
	// Extract class declarations, function declarations, exports, etc.
	// This is simplified - a full implementation would use proper AST parsing
	
	return summary.String()
}

func (c *DefaultContextCompressor) summarizePythonFile(content string) string {
	var summary strings.Builder
	
	summary.WriteString("# Python summary\n")
	
	// Extract imports, class definitions, function definitions
	
	return summary.String()
}

func (c *DefaultContextCompressor) summarizeGenericFile(content string) string {
	// Generic summary for unknown file types
	lines := strings.Split(content, "\n")
	
	summary := fmt.Sprintf("File with %d lines\n", len(lines))
	
	// Include first and last few lines
	if len(lines) > 10 {
		summary += "// First 3 lines:\n"
		for i := 0; i < 3 && i < len(lines); i++ {
			summary += lines[i] + "\n"
		}
		summary += "// ... content truncated ...\n"
		summary += "// Last 3 lines:\n"
		for i := len(lines) - 3; i < len(lines); i++ {
			if i >= 0 {
				summary += lines[i] + "\n"
			}
		}
	} else {
		summary += content
	}
	
	return summary
}

// Helper methods for content analysis
func (c *DefaultContextCompressor) isImportLine(line, language string) bool {
	trimmed := strings.TrimSpace(line)
	
	switch language {
	case "go":
		return strings.HasPrefix(trimmed, "import ") || 
			   (strings.HasPrefix(trimmed, "\"") && strings.Contains(line, "import"))
	case "javascript", "typescript":
		return strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "require(")
	case "python":
		return strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ")
	default:
		return strings.Contains(strings.ToLower(trimmed), "import") ||
			   strings.Contains(strings.ToLower(trimmed), "include")
	}
}

func (c *DefaultContextCompressor) isFunctionStart(line, language string) bool {
	trimmed := strings.TrimSpace(line)
	
	switch language {
	case "go":
		return strings.HasPrefix(trimmed, "func ")
	case "javascript", "typescript":
		return strings.Contains(trimmed, "function ") || 
			   (strings.Contains(trimmed, "=>") && strings.Contains(trimmed, "="))
	case "python":
		return strings.HasPrefix(trimmed, "def ")
	default:
		return false
	}
}

func (c *DefaultContextCompressor) isFunctionEnd(line, language string) bool {
	trimmed := strings.TrimSpace(line)
	
	switch language {
	case "go":
		return trimmed == "}"
	case "javascript", "typescript":
		return trimmed == "}" || trimmed == "};"
	case "python":
		// Python doesn't have explicit function end markers
		return !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != ""
	default:
		return trimmed == "}"
	}
}

func (c *DefaultContextCompressor) removeComments(content, language string) string {
	switch language {
	case "go", "javascript", "typescript":
		// Remove // comments
		lines := strings.Split(content, "\n")
		filtered := []string{}
		for _, line := range lines {
			if idx := strings.Index(line, "//"); idx != -1 {
				line = line[:idx]
			}
			filtered = append(filtered, line)
		}
		content = strings.Join(filtered, "\n")
		
		// Remove /* */ comments (simplified)
		re := regexp.MustCompile(`/\*.*?\*/`)
		content = re.ReplaceAllString(content, "")
		
	case "python":
		// Remove # comments
		lines := strings.Split(content, "\n")
		filtered := []string{}
		for _, line := range lines {
			if idx := strings.Index(line, "#"); idx != -1 {
				line = line[:idx]
			}
			filtered = append(filtered, line)
		}
		content = strings.Join(filtered, "\n")
	}
	
	return content
}

func (c *DefaultContextCompressor) removeExcessiveWhitespace(content string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(` +`)
	content = re.ReplaceAllString(content, " ")
	
	// Replace multiple newlines with single newline
	re = regexp.MustCompile(`\n+`)
	content = re.ReplaceAllString(content, "\n")
	
	return content
}

func (c *DefaultContextCompressor) extractPackageDeclaration(content, language string) string {
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch language {
		case "go":
			if strings.HasPrefix(trimmed, "package ") {
				return line
			}
		}
	}
	
	return ""
}

func (c *DefaultContextCompressor) extractAndGroupImports(content, language string) []string {
	imports := []string{}
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		if c.isImportLine(line, language) {
			imports = append(imports, strings.TrimSpace(line))
		}
	}
	
	return imports
}

func (c *DefaultContextCompressor) extractTypeDefinitions(content, language string) []string {
	types := []string{}
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch language {
		case "go":
			if strings.HasPrefix(trimmed, "type ") {
				types = append(types, line)
			}
		}
	}
	
	return types
}

func (c *DefaultContextCompressor) extractFunctionSignatures(content, language string) []string {
	functions := []string{}
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		if c.isFunctionStart(line, language) {
			functions = append(functions, line+" { /* ... */ }")
		}
	}
	
	return functions
}

func (c *DefaultContextCompressor) estimateQualityImpact(strategy CompressionStrategy, ratio float64) float64 {
	// Estimate quality impact based on compression strategy and ratio
	
	switch strategy {
	case CompressionNone:
		return 1.0
	case CompressionMinify:
		return 0.95 // Minimal quality impact
	case CompressionSnippet:
		return 0.8 - (1.0-ratio)*0.3 // Quality decreases with higher compression
	case CompressionSummary:
		return 0.6 - (1.0-ratio)*0.2
	case CompressionSemantic:
		return 0.75 - (1.0-ratio)*0.25
	default:
		return 0.7
	}
}

func getDefaultLanguageRules() map[string]*LanguageCompressionRules {
	return map[string]*LanguageCompressionRules{
		"go": {
			ImportPatterns:    []string{`^import\s+`, `^\s*".*"$`},
			FunctionPatterns:  []string{`^func\s+`},
			CommentPatterns:   []string{`//.*$`, `/\*.*?\*/`},
			PreservePatterns:  []string{`^package\s+`, `^type\s+`, `^var\s+`, `^const\s+`},
			RemovablePatterns: []string{`^\s*$`, `^\s*//`},
		},
		"javascript": {
			ImportPatterns:    []string{`^import\s+`, `^require\(`},
			FunctionPatterns:  []string{`function\s+`, `=>\s*{`, `^.*function`},
			CommentPatterns:   []string{`//.*$`, `/\*.*?\*/`},
			PreservePatterns:  []string{`^class\s+`, `^export\s+`, `^const\s+`, `^let\s+`, `^var\s+`},
			RemovablePatterns: []string{`^\s*$`, `^\s*//`, `^\s*/\*`},
		},
	}
}