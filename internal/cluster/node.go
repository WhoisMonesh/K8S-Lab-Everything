package cluster

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// NodeShell runs a command inside a node container via `docker exec`. This is
// how you interact with a cluster node that does not expose SSH (e.g. kind).
// The node is addressed by its Docker container name, usually derived from the
// cluster name, e.g. "<cluster>-control-plane".
func NodeShell(ctx context.Context, nodeName string, command string) (string, error) {
	args := []string{"exec", nodeName, "sh", "-c", command}
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("node %s: %w", nodeName, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ControlPlaneNodeName returns the Docker container name of the control-plane
// node for a kind cluster.
func ControlPlaneNodeName(clusterName string) string {
	return clusterName + "-control-plane"
}

// WorkerNodeNames returns the Docker container names for the worker nodes of a
// kind cluster. kind names workers <cluster>-worker, <cluster>-worker2, ...
func WorkerNodeNames(clusterName string, workers int) []string {
	names := make([]string, 0, workers)
	for i := 0; i < workers; i++ {
		if i == 0 {
			names = append(names, clusterName+"-worker")
		} else {
			names = append(names, fmt.Sprintf("%s-worker%d", clusterName, i+1))
		}
	}
	return names
}

// NodeNamesForContainer returns the container name for an unregistered
// "pending" node (used by join labs), derived from a base node label.
func PendingNodeName(clusterName string) string {
	return clusterName + "-pending-worker"
}

// EnsureNodeContainer ensures a node container with the given name exists,
// provisioning it from a base image if it does not. Returns true if created.
// This is used by join labs to provide an extra "pending" node to practice
// kubeadm join against.
func EnsureNodeContainer(ctx context.Context, name, image string) (bool, error) {
	exists, err := containerExists(ctx, name)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	if image == "" {
		image = "kindest/node:v1.28.0"
	}

	// Provision a fresh, pre-booted container so it can be joined.
	args := []string{
		"run", "-d", "--name", name,
		"--privileged",
		"--hostname", name,
		"--tmpfs", "/tmp",
		"--volume", "/var",
		"--volume", "/lib/modules:/lib/modules:ro",
		image,
		"/sbin/init",
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("creating pending node container: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return true, nil
}

func containerExists(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// container not found
		return false, nil
	}
	if strings.TrimSpace(string(out)) == "true" {
		return true, nil
	}
	return false, nil
}
