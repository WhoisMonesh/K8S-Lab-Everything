package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&ContainerRuntimeMigrationLab{})
}

type ContainerRuntimeMigrationLab struct {
	BaseLab
}

func (l *ContainerRuntimeMigrationLab) ID() string {
	return "cka_container_runtime_migration"
}
func (l *ContainerRuntimeMigrationLab) Title() string {
	return "Migrate from dockershim to containerd"
}
func (l *ContainerRuntimeMigrationLab) Category() Category     { return CategoryClusterArchitecture }
func (l *ContainerRuntimeMigrationLab) Difficulty() Difficulty { return DifficultyHard }
func (l *ContainerRuntimeMigrationLab) EstimatedTime() int     { return 35 }
func (l *ContainerRuntimeMigrationLab) Tags() []string {
	return []string{"containerd", "dockershim", "runtime", "migration"}
}
func (l *ContainerRuntimeMigrationLab) Cert() Cert        { return CertCKA }
func (l *ContainerRuntimeMigrationLab) DomainWeight() int { return 25 }

func (l *ContainerRuntimeMigrationLab) Description() string {
	return `The cluster is using dockershim which is deprecated. Migrate a worker
node from Docker to containerd runtime. Stop Docker, configure containerd,
update kubelet, and verify pods are running.`
}

func (l *ContainerRuntimeMigrationLab) Hints() []string {
	return []string{
		"Install containerd on the worker node",
		"Update kubelet to use --container-runtime-endpoint",
		"Restart kubelet after changes",
	}
}

func (l *ContainerRuntimeMigrationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ContainerRuntimeMigrationLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ContainerRuntimeMigrationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
		"jsonpath={.items[*].status.nodeInfo.containerRuntimeVersion}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "docker") {
		return fmt.Errorf("node still using docker runtime")
	}
	return nil
}

func (l *ContainerRuntimeMigrationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current runtime", Command: "kubectl get nodes -o wide"},
		{Description: "Stop Docker", Command: "sudo systemctl stop docker"},
		{Description: "Install containerd", Command: "sudo apt-get install -y containerd"},
		{Description: "Configure containerd", Command: "sudo mkdir -p /etc/containerd && containerd config default | sudo tee /etc/containerd/config.toml"},
		{Description: "Update kubelet config", Command: "sudo sed -i 's|--container-runtime=remote|--container-runtime-endpoint=unix:///run/containerd/containerd.sock|' /var/lib/kubelet/config.yaml"},
		{Description: "Restart services", Command: "sudo systemctl restart containerd && sudo systemctl restart kubelet"},
	}
}
