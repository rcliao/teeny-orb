package container

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// hostSession implements Session interface using the host system
type hostSession struct {
	id      string
	workDir string
	env     map[string]string
	status  SessionStatus
}

// hostManager implements the Manager interface using host processes
type hostManager struct {
	sessions map[string]Session
}

// NewHostManager creates a new host-based session manager
func NewHostManager() Manager {
	return &hostManager{
		sessions: make(map[string]Session),
	}
}

// NewHostSession creates a new host-based session
func NewHostSession(config SessionConfig) (Session, error) {
	return NewHostSessionWithIDGen(config, &DefaultIDGenerator{})
}

// NewHostSessionWithIDGen creates a new host-based session with custom ID generator
func NewHostSessionWithIDGen(config SessionConfig, idGen IDGenerator) (Session, error) {
	sessionID := idGen.GenerateID()

	// Use current working directory if no workDir specified
	workDir := config.WorkDir
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		workDir = cwd
	}

	// Verify working directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("working directory does not exist: %s", workDir)
	}

	session := &hostSession{
		id:      sessionID,
		workDir: workDir,
		env:     config.Environment,
		status:  StatusRunning,
	}

	return session, nil
}

func (m *hostManager) CreateSession(ctx context.Context, config SessionConfig) (Session, error) {
	session, err := NewHostSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create host session: %w", err)
	}

	m.sessions[session.ID()] = session
	return session, nil
}

func (m *hostManager) GetSession(id string) (Session, error) {
	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return session, nil
}

func (m *hostManager) ListSessions() []Session {
	sessions := make([]Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func (m *hostManager) Cleanup(ctx context.Context) error {
	for id, session := range m.sessions {
		if session.Status() == StatusStopped || session.Status() == StatusError {
			delete(m.sessions, id)
		}
	}

	return nil
}

func (s *hostSession) ID() string {
	return s.id
}

func (s *hostSession) Status() SessionStatus {
	return s.status
}

func (s *hostSession) Execute(ctx context.Context, cmd []string) (*ExecResult, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("no command provided")
	}

	start := time.Now()

	// Create command
	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	execCmd.Dir = s.workDir

	// Set environment variables
	execCmd.Env = os.Environ()
	for key, value := range s.env {
		execCmd.Env = append(execCmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Run command and capture output
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := execCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for command to complete
	err = execCmd.Wait()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	result := &ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Duration: time.Since(start),
	}

	return result, nil
}

func (s *hostSession) SyncFiles(ctx context.Context, direction SyncDirection) error {
	// For host sessions, file sync is not needed since we're working directly on the host
	switch direction {
	case SyncToContainer, SyncFromContainer, SyncBidirectional:
		// No-op for host sessions
		return nil
	}
	return nil
}

func (s *hostSession) Close() error {
	s.status = StatusStopped
	return nil
}
