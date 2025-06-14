package container

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// dockerSession implements Session interface using Docker
type dockerSession struct {
	id          string
	containerID string
	client      *client.Client
	config      SessionConfig
	status      SessionStatus
	idGen       IDGenerator
}

// NewDockerSession creates a new Docker-based session
func NewDockerSession(ctx context.Context, cli *client.Client, config SessionConfig) (Session, error) {
	return NewDockerSessionWithIDGen(ctx, cli, config, &DefaultIDGenerator{})
}

// NewDockerSessionWithIDGen creates a new Docker-based session with custom ID generator
func NewDockerSessionWithIDGen(ctx context.Context, cli *client.Client, config SessionConfig, idGen IDGenerator) (Session, error) {
	sessionID := idGen.GenerateID()

	// Create container
	containerConfig := &container.Config{
		Image:      config.Image,
		WorkingDir: config.WorkDir,
		Env:        mapToEnvSlice(config.Environment),
		Tty:        true,
		OpenStdin:  true,
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			CPUShares: config.Limits.CPUShares,
			Memory:    config.Limits.Memory,
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	session := &dockerSession{
		id:          sessionID,
		containerID: resp.ID,
		client:      cli,
		config:      config,
		status:      StatusRunning,
		idGen:       idGen,
	}

	return session, nil
}

func (s *dockerSession) ID() string {
	return s.id
}

func (s *dockerSession) Status() SessionStatus {
	return s.status
}

func (s *dockerSession) Execute(ctx context.Context, cmd []string) (*ExecResult, error) {
	start := time.Now()

	// Create exec configuration
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create exec instance
	execResp, err := s.client.ContainerExecCreate(ctx, s.containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to exec
	attachResp, err := s.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attachResp.Close()

	// Start exec
	if err := s.client.ContainerExecStart(ctx, execResp.ID, container.ExecStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start exec: %w", err)
	}

	// Read output
	stdout, stderr := separateOutput(attachResp.Reader)

	// Get exit code
	inspectResp, err := s.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec: %w", err)
	}

	result := &ExecResult{
		ExitCode: inspectResp.ExitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Duration: time.Since(start),
	}

	return result, nil
}

func (s *dockerSession) SyncFiles(ctx context.Context, direction SyncDirection) error {
	// Basic file sync implementation - will be enhanced in file sync POC
	switch direction {
	case SyncToContainer:
		fmt.Printf("Syncing files to container %s\n", s.containerID)
	case SyncFromContainer:
		fmt.Printf("Syncing files from container %s\n", s.containerID)
	case SyncBidirectional:
		fmt.Printf("Bidirectional sync for container %s\n", s.containerID)
	}
	return nil
}

func (s *dockerSession) Close() error {
	ctx := context.Background()

	// Stop container
	if err := s.client.ContainerStop(ctx, s.containerID, container.StopOptions{}); err != nil {
		fmt.Printf("Warning: failed to stop container %s: %v\n", s.containerID, err)
	}

	// Remove container
	if err := s.client.ContainerRemove(ctx, s.containerID, container.RemoveOptions{Force: true}); err != nil {
		fmt.Printf("Warning: failed to remove container %s: %v\n", s.containerID, err)
	}

	s.status = StatusStopped
	return nil
}
