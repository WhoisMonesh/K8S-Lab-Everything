package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&ClusterUpgradeLab{})
}

type ClusterUpgradeLab struct {
	BaseLab
}

func (l *ClusterUpgradeLab) ID() string             { return "cka_cluster_upgrade_v128_to_v129" }
func (l *ClusterUpgradeLab) Title() string          { return "Upgrade Cluster from 1.28 to 1.29" }
func (l *ClusterUpgradeLab) Category() Category     { return CategoryClusterArchitecture }
func (l *ClusterUpgradeLab) Difficulty() Difficulty { return DifficultyHard }
func (l *ClusterUpgradeLab) EstimatedTime() int     { return 45 }
func (l *ClusterUpgradeLab) Tags() []string {
	return []string{"cluster-upgrade", "kubeadm", "control-plane"}
}
func (l *ClusterUpgradeLab) Cert() Cert        { return CertCKA }
func (l *ClusterUpgradeLab) DomainWeight() int { return 25 }

// ClusterSpec declares the cluster this lab needs: a multi-node v1.28 cluster
// so the 1.28 -> 1.29 upgrade scenario is real. The runner auto-provisions it.
func (l *ClusterUpgradeLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           1,
	}
}

func (l *ClusterUpgradeLab) Description() string {
	return `The cluster is running Kubernetes 1.28 and needs to be upgraded to 1.29.
Upgrade the control plane first, then upgrade worker nodes. Ensure all components
are running the new version before finishing.

NOTE: This lab runs on a kind cluster. kind nodes are Docker containers that do
not expose SSH, so you access the node shell with:
    docker exec -it <cluster>-control-plane bash
    docker exec -it <cluster>-worker bash
Run the kubeadm/apt/systemctl commands inside that node shell. A real kubeadm
in-place upgrade is not possible in kind (the node image pins the version), so
this lab focuses on the exact command sequence you would run on a real node.`
}

func (l *ClusterUpgradeLab) Hints() []string {
	return []string{
		"Enter the control-plane node: docker exec -it <cluster>-control-plane bash",
		"Run kubeadm upgrade plan to see available upgrades",
		"Upgrade control plane with kubeadm upgrade apply",
		"Drain, upgrade, and uncordon each worker node",
	}
}

func (l *ClusterUpgradeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ClusterUpgradeLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ClusterUpgradeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
		"jsonpath={.items[*].status.nodeInfo.kubeletVersion}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "v1.28") {
		return fmt.Errorf("cluster still running v1.28")
	}
	return nil
}

func (l *ClusterUpgradeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current version", Command: "kubectl get nodes"},
		{Description: "Enter the control-plane node shell (kind nodes are containers)", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Plan the upgrade", Command: "kubeadm upgrade plan"},
		{Description: "Upgrade control plane", Command: "kubeadm upgrade apply v1.29.0"},
		{Description: "Exit and drain node", Command: "exit && kubectl drain <worker-node> --ignore-daemonsets --delete-emptydir-data"},
		{Description: "Enter worker node shell and upgrade kubelet", Command: "docker exec -it <cluster>-worker bash"},
		{Description: "Upgrade kubelet packages", Command: "apt-mark unhold kubelet && apt-get update && apt-get install -y kubelet=1.29.0-1.1 && apt-mark hold kubelet"},
		{Description: "Restart kubelet", Command: "systemctl daemon-reload && systemctl restart kubelet"},
		{Description: "Exit and uncordon node", Command: "exit && kubectl uncordon <worker-node>"},
	}
}
