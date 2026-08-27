package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&KubeadmJoinLab{})
}

type KubeadmJoinLab struct {
	BaseLab
}

func (l *KubeadmJoinLab) ID() string             { return "cka_kubeadm_join_worker" }
func (l *KubeadmJoinLab) Title() string          { return "Join New Worker Node Using kubeadm" }
func (l *KubeadmJoinLab) Category() Category     { return CategoryClusterArchitecture }
func (l *KubeadmJoinLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *KubeadmJoinLab) EstimatedTime() int     { return 30 }
func (l *KubeadmJoinLab) Tags() []string {
	return []string{"kubeadm", "join", "worker-node", "cluster"}
}
func (l *KubeadmJoinLab) Cert() Cert        { return CertCKA }
func (l *KubeadmJoinLab) DomainWeight() int { return 25 }

func (l *KubeadmJoinLab) Description() string {
	return `A new worker node needs to be joined to the cluster. The join token has expired
and needs to be regenerated. Generate a new join token and prepare the worker node
to join the cluster.`
}

func (l *KubeadmJoinLab) Hints() []string {
	return []string{
		"Use kubeadm token create to generate a new token",
		"Run kubeadm token create --print-join-command to get the full command",
		"Ensure the new node can reach the API server",
	}
}

func (l *KubeadmJoinLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeadmJoinLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *KubeadmJoinLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o", "name")
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("cluster still has only one node")
	}
	return nil
}

func (l *KubeadmJoinLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Generate new join token", Command: "sudo kubeadm token create --print-join-command"},
		{Description: "Run join command on worker", Command: "sudo kubeadm join <api-server>:6443 --token <token> --discovery-token-ca-cert-hash sha256:<hash>"},
		{Description: "Verify node joined", Command: "kubectl get nodes"},
	}
}
