package commands

import (
	"github.com/spf13/cobra"
)

func NewReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review [file]",
		Short: "Review code file for improvements",
		Long:  "Analyze existing code for improvements and suggestions.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			cmd.Printf("Reviewing file: %s\n", file)
			cmd.Println("(This feature will be implemented in Phase 2: LLM Integration)")
			return nil
		},
	}

	return cmd
}
