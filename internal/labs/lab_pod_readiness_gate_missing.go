package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodReadinessGateMissingLab{})
}

type PodReadinessGateMissingLab struct {
	BaseLab
}

func (l *PodReadinessGateMissingLab) ID() string {
	return "pod_readiness_gate_missing"
}

func (l *PodReadinessGateMissingLab) Title() string {
	return "Pod Readiness Gate Missing"
}

func (l *PodReadinessGateMissingLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodReadinessGateMissingLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *PodReadinessGateMissingLab) Description() string {
	return `A pod 'custom-app' is running but not passing readiness checks because
it's missing a required readiness gate. The application requires a custom
readiness condition to be set before it can receive traffic.

Your task: Add the missing readiness gate to the pod spec.`
}

func (l *PodReadinessGateMissingLab) Hints() []string {
	return []string{
		"Check the pod conditions",
		"Readiness gates allow custom conditions for pod readiness",
		"The readinessGate field goes in spec.readinessGates",
	}
}

func (l *PodReadinessGateMissingLab) EstimatedTime() int {
	return 15
}

func (l *PodReadinessGateMissingLab) Tags() []string {
	return []string{"pod", "readiness", "readiness-gate", "workloads"}
}

func (l *PodReadinessGateMissingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodReadinessGateMissingLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: custom-app
  namespace: default
  labels:
    app: custom-app
spec:
  containers:
  - name: app
    image: nginx:alpine
    readinessProbe:
      httpGet:
        path: /
        port: 80
      initialDelaySeconds: 5
      periodSeconds: 5
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodReadinessGateMissingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodReadinessGateMissingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "custom-app",
		"-o", "jsonpath={.spec.readinessGates}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "null" {
		return fmt.Errorf("readinessGates not configured")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "custom-app",
		"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return fmt.Errorf("failed to check pod conditions: %w", err)
	}

	if strings.TrimSpace(output) != "True" {
		return fmt.Errorf("pod is not ready")
	}

	return nil
}

func (l *PodReadinessGateMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod custom-app",
			Notes:       "Pod should be running but not ready",
		},
		{
			Description: "Check pod conditions",
			Command:     "kubectl get pod custom-app -o jsonpath='{.status.conditions}'",
			Notes:       "Look for custom conditions that might be False",
		},
		{
			Description: "Add readiness gate",
			Command:     "kubectl edit pod custom-app",
			Notes:       "Add spec.readinessGates with the required condition type",
		},
		{
			Description: "Verify pod is ready",
			Command:     "kubectl get pod custom-app",
			Notes:       "Pod should now show 1/1 READY",
		},
	}
}
