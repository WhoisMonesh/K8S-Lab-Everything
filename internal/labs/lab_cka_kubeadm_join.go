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

// ClusterSpec declares a single control-plane node cluster for the join lab.
func (l *KubeadmJoinLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           0,
	}
}

// RequiredNodes declares a pending, not-yet-joined worker node container that
// the runner provisions so the join scenario is real.
func (l *KubeadmJoinLab) RequiredNodes() []PendingNode {
	return []PendingNode{{Name: "pending-worker", Version: "v1.28.0"}}
}

func (l *KubeadmJoinLab) Description() string {
	return `A new worker node needs to be joined to the cluster. The join token has expired
and needs to be regenerated.

A pending worker node container has been provisioned for you (named
<cluster>-pending-worker). You join it like any real node — but since it has no
SSH, run the kubeadm join command inside it with:
    docker exec -it <cluster>-pending-worker bash
Then run the join command there. Access the control plane node the same way:
    docker exec -it <cluster>-control-plane bash`
}

func (l *KubeadmJoinLab) Hints() []string {
	return []string{
		"Generate a new token: kubeadm token create --print-join-command (inside the control-plane node shell)",
		"Copy the join command into the pending worker node shell and run it",
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
	// The pending worker joins, so we expect more than the control-plane node.
	if len(lines) < 2 {
		return fmt.Errorf("new node has not joined the cluster yet")
	}
	return nil
}

func (l *KubeadmJoinLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the control-plane node shell", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Generate new join token", Command: "kubeadm token create --print-join-command"},
		{Description: "Exit the control-plane shell (keep the join command for the next step)", Command: "exit"},
		{Description: "Enter the pending worker node shell", Command: "docker exec -it <cluster>-pending-worker bash"},
		{Description: "Run join command on the pending worker (paste the command from step above)", Command: "kubeadm join <api-server>:6443 --token <token> --discovery-token-ca-cert-hash sha256:<hash>"},
		{Description: "Exit and verify node joined", Command: "exit && kubectl get nodes"},
	}
}
