package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DeploymentStrategyRollingMaxUnavailableLab{})
}

type DeploymentStrategyRollingMaxUnavailableLab struct {
	BaseLab
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) ID() string {
	return "deployment_strategy_rolling_max_unavailable"
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Title() string {
	return "Deployment maxUnavailable Too Aggressive"
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Category() Category {
	return CategoryWorkloads
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Description() string {
	return `A deployment 'frontend' has maxUnavailable set to 100%, which means
all pods are terminated before new ones are created. This causes downtime
during rolling updates.

Your task: Fix the deployment strategy to maintain availability during updates.`
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Hints() []string {
	return []string{
		"Check the deployment strategy",
		"maxUnavailable=100% terminates all old pods before creating new ones",
		"Use maxUnavailable=25% or a fixed number like 1",
	}
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) EstimatedTime() int {
	return 10
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Tags() []string {
	return []string{"deployment", "rolling-update", "maxUnavailable", "workloads"}
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: default
spec:
  replicas: 4
  selector:
    matchLabels:
      app: frontend
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 100%
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: web
        image: nginx:1.19
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	time.Sleep(10 * time.Second)

	_, _ = kubectl(ctx, kubeconfigPath, "set", "image", "deployment/frontend",
		"web=nginx:1.20")

	return nil
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "frontend",
		"-o", "jsonpath={.spec.strategy.rollingUpdate.maxUnavailable}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "100%" {
		return fmt.Errorf("maxUnavailable is still 100%%")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "deployment", "frontend",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check ready replicas: %w", err)
	}

	if strings.TrimSpace(output) != "4" {
		return fmt.Errorf("deployment not fully ready (ready replicas: %s)", output)
	}

	return nil
}

func (l *DeploymentStrategyRollingMaxUnavailableLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment strategy",
			Command:     "kubectl get deployment frontend -o yaml | grep -A 5 strategy",
			Notes:       "maxUnavailable is 100% which is too aggressive",
		},
		{
			Description: "Fix maxUnavailable",
			Command:     "kubectl patch deployment frontend --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/strategy/rollingUpdate/maxUnavailable\",\"value\":\"25%\"}]'",
			Notes:       "Set to 25% to maintain availability during updates",
		},
		{
			Description: "Verify rollout completes",
			Command:     "kubectl rollout status deployment/frontend --timeout=120s",
			Notes:       "Rollout should complete with pods always available",
		},
	}
}
