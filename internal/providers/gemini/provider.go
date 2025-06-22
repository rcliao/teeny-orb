package gemini

import (
	"context"
	"fmt"

	"github.com/rcliao/teeny-orb/internal/providers"
)

// GeminiToolProvider integrates Gemini with tool calling through MCP or direct
type GeminiToolProvider struct {
	client       *GeminiClient
	toolProvider providers.ToolProvider
	mode         string // "direct" or "mcp"
}

// NewGeminiToolProvider creates a new Gemini tool provider
func NewGeminiToolProvider(apiKey, model, mode string, toolProvider providers.ToolProvider) *GeminiToolProvider {
	client := NewGeminiClient(apiKey, model)
	client.SetToolProvider(toolProvider)
	
	return &GeminiToolProvider{
		client:       client,
		toolProvider: toolProvider,
		mode:         mode,
	}
}

// ChatWithTools performs a chat request with tool calling capability
func (g *GeminiToolProvider) ChatWithTools(ctx context.Context, messages []providers.Message) (*providers.ChatResponse, error) {
	// Get available tools
	tools := g.toolProvider.ListTools()
	
	// Convert to tool definitions
	toolDefs := make([]providers.ToolDefinition, len(tools))
	for i, tool := range tools {
		toolDefs[i] = providers.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  g.generateToolSchema(tool),
		}
	}
	
	// Create chat request
	request := &providers.ChatRequest{
		Messages: messages,
		Tools:    toolDefs,
		Model:    g.client.model,
	}
	
	// Make initial request
	response, err := g.client.Chat(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("Gemini chat failed: %w", err)
	}
	
	// If no tool calls, return response directly
	if len(response.ToolCalls) == 0 {
		return response, nil
	}
	
	// Execute tool calls
	toolResults := make([]string, 0, len(response.ToolCalls))
	for _, toolCall := range response.ToolCalls {
		result, err := g.toolProvider.CallTool(ctx, toolCall.Name, toolCall.Arguments)
		if err != nil {
			toolResults = append(toolResults, fmt.Sprintf("Error calling %s: %v", toolCall.Name, err))
		} else if !result.Success {
			toolResults = append(toolResults, fmt.Sprintf("Tool %s failed: %s", toolCall.Name, result.Error))
		} else {
			toolResults = append(toolResults, result.Output)
		}
	}
	
	// Create follow-up message with tool results
	toolResultMessage := "Tool execution results:\n"
	for i, result := range toolResults {
		toolResultMessage += fmt.Sprintf("%d. %s\n", i+1, result)
	}
	
	// Make follow-up request to get final response
	followUpMessages := append(messages, 
		providers.Message{Role: "assistant", Content: response.Content},
		providers.Message{Role: "user", Content: toolResultMessage},
	)
	
	followUpRequest := &providers.ChatRequest{
		Messages: followUpMessages,
		Model:    g.client.model,
	}
	
	finalResponse, err := g.client.Chat(ctx, followUpRequest)
	if err != nil {
		return nil, fmt.Errorf("Gemini follow-up failed: %w", err)
	}
	
	// Combine responses
	combinedResponse := &providers.ChatResponse{
		Content: response.Content + "\n\n" + finalResponse.Content,
		Usage: providers.Usage{
			PromptTokens:     response.Usage.PromptTokens + finalResponse.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens + finalResponse.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens + finalResponse.Usage.TotalTokens,
		},
		Model: finalResponse.Model,
	}
	
	return combinedResponse, nil
}

// generateToolSchema creates a JSON schema for a tool
func (g *GeminiToolProvider) generateToolSchema(tool providers.Tool) map[string]interface{} {
	// Basic schema generation based on tool type
	switch tool.Name() {
	case "filesystem":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type": "string",
					"enum": []string{"read", "write", "list"},
					"description": "The file system operation to perform",
				},
				"path": map[string]interface{}{
					"type": "string",
					"description": "The file or directory path",
				},
				"content": map[string]interface{}{
					"type": "string",
					"description": "Content to write (for write operation)",
				},
			},
			"required": []string{"operation"},
		}
	case "command":
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type": "string",
					"description": "The command to execute",
				},
			},
			"required": []string{"command"},
		}
	default:
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		}
	}
}

// GetClient returns the underlying Gemini client
func (g *GeminiToolProvider) GetClient() *GeminiClient {
	return g.client
}

// GetMode returns the tool calling mode
func (g *GeminiToolProvider) GetMode() string {
	return g.mode
}

// Close performs cleanup
func (g *GeminiToolProvider) Close() error {
	if g.toolProvider != nil {
		return g.toolProvider.Close()
	}
	return nil
}