package container

import (
	"context"
	"io"
	"time"
)

// SessionConfig holds configuration for a container session
type SessionConfig struct {
	Image       string
	WorkDir     string
	ProjectPath string
	Environment map[string]string
	Limits      ResourceLimits
}

// ResourceLimits defines resource constraints for containers
type ResourceLimits struct {
	CPUShares int64
	Memory    int64 // in bytes
}

// Session represents an active container session
type Session interface {
	// ID returns the unique session identifier
	ID() string
	
	// Status returns the current session status
	Status() SessionStatus
	
	// Execute runs a command in the session container
	Execute(ctx context.Context, cmd []string) (*ExecResult, error)
	
	// SyncFiles synchronizes files between host and container
	SyncFiles(ctx context.Context, direction SyncDirection) error
	
	// Close terminates the session and cleans up resources
	Close() error
}

// Manager handles container lifecycle and session management
type Manager interface {
	// CreateSession creates a new container session
	CreateSession(ctx context.Context, config SessionConfig) (Session, error)
	
	// GetSession retrieves an existing session by ID
	GetSession(id string) (Session, error)
	
	// ListSessions returns all active sessions
	ListSessions() []Session
	
	// Cleanup removes orphaned containers and resources
	Cleanup(ctx context.Context) error
}

// SessionStatus represents the state of a session
type SessionStatus string

const (
	StatusCreating SessionStatus = "creating"
	StatusRunning  SessionStatus = "running"
	StatusStopped  SessionStatus = "stopped"
	StatusError    SessionStatus = "error"
)

// SyncDirection specifies the direction of file synchronization
type SyncDirection string

const (
	SyncToContainer   SyncDirection = "to_container"
	SyncFromContainer SyncDirection = "from_container"
	SyncBidirectional SyncDirection = "bidirectional"
)

// ExecResult contains the result of command execution
type ExecResult struct {
	ExitCode int
	Stdout   io.Reader
	Stderr   io.Reader
	Duration time.Duration
}