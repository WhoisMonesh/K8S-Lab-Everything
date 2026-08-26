package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&BadImageUndo{})
}

type BadImageUndo struct {
	BaseLab
}

func (l *BadImageUndo) ID() string             { return "bad_image_undo2" }
func (l *BadImageUndo) Title() string          { return "Roll Back a Bad Image Update" }
func (l *BadImageUndo) Category() Category     { return CategoryWorkloads }
func (l *BadImageUndo) Difficulty() Difficulty { return DifficultyMedium }
func (l *BadImageUndo) EstimatedTime() int     { return 15 }
func (l *BadImageUndo) Tags() []string         { return []string{"deployment", "rollback", "image"} }

func (l *BadImageUndo) Description() string {
	return `A deployment was updated with a broken image tag.
Roll back to the previous working revision.`
}

func (l *BadImageUndo) Hints() []string {
	return []string{
		"Check deployment rollout history",
		"Look at the current and previous images",
		"Use kubectl rollout undo",
	}
}

func (l *BadImageUndo) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *BadImageUndo) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: rolling-app
  annotations:
    deployment.kubernetes.io/revision: "2"
spec:
  replicas: 2
  selector:
    matchLabels:
      app: rolling-app
  template:
    metadata:
      labels:
        app: rolling-app
    spec:
      containers:
      - name: nginx
        image: nginx:broken-tag
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *BadImageUndo) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/rolling-app",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return err
	}
	if output == "nginx:broken-tag" {
		return fmt.Errorf("deployment still has broken image")
	}
	return nil
}

func (l *BadImageUndo) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check rollout history", Command: "kubectl rollout history deployment/rolling-app"},
		{Description: "Undo rollback", Command: "kubectl rollout undo deployment/rolling-app"},
		{Description: "Check status", Command: "kubectl rollout status deployment/rolling-app"},
	}
}
