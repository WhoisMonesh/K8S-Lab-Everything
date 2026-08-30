package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CronJobSchedulingLab{})
}

type CronJobSchedulingLab struct {
	BaseLab
}

func (l *CronJobSchedulingLab) ID() string {
	return "cronjob_scheduling"
}

func (l *CronJobSchedulingLab) Title() string {
	return "CronJob Scheduling Issues"
}

func (l *CronJobSchedulingLab) Category() Category {
	return CategoryWorkloadsScheduling
}

func (l *CronJobSchedulingLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *CronJobSchedulingLab) Description() string {
	return `A CronJob is configured but never runs. The schedule looks correct
but jobs are not being created.

Your task: Diagnose and fix why the CronJob is not scheduling any jobs.`
}

func (l *CronJobSchedulingLab) Hints() []string {
	return []string{
		"Check the CronJob schedule format",
		"Verify the CronJob status with kubectl get cronjob",
		"Check if there are any failed job history entries",
	}
}

func (l *CronJobSchedulingLab) EstimatedTime() int {
	return 15
}

func (l *CronJobSchedulingLab) Tags() []string {
	return []string{"cronjob", "scheduling", "batch", "workloads"}
}

func (l *CronJobSchedulingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CronJobSchedulingLab) Break(ctx context.Context, kubeconfigPath string) error {
	cronjob := `apiVersion: batch/v1
kind: CronJob
metadata:
  name: scheduled-task
  namespace: default
spec:
  schedule: "60 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: task
            image: busybox:1.36
            command: ['sh', '-c', 'echo task completed']
          restartPolicy: OnFailure
`
	return kubectlApply(ctx, kubeconfigPath, cronjob)
}

func (l *CronJobSchedulingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "scheduled-task",
		"-o", "jsonpath={.spec.schedule}")
	if err != nil {
		return fmt.Errorf("checking cronjob: %w", err)
	}

	schedule := strings.TrimSpace(output)
	if schedule == "60 * * * *" {
		return nil
	}

	return fmt.Errorf("schedule is valid (expected broken)")
}

func (l *CronJobSchedulingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "scheduled-task",
		"-o", "jsonpath={.spec.schedule}")
	if err != nil {
		return fmt.Errorf("checking cronjob: %w", err)
	}

	schedule := strings.TrimSpace(output)
	if schedule == "60 * * * *" {
		return fmt.Errorf("CronJob schedule is still invalid (60 is not a valid minute)")
	}

	return nil
}

func (l *CronJobSchedulingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check CronJob schedule",
			Command:     "kubectl get cronjob scheduled-task -o yaml | grep schedule",
			Notes:       "Notice the schedule is '60 * * * *' - minute must be 0-59",
		},
		{
			Description: "Fix: Correct the schedule",
			Command:     `kubectl patch cronjob scheduled-task --type='json' -p='[{"op":"replace","path":"/spec/schedule","value":"0 * * * *"}]'`,
			Notes:       "Use minute 0 to run at the start of each hour",
		},
		{
			Description: "Verify CronJob is scheduled",
			Command:     "kubectl get cronjob scheduled-task",
			Notes:       "SCHEDULE column should show the corrected time",
		},
	}
}
