package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DaemonSetStrategyLab{})
}

type DaemonSetStrategyLab struct {
	BaseLab
}

func (l *DaemonSetStrategyLab) ID() string             { return "cka_daemonset_strategy" }
func (l *DaemonSetStrategyLab) Title() string          { return "Configure DaemonSet Update Strategy" }
func (l *DaemonSetStrategyLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *DaemonSetStrategyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *DaemonSetStrategyLab) EstimatedTime() int     { return 20 }
func (l *DaemonSetStrategyLab) Tags() []string {
	return []string{"daemonset", "update-strategy", "rolling"}
}
func (l *DaemonSetStrategyLab) Cert() Cert        { return CertCKA }
func (l *DaemonSetStrategyLab) DomainWeight() int { return 15 }

func (l *DaemonSetStrategyLab) Description() string {
	return `A DaemonSet is using OnDelete strategy which requires manual pod deletion.
Change it to RollingUpdate with a maxUnavailable of 1 to automate updates.`
}

func (l *DaemonSetStrategyLab) Hints() []string {
	return []string{
		"Check the DaemonSet update strategy",
		"Change type from OnDelete to RollingUpdate",
		"Set maxUnavailable appropriately",
	}
}

func (l *DaemonSetStrategyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DaemonSetStrategyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *DaemonSetStrategyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "daemonset", "logging-ds",
		"-n", "daemonset-ns", "-o",
		"jsonpath={.spec.updateStrategy.type}")
	if err != nil {
		return err
	}
	if output == "OnDelete" {
		return fmt.Errorf("DaemonSet still using OnDelete strategy")
	}
	return nil
}

func (l *DaemonSetStrategyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check DaemonSet strategy", Command: "kubectl get daemonset logging-ds -n daemonset-ns -o yaml"},
		{Description: "Patch DaemonSet", Command: "kubectl patch daemonset logging-ds -n daemonset-ns -p '{\"spec\":{\"updateStrategy\":{\"type\":\"RollingUpdate\",\"rollingUpdate\":{\"maxUnavailable\":1}}}}'"},
		{Description: "Trigger update", Command: "kubectl set image daemonset/logging-ds logging=fluentd:v1.16 -n daemonset-ns"},
	}
}
