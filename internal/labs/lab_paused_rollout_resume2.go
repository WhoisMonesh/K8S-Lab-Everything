package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PausedRolloutResume{})
}

type PausedRolloutResume struct {
	BaseLab
}

func (l *PausedRolloutResume) ID() string            { return "paused_rollout_resume2" }
func (l *PausedRolloutResume) Title() string         { return "Deployment Paused Mid-Rollout" }
func (l *PausedRolloutResume) Category() Category    { return CategoryWorkloads }
func (l *PausedRolloutResume) Difficulty() Difficulty { return DifficultyEasy }
func (l *PausedRolloutResume) EstimatedTime() int    { return 10 }
func (l *PausedRolloutResume) Tags() []string        { return []string{"deployment", "rollout", "paused"} }

func (l *PausedRolloutResume) Description() string {
	return `A deployment rollout is paused. The new version cannot complete.
Resume the rollout to allow it to finish.`
}

func (l *PausedRolloutResume) Hints() []string {
	return []string{
		"Check if the deployment is paused",
		"Look at rollout status",
		"Use kubectl rollout resume",
	}
}

func (l *PausedRolloutResume) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PausedRolloutResume) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: paused-app
spec:
  replicas: 3
  paused: true
  selector:
    matchLabels:
      app: paused-app
  template:
    metadata:
      labels:
        app: paused-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.19
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PausedRolloutResume) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/paused-app",
		"-o", "jsonpath={.spec.paused}")
	if err != nil {
		return err
	}
	if output == "true" {
		return fmt.Errorf("deployment still paused")
	}
	return nil
}

func (l *PausedRolloutResume) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check deployment", Command: "kubectl get deploy paused-app"},
		{Description: "Resume rollout", Command: "kubectl rollout resume deployment paused-app"},
		{Description: "Check status", Command: "kubectl rollout status deployment paused-app"},
	}
}
