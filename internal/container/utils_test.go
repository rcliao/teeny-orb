package container

import (
	"strings"
	"testing"
)

func TestDefaultIDGenerator_GenerateID(t *testing.T) {
	gen := &DefaultIDGenerator{}

	id1 := gen.GenerateID()
	id2 := gen.GenerateID()

	if id1 == id2 {
		t.Error("GenerateID should return unique IDs")
	}

	if !strings.HasPrefix(id1, "teeny-orb-") {
		t.Errorf("Generated ID should have prefix 'teeny-orb-', got %s", id1)
	}
}

func TestStaticIDGenerator_GenerateID(t *testing.T) {
	gen := NewStaticIDGenerator("test")

	id1 := gen.GenerateID()
	id2 := gen.GenerateID()

	if id1 != "test-1" {
		t.Errorf("First ID = %s, want test-1", id1)
	}

	if id2 != "test-2" {
		t.Errorf("Second ID = %s, want test-2", id2)
	}
}

func TestMapToEnvSlice(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want []string
	}{
		{
			name: "empty map",
			env:  map[string]string{},
			want: []string{},
		},
		{
			name: "single entry",
			env:  map[string]string{"KEY": "value"},
			want: []string{"KEY=value"},
		},
		{
			name: "multiple entries",
			env:  map[string]string{"KEY1": "value1", "KEY2": "value2"},
			want: []string{"KEY1=value1", "KEY2=value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapToEnvSlice(tt.env)

			if len(got) != len(tt.want) {
				t.Errorf("mapToEnvSlice() length = %v, want %v", len(got), len(tt.want))
				return
			}

			// Convert to map for comparison since order may vary
			wantMap := make(map[string]bool)
			for _, item := range tt.want {
				wantMap[item] = true
			}

			for _, item := range got {
				if !wantMap[item] {
					t.Errorf("mapToEnvSlice() contains unexpected item: %s", item)
				}
			}
		})
	}
}

func TestSeparateOutput(t *testing.T) {
	input := strings.NewReader("test output")

	stdout, stderr := separateOutput(input)

	if stdout == nil {
		t.Error("stdout should not be nil")
	}

	if stderr == nil {
		t.Error("stderr should not be nil")
	}

	// Verify stdout contains the input
	stdoutData := make([]byte, 11)
	n, err := stdout.Read(stdoutData)
	if err != nil {
		t.Errorf("Reading stdout failed: %v", err)
	}

	if string(stdoutData[:n]) != "test output" {
		t.Errorf("stdout content = %s, want 'test output'", string(stdoutData[:n]))
	}
}
