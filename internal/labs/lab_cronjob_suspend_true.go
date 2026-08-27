package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CronJobSuspendTrueLab{})
}

type CronJobSuspendTrueLab struct {
	BaseLab
}

func (l *CronJobSuspendTrueLab) ID() string {
	return "cronjob_suspend_true"
}

func (l *CronJobSuspendTrueLab) Title() string {
	return "CronJob Suspended"
}

func (l *CronJobSuspendTrueLab) Category() Category {
	return CategoryWorkloads
}

func (l *CronJobSuspendTrueLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *CronJobSuspendTrueLab) Description() string {
	return `A CronJob 'backup-job' is configured but never creates any Jobs.
The CronJob has suspend set to true, preventing it from running.

Your task: Unsuspend the CronJob so it can run on schedule.`
}

func (l *CronJobSuspendTrueLab) Hints() []string {
	return []string{
		"Check the CronJob configuration",
		"Look for the suspend field in the CronJob spec",
		"Set suspend to false to allow the CronJob to run",
	}
}

func (l *CronJobSuspendTrueLab) EstimatedTime() int {
	return 5
}

func (l *CronJobSuspendTrueLab) Tags() []string {
	return []string{"cronjob", "suspend", "batch", "workloads"}
}

func (l *CronJobSuspendTrueLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CronJobSuspendTrueLab) Break(ctx context.Context, kubeconfigPath string) error {
	cronjob := `apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-job
  namespace: default
spec:
  schedule: "*/1 * * * *"
  suspend: true
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: busybox:1.36
            command: ['sh', '-c', 'echo "Backup completed"']
          restartPolicy: Never
`
	if err := kubectlApply(ctx, kubeconfigPath, cronjob); err != nil {
		return fmt.Errorf("creating cronjob: %w", err)
	}

	return nil
}

func (l *CronJobSuspendTrueLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *CronJobSuspendTrueLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "backup-job",
		"-o", "jsonpath={.spec.suspend}")
	if err != nil {
		return fmt.Errorf("failed to check cronjob: %w", err)
	}

	if strings.TrimSpace(output) == "true" {
		return fmt.Errorf("cronjob is still suspended")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "jobs",
		"-l", "job-name=backup-job", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to check jobs: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no jobs created yet")
	}

	return nil
}

func (l *CronJobSuspendTrueLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check CronJob status",
			Command:     "kubectl get cronjob backup-job",
			Notes:       "SUSPEND column should show 'True'",
		},
		{
			Description: "Unsuspend the CronJob",
			Command:     "kubectl patch cronjob backup-job --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/suspend\",\"value\":false}]'",
			Notes:       "Setting suspend to false allows the CronJob to run",
		},
		{
			Description: "Verify CronJob is running",
			Command:     "kubectl get cronjob backup-job",
			Notes:       "SUSPEND column should now show 'False'",
		},
		{
			Description: "Check for created Jobs",
			Command:     "kubectl get jobs -l job-name=backup-job",
			Notes:       "A new Job should be created within a minute",
		},
	}
}
