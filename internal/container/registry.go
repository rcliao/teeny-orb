package container

import (
	"fmt"
	"sync"
)

// ManagerRegistry manages both host and Docker session managers
type ManagerRegistry struct {
	hostManager   Manager
	dockerManager Manager
	mutex         sync.RWMutex
}

var (
	registry     *ManagerRegistry
	registryOnce sync.Once
)

// GetRegistry returns the singleton manager registry
func GetRegistry() *ManagerRegistry {
	registryOnce.Do(func() {
		registry = &ManagerRegistry{
			hostManager: NewHostManager(),
		}
	})
	return registry
}

// GetDockerManager returns the Docker manager, creating it if necessary
func (r *ManagerRegistry) GetDockerManager() (Manager, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.dockerManager == nil {
		manager, err := NewDockerManager()
		if err != nil {
			return nil, err
		}
		r.dockerManager = manager
	}

	return r.dockerManager, nil
}

// GetHostManager returns the host manager
func (r *ManagerRegistry) GetHostManager() Manager {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.hostManager
}

// GetAllSessions returns sessions from both managers
func (r *ManagerRegistry) GetAllSessions() []Session {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var allSessions []Session

	// Get host sessions
	hostSessions := r.hostManager.ListSessions()
	allSessions = append(allSessions, hostSessions...)

	// Get Docker sessions if manager exists
	if r.dockerManager != nil {
		dockerSessions := r.dockerManager.ListSessions()
		allSessions = append(allSessions, dockerSessions...)
	}

	return allSessions
}

// GetSession returns a session by ID from either manager
func (r *ManagerRegistry) GetSession(id string) (Session, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Try host manager first
	if session, err := r.hostManager.GetSession(id); err == nil {
		return session, nil
	}

	// Try Docker manager if it exists
	if r.dockerManager != nil {
		if session, err := r.dockerManager.GetSession(id); err == nil {
			return session, nil
		}
	}

	return nil, fmt.Errorf("session %s not found", id)
}
