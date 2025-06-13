package container

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// mockSession is a temporary implementation for testing
// Will be replaced with real Docker integration
type mockSession struct {
	id     string
	status SessionStatus
	config SessionConfig
}

func (s *mockSession) ID() string {
	return s.id
}

func (s *mockSession) Status() SessionStatus {
	return s.status
}

func (s *mockSession) Execute(ctx context.Context, cmd []string) (*ExecResult, error) {
	// Mock execution - just echo the command
	result := &ExecResult{
		ExitCode: 0,
		Stdout:   strings.NewReader(fmt.Sprintf("Mock execution: %s", strings.Join(cmd, " "))),
		Stderr:   strings.NewReader(""),
		Duration: time.Millisecond * 100,
	}
	
	return result, nil
}

func (s *mockSession) SyncFiles(ctx context.Context, direction SyncDirection) error {
	fmt.Printf("Mock file sync: %s for session %s\n", direction, s.id)
	return nil
}

func (s *mockSession) Close() error {
	s.status = StatusStopped
	fmt.Printf("Session %s closed\n", s.id)
	return nil
}