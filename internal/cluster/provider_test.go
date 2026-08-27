package cluster

import (
	"runtime"
	"testing"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantErr     bool
		wantType    string
		description string
	}{
		{
			name: "kind provider",
			cfg: Config{
				Provider:          "kind",
				Name:              "test-cluster",
				KubernetesVersion: "v1.30.0",
			},
			wantErr:     false,
			wantType:    "*cluster.KindProvider",
			description: "Should create a kind provider",
		},
		{
			name: "k3d provider",
			cfg: Config{
				Provider:          "k3d",
				Name:              "test-cluster",
				KubernetesVersion: "v1.30.0",
			},
			wantErr:     false,
			wantType:    "*cluster.K3dProvider",
			description: "Should create a k3d provider",
		},
		{
			name: "minikube provider",
			cfg: Config{
				Provider:          "minikube",
				Name:              "test-cluster",
				KubernetesVersion: "v1.30.0",
			},
			wantErr:     false,
			wantType:    "*cluster.MinikubeProvider",
			description: "Should create a minikube provider",
		},
		{
			name: "invalid provider",
			cfg: Config{
				Provider:          "invalid",
				Name:              "test-cluster",
				KubernetesVersion: "v1.30.0",
			},
			wantErr:     true,
			description: "Should error on invalid provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("NewProvider() returned nil provider without error")
			}
			if !tt.wantErr && provider.Name() != tt.cfg.Name {
				t.Errorf("Provider.Name() = %v, want %v", provider.Name(), tt.cfg.Name)
			}
		})
	}
}

func TestDetectProvider(t *testing.T) {
	// This test just ensures DetectProvider returns one of the valid providers
	// or an error if none are available
	detected, err := DetectProvider()
	if err != nil {
		// It's OK if no providers are installed
		t.Logf("No cluster provider detected (this is OK in CI): %v", err)
		return
	}

	validProviders := map[string]bool{
		"kind":     true,
		"k3d":      true,
		"minikube": true,
	}

	if !validProviders[detected] {
		t.Errorf("DetectProvider() returned invalid provider: %s", detected)
	}
}

func TestIsCommandAvailable(t *testing.T) {
	// Test with a command that should always exist on the current OS
	shell := "sh"
	if runtime.GOOS == "windows" {
		shell = "cmd"
	}
	if !isCommandAvailable(shell) {
		t.Errorf("isCommandAvailable(%q) should return true", shell)
	}

	// Test with a command that definitely doesn't exist
	if isCommandAvailable("this-command-definitely-does-not-exist-12345") {
		t.Error("isCommandAvailable() should return false for non-existent command")
	}
}

func TestProviderName(t *testing.T) {
	providers := []struct {
		name     string
		provider Provider
	}{
		{
			name:     "kind",
			provider: NewKindProvider("test", "v1.30.0"),
		},
		{
			name:     "k3d",
			provider: NewK3dProvider("test", "v1.30.0"),
		},
		{
			name:     "minikube",
			provider: NewMinikubeProvider("test", "v1.30.0"),
		},
	}

	for _, tt := range providers {
		t.Run(tt.name, func(t *testing.T) {
			if tt.provider.Name() != "test" {
				t.Errorf("%s provider Name() = %v, want 'test'", tt.name, tt.provider.Name())
			}
		})
	}
}
