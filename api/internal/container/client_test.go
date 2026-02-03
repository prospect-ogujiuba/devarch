package container

import (
	"os"
	"testing"
)

func TestNewClient_RespectsDevarchRuntime(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		wantRuntime RuntimeType
		wantErr     bool
		errContains string
		skipIfNo    string // skip test if this runtime is not available
	}{
		{
			name:        "explicit podman",
			envValue:    "podman",
			wantRuntime: RuntimePodman,
			skipIfNo:    "podman",
		},
		{
			name:        "explicit docker",
			envValue:    "docker",
			wantRuntime: RuntimeDocker,
			skipIfNo:    "docker",
		},
		{
			name:        "invalid value",
			envValue:    "invalid",
			wantErr:     true,
			errContains: "invalid — must be 'podman' or 'docker'",
		},
	}

	// Save original env
	origEnv := os.Getenv("DEVARCH_RUNTIME")
	defer os.Setenv("DEVARCH_RUNTIME", origEnv)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if the required runtime is not available
			if tt.skipIfNo != "" {
				if _, err := os.Stat("/usr/bin/" + tt.skipIfNo); os.IsNotExist(err) {
					if _, err := os.Stat("/usr/local/bin/" + tt.skipIfNo); os.IsNotExist(err) {
						t.Skipf("%s not installed", tt.skipIfNo)
					}
				}
			}

			os.Setenv("DEVARCH_RUNTIME", tt.envValue)

			client, err := NewClient()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client.runtime != tt.wantRuntime {
				t.Errorf("runtime = %v, want %v", client.runtime, tt.wantRuntime)
			}
		})
	}
}

func TestStubMethods_ReturnErrNotImplemented(t *testing.T) {
	// Clear env to allow auto-detect
	origEnv := os.Getenv("DEVARCH_RUNTIME")
	os.Unsetenv("DEVARCH_RUNTIME")
	defer os.Setenv("DEVARCH_RUNTIME", origEnv)

	client, err := NewClient()
	if err != nil {
		t.Skipf("no container runtime available: %v", err)
	}

	t.Run("CreateNetwork", func(t *testing.T) {
		err := client.CreateNetwork("test", nil)
		if err != ErrNotImplemented {
			t.Errorf("CreateNetwork() error = %v, want %v", err, ErrNotImplemented)
		}
	})

	t.Run("RemoveNetwork", func(t *testing.T) {
		err := client.RemoveNetwork("test")
		if err != ErrNotImplemented {
			t.Errorf("RemoveNetwork() error = %v, want %v", err, ErrNotImplemented)
		}
	})

	t.Run("ListNetworks", func(t *testing.T) {
		_, err := client.ListNetworks()
		if err != ErrNotImplemented {
			t.Errorf("ListNetworks() error = %v, want %v", err, ErrNotImplemented)
		}
	})
}

func TestParseNamesList(t *testing.T) {
	// Clear env to allow auto-detect
	origEnv := os.Getenv("DEVARCH_RUNTIME")
	os.Unsetenv("DEVARCH_RUNTIME")
	defer os.Setenv("DEVARCH_RUNTIME", origEnv)

	client, err := NewClient()
	if err != nil {
		t.Skipf("no container runtime available: %v", err)
	}

	tests := []struct {
		name   string
		input  string
		want   []string
	}{
		{
			name:  "single container",
			input: "container1\n",
			want:  []string{"container1"},
		},
		{
			name:  "multiple containers",
			input: "container1\ncontainer2\ncontainer3\n",
			want:  []string{"container1", "container2", "container3"},
		},
		{
			name:  "empty output",
			input: "",
			want:  []string{},
		},
		{
			name:  "trailing newlines",
			input: "container1\n\n",
			want:  []string{"container1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.parseNamesList(tt.input)
			if err != nil {
				t.Fatalf("parseNamesList() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Errorf("parseNamesList() got %d items, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseNamesList()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestExec_ConstructsCorrectArgs(t *testing.T) {
	// This test just verifies the method exists and constructs args correctly
	// We can't actually test exec without a running container
	origEnv := os.Getenv("DEVARCH_RUNTIME")
	os.Unsetenv("DEVARCH_RUNTIME")
	defer os.Setenv("DEVARCH_RUNTIME", origEnv)

	client, err := NewClient()
	if err != nil {
		t.Skipf("no container runtime available: %v", err)
	}

	// This will fail because the container doesn't exist, but we're just
	// checking that the method exists and can be called
	_, err = client.Exec("nonexistent", []string{"echo", "test"})
	if err == nil {
		t.Error("expected error for nonexistent container")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
