package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodPendingSchedulingLab{})
}

type PodPendingSchedulingLab struct {
	BaseLab
}

func (l *PodPendingSchedulingLab) ID() string { return "cka_pod_pending_scheduling" }
func (l *PodPendingSchedulingLab) Title() string {
	return "Debug Pending Pods (Scheduling)"
}
func (l *PodPendingSchedulingLab) Category() Category     { return CategoryTroubleshooting }
func (l *PodPendingSchedulingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodPendingSchedulingLab) EstimatedTime() int     { return 20 }
func (l *PodPendingSchedulingLab) Tags() []string {
	return []string{"pod", "pending", "scheduling", "troubleshooting"}
}
func (l *PodPendingSchedulingLab) Cert() Cert        { return CertCKA }
func (l *PodPendingSchedulingLab) DomainWeight() int { return 30 }

func (l *PodPendingSchedulingLab) Description() string {
	return `A pod is stuck in Pending state due to unsatisfiable scheduling
constraints. Diagnose why the scheduler cannot place the pod and fix
the issue.`
}

func (l *PodPendingSchedulingLab) Hints() []string {
	return []string{
		"Check pod events for scheduling failures",
		"Verify node labels match pod requirements",
		"Check resource availability on nodes",
	}
}

func (l *PodPendingSchedulingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodPendingSchedulingLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *PodPendingSchedulingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "pending-ns",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return err
	}
	if output == "Pending" {
		return fmt.Errorf("pod still pending")
	}
	return nil
}

func (l *PodPendingSchedulingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pods -n pending-ns"},
		{Description: "Describe pod", Command: "kubectl describe pod -n pending-ns <pod-name>"},
		{Description: "Check events", Command: "kubectl get events -n pending-ns --field-selector reason=FailedScheduling"},
		{Description: "Fix scheduling", Command: "Add required labels or resources to nodes"},
	}
}
