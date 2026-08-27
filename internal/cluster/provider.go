package cluster

import (
	"context"
	"fmt"
	"os/exec"
)

// Provider defines the interface for cluster providers (kind, k3d, etc.)
type Provider interface {
	// Up creates or ensures the cluster exists
	Up(ctx context.Context) error

	// Down deletes the cluster
	Down(ctx context.Context) error

	// Exists checks if the cluster already exists
	Exists(ctx context.Context) (bool, error)

	// KubeconfigPath returns the path to the kubeconfig file
	KubeconfigPath(ctx context.Context) (string, error)

	// Name returns the cluster name
	Name() string
}

// Config holds configuration for creating a cluster provider
type Config struct {
	Provider          string
	Name              string
	KubernetesVersion string
}

// NewProvider creates a new cluster provider based on the config
func NewProvider(cfg Config) (Provider, error) {
	providerName := cfg.Provider

	// Auto-detect provider if set to "auto"
	if providerName == "auto" || providerName == "" {
		detected, err := DetectProvider()
		if err != nil {
			return nil, fmt.Errorf("auto-detecting provider: %w", err)
		}
		providerName = detected
	}

	switch providerName {
	case "kind":
		return NewKindProvider(cfg.Name, cfg.KubernetesVersion), nil
	case "k3d":
		return NewK3dProvider(cfg.Name, cfg.KubernetesVersion), nil
	case "minikube":
		return NewMinikubeProvider(cfg.Name, cfg.KubernetesVersion), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s (supported: kind, k3d, minikube, auto)", providerName)
	}
}

// DetectProvider attempts to auto-detect which cluster provider is available
// It checks in order: kind, k3d, minikube
func DetectProvider() (string, error) {
	// Check for kind
	if isCommandAvailable("kind") {
		return "kind", nil
	}

	// Check for k3d
	if isCommandAvailable("k3d") {
		return "k3d", nil
	}

	// Check for minikube
	if isCommandAvailable("minikube") {
		return "minikube", nil
	}

	return "", fmt.Errorf("no supported cluster provider found (checked: kind, k3d, minikube)")
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
