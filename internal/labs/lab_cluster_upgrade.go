package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&ClusterUpgradeLab{})
}

type ClusterUpgradeLab struct {
	BaseLab
}

func (l *ClusterUpgradeLab) ID() string {
	return "cluster_upgrade"
}

func (l *ClusterUpgradeLab) Title() string {
	return "Cluster Upgrade Simulation"
}

func (l *ClusterUpgradeLab) Category() Category {
	return CategoryControlPlane
}

func (l *ClusterUpgradeLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *ClusterUpgradeLab) Description() string {
	return `The Kubernetes control plane needs to be upgraded from the current version.
However, the kubelet version on the control plane node is mismatched with the API server version,
causing upgrade issues.

Your task: Identify the version mismatch and simulate a proper upgrade path by ensuring
all components are at compatible versions.`
}

func (l *ClusterUpgradeLab) Hints() []string {
	return []string{
		"Check the versions of all Kubernetes components (kubelet, kube-apiserver, etc.)",
		"Use 'kubectl get nodes' to see the kubelet version",
		"Check the control plane component versions in the manifests",
		"The kubelet should be at most one minor version behind the API server",
	}
}

func (l *ClusterUpgradeLab) EstimatedTime() int {
	return 30
}

func (l *ClusterUpgradeLab) Tags() []string {
	return []string{"upgrade", "version", "control-plane", "kubelet", "troubleshooting"}
}

func (l *ClusterUpgradeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ClusterUpgradeLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Get the control plane node name
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Simulate a version mismatch by modifying the kube-apiserver manifest
	// to report a different version (this is simulation - in real scenarios
	// you'd actually upgrade components)

	// Add a label to the node indicating an upgrade is needed
	_, err = kubectl(ctx, kubeconfigPath, "label", "node", nodeName,
		"upgrade-needed=true", "--overwrite")
	if err != nil {
		return fmt.Errorf("labeling node: %w", err)
	}

	// Create a ConfigMap with version information
	versionInfo := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-version-info
  namespace: kube-system
data:
  target-version: "v1.30.0"
  current-version: "v1.29.0"
  upgrade-status: "pending"
`
	if err := kubectlApply(ctx, kubeconfigPath, versionInfo); err != nil {
		return fmt.Errorf("creating version ConfigMap: %w", err)
	}

	// Modify the kubelet version file to simulate an old version
	_, err = dockerExec(ctx, containerName, "sh", "-c",
		"echo 'v1.29.0' > /var/lib/kubelet/version-marker")
	if err != nil {
		return fmt.Errorf("creating version marker: %w", err)
	}

	return nil
}

func (l *ClusterUpgradeLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Check if the version ConfigMap exists
	_, err := kubectl(ctx, kubeconfigPath, "get", "configmap",
		"cluster-version-info", "-n", "kube-system")
	if err != nil {
		return fmt.Errorf("version ConfigMap not found")
	}
	return nil
}

func (l *ClusterUpgradeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if the upgrade has been marked as complete
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap",
		"cluster-version-info", "-n", "kube-system",
		"-o", "jsonpath={.data.upgrade-status}")
	if err != nil {
		return fmt.Errorf("checking upgrade status: %w", err)
	}

	if !strings.Contains(output, "complete") {
		return fmt.Errorf("upgrade not marked as complete (current status: %s)", output)
	}

	// Check if the upgrade-needed label has been removed
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "node", nodeName,
		"-o", "jsonpath={.metadata.labels.upgrade-needed}")
	if err == nil && output != "" {
		return fmt.Errorf("upgrade-needed label still present on node")
	}

	return nil
}

func (l *ClusterUpgradeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the current cluster version information",
			Command:     "kubectl get configmap cluster-version-info -n kube-system -o yaml",
			Notes:       "This shows the current and target versions for the upgrade",
		},
		{
			Description: "Check the node version and labels",
			Command:     "kubectl get nodes -o wide",
			Notes:       "Look for version information and any upgrade-related labels",
		},
		{
			Description: "Verify component versions",
			Command:     "kubectl version --short",
			Notes:       "Check both client and server versions",
		},
		{
			Description: "Access the control plane node to check kubelet version",
			Command:     "docker exec -it <cluster-name>-control-plane cat /var/lib/kubelet/version-marker",
			Notes:       "Verify the kubelet version marker file",
		},
		{
			Description: "Update the upgrade status to complete",
			Command:     "kubectl patch configmap cluster-version-info -n kube-system --type merge -p '{\"data\":{\"upgrade-status\":\"complete\"}}'",
			Notes:       "Mark the upgrade as complete after verifying versions are compatible",
		},
		{
			Description: "Remove the upgrade-needed label from the node",
			Command:     "kubectl label node <node-name> upgrade-needed-",
			Notes:       "The trailing dash removes the label",
		},
		{
			Description: "Verify the cluster is healthy after the upgrade simulation",
			Command:     "kubectl get pods -A",
			Notes:       "All system pods should be running",
		},
	}
}
