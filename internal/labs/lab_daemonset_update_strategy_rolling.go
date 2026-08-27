package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DaemonSetUpdateStrategyLab{})
}

type DaemonSetUpdateStrategyLab struct {
	BaseLab
}

func (l *DaemonSetUpdateStrategyLab) ID() string {
	return "daemonset_update_strategy_rolling"
}

func (l *DaemonSetUpdateStrategyLab) Title() string {
	return "DaemonSet Stuck in Rolling Update"
}

func (l *DaemonSetUpdateStrategyLab) Category() Category {
	return CategoryWorkloads
}

func (l *DaemonSetUpdateStrategyLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *DaemonSetUpdateStrategyLab) Description() string {
	return `A DaemonSet 'log-collector' is stuck in a rolling update. The update
strategy is configured with maxUnavailable=0, preventing any nodes from
being updated.

Your task: Fix the DaemonSet update strategy so it can complete the rollout.`
}

func (l *DaemonSetUpdateStrategyLab) Hints() []string {
	return []string{
		"Check the DaemonSet update strategy",
		"maxUnavailable=0 prevents any pods from being updated",
		"Set maxUnavailable to at least 1 or use a percentage",
	}
}

func (l *DaemonSetUpdateStrategyLab) EstimatedTime() int {
	return 15
}

func (l *DaemonSetUpdateStrategyLab) Tags() []string {
	return []string{"daemonset", "rolling-update", "strategy", "workloads"}
}

func (l *DaemonSetUpdateStrategyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DaemonSetUpdateStrategyLab) Break(ctx context.Context, kubeconfigPath string) error {
	daemonset := `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: log-collector
  namespace: default
spec:
  selector:
    matchLabels:
      app: log-collector
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: log-collector
    spec:
      containers:
      - name: collector
        image: fluentd:v1.14
        resources:
          limits:
            memory: 128Mi
          requests:
            memory: 64Mi
`
	if err := kubectlApply(ctx, kubeconfigPath, daemonset); err != nil {
		return fmt.Errorf("creating daemonset: %w", err)
	}

	time.Sleep(10 * time.Second)

	_, _ = kubectl(ctx, kubeconfigPath, "set", "image", "daemonset/log-collector",
		"collector=fluentd:v1.16")

	return nil
}

func (l *DaemonSetUpdateStrategyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *DaemonSetUpdateStrategyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "daemonset", "log-collector",
		"-o", "jsonpath={.status.updatedNumberScheduled}")
	if err != nil {
		return fmt.Errorf("failed to check daemonset: %w", err)
	}

	updated := strings.TrimSpace(output)
	if updated == "" || updated == "0" {
		return fmt.Errorf("daemonset not updated yet")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "daemonset", "log-collector",
		"-o", "jsonpath={.status.numberReady}")
	if err != nil {
		return fmt.Errorf("failed to check ready count: %w", err)
	}

	if strings.TrimSpace(output) == "0" {
		return fmt.Errorf("no ready pods")
	}

	return nil
}

func (l *DaemonSetUpdateStrategyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check DaemonSet status",
			Command:     "kubectl get daemonset log-collector",
			Notes:       "UPDATED column should show 0 or not match DESIRED",
		},
		{
			Description: "Check update strategy",
			Command:     "kubectl get daemonset log-collector -o yaml | grep -A 5 updateStrategy",
			Notes:       "maxUnavailable=0 prevents any updates",
		},
		{
			Description: "Fix the update strategy",
			Command:     "kubectl patch daemonset log-collector --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/updateStrategy/rollingUpdate/maxUnavailable\",\"value\":1}]'",
			Notes:       "Set maxUnavailable to 1 to allow one node to update at a time",
		},
		{
			Description: "Verify rollout completes",
			Command:     "kubectl rollout status daemonset/log-collector --timeout=120s",
			Notes:       "All nodes should now be updated",
		},
	}
}
