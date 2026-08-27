package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ResourceRequestTooHigh{})
}

type ResourceRequestTooHigh struct {
	BaseLab
}

func (l *ResourceRequestTooHigh) ID() string             { return "resource_request_too_high" }
func (l *ResourceRequestTooHigh) Title() string          { return "Pod Pending - Resource Request Too High" }
func (l *ResourceRequestTooHigh) Category() Category     { return CategoryScheduling }
func (l *ResourceRequestTooHigh) Difficulty() Difficulty { return DifficultyEasy }
func (l *ResourceRequestTooHigh) EstimatedTime() int     { return 10 }
func (l *ResourceRequestTooHigh) Tags() []string {
	return []string{"scheduling", "resources", "requests"}
}

func (l *ResourceRequestTooHigh) Description() string {
	return `A pod is stuck in Pending because the resource requests exceed available node resources.
Reduce the resource requests to fit within available capacity.`
}

func (l *ResourceRequestTooHigh) Hints() []string {
	return []string{
		"Check pod resource requests",
		"Check node allocatable resources",
		"Reduce memory/CPU requests",
	}
}

func (l *ResourceRequestTooHigh) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceRequestTooHigh) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: big-pod
spec:
  containers:
  - name: nginx
    image: nginx:alpine
    resources:
      requests:
        memory: "100Gi"
        cpu: "50"`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ResourceRequestTooHigh) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "big-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not running: %s", output)
	}
	return nil
}

func (l *ResourceRequestTooHigh) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod requests", Command: "kubectl get pod big-pod -o jsonpath='{.spec.containers[0].resources}'"},
		{Description: "Check node resources", Command: "kubectl describe nodes | grep -A 5 Allocatable"},
		{Description: "Fix resource requests", Command: "kubectl edit pod big-pod and reduce memory to 128Mi and cpu to 0.1"},
	}
}
