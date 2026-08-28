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
	workers    int
}

// NewKindProvider creates a new kind cluster provider
func NewKindProvider(name, k8sVersion string, workers int) *KindProvider {
	return &KindProvider{
		name:       name,
		k8sVersion: k8sVersion,
		workers:    workers,
	}
}

// Name returns the cluster name
func (k *KindProvider) Name() string {
	return k.name
}

// kindConfigTemplate is the template for single-node kind cluster configuration
const kindConfigTemplate = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
`

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

	// Generate kind config for multi-node clusters (includes image)
	if k.workers > 0 {
		configPath, err := k.generateConfig()
		if err != nil {
			return fmt.Errorf("generating kind config: %w", err)
		}
		defer os.Remove(configPath)
		args = append(args, "--config", configPath)
	} else if k.k8sVersion != "" {
		// Single-node: use --image flag
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

// generateConfig creates a temporary kind config file for multi-node clusters
func (k *KindProvider) generateConfig() (string, error) {
	var config strings.Builder
	config.WriteString("kind: Cluster\n")
	config.WriteString("apiVersion: kind.x-k8s.io/v1alpha4\n")

	if k.k8sVersion != "" {
		nodeImage := k.versionToNodeImage(k.k8sVersion)
		config.WriteString(fmt.Sprintf("nodes:\n"))
		config.WriteString(fmt.Sprintf("- role: control-plane\n  image: %s\n", nodeImage))
		for i := 0; i < k.workers; i++ {
			config.WriteString(fmt.Sprintf("- role: worker\n  image: %s\n", nodeImage))
		}
	} else {
		config.WriteString("nodes:\n")
		config.WriteString("- role: control-plane\n")
		for i := 0; i < k.workers; i++ {
			config.WriteString("- role: worker\n")
		}
	}

	tmpFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return "", err
	}

	if _, err := tmpFile.WriteString(config.String()); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
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

	tmpDir := os.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, fmt.Sprintf("kind-%s-kubeconfig", k.name))

	if err := os.WriteFile(kubeconfigPath, output, 0600); err != nil {
		return "", fmt.Errorf("writing kubeconfig: %w", err)
	}

	return kubeconfigPath, nil
}

// versionToNodeImage converts a Kubernetes version to a kind node image
func (k *KindProvider) versionToNodeImage(version string) string {
	version = strings.TrimPrefix(version, "v")
	return fmt.Sprintf("kindest/node:v%s", version)
}
