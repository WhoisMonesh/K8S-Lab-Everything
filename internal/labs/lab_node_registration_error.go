package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NodeRegistrationError{})
}

type NodeRegistrationError struct {
	BaseLab
}

func (l *NodeRegistrationError) ID() string            { return "node_registration_error" }
func (l *NodeRegistrationError) Title() string         { return "Node Registration Error" }
func (l *NodeRegistrationError) Category() Category    { return CategoryControlPlane }
func (l *NodeRegistrationError) Difficulty() Difficulty { return DifficultyMedium }
func (l *NodeRegistrationError) EstimatedTime() int    { return 20 }
func (l *NodeRegistrationError) Tags() []string        { return []string{"nodes", "registration", "kubelet"} }

func (l *NodeRegistrationError) Description() string {
	return `A node is in NotReady state due to registration errors.
The kubelet cannot register with the API server. Debug and fix the issue.`
}

func (l *NodeRegistrationError) Hints() []string {
	return []string{
		"Check node status and conditions",
		"Look at kubelet logs",
		"Verify kubelet configuration",
	}
}

func (l *NodeRegistrationError) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeRegistrationError) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *NodeRegistrationError) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes",
		"-o", "jsonpath={.items[*].status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return err
	}
	if containsAny(output, "False") {
		return fmt.Errorf("node still NotReady")
	}
	return nil
}

func (l *NodeRegistrationError) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status", Command: "kubectl get nodes"},
		{Description: "Describe node", Command: "kubectl describe node <node-name>"},
		{Description: "Check kubelet logs", Command: "journalctl -u kubelet -f"},
		{Description: "Restart kubelet", Command: "sudo systemctl restart kubelet"},
	}
}
