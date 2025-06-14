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

## Project Planning & Status Tracking

### Planning Document Maintenance

The project maintains a comprehensive planning document at `/docs/planning/v0.md` that tracks development progress across all phases. **You MUST update this document whenever making significant changes to the codebase.**

#### When to Update Planning Status

**Always update the planning document when:**
- Completing any task listed in the phase breakdowns
- Adding new features or components
- Implementing new tests or infrastructure
- Fixing major bugs or technical debt
- Making architectural changes
- Adding or updating dependencies

#### How to Update Planning Status

1. **Update the Status Overview Table** (`docs/planning/v0.md` lines ~15-25):
   - Change phase status symbols: ⏳ (not started) → 🔄 (in progress) → ✅ (completed)
   - Update progress percentages based on completed tasks
   - Update "Key Achievements" and "Remaining Work" columns

2. **Update Individual Task Checkboxes**:
   - Mark completed tasks with `[x]` and add ✅ emoji
   - Keep incomplete tasks as `[ ]`

3. **Update Status Sections**:
   - **Test Coverage Status**: Update test counts and coverage information
   - **Build Status**: Update build/test/lint status
   - **Dependencies Status**: Mark new dependencies as ✅ or update versions
   - **Architecture Implementation Status**: Move components between sections as they progress

4. **Update "Last Updated" Date**: Change the date in the status section header

5. **Update Next Immediate Actions**: Revise based on current priorities

#### Example Status Update Workflow

```bash
# After implementing a new feature, always:
1. Test the changes: make test
2. Update planning status in docs/planning/v0.md
3. Commit both code and documentation changes together
```

#### Status Tracking Guidelines

- **Be Accurate**: Only mark tasks as complete when fully implemented and tested
- **Be Specific**: Update achievement descriptions to reflect actual implementation
- **Be Current**: Always update the "Last Updated" date
- **Be Comprehensive**: Include both positive progress and known issues/technical debt

### Planning Document Structure

The planning document includes:
- **Status Overview**: High-level progress tracking with tables and metrics
- **Phase Breakdowns**: Detailed task lists for each development phase  
- **Architecture Tracking**: Component implementation status
- **Technical Debt**: Known issues and areas needing improvement

This approach ensures the planning document serves as a living, accurate reflection of the project's current state and helps maintain development momentum.

### Planning Document Archiving

When development phases or the overall plan are completed, planning documents should be archived to maintain a clean and current documentation structure.

#### Archive Triggers

**Archive planning documents when:**
- **Individual Phase Complete**: When a major phase (Phase 0-8) reaches 100% completion
- **Overall Plan Complete**: When the entire v0 development plan is finished
- **Major Milestone Reached**: When transitioning to a new major version or development cycle
- **Document Superseded**: When a new planning document replaces the current one

#### Archiving Process

1. **Create Archive Directory Structure**:
   ```bash
   mkdir -p docs/archive/planning/
   ```

2. **Move Completed Documents**:
   ```bash
   # For completed phases
   mv docs/planning/v0.md docs/archive/planning/v0-completed-YYYY-MM-DD.md
   
   # For superseded documents  
   mv docs/planning/v0.md docs/archive/planning/v0-superseded-YYYY-MM-DD.md
   ```

3. **Update Archive Document**:
   - Add completion date and final status summary at the top
   - Mark all phases as completed ✅
   - Add final metrics (test coverage, build status, etc.)
   - Include lessons learned or retrospective notes

4. **Extract Key Decisions** (see Decision Records section below):
   - Identify technical architecture decisions made during development
   - Document functional requirement decisions and trade-offs
   - Create decision records in `/docs/architecture/` or `/docs/requirements/`
   - Reference decision records from the archived planning document

5. **Create New Planning Document** (if continuing development):
   - Start fresh planning document for next phase/version
   - Reference archived document for context
   - Reset status tracking for new objectives

#### Archive Document Header Example

```markdown
# Teeny-Orb Development Plan v0 - COMPLETED

**Completed**: 2025-XX-XX  
**Duration**: X weeks  
**Final Status**: ✅ All 8 phases completed successfully

## Final Metrics
- **Total Tests**: XX passing
- **Coverage**: XX%  
- **Features Delivered**: X major features
- **Technical Debt**: Resolved/Documented

## Lessons Learned
- Key insights from development process
- What worked well
- Areas for improvement in future planning

---

[Original planning document content below...]
```

#### Archive Management

- **Naming Convention**: `vX-{status}-YYYY-MM-DD.md`
  - Status: `completed`, `superseded`, `cancelled`
  - Date: Archive date
- **Index File**: Maintain `docs/archive/README.md` listing all archived documents
- **Retention**: Keep archived planning documents indefinitely for historical reference
- **Access**: Archived documents remain searchable and accessible for future reference

#### Current vs. Archive Structure

```
docs/
├── planning/           # Current active planning documents
│   └── v1.md          # Next version planning (after v0 archived)
├── archive/           # Completed/superseded documents
│   ├── README.md      # Archive index
│   └── planning/
│       ├── v0-completed-2025-06-14.md
│       └── v0-phase1-completed-2025-05-01.md
└── requirements/      # Current requirements (keep active)
```

This archiving system ensures:
- ✅ Current planning documents stay focused and actionable
- ✅ Historical development progress is preserved  
- ✅ Documentation doesn't become cluttered with completed items
- ✅ Future planning can reference past lessons learned
- ✅ Project evolution is clearly tracked over time

### Decision Records Management

During planning document reviews and archiving, extract key decisions to maintain long-term decision records that survive planning document cycles.

#### When to Extract Decisions

**Extract decisions during:**
- **Planning Document Reviews** - Regular review of active planning documents
- **Phase Completions** - When archiving completed phases
- **Architecture Changes** - When making significant technical decisions
- **Requirement Changes** - When functional requirements evolve or are clarified
- **Problem Resolution** - When solving complex technical or design problems

#### Types of Decisions to Document

**Technical Architecture Decisions:**
- Interface design choices (e.g., Session interface, Manager interface)
- Technology selections (e.g., Docker API, Cobra CLI framework)
- Design pattern adoptions (e.g., Registry pattern, dependency injection)
- Security model decisions (e.g., container isolation, resource limits)
- Performance trade-offs (e.g., session management approach)

**Functional Requirement Decisions:**
- Feature scope definitions (e.g., host vs container execution modes)
- User experience choices (e.g., CLI command structure)
- Integration decisions (e.g., MCP protocol adoption)
- Configuration management approach (e.g., Viper with YAML/JSON/ENV)

#### Decision Record Format

Create decision records using this template:

```markdown
# ADR-XXX: [Decision Title]

**Date**: YYYY-MM-DD  
**Status**: Accepted | Superseded | Deprecated  
**Context**: Planning Phase X | Issue Resolution | Architecture Review

## Context and Problem Statement

Brief description of the problem or decision that needs to be made.

## Decision Drivers

- Driver 1 (e.g., performance requirement)
- Driver 2 (e.g., security constraint)  
- Driver 3 (e.g., maintainability goal)

## Considered Options

1. **Option A**: Brief description
2. **Option B**: Brief description
3. **Option C**: Brief description

## Decision Outcome

**Chosen Option**: Option X

**Rationale**:
- Why this option was selected
- Key benefits and trade-offs
- Alignment with project goals

## Implementation Details

- Specific implementation approach
- Key components or interfaces involved
- Integration points with existing architecture

## Consequences

**Positive**:
- Benefit 1
- Benefit 2

**Negative**:
- Trade-off 1
- Trade-off 2

**Neutral**:
- Impact 1
- Impact 2

## References

- Link to planning document section
- Related ADRs
- External documentation or research
```

#### Decision Record Organization

```
docs/
├── architecture/
│   ├── README.md              # Architecture decision index
│   ├── ADR-001-session-interface.md
│   ├── ADR-002-registry-pattern.md
│   ├── ADR-003-container-security.md
│   └── ADR-004-dependency-injection.md
├── requirements/
│   ├── README.md              # Requirements decision index  
│   ├── RDR-001-cli-structure.md
│   ├── RDR-002-configuration-management.md
│   └── RDR-003-execution-modes.md
└── planning/
    └── v0.md                  # References relevant ADRs/RDRs
```

#### Decision Extraction Process

1. **Review Planning Documents**:
   - Scan for "Key Design Patterns", "Architecture" sections
   - Look for problem statements and solution explanations
   - Identify trade-offs and alternatives mentioned

2. **Identify Decision Points**:
   - Major technical choices (frameworks, patterns, interfaces)
   - Functional requirement clarifications
   - Problem resolution approaches

3. **Create Decision Records**:
   - Use appropriate prefix: ADR (Architecture) or RDR (Requirements)
   - Number sequentially: ADR-001, ADR-002, etc.
   - Cross-reference with planning documents

4. **Update Index Files**:
   - Maintain `docs/architecture/README.md` with ADR list
   - Maintain `docs/requirements/README.md` with RDR list
   - Include brief descriptions and status

5. **Link Decision Records**:
   - Reference ADRs/RDRs from planning documents
   - Link related decisions to each other
   - Update status when decisions are superseded

#### Example Decision Extraction

From the current v0 planning, key decisions to extract:

**Architecture Decisions (ADRs)**:
- Session interface design with Docker/host implementations
- Registry pattern for unified session management  
- Interface-based dependency injection for testability
- Container security configuration approach

**Requirements Decisions (RDRs)**:
- CLI command structure (generate, review, session)
- Configuration management with Viper
- Dual execution modes (host vs containerized)
- MCP protocol adoption for AI integration

This decision record system ensures:
- ✅ **Persistent Knowledge**: Key decisions survive planning document cycles
- ✅ **Context Preservation**: Rationale and alternatives are documented
- ✅ **Future Reference**: New team members can understand past decisions
- ✅ **Decision Evolution**: Superseded decisions are tracked and replaced
- ✅ **Architectural Coherence**: System design remains consistent over time