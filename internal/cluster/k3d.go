package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// K3dProvider implements the Provider interface for k3d clusters
type K3dProvider struct {
	name       string
	k8sVersion string
}

// NewK3dProvider creates a new k3d cluster provider
func NewK3dProvider(name, k8sVersion string) *K3dProvider {
	return &K3dProvider{
		name:       name,
		k8sVersion: k8sVersion,
	}
}

// Name returns the cluster name
func (k *K3dProvider) Name() string {
	return k.name
}

// Up creates the k3d cluster
func (k *K3dProvider) Up(ctx context.Context) error {
	exists, err := k.Exists(ctx)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if exists {
		return nil
	}

	args := []string{
		"cluster", "create", k.name,
		"--servers", "1",
		"--agents", "0",
	}

	if k.k8sVersion != "" {
		// k3d uses rancher/k3s images
		imageVersion := k.versionToImage(k.k8sVersion)
		args = append(args, "--image", imageVersion)
	}

	cmd := exec.CommandContext(ctx, "k3d", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating k3d cluster: %w", err)
	}

	return nil
}

// Down deletes the k3d cluster
func (k *K3dProvider) Down(ctx context.Context) error {
	exists, err := k.Exists(ctx)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if !exists {
		return nil
	}

	cmd := exec.CommandContext(ctx, "k3d", "cluster", "delete", k.name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("deleting k3d cluster: %w", err)
	}

	return nil
}

// Exists checks if the k3d cluster exists
func (k *K3dProvider) Exists(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "k3d", "cluster", "list", "--no-headers")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("listing k3d clusters: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == k.name {
			return true, nil
		}
	}

	return false, nil
}

// KubeconfigPath returns the path to the kubeconfig for this cluster
func (k *K3dProvider) KubeconfigPath(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "k3d", "kubeconfig", "get", k.name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("getting kubeconfig: %w", err)
	}

	// Write kubeconfig to a temp file
	tmpDir := os.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, fmt.Sprintf("k3d-%s-kubeconfig", k.name))

	if err := os.WriteFile(kubeconfigPath, output, 0600); err != nil {
		return "", fmt.Errorf("writing kubeconfig: %w", err)
	}

	return kubeconfigPath, nil
}

// versionToImage converts a Kubernetes version to a k3d image
func (k *K3dProvider) versionToImage(version string) string {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")
	// k3d uses rancher/k3s images
	return fmt.Sprintf("rancher/k3s:v%s-k3s1", version)
}
