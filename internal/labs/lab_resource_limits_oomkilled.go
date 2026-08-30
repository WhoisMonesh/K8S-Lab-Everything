package labs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register(&ResourceLimitsOOMKilledLab{})
}

type ResourceLimitsOOMKilledLab struct {
	BaseLab
}

func (l *ResourceLimitsOOMKilledLab) ID() string {
	return "resource_limits_oomkilled"
}

func (l *ResourceLimitsOOMKilledLab) Title() string {
	return "Resource Limits Causing OOMKilled"
}

func (l *ResourceLimitsOOMKilledLab) Category() Category {
	return CategoryTroubleshooting
}

func (l *ResourceLimitsOOMKilledLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ResourceLimitsOOMKilledLab) Description() string {
	return `A pod keeps getting OOMKilled because its memory limit is too low
for the workload. The pod restarts repeatedly with exit code 137.

Your task: Increase the memory limit to prevent OOMKilled errors while
keeping resource usage reasonable.`
}

func (l *ResourceLimitsOOMKilledLab) Hints() []string {
	return []string{
		"Check the pod status for OOMKilled restarts",
		"Look at the current memory limits",
		"Exit code 137 means the process was killed by OOM killer",
	}
}

func (l *ResourceLimitsOOMKilledLab) EstimatedTime() int {
	return 15
}

func (l *ResourceLimitsOOMKilledLab) Tags() []string {
	return []string{"oomkilled", "resource-limits", "memory", "troubleshooting"}
}

func (l *ResourceLimitsOOMKilledLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceLimitsOOMKilledLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-hungry
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: memory-hungry
  template:
    metadata:
      labels:
        app: memory-hungry
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'dd if=/dev/zero bs=1M count=50 2>/dev/null; while true; do echo working; sleep 15; done']
        resources:
          limits:
            memory: "10Mi"
            cpu: 100m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *ResourceLimitsOOMKilledLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=memory-hungry",
		"-o", "jsonpath={.items[*].status.containerStatuses[0].restartCount}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	restarts, _ := strconv.Atoi(strings.TrimSpace(output))
	if restarts > 0 {
		return nil
	}

	return fmt.Errorf("no restarts yet (expected OOMKilled)")
}

func (l *ResourceLimitsOOMKilledLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "memory-hungry",
		"-o", "jsonpath={.spec.template.spec.containers[0].resources.limits.memory}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	memLimit := strings.TrimSpace(output)
	if memLimit == "10Mi" || memLimit == "" {
		return fmt.Errorf("memory limit still too low (%s)", memLimit)
	}

	return nil
}

func (l *ResourceLimitsOOMKilledLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=memory-hungry",
			Notes:       "Look for OOMKilled in RESTARTS column",
		},
		{
			Description: "Check container memory limits",
			Command:     "kubectl get deploy memory-hungry -o yaml | grep -A 3 resources",
			Notes:       "Memory limit is only 10Mi - too low",
		},
		{
			Description: "Fix: Increase memory limit",
			Command:     `kubectl patch deploy memory-hungry --type='json' -p='[{"op":"replace","path":"/spec/template/spec/containers/0/resources/limits/memory","value":"128Mi"}]'`,
			Notes:       "Increase to 128Mi to handle the workload",
		},
		{
			Description: "Verify pods are running",
			Command:     "kubectl rollout status deploy/memory-hungry",
			Notes:       "Pod should stabilize without OOMKilled restarts",
		},
	}
}
