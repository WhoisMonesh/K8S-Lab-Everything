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
	return `A worker node has been cordoned (marked unschedulable) for maintenance,
which prevents new pods from scheduling on it. Drain any workload from the
node safely, verify the maintenance is done, and then uncordon the node so it
resumes scheduling again.

kind nodes are containers (no SSH); run node-level commands inside the node
shell with:  docker exec -it <cluster>-worker bash`
}

func (l *NodeDrainUncordonLab) Hints() []string {
	return []string{
		"Check which node is unschedulable (SchedulingDisabled)",
		"Use kubectl drain <node> --ignore-daemonsets --delete-emptydir-data --force",
		"Remember to uncordon: kubectl uncordon <node>",
	}
}

func (l *NodeDrainUncordonLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           1,
	}
}

func (l *NodeDrainUncordonLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeDrainUncordonLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Real scenario: cordon a worker node so the learner must uncordon it.
	node, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	if _, err := kubectl(ctx, kubeconfigPath, "cordon", node); err != nil {
		return fmt.Errorf("cordoning node %s: %w", node, err)
	}
	return nil
}

func (l *NodeDrainUncordonLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Pass once no node is unschedulable (the learner uncordoned the node).
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
		"jsonpath={.items[*].spec.unschedulable}")
	if err != nil {
		return err
	}
	if strings.Contains(strings.ToLower(output), "true") {
		return fmt.Errorf("node is still cordoned (unschedulable)")
	}
	return nil
}

func (l *NodeDrainUncordonLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status (one is SchedulingDisabled)", Command: "kubectl get nodes"},
		{Description: "Drain the cordoned node safely", Command: "kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data --force"},
		{Description: "Perform the node maintenance", Command: "docker exec -it <cluster>-worker bash  (then run systemctl/apt inside)"},
		{Description: "Uncordon the node to resume scheduling", Command: "kubectl uncordon <node-name>"},
		{Description: "Verify the node is schedulable again", Command: "kubectl get nodes"},
	}
}
