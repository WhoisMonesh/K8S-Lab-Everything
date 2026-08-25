package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodOOMKilledMemoryLab{})
}

type PodOOMKilledMemoryLab struct {
	BaseLab
}

func (l *PodOOMKilledMemoryLab) ID() string {
	return "pod_oomkilled_memory"
}

func (l *PodOOMKilledMemoryLab) Title() string {
	return "Pod OOMKilled Due to Low Memory Limits"
}

func (l *PodOOMKilledMemoryLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodOOMKilledMemoryLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *PodOOMKilledMemoryLab) Description() string {
	return `A deployment 'cache-server' has pods being OOMKilled repeatedly.
The container's memory limit is too low for the application.

Your task: Fix the memory limits so the pods can run without being OOMKilled.`
}

func (l *PodOOMKilledMemoryLab) Hints() []string {
	return []string{
		"Check the pod status and restart count",
		"Look at the last state of the container",
		"Check the memory limits in the deployment spec",
		"The memory limit is too low for the application",
	}
}

func (l *PodOOMKilledMemoryLab) EstimatedTime() int {
	return 10
}

func (l *PodOOMKilledMemoryLab) Tags() []string {
	return []string{"oomkilled", "memory", "limits", "resources"}
}

func (l *PodOOMKilledMemoryLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodOOMKilledMemoryLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create deployment with very low memory limit
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cache-server
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cache-server
  template:
    metadata:
      labels:
        app: cache-server
    spec:
      containers:
      - name: cache
        image: redis:alpine
        resources:
          requests:
            memory: "10Mi"
            cpu: "50m"
          limits:
            memory: "10Mi"
            cpu: "100m"
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *PodOOMKilledMemoryLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)
	return nil
}

func (l *PodOOMKilledMemoryLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if deployment has all replicas ready
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "cache-server",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) != "2" {
		return fmt.Errorf("deployment not fully ready (ready replicas: %s, expected: 2)", output)
	}

	// Check memory limit
	output, err = kubectl(ctx, kubeconfigPath, "get", "deployment", "cache-server",
		"-o", "jsonpath={.spec.template.spec.containers[0].resources.limits.memory}")
	if err != nil {
		return fmt.Errorf("failed to check memory limit: %w", err)
	}

	// Memory should be at least 50Mi for redis
	memStr := strings.TrimSpace(output)
	if memStr == "10Mi" || memStr == "" {
		return fmt.Errorf("memory limit is still too low (%s)", memStr)
	}

	return nil
}

func (l *PodOOMKilledMemoryLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment status",
			Command:     "kubectl get deployment cache-server",
			Notes:       "Notice the pods keep restarting",
		},
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=cache-server",
			Notes:       "Pods will show high restart counts",
		},
		{
			Description: "Check pod last state",
			Command:     "kubectl get pod -l app=cache-server -o jsonpath='{.items[*].status.containerStatuses[*].lastState.terminated.reason}'",
			Notes:       "The reason will be OOMKilled",
		},
		{
			Description: "Check memory limit",
			Command:     "kubectl get deployment cache-server -o yaml | grep -A 5 resources",
			Notes:       "Memory limit is only 10Mi, which is too low for Redis",
		},
		{
			Description: "Fix memory limits",
			Command:     "kubectl set resources deployment cache-server --limits=memory=256Mi --requests=memory=128Mi",
			Notes:       "Increase memory to a reasonable value for Redis",
		},
		{
			Description: "Wait for rollout",
			Command:     "kubectl rollout status deployment cache-server",
			Notes:       "Wait for new pods with correct memory limits to start",
		},
		{
			Description: "Verify pods are stable",
			Command:     "kubectl get pods -l app=cache-server",
			Notes:       "Pods should be Running with 0 restarts",
		},
	}
}
