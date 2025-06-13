package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/rcliao/teeny-orb/internal/container"
)

func NewSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage container sessions",
		Long:  "Create and manage isolated container sessions for coding.",
	}

	cmd.AddCommand(newSessionCreateCmd())
	cmd.AddCommand(newSessionListCmd())
	cmd.AddCommand(newSessionStopCmd())

	return cmd
}

func newSessionCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new container session",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := container.NewDockerManager()
			if err != nil {
				return fmt.Errorf("failed to create container manager: %w", err)
			}

			config := container.SessionConfig{
				Image:   "alpine:latest",
				WorkDir: "/workspace",
				Environment: map[string]string{
					"TERM": "xterm-256color",
				},
				Limits: container.ResourceLimits{
					CPUShares: 512,
					Memory:    536870912, // 512MB
				},
			}

			session, err := manager.CreateSession(context.Background(), config)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			fmt.Printf("Created session: %s\n", session.ID())
			fmt.Printf("Status: %s\n", session.Status())
			return nil
		},
	}
}

func newSessionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := container.NewDockerManager()
			if err != nil {
				return fmt.Errorf("failed to create container manager: %w", err)
			}

			sessions := manager.ListSessions()
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

			manager, err := container.NewDockerManager()
			if err != nil {
				return fmt.Errorf("failed to create container manager: %w", err)
			}

			session, err := manager.GetSession(sessionID)
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