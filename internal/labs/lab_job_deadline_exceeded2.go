package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&JobDeadlineExceeded{})
}

type JobDeadlineExceeded struct {
	BaseLab
}

func (l *JobDeadlineExceeded) ID() string            { return "job_deadline_exceeded2" }
func (l *JobDeadlineExceeded) Title() string         { return "Job Killed By activeDeadlineSeconds" }
func (l *JobDeadlineExceeded) Category() Category    { return CategoryWorkloads }
func (l *JobDeadlineExceeded) Difficulty() Difficulty { return DifficultyMedium }
func (l *JobDeadlineExceeded) EstimatedTime() int    { return 15 }
func (l *JobDeadlineExceeded) Tags() []string        { return []string{"jobs", "deadline", "workloads"} }

func (l *JobDeadlineExceeded) Description() string {
	return `A job is failing because activeDeadlineSeconds is too short.
The job is being killed before it can complete. Increase the deadline.`
}

func (l *JobDeadlineExceeded) Hints() []string {
	return []string{
		"Check the job activeDeadlineSeconds",
		"Look at job status and conditions",
		"Increase the deadline to allow completion",
	}
}

func (l *JobDeadlineExceeded) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *JobDeadlineExceeded) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: batch/v1
kind: Job
metadata:
  name: long-job
spec:
  activeDeadlineSeconds: 10
  template:
    spec:
      containers:
      - name: worker
        image: busybox:1.36
        command: ["sh", "-c", "echo Working... && sleep 60 && echo Done"]
      restartPolicy: Never`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *JobDeadlineExceeded) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "job/long-job",
		"-o", "jsonpath={.spec.activeDeadlineSeconds}")
	if err != nil {
		return err
	}
	if output == "10" {
		return fmt.Errorf("deadline still 10 seconds")
	}
	return nil
}

func (l *JobDeadlineExceeded) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check job", Command: "kubectl get job long-job -o yaml"},
		{Description: "Fix deadline", Command: "kubectl patch job long-job -p '{\"spec\":{\"activeDeadlineSeconds\":300}}'"},
	}
}
