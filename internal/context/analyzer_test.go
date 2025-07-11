package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestSimpleTokenCounter tests basic token counting functionality
func TestSimpleTokenCounter(t *testing.T) {
	counter := NewSimpleTokenCounter()
	
	tests := []struct {
		name     string
		content  string
		expected int
		delta    int // Allow some variance
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: 0,
			delta:    0,
		},
		{
			name:     "Simple text",
			content:  "Hello world",
			expected: 2,
			delta:    1,
		},
		{
			name:     "Go code",
			content:  "func main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			expected: 25, // More accurate expectation
			delta:    5,
		},
		{
			name:     "Complex code with symbols",
			content:  "if x := getValue(); x > 0 && x < 100 { return x * 2 }",
			expected: 35, // More accurate expectation
			delta:    5,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := counter.CountTokens(tt.content)
			if err != nil {
				t.Fatalf("CountTokens failed: %v", err)
			}
			
			if tokens < tt.expected-tt.delta || tokens > tt.expected+tt.delta {
				t.Errorf("CountTokens() = %d, expected %d ± %d", tokens, tt.expected, tt.delta)
			}
		})
	}
}

// TestDefaultAnalyzer tests project analysis functionality
func TestDefaultAnalyzer(t *testing.T) {
	// Create a temporary test project
	tmpDir, err := os.MkdirTemp("", "test-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create test files
	testFiles := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`,
		"utils.go": `package main

func add(a, b int) int {
	return a + b
}`,
		"README.md": `# Test Project

This is a test project for context analysis.`,
		"config.yaml": `name: test
version: 1.0.0`,
	}
	
	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}
	
	// Create analyzer
	tokenCounter := NewSimpleTokenCounter()
	analyzer := NewDefaultAnalyzer(tokenCounter, nil)
	
	// Analyze project
	ctx := context.Background()
	projectCtx, err := analyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}
	
	// Verify results
	if projectCtx.TotalFiles != 4 {
		t.Errorf("Expected 4 files, got %d", projectCtx.TotalFiles)
	}
	
	if projectCtx.TotalTokens == 0 {
		t.Errorf("Expected non-zero total tokens")
	}
	
	// Check language detection
	if projectCtx.Languages["go"] != 2 {
		t.Errorf("Expected 2 Go files, got %d", projectCtx.Languages["go"])
	}
	
	if projectCtx.Languages["markdown"] != 1 {
		t.Errorf("Expected 1 Markdown file, got %d", projectCtx.Languages["markdown"])
	}
	
	if projectCtx.Languages["yaml"] != 1 {
		t.Errorf("Expected 1 YAML file, got %d", projectCtx.Languages["yaml"])
	}
	
	// Check analysis
	if projectCtx.Analysis == nil {
		t.Fatalf("Expected analysis to be populated")
	}
	
	// Verify entry points detection
	foundMain := false
	for _, entry := range projectCtx.Analysis.EntryPoints {
		if filepath.Base(entry) == "main.go" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Errorf("Expected main.go to be detected as entry point")
	}
}

// TestFileRelevanceScoring tests file relevance scoring
func TestFileRelevanceScoring(t *testing.T) {
	tokenCounter := NewSimpleTokenCounter()
	analyzer := NewDefaultAnalyzer(tokenCounter, nil)
	
	file := &FileInfo{
		Path:     "internal/auth/handler.go",
		FileType: "source",
		Language: "go",
	}
	
	tests := []struct {
		taskType    TaskType
		description string
		minScore    float64
	}{
		{
			taskType:    TaskTypeFeature,
			description: "Add authentication to handler",
			minScore:    0.3,
		},
		{
			taskType:    TaskTypeDebug,
			description: "Fix memory leak",
			minScore:    0.3,
		},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.taskType), func(t *testing.T) {
			score := analyzer.ScoreFileRelevance(file, tt.taskType, tt.description)
			if score < tt.minScore {
				t.Errorf("Expected score >= %f, got %f", tt.minScore, score)
			}
		})
	}
}