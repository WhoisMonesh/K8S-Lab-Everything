package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DeploymentStrategyLab{})
}

type DeploymentStrategyLab struct {
	BaseLab
}

func (l *DeploymentStrategyLab) ID() string {
	return "deployment_wrong_strategy"
}

func (l *DeploymentStrategyLab) Title() string {
	return "Deployment Using Wrong Update Strategy"
}

func (l *DeploymentStrategyLab) Category() Category {
	return CategoryWorkloads
}

func (l *DeploymentStrategyLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DeploymentStrategyLab) Description() string {
	return `A deployment 'web' is using Recreate strategy, causing downtime during updates.
All old pods are killed before new pods start.

Your task: Change the deployment to use RollingUpdate strategy with zero downtime.`
}

func (l *DeploymentStrategyLab) Hints() []string {
	return []string{
		"Check the deployment strategy",
		"Look at spec.strategy.type",
		"Change from Recreate to RollingUpdate",
		"Set maxSurge and maxUnavailable appropriately",
	}
}

func (l *DeploymentStrategyLab) EstimatedTime() int {
	return 15
}

func (l *DeploymentStrategyLab) Tags() []string {
	return []string{"deployment", "strategy", "rolling-update", "workloads"}
}

func (l *DeploymentStrategyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentStrategyLab) Break(ctx context.Context, kubeconfigPath string) error {
	ds := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: web
        image: nginx:1.25-alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, ds); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	return nil
}

func (l *DeploymentStrategyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *DeploymentStrategyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "web",
		"-o", "jsonpath={.spec.strategy.type}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if strings.TrimSpace(output) != "RollingUpdate" {
		return fmt.Errorf("strategy is still %s, expected RollingUpdate", output)
	}
	return nil
}

func (l *DeploymentStrategyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check current strategy",
			Command:     "kubectl get deployment web -o yaml | grep -A 5 strategy",
			Notes:       "Shows type: Recreate",
		},
		{
			Description: "Edit the deployment",
			Command:     "kubectl edit deployment web",
			Notes:       "Change strategy.type from Recreate to RollingUpdate",
		},
		{
			Description: "Add RollingUpdate config",
			Command:     "kubectl patch deployment web -p '{\"spec\":{\"strategy\":{\"type\":\"RollingUpdate\",\"rollingUpdate\":{\"maxSurge\":1,\"maxUnavailable\":0}}}'",
			Notes:       "Set maxSurge=1 and maxUnavailable=0 for zero-downtime updates",
		},
		{
			Description: "Verify strategy change",
			Command:     "kubectl get deployment web -o yaml | grep -A 5 strategy",
			Notes:       "Should show RollingUpdate with proper settings",
		},
	}
}
