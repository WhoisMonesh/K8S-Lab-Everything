package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodSchedulingFailedLab{})
}

type PodSchedulingFailedLab struct {
	BaseLab
}

func (l *PodSchedulingFailedLab) ID() string {
	return "pod_scheduling_failed"
}

func (l *PodSchedulingFailedLab) Title() string {
	return "Pod Stuck in Pending - Scheduling Failed"
}

func (l *PodSchedulingFailedLab) Category() Category {
	return CategoryScheduling
}

func (l *PodSchedulingFailedLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *PodSchedulingFailedLab) Description() string {
	return `A pod named 'special-app' in the 'apps' namespace is stuck in Pending state.
The pod has a nodeSelector that doesn't match any node labels.

Your task: Fix the pod scheduling by updating the node labels or pod configuration.`
}

func (l *PodSchedulingFailedLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the pod's nodeSelector field",
		"Check the node labels",
		"Either add the label to a node or remove the nodeSelector",
	}
}

func (l *PodSchedulingFailedLab) EstimatedTime() int {
	return 10
}

func (l *PodSchedulingFailedLab) Tags() []string {
	return []string{"scheduling", "nodeselector", "pending", "nodes"}
}

func (l *PodSchedulingFailedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodSchedulingFailedLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create namespace
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: apps
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	// Create pod with nodeSelector that won't match any node
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: special-app
  namespace: apps
spec:
  nodeSelector:
    gpu: "true"
  containers:
  - name: app
    image: nginx:alpine
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodSchedulingFailedLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodSchedulingFailedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "special-app", "-n", "apps",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodSchedulingFailedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod special-app -n apps",
			Notes:       "The pod is stuck in Pending state",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod special-app -n apps | grep -A 10 Events",
			Notes:       "Look for '0/N nodes are available: node(s) had taint(s), pod didn't tolerate' or 'no nodes match node selector'",
		},
		{
			Description: "Check node labels",
			Command:     "kubectl get nodes --show-labels",
			Notes:       "No node has the 'gpu=true' label",
		},
		{
			Description: "Option A: Add label to a node",
			Command:     "kubectl label node <node-name> gpu=true",
			Notes:       "Replace <node-name> with an actual node name from your cluster",
		},
		{
			Description: "Option B: Remove the nodeSelector from the pod",
			Command:     "kubectl edit pod special-app -n apps",
			Notes:       "Remove or modify the nodeSelector to match available nodes",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod special-app -n apps",
			Notes:       "The pod should now be in Running state",
		},
	}
}
