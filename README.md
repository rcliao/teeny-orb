# Teeny-Orb

An AI-powered coding assistant that executes operations within containerized environments for security and isolation. Built with Go and integrating LLM capabilities through the Model Context Protocol (MCP).

## Features

- **Interactive Coding Sessions** - Chat-based interface for iterative development
- **Container-Based Execution** - Isolated environments for secure code execution
- **LLM Integration** - Pluggable providers (OpenAI, Anthropic) for AI assistance
- **MCP Protocol** - Embedded Model Context Protocol server for tool extensibility
- **Context Management** - Intelligent context window management and persistence
- **Dual Interface** - Both CLI and TUI (Terminal UI) modes

## Quick Start

```bash
# Build the application
go build ./cmd/teeny-orb

# Start an interactive session
./teeny-orb

# Non-interactive mode
./teeny-orb generate "create a REST API handler"

# Work with a specific project
./teeny-orb --project ./myapp

# Review existing code
./teeny-orb review main.go
```

## Architecture

The project follows domain-driven design with these core domains:

- **Container Management** (`internal/container/`) - Docker-based session isolation and file sync
- **LLM Integration** (`internal/llm/`) - Pluggable providers for AI capabilities
- **MCP Server** (`internal/mcp/`) - Embedded Model Context Protocol server
- **Context Management** (`internal/context/`) - Context window management and persistence
- **CLI Interface** (`internal/cli/`) - Cobra-based commands
- **TUI Interface** (`internal/tui/`) - Bubble Tea terminal UI

## Requirements

- Go 1.21+
- Docker 20.10+ or containerd 1.5+

## Development

```bash
# Run tests
go test ./...

# Clean up dependencies
go mod tidy

# Build for development
go build ./cmd/teeny-orb
```

## Security

All code execution happens within isolated containers with:
- No privileged access
- Explicit volume mounts only
- Network isolation from host
- Resource limits per session

## Contributing

This is an early-stage project focused on demonstrating AI-powered coding assistance patterns with container isolation and MCP integration.

## License

[License information to be added]