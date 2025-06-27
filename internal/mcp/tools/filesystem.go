package tools

import (
	"context"
	"fmt"

	"github.com/rcliao/teeny-orb/internal/mcp"
)

// FileSystemTool provides MCP-compatible file system operations
type FileSystemTool struct {
	baseDir string
}

// NewFileSystemTool creates a new MCP filesystem tool
func NewFileSystemTool(baseDir string) *FileSystemTool {
	return &FileSystemTool{baseDir: baseDir}
}

// Name returns the tool name
func (f *FileSystemTool) Name() string {
	return "filesystem"
}

// Description returns the tool description
func (f *FileSystemTool) Description() string {
	return "Provides secure file system operations including read, write, and list operations"
}

// InputSchema returns the JSON schema for tool inputs
func (f *FileSystemTool) InputSchema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"read", "write", "list"},
				"description": "The file system operation to perform",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The file or directory path relative to the workspace",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write (required for write operation)",
			},
		},
		Required: []string{"operation"},
	}
}

// Handle executes the filesystem operation
func (f *FileSystemTool) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	operation, ok := arguments["operation"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Error: operation parameter is required and must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	switch operation {
	case "read":
		return f.handleRead(arguments)
	case "write":
		return f.handleWrite(arguments)
	case "list":
		return f.handleList(arguments)
	default:
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error: unsupported operation '%s'. Supported operations: read, write, list", operation),
				},
			},
			IsError: true,
		}, nil
	}
}

func (f *FileSystemTool) handleRead(arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Error: path parameter is required for read operation",
				},
			},
			IsError: true,
		}, nil
	}

	// In a real implementation, this would:
	// 1. Validate path is within baseDir
	// 2. Read the actual file
	// 3. Return file contents
	
	// For experiment purposes, return simulated content
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("File contents of %s:\n// This is simulated file content for MCP experiment\npackage main\n\nfunc main() {\n\tprintln(\"Hello from MCP!\")\n}", path),
			},
		},
		IsError: false,
	}, nil
}

func (f *FileSystemTool) handleWrite(arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Error: path parameter is required for write operation",
				},
			},
			IsError: true,
		}, nil
	}

	content, ok := arguments["content"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Error: content parameter is required for write operation",
				},
			},
			IsError: true,
		}, nil
	}

	// In a real implementation, this would:
	// 1. Validate path is within baseDir
	// 2. Write content to file securely
	// 3. Return success confirmation
	
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path),
			},
		},
		IsError: false,
	}, nil
}

func (f *FileSystemTool) handleList(arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		path = "." // Default to current directory
	}

	// In a real implementation, this would:
	// 1. Validate path is within baseDir
	// 2. List directory contents
	// 3. Return file/directory information
	
	// For experiment purposes, return simulated directory listing
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Directory listing for %s:\n- main.go (file, 1.2KB)\n- README.md (file, 3.4KB)\n- internal/ (directory)\n- experiments/ (directory)", path),
			},
		},
		IsError: false,
	}, nil
}

// CommandTool provides MCP-compatible command execution
type CommandTool struct {
	allowedCommands []string
}

// NewCommandTool creates a new MCP command tool
func NewCommandTool(allowedCommands []string) *CommandTool {
	return &CommandTool{allowedCommands: allowedCommands}
}

// Name returns the tool name
func (c *CommandTool) Name() string {
	return "command"
}

// Description returns the tool description
func (c *CommandTool) Description() string {
	return "Executes allowed shell commands with security restrictions"
}

// InputSchema returns the JSON schema for tool inputs
func (c *CommandTool) InputSchema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to execute (must be in allowed list)",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Command arguments (optional)",
			},
		},
		Required: []string{"command"},
	}
}

// Handle executes the command
func (c *CommandTool) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	command, ok := arguments["command"].(string)
	if !ok {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Error: command parameter is required and must be a string",
				},
			},
			IsError: true,
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
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Error: command '%s' is not allowed. Allowed commands: %v", command, c.allowedCommands),
				},
			},
			IsError: true,
		}, nil
	}

	// In a real implementation, this would:
	// 1. Execute the command securely
	// 2. Capture output and error streams
	// 3. Return execution results
	
	// For experiment purposes, return simulated output
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Command '%s' executed successfully:\nSimulated output for MCP experiment\nExit code: 0", command),
			},
		},
		IsError: false,
	}, nil
}