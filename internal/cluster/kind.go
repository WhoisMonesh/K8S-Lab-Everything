package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// KindProvider implements the Provider interface for kind clusters
type KindProvider struct {
	name       string
	k8sVersion string
}

// NewKindProvider creates a new kind cluster provider
func NewKindProvider(name, k8sVersion string) *KindProvider {
	return &KindProvider{
		name:       name,
		k8sVersion: k8sVersion,
	}
}

// Name returns the cluster name
func (k *KindProvider) Name() string {
	return k.name
}

// Up creates the kind cluster
func (k *KindProvider) Up(ctx context.Context) error {
	exists, err := k.Exists(ctx)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if exists {
		return nil
	}

	args := []string{
		"create", "cluster",
		"--name", k.name,
	}

	if k.k8sVersion != "" {
		// Convert version to kind node image
		nodeImage := k.versionToNodeImage(k.k8sVersion)
		args = append(args, "--image", nodeImage)
	}

	cmd := exec.CommandContext(ctx, "kind", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating kind cluster: %w", err)
	}

	return nil
}

// Down deletes the kind cluster
func (k *KindProvider) Down(ctx context.Context) error {
	exists, err := k.Exists(ctx)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if !exists {
		return nil
	}

	cmd := exec.CommandContext(ctx, "kind", "delete", "cluster", "--name", k.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deleting kind cluster: %w", err)
	}

	return nil
}

// Exists checks if the kind cluster exists
func (k *KindProvider) Exists(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// kind get clusters can fail if Docker is not running.
		// Check if Docker is the problem and give a clear message.
		dockerCmd := exec.CommandContext(ctx, "docker", "info")
		if dErr := dockerCmd.Run(); dErr != nil {
			return false, fmt.Errorf("Docker is not running. Please start Docker Desktop and try again")
		}
		return false, fmt.Errorf("listing kind clusters: %s", strings.TrimSpace(string(output)))
	}

	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if strings.TrimSpace(cluster) == k.name {
			return true, nil
		}
	}

	return false, nil
}

// KubeconfigPath returns the path to the kubeconfig for this cluster
func (k *KindProvider) KubeconfigPath(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "kind", "get", "kubeconfig", "--name", k.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getting kubeconfig: %w", err)
	}

	// Write kubeconfig to a temp file
	tmpDir := os.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, fmt.Sprintf("kind-%s-kubeconfig", k.name))

	if err := os.WriteFile(kubeconfigPath, output, 0600); err != nil {
		return "", fmt.Errorf("writing kubeconfig: %w", err)
	}

	return kubeconfigPath, nil
}

// versionToNodeImage converts a Kubernetes version to a kind node image
func (k *KindProvider) versionToNodeImage(version string) string {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")
	return fmt.Sprintf("kindest/node:v%s", version)
}
