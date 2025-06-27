package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rcliao/teeny-orb/internal/mcp/security"
	"github.com/rcliao/teeny-orb/internal/mcp/server"
	"github.com/rcliao/teeny-orb/internal/mcp/tools"
	"github.com/rcliao/teeny-orb/internal/mcp/transport"
)

func main() {
	var (
		port    = flag.String("port", "8080", "HTTP server port")
		host    = flag.String("host", "localhost", "HTTP server host")
		name    = flag.String("name", "teeny-orb-mcp-http-server", "Server name")
		version = flag.String("version", "0.1.0", "Server version")
		debug   = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	// Set up logging
	if *debug {
		log.SetOutput(os.Stderr)
		log.Println("Starting MCP HTTP server in debug mode")
	} else {
		log.SetOutput(os.Stderr) // Keep logging for HTTP server
	}

	// Create MCP server
	mcpServer := server.NewServer(*name, *version)

	// Register tools
	if err := registerTools(mcpServer, *debug); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Create HTTP transport
	addr := fmt.Sprintf("%s:%s", *host, *port)
	httpTransport := transport.NewHTTPTransport(addr, mcpServer, *debug)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		if *debug {
			log.Println("Received shutdown signal")
		}
		cancel()
	}()

	// Start HTTP server
	fmt.Printf("🚀 MCP HTTP Server starting on http://%s\n", addr)
	fmt.Printf("📡 MCP endpoint: http://%s/mcp\n", addr)
	fmt.Printf("💚 Health check: http://%s/health\n", addr)
	fmt.Printf("📊 Status info: http://%s/status\n", addr)
	fmt.Println()

	if err := httpTransport.Start(ctx); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}

	if *debug {
		log.Println("MCP HTTP server shutdown complete")
	}
}

// registerTools registers all available tools with the server
func registerTools(server *server.Server, debug bool) error {
	// Get working directory - check environment variable first, then current directory
	workDir := os.Getenv("WORKSPACE_PATH")
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			workDir = "."
		}
	}

	if debug {
		log.Printf("Setting up tools with working directory: %s", workDir)
	}

	// Create security policy - permissive for development but with key restrictions
	policy := &security.SecurityPolicy{
		AllowedPermissions: []security.Permission{
			security.PermissionReadFile,
			security.PermissionWriteFile,
			security.PermissionListDir,
			security.PermissionExecCommand,
		},
		DeniedPermissions: []security.Permission{
			security.PermissionDeleteFile,
			security.PermissionExecSystem,
		},
		PathRestrictions: security.PathRestrictions{
			RequireBasePath: workDir,
			DeniedPaths: []string{
				"/etc",
				"/var",
				"/usr",
				"/bin",
				"/sbin",
				"/root",
				"/proc",
				"/sys",
			},
		},
		CommandWhitelist: []string{
			// Basic commands
			"echo", "pwd", "ls", "date", "whoami", "cat", "grep", "find", "wc", "sort",
			// Development tools
			"git", "go", "make", "npm", "yarn", "python", "node", "pip", "cargo",
			// Build tools
			"docker", "kubectl", "terraform", "ansible",
			// Editor commands
			"vim", "nano", "code",
		},
		ResourceLimits: security.ResourceLimits{
			MaxMemoryMB:     500,
			MaxCPUPercent:   80,
			MaxExecutionSec: 300, // 5 minutes for longer operations
			MaxFileSize:     50 * 1024 * 1024, // 50MB
		},
		AuditLog: true,
	}

	// Create security validator
	validator := security.NewSecurityValidator(policy, "mcp-http-server", "main-session")

	// Register real filesystem tool with security
	fsTools := tools.NewRealFileSystemTool(workDir, validator)
	if err := server.RegisterTool(fsTools); err != nil {
		return fmt.Errorf("failed to register filesystem tool: %w", err)
	}

	// Register real command tool with security
	cmdTool := tools.NewRealCommandTool(validator, workDir)
	if err := server.RegisterTool(cmdTool); err != nil {
		return fmt.Errorf("failed to register command tool: %w", err)
	}

	if debug {
		log.Printf("Successfully registered %d tools", 2)
	}

	return nil
}