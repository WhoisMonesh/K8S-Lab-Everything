package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&JobBackoffLimitHitLab{})
}

type JobBackoffLimitHitLab struct {
	BaseLab
}

func (l *JobBackoffLimitHitLab) ID() string {
	return "job_backoff_limit_hit"
}

func (l *JobBackoffLimitHitLab) Title() string {
	return "Job Hit backoffLimit"
}

func (l *JobBackoffLimitHitLab) Category() Category {
	return CategoryWorkloads
}

func (l *JobBackoffLimitHitLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *JobBackoffLimitHitLab) Description() string {
	return `A Job 'data-migration' has failed because it hit its backoffLimit.
The job's command has a typo causing it to fail on every attempt.

Your task: Fix the job command and reset the job so it can run successfully.`
}

func (l *JobBackoffLimitHitLab) Hints() []string {
	return []string{
		"Check the job status with kubectl get jobs",
		"Look at the job's pod logs to see the error",
		"Delete the failed job and recreate with the correct command",
	}
}

func (l *JobBackoffLimitHitLab) EstimatedTime() int {
	return 15
}

func (l *JobBackoffLimitHitLab) Tags() []string {
	return []string{"job", "backoff", "batch", "workloads"}
}

func (l *JobBackoffLimitHitLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *JobBackoffLimitHitLab) Break(ctx context.Context, kubeconfigPath string) error {
	job := `apiVersion: batch/v1
kind: Job
metadata:
  name: data-migration
  namespace: default
spec:
  backoffLimit: 3
  template:
    spec:
      containers:
      - name: migrate
        image: busybox:1.36
        command: ['sh', '-c', 'echo "Running migration..." && exi 1']
      restartPolicy: Never
`
	if err := kubectlApply(ctx, kubeconfigPath, job); err != nil {
		return fmt.Errorf("creating job: %w", err)
	}

	time.Sleep(30 * time.Second)

	return nil
}

func (l *JobBackoffLimitHitLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "job", "data-migration",
		"-o", "jsonpath={.status.failed}")
	if err != nil {
		return nil
	}
	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "0" {
		return fmt.Errorf("job should have failed pods")
	}
	return nil
}

func (l *JobBackoffLimitHitLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "job", "data-migration",
		"-o", "jsonpath={.status.succeeded}")
	if err != nil {
		return fmt.Errorf("failed to check job: %w", err)
	}

	if strings.TrimSpace(output) != "1" {
		return fmt.Errorf("job not succeeded (succeeded: %s)", output)
	}

	return nil
}

func (l *JobBackoffLimitHitLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check job status",
			Command:     "kubectl get jobs",
			Notes:       "Job should show 3/3 failures or similar",
		},
		{
			Description: "Check pod logs",
			Command:     "kubectl logs -l job-name=data-migration --tail=20",
			Notes:       "Command has typo: 'exi 1' instead of 'exit 1'",
		},
		{
			Description: "Delete the failed job",
			Command:     "kubectl delete job data-migration",
			Notes:       "Need to delete to reset the backoff counter",
		},
		{
			Description: "Recreate with correct command",
			Command:     "kubectl create job data-migration --from=cronjob/data-migration --dry-run=client -o yaml | sed 's/exi 1/exit 0/' | kubectl apply -f -",
			Notes:       "Fix the typo from 'exi' to 'exit' and use exit 0 for success",
		},
		{
			Description: "Verify job completes",
			Command:     "kubectl wait --for=condition=complete job/data-migration --timeout=60s",
			Notes:       "Job should complete successfully",
		},
	}
}
