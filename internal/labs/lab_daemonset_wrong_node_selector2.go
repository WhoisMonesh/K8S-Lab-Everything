package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DaemonsetWrongNodeSelector2{})
}

type DaemonsetWrongNodeSelector2 struct {
	BaseLab
}

func (l *DaemonsetWrongNodeSelector2) ID() string            { return "daemonset_wrong_node_selector2" }
func (l *DaemonsetWrongNodeSelector2) Title() string         { return "DaemonSet NodeSelector Mismatch" }
func (l *DaemonsetWrongNodeSelector2) Category() Category    { return CategoryWorkloads }
func (l *DaemonsetWrongNodeSelector2) Difficulty() Difficulty { return DifficultyMedium }
func (l *DaemonsetWrongNodeSelector2) EstimatedTime() int    { return 15 }
func (l *DaemonsetWrongNodeSelector2) Tags() []string        { return []string{"daemonset", "nodeselector", "scheduling"} }

func (l *DaemonsetWrongNodeSelector2) Description() string {
	return `A DaemonSet is not scheduling because the nodeSelector doesn't match any nodes.
Fix the nodeSelector to match node labels.`
}

func (l *DaemonsetWrongNodeSelector2) Hints() []string {
	return []string{
		"Check DaemonSet nodeSelector",
		"List node labels",
		"Update nodeSelector to match",
	}
}

func (l *DaemonsetWrongNodeSelector2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DaemonsetWrongNodeSelector2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: logging-ds
spec:
  selector:
    matchLabels:
      app: logging
  template:
    metadata:
      labels:
        app: logging
    spec:
      nodeSelector:
        role: logging
      containers:
      - name: fluentd
        image: fluentd:v1.14
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *DaemonsetWrongNodeSelector2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "daemonset/logging-ds",
		"-o", "jsonpath={.status.desiredNumberScheduled}")
	if err != nil {
		return err
	}
	if output == "0" {
		return fmt.Errorf("daemonset not scheduled on any nodes")
	}
	return nil
}

func (l *DaemonsetWrongNodeSelector2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check nodeSelector", Command: "kubectl get daemonset logging-ds -o jsonpath='{.spec.template.spec.nodeSelector}'"},
		{Description: "List node labels", Command: "kubectl get nodes --show-labels"},
		{Description: "Add label or fix selector", Command: "kubectl label node <node-name> role=logging or kubectl patch daemonset logging-ds -p '{\"spec\":{\"template\":{\"spec\":{\"nodeSelector\":{\"role\":\"worker\"}}}}}'"},
	}
}
