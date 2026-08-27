package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodTerminationGraceZeroLab{})
}

type PodTerminationGraceZeroLab struct {
	BaseLab
}

func (l *PodTerminationGraceZeroLab) ID() string {
	return "pod_termination_grace_zero"
}

func (l *PodTerminationGraceZeroLab) Title() string {
	return "Pod terminationGracePeriodSeconds=0"
}

func (l *PodTerminationGraceZeroLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodTerminationGraceZeroLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *PodTerminationGraceZeroLab) Description() string {
	return `A pod 'graceful-app' has terminationGracePeriodSeconds set to 0.
This causes the pod to be killed immediately without allowing time for
graceful shutdown, which can lead to data loss.

Your task: Set a reasonable terminationGracePeriodSeconds value.`
}

func (l *PodTerminationGraceZeroLab) Hints() []string {
	return []string{
		"Check the pod configuration",
		"terminationGracePeriodSeconds=0 gives no time for cleanup",
		"A typical value is 30 seconds",
	}
}

func (l *PodTerminationGraceZeroLab) EstimatedTime() int {
	return 5
}

func (l *PodTerminationGraceZeroLab) Tags() []string {
	return []string{"pod", "termination", "grace-period", "workloads"}
}

func (l *PodTerminationGraceZeroLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodTerminationGraceZeroLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: graceful-app
  namespace: default
spec:
  terminationGracePeriodSeconds: 0
  containers:
  - name: app
    image: nginx:alpine
    ports:
    - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodTerminationGraceZeroLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *PodTerminationGraceZeroLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "graceful-app",
		"-o", "jsonpath={.spec.terminationGracePeriodSeconds}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "0" {
		return fmt.Errorf("terminationGracePeriodSeconds is still 0")
	}

	return nil
}

func (l *PodTerminationGraceZeroLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod configuration",
			Command:     "kubectl get pod graceful-app -o yaml | grep terminationGracePeriodSeconds",
			Notes:       "terminationGracePeriodSeconds is 0",
		},
		{
			Description: "Fix termination grace period",
			Command:     "kubectl edit pod graceful-app",
			Notes:       "Change terminationGracePeriodSeconds from 0 to 30",
		},
		{
			Description: "Verify configuration",
			Command:     "kubectl get pod graceful-app -o yaml | grep terminationGracePeriodSeconds",
			Notes:       "Should now be 30",
		},
	}
}
