package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&KubeletConfigLab{})
}

type KubeletConfigLab struct {
	BaseLab
}

func (l *KubeletConfigLab) ID() string             { return "cka_kubelet_config" }
func (l *KubeletConfigLab) Title() string          { return "Fix Kubelet Configuration Issues" }
func (l *KubeletConfigLab) Category() Category     { return CategoryClusterArchitecture }
func (l *KubeletConfigLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *KubeletConfigLab) EstimatedTime() int     { return 20 }
func (l *KubeletConfigLab) Tags() []string {
	return []string{"kubelet", "configuration", "node"}
}
func (l *KubeletConfigLab) Cert() Cert        { return CertCKA }
func (l *KubeletConfigLab) DomainWeight() int { return 25 }

func (l *KubeletConfigLab) Description() string {
	return `A worker node's kubelet is misconfigured with an incorrect --cluster-dns
setting causing DNS resolution failures in pods. Fix the kubelet configuration
to use the correct cluster DNS IP.

kind nodes are containers (no SSH), so access the worker node shell with:
    docker exec -it <cluster>-worker bash
The kubelet config lives at /var/lib/kubelet/config.yaml and kubelet is managed
by the node's kubelet process. Restart it with: systemctl restart kubelet`
}

func (l *KubeletConfigLab) Hints() []string {
	return []string{
		"Enter the worker node: docker exec -it <cluster>-worker bash",
		"Check the kubelet config file at /var/lib/kubelet/config.yaml",
		"Verify the clusterDNS setting",
		"Restart kubelet after fixing",
	}
}

func (l *KubeletConfigLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           1,
	}
}

func (l *KubeletConfigLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeletConfigLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Real scenario: point the worker's kubelet at a wrong cluster DNS server by
	// editing the actual per-node kubelet config, then restart kubelet.
	node, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	// Replace the clusterDNS IP in /var/lib/kubelet/config.yaml with a bogus one.
	// The value can be written inline ([10.96.0.10]) or as a block list
	// (- 10.96.0.10), so replace the IP token itself to match either form.
	cmd := "sed -i 's/10.96.0.10/10.96.0.99/g' /var/lib/kubelet/config.yaml && " +
		"grep -q '10.96.0.99' /var/lib/kubelet/config.yaml && " +
		"systemctl restart kubelet 2>/dev/null || true"
	if _, err := dockerCommand(node, cmd); err != nil {
		return fmt.Errorf("breaking kubelet config on node %s: %w", node, err)
	}
	return nil
}

func (l *KubeletConfigLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// The per-node kubelet config must point at the correct CoreDNS IP again
	// and must no longer reference the broken value.
	node, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	output, err := dockerCommand(node, "cat /var/lib/kubelet/config.yaml")
	if err != nil {
		return fmt.Errorf("could not read kubelet config on %s: %w", node, err)
	}
	if !strings.Contains(output, "10.96.0.10") {
		return fmt.Errorf("kubelet config on %s not using the correct cluster DNS (10.96.0.10)", node)
	}
	if strings.Contains(output, "10.96.0.99") {
		return fmt.Errorf("kubelet config on %s still references the wrong cluster DNS (10.96.0.99)", node)
	}
	return nil
}

func (l *KubeletConfigLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Get the correct CoreDNS cluster IP", Command: "kubectl get svc -n kube-system kube-dns -o jsonpath='{.spec.clusterIP}'"},
		{Description: "Enter the worker node shell (kind has no SSH)", Command: "docker exec -it <cluster>-worker bash"},
		{Description: "Inspect the kubelet config", Command: "grep -A2 clusterDNS /var/lib/kubelet/config.yaml"},
		{Description: "Fix the clusterDNS back to 10.96.0.10", Command: "sed -i 's/10.96.0.99/10.96.0.10/g' /var/lib/kubelet/config.yaml"},
		{Description: "Restart kubelet", Command: "systemctl daemon-reload && systemctl restart kubelet"},
		{Description: "Exit and verify", Command: "exit && kubectl get nodes"},
	}
}
