package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DeploymentProgressDeadline{})
}

type DeploymentProgressDeadline struct {
	BaseLab
}

func (l *DeploymentProgressDeadline) ID() string             { return "deployment_progress_deadline" }
func (l *DeploymentProgressDeadline) Title() string          { return "Deployment Progress Deadline Exceeded" }
func (l *DeploymentProgressDeadline) Category() Category     { return CategoryWorkloads }
func (l *DeploymentProgressDeadline) Difficulty() Difficulty { return DifficultyMedium }
func (l *DeploymentProgressDeadline) EstimatedTime() int     { return 20 }
func (l *DeploymentProgressDeadline) Tags() []string {
	return []string{"deployment", "rollout", "workloads"}
}

func (l *DeploymentProgressDeadline) Description() string {
	return `A deployment is stuck progressing because the new pods cannot start.
The progress deadline has been exceeded. Debug and fix the deployment.`
}

func (l *DeploymentProgressDeadline) Hints() []string {
	return []string{
		"Check the deployment status",
		"Look at the new replica set events",
		"Check if pods are failing to start",
	}
}

func (l *DeploymentProgressDeadline) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentProgressDeadline) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 3
  progressDeadlineSeconds: 60
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:broken-tag
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *DeploymentProgressDeadline) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "web",
		"-o", "jsonpath={.status.conditions[?(@.type=='Progressing')].status}")
	if err != nil {
		return err
	}
	if output == "False" {
		return fmt.Errorf("deployment is still Progressing=False")
	}
	return nil
}

func (l *DeploymentProgressDeadline) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check deployment status", Command: "kubectl describe deployment web"},
		{Description: "Check events", Command: "kubectl get events --field-selector involvedObject.name=web"},
		{Description: "Fix image tag", Command: "kubectl set image deployment/web nginx=nginx:stable"},
	}
}
