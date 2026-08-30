package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&KubeletNotRegisteredLab{})
}

type KubeletNotRegisteredLab struct {
	BaseLab
}

func (l *KubeletNotRegisteredLab) ID() string { return "cka_kubelet_not_registered" }
func (l *KubeletNotRegisteredLab) Title() string {
	return "Debug Kubelet Registration Issues"
}
func (l *KubeletNotRegisteredLab) Category() Category     { return CategoryTroubleshooting }
func (l *KubeletNotRegisteredLab) Difficulty() Difficulty { return DifficultyHard }
func (l *KubeletNotRegisteredLab) EstimatedTime() int     { return 25 }
func (l *KubeletNotRegisteredLab) Tags() []string {
	return []string{"kubelet", "registration", "node", "troubleshooting"}
}
func (l *KubeletNotRegisteredLab) Cert() Cert        { return CertCKA }
func (l *KubeletNotRegisteredLab) DomainWeight() int { return 30 }

func (l *KubeletNotRegisteredLab) Description() string {
	return `A kubelet on a worker node has been stopped and is no longer registered
with the API server. The node is missing or NotReady in kubectl get nodes.
Diagnose and restart the kubelet to re-register the node.

kind nodes are containers (no SSH); access the worker node shell with:
    docker exec -it <cluster>-worker bash`
}

func (l *KubeletNotRegisteredLab) Hints() []string {
	return []string{
		"Enter the worker node: docker exec -it <cluster>-worker bash",
		"Check kubelet logs: journalctl -u kubelet",
		"Verify kubelet is running: systemctl status kubelet",
		"Restart kubelet: systemctl restart kubelet",
	}
}

func (l *KubeletNotRegisteredLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           1,
	}
}

func (l *KubeletNotRegisteredLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeletNotRegisteredLab) Break(ctx context.Context, kubeconfigPath string) error {
	node, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	if _, err := dockerExec(ctx, node, "sh", "-c", "pkill -x kubelet || systemctl stop kubelet"); err != nil {
		return fmt.Errorf("failed to stop kubelet on node %s: %w", node, err)
	}
	time.Sleep(5 * time.Second)
	return nil
}

func (l *KubeletNotRegisteredLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *KubeletNotRegisteredLab) Verify(ctx context.Context, kubeconfigPath string) error {
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
	return fmt.Errorf("worker node not ready/registered")
}

func (l *KubeletNotRegisteredLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status", Command: "kubectl get nodes"},
		{Description: "Describe node to confirm it is NotReady", Command: "kubectl describe node <node-name> | grep -A5 Conditions"},
		{Description: "Enter the node shell (kind has no SSH)", Command: "docker exec -it <cluster>-worker bash"},
		{Description: "Check kubelet status", Command: "systemctl status kubelet || ps aux | grep kubelet"},
		{Description: "Restart the stopped kubelet", Command: "systemctl restart kubelet || (/usr/local/bin/kubelet & )"},
		{Description: "Exit the node and verify it is Ready", Command: "exit && kubectl wait --for=condition=Ready node --all --timeout=120s"},
	}
}
