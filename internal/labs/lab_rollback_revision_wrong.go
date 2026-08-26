package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&RollbackRevisionWrong{})
}

type RollbackRevisionWrong struct {
	BaseLab
}

func (l *RollbackRevisionWrong) ID() string            { return "rollback_revision_wrong" }
func (l *RollbackRevisionWrong) Title() string         { return "Deployment Rollback to Wrong Revision" }
func (l *RollbackRevisionWrong) Category() Category    { return CategoryWorkloads }
func (l *RollbackRevisionWrong) Difficulty() Difficulty { return DifficultyMedium }
func (l *RollbackRevisionWrong) EstimatedTime() int    { return 15 }
func (l *RollbackRevisionWrong) Tags() []string        { return []string{"deployment", "rollback", "workloads"} }

func (l *RollbackRevisionWrong) Description() string {
	return `A deployment was rolled back to the wrong revision. The current version is broken.
Find the correct revision and rollback to it.`
}

func (l *RollbackRevisionWrong) Hints() []string {
	return []string{
		"Check deployment rollout history",
		"Look at each revision's configuration",
		"Identify which revision was working",
	}
}

func (l *RollbackRevisionWrong) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *RollbackRevisionWrong) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  annotations:
    deployment.kubernetes.io/revision: "3"
spec:
  replicas: 2
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: nginx
        image: nginx:broken
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *RollbackRevisionWrong) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "app",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return err
	}
	if output == "nginx:broken" {
		return fmt.Errorf("deployment still has broken image")
	}
	return nil
}

func (l *RollbackRevisionWrong) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check rollout history", Command: "kubectl rollout history deployment/app"},
		{Description: "Check current image", Command: "kubectl get deployment app -o jsonpath='{.spec.template.spec.containers[0].image}'"},
		{Description: "Fix image", Command: "kubectl set image deployment/app nginx=nginx:stable"},
	}
}
