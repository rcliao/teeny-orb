package container

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// FileSyncer handles bidirectional file synchronization between host and container
type FileSyncer struct {
	client      *client.Client
	containerID string
	hostPath    string
	containerPath string
}

// NewFileSyncer creates a new file synchronizer
func NewFileSyncer(client *client.Client, containerID, hostPath, containerPath string) *FileSyncer {
	return &FileSyncer{
		client:        client,
		containerID:   containerID,
		hostPath:      hostPath,
		containerPath: containerPath,
	}
}

// SyncToContainer copies files from host to container
func (fs *FileSyncer) SyncToContainer(ctx context.Context) error {
	// Create tar archive of host directory
	tarBuffer, err := fs.createTarFromHost()
	if err != nil {
		return fmt.Errorf("failed to create tar from host: %w", err)
	}

	// Copy tar to container  
	err = fs.client.CopyToContainer(ctx, fs.containerID, fs.containerPath, tarBuffer, container.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	return nil
}

// SyncFromContainer copies files from container to host
func (fs *FileSyncer) SyncFromContainer(ctx context.Context) error {
	// Get tar archive from container
	reader, _, err := fs.client.CopyFromContainer(ctx, fs.containerID, fs.containerPath)
	if err != nil {
		return fmt.Errorf("failed to copy from container: %w", err)
	}
	defer reader.Close()

	// Extract tar to host
	err = fs.extractTarToHost(reader)
	if err != nil {
		return fmt.Errorf("failed to extract tar to host: %w", err)
	}

	return nil
}

// createTarFromHost creates a tar archive from the host directory
func (fs *FileSyncer) createTarFromHost() (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	err := filepath.Walk(fs.hostPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories for this POC
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path from host path
		relPath, err := filepath.Rel(fs.hostPath, path)
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// extractTarToHost extracts a tar archive to the host directory
func (fs *FileSyncer) extractTarToHost(reader io.Reader) error {
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Skip hidden files for this POC
		if strings.HasPrefix(header.Name, ".") {
			continue
		}

		targetPath := filepath.Join(fs.hostPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create file
			dir := filepath.Dir(targetPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			_, err = io.Copy(file, tr)
			file.Close()
			if err != nil {
				return err
			}
		default:
			fmt.Printf("Skipping unsupported file type: %c for %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

// Enhanced Docker session with file sync
func (s *dockerSession) SyncFilesAdvanced(ctx context.Context, direction SyncDirection) error {
	if s.config.ProjectPath == "" {
		return fmt.Errorf("no project path configured for file sync")
	}

	syncer := NewFileSyncer(s.client, s.containerID, s.config.ProjectPath, s.config.WorkDir)

	switch direction {
	case SyncToContainer:
		return syncer.SyncToContainer(ctx)
	case SyncFromContainer:
		return syncer.SyncFromContainer(ctx)
	case SyncBidirectional:
		// For POC, just sync to container
		// In production, would implement change detection
		return syncer.SyncToContainer(ctx)
	default:
		return fmt.Errorf("unsupported sync direction: %s", direction)
	}
}