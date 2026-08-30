package labs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodeDaemonsetSelectorLab{})
}

type MultiNodeDaemonsetSelectorLab struct {
	BaseLab
}

func (l *MultiNodeDaemonsetSelectorLab) ID() string {
	return "multi_node_daemonset_selector"
}

func (l *MultiNodeDaemonsetSelectorLab) Title() string {
	return "DaemonSet Node Selector Mismatch"
}

func (l *MultiNodeDaemonsetSelectorLab) Category() Category {
	return CategoryWorkloadsScheduling
}

func (l *MultiNodeDaemonsetSelectorLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *MultiNodeDaemonsetSelectorLab) Description() string {
	return `A DaemonSet should run a logging agent on all worker nodes, but
it's only running on some nodes because of a mismatched node selector.

Your task: Fix the DaemonSet so it runs on all worker nodes in the cluster.`
}

func (l *MultiNodeDaemonsetSelectorLab) Hints() []string {
	return []string{
		"Check the DaemonSet's nodeSelector",
		"Compare node labels across all nodes",
		"Either fix the selector or label the nodes",
	}
}

func (l *MultiNodeDaemonsetSelectorLab) EstimatedTime() int {
	return 15
}

func (l *MultiNodeDaemonsetSelectorLab) Tags() []string {
	return []string{"daemonset", "node-selector", "multi-node", "logging"}
}

func (l *MultiNodeDaemonsetSelectorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: logging
`
	return kubectlApply(ctx, kubeconfigPath, namespace)
}

// ClusterSpec declares a multi-worker cluster so node scheduling/scaling
// scenarios are real on kind.
func (l *MultiNodeDaemonsetSelectorLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           2,
	}
}

func (l *MultiNodeDaemonsetSelectorLab) Break(ctx context.Context, kubeconfigPath string) error {
	daemonset := `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: log-agent
  namespace: logging
spec:
  selector:
    matchLabels:
      app: log-agent
  template:
    metadata:
      labels:
        app: log-agent
    spec:
      nodeSelector:
        logging-enabled: "true"
      containers:
      - name: agent
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo logging; sleep 30; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, daemonset)
}

func (l *MultiNodeDaemonsetSelectorLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "daemonset", "log-agent",
		"-n", "logging", "-o", "jsonpath={.status.desiredNumberScheduled}")
	if err != nil {
		return fmt.Errorf("checking daemonset: %w", err)
	}

	desired, _ := strconv.Atoi(strings.TrimSpace(output))
	if desired <= 1 {
		return nil
	}

	ready, _ := kubectl(ctx, kubeconfigPath, "get", "daemonset", "log-agent",
		"-n", "logging", "-o", "jsonpath={.status.numberReady}")
	readyNum, _ := strconv.Atoi(strings.TrimSpace(ready))

	if readyNum < desired {
		return nil
	}

	return fmt.Errorf("daemonset is running on all nodes (expected broken)")
}

func (l *MultiNodeDaemonsetSelectorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	nodes, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o", "name")
	if err != nil {
		return fmt.Errorf("getting nodes: %w", err)
	}

	nodeCount := len(strings.Split(strings.TrimSpace(nodes), "\n"))

	output, err := kubectl(ctx, kubeconfigPath, "get", "daemonset", "log-agent",
		"-n", "logging", "-o", "jsonpath={.status.numberReady}")
	if err != nil {
		return fmt.Errorf("checking daemonset: %w", err)
	}

	ready, _ := strconv.Atoi(strings.TrimSpace(output))
	if ready < nodeCount {
		return fmt.Errorf("daemonset not on all nodes (ready: %d, nodes: %d)", ready, nodeCount)
	}

	return nil
}

func (l *MultiNodeDaemonsetSelectorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check DaemonSet status",
			Command:     "kubectl get daemonset -n logging",
			Notes:       "See how many pods are scheduled vs desired",
		},
		{
			Description: "Check node labels",
			Command:     "kubectl get nodes --show-labels",
			Notes:       "Not all nodes have logging-enabled=true",
		},
		{
			Description: "Fix: Remove the nodeSelector",
			Command:     `kubectl patch daemonset log-agent -n logging --type json -p='[{"op": "remove", "path": "/spec/template/spec/nodeSelector"}]'`,
			Notes:       "Or label all nodes: kubectl label nodes --all logging-enabled=true",
		},
		{
			Description: "Verify DaemonSet runs on all nodes",
			Command:     "kubectl get daemonset -n logging",
			Notes:       "Should show desired == ready == number of nodes",
		},
	}
}
