package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rcliao/teeny-orb/internal/mcp"
	"github.com/rcliao/teeny-orb/internal/mcp/security"
)

// RealFileSystemTool provides actual file system operations with security
type RealFileSystemTool struct {
	baseDir   string
	validator *security.SecurityValidator
}

// NewRealFileSystemTool creates a new real filesystem tool
func NewRealFileSystemTool(baseDir string, validator *security.SecurityValidator) *RealFileSystemTool {
	// Ensure baseDir is absolute
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		absBaseDir = baseDir
	}
	
	return &RealFileSystemTool{
		baseDir:   absBaseDir,
		validator: validator,
	}
}

// Name returns the tool name
func (f *RealFileSystemTool) Name() string {
	return "filesystem"
}

// Description returns the tool description
func (f *RealFileSystemTool) Description() string {
	return "Provides real file system operations including read, write, and list with security validation"
}

// InputSchema returns the JSON schema for tool inputs
func (f *RealFileSystemTool) InputSchema() mcp.InputSchema {
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
func (f *RealFileSystemTool) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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
		return f.handleRead(ctx, arguments)
	case "write":
		return f.handleWrite(ctx, arguments)
	case "list":
		return f.handleList(ctx, arguments)
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

func (f *RealFileSystemTool) handleRead(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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

	// Resolve path relative to base directory
	fullPath := f.resolvePath(path)

	// Validate security permissions
	if f.validator != nil {
		if err := f.validator.ValidateFileOperation(ctx, "read", fullPath); err != nil {
			return &mcp.CallToolResponse{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: fmt.Sprintf("Access denied: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
	}

	// Read the actual file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to read file '%s': %v", path, err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("File: %s\n%s", path, string(content)),
			},
		},
		IsError: false,
	}, nil
}

func (f *RealFileSystemTool) handleWrite(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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

	// Resolve path relative to base directory
	fullPath := f.resolvePath(path)

	// Validate security permissions
	if f.validator != nil {
		if err := f.validator.ValidateFileOperation(ctx, "write", fullPath); err != nil {
			return &mcp.CallToolResponse{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: fmt.Sprintf("Access denied: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to create directory '%s': %v", dir, err),
				},
			},
			IsError: true,
		}, nil
	}

	// Write the actual file
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to write file '%s': %v", path, err),
				},
			},
			IsError: true,
		}, nil
	}

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

func (f *RealFileSystemTool) handleList(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		path = "." // Default to current directory
	}

	// Resolve path relative to base directory
	fullPath := f.resolvePath(path)

	// Validate security permissions
	if f.validator != nil {
		if err := f.validator.ValidateFileOperation(ctx, "list", fullPath); err != nil {
			return &mcp.CallToolResponse{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: fmt.Sprintf("Access denied: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
	}

	// Read directory contents
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Failed to list directory '%s': %v", path, err),
				},
			},
			IsError: true,
		}, nil
	}

	// Format directory listing
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Directory listing for %s:\n", path))
	
	if len(entries) == 0 {
		result.WriteString("(empty directory)")
	} else {
		for _, entry := range entries {
			entryType := "file"
			var size string
			
			if entry.IsDir() {
				entryType = "directory"
				size = ""
			} else {
				// Get file size
				if info, err := entry.Info(); err == nil {
					size = fmt.Sprintf(" (%d bytes)", info.Size())
				}
			}
			
			result.WriteString(fmt.Sprintf("- %s (%s)%s\n", entry.Name(), entryType, size))
		}
	}

	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: result.String(),
			},
		},
		IsError: false,
	}, nil
}

// resolvePath resolves a path relative to the base directory
func (f *RealFileSystemTool) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(f.baseDir, path)
}

// RealCommandTool provides actual command execution with security
type RealCommandTool struct {
	validator *security.SecurityValidator
	workDir   string
}

// NewRealCommandTool creates a new real command tool
func NewRealCommandTool(validator *security.SecurityValidator, workDir string) *RealCommandTool {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	
	return &RealCommandTool{
		validator: validator,
		workDir:   workDir,
	}
}

// Name returns the tool name
func (c *RealCommandTool) Name() string {
	return "command"
}

// Description returns the tool description
func (c *RealCommandTool) Description() string {
	return "Executes shell commands with security validation"
}

// InputSchema returns the JSON schema for tool inputs
func (c *RealCommandTool) InputSchema() mcp.InputSchema {
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

// Handle executes the command
func (c *RealCommandTool) Handle(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
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
	if c.validator != nil {
		if err := c.validator.ValidateCommandExecution(ctx, command, args); err != nil {
			return &mcp.CallToolResponse{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: fmt.Sprintf("Access denied: %v", err),
					},
				},
				IsError: true,
			}, nil
		}
	}

	// Execute the actual command
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = c.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output)),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: fmt.Sprintf("Command: %s %v\n%s", command, args, string(output)),
			},
		},
		IsError: false,
	}, nil
}