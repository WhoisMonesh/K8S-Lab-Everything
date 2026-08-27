package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADDeploymentStrategyRollingLab{})
}

type CKADDeploymentStrategyRollingLab struct {
	BaseLab
}

func (l *CKADDeploymentStrategyRollingLab) ID() string {
	return "ckad_deployment_strategy_rolling"
}

func (l *CKADDeploymentStrategyRollingLab) Title() string {
	return "Configure Rolling Update Strategy"
}

func (l *CKADDeploymentStrategyRollingLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADDeploymentStrategyRollingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADDeploymentStrategyRollingLab) Cert() Cert             { return CertCKAD }
func (l *CKADDeploymentStrategyRollingLab) DomainWeight() int      { return 20 }
func (l *CKADDeploymentStrategyRollingLab) EstimatedTime() int     { return 20 }
func (l *CKADDeploymentStrategyRollingLab) Tags() []string {
	return []string{"rolling-update", "strategy", "deployment"}
}

func (l *CKADDeploymentStrategyRollingLab) Description() string {
	return `A deployment needs a specific rolling update strategy. Configure it to
use maxSurge: 25% and maxUnavailable: 25% for controlled rollouts.

Your task: Update the deployment strategy configuration.`
}

func (l *CKADDeploymentStrategyRollingLab) Hints() []string {
	return []string{
		"Use spec.strategy.type: RollingUpdate",
		"Set maxSurge and maxUnavailable in strategy.rollingUpdate",
		"Values can be percentages or absolute numbers",
	}
}

func (l *CKADDeploymentStrategyRollingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADDeploymentStrategyRollingLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
spec:
  replicas: 4
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: nginx:1.21
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADDeploymentStrategyRollingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.strategy.type}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.TrimSpace(output) != "RollingUpdate" {
		return fmt.Errorf("strategy type is not RollingUpdate (current: %s)", output)
	}
	return nil
}

func (l *CKADDeploymentStrategyRollingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current strategy", Command: "kubectl get deployment webapp -o yaml | grep -A 5 strategy"},
		{Description: "Edit deployment", Command: "kubectl edit deployment webapp"},
		{Description: "Update strategy", Command: "Change strategy type to RollingUpdate and set maxSurge/maxUnavailable to 25%"},
		{Description: "Verify strategy", Command: "kubectl get deployment webapp -o yaml | grep -A 5 strategy"},
	}
}
