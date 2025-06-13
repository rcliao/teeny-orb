package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [prompt]",
		Short: "Generate code from natural language description",
		Long:  "Generate code from a natural language description using AI assistance.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]
			fmt.Printf("Generating code for: %s\n", prompt)
			fmt.Println("(This feature will be implemented in Phase 2: LLM Integration)")
			return nil
		},
	}

	return cmd
}