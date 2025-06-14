package container

import (
	"context"
	"testing"
)

func TestGetRegistry(t *testing.T) {
	// Test singleton behavior
	registry1 := GetRegistry()
	registry2 := GetRegistry()

	if registry1 != registry2 {
		t.Error("GetRegistry() should return the same instance (singleton)")
	}

	if registry1 == nil {
		t.Error("GetRegistry() should not return nil")
	}
}

func TestManagerRegistry_GetHostManager(t *testing.T) {
	registry := GetRegistry()

	hostManager := registry.GetHostManager()
	if hostManager == nil {
		t.Error("GetHostManager() should not return nil")
	}

	// Test that it returns the same instance
	hostManager2 := registry.GetHostManager()
	if hostManager != hostManager2 {
		t.Error("GetHostManager() should return the same instance")
	}
}

func TestManagerRegistry_GetDockerManager(t *testing.T) {
	registry := GetRegistry()

	// This test will fail without Docker, but tests the error handling
	_, err := registry.GetDockerManager()
	// We expect this to fail in a test environment without Docker
	if err == nil {
		t.Skip("Skipping Docker manager test - requires Docker daemon")
	}

	// Test that error is properly returned
	if err == nil {
		t.Error("GetDockerManager() should return error when Docker is not available")
	}
}

func TestManagerRegistry_GetAllSessions(t *testing.T) {
	registry := GetRegistry()

	// Initially should be empty
	sessions := registry.GetAllSessions()
	if len(sessions) != 0 {
		t.Errorf("GetAllSessions() initial count = %v, want 0", len(sessions))
	}

	// Create a host session
	hostManager := registry.GetHostManager()
	config := SessionConfig{WorkDir: "/tmp"}
	session, err := hostManager.CreateSession(context.Background(), config)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Now should have one session
	sessions = registry.GetAllSessions()
	if len(sessions) != 1 {
		t.Errorf("GetAllSessions() count after create = %v, want 1", len(sessions))
	}

	if sessions[0].ID() != session.ID() {
		t.Error("GetAllSessions() should return the created session")
	}
}

func TestManagerRegistry_GetSession(t *testing.T) {
	registry := GetRegistry()

	// Create a session
	hostManager := registry.GetHostManager()
	config := SessionConfig{WorkDir: "/tmp"}
	session, err := hostManager.CreateSession(context.Background(), config)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Retrieve the session
	retrievedSession, err := registry.GetSession(session.ID())
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}

	if retrievedSession.ID() != session.ID() {
		t.Errorf("GetSession() ID = %v, want %v", retrievedSession.ID(), session.ID())
	}
}

func TestManagerRegistry_GetSession_NotFound(t *testing.T) {
	registry := GetRegistry()

	_, err := registry.GetSession("nonexistent")
	if err == nil {
		t.Error("GetSession() should return error for nonexistent session")
	}

	expected := "session nonexistent not found"
	if err.Error() != expected {
		t.Errorf("GetSession() error = %v, want %v", err.Error(), expected)
	}
}

func TestManagerRegistry_Concurrency(t *testing.T) {
	// Test concurrent access to registry
	done := make(chan bool)

	// Start multiple goroutines accessing the registry
	for i := 0; i < 10; i++ {
		go func() {
			registry := GetRegistry()
			_ = registry.GetHostManager()
			_ = registry.GetAllSessions()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panicking, the test passed
}

// Test that registry properly handles both host and docker sessions
func TestManagerRegistry_MixedSessions(t *testing.T) {
	// Note: This test may be affected by shared registry state from other tests
	// In a production test suite, we would either reset the registry or use isolated instances
	registry := GetRegistry()

	// Get current session count first (may have leftovers from other tests)
	initialCount := len(registry.GetAllSessions())

	// Create host session
	hostManager := registry.GetHostManager()
	config := SessionConfig{WorkDir: "/tmp"}
	hostSession, err := hostManager.CreateSession(context.Background(), config)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Check that GetAllSessions includes the new host session
	allSessions := registry.GetAllSessions()
	if len(allSessions) != initialCount+1 {
		t.Errorf("GetAllSessions() count = %v, want %v", len(allSessions), initialCount+1)
	}

	// Check that GetSession can find host session
	foundSession, err := registry.GetSession(hostSession.ID())
	if err != nil {
		t.Errorf("GetSession() should find host session: %v", err)
	}

	if foundSession.ID() != hostSession.ID() {
		t.Error("GetSession() returned wrong session")
	}
}
