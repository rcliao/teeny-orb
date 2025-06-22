package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rcliao/teeny-orb/internal/mcp"
	"github.com/rcliao/teeny-orb/internal/mcp/server"
	"github.com/rcliao/teeny-orb/internal/mcp/tools"
	"github.com/rcliao/teeny-orb/internal/mcp/transport"
)

func main() {
	var (
		name    = flag.String("name", "teeny-orb-mcp-server", "Server name")
		version = flag.String("version", "0.1.0", "Server version")
		debug   = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	if *debug {
		log.SetOutput(os.Stderr)
		log.Println("Starting MCP server in debug mode")
	} else {
		// Disable logging to avoid interfering with MCP protocol
		log.SetOutput(io.Discard)
	}

	// Create MCP server
	mcpServer := server.NewServer(*name, *version)

	// Register tools
	if err := registerTools(mcpServer); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Create stdio transport
	transport := transport.NewStdioTransport()
	defer transport.Close()

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

	// Run server
	if err := runServer(ctx, mcpServer, transport, *debug); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	if *debug {
		log.Println("MCP server shutdown complete")
	}
}

// registerTools registers all available tools with the server
func registerTools(server *server.Server) error {
	// Register filesystem tool
	fsTools := tools.NewFileSystemTool(".")
	if err := server.RegisterTool(fsTools); err != nil {
		return fmt.Errorf("failed to register filesystem tool: %w", err)
	}

	// Register command tool with safe commands
	cmdTool := tools.NewCommandTool([]string{"echo", "pwd", "ls", "date", "whoami"})
	if err := server.RegisterTool(cmdTool); err != nil {
		return fmt.Errorf("failed to register command tool: %w", err)
	}

	return nil
}

// runServer runs the MCP server with the given transport
func runServer(ctx context.Context, server *server.Server, transport mcp.Transport, debug bool) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Receive message
		msg, err := transport.Receive(ctx)
		if err != nil {
			if err == io.EOF {
				if debug {
					log.Println("Client disconnected")
				}
				return nil
			}
			return fmt.Errorf("failed to receive message: %w", err)
		}

		if debug {
			log.Printf("Received: %s %v", msg.Method, msg.ID)
		}

		// Process message
		response, err := server.HandleMessage(ctx, msg)
		if err != nil {
			if debug {
				log.Printf("Error handling message: %v", err)
			}
			continue
		}

		// Send response (if not a notification)
		if response != nil {
			if err := transport.Send(ctx, response); err != nil {
				if debug {
					log.Printf("Failed to send response: %v", err)
				}
				return fmt.Errorf("failed to send response: %w", err)
			}

			if debug {
				log.Printf("Sent response for %v", response.ID)
			}
		}
	}
}