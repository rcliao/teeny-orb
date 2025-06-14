package container

import "testing"

func TestSessionStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status SessionStatus
		want   string
	}{
		{"creating", StatusCreating, "creating"},
		{"running", StatusRunning, "running"},
		{"stopped", StatusStopped, "stopped"},
		{"error", StatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.status); got != tt.want {
				t.Errorf("SessionStatus = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncDirection_String(t *testing.T) {
	tests := []struct {
		name      string
		direction SyncDirection
		want      string
	}{
		{"to_container", SyncToContainer, "to_container"},
		{"from_container", SyncFromContainer, "from_container"},
		{"bidirectional", SyncBidirectional, "bidirectional"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.direction); got != tt.want {
				t.Errorf("SyncDirection = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SessionConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SessionConfig{
				Image:       "ubuntu:latest",
				WorkDir:     "/workspace",
				ProjectPath: "/host/project",
				Environment: map[string]string{"ENV": "test"},
				Limits: ResourceLimits{
					CPUShares: 1024,
					Memory:    1024 * 1024 * 1024,
				},
			},
			wantErr: false,
		},
		{
			name: "empty image",
			config: SessionConfig{
				Image:   "",
				WorkDir: "/workspace",
			},
			wantErr: true,
		},
		{
			name: "empty workdir",
			config: SessionConfig{
				Image:   "ubuntu:latest",
				WorkDir: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SessionConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
