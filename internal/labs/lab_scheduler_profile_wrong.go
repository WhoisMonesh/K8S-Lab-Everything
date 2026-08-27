package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&SchedulerProfileWrongLab{})
}

type SchedulerProfileWrongLab struct {
	BaseLab
}

func (l *SchedulerProfileWrongLab) ID() string {
	return "scheduler_profile_wrong"
}

func (l *SchedulerProfileWrongLab) Title() string {
	return "Scheduler Profile Misconfigured"
}

func (l *SchedulerProfileWrongLab) Category() Category {
	return CategoryScheduling
}

func (l *SchedulerProfileWrongLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *SchedulerProfileWrongLab) Description() string {
	return `A Pod is configured with schedulerName: custom-scheduler but no
scheduler with that name exists in the cluster. The pod remains Pending.

Your task: Fix the pod to use the default scheduler.`
}

func (l *SchedulerProfileWrongLab) Hints() []string {
	return []string{
		"Check the pod's schedulerName",
		"List available schedulers in the cluster",
		"Change schedulerName to 'default-scheduler' or remove it",
	}
}

func (l *SchedulerProfileWrongLab) EstimatedTime() int {
	return 10
}

func (l *SchedulerProfileWrongLab) Tags() []string {
	return []string{"scheduler", "scheduler-name", "scheduling"}
}

func (l *SchedulerProfileWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SchedulerProfileWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: custom-scheduled
  namespace: default
spec:
  schedulerName: custom-scheduler
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *SchedulerProfileWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "custom-scheduled",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *SchedulerProfileWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "custom-scheduled",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *SchedulerProfileWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod scheduler",
			Command:     "kubectl get pod custom-scheduled -o yaml | grep schedulerName",
			Notes:       "schedulerName is 'custom-scheduler' which doesn't exist",
		},
		{
			Description: "Check available schedulers",
			Command:     "kubectl get pods -n kube-system | grep scheduler",
			Notes:       "Only 'default-scheduler' exists",
		},
		{
			Description: "Fix scheduler name",
			Command:     "kubectl edit pod custom-scheduled",
			Notes:       "Change schedulerName to 'default-scheduler' or remove the field",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod custom-scheduled",
			Notes:       "Pod should now be Running",
		},
	}
}
