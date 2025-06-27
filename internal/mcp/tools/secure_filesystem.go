package tools

import (
	"context"
	"fmt"

	"github.com/rcliao/teeny-orb/internal/mcp"
	"github.com/rcliao/teeny-orb/internal/mcp/security"
)

// SecureFileSystemTool provides MCP-compatible file system operations with security
type SecureFileSystemTool struct {
	baseDir   string
	validator *security.SecurityValidator
}

// NewSecureFileSystemTool creates a new secure MCP filesystem tool
func NewSecureFileSystemTool(baseDir string, validator *security.SecurityValidator) *SecureFileSystemTool {
	return &SecureFileSystemTool{
		baseDir:   baseDir,
		validator: validator,
	}
}

// Name returns the tool name
func (f *SecureFileSystemTool) Name() string {
	return "secure_filesystem"
}

// Description returns the tool description
func (f *SecureFileSystemTool) Description() string {
	return "Provides secure file system operations with permission validation and audit logging"
}

// InputSchema returns the JSON schema for tool inputs
func (f *SecureFileSystemTool) InputSchema() mcp.InputSchema {
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

// Handle executes the filesystem operation with security validation
func (f *SecureFileSystemTool) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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
		return f.handleSecureRead(ctx, arguments)
	case "write":
		return f.handleSecureWrite(ctx, arguments)
	case "list":
		return f.handleSecureList(ctx, arguments)
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

func (f *SecureFileSystemTool) handleSecureRead(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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

	// Validate security permissions
	if err := f.validator.ValidateFileOperation(ctx, "read", path); err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Security violation: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Simulate file read (in real implementation, would read actual file)
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("SECURE READ: File contents of %s:\n// Security-validated file content\npackage main\n\nfunc main() {\n\tprintln(\"Secure file access!\")\n}", path),
			},
		},
		IsError: false,
	}, nil
}

func (f *SecureFileSystemTool) handleSecureWrite(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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

	// Validate security permissions
	if err := f.validator.ValidateFileOperation(ctx, "write", path); err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Security violation: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Simulate file write (in real implementation, would write actual file)
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("SECURE WRITE: Successfully wrote %d bytes to %s (security validated)", len(content), path),
			},
		},
		IsError: false,
	}, nil
}

func (f *SecureFileSystemTool) handleSecureList(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		path = "." // Default to current directory
	}

	// Validate security permissions
	if err := f.validator.ValidateFileOperation(ctx, "list", path); err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Security violation: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Simulate directory listing (in real implementation, would list actual directory)
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("SECURE LIST: Directory listing for %s (security validated):\n- main.go (file, 1.2KB)\n- config.json (file, 256B)\n- data/ (directory)\n- logs/ (directory)", path),
			},
		},
		IsError: false,
	}, nil
}

// SecureCommandTool provides MCP-compatible command execution with security
type SecureCommandTool struct {
	validator *security.SecurityValidator
}

// NewSecureCommandTool creates a new secure MCP command tool
func NewSecureCommandTool(validator *security.SecurityValidator) *SecureCommandTool {
	return &SecureCommandTool{
		validator: validator,
	}
}

// Name returns the tool name
func (c *SecureCommandTool) Name() string {
	return "secure_command"
}

// Description returns the tool description
func (c *SecureCommandTool) Description() string {
	return "Executes shell commands with security validation and audit logging"
}

// InputSchema returns the JSON schema for tool inputs
func (c *SecureCommandTool) InputSchema() mcp.InputSchema {
	return mcp.InputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to execute (must be in security whitelist)",
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

// Handle executes the command with security validation
func (c *SecureCommandTool) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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

	// Extract args if provided
	var args []string
	if argsInterface, ok := arguments["args"]; ok {
		if argsSlice, ok := argsInterface.([]interface{}); ok {
			args = make([]string, len(argsSlice))
			for i, arg := range argsSlice {
				if argStr, ok := arg.(string); ok {
					args[i] = argStr
				}
			}
		}
	}

	// Validate security permissions
	if err := c.validator.ValidateCommandExecution(ctx, command, args); err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Security violation: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Simulate command execution (in real implementation, would execute actual command)
	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("SECURE EXEC: Command '%s %v' executed successfully (security validated)\nSimulated output for security testing", command, args),
			},
		},
		IsError: false,
	}, nil
}