package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NodeNotReadyLab{})
}

type NodeNotReadyLab struct {
	BaseLab
}

func (l *NodeNotReadyLab) ID() string {
	return "node_not_ready"
}

func (l *NodeNotReadyLab) Title() string {
	return "Node Not Ready"
}

func (l *NodeNotReadyLab) Category() Category {
	return CategoryControlPlane
}

func (l *NodeNotReadyLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NodeNotReadyLab) Description() string {
	return `A node in the cluster is in 'NotReady' state. Pods scheduled on this node cannot start.
The kubelet on this node has stopped working.

Your task: Fix the kubelet configuration and restart it to bring the node back to Ready state.`
}

func (l *NodeNotReadyLab) Hints() []string {
	return []string{
		"Check node status with kubectl get nodes",
		"Describe the node to see the conditions and events",
		"Check kubelet logs on the node",
		"The kubelet configuration file might be corrupted",
	}
}

func (l *NodeNotReadyLab) EstimatedTime() int {
	return 20
}

func (l *NodeNotReadyLab) Tags() []string {
	return []string{"kubelet", "node", "control-plane", "troubleshooting"}
}

func (l *NodeNotReadyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeNotReadyLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}

	// Backup and corrupt the kubelet config
	_, _ = dockerExec(ctx, nodeName, "cp", "/var/lib/kubelet/config.yaml", "/var/lib/kubelet/config.yaml.bak")

	// Corrupt the kubelet config by adding invalid YAML
	_, err = dockerExec(ctx, nodeName, "sh", "-c",
		`echo "invalid: [yaml: {broken" > /var/lib/kubelet/config.yaml && echo "yaml corruption done"`)
	if err != nil {
		return fmt.Errorf("corrupting kubelet config: %w", err)
	}

	// Restart kubelet to make it fail
	_, err = dockerExec(ctx, nodeName, "systemctl", "restart", "kubelet")
	if err != nil {
		return fmt.Errorf("restarting kubelet: %w", err)
	}

	return nil
}

func (l *NodeNotReadyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)
	return nil
}

func (l *NodeNotReadyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if node is Ready
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes",
		"-o", "jsonpath={.items[*].status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return fmt.Errorf("failed to check nodes: %w", err)
	}

	if !strings.Contains(output, "True") {
		return fmt.Errorf("no nodes are Ready (status: %s)", output)
	}

	return nil
}

func (l *NodeNotReadyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check node status",
			Command:     "kubectl get nodes",
			Notes:       "One node will show as NotReady",
		},
		{
			Description: "Describe the NotReady node",
			Command:     "kubectl describe node <node-name>",
			Notes:       "Look at the Conditions section and Events",
		},
		{
			Description: "SSH into the node or use docker exec for kind",
			Command:     "docker exec -it <node-name> bash",
			Notes:       "For kind clusters, use docker exec to access the node",
		},
		{
			Description: "Check kubelet logs",
			Command:     "journalctl -u kubelet -n 50",
			Notes:       "Look for configuration parsing errors",
		},
		{
			Description: "Restore the kubelet config",
			Command:     "cp /var/lib/kubelet/config.yaml.bak /var/lib/kubelet/config.yaml",
			Notes:       "Restore the backup of the kubelet configuration",
		},
		{
			Description: "Restart kubelet",
			Command:     "systemctl restart kubelet",
			Notes:       "Wait for kubelet to start and node to become Ready",
		},
		{
			Description: "Verify node is Ready",
			Command:     "kubectl get nodes",
			Notes:       "The node should now show as Ready",
		},
	}
}
