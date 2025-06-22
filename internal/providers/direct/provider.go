package direct

import (
	"context"
	"fmt"
	"sync"

	"github.com/rcliao/teeny-orb/internal/providers"
)

// DirectToolProvider implements direct tool calling without protocol overhead
type DirectToolProvider struct {
	tools map[string]providers.Tool
	mutex sync.RWMutex
}

// NewDirectToolProvider creates a new direct tool provider
func NewDirectToolProvider() *DirectToolProvider {
	return &DirectToolProvider{
		tools: make(map[string]providers.Tool),
	}
}

// RegisterTool registers a tool with the provider
func (d *DirectToolProvider) RegisterTool(tool providers.Tool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	name := tool.Name()
	if _, exists := d.tools[name]; exists {
		return fmt.Errorf("tool already registered: %s", name)
	}
	
	d.tools[name] = tool
	return nil
}

// ListTools returns all registered tools
func (d *DirectToolProvider) ListTools() []providers.Tool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	tools := make([]providers.Tool, 0, len(d.tools))
	for _, tool := range d.tools {
		tools = append(tools, tool)
	}
	return tools
}

// CallTool executes a tool directly by name
func (d *DirectToolProvider) CallTool(ctx context.Context, name string, args map[string]interface{}) (*providers.ToolResult, error) {
	d.mutex.RLock()
	tool, exists := d.tools[name]
	d.mutex.RUnlock()
	
	if !exists {
		return &providers.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool not found: %s", name),
		}, nil
	}
	
	// Direct execution - no protocol overhead
	return tool.Execute(ctx, args)
}

// Close performs cleanup
func (d *DirectToolProvider) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	// Clear all tools
	d.tools = make(map[string]providers.Tool)
	return nil
}

// GetToolDefinitions returns tool definitions for AI providers
func (d *DirectToolProvider) GetToolDefinitions() []providers.ToolDefinition {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	definitions := make([]providers.ToolDefinition, 0, len(d.tools))
	for _, tool := range d.tools {
		definitions = append(definitions, providers.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  generateParameters(tool),
		})
	}
	return definitions
}

// generateParameters creates parameter schema for a tool
func generateParameters(tool providers.Tool) map[string]interface{} {
	// Basic parameter schema - in a real implementation, this would be more sophisticated
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

// DirectAIProvider implements a simple AI provider for testing
type DirectAIProvider struct {
	model    string
	provider *DirectToolProvider
}

// NewDirectAIProvider creates a new direct AI provider
func NewDirectAIProvider(model string, toolProvider *DirectToolProvider) *DirectAIProvider {
	return &DirectAIProvider{
		model:    model,
		provider: toolProvider,
	}
}

// Chat implements a simple chat interface
func (d *DirectAIProvider) Chat(ctx context.Context, request *providers.ChatRequest) (*providers.ChatResponse, error) {
	// Simple implementation for testing - just echo the request with tool information
	toolCount := len(d.provider.ListTools())
	
	response := &providers.ChatResponse{
		Content: fmt.Sprintf("Received message with %d available tools. Direct implementation ready.", toolCount),
		Model:   d.model,
		Usage: providers.Usage{
			PromptTokens:     len(request.Messages) * 10, // Rough estimate
			CompletionTokens: 20,
			TotalTokens:      len(request.Messages)*10 + 20,
		},
	}
	
	// If tools are available, demonstrate tool calling
	if len(request.Tools) > 0 {
		response.ToolCalls = []providers.ToolCall{
			{
				ID:   "call_1",
				Name: "filesystem",
				Arguments: map[string]interface{}{
					"operation": "list",
					"path":      ".",
				},
			},
		}
	}
	
	return response, nil
}

// ChatStream implements streaming (simplified for testing)
func (d *DirectAIProvider) ChatStream(ctx context.Context, request *providers.ChatRequest) (<-chan *providers.StreamChunk, error) {
	ch := make(chan *providers.StreamChunk, 2)
	
	go func() {
		defer close(ch)
		
		ch <- &providers.StreamChunk{
			Content: "Direct streaming response",
			Done:    false,
		}
		
		ch <- &providers.StreamChunk{
			Content: " - implementation ready!",
			Done:    true,
		}
	}()
	
	return ch, nil
}

// CountTokens provides a simple token counting implementation
func (d *DirectAIProvider) CountTokens(text string) (int, error) {
	// Simple word-based approximation: ~1.3 tokens per word
	wordCount := len(text) / 5 // Rough word count
	return int(float64(wordCount) * 1.3), nil
}

// GetModel returns model information
func (d *DirectAIProvider) GetModel() *providers.ModelInfo {
	return &providers.ModelInfo{
		Name:         d.model,
		Provider:     "direct",
		MaxTokens:    4096,
		SupportsTools: true,
	}
}