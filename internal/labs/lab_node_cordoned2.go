package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NodeCordoned2{})
}

type NodeCordoned2 struct {
	BaseLab
}

func (l *NodeCordoned2) ID() string             { return "node_cordoned2" }
func (l *NodeCordoned2) Title() string          { return "Node Cordoned - Pods Cannot Schedule" }
func (l *NodeCordoned2) Category() Category     { return CategoryControlPlane }
func (l *NodeCordoned2) Difficulty() Difficulty { return DifficultyEasy }
func (l *NodeCordoned2) EstimatedTime() int     { return 10 }
func (l *NodeCordoned2) Tags() []string         { return []string{"nodes", "cordon", "scheduling"} }

func (l *NodeCordoned2) Description() string {
	return `A node has been cordoned and pods cannot be scheduled to it.
Uncordon the node to allow pod scheduling.`
}

func (l *NodeCordoned2) Hints() []string {
	return []string{
		"Check node status",
		"Look for the Unschedulable condition",
		"Use kubectl uncordon to fix",
	}
}

func (l *NodeCordoned2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeCordoned2) Break(ctx context.Context, kubeconfigPath string) error {
	nodes, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return err
	}
	_, err = kubectl(ctx, kubeconfigPath, "cordon", nodes)
	return err
}

func (l *NodeCordoned2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes",
		"-o", "jsonpath={.items[*].spec.unschedulable}")
	if err != nil {
		return err
	}
	if containsAny(output, "true") {
		return fmt.Errorf("node still cordoned")
	}
	return nil
}

func (l *NodeCordoned2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status", Command: "kubectl get nodes"},
		{Description: "Find cordoned node", Command: "kubectl get nodes | grep SchedulingDisabled"},
		{Description: "Uncordon node", Command: "kubectl uncordon <node-name>"},
	}
}
