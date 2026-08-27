package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DeploymentRollingUpdateLab{})
}

type DeploymentRollingUpdateLab struct {
	BaseLab
}

func (l *DeploymentRollingUpdateLab) ID() string { return "cka_deployment_rolling_update" }
func (l *DeploymentRollingUpdateLab) Title() string {
	return "Perform Rolling Update with Zero Downtime"
}
func (l *DeploymentRollingUpdateLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *DeploymentRollingUpdateLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *DeploymentRollingUpdateLab) EstimatedTime() int     { return 20 }
func (l *DeploymentRollingUpdateLab) Tags() []string {
	return []string{"deployment", "rolling-update", "zero-downtime"}
}
func (l *DeploymentRollingUpdateLab) Cert() Cert        { return CertCKA }
func (l *DeploymentRollingUpdateLab) DomainWeight() int { return 15 }

func (l *DeploymentRollingUpdateLab) Description() string {
	return `A deployment is currently broken with maxSurge=0 and maxUnavailable=1,
causing downtime during updates. Fix the deployment strategy to ensure zero
downtime during rolling updates.`
}

func (l *DeploymentRollingUpdateLab) Hints() []string {
	return []string{
		"Check the deployment strategy",
		"Set maxSurge to at least 1",
		"Set maxUnavailable to 0 for zero downtime",
	}
}

func (l *DeploymentRollingUpdateLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentRollingUpdateLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *DeploymentRollingUpdateLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "rolling-app",
		"-n", "rolling-ns", "-o",
		"jsonpath={.spec.strategy.rollingUpdate.maxSurge}")
	if err != nil {
		return err
	}
	if output == "0" {
		return fmt.Errorf("maxSurge still set to 0")
	}
	return nil
}

func (l *DeploymentRollingUpdateLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check deployment strategy", Command: "kubectl get deployment rolling-app -n rolling-ns -o yaml"},
		{Description: "Fix strategy", Command: "kubectl patch deployment rolling-app -n rolling-ns -p '{\"spec\":{\"strategy\":{\"rollingUpdate\":{\"maxSurge\":1,\"maxUnavailable\":0}}}}'"},
		{Description: "Trigger update", Command: "kubectl set image deployment/rolling-app app=nginx:1.25 -n rolling-ns"},
		{Description: "Verify rollout", Command: "kubectl rollout status deployment/rolling-app -n rolling-ns"},
	}
}
