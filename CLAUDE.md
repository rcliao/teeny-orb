# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

teeny-orb is a Go CLI application for AI-powered coding assistance that executes operations within containerized environments for security and isolation. It integrates LLM capabilities with local development through the Model Context Protocol (MCP).

## Commands

### Build and Development
- `go build ./cmd/teeny-orb` - Build the main binary
- `go mod tidy` - Clean up dependencies

### Testing (Makefile available)
- `make test` - Run all tests
- `make test-short` - Run unit tests only (skip integration tests)
- `make test-coverage` - Run tests with coverage report
- `make test-race` - Run tests with race detection
- `make bench` - Run benchmarks
- `go test -v ./internal/container/` - Run specific package tests
- `go test -run TestSpecificFunction` - Run single test

### Quality Checks
- `make check` - Quick quality check (fmt + vet + test-short)
- `make fmt` - Format code
- `make vet` - Vet code
- `make lint` - Run linter (requires golangci-lint)

### Application Usage
- `./teeny-orb` - Start interactive session
- `./teeny-orb generate "prompt"` - Generate code from prompt
- `./teeny-orb review main.go` - Review code file
- `./teeny-orb session create --docker --image alpine:latest` - Create containerized session

## Architecture

The project follows domain-driven design with interface-based dependency injection:

### Core Components
- **Container Management** (`internal/container/`) - Session lifecycle with Docker and host execution
- **CLI Interface** (`internal/cli/`) - Cobra-based commands with subcommands for generate, review, session
- **Registry Pattern** - Singleton ManagerRegistry manages both host and Docker session managers

### Key Design Patterns

**Session Management:**
- `Session` interface with `dockerSession` and `hostSession` implementations
- `Manager` interface with `dockerManager` and `hostManager` implementations
- Sessions have unique IDs, status tracking, command execution, and file sync capabilities

**Dependency Injection:**
- `IDGenerator` interface with `DefaultIDGenerator` (time-based) and `StaticIDGenerator` (for testing)
- Constructor functions accept interfaces for testability (`NewDockerSessionWithIDGen`, `NewHostSessionWithIDGen`)

**Registry Pattern:**
- `ManagerRegistry` provides unified access to both host and Docker managers
- Thread-safe with mutex protection
- Lazy initialization of Docker manager (only when needed)

### Testing Architecture

**Mock Infrastructure:**
- `testutils.go` contains `MockSession` and `MockManager` implementations
- Interface-based mocking for all external dependencies
- Error injection capabilities for testing failure scenarios

**Test Organization:**
- Unit tests: `*_test.go` files alongside code
- Integration tests: Use `//go:build integration` build tags
- Benchmarks: Focus on ID generation and session management performance
- Coverage: Aim for >45% (current range: 17.6% to 90.9%)

**Testing Commands:**
- Use `make test-short` for unit tests during development
- Docker integration tests require Docker daemon running
- Tests use temporary directories and clean up resources

### Container Execution Model

**Host Sessions:**
- Execute commands directly on host using `os/exec`
- Working directory validation and environment variable injection
- No file sync needed (direct host access)

**Docker Sessions:**
- Create isolated containers with resource limits
- Command execution via Docker exec API
- File synchronization between host and container (basic implementation)
- Automatic cleanup on session close

### CLI Command Structure

**Root Command:**
- Supports `--config`, `--project`, `--verbose` flags
- Default behavior starts interactive session

**Subcommands:**
- `generate [prompt]` - AI code generation (Phase 2 implementation planned)
- `review [file]` - Code review assistance (Phase 2 implementation planned)
- `session create|list|stop` - Session management with optional Docker isolation

Key architectural decisions:
- Session-based containers for persistent state
- Interface-based design enables comprehensive testing and dependency injection
- Registry pattern provides unified session management across execution environments
- Separation between host and containerized execution with identical interfaces