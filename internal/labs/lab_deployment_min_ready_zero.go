package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DeploymentMinReadyZeroLab{})
}

type DeploymentMinReadyZeroLab struct {
	BaseLab
}

func (l *DeploymentMinReadyZeroLab) ID() string {
	return "deployment_min_ready_zero"
}

func (l *DeploymentMinReadyZeroLab) Title() string {
	return "Deployment minReadySeconds Too High"
}

func (l *DeploymentMinReadyZeroLab) Category() Category {
	return CategoryWorkloads
}

func (l *DeploymentMinReadyZeroLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DeploymentMinReadyZeroLab) Description() string {
	return `A deployment 'api-server' has minReadySeconds set to 300 (5 minutes).
This causes the rollout to appear stuck as it waits excessively long
before considering a pod ready.

The application starts and is ready within seconds, but the deployment
takes 5+ minutes to roll out each replica.

Your task: Fix the minReadySeconds to a reasonable value.`
}

func (l *DeploymentMinReadyZeroLab) Hints() []string {
	return []string{
		"Check the deployment configuration for minReadySeconds",
		"minReadySeconds=300 means each pod must be ready for 5 minutes before proceeding",
		"A typical value is 0-10 seconds for most applications",
	}
}

func (l *DeploymentMinReadyZeroLab) EstimatedTime() int {
	return 10
}

func (l *DeploymentMinReadyZeroLab) Tags() []string {
	return []string{"deployment", "minReadySeconds", "rollout", "workloads"}
}

func (l *DeploymentMinReadyZeroLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentMinReadyZeroLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: default
spec:
  replicas: 3
  minReadySeconds: 300
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      containers:
      - name: api
        image: nginx:1.19
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	time.Sleep(10 * time.Second)

	_, _ = kubectl(ctx, kubeconfigPath, "set", "image", "deployment/api-server",
		"api=nginx:1.20")

	return nil
}

func (l *DeploymentMinReadyZeroLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *DeploymentMinReadyZeroLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "api-server",
		"-o", "jsonpath={.spec.minReadySeconds}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "300" {
		return fmt.Errorf("minReadySeconds is still 300")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "deployment", "api-server",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check ready replicas: %w", err)
	}

	if strings.TrimSpace(output) != "3" {
		return fmt.Errorf("deployment not fully ready (ready replicas: %s)", output)
	}

	return nil
}

func (l *DeploymentMinReadyZeroLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment configuration",
			Command:     "kubectl get deployment api-server -o yaml | grep minReadySeconds",
			Notes:       "minReadySeconds is set to 300 (5 minutes)",
		},
		{
			Description: "Fix minReadySeconds",
			Command:     "kubectl patch deployment api-server --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/minReadySeconds\",\"value\":5}]'",
			Notes:       "Set to 5 seconds for typical applications",
		},
		{
			Description: "Verify rollout completes quickly",
			Command:     "kubectl rollout status deployment/api-server --timeout=60s",
			Notes:       "Rollout should now complete in seconds instead of minutes",
		},
	}
}
