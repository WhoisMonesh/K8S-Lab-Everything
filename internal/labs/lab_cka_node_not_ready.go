package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&NodeNotReadyLab{})
}

type NodeNotReadyLab struct {
	BaseLab
}

func (l *NodeNotReadyLab) ID() string             { return "cka_node_not_ready" }
func (l *NodeNotReadyLab) Title() string          { return "Troubleshoot Node NotReady Status" }
func (l *NodeNotReadyLab) Category() Category     { return CategoryTroubleshooting }
func (l *NodeNotReadyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NodeNotReadyLab) EstimatedTime() int     { return 25 }
func (l *NodeNotReadyLab) Tags() []string {
	return []string{"node", "not-ready", "kubelet", "troubleshooting"}
}
func (l *NodeNotReadyLab) Cert() Cert        { return CertCKA }
func (l *NodeNotReadyLab) DomainWeight() int { return 30 }

func (l *NodeNotReadyLab) Description() string {
	return `A worker node is in NotReady status. Diagnose the issue by checking
the node conditions and fix the underlying problem. The kubelet may have
configuration issues or may have stopped running.`
}

func (l *NodeNotReadyLab) Hints() []string {
	return []string{
		"Check node conditions with kubectl describe node",
		"Look at kubelet logs on the worker node",
		"Verify kubelet is running with systemctl status kubelet",
	}
}

func (l *NodeNotReadyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeNotReadyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *NodeNotReadyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
		"jsonpath={.items[*].status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "False") {
		return fmt.Errorf("node still in NotReady status")
	}
	return nil
}

func (l *NodeNotReadyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status", Command: "kubectl get nodes"},
		{Description: "Describe node", Command: "kubectl describe node <node-name>"},
		{Description: "Check kubelet logs", Command: "ssh <node> journalctl -u kubelet -f"},
		{Description: "Restart kubelet", Command: "ssh <node> systemctl restart kubelet"},
	}
}
