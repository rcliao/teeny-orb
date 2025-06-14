package container

import (
	"testing"
)

// MockDockerClient implements basic docker client interface for testing
// In a real implementation, we would mock the full Docker client interface
type MockDockerClient struct {
	createError error
	startError  error
	stopError   error
	removeError error
	execError   error
	containerID string
}

func TestDockerSession_ID(t *testing.T) {
	idGen := NewStaticIDGenerator("test")
	session := &dockerSession{
		id:    idGen.GenerateID(),
		idGen: idGen,
	}

	if session.ID() != "test-1" {
		t.Errorf("ID() = %v, want test-1", session.ID())
	}
}

func TestDockerSession_Status(t *testing.T) {
	session := &dockerSession{
		status: StatusRunning,
	}

	if session.Status() != StatusRunning {
		t.Errorf("Status() = %v, want %v", session.Status(), StatusRunning)
	}
}

func TestDockerSession_Close(t *testing.T) {
	// This test focuses on the status change behavior
	// We can't test the Docker client interaction without mocking
	t.Skip("Docker session Close() requires Docker client mock - skipping in unit tests")
}

func TestNewDockerSessionWithIDGen(t *testing.T) {
	// This test would require a mock Docker client for proper testing
	t.Skip("NewDockerSessionWithIDGen requires Docker client mock - skipping in unit tests")
}

// Integration test structure - would need real Docker for full testing
func TestDockerSession_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would be an integration test that requires Docker to be running
	// For now, we'll skip it to focus on unit tests
	t.Skip("Integration tests require Docker daemon")
}

// Benchmark for ID generation
func BenchmarkDefaultIDGenerator(b *testing.B) {
	gen := &DefaultIDGenerator{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.GenerateID()
	}
}

func BenchmarkStaticIDGenerator(b *testing.B) {
	gen := NewStaticIDGenerator("bench")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.GenerateID()
	}
}
