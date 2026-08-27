package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CronJobConcurrencyForbidLab{})
}

type CronJobConcurrencyForbidLab struct {
	BaseLab
}

func (l *CronJobConcurrencyForbidLab) ID() string {
	return "cronjob_concurrency_forbid"
}

func (l *CronJobConcurrencyForbidLab) Title() string {
	return "CronJob Concurrency Policy Conflict"
}

func (l *CronJobConcurrencyForbidLab) Category() Category {
	return CategoryWorkloads
}

func (l *CronJobConcurrencyForbidLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *CronJobConcurrencyForbidLab) Description() string {
	return `A CronJob 'report-gen' has concurrencyPolicy set to Forbid but the
schedule is too frequent (every minute), causing most runs to be skipped.
The job needs to run at least every 5 minutes.

Your task: Fix the CronJob schedule to prevent excessive skipping.`
}

func (l *CronJobConcurrencyForbidLab) Hints() []string {
	return []string{
		"Check the CronJob schedule and concurrency policy",
		"Forbid policy skips new runs if previous is still running",
		"Adjust the schedule to give enough time between runs",
	}
}

func (l *CronJobConcurrencyForbidLab) EstimatedTime() int {
	return 10
}

func (l *CronJobConcurrencyForbidLab) Tags() []string {
	return []string{"cronjob", "concurrency", "schedule", "batch", "workloads"}
}

func (l *CronJobConcurrencyForbidLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CronJobConcurrencyForbidLab) Break(ctx context.Context, kubeconfigPath string) error {
	cronjob := `apiVersion: batch/v1
kind: CronJob
metadata:
  name: report-gen
  namespace: default
spec:
  schedule: "*/1 * * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      activeDeadlineSeconds: 120
      template:
        spec:
          containers:
          - name: reporter
            image: busybox:1.36
            command: ['sh', '-c', 'echo "Generating report..." && sleep 90 && echo "Report done"']
          restartPolicy: Never
`
	if err := kubectlApply(ctx, kubeconfigPath, cronjob); err != nil {
		return fmt.Errorf("creating cronjob: %w", err)
	}

	return nil
}

func (l *CronJobConcurrencyForbidLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *CronJobConcurrencyForbidLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "report-gen",
		"-o", "jsonpath={.spec.schedule}")
	if err != nil {
		return fmt.Errorf("failed to check cronjob: %w", err)
	}

	if strings.TrimSpace(output) == "*/1 * * * *" {
		return fmt.Errorf("schedule is still every minute")
	}

	return nil
}

func (l *CronJobConcurrencyForbidLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check CronJob status",
			Command:     "kubectl get cronjob report-gen",
			Notes:       "SCHEDULE should show every minute",
		},
		{
			Description: "Check job history",
			Command:     "kubectl get jobs -l job-name=report-gen",
			Notes:       "Many jobs should be failing or skipped",
		},
		{
			Description: "Fix the schedule",
			Command:     "kubectl patch cronjob report-gen --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/schedule\",\"value\":\"*/5 * * * *\"}]'",
			Notes:       "Change to every 5 minutes to allow jobs to complete",
		},
		{
			Description: "Verify schedule is updated",
			Command:     "kubectl get cronjob report-gen",
			Notes:       "SCHEDULE should now show */5 * * * *",
		},
	}
}
