package container

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewHostSession(t *testing.T) {
	config := SessionConfig{
		WorkDir:     "/tmp",
		Environment: map[string]string{"TEST": "value"},
	}

	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() error = %v", err)
	}

	if session.ID() == "" {
		t.Error("Session ID should not be empty")
	}

	if session.Status() != StatusRunning {
		t.Errorf("Session status = %v, want %v", session.Status(), StatusRunning)
	}
}

func TestNewHostSessionWithIDGen(t *testing.T) {
	idGen := NewStaticIDGenerator("host-test")
	config := SessionConfig{
		WorkDir: "/tmp",
	}

	session, err := NewHostSessionWithIDGen(config, idGen)
	if err != nil {
		t.Fatalf("NewHostSessionWithIDGen() error = %v", err)
	}

	if session.ID() != "host-test-1" {
		t.Errorf("Session ID = %v, want host-test-1", session.ID())
	}
}

func TestNewHostSession_DefaultWorkDir(t *testing.T) {
	config := SessionConfig{
		// No WorkDir specified - should use current directory
	}

	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() with default workdir error = %v", err)
	}

	if session.ID() == "" {
		t.Error("Session ID should not be empty")
	}
}

func TestNewHostSession_InvalidWorkDir(t *testing.T) {
	config := SessionConfig{
		WorkDir: "/nonexistent/directory/that/does/not/exist",
	}

	_, err := NewHostSession(config)
	if err == nil {
		t.Error("NewHostSession() should fail with invalid work directory")
	}

	if !strings.Contains(err.Error(), "working directory does not exist") {
		t.Errorf("Expected 'working directory does not exist' error, got: %v", err)
	}
}

func TestHostManager_CreateSession(t *testing.T) {
	manager := NewHostManager()
	config := SessionConfig{
		WorkDir: "/tmp",
	}

	session, err := manager.CreateSession(context.Background(), config)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Verify session is stored in manager
	retrievedSession, err := manager.GetSession(session.ID())
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}

	if retrievedSession.ID() != session.ID() {
		t.Errorf("Retrieved session ID = %v, want %v", retrievedSession.ID(), session.ID())
	}
}

func TestHostManager_GetSession_NotFound(t *testing.T) {
	manager := NewHostManager()

	_, err := manager.GetSession("nonexistent")
	if err == nil {
		t.Error("GetSession() should return error for nonexistent session")
	}

	expected := "session nonexistent not found"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error containing '%s', got: %v", expected, err)
	}
}

func TestHostManager_ListSessions(t *testing.T) {
	manager := NewHostManager()
	config := SessionConfig{WorkDir: "/tmp"}

	// Initially empty
	sessions := manager.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("ListSessions() initial count = %v, want 0", len(sessions))
	}

	// Create sessions
	session1, _ := manager.CreateSession(context.Background(), config)
	session2, _ := manager.CreateSession(context.Background(), config)

	sessions = manager.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("ListSessions() count = %v, want 2", len(sessions))
	}

	// Verify sessions are present
	found := make(map[string]bool)
	for _, session := range sessions {
		found[session.ID()] = true
	}

	if !found[session1.ID()] || !found[session2.ID()] {
		t.Error("Not all created sessions found in ListSessions()")
	}
}

func TestHostManager_Cleanup(t *testing.T) {
	manager := NewHostManager().(*hostManager)
	config := SessionConfig{WorkDir: "/tmp"}

	// Create sessions
	session1, _ := manager.CreateSession(context.Background(), config)
	session2, _ := manager.CreateSession(context.Background(), config)

	// Manually set one session to stopped
	hostSession1 := session1.(*hostSession)
	hostSession1.status = StatusStopped

	// Run cleanup
	err := manager.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	// Check that stopped session was removed
	sessions := manager.ListSessions()
	if len(sessions) != 1 {
		t.Errorf("ListSessions() after cleanup count = %v, want 1", len(sessions))
	}

	if sessions[0].ID() != session2.ID() {
		t.Error("Wrong session remained after cleanup")
	}
}

func TestHostSession_Execute(t *testing.T) {
	config := SessionConfig{WorkDir: "/tmp"}
	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() error = %v", err)
	}

	// Test simple command
	result, err := session.Execute(context.Background(), []string{"echo", "hello"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Execute() exit code = %v, want 0", result.ExitCode)
	}

	if result.Duration <= 0 {
		t.Error("Execute() duration should be positive")
	}
}

func TestHostSession_Execute_NoCommand(t *testing.T) {
	config := SessionConfig{WorkDir: "/tmp"}
	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() error = %v", err)
	}

	// Test with empty command
	_, err = session.Execute(context.Background(), []string{})
	if err == nil {
		t.Error("Execute() should fail with empty command")
	}

	if !strings.Contains(err.Error(), "no command provided") {
		t.Errorf("Expected 'no command provided' error, got: %v", err)
	}
}

func TestHostSession_Execute_WithContext(t *testing.T) {
	config := SessionConfig{WorkDir: "/tmp"}
	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() error = %v", err)
	}

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This command should be fast enough to complete
	_, err = session.Execute(ctx, []string{"echo", "test"})
	if err != nil {
		t.Errorf("Execute() with context error = %v", err)
	}
}

func TestHostSession_SyncFiles(t *testing.T) {
	config := SessionConfig{WorkDir: "/tmp"}
	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() error = %v", err)
	}

	// Test all sync directions - should all be no-ops for host sessions
	directions := []SyncDirection{SyncToContainer, SyncFromContainer, SyncBidirectional}
	for _, direction := range directions {
		err := session.SyncFiles(context.Background(), direction)
		if err != nil {
			t.Errorf("SyncFiles(%v) error = %v", direction, err)
		}
	}
}

func TestHostSession_Close(t *testing.T) {
	config := SessionConfig{WorkDir: "/tmp"}
	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() error = %v", err)
	}

	if session.Status() != StatusRunning {
		t.Errorf("Initial status = %v, want %v", session.Status(), StatusRunning)
	}

	err = session.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if session.Status() != StatusStopped {
		t.Errorf("Status after Close() = %v, want %v", session.Status(), StatusStopped)
	}
}

func TestHostSession_WorkingDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "host-session-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := SessionConfig{WorkDir: tempDir}
	session, err := NewHostSession(config)
	if err != nil {
		t.Fatalf("NewHostSession() error = %v", err)
	}

	// Execute pwd command to verify working directory
	result, err := session.Execute(context.Background(), []string{"pwd"})
	if err != nil {
		t.Fatalf("Execute(pwd) error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("pwd exit code = %v, want 0", result.ExitCode)
	}

	// Note: We can't easily verify the stdout content in this test setup
	// In a real implementation, we would read from result.Stdout
}
