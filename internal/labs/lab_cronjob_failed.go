package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CronJobFailedLab{})
}

type CronJobFailedLab struct {
	BaseLab
}

func (l *CronJobFailedLab) ID() string {
	return "cronjob_failed"
}

func (l *CronJobFailedLab) Title() string {
	return "CronJob Not Creating Pods"
}

func (l *CronJobFailedLab) Category() Category {
	return CategoryWorkloads
}

func (l *CronJobFailedLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *CronJobFailedLab) Description() string {
	return `A CronJob 'backup-job' is configured but not creating any jobs or pods.
The CronJob schedule is correct but jobs fail to start.

Your task: Fix the CronJob configuration so it creates jobs successfully.`
}

func (l *CronJobFailedLab) Hints() []string {
	return []string{
		"Check the CronJob status",
		"Look at the CronJob events",
		"Check the job template spec",
		"The container image or command might be wrong",
	}
}

func (l *CronJobFailedLab) EstimatedTime() int {
	return 20
}

func (l *CronJobFailedLab) Tags() []string {
	return []string{"cronjob", "batch", "scheduling", "workloads"}
}

func (l *CronJobFailedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CronJobFailedLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create CronJob with broken image
	cronjob := `apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-job
  namespace: default
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: busybox:99.99-doesnotexist
            command: ['sh', '-c', 'echo "Running backup..." && date']
          restartPolicy: OnFailure
      backoffLimit: 3
`
	if err := kubectlApply(ctx, kubeconfigPath, cronjob); err != nil {
		return fmt.Errorf("creating CronJob: %w", err)
	}

	return nil
}

func (l *CronJobFailedLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(70 * time.Second)
	return nil
}

func (l *CronJobFailedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if CronJob has created successful jobs
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "backup-job",
		"-o", "jsonpath={.status.lastScheduleTime}")
	if err != nil {
		return fmt.Errorf("failed to check CronJob: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("CronJob has not scheduled any jobs")
	}

	// Check if there are successful jobs
	output, err = kubectl(ctx, kubeconfigPath, "get", "jobs",
		"-l", "job-name=backup-job",
		"-o", "jsonpath={.items[*].status.conditions[*].type}")
	if err != nil {
		return fmt.Errorf("failed to check jobs: %w", err)
	}

	if !strings.Contains(output, "Complete") {
		return fmt.Errorf("no successful jobs found")
	}

	return nil
}

func (l *CronJobFailedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check CronJob status",
			Command:     "kubectl get cronjob backup-job",
			Notes:       "The CronJob exists but might show 0 successful jobs",
		},
		{
			Description: "Check CronJob events",
			Command:     "kubectl describe cronjob backup-job",
			Notes:       "Look for events about failed jobs",
		},
		{
			Description: "Check recent jobs",
			Command:     "kubectl get jobs -l job-name=backup-job",
			Notes:       "Jobs will show 0/1 completions or failures",
		},
		{
			Description: "Check job pod logs",
			Command:     "kubectl logs -l job-name=backup-job --tail=20",
			Notes:       "Look for ImagePullBackOff errors",
		},
		{
			Description: "Fix the CronJob image",
			Command:     "kubectl edit cronjob backup-job",
			Notes:       "Change the image from busybox:99.99-doesnotexist to busybox:1.28",
		},
		{
			Description: "Wait for next scheduled run",
			Command:     "kubectl get cronjob backup-job -w",
			Notes:       "Wait for the next minute to see if the job runs successfully",
		},
		{
			Description: "Verify job completed",
			Command:     "kubectl get jobs -l job-name=backup-job",
			Notes:       "Look for a job with 1/1 completions",
		},
	}
}
