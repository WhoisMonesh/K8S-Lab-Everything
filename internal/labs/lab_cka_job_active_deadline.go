package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&JobActiveDeadlineLab{})
}

type JobActiveDeadlineLab struct {
	BaseLab
}

func (l *JobActiveDeadlineLab) ID() string { return "cka_job_active_deadline" }
func (l *JobActiveDeadlineLab) Title() string {
	return "Set activeDeadlineSeconds for Jobs"
}
func (l *JobActiveDeadlineLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *JobActiveDeadlineLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *JobActiveDeadlineLab) EstimatedTime() int     { return 15 }
func (l *JobActiveDeadlineLab) Tags() []string {
	return []string{"job", "deadline", "timeout"}
}
func (l *JobActiveDeadlineLab) Cert() Cert        { return CertCKA }
func (l *JobActiveDeadlineLab) DomainWeight() int { return 15 }

func (l *JobActiveDeadlineLab) Description() string {
	return `A Job is running indefinitely because it has no activeDeadlineSeconds
set. Add a deadline of 30 seconds to ensure the job is terminated if it
doesn't complete in time.`
}

func (l *JobActiveDeadlineLab) Hints() []string {
	return []string{
		"Check the Job spec",
		"Add activeDeadlineSeconds field",
		"Set the value to 30",
	}
}

func (l *JobActiveDeadlineLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *JobActiveDeadlineLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *JobActiveDeadlineLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "job", "long-running-job",
		"-n", "job-ns", "-o", "jsonpath={.spec.activeDeadlineSeconds}")
	if err != nil {
		return err
	}
	if output == "" || output == "0" {
		return fmt.Errorf("activeDeadlineSeconds not set")
	}
	return nil
}

func (l *JobActiveDeadlineLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check Job spec", Command: "kubectl get job long-running-job -n job-ns -o yaml"},
		{Description: "Patch Job", Command: "kubectl patch job long-running-job -n job-ns -p '{\"spec\":{\"activeDeadlineSeconds\":30}}'"},
		{Description: "Verify", Command: "kubectl get job long-running-job -n job-ns -o yaml"},
	}
}
