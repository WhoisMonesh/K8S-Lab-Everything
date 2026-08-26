package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&SlowPodTermination{})
}

type SlowPodTermination struct {
	BaseLab
}

func (l *SlowPodTermination) ID() string             { return "slow_pod_termination2" }
func (l *SlowPodTermination) Title() string          { return "Pod Stuck Terminating" }
func (l *SlowPodTermination) Category() Category     { return CategoryWorkloads }
func (l *SlowPodTermination) Difficulty() Difficulty { return DifficultyMedium }
func (l *SlowPodTermination) EstimatedTime() int     { return 15 }
func (l *SlowPodTermination) Tags() []string         { return []string{"pods", "termination", "graceperiod"} }

func (l *SlowPodTermination) Description() string {
	return `A pod is taking too long to terminate because the terminationGracePeriodSeconds is too high.
Reduce the grace period to speed up pod termination.`
}

func (l *SlowPodTermination) Hints() []string {
	return []string{
		"Check terminationGracePeriodSeconds",
		"Look at the pod spec",
		"Reduce the grace period",
	}
}

func (l *SlowPodTermination) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SlowPodTermination) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: slow-terminate
spec:
  replicas: 1
  selector:
    matchLabels:
      app: slow-terminate
  template:
    metadata:
      labels:
        app: slow-terminate
    spec:
      terminationGracePeriodSeconds: 600
      containers:
      - name: nginx
        image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *SlowPodTermination) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/slow-terminate",
		"-o", "jsonpath={.spec.template.spec.terminationGracePeriodSeconds}")
	if err != nil {
		return err
	}
	if output == "600" {
		return fmt.Errorf("grace period still 600")
	}
	return nil
}

func (l *SlowPodTermination) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check grace period", Command: "kubectl get deploy slow-terminate -o jsonpath='{.spec.template.spec.terminationGracePeriodSeconds}'"},
		{Description: "Fix grace period", Command: "kubectl edit deploy slow-terminate"},
		{Description: "Reduce to 30", Command: "Change terminationGracePeriodSeconds from 600 to 30"},
	}
}
