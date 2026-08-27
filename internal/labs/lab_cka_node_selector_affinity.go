package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&NodeSelectorAffinityLab{})
}

type NodeSelectorAffinityLab struct {
	BaseLab
}

func (l *NodeSelectorAffinityLab) ID() string { return "cka_node_selector_affinity" }
func (l *NodeSelectorAffinityLab) Title() string {
	return "Schedule Using nodeSelector and nodeAffinity"
}
func (l *NodeSelectorAffinityLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *NodeSelectorAffinityLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NodeSelectorAffinityLab) EstimatedTime() int     { return 20 }
func (l *NodeSelectorAffinityLab) Tags() []string {
	return []string{"nodeselector", "nodeaffinity", "scheduling"}
}
func (l *NodeSelectorAffinityLab) Cert() Cert        { return CertCKA }
func (l *NodeSelectorAffinityLab) DomainWeight() int { return 15 }

func (l *NodeSelectorAffinityLab) Description() string {
	return `A pod needs to be scheduled on nodes with specific labels. Use both
nodeSelector for simple matching and nodeAffinity for more complex rules
to ensure proper placement.`
}

func (l *NodeSelectorAffinityLab) Hints() []string {
	return []string{
		"Add nodeSelector with disktype=ssd",
		"Add nodeAffinity with requiredDuringSchedulingIgnoredDuringExecution",
		"Match zone=east label with In operator",
	}
}

func (l *NodeSelectorAffinityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeSelectorAffinityLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *NodeSelectorAffinityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "selector-pod",
		"-n", "selector-ns", "-o", "jsonpath={.spec.nodeSelector}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "disktype") {
		return fmt.Errorf("nodeSelector not configured correctly")
	}
	return nil
}

func (l *NodeSelectorAffinityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Label nodes", Command: "kubectl label nodes <node> disktype=ssd zone=east"},
		{Description: "Add nodeSelector", Command: "Add nodeSelector:\n  disktype: ssd"},
		{Description: "Add nodeAffinity", Command: "Add nodeAffinity with requiredDuringSchedulingIgnoredDuringExecution for zone=east"},
		{Description: "Verify", Command: "kubectl get pod selector-pod -n selector-ns -o wide"},
	}
}
