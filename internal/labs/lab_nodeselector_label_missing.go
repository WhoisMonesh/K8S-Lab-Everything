package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NodeSelectorLabelMissing{})
}

type NodeSelectorLabelMissing struct {
	BaseLab
}

func (l *NodeSelectorLabelMissing) ID() string             { return "nodeselector_label_missing" }
func (l *NodeSelectorLabelMissing) Title() string          { return "Pod Pending - Missing Node Label" }
func (l *NodeSelectorLabelMissing) Category() Category     { return CategoryScheduling }
func (l *NodeSelectorLabelMissing) Difficulty() Difficulty { return DifficultyEasy }
func (l *NodeSelectorLabelMissing) EstimatedTime() int     { return 10 }
func (l *NodeSelectorLabelMissing) Tags() []string {
	return []string{"scheduling", "nodeselector", "labels"}
}

func (l *NodeSelectorLabelMissing) Description() string {
	return `A pod is stuck in Pending state because the required node label doesn't exist.
Add the missing label to the node.`
}

func (l *NodeSelectorLabelMissing) Hints() []string {
	return []string{
		"Check the pod nodeSelector",
		"List node labels",
		"Add the missing label to a node",
	}
}

func (l *NodeSelectorLabelMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeSelectorLabelMissing) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: selector-pod
spec:
  nodeSelector:
    disktype: ssd
  containers:
  - name: nginx
    image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *NodeSelectorLabelMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "selector-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not running: %s", output)
	}
	return nil
}

func (l *NodeSelectorLabelMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod nodeSelector", Command: "kubectl get pod selector-pod -o jsonpath='{.spec.nodeSelector}'"},
		{Description: "List node labels", Command: "kubectl get nodes --show-labels"},
		{Description: "Add label to node", Command: "kubectl label node <node-name> disktype=ssd"},
	}
}
