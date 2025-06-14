package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewReviewCmd(t *testing.T) {
	cmd := NewReviewCmd()

	if cmd.Use != "review [file]" {
		t.Errorf("Review command Use = %v, want 'review [file]'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Review command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Review command should have a long description")
	}
}

func TestReviewCmd_RequiresArgument(t *testing.T) {
	cmd := NewReviewCmd()

	// Test without arguments - should fail
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err == nil {
		t.Error("Review command should require a file argument")
	}

	// Verify it's the right kind of error
	if !strings.Contains(err.Error(), "arg") && !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected argument error, got: %v", err)
	}
}

func TestReviewCmd_WithArgument(t *testing.T) {
	cmd := NewReviewCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)

	// Test with valid file argument
	cmd.SetArgs([]string{"main.go"})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("Review command should not error with valid argument: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "main.go") {
		t.Errorf("Output should contain the filename, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Phase 2") {
		t.Errorf("Output should mention Phase 2 implementation, got: %s", outputStr)
	}
}

func TestReviewCmd_MultipleArgs(t *testing.T) {
	cmd := NewReviewCmd()

	// Test with multiple arguments - should fail due to ExactArgs(1)
	cmd.SetArgs([]string{"file1.go", "file2.go"})
	err := cmd.Execute()

	if err == nil {
		t.Error("Review command should only accept exactly one argument")
	}
}

func TestReviewCmd_DifferentFileTypes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"Go file", "main.go"},
		{"Python file", "script.py"},
		{"JavaScript file", "app.js"},
		{"Text file", "README.md"},
		{"No extension", "Dockerfile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewReviewCmd()
			var output bytes.Buffer
			cmd.SetOut(&output)

			cmd.SetArgs([]string{tt.filename})
			err := cmd.Execute()

			if err != nil {
				t.Errorf("Review command should handle %s: %v", tt.filename, err)
			}

			outputStr := output.String()
			if !strings.Contains(outputStr, tt.filename) {
				t.Errorf("Output should contain the filename %s, got: %s", tt.filename, outputStr)
			}
		})
	}
}

func TestReviewCmd_WithPath(t *testing.T) {
	cmd := NewReviewCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)

	// Test with file path
	cmd.SetArgs([]string{"/path/to/file.go"})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("Review command should handle file paths: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "/path/to/file.go") {
		t.Errorf("Output should contain the file path, got: %s", outputStr)
	}
}
