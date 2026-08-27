package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ControllerNotSyncingLab{})
}

type ControllerNotSyncingLab struct {
	BaseLab
}

func (l *ControllerNotSyncingLab) ID() string { return "cka_controller_not_syncing" }
func (l *ControllerNotSyncingLab) Title() string {
	return "Debug Controller Manager Issues"
}
func (l *ControllerNotSyncingLab) Category() Category     { return CategoryTroubleshooting }
func (l *ControllerNotSyncingLab) Difficulty() Difficulty { return DifficultyHard }
func (l *ControllerNotSyncingLab) EstimatedTime() int     { return 25 }
func (l *ControllerNotSyncingLab) Tags() []string {
	return []string{"controller-manager", "sync", "troubleshooting"}
}
func (l *ControllerNotSyncingLab) Cert() Cert        { return CertCKA }
func (l *ControllerNotSyncingLab) DomainWeight() int { return 30 }

func (l *ControllerNotSyncingLab) Description() string {
	return `The controller manager is not syncing resources. Deployments are not
creating replicasets. Debug the controller manager by checking its logs
and configuration.`
}

func (l *ControllerNotSyncingLab) Hints() []string {
	return []string{
		"Check controller manager pod status",
		"Review controller manager logs",
		"Verify cluster CIDR and service CIDR settings",
	}
}

func (l *ControllerNotSyncingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ControllerNotSyncingLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ControllerNotSyncingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "controller-ns",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("controller pod not running")
	}
	return nil
}

func (l *ControllerNotSyncingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check controller manager", Command: "kubectl get pods -n kube-system -l component=kube-controller-manager"},
		{Description: "Check logs", Command: "kubectl logs -n kube-system -l component=kube-controller-manager --tail=50"},
		{Description: "Verify config", Command: "Check static pod manifest arguments"},
		{Description: "Restart", Command: "Remove and recreate static pod manifest"},
	}
}
