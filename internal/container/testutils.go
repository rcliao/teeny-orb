package container

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockSession implements Session interface for testing
type MockSession struct {
	id       string
	status   SessionStatus
	config   SessionConfig
	commands [][]string
	closed   bool
}

// NewMockSession creates a new mock session for testing
func NewMockSession(id string, config SessionConfig) *MockSession {
	return &MockSession{
		id:       id,
		status:   StatusRunning,
		config:   config,
		commands: make([][]string, 0),
	}
}

func (s *MockSession) ID() string {
	return s.id
}

func (s *MockSession) Status() SessionStatus {
	return s.status
}

func (s *MockSession) Execute(ctx context.Context, cmd []string) (*ExecResult, error) {
	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	s.commands = append(s.commands, cmd)

	result := &ExecResult{
		ExitCode: 0,
		Stdout:   strings.NewReader(fmt.Sprintf("Mock execution: %s", strings.Join(cmd, " "))),
		Stderr:   strings.NewReader(""),
		Duration: time.Millisecond * 100,
	}

	return result, nil
}

func (s *MockSession) SyncFiles(ctx context.Context, direction SyncDirection) error {
	if s.closed {
		return fmt.Errorf("session is closed")
	}
	return nil
}

func (s *MockSession) Close() error {
	s.status = StatusStopped
	s.closed = true
	return nil
}

// GetExecutedCommands returns all commands executed on this session
func (s *MockSession) GetExecutedCommands() [][]string {
	return s.commands
}

// SetStatus sets the session status (for testing)
func (s *MockSession) SetStatus(status SessionStatus) {
	s.status = status
}

// MockManager implements Manager interface for testing
type MockManager struct {
	sessions      map[string]Session
	createError   error
	cleanupCalled bool
}

// NewMockManager creates a new mock manager for testing
func NewMockManager() *MockManager {
	return &MockManager{
		sessions: make(map[string]Session),
	}
}

func (m *MockManager) CreateSession(ctx context.Context, config SessionConfig) (Session, error) {
	if m.createError != nil {
		return nil, m.createError
	}

	id := fmt.Sprintf("mock-session-%d", len(m.sessions))
	session := NewMockSession(id, config)
	m.sessions[id] = session
	return session, nil
}

func (m *MockManager) GetSession(id string) (Session, error) {
	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}
	return session, nil
}

func (m *MockManager) ListSessions() []Session {
	sessions := make([]Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (m *MockManager) Cleanup(ctx context.Context) error {
	m.cleanupCalled = true
	for id, session := range m.sessions {
		if session.Status() == StatusStopped || session.Status() == StatusError {
			delete(m.sessions, id)
		}
	}
	return nil
}

// SetCreateError sets an error to be returned by CreateSession (for testing)
func (m *MockManager) SetCreateError(err error) {
	m.createError = err
}

// WasCleanupCalled returns whether Cleanup was called
func (m *MockManager) WasCleanupCalled() bool {
	return m.cleanupCalled
}
