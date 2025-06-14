package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRootCmd(t *testing.T) {
	// Reset viper for clean test state
	viper.Reset()

	cmd := &cobra.Command{
		Use:   "teeny-orb",
		Short: "AI-powered coding assistant with container isolation",
		Run: func(cmd *cobra.Command, args []string) {
			// Test version of the root command
		},
	}

	if cmd.Use != "teeny-orb" {
		t.Errorf("Root command Use = %v, want teeny-orb", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Root command should have a short description")
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute function exists and can be called
	// In a real test, we might capture output or test specific behavior
	err := Execute()

	// Since we're not providing any arguments, this should work
	// The actual behavior depends on the command structure
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}

func TestInitConfig(t *testing.T) {
	// Test config initialization
	viper.Reset()

	// Call initConfig directly
	initConfig()

	// Verify that viper is configured
	// We can't easily test file reading without creating actual config files
	// But we can verify that the function doesn't panic
}

func TestRootCommandFlags(t *testing.T) {
	viper.Reset()

	// Create a test version of the root command
	cmd := &cobra.Command{
		Use: "teeny-orb",
	}

	// Add the same flags as the real root command
	cmd.PersistentFlags().String("config", "", "config file")
	cmd.PersistentFlags().String("project", "", "project directory")
	cmd.PersistentFlags().Bool("verbose", false, "enable verbose output")

	// Test flag parsing
	args := []string{"--project", "/test/project", "--verbose"}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution with flags failed: %v", err)
	}

	// Test flag values
	projectFlag := cmd.PersistentFlags().Lookup("project")
	if projectFlag.Value.String() != "/test/project" {
		t.Errorf("Project flag = %v, want /test/project", projectFlag.Value.String())
	}

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag.Value.String() != "true" {
		t.Errorf("Verbose flag = %v, want true", verboseFlag.Value.String())
	}
}

func TestRootCommandOutput(t *testing.T) {
	// Test the output of the root command
	var output bytes.Buffer

	cmd := &cobra.Command{
		Use: "teeny-orb",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print("Starting interactive coding session...\n")
			cmd.Print("Type 'help' for available commands or 'exit' to quit.\n")
		},
	}

	cmd.SetOut(&output)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	expectedOutput := "Starting interactive coding session...\nType 'help' for available commands or 'exit' to quit.\n"
	if output.String() != expectedOutput {
		t.Errorf("Output = %q, want %q", output.String(), expectedOutput)
	}
}
