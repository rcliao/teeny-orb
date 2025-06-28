package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
			"env": map[string]interface{}{
				"type":        "object",
				"description": "Environment variables to set for the command (optional)",
				"additionalProperties": map[string]interface{}{"type": "string"},
			},
		},
		Required: []string{"command"},
	}
}

// Handle executes the command with enhanced cross-platform support
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

	// Extract environment variables if provided
	var envVars map[string]string
	if envInterface, ok := arguments["env"]; ok {
		if envMap, ok := envInterface.(map[string]interface{}); ok {
			envVars = make(map[string]string)
			for k, v := range envMap {
				if vStr, ok := v.(string); ok {
					envVars[k] = vStr
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

	// Execute the command with enhanced configuration
	result, err := c.executeCommand(ctx, command, args, envVars)
	if err != nil {
		return &mcp.CallToolResponse{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: result,
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResponse{
		Content: []mcp.Content{
			{
				Type: "text",
				Text: result,
			},
		},
		IsError: false,
	}, nil
}

// executeCommand performs cross-platform command execution with enhanced environment management
func (c *RealCommandTool) executeCommand(ctx context.Context, command string, args []string, envVars map[string]string) (string, error) {
	// Prepare command execution based on platform
	cmd, err := c.prepareCommand(ctx, command, args)
	if err != nil {
		return "", fmt.Errorf("failed to prepare command: %w", err)
	}

	// Set working directory
	cmd.Dir = c.workDir

	// Configure environment
	if err := c.configureEnvironment(cmd, command, envVars); err != nil {
		return "", fmt.Errorf("failed to configure environment: %w", err)
	}

	// Execute with timeout
	start := time.Now()
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	// Format result
	result := c.formatCommandResult(command, args, output, err, duration)

	if err != nil {
		return result, fmt.Errorf("command execution failed")
	}

	return result, nil
}

// prepareCommand creates the appropriate command for the current platform
func (c *RealCommandTool) prepareCommand(ctx context.Context, command string, args []string) (*exec.Cmd, error) {
	// Handle shell commands differently on Windows
	if runtime.GOOS == "windows" {
		return c.prepareWindowsCommand(ctx, command, args)
	}
	return c.prepareUnixCommand(ctx, command, args)
}

// prepareWindowsCommand handles Windows-specific command preparation
func (c *RealCommandTool) prepareWindowsCommand(ctx context.Context, command string, args []string) (*exec.Cmd, error) {
	// Check for shell built-ins that need cmd.exe
	shellBuiltins := map[string]bool{
		"dir": true, "cd": true, "copy": true, "move": true, "del": true,
		"type": true, "echo": true, "set": true, "where": true,
	}

	if shellBuiltins[strings.ToLower(command)] {
		// Use cmd.exe for shell built-ins
		cmdArgs := append([]string{"/c", command}, args...)
		return exec.CommandContext(ctx, "cmd.exe", cmdArgs...), nil
	}

	// For regular executables, try to find with extension
	if !strings.Contains(command, ".") {
		// Try common Windows executable extensions
		extensions := []string{".exe", ".bat", ".cmd", ".com"}
		for _, ext := range extensions {
			if _, err := exec.LookPath(command + ext); err == nil {
				command = command + ext
				break
			}
		}
	}

	return exec.CommandContext(ctx, command, args...), nil
}

// prepareUnixCommand handles Unix-like command preparation
func (c *RealCommandTool) prepareUnixCommand(ctx context.Context, command string, args []string) (*exec.Cmd, error) {
	return exec.CommandContext(ctx, command, args...), nil
}

// configureEnvironment sets up the command environment with proper variable handling
func (c *RealCommandTool) configureEnvironment(cmd *exec.Cmd, command string, envVars map[string]string) error {
	// Start with system environment
	cmd.Env = os.Environ()

	// Add command-specific environment variables
	c.addCommandSpecificEnv(cmd, command)

	// Add user-provided environment variables
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	return nil
}

// addCommandSpecificEnv adds environment variables specific to certain commands
func (c *RealCommandTool) addCommandSpecificEnv(cmd *exec.Cmd, command string) {
	switch command {
	case "go":
		// Ensure Go has proper cache directories
		goCacheDir := filepath.Join(c.workDir, ".go-cache")
		goModCacheDir := filepath.Join(c.workDir, ".go-mod-cache")
		goTmpDir := filepath.Join(c.workDir, ".go-tmp")

		// Create directories if they don't exist
		os.MkdirAll(goCacheDir, 0755)
		os.MkdirAll(goModCacheDir, 0755)
		os.MkdirAll(goTmpDir, 0755)

		// Set Go environment variables
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("GOCACHE=%s", goCacheDir),
			fmt.Sprintf("GOMODCACHE=%s", goModCacheDir),
			fmt.Sprintf("GOTMPDIR=%s", goTmpDir),
		)

	case "npm", "yarn", "node":
		// Set Node.js cache directories
		npmCacheDir := filepath.Join(c.workDir, ".npm-cache")
		os.MkdirAll(npmCacheDir, 0755)
		cmd.Env = append(cmd.Env, fmt.Sprintf("npm_config_cache=%s", npmCacheDir))

	case "python", "python3", "pip", "pip3":
		// Set Python cache directories
		pythonCacheDir := filepath.Join(c.workDir, ".python-cache")
		os.MkdirAll(pythonCacheDir, 0755)
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("PYTHONPYCACHEPREFIX=%s", pythonCacheDir),
			fmt.Sprintf("PIP_CACHE_DIR=%s", pythonCacheDir),
		)
	}
}

// formatCommandResult creates a standardized command result format
func (c *RealCommandTool) formatCommandResult(command string, args []string, output []byte, err error, duration time.Duration) string {
	var result strings.Builder

	// Command header
	result.WriteString(fmt.Sprintf("Command: %s", command))
	if len(args) > 0 {
		result.WriteString(fmt.Sprintf(" %s", strings.Join(args, " ")))
	}
	result.WriteString(fmt.Sprintf("\nDuration: %v\n", duration.Round(time.Millisecond)))
	result.WriteString(fmt.Sprintf("Working Directory: %s\n", c.workDir))

	// Output section
	if len(output) > 0 {
		result.WriteString("\nOutput:\n")
		result.WriteString(strings.TrimSpace(string(output)))
		result.WriteString("\n")
	}

	// Error section
	if err != nil {
		result.WriteString(fmt.Sprintf("\nError: %v\n", err))
		if exitError, ok := err.(*exec.ExitError); ok {
			result.WriteString(fmt.Sprintf("Exit Code: %d\n", exitError.ExitCode()))
		}
	}

	return result.String()
}