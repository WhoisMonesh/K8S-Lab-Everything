package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADNetworkPolicyAllowNSLab{})
}

type CKADNetworkPolicyAllowNSLab struct {
	BaseLab
}

func (l *CKADNetworkPolicyAllowNSLab) ID() string {
	return "ckad_network_policy_allow_ns"
}

func (l *CKADNetworkPolicyAllowNSLab) Title() string {
	return "Allow Traffic Between Namespaces"
}

func (l *CKADNetworkPolicyAllowNSLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADNetworkPolicyAllowNSLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADNetworkPolicyAllowNSLab) Cert() Cert             { return CertCKAD }
func (l *CKADNetworkPolicyAllowNSLab) DomainWeight() int      { return 20 }
func (l *CKADNetworkPolicyAllowNSLab) EstimatedTime() int     { return 25 }
func (l *CKADNetworkPolicyAllowNSLab) Tags() []string {
	return []string{"network-policy", "namespace", "cross-namespace"}
}

func (l *CKADNetworkPolicyAllowNSLab) Description() string {
	return `Two namespaces need to communicate. The frontend namespace should be
able to send traffic to the backend namespace on port 8080.

Your task: Create a NetworkPolicy that allows cross-namespace traffic.`
}

func (l *CKADNetworkPolicyAllowNSLab) Hints() []string {
	return []string{
		"Use namespaceSelector to match source namespace",
		"Use podSelector to match target pods",
		"Apply the policy in the target namespace",
	}
}

func (l *CKADNetworkPolicyAllowNSLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADNetworkPolicyAllowNSLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADNetworkPolicyAllowNSLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "backend",
		"-o", "jsonpath={.items[*].spec.ingress[*].from[*].namespaceSelector.matchLabels}")
	if err != nil {
		return fmt.Errorf("failed to get networkpolicies: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no cross-namespace policy found")
	}
	return nil
}

func (l *CKADNetworkPolicyAllowNSLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create NetworkPolicy", Command: "Create policy in backend namespace with namespaceSelector matching frontend"},
		{Description: "Verify", Command: "kubectl get networkpolicies -n backend"},
	}
}
