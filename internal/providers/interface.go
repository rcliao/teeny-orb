package providers

import (
	"context"
	"fmt"
	"io"
)

// Tool represents a single tool that can be called by an AI provider
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Output  string                 `json:"output,omitempty"`
}

// ToolProvider defines the interface for different tool calling implementations
type ToolProvider interface {
	// RegisterTool registers a tool with the provider
	RegisterTool(tool Tool) error
	
	// ListTools returns all registered tools
	ListTools() []Tool
	
	// CallTool executes a tool by name with the given arguments
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error)
	
	// Close performs cleanup
	Close() error
}

// AIProvider defines the interface for AI service providers
type AIProvider interface {
	// Chat sends a chat request and returns a response
	Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)
	
	// ChatStream sends a chat request and returns a streaming response
	ChatStream(ctx context.Context, request *ChatRequest) (<-chan *StreamChunk, error)
	
	// CountTokens estimates token count for the given text
	CountTokens(text string) (int, error)
	
	// GetModel returns information about the current model
	GetModel() *ModelInfo
}

// ChatRequest represents a chat request to an AI provider
type ChatRequest struct {
	Messages []Message              `json:"messages"`
	Tools    []ToolDefinition       `json:"tools,omitempty"`
	Model    string                 `json:"model,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ChatResponse represents a response from an AI provider
type ChatResponse struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     Usage      `json:"usage"`
	Model     string     `json:"model"`
}

// StreamChunk represents a chunk in a streaming response
type StreamChunk struct {
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
	Error     error      `json:"error,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// ToolDefinition describes a tool for the AI provider
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a tool call from the AI
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Usage tracks token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelInfo contains information about a model
type ModelInfo struct {
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	MaxTokens    int    `json:"max_tokens"`
	SupportsTools bool   `json:"supports_tools"`
}

// FileSystemTool provides basic file system operations
type FileSystemTool struct {
	baseDir string
}

// NewFileSystemTool creates a new file system tool
func NewFileSystemTool(baseDir string) *FileSystemTool {
	return &FileSystemTool{baseDir: baseDir}
}

// Name returns the tool name
func (f *FileSystemTool) Name() string {
	return "filesystem"
}

// Description returns the tool description
func (f *FileSystemTool) Description() string {
	return "Provides file system operations like read, write, and list files"
}

// Execute performs the file system operation
func (f *FileSystemTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "operation parameter is required",
		}, nil
	}

	switch operation {
	case "read":
		return f.readFile(args)
	case "write":
		return f.writeFile(args)
	case "list":
		return f.listFiles(args)
	default:
		return &ToolResult{
			Success: false,
			Error:   "unsupported operation: " + operation,
		}, nil
	}
}

func (f *FileSystemTool) readFile(args map[string]interface{}) (*ToolResult, error) {
	path, ok := args["path"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required for read operation",
		}, nil
	}

	// In a real implementation, this would read the file
	// For now, return a placeholder
	return &ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":    path,
			"content": "// File content would be here",
		},
		Output: fmt.Sprintf("Successfully read file: %s", path),
	}, nil
}

func (f *FileSystemTool) writeFile(args map[string]interface{}) (*ToolResult, error) {
	path, ok := args["path"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required for write operation",
		}, nil
	}

	content, ok := args["content"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "content parameter is required for write operation",
		}, nil
	}

	// In a real implementation, this would write the file
	return &ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":    path,
			"size":    len(content),
		},
		Output: fmt.Sprintf("Successfully wrote %d bytes to: %s", len(content), path),
	}, nil
}

func (f *FileSystemTool) listFiles(args map[string]interface{}) (*ToolResult, error) {
	path, ok := args["path"].(string)
	if !ok {
		path = "."
	}

	// In a real implementation, this would list the directory
	return &ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"path":  path,
			"files": []string{"main.go", "README.md", "Makefile"},
		},
		Output: fmt.Sprintf("Successfully listed directory: %s", path),
	}, nil
}

// CommandTool provides command execution capabilities
type CommandTool struct {
	allowedCommands []string
}

// NewCommandTool creates a new command execution tool
func NewCommandTool(allowedCommands []string) *CommandTool {
	return &CommandTool{allowedCommands: allowedCommands}
}

// Name returns the tool name
func (c *CommandTool) Name() string {
	return "command"
}

// Description returns the tool description
func (c *CommandTool) Description() string {
	return "Executes allowed shell commands"
}

// Execute runs the specified command
func (c *CommandTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "command parameter is required",
		}, nil
	}

	// Check if command is allowed
	allowed := false
	for _, allowedCmd := range c.allowedCommands {
		if command == allowedCmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("command not allowed: %s", command),
		}, nil
	}

	// In a real implementation, this would execute the command
	return &ToolResult{
		Success: true,
		Data: map[string]interface{}{
			"command":   command,
			"exit_code": 0,
		},
		Output: fmt.Sprintf("Successfully executed: %s", command),
	}, nil
}

// StreamWriter provides a writer interface for streaming output
type StreamWriter interface {
	io.Writer
	Flush() error
}