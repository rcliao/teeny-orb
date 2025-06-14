package container

import (
	"context"
	"fmt"
	"testing"
)

func TestMockManager_CreateSession(t *testing.T) {
	manager := NewMockManager()
	config := SessionConfig{
		Image:   "ubuntu:latest",
		WorkDir: "/workspace",
	}

	session, err := manager.CreateSession(context.Background(), config)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if session.ID() == "" {
		t.Error("Session ID should not be empty")
	}

	if session.Status() != StatusRunning {
		t.Errorf("Session status = %v, want %v", session.Status(), StatusRunning)
	}
}

func TestMockManager_CreateSession_Error(t *testing.T) {
	manager := NewMockManager()
	expectedErr := fmt.Errorf("create error")
	manager.SetCreateError(expectedErr)

	config := SessionConfig{
		Image:   "ubuntu:latest",
		WorkDir: "/workspace",
	}

	_, err := manager.CreateSession(context.Background(), config)
	if err == nil {
		t.Fatal("CreateSession() should return error")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("CreateSession() error = %v, want %v", err, expectedErr)
	}
}

func TestMockManager_GetSession(t *testing.T) {
	manager := NewMockManager()
	config := SessionConfig{
		Image:   "ubuntu:latest",
		WorkDir: "/workspace",
	}

	createdSession, err := manager.CreateSession(context.Background(), config)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	retrievedSession, err := manager.GetSession(createdSession.ID())
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}

	if retrievedSession.ID() != createdSession.ID() {
		t.Errorf("GetSession() ID = %v, want %v", retrievedSession.ID(), createdSession.ID())
	}
}

func TestMockManager_GetSession_NotFound(t *testing.T) {
	manager := NewMockManager()

	_, err := manager.GetSession("nonexistent")
	if err == nil {
		t.Fatal("GetSession() should return error for nonexistent session")
	}

	expected := "session nonexistent not found"
	if err.Error() != expected {
		t.Errorf("GetSession() error = %v, want %v", err.Error(), expected)
	}
}

func TestMockManager_ListSessions(t *testing.T) {
	manager := NewMockManager()
	config := SessionConfig{
		Image:   "ubuntu:latest",
		WorkDir: "/workspace",
	}

	// Initially empty
	sessions := manager.ListSessions()
	if len(sessions) != 0 {
		t.Errorf("ListSessions() count = %v, want 0", len(sessions))
	}

	// Create sessions
	session1, _ := manager.CreateSession(context.Background(), config)
	session2, _ := manager.CreateSession(context.Background(), config)

	sessions = manager.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("ListSessions() count = %v, want 2", len(sessions))
	}

	// Verify sessions are present
	found1, found2 := false, false
	for _, session := range sessions {
		if session.ID() == session1.ID() {
			found1 = true
		}
		if session.ID() == session2.ID() {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Not all created sessions found in ListSessions()")
	}
}

func TestMockManager_Cleanup(t *testing.T) {
	manager := NewMockManager()
	config := SessionConfig{
		Image:   "ubuntu:latest",
		WorkDir: "/workspace",
	}

	// Create sessions
	session1, _ := manager.CreateSession(context.Background(), config)
	session2, _ := manager.CreateSession(context.Background(), config)

	// Set one session to stopped
	mockSession1 := session1.(*MockSession)
	mockSession1.SetStatus(StatusStopped)

	// Run cleanup
	err := manager.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if !manager.WasCleanupCalled() {
		t.Error("Cleanup was not called")
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
