package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DeploymentMaxSurgeZeroLab{})
}

type DeploymentMaxSurgeZeroLab struct {
	BaseLab
}

func (l *DeploymentMaxSurgeZeroLab) ID() string {
	return "deployment_max_surge_zero"
}

func (l *DeploymentMaxSurgeZeroLab) Title() string {
	return "Deployment Stuck with maxSurge=0"
}

func (l *DeploymentMaxSurgeZeroLab) Category() Category {
	return CategoryWorkloads
}

func (l *DeploymentMaxSurgeZeroLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DeploymentMaxSurgeZeroLab) Description() string {
	return `A deployment 'web-app' has maxSurge=0 and maxUnavailable=1, making
it impossible to perform a rolling update. New pods cannot be created
before old ones are terminated, but with maxUnavailable=1 only one pod
can be unavailable at a time.

Your task: Fix the deployment strategy to allow proper rolling updates.`
}

func (l *DeploymentMaxSurgeZeroLab) Hints() []string {
	return []string{
		"Check the deployment strategy configuration",
		"maxSurge=0 with maxUnavailable=1 prevents new pods from being created first",
		"Increase maxSurge to at least 1 to allow new pods before terminating old ones",
	}
}

func (l *DeploymentMaxSurgeZeroLab) EstimatedTime() int {
	return 15
}

func (l *DeploymentMaxSurgeZeroLab) Tags() []string {
	return []string{"deployment", "rolling-update", "strategy", "workloads"}
}

func (l *DeploymentMaxSurgeZeroLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentMaxSurgeZeroLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web-app
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 0
      maxUnavailable: 1
  template:
    metadata:
      labels:
        app: web-app
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

	// Wait for rollout to complete
	time.Sleep(10 * time.Second)

	// Trigger an update that will get stuck
	_, _ = kubectl(ctx, kubeconfigPath, "set", "image", "deployment/web-app",
		"web=nginx:1.20")

	return nil
}

func (l *DeploymentMaxSurgeZeroLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "deployment", "web-app",
		"-o", "jsonpath={.status.observedGeneration}")
	_ = output
	return nil
}

func (l *DeploymentMaxSurgeZeroLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "web-app",
		"-o", "jsonpath={.spec.strategy.rollingUpdate.maxSurge}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) == "0" {
		return fmt.Errorf("maxSurge is still 0")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "deployment", "web-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check ready replicas: %w", err)
	}

	if strings.TrimSpace(output) != "3" {
		return fmt.Errorf("deployment not fully ready (ready replicas: %s)", output)
	}

	return nil
}

func (l *DeploymentMaxSurgeZeroLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment rollout status",
			Command:     "kubectl rollout status deployment/web-app",
			Notes:       "Rollout should be stuck or failing",
		},
		{
			Description: "Check deployment strategy",
			Command:     "kubectl get deployment web-app -o yaml | grep -A 5 strategy",
			Notes:       "maxSurge=0 with maxUnavailable=1 is problematic",
		},
		{
			Description: "Fix the strategy configuration",
			Command:     "kubectl patch deployment web-app --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/strategy/rollingUpdate/maxSurge\",\"value\":1}]'",
			Notes:       "Set maxSurge to at least 1 to allow new pods to be created",
		},
		{
			Description: "Verify rollout completes",
			Command:     "kubectl rollout status deployment/web-app",
			Notes:       "Rollout should now complete successfully",
		},
	}
}
