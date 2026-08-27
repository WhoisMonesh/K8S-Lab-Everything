package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodLifecycleIssuesLab{})
}

type PodLifecycleIssuesLab struct {
	BaseLab
}

func (l *PodLifecycleIssuesLab) ID() string { return "cka_pod_lifecycle_issues" }
func (l *PodLifecycleIssuesLab) Title() string {
	return "Debug Pod Lifecycle Problems"
}
func (l *PodLifecycleIssuesLab) Category() Category     { return CategoryTroubleshooting }
func (l *PodLifecycleIssuesLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodLifecycleIssuesLab) EstimatedTime() int     { return 20 }
func (l *PodLifecycleIssuesLab) Tags() []string {
	return []string{"pod", "lifecycle", "probes", "troubleshooting"}
}
func (l *PodLifecycleIssuesLab) Cert() Cert        { return CertCKA }
func (l *PodLifecycleIssuesLab) DomainWeight() int { return 30 }

func (l *PodLifecycleIssuesLab) Description() string {
	return `A pod is failing readiness probes and never becomes ready. Debug the
probe configuration and fix the readiness probe to match the actual
application health endpoint.`
}

func (l *PodLifecycleIssuesLab) Hints() []string {
	return []string{
		"Check pod readiness probe configuration",
		"Verify the probe endpoint is accessible",
		"Check probe timing parameters",
	}
}

func (l *PodLifecycleIssuesLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodLifecycleIssuesLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *PodLifecycleIssuesLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "lifecycle-ns",
		"-o", "jsonpath={.items[0].status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return err
	}
	if output != "True" {
		return fmt.Errorf("pod not ready")
	}
	return nil
}

func (l *PodLifecycleIssuesLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pods -n lifecycle-ns"},
		{Description: "Describe pod", Command: "kubectl describe pod -n lifecycle-ns <pod-name>"},
		{Description: "Check probe config", Command: "kubectl get pod -n lifecycle-ns -o yaml | grep -A10 readinessProbe"},
		{Description: "Fix probe", Command: "Update readiness probe to correct endpoint and timing"},
	}
}
