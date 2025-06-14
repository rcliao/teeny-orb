package container

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/client"
)

// dockerManager implements the Manager interface using Docker
type dockerManager struct {
	client   *client.Client
	sessions map[string]Session
	mutex    sync.RWMutex
}

// NewDockerManager creates a new Docker-based container manager
func NewDockerManager() (Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &dockerManager{
		client:   cli,
		sessions: make(map[string]Session),
	}, nil
}

func (m *dockerManager) CreateSession(ctx context.Context, config SessionConfig) (Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, err := NewDockerSession(ctx, m.client, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker session: %w", err)
	}

	m.sessions[session.ID()] = session
	return session, nil
}

func (m *dockerManager) GetSession(id string) (Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session %s not found", id)
	}

	return session, nil
}

func (m *dockerManager) ListSessions() []Session {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessions := make([]Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func (m *dockerManager) Cleanup(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, session := range m.sessions {
		if session.Status() == StatusStopped || session.Status() == StatusError {
			if err := session.Close(); err != nil {
				fmt.Printf("Error cleaning up session %s: %v\n", id, err)
			}
			delete(m.sessions, id)
		}
	}

	return nil
}

// Moved generateSessionID to docker.go as generateUniqueID
