package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DeploymentRollingUpdateStuckLab{})
}

type DeploymentRollingUpdateStuckLab struct {
	BaseLab
}

func (l *DeploymentRollingUpdateStuckLab) ID() string {
	return "deployment_rolling_update_stuck"
}

func (l *DeploymentRollingUpdateStuckLab) Title() string {
	return "Deployment Stuck in Rolling Update"
}

func (l *DeploymentRollingUpdateStuckLab) Category() Category {
	return CategoryWorkloads
}

func (l *DeploymentRollingUpdateStuckLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DeploymentRollingUpdateStuckLab) Description() string {
	return `A deployment 'api-server' has been updated to a new image version, but the rolling update is stuck.
New pods are failing to start while old pods are still running.

Your task: Fix the deployment so the rolling update completes successfully.`
}

func (l *DeploymentRollingUpdateStuckLab) Hints() []string {
	return []string{
		"Check the deployment status",
		"Look at the rollout history",
		"Check the new pod events and logs",
		"The new image might not exist or might have a bad configuration",
	}
}

func (l *DeploymentRollingUpdateStuckLab) EstimatedTime() int {
	return 20
}

func (l *DeploymentRollingUpdateStuckLab) Tags() []string {
	return []string{"deployment", "rolling-update", "rollout", "workloads"}
}

func (l *DeploymentRollingUpdateStuckLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentRollingUpdateStuckLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create initial deployment with version 1
	initialDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: default
  labels:
    app: api-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      labels:
        app: api-server
        version: v1
    spec:
      containers:
      - name: api
        image: nginx:1.25-alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, initialDeployment); err != nil {
		return fmt.Errorf("creating initial deployment: %w", err)
	}

	// Wait for rollout to complete
	_, _ = kubectl(ctx, kubeconfigPath, "rollout", "status", "deployment/api-server", "--timeout=30s")

	// Now update to a broken image (non-existent tag)
	brokenUpdate := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: default
  labels:
    app: api-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      labels:
        app: api-server
        version: v2
    spec:
      containers:
      - name: api
        image: nginx:99.99.99-doesnotexist
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, brokenUpdate); err != nil {
		return fmt.Errorf("applying broken update: %w", err)
	}

	return nil
}

func (l *DeploymentRollingUpdateStuckLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *DeploymentRollingUpdateStuckLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if deployment has all replicas ready
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "api-server",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) != "3" {
		return fmt.Errorf("deployment not fully ready (ready replicas: %s, expected: 3)", output)
	}

	// Check that rollout is complete
	output, err = kubectl(ctx, kubeconfigPath, "rollout", "status", "deployment/api-server", "--timeout=5s")
	if err != nil {
		return fmt.Errorf("rollout not complete: %w", err)
	}

	if !strings.Contains(output, "successfully") {
		return fmt.Errorf("rollout not complete yet")
	}

	return nil
}

func (l *DeploymentRollingUpdateStuckLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment status",
			Command:     "kubectl get deployment api-server",
			Notes:       "Notice only 2 of 3 replicas are ready",
		},
		{
			Description: "Check rollout status",
			Command:     "kubectl rollout status deployment api-server",
			Notes:       "The rollout will be stuck, showing waiting for new pods",
		},
		{
			Description: "Check new pod events",
			Command:     "kubectl get pods -l app=api-server",
			Notes:       "New pods will be in ImagePullBackOff or CrashLoopBackOff",
		},
		{
			Description: "Check rollout history",
			Command:     "kubectl rollout history deployment api-server",
			Notes:       "See the revision history to understand what changed",
		},
		{
			Description: "Identify the broken revision",
			Command:     "kubectl rollout undo deployment api-server --to-revision=1",
			Notes:       "Undo to the last working revision (v1)",
		},
		{
			Description: "Alternative: Fix the image directly",
			Command:     "kubectl set image deployment api-server api=nginx:1.26-alpine",
			Notes:       "Update to a valid nginx image tag",
		},
		{
			Description: "Verify rollout completes",
			Command:     "kubectl rollout status deployment api-server",
			Notes:       "The rollout should now complete successfully",
		},
	}
}
