package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	contextpkg "github.com/rcliao/teeny-orb/internal/context"
	"github.com/rcliao/teeny-orb/internal/mcp"
)

// ContextAnalysisHandler implements MCP tool for project context analysis
type ContextAnalysisHandler struct {
	analyzer contextpkg.ContextAnalyzer
}

// NewContextAnalysisHandler creates a new context analysis MCP tool handler
func NewContextAnalysisHandler(analyzer contextpkg.ContextAnalyzer) *ContextAnalysisHandler {
	return &ContextAnalysisHandler{
		analyzer: analyzer,
	}
}

// Name returns the tool name
func (h *ContextAnalysisHandler) Name() string {
	return "analyze_context"
}

// Description returns the tool description
func (h *ContextAnalysisHandler) Description() string {
	return "Analyzes project context to measure token usage, file dependencies, and project structure for intelligent context optimization"
}

// InputSchema returns the tool input schema
func (h *ContextAnalysisHandler) InputSchema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"project_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the project root directory to analyze",
			},
			"include_dependencies": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to build dependency graph analysis",
				"default":     true,
			},
			"max_file_size": map[string]interface{}{
				"type":        "number",
				"description": "Maximum file size in bytes to include in analysis",
				"default":     1048576, // 1MB
			},
		},
		Required: []string{"project_path"},
	}
}

// Handle executes the context analysis tool
func (h *ContextAnalysisHandler) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	projectPath, ok := arguments["project_path"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: project_path is required and must be a string",
			}},
			IsError: true,
		}, nil
	}

	// Make path absolute
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: invalid project path: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Perform analysis
	projectContext, err := h.analyzer.AnalyzeProject(ctx, absPath)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error analyzing project: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Format response
	analysisText := h.formatAnalysisResults(projectContext)
	analysisJSON, _ := json.MarshalIndent(projectContext, "", "  ")

	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: analysisText,
			},
			{
				Type:     "text",
				Text:     string(analysisJSON),
				MimeType: "application/json",
			},
		},
	}, nil
}

func (h *ContextAnalysisHandler) formatAnalysisResults(projectCtx *contextpkg.ProjectContext) string {
	var result strings.Builder
	
	result.WriteString("# Project Context Analysis\n\n")
	result.WriteString(fmt.Sprintf("**Project Path:** %s\n", projectCtx.RootPath))
	result.WriteString(fmt.Sprintf("**Analysis Date:** %s\n\n", projectCtx.CreatedAt.Format("2006-01-02 15:04:05")))
	
	result.WriteString("## Summary Statistics\n")
	result.WriteString(fmt.Sprintf("- **Total Files:** %d\n", projectCtx.TotalFiles))
	result.WriteString(fmt.Sprintf("- **Total Tokens:** %d\n", projectCtx.TotalTokens))
	if projectCtx.TotalFiles > 0 {
		avgTokens := projectCtx.TotalTokens / projectCtx.TotalFiles
		result.WriteString(fmt.Sprintf("- **Average Tokens per File:** %d\n", avgTokens))
	}
	result.WriteString("\n")
	
	// Language breakdown
	if len(projectCtx.Languages) > 0 {
		result.WriteString("## Language Distribution\n")
		for lang, count := range projectCtx.Languages {
			result.WriteString(fmt.Sprintf("- **%s:** %d files\n", lang, count))
		}
		result.WriteString("\n")
	}
	
	// Analysis insights
	if projectCtx.Analysis != nil {
		result.WriteString("## Project Structure\n")
		
		if len(projectCtx.Analysis.EntryPoints) > 0 {
			result.WriteString("**Entry Points:**\n")
			for _, entry := range projectCtx.Analysis.EntryPoints {
				result.WriteString(fmt.Sprintf("- %s\n", entry))
			}
			result.WriteString("\n")
		}
		
		if len(projectCtx.Analysis.TestFiles) > 0 {
			result.WriteString(fmt.Sprintf("**Test Files:** %d\n", len(projectCtx.Analysis.TestFiles)))
		}
		
		if len(projectCtx.Analysis.ConfigFiles) > 0 {
			result.WriteString(fmt.Sprintf("**Configuration Files:** %d\n", len(projectCtx.Analysis.ConfigFiles)))
		}
		
		if len(projectCtx.Analysis.Recommendations) > 0 {
			result.WriteString("\n## Recommendations\n")
			for _, rec := range projectCtx.Analysis.Recommendations {
				result.WriteString(fmt.Sprintf("- %s\n", rec))
			}
		}
	}
	
	return result.String()
}

// ContextOptimizationHandler implements MCP tool for context optimization
type ContextOptimizationHandler struct {
	optimizer contextpkg.ContextOptimizer
	analyzer  contextpkg.ContextAnalyzer
}

// NewContextOptimizationHandler creates a new context optimization MCP tool handler
func NewContextOptimizationHandler(optimizer contextpkg.ContextOptimizer, analyzer contextpkg.ContextAnalyzer) *ContextOptimizationHandler {
	return &ContextOptimizationHandler{
		optimizer: optimizer,
		analyzer:  analyzer,
	}
}

// Name returns the tool name
func (h *ContextOptimizationHandler) Name() string {
	return "optimize_context"
}

// Description returns the tool description
func (h *ContextOptimizationHandler) Description() string {
	return "Optimizes project context for a specific task by intelligently selecting relevant files within token budget constraints"
}

// InputSchema returns the tool input schema
func (h *ContextOptimizationHandler) InputSchema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"project_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the project root directory",
			},
			"task_description": map[string]interface{}{
				"type":        "string",
				"description": "Description of the coding task for context optimization",
			},
			"task_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of task: general, debug, refactor, feature, test, documentation",
				"enum":        []string{"general", "debug", "refactor", "feature", "test", "documentation"},
				"default":     "general",
			},
			"token_budget": map[string]interface{}{
				"type":        "number",
				"description": "Maximum number of tokens to include in optimized context",
				"default":     8000,
			},
			"include_tests": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include test files in the context",
				"default":     false,
			},
			"include_docs": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include documentation files in the context",
				"default":     false,
			},
			"strategy": map[string]interface{}{
				"type":        "string",
				"description": "Context selection strategy",
				"enum":        []string{"relevance", "dependency", "freshness", "compactness", "balanced"},
				"default":     "balanced",
			},
		},
		Required: []string{"project_path", "task_description"},
	}
}

// Handle executes the context optimization tool
func (h *ContextOptimizationHandler) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	projectPath, ok := arguments["project_path"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: project_path is required and must be a string",
			}},
			IsError: true,
		}, nil
	}

	taskDescription, ok := arguments["task_description"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: task_description is required and must be a string",
			}},
			IsError: true,
		}, nil
	}

	// Parse optional parameters
	tokenBudget := 8000
	if budget, ok := arguments["token_budget"]; ok {
		if budgetFloat, ok := budget.(float64); ok {
			tokenBudget = int(budgetFloat)
		}
	}

	taskTypeStr := "general"
	if tt, ok := arguments["task_type"].(string); ok {
		taskTypeStr = tt
	}

	includeTests := false
	if it, ok := arguments["include_tests"]; ok {
		if itBool, ok := it.(bool); ok {
			includeTests = itBool
		}
	}

	includeDocs := false
	if id, ok := arguments["include_docs"]; ok {
		if idBool, ok := id.(bool); ok {
			includeDocs = idBool
		}
	}

	strategy := "balanced"
	if s, ok := arguments["strategy"].(string); ok {
		strategy = s
	}

	// Make path absolute
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error: invalid project path: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Analyze project first
	projectContext, err := h.analyzer.AnalyzeProject(ctx, absPath)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error analyzing project: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Create task
	task := &contextpkg.Task{
		Type:        contextpkg.TaskType(taskTypeStr),
		Description: taskDescription,
		Priority:    contextpkg.PriorityMedium,
		Scope:       contextpkg.ScopeProject,
	}

	// Create constraints
	constraints := &contextpkg.ContextConstraints{
		MaxTokens:         tokenBudget,
		MaxFiles:          50,
		MinRelevanceScore: 0.1,
		PreferredTypes:    []string{"source", "configuration"},
		IncludeTests:      includeTests,
		IncludeDocs:       includeDocs,
		FreshnessBias:     0.2,
		DependencyDepth:   2,
		Strategy:          contextpkg.SelectionStrategy(strategy),
	}

	// Optimize context
	selectedContext, err := h.optimizer.SelectOptimalContext(ctx, projectContext, task, constraints)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error optimizing context: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Format response
	optimizationText := h.formatOptimizationResults(selectedContext, projectContext)
	optimizationJSON, _ := json.MarshalIndent(selectedContext, "", "  ")

	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: optimizationText,
			},
			{
				Type:     "text",
				Text:     string(optimizationJSON),
				MimeType: "application/json",
			},
		},
	}, nil
}

func (h *ContextOptimizationHandler) formatOptimizationResults(selectedCtx *contextpkg.SelectedContext, projectCtx *contextpkg.ProjectContext) string {
	var result strings.Builder
	
	result.WriteString("# Context Optimization Results\n\n")
	result.WriteString(fmt.Sprintf("**Task:** %s\n", selectedCtx.Task.Description))
	result.WriteString(fmt.Sprintf("**Task Type:** %s\n", selectedCtx.Task.Type))
	result.WriteString(fmt.Sprintf("**Strategy:** %s\n", selectedCtx.Strategy))
	result.WriteString(fmt.Sprintf("**Selection Time:** %v\n\n", selectedCtx.SelectionTime))
	
	result.WriteString("## Optimization Summary\n")
	result.WriteString(fmt.Sprintf("- **Files Selected:** %d / %d (%.1f%%)\n", 
		selectedCtx.TotalFiles, 
		projectCtx.TotalFiles,
		float64(selectedCtx.TotalFiles)/float64(projectCtx.TotalFiles)*100))
	result.WriteString(fmt.Sprintf("- **Tokens Selected:** %d / %d (%.1f%%)\n", 
		selectedCtx.TotalTokens, 
		projectCtx.TotalTokens,
		float64(selectedCtx.TotalTokens)/float64(projectCtx.TotalTokens)*100))
	result.WriteString(fmt.Sprintf("- **Selection Score:** %.3f\n", selectedCtx.SelectionScore))
	result.WriteString(fmt.Sprintf("- **Token Budget:** %d\n", selectedCtx.Constraints.MaxTokens))
	
	tokenReduction := float64(projectCtx.TotalTokens-selectedCtx.TotalTokens) / float64(projectCtx.TotalTokens) * 100
	result.WriteString(fmt.Sprintf("- **Token Reduction:** %.1f%%\n\n", tokenReduction))
	
	result.WriteString("## Selected Files\n")
	for i, file := range selectedCtx.Files {
		result.WriteString(fmt.Sprintf("%d. **%s** (%.3f relevance, %d tokens)\n", 
			i+1, 
			file.FileInfo.Path, 
			file.RelevanceScore, 
			file.FileInfo.TokenCount))
		if file.InclusionReason != "" {
			result.WriteString(fmt.Sprintf("   - Reason: %s\n", file.InclusionReason))
		}
	}
	
	return result.String()
}

// TokenCountHandler implements MCP tool for token counting
type TokenCountHandler struct {
	analyzer contextpkg.ContextAnalyzer
}

// NewTokenCountHandler creates a new token counting MCP tool handler
func NewTokenCountHandler(analyzer contextpkg.ContextAnalyzer) *TokenCountHandler {
	return &TokenCountHandler{
		analyzer: analyzer,
	}
}

// Name returns the tool name
func (h *TokenCountHandler) Name() string {
	return "count_tokens"
}

// Description returns the tool description
func (h *TokenCountHandler) Description() string {
	return "Counts tokens in text content or files for context optimization planning"
}

// InputSchema returns the tool input schema
func (h *TokenCountHandler) InputSchema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Text content to count tokens for",
			},
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to file to count tokens for",
			},
		},
	}
}

// Handle executes the token counting tool
func (h *TokenCountHandler) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	var tokenCount int
	var err error
	var source string

	if content, ok := arguments["content"].(string); ok {
		tokenCount, err = h.analyzer.CountTokens(content)
		source = "provided content"
	} else if filePath, ok := arguments["file_path"].(string); ok {
		fileInfo, fileErr := h.analyzer.GetFileInfo(ctx, filePath)
		if fileErr != nil {
			return &mcp.CallToolResponse{
				Content: []mcp.Content{{
					Type: "text",
					Text: fmt.Sprintf("Error reading file: %v", fileErr),
				}},
				IsError: true,
			}, nil
		}
		tokenCount = fileInfo.TokenCount
		source = filePath
	} else {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: "Error: either 'content' or 'file_path' must be provided",
			}},
			IsError: true,
		}, nil
	}

	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{{
				Type: "text",
				Text: fmt.Sprintf("Error counting tokens: %v", err),
			}},
			IsError: true,
		}, nil
	}

	resultText := fmt.Sprintf("Token count for %s: %d tokens", source, tokenCount)
	resultJSON, _ := json.Marshal(map[string]interface{}{
		"source":      source,
		"token_count": tokenCount,
	})

	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: resultText,
			},
			{
				Type:     "text",
				Text:     string(resultJSON),
				MimeType: "application/json",
			},
		},
	}, nil
}