package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&JobDeadlineExceededLab{}) }

type JobDeadlineExceededLab struct{ BaseLab }

func (l *JobDeadlineExceededLab) ID() string          { return "job_deadline_exceeded" }
func (l *JobDeadlineExceededLab) Title() string        { return "Job Killed By activeDeadlineSeconds" }
func (l *JobDeadlineExceededLab) Category() Category   { return CategoryWorkloads }
func (l *JobDeadlineExceededLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *JobDeadlineExceededLab) EstimatedTime() int   { return 15 }
func (l *JobDeadlineExceededLab) Tags() []string {
	return []string{"job", "deadline", "activeDeadline", "workloads"}
}
func (l *JobDeadlineExceededLab) Description() string {
	return `A batch job named 'data-processor' fails immediately with status reason
DeadlineExceeded. The job needs ~60 seconds of work but the activeDeadlineSeconds
is set to 10.

Your task: Increase the deadline so the job can complete. Delete the old
failed job first, then re-apply with the corrected value.`
}
func (l *JobDeadlineExceededLab) Hints() []string {
	return []string{
		"kubectl get job data-processor shows DeadlineExceeded",
		"Check spec.activeDeadlineSeconds — it's too low",
		"Jobs are immutable for certain fields; delete and recreate",
	}
}

func (l *JobDeadlineExceededLab) Break(ctx context.Context, kp string) error {
	job := `apiVersion: batch/v1
kind: Job
metadata:
  name: data-processor
  namespace: default
spec:
  activeDeadlineSeconds: 10
  backoffLimit: 0
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: work
        image: busybox:1.36
        command: ["sh","-c","echo processing...; sleep 60; echo done"]
`
	return kubectlApply(ctx, kp, job)
}

func (l *JobDeadlineExceededLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *JobDeadlineExceededLab) Verify(ctx context.Context, kp string) error {
	deadline, _ := kubectl(ctx, kp, "get", "job", "data-processor", "-o",
		"jsonpath={.spec.activeDeadlineSeconds}")
	if deadline == "10" {
		return fmt.Errorf("deadline is still 10s")
	}
	succeeded, _ := kubectl(ctx, kp, "get", "job", "data-processor", "-o",
		"jsonpath={.status.succeeded}")
	if succeeded != "1" {
		return fmt.Errorf("job has not completed successfully (succeeded: %s)", succeeded)
	}
	return nil
}

func (l *JobDeadlineExceededLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check job status", Command: "kubectl get job data-processor -o yaml | grep -A3 status", Notes: "status.reason = DeadlineExceeded, succeeded = 0"},
		{Description: "Delete the stuck job", Command: "kubectl delete job data-processor", Notes: "Required because activeDeadlineSeconds is immutable on existing pods"},
		{Description: "Recreate with a longer deadline", Command: `kubectl create job data-processor --image=busybox:1.36 -- sh -c "echo processing...; sleep 60; echo done" --dry-run=client -o yaml | sed 's/activeDeadlineSeconds: 10/activeDeadlineSeconds: 120/' | kubectl apply -f -`, Notes: "Or kubectl create from a fixed YAML with activeDeadlineSeconds: 120"},
		{Description: "Verify completion", Command: "kubectl wait --for=condition=complete job/data-processor --timeout=90s", Notes: "Job completes successfully"},
	}
}
