package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DeploymentWrongStrategy2{})
}

type DeploymentWrongStrategy2 struct {
	BaseLab
}

func (l *DeploymentWrongStrategy2) ID() string             { return "deployment_wrong_strategy2" }
func (l *DeploymentWrongStrategy2) Title() string          { return "Deployment Using Wrong Update Strategy" }
func (l *DeploymentWrongStrategy2) Category() Category     { return CategoryWorkloads }
func (l *DeploymentWrongStrategy2) Difficulty() Difficulty { return DifficultyMedium }
func (l *DeploymentWrongStrategy2) EstimatedTime() int     { return 15 }
func (l *DeploymentWrongStrategy2) Tags() []string {
	return []string{"deployment", "strategy", "rolling"}
}

func (l *DeploymentWrongStrategy2) Description() string {
	return `A deployment is using Recreate strategy but should use RollingUpdate.
Change the strategy to RollingUpdate for zero-downtime deployments.`
}

func (l *DeploymentWrongStrategy2) Hints() []string {
	return []string{
		"Check the deployment strategy",
		"Look at the strategy type",
		"Change from Recreate to RollingUpdate",
	}
}

func (l *DeploymentWrongStrategy2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentWrongStrategy2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: strategy-app
spec:
  replicas: 3
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: strategy-app
  template:
    metadata:
      labels:
        app: strategy-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.19
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *DeploymentWrongStrategy2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/strategy-app",
		"-o", "jsonpath={.spec.strategy.type}")
	if err != nil {
		return err
	}
	if output == "Recreate" {
		return fmt.Errorf("strategy still Recreate")
	}
	return nil
}

func (l *DeploymentWrongStrategy2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check strategy", Command: "kubectl get deploy strategy-app -o jsonpath='{.spec.strategy.type}'"},
		{Description: "Fix strategy", Command: "kubectl patch deploy strategy-app -p '{\"spec\":{\"strategy\":{\"type\":\"RollingUpdate\",\"rollingUpdate\":{\"maxSurge\":1,\"maxUnavailable\":1}}}}'"},
	}
}
