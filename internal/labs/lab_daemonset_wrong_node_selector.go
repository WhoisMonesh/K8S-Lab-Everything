package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DaemonSetWrongNodeSelectorLab{})
}

type DaemonSetWrongNodeSelectorLab struct {
	BaseLab
}

func (l *DaemonSetWrongNodeSelectorLab) ID() string {
	return "daemonset_wrong_node_selector"
}

func (l *DaemonSetWrongNodeSelectorLab) Title() string {
	return "DaemonSet Not Scheduling Due to Wrong Node Selector"
}

func (l *DaemonSetWrongNodeSelectorLab) Category() Category {
	return CategoryWorkloads
}

func (l *DaemonSetWrongNodeSelectorLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DaemonSetWrongNodeSelectorLab) Description() string {
	return `A DaemonSet 'log-collector' is not scheduling any pods because its nodeSelector doesn't match any nodes.

Your task: Fix the DaemonSet nodeSelector to schedule pods on the nodes.`
}

func (l *DaemonSetWrongNodeSelectorLab) Hints() []string {
	return []string{
		"Check the DaemonSet status",
		"Look at the nodeSelector in the DaemonSet spec",
		"Check node labels",
		"The nodeSelector doesn't match any node labels",
	}
}

func (l *DaemonSetWrongNodeSelectorLab) EstimatedTime() int {
	return 15
}

func (l *DaemonSetWrongNodeSelectorLab) Tags() []string {
	return []string{"daemonset", "nodeselector", "scheduling", "workloads"}
}

func (l *DaemonSetWrongNodeSelectorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DaemonSetWrongNodeSelectorLab) Break(ctx context.Context, kubeconfigPath string) error {
	ds := `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: log-collector
  namespace: default
spec:
  selector:
    matchLabels:
      app: log-collector
  template:
    metadata:
      labels:
        app: log-collector
    spec:
      nodeSelector:
        role: monitoring
      containers:
      - name: collector
        image: busybox:1.28
        command: ['sh', '-c', 'while true; do sleep 3600; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, ds); err != nil {
		return fmt.Errorf("creating DaemonSet: %w", err)
	}
	return nil
}

func (l *DaemonSetWrongNodeSelectorLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *DaemonSetWrongNodeSelectorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ds", "log-collector",
		"-o", "jsonpath={.status.desiredNumberScheduled}")
	if err != nil {
		return fmt.Errorf("failed to check DaemonSet: %w", err)
	}
	if strings.TrimSpace(output) == "0" {
		return fmt.Errorf("DaemonSet has 0 desired scheduled pods")
	}
	output2, err := kubectl(ctx, kubeconfigPath, "get", "ds", "log-collector",
		"-o", "jsonpath={.status.numberReady}")
	if err != nil {
		return fmt.Errorf("failed to check DaemonSet ready: %w", err)
	}
	if strings.TrimSpace(output2) == "0" {
		return fmt.Errorf("DaemonSet has 0 ready pods")
	}
	return nil
}

func (l *DaemonSetWrongNodeSelectorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check DaemonSet status",
			Command:     "kubectl get ds log-collector",
			Notes:       "DESIRED will be 0 because no nodes match",
		},
		{
			Description: "Check node labels",
			Command:     "kubectl get nodes --show-labels",
			Notes:       "No node has role=monitoring label",
		},
		{
			Description: "Fix the nodeSelector",
			Command:     "kubectl edit ds log-collector",
			Notes:       "Change nodeSelector to match available node labels, or remove it entirely",
		},
		{
			Description: "Verify pods are scheduled",
			Command:     "kubectl get ds log-collector",
			Notes:       "DESIRED and READY should be > 0",
		},
	}
}
