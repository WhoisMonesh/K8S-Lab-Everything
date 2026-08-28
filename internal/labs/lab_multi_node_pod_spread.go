package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodePodSpreadLab{})
}

type MultiNodePodSpreadLab struct {
	BaseLab
}

func (l *MultiNodePodSpreadLab) ID() string {
	return "multi_node_pod_spread"
}

func (l *MultiNodePodSpreadLab) Title() string {
	return "Pod Topology Spread Across Nodes"
}

func (l *MultiNodePodSpreadLab) Category() Category {
	return CategoryWorkloadsScheduling
}

func (l *MultiNodePodSpreadLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *MultiNodePodSpreadLab) Description() string {
	return `A deployment should spread its pods evenly across all worker nodes
for high availability, but currently all pods are landing on a single node.

Your task: Configure topology spread constraints to ensure pods are
distributed evenly across all available worker nodes.`
}

func (l *MultiNodePodSpreadLab) Hints() []string {
	return []string{
		"Check how many worker nodes are available",
		"Use topologySpreadConstraints with maxSkew",
		"Set the topologyKey to kubernetes.io/hostname",
	}
}

func (l *MultiNodePodSpreadLab) EstimatedTime() int {
	return 20
}

func (l *MultiNodePodSpreadLab) Tags() []string {
	return []string{"topology-spread", "multi-node", "scheduling", "high-availability"}
}

func (l *MultiNodePodSpreadLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *MultiNodePodSpreadLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: spread-app
  namespace: default
spec:
  replicas: 6
  selector:
    matchLabels:
      app: spread-app
  template:
    metadata:
      labels:
        app: spread-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo spreading; sleep 15; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *MultiNodePodSpreadLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=spread-app",
		"-o", "jsonpath={range .items[*]}{.spec.nodeName}{'\n'}{end}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	nodes := strings.Split(strings.TrimSpace(output), "\n")
	nodeCount := make(map[string]int)
	for _, n := range nodes {
		if n != "" {
			nodeCount[n]++
		}
	}

	if len(nodeCount) <= 1 {
		return nil
	}

	return fmt.Errorf("pods are spread across nodes (broken state expected all on one node)")
}

func (l *MultiNodePodSpreadLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=spread-app",
		"-o", "jsonpath={range .items[*]}{.spec.nodeName}{'\n'}{end}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	nodes := strings.Split(strings.TrimSpace(output), "\n")
	nodeCount := make(map[string]int)
	for _, n := range nodes {
		if n != "" {
			nodeCount[n]++
		}
	}

	totalPods := len(nodes)
	if totalPods < 2 {
		return fmt.Errorf("not enough pods running (got %d)", totalPods)
	}

	if len(nodeCount) < 2 {
		return fmt.Errorf("pods are not spread across nodes (all on same node)")
	}

	for node, count := range nodeCount {
		if count > totalPods/len(nodeCount)+1 {
			return fmt.Errorf("node %s has too many pods (%d/%d)", node, count, totalPods)
		}
	}

	return nil
}

func (l *MultiNodePodSpreadLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check available nodes",
			Command:     "kubectl get nodes",
			Notes:       "Verify multiple worker nodes exist",
		},
		{
			Description: "Check current pod distribution",
			Command:     "kubectl get pods -l app=spread-app -o wide",
			Notes:       "Notice all pods on same node",
		},
		{
			Description: "Fix: Add topology spread constraints",
			Command:     `kubectl patch deploy spread-app --type='json' -p='[{"op":"add","path":"/spec/template/spec/topologySpreadConstraints","value":[{"maxSkew":1,"topologyKey":"kubernetes.io/hostname","whenUnsatisfiable":"DoNotSchedule","labelSelector":{"matchLabels":{"app":"spread-app"}}}]}]'`,
			Notes:       "Spread pods evenly across nodes",
		},
		{
			Description: "Verify pods are spread",
			Command:     "kubectl get pods -l app=spread-app -o wide",
			Notes:       "Pods should be distributed across nodes",
		},
	}
}
