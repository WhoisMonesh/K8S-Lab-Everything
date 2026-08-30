package labs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func init() {
	Register(&KubeletPodMaxPodsLab{})
}

type KubeletPodMaxPodsLab struct {
	BaseLab
}

func (l *KubeletPodMaxPodsLab) ID() string { return "cka_kubelet_pod_max_pods" }
func (l *KubeletPodMaxPodsLab) Title() string {
	return "Configure kubelet maxPods Setting"
}
func (l *KubeletPodMaxPodsLab) Category() Category     { return CategoryClusterArchitecture }
func (l *KubeletPodMaxPodsLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *KubeletPodMaxPodsLab) EstimatedTime() int     { return 15 }
func (l *KubeletPodMaxPodsLab) Tags() []string {
	return []string{"kubelet", "maxpods", "configuration"}
}
func (l *KubeletPodMaxPodsLab) Cert() Cert        { return CertCKA }
func (l *KubeletPodMaxPodsLab) DomainWeight() int { return 25 }

func (l *KubeletPodMaxPodsLab) Description() string {
	return `The kubelet on a worker node has maxPods set too low (10), preventing
additional pods from being scheduled. Increase maxPods to 110 and restart
the kubelet.

kind nodes are containers (no SSH); access the worker node shell with:
    docker exec -it <cluster>-worker bash
The kubelet config lives at /var/lib/kubelet/config.yaml.`
}

func (l *KubeletPodMaxPodsLab) Hints() []string {
	return []string{
		"Enter the worker node: docker exec -it <cluster>-worker bash",
		"Edit /var/lib/kubelet/config.yaml and look for maxPods",
		"Set maxPods to 110 and restart kubelet",
	}
}

func (l *KubeletPodMaxPodsLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           1,
	}
}

func (l *KubeletPodMaxPodsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeletPodMaxPodsLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Real scenario: clamp maxPods to 10 in the worker's kubelet config and
	// restart kubelet.
	node, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	cmd := "if grep -q '^maxPods:' /var/lib/kubelet/config.yaml; then " +
		"sed -i 's/^maxPods:.*/maxPods: 10/' /var/lib/kubelet/config.yaml; " +
		"else sed -i '1i maxPods: 10' /var/lib/kubelet/config.yaml; fi && " +
		"grep -q '^maxPods: 10' /var/lib/kubelet/config.yaml && " +
		"systemctl restart kubelet 2>/dev/null || true"
	if _, err := dockerCommand(node, cmd); err != nil {
		return fmt.Errorf("breaking maxPods on node %s: %w", node, err)
	}
	return nil
}

func (l *KubeletPodMaxPodsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	node, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	output, err := dockerCommand(node, "grep -E '^maxPods:' /var/lib/kubelet/config.yaml || grep -E 'maxPods:' /var/lib/kubelet/config.yaml")
	if err != nil {
		return fmt.Errorf("could not read maxPods from node %s: %w", node, err)
	}
	valStr := strings.TrimSpace(strings.ReplaceAll(output, "maxPods:", ""))
	if valStr == "" {
		return fmt.Errorf("could not read maxPods from node %s", node)
	}
	val, convErr := strconv.Atoi(valStr)
	if convErr != nil {
		return fmt.Errorf("could not parse maxPods on node %s (value %q)", node, strings.TrimSpace(output))
	}
	if val < 110 {
		return fmt.Errorf("maxPods on node %s is still too low: %d (need >= 110)", node, val)
	}
	return nil
}

func (l *KubeletPodMaxPodsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the node shell (kind has no SSH)", Command: "docker exec -it <cluster>-worker bash"},
		{Description: "Check the current maxPods setting", Command: "grep maxPods /var/lib/kubelet/config.yaml"},
		{Description: "Update maxPods to 110 (only set maxPods if absent)", Command: "sed -i 's/maxPods: *[0-9]*/maxPods: 110/' /var/lib/kubelet/config.yaml; grep -q '^maxPods:' /var/lib/kubelet/config.yaml || sed -i '1i maxPods: 110' /var/lib/kubelet/config.yaml"},
		{Description: "Restart kubelet", Command: "systemctl daemon-reload && systemctl restart kubelet"},
		{Description: "Exit and verify", Command: "exit && kubectl get nodes"},
	}
}
