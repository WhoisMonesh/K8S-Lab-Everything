package labs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodeNodeAffinityLab{})
}

type MultiNodeNodeAffinityLab struct {
	BaseLab
}

func (l *MultiNodeNodeAffinityLab) ID() string {
	return "multi_node_node_affinity"
}

func (l *MultiNodeNodeAffinityLab) Title() string {
	return "Node Affinity Rules Across Workers"
}

func (l *MultiNodeNodeAffinityLab) Category() Category {
	return CategoryWorkloadsScheduling
}

func (l *MultiNodeNodeAffinityLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *MultiNodeNodeAffinityLab) Description() string {
	return `A deployment needs to schedule pods on nodes with specific hardware
characteristics. The requiredDuringSchedulingIgnoredDuringExecution affinity
rule is too restrictive and no nodes match.

Your task: Fix the node affinity rules so pods can be scheduled on
available worker nodes.`
}

func (l *MultiNodeNodeAffinityLab) Hints() []string {
	return []string{
		"Check the current node labels",
		"requiredDuringSchedulingIgnoredDuringExecution is very strict",
		"Consider using preferredDuringSchedulingIgnoredDuringExecution instead",
	}
}

func (l *MultiNodeNodeAffinityLab) EstimatedTime() int {
	return 20
}

func (l *MultiNodeNodeAffinityLab) Tags() []string {
	return []string{"node-affinity", "scheduling", "multi-node", "constraints"}
}

func (l *MultiNodeNodeAffinityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

// ClusterSpec declares a multi-worker cluster so node scheduling/scaling
// scenarios are real on kind.
func (l *MultiNodeNodeAffinityLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           2,
	}
}

func (l *MultiNodeNodeAffinityLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: affinity-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: affinity-app
  template:
    metadata:
      labels:
        app: affinity-app
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: gpu-type
                operator: In
                values:
                - nvidia-a100
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo affinity; sleep 15; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *MultiNodeNodeAffinityLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=affinity-app",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	if strings.Contains(output, "Running") {
		return fmt.Errorf("pods are running (expected pending)")
	}

	return nil
}

func (l *MultiNodeNodeAffinityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "affinity-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	ready, _ := strconv.Atoi(strings.TrimSpace(output))
	if ready < 3 {
		return fmt.Errorf("deployment not ready (ready: %d, expected: 3)", ready)
	}

	return nil
}

func (l *MultiNodeNodeAffinityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=affinity-app",
			Notes:       "Pods should be in Pending state",
		},
		{
			Description: "Check node labels",
			Command:     "kubectl get nodes --show-labels",
			Notes:       "No nodes have gpu-type=nvidia-a100",
		},
		{
			Description: "Fix: Change to preferredDuringScheduling",
			Command:     `kubectl patch deploy affinity-app --type='json' -p='[{"op":"replace","path":"/spec/template/spec/affinity/nodeAffinity","value":{"preferredDuringSchedulingIgnoredDuringExecution":[{"weight":1,"preference":{"matchExpressions":[{"key":"gpu-type","operator":"In","values":["nvidia-a100"]}]}}]}}]'`,
			Notes:       "Use preferred instead of required scheduling",
		},
		{
			Description: "Verify pods are running",
			Command:     "kubectl rollout status deploy/affinity-app",
			Notes:       "All 3 replicas should be ready",
		},
	}
}
