package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NodeNotReadyLab{})
}

type NodeNotReadyLab struct {
	BaseLab
}

func (l *NodeNotReadyLab) ID() string             { return "cka_node_not_ready" }
func (l *NodeNotReadyLab) Title() string          { return "Troubleshoot Node NotReady Status" }
func (l *NodeNotReadyLab) Category() Category     { return CategoryTroubleshooting }
func (l *NodeNotReadyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NodeNotReadyLab) EstimatedTime() int     { return 25 }
func (l *NodeNotReadyLab) Tags() []string {
	return []string{"node", "not-ready", "kubelet", "troubleshooting"}
}
func (l *NodeNotReadyLab) Cert() Cert        { return CertCKA }
func (l *NodeNotReadyLab) DomainWeight() int { return 30 }

func (l *NodeNotReadyLab) Description() string {
	return `A worker node is in NotReady status. Diagnose the issue by checking
the node conditions and fix the underlying problem. The kubelet may have
configuration issues or may have stopped running.

kind nodes are containers (no SSH); access the worker node shell with:
    docker exec -it <cluster>-worker bash`
}

func (l *NodeNotReadyLab) Hints() []string {
	return []string{
		"Check node conditions with kubectl describe node",
		"Enter the worker node: docker exec -it <cluster>-worker bash",
		"Look at kubelet logs: journalctl -u kubelet",
		"Verify kubelet is running: systemctl status kubelet",
	}
}

func (l *NodeNotReadyLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           1,
	}
}

func (l *NodeNotReadyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeNotReadyLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Real scenario: stop kubelet inside the worker node's container. The node
	// reports NotReady until the learner restarts it.
	node, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	// The node container is addressed by its kind name (same as the node name
	// in a kind cluster).
	if _, err := dockerExec(ctx, node, "sh", "-c", "pkill -x kubelet || systemctl stop kubelet"); err != nil {
		return fmt.Errorf("failed to stop kubelet on node %s: %w", node, err)
	}
	// Give the node time to report NotReady.
	time.Sleep(5 * time.Second)
	return nil
}

func (l *NodeNotReadyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Wait for all nodes to be Ready (learner restarted kubelet).
	for i := 0; i < 30; i++ {
		output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
			"jsonpath={.items[*].status.conditions[?(@.type=='Ready')].status}")
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		ready := strings.Fields(output)
		if len(ready) > 0 {
			allReady := true
			for _, s := range ready {
				if strings.TrimSpace(s) != "True" {
					allReady = false
					break
				}
			}
			if allReady {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("node is still in NotReady status")
}

func (l *NodeNotReadyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status", Command: "kubectl get nodes"},
		{Description: "Describe node to confirm it is NotReady", Command: "kubectl describe node <node-name> | grep -A5 Conditions"},
		{Description: "Enter the node shell (kind has no SSH)", Command: "docker exec -it <cluster>-worker bash"},
		{Description: "Check kubelet status", Command: "systemctl status kubelet || ps aux | grep kubelet"},
		{Description: "Restart the stopped kubelet", Command: "systemctl restart kubelet || (/usr/local/bin/kubelet & )"},
		{Description: "Exit the node and verify it is Ready", Command: "exit && kubectl wait --for=condition=Ready node --all --timeout=120s"},
	}
}
