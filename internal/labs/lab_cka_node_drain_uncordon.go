package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&NodeDrainUncordonLab{})
}

type NodeDrainUncordonLab struct {
	BaseLab
}

func (l *NodeDrainUncordonLab) ID() string { return "cka_node_drain_uncordon" }
func (l *NodeDrainUncordonLab) Title() string {
	return "Drain and Uncordon Nodes Safely"
}
func (l *NodeDrainUncordonLab) Category() Category     { return CategoryClusterArchitecture }
func (l *NodeDrainUncordonLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NodeDrainUncordonLab) EstimatedTime() int     { return 20 }
func (l *NodeDrainUncordonLab) Tags() []string {
	return []string{"node", "drain", "uncordon", "maintenance"}
}
func (l *NodeDrainUncordonLab) Cert() Cert        { return CertCKA }
func (l *NodeDrainUncordonLab) DomainWeight() int { return 25 }

func (l *NodeDrainUncordonLab) Description() string {
	return `A node needs maintenance but has a pod with a local persistent volume
attached. Drain the node safely by handling the local PV, perform maintenance,
and then uncordon the node.`
}

func (l *NodeDrainUncordonLab) Hints() []string {
	return []string{
		"Check for pods with local persistent volumes",
		"Use --force to override pod eviction failures",
		"Remember to uncordon after maintenance",
	}
}

func (l *NodeDrainUncordonLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeDrainUncordonLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *NodeDrainUncordonLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
		"jsonpath={.items[*].spec.unschedulable}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "true") {
		return fmt.Errorf("node is still cordoned")
	}
	return nil
}

func (l *NodeDrainUncordonLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status", Command: "kubectl get nodes"},
		{Description: "Drain node", Command: "kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data --force"},
		{Description: "Perform maintenance", Command: "sudo apt-get update && sudo apt-get upgrade"},
		{Description: "Uncordon node", Command: "kubectl uncordon <node-name>"},
	}
}
