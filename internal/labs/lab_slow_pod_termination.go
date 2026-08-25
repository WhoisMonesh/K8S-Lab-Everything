package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&SlowPodTerminationLab{}) }

type SlowPodTerminationLab struct{ BaseLab }

func (l *SlowPodTerminationLab) ID() string          { return "slow_pod_termination" }
func (l *SlowPodTerminationLab) Title() string        { return "Pod Stuck Terminating" }
func (l *SlowPodTerminationLab) Category() Category   { return CategoryWorkloads }
func (l *SlowPodTerminationLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *SlowPodTerminationLab) EstimatedTime() int   { return 15 }
func (l *SlowPodTerminationLab) Tags() []string {
	return []string{"termination", "grace-period", "deletion", "workloads"}
}
func (l *SlowPodTerminationLab) Description() string {
	return `A deployment named 'legacy-worker' has a pod stuck in Terminating for
minutes after deletion. The terminationGracePeriodSeconds is set to 3600
(1 hour) and the app doesn't handle SIGTERM gracefully.

Your task: Fix the pod spec so that when the pod is deleted, it terminates
quickly (under 30 seconds). Then delete the stuck pod.`
}
func (l *SlowPodTerminationLab) Hints() []string {
	return []string{
		"Check spec.terminationGracePeriodSeconds on the deployment",
		"1 hour = 3600s is far too generous",
		"Patch it down to a sane value like 10",
		"Then delete the stuck pod so a replacement with the new spec starts",
	}
}

func (l *SlowPodTerminationLab) Break(ctx context.Context, kp string) error {
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: legacy-worker
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: legacy-worker
  template:
    metadata:
      labels:
        app: legacy-worker
    spec:
      terminationGracePeriodSeconds: 3600
      containers:
      - name: worker
        image: busybox:1.36
        command: ["sh","-c","trap 'echo caught; sleep 9999' TERM; while true; do sleep 1; done"]
`
	return kubectlApply(ctx, kp, deploy)
}

func (l *SlowPodTerminationLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *SlowPodTerminationLab) Verify(ctx context.Context, kp string) error {
	grace, _ := kubectl(ctx, kp, "get", "deploy", "legacy-worker", "-o",
		"jsonpath={.spec.template.spec.terminationGracePeriodSeconds}")
	if grace == "3600" || grace == "" {
		return fmt.Errorf("terminationGracePeriodSeconds is still 3600")
	}
	pods, _ := kubectl(ctx, kp, "get", "pods", "-l", "app=legacy-worker", "-o",
		"jsonpath={.items[*].status.phase}")
	for _, p := range splitFields(pods) {
		if p != "Running" {
			return fmt.Errorf("pod not running yet (phase: %s)", p)
		}
	}
	return nil
}

func (l *SlowPodTerminationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check grace period", Command: "kubectl get deploy legacy-worker -o jsonpath='{.spec.template.spec.terminationGracePeriodSeconds}'", Notes: "Shows 3600 — 1 hour!"},
		{Description: "Patch to 10 seconds", Command: `kubectl patch deploy legacy-worker -p '{"spec":{"template":{"spec":{"terminationGracePeriodSeconds":10}}}}'`, Notes: "New pods will use 10s grace"},
		{Description: "Delete the stuck pod", Command: "kubectl delete pod -l app=legacy-worker --grace-period=10", Notes: "New replacement pod starts with 10s grace"},
		{Description: "Verify", Command: "kubectl get pods -l app=legacy-worker", Notes: "New pod Running, no stuck Terminating pods"},
	}
}
