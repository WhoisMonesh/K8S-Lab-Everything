package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MinikubeProvider implements the Provider interface for minikube clusters
type MinikubeProvider struct {
	name       string
	k8sVersion string
}

// NewMinikubeProvider creates a new minikube cluster provider
func NewMinikubeProvider(name, k8sVersion string) *MinikubeProvider {
	return &MinikubeProvider{
		name:       name,
		k8sVersion: k8sVersion,
	}
}

// Name returns the cluster name
func (m *MinikubeProvider) Name() string {
	return m.name
}

// Up creates the minikube cluster
func (m *MinikubeProvider) Up(ctx context.Context) error {
	exists, err := m.Exists(ctx)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if exists {
		// Start the cluster if it's stopped
		cmd := exec.CommandContext(ctx, "minikube", "start", "-p", m.name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("starting minikube cluster: %w", err)
		}
		return nil
	}

	args := []string{
		"start",
		"-p", m.name,
		"--driver=docker",
	}

	if m.k8sVersion != "" {
		args = append(args, "--kubernetes-version", m.k8sVersion)
	}

	cmd := exec.CommandContext(ctx, "minikube", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating minikube cluster: %w", err)
	}

	return nil
}

// Down deletes the minikube cluster
func (m *MinikubeProvider) Down(ctx context.Context) error {
	exists, err := m.Exists(ctx)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if !exists {
		return nil
	}

	cmd := exec.CommandContext(ctx, "minikube", "delete", "-p", m.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deleting minikube cluster: %w", err)
	}

	return nil
}

// Exists checks if the minikube cluster exists
func (m *MinikubeProvider) Exists(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "minikube", "profile", "list", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If no profiles exist, minikube returns an error
		if strings.Contains(string(output), "no minikube profile") {
			return false, nil
		}
		return false, fmt.Errorf("listing minikube profiles: %w", err)
	}

	// Check if our profile name is in the output
	if strings.Contains(string(output), fmt.Sprintf(`"Name":"%s"`, m.name)) ||
		strings.Contains(string(output), fmt.Sprintf(`"name":"%s"`, m.name)) {
		return true, nil
	}

	return false, nil
}

// KubeconfigPath returns the path to the kubeconfig for this cluster
func (m *MinikubeProvider) KubeconfigPath(ctx context.Context) (string, error) {
	// Minikube stores kubeconfig in a standard location
	// We can get it via minikube kubectl command or directly
	cmd := exec.CommandContext(ctx, "minikube", "kubectl", "--", "config", "view", "--raw", "-p", m.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try alternative approach
		homeDir, _ := os.UserHomeDir()
		kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); err == nil {
			return kubeconfigPath, nil
		}
		return "", fmt.Errorf("getting kubeconfig: %w", err)
	}

	// Write kubeconfig to a temp file
	tmpDir := os.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, fmt.Sprintf("minikube-%s-kubeconfig", m.name))

	if err := os.WriteFile(kubeconfigPath, output, 0600); err != nil {
		return "", fmt.Errorf("writing kubeconfig: %w", err)
	}

	return kubeconfigPath, nil
}
