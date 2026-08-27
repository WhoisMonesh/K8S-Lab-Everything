package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADNetworkPolicyEgressLab{})
}

type CKADNetworkPolicyEgressLab struct {
	BaseLab
}

func (l *CKADNetworkPolicyEgressLab) ID() string {
	return "ckad_network_policy_egress"
}

func (l *CKADNetworkPolicyEgressLab) Title() string {
	return "Configure Egress NetworkPolicy"
}

func (l *CKADNetworkPolicyEgressLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADNetworkPolicyEgressLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADNetworkPolicyEgressLab) Cert() Cert             { return CertCKAD }
func (l *CKADNetworkPolicyEgressLab) DomainWeight() int      { return 20 }
func (l *CKADNetworkPolicyEgressLab) EstimatedTime() int     { return 25 }
func (l *CKADNetworkPolicyEgressLab) Tags() []string {
	return []string{"network-policy", "egress", "outbound"}
}

func (l *CKADNetworkPolicyEgressLab) Description() string {
	return `A pod should only be able to make DNS queries and connect to a specific
database service. Configure an egress NetworkPolicy to restrict outbound traffic.

Your task: Create an egress policy allowing DNS and database access only.`
}

func (l *CKADNetworkPolicyEgressLab) Hints() []string {
	return []string{
		"Use policyTypes: [Egress] to control outbound traffic",
		"Allow DNS traffic on port 53 for kube-dns",
		"Allow traffic to the database service on its port",
	}
}

func (l *CKADNetworkPolicyEgressLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADNetworkPolicyEgressLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADNetworkPolicyEgressLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies",
		"-o", "jsonpath={.items[*].spec.policyTypes[*]}")
	if err != nil {
		return fmt.Errorf("failed to get networkpolicies: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no NetworkPolicy found")
	}
	if !strings.Contains(output, "Egress") {
		return fmt.Errorf("Egress policyType not configured")
	}
	return nil
}

func (l *CKADNetworkPolicyEgressLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create NetworkPolicy", Command: "Create policy with egress rules for DNS (port 53) and database service"},
		{Description: "Verify", Command: "kubectl get networkpolicies -o yaml | grep Egress"},
	}
}
