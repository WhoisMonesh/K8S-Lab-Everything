package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&JobActiveDeadlineLowLab{})
}

type JobActiveDeadlineLowLab struct {
	BaseLab
}

func (l *JobActiveDeadlineLowLab) ID() string {
	return "job_active_deadline_low"
}

func (l *JobActiveDeadlineLowLab) Title() string {
	return "Job activeDeadlineSeconds Too Short"
}

func (l *JobActiveDeadlineLowLab) Category() Category {
	return CategoryWorkloads
}

func (l *JobActiveDeadlineLowLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *JobActiveDeadlineLowLab) Description() string {
	return `A Job 'data-export' keeps getting terminated because
activeDeadlineSeconds is set too low (10 seconds). The actual export
takes about 30 seconds to complete.

Your task: Increase the activeDeadlineSeconds to allow the job to complete.`
}

func (l *JobActiveDeadlineLowLab) Hints() []string {
	return []string{
		"Check the job configuration",
		"activeDeadlineSeconds kills the job after the specified time",
		"Set it to a value higher than the actual job duration",
	}
}

func (l *JobActiveDeadlineLowLab) EstimatedTime() int {
	return 10
}

func (l *JobActiveDeadlineLowLab) Tags() []string {
	return []string{"job", "deadline", "timeout", "batch", "workloads"}
}

func (l *JobActiveDeadlineLowLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *JobActiveDeadlineLowLab) Break(ctx context.Context, kubeconfigPath string) error {
	job := `apiVersion: batch/v1
kind: Job
metadata:
  name: data-export
  namespace: default
spec:
  activeDeadlineSeconds: 10
  backoffLimit: 5
  template:
    spec:
      containers:
      - name: exporter
        image: busybox:1.36
        command: ['sh', '-c', 'echo "Starting export..." && sleep 30 && echo "Export complete"']
      restartPolicy: Never
`
	if err := kubectlApply(ctx, kubeconfigPath, job); err != nil {
		return fmt.Errorf("creating job: %w", err)
	}

	return nil
}

func (l *JobActiveDeadlineLowLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "job", "data-export",
		"-o", "jsonpath={.status.failed}")
	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "0" {
		return fmt.Errorf("job should have failed due to deadline")
	}
	return nil
}

func (l *JobActiveDeadlineLowLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "job", "data-export",
		"-o", "jsonpath={.status.succeeded}")
	if err != nil {
		return fmt.Errorf("failed to check job: %w", err)
	}

	if strings.TrimSpace(output) != "1" {
		return fmt.Errorf("job not succeeded (succeeded: %s)", output)
	}

	return nil
}

func (l *JobActiveDeadlineLowLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check job status",
			Command:     "kubectl get jobs",
			Notes:       "Job should show failures increasing",
		},
		{
			Description: "Check job configuration",
			Command:     "kubectl get job data-export -o yaml | grep activeDeadlineSeconds",
			Notes:       "activeDeadlineSeconds is 10, but job needs 30 seconds",
		},
		{
			Description: "Delete and recreate the job",
			Command:     "kubectl delete job data-export && kubectl create job data-export --image=busybox:1.36 --dry-run=client -o yaml | sed 's/activeDeadlineSeconds: 10/activeDeadlineSeconds: 60/' | kubectl apply -f -",
			Notes:       "Set activeDeadlineSeconds to 60 to give enough time",
		},
		{
			Description: "Wait for job completion",
			Command:     "kubectl wait --for=condition=complete job/data-export --timeout=120s",
			Notes:       "Job should now complete successfully",
		},
	}
}
