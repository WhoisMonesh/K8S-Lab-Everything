package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&SchedulerExtenderMissingLab{})
}

type SchedulerExtenderMissingLab struct {
	BaseLab
}

func (l *SchedulerExtenderMissingLab) ID() string {
	return "scheduler_extender_missing"
}

func (l *SchedulerExtenderMissingLab) Title() string {
	return "Scheduler Extender Not Available"
}

func (l *SchedulerExtenderMissingLab) Category() Category {
	return CategoryScheduling
}

func (l *SchedulerExtenderMissingLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *SchedulerExtenderMissingLab) Description() string {
	return `A scheduler configuration references an extender service at
http://extender-svc:9001 that doesn't exist. Pods are stuck in Pending
because the scheduler cannot reach the extender.

Your task: Fix the scheduler configuration to remove or fix the
extender reference.`
}

func (l *SchedulerExtenderMissingLab) Hints() []string {
	return []string{
		"Check the scheduler configuration",
		"The extender endpoint doesn't exist",
		"Remove the extender configuration from the scheduler config",
	}
}

func (l *SchedulerExtenderMissingLab) EstimatedTime() int {
	return 15
}

func (l *SchedulerExtenderMissingLab) Tags() []string {
	return []string{"scheduler", "extender", "scheduling"}
}

func (l *SchedulerExtenderMissingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SchedulerExtenderMissingLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: stuck-scheduler
  namespace: default
spec:
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

func (l *SchedulerExtenderMissingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *SchedulerExtenderMissingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "stuck-scheduler",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *SchedulerExtenderMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check scheduler configuration",
			Command:     "kubectl get cm -n kube-system | grep scheduler",
			Notes:       "Look for scheduler configuration configmap",
		},
		{
			Description: "Check scheduler logs",
			Command:     "kubectl logs -n kube-system -l component=kube-scheduler --tail=50",
			Notes:       "Look for extender connection errors",
		},
		{
			Description: "Fix scheduler config",
			Command:     "kubectl edit cm -n kube-system kube-scheduler-config",
			Notes:       "Remove or fix the extender configuration section",
		},
		{
			Description: "Restart scheduler",
			Command:     "kubectl rollout restart deployment kube-scheduler -n kube-system",
			Notes:       "Restart to apply new configuration",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod stuck-scheduler",
			Notes:       "Pod should now be Running",
		},
	}
}
