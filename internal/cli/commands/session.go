package commands

import (
	"context"
	"fmt"

	"github.com/rcliao/teeny-orb/internal/container"
	"github.com/spf13/cobra"
)

func NewSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage sessions",
		Long:  "Create and manage sessions for coding (host-based by default, containerized with --docker).",
	}

	cmd.AddCommand(newSessionCreateCmd())
	cmd.AddCommand(newSessionListCmd())
	cmd.AddCommand(newSessionStopCmd())

	return cmd
}

func newSessionCreateCmd() *cobra.Command {
	var useDocker bool
	var workDir string
	var image string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new session",
		Long:  "Create a new session for coding (runs on host by default, use --docker for containerized execution)",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := container.GetRegistry()
			var manager container.Manager
			var err error

			if useDocker {
				manager, err = registry.GetDockerManager()
				if err != nil {
					return fmt.Errorf("failed to get Docker manager: %w", err)
				}
			} else {
				manager = registry.GetHostManager()
			}

			config := container.SessionConfig{
				WorkDir: workDir,
				Environment: map[string]string{
					"TERM": "xterm-256color",
				},
			}

			// Docker-specific configuration
			if useDocker {
				config.Image = image
				config.Limits = container.ResourceLimits{
					CPUShares: 512,
					Memory:    536870912, // 512MB
				}
			}

			session, err := manager.CreateSession(context.Background(), config)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			sessionType := "host"
			if useDocker {
				sessionType = "container"
			}

			fmt.Printf("Created %s session: %s\n", sessionType, session.ID())
			fmt.Printf("Status: %s\n", session.Status())
			if !useDocker && workDir != "" {
				fmt.Printf("Working directory: %s\n", workDir)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&useDocker, "docker", false, "Use Docker containers for session isolation")
	cmd.Flags().StringVar(&workDir, "workdir", "", "Working directory for the session (defaults to current directory for host sessions)")
	cmd.Flags().StringVar(&image, "image", "alpine:latest", "Docker image to use (only applies when --docker is set)")

	return cmd
}

func newSessionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := container.GetRegistry()
			sessions := registry.GetAllSessions()

			if len(sessions) == 0 {
				fmt.Println("No active sessions")
				return nil
			}

			fmt.Println("Active sessions:")
			for _, session := range sessions {
				fmt.Printf("  ID: %s, Status: %s\n", session.ID(), session.Status())
			}
			return nil
		},
	}
}

func newSessionStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [session-id]",
		Short: "Stop a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			registry := container.GetRegistry()
			session, err := registry.GetSession(sessionID)
			if err != nil {
				return fmt.Errorf("session not found: %w", err)
			}

			if err := session.Close(); err != nil {
				return fmt.Errorf("failed to stop session: %w", err)
			}

			fmt.Printf("Session %s stopped\n", sessionID)
			return nil
		},
	}
}
