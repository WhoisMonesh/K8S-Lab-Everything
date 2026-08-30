package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&PausedRolloutResumeLab{}) }

type PausedRolloutResumeLab struct{ BaseLab }

func (l *PausedRolloutResumeLab) ID() string             { return "paused_rollout_resume" }
func (l *PausedRolloutResumeLab) Title() string          { return "Deployment Paused Mid-Rollout" }
func (l *PausedRolloutResumeLab) Category() Category     { return CategoryWorkloads }
func (l *PausedRolloutResumeLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *PausedRolloutResumeLab) EstimatedTime() int     { return 10 }
func (l *PausedRolloutResumeLab) Tags() []string {
	return []string{"deployment", "rollout", "pause", "workloads"}
}
func (l *PausedRolloutResumeLab) Description() string {
	return `A deployment named 'frontend' was being updated to nginx:1.27 but the
rollout appears frozen — new replicas never finish scaling up.

kubectl rollout status shows: Waiting for rollout to finish: 1 out of 2 new
replicas have been updated… indefinitely.

Your task: Resume the paused rollout so all replicas come up.`
}
func (l *PausedRolloutResumeLab) Hints() []string {
	return []string{
		"kubectl rollout status deploy/frontend shows the stall",
		"Check the deployment's spec.paused field",
		"kubectl rollout resume fixes this in one command",
		"Alternatively: kubectl patch deploy frontend -p '{\"spec\":{\"paused\":false}}'",
	}
}

func (l *PausedRolloutResumeLab) Break(ctx context.Context, kp string) error {
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  paused: true
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: web
        image: nginx:1.27-alpine
`
	if err := kubectlApply(ctx, kp, deploy); err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	// Scale old RS to 2, update new RS to 1 to simulate mid-rollout
	if _, err := kubectl(ctx, kp, "scale", "rs", "-l", "app=frontend,app.kubernetes.io/version!=1.27", "--replicas=2", "-n", "default"); err != nil {
		return fmt.Errorf("scaling old RS: %w", err)
	}
	if _, err := kubectl(ctx, kp, "scale", "rs", "-l", "app=frontend,app.kubernetes.io/version=1.27", "--replicas=1", "-n", "default"); err != nil {
		return fmt.Errorf("scaling new RS: %w", err)
	}
	return nil
}

func (l *PausedRolloutResumeLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *PausedRolloutResumeLab) Verify(ctx context.Context, kp string) error {
	paused, _ := kubectl(ctx, kp, "get", "deploy", "frontend", "-o", "jsonpath={.spec.paused}")
	if paused == "true" {
		return fmt.Errorf("deployment is still paused")
	}
	ready, _ := kubectl(ctx, kp, "get", "deploy", "frontend", "-o", "jsonpath={.status.readyReplicas}")
	if ready != "2" {
		return fmt.Errorf("deployment not fully ready (ready: %s)", ready)
	}
	return nil
}

func (l *PausedRolloutResumeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Inspect rollout status", Command: "kubectl rollout status deploy/frontend", Notes: "Shows stalled — waiting for new replicas"},
		{Description: "Check if rollout is paused", Command: "kubectl get deploy frontend -o jsonpath='{.spec.paused}'", Notes: "Returns 'true' — rollout was paused"},
		{Description: "Resume the rollout", Command: "kubectl rollout resume deploy/frontend", Notes: "Sets paused=false; rollout proceeds immediately"},
		{Description: "Verify", Command: "kubectl rollout status deploy/frontend && kubectl get deploy frontend", Notes: "Shows ready 2/2, rollout complete"},
	}
}
