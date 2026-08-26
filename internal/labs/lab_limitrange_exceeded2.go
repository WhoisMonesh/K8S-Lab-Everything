package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&LimitRangeExceeded{})
}

type LimitRangeExceeded struct {
	BaseLab
}

func (l *LimitRangeExceeded) ID() string             { return "limitrange_exceeded2" }
func (l *LimitRangeExceeded) Title() string          { return "Pod Exceeds LimitRange" }
func (l *LimitRangeExceeded) Category() Category     { return CategoryScheduling }
func (l *LimitRangeExceeded) Difficulty() Difficulty { return DifficultyMedium }
func (l *LimitRangeExceeded) EstimatedTime() int     { return 15 }
func (l *LimitRangeExceeded) Tags() []string {
	return []string{"scheduling", "limitrange", "resources"}
}

func (l *LimitRangeExceeded) Description() string {
	return `A pod is being rejected because it exceeds the LimitRange for the namespace.
Fix the pod resource limits to comply with the LimitRange.`
}

func (l *LimitRangeExceeded) Hints() []string {
	return []string{
		"Check the LimitRange in the namespace",
		"Compare pod resources with LimitRange limits",
		"Reduce pod resource limits to comply",
	}
}

func (l *LimitRangeExceeded) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *LimitRangeExceeded) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: limitrange-test
---
apiVersion: v1
kind: LimitRange
metadata:
  name: limits
  namespace: limitrange-test
spec:
  limits:
  - default:
      cpu: "500m"
      memory: "256Mi"
    defaultRequest:
      cpu: "100m"
      memory: "128Mi"
    max:
      cpu: "1"
      memory: "512Mi"
    type: Container
---
apiVersion: v1
kind: Pod
metadata:
  name: big-pod
  namespace: limitrange-test
spec:
  containers:
  - name: nginx
    image: nginx:alpine
    resources:
      limits:
        cpu: "2"
        memory: "1Gi"`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *LimitRangeExceeded) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "big-pod", "-n", "limitrange-test",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not running: %s", output)
	}
	return nil
}

func (l *LimitRangeExceeded) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check LimitRange", Command: "kubectl get limitrange -n limitrange-test -o yaml"},
		{Description: "Fix pod resources", Command: "kubectl edit pod big-pod -n limitrange-test"},
		{Description: "Reduce limits", Command: "Change cpu to 500m and memory to 256Mi"},
	}
}
