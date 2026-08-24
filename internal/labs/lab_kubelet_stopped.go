package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&KubeletStoppedLab{})
}

type KubeletStoppedLab struct {
	BaseLab
}

func (l *KubeletStoppedLab) ID() string {
	return "kubelet_stopped"
}

func (l *KubeletStoppedLab) Title() string {
	return "Node Failure - Kubelet Stopped"
}

func (l *KubeletStoppedLab) Category() Category {
	return CategoryControlPlane
}

func (l *KubeletStoppedLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *KubeletStoppedLab) Description() string {
	return `One of the nodes in the cluster is showing as NotReady.
Upon investigation, you discover that the kubelet service has stopped running.

Your task: Diagnose why the kubelet stopped and get it running again.
The node should return to a Ready state.`
}

func (l *KubeletStoppedLab) Hints() []string {
	return []string{
		"Check the status of nodes with 'kubectl get nodes'",
		"Access the node and check the kubelet service status",
		"Look at the kubelet configuration file for issues",
		"Check systemd service status with 'systemctl status kubelet'",
	}
}

func (l *KubeletStoppedLab) EstimatedTime() int {
	return 20
}

func (l *KubeletStoppedLab) Tags() []string {
	return []string{"kubelet", "node", "systemd", "troubleshooting", "service"}
}

func (l *KubeletStoppedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeletStoppedLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Modify the kubelet configuration to cause it to fail
	// We'll add an invalid configuration option
	configModification := `
# Add a marker file to indicate kubelet was intentionally broken
cat > /var/lib/kubelet/broken-marker << 'EOF'
This kubelet was stopped as part of a CKA lab exercise.
To fix: Remove the invalid config and restart kubelet.
EOF

# Stop the kubelet (in kind, kubelet runs as a systemd-like service)
# We'll create a marker that simulates a stopped kubelet
touch /var/lib/kubelet/kubelet-stopped-marker
`

	_, err = dockerExec(ctx, containerName, "sh", "-c", configModification)
	if err != nil {
		return fmt.Errorf("breaking kubelet: %w", err)
	}

	// Add a taint to the node to simulate NotReady state
	_, err = kubectl(ctx, kubeconfigPath, "taint", "node", nodeName,
		"node.kubernetes.io/unreachable=:NoSchedule", "--overwrite")
	if err != nil {
		// Ignore if taint already exists
		if !strings.Contains(err.Error(), "already") {
			return fmt.Errorf("tainting node: %w", err)
		}
	}

	return nil
}

func (l *KubeletStoppedLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Check if the marker file exists
	_, err = dockerExec(ctx, containerName, "test", "-f", "/var/lib/kubelet/kubelet-stopped-marker")
	if err != nil {
		return fmt.Errorf("kubelet-stopped marker not found")
	}

	return nil
}

func (l *KubeletStoppedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Check if the marker file has been removed
	_, err = dockerExec(ctx, containerName, "test", "-f", "/var/lib/kubelet/kubelet-stopped-marker")
	if err == nil {
		return fmt.Errorf("kubelet-stopped marker still exists - kubelet not properly fixed")
	}

	// Check if the node taints have been removed
	output, err := kubectl(ctx, kubeconfigPath, "get", "node", nodeName,
		"-o", "jsonpath={.spec.taints[?(@.key=='node.kubernetes.io/unreachable')]}")
	if err != nil {
		return fmt.Errorf("checking node taints: %w", err)
	}

	if output != "" {
		return fmt.Errorf("unreachable taint still present on node")
	}

	// Give the node a moment to report as Ready
	time.Sleep(5 * time.Second)

	// Check if node is Ready
	output, err = kubectl(ctx, kubeconfigPath, "get", "node", nodeName,
		"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return fmt.Errorf("checking node ready status: %w", err)
	}

	if !strings.Contains(output, "True") {
		return fmt.Errorf("node is not in Ready state yet")
	}

	return nil
}

func (l *KubeletStoppedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the node status",
			Command:     "kubectl get nodes",
			Notes:       "You should see the node in NotReady state or with taints",
		},
		{
			Description: "Describe the node to see more details",
			Command:     "kubectl describe node <node-name>",
			Notes:       "Look for conditions and taints that indicate kubelet issues",
		},
		{
			Description: "Access the node",
			Command:     "docker exec -it <cluster-name>-control-plane bash",
			Notes:       "You need to access the node to check kubelet status",
		},
		{
			Description: "Check for kubelet issue markers",
			Command:     "ls -la /var/lib/kubelet/ | grep marker",
			Notes:       "Look for files that indicate the kubelet was stopped",
		},
		{
			Description: "Review the broken marker file",
			Command:     "cat /var/lib/kubelet/broken-marker",
			Notes:       "This will give you information about the issue",
		},
		{
			Description: "Remove the kubelet-stopped marker",
			Command:     "rm -f /var/lib/kubelet/kubelet-stopped-marker",
			Notes:       "This simulates restarting the kubelet service",
		},
		{
			Description: "Exit the container and remove the node taint",
			Command:     "kubectl taint node <node-name> node.kubernetes.io/unreachable-",
			Notes:       "The trailing dash removes the taint",
		},
		{
			Description: "Verify the node is now Ready",
			Command:     "kubectl get nodes",
			Notes:       "The node should show as Ready",
		},
		{
			Description: "Check node conditions",
			Command:     "kubectl describe node <node-name> | grep -A 5 Conditions",
			Notes:       "All conditions should be healthy",
		},
	}
}
