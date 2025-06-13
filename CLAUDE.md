# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

teeny-orb is a Go CLI application for AI-powered coding assistance that executes operations within containerized environments for security and isolation. It integrates LLM capabilities with local development through the Model Context Protocol (MCP).

## Commands

This is an early-stage project. Standard Go commands will apply:

- `go build ./cmd/teeny-orb` - Build the main binary
- `go test ./...` - Run all tests  
- `go mod tidy` - Clean up dependencies

## Architecture

The project follows domain-driven design with these core domains:

- **Container Management** (`internal/container/`) - Docker-based session isolation and file sync
- **LLM Integration** (`internal/llm/`) - Pluggable providers (OpenAI, Anthropic)
- **MCP Server** (`internal/mcp/`) - Embedded Model Context Protocol server
- **Context Management** (`internal/context/`) - Context window management and persistence
- **CLI Interface** (`internal/cli/`) - Cobra-based commands
- **TUI Interface** (`internal/tui/`) - Bubble Tea terminal UI

Key architectural decisions:
- Session-based containers for persistent state
- Embedded MCP server in main process
- Single LLM provider per configuration
- Interface-based design for dependency injection