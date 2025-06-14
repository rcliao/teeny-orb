package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewGenerateCmd(t *testing.T) {
	cmd := NewGenerateCmd()

	if cmd.Use != "generate [prompt]" {
		t.Errorf("Generate command Use = %v, want 'generate [prompt]'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Generate command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Generate command should have a long description")
	}
}

func TestGenerateCmd_RequiresArgument(t *testing.T) {
	cmd := NewGenerateCmd()

	// Test without arguments - should fail
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Error("Generate command should require an argument")
	}

	// Verify it's the right kind of error
	if !strings.Contains(err.Error(), "arg") && !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected argument error, got: %v", err)
	}
}

func TestGenerateCmd_WithArgument(t *testing.T) {
	cmd := NewGenerateCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)

	// Test with valid argument
	cmd.SetArgs([]string{"create a hello world function"})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("Generate command should not error with valid argument: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "create a hello world function") {
		t.Errorf("Output should contain the prompt, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Phase 2") {
		t.Errorf("Output should mention Phase 2 implementation, got: %s", outputStr)
	}
}

func TestGenerateCmd_MultipleArgs(t *testing.T) {
	cmd := NewGenerateCmd()

	// Test with multiple arguments - should fail due to ExactArgs(1)
	cmd.SetArgs([]string{"create", "hello", "world"})
	err := cmd.Execute()

	if err == nil {
		t.Error("Generate command should only accept exactly one argument")
	}
}

func TestGenerateCmd_EmptyArgument(t *testing.T) {
	cmd := NewGenerateCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)

	// Test with empty string argument
	cmd.SetArgs([]string{""})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("Generate command should accept empty string: %v", err)
	}

	// Should still produce output
	if output.String() == "" {
		t.Error("Generate command should produce output even with empty prompt")
	}
}

func TestGenerateCmd_LongPrompt(t *testing.T) {
	cmd := NewGenerateCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)

	// Test with a long prompt
	longPrompt := strings.Repeat("create a complex function that does many things ", 10)
	cmd.SetArgs([]string{longPrompt})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("Generate command should handle long prompts: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "create a complex function") {
		t.Errorf("Output should contain part of the long prompt, got: %s", outputStr)
	}
}
