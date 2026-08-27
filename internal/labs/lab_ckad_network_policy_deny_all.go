package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADNetworkPolicyDenyAllLab{})
}

type CKADNetworkPolicyDenyAllLab struct {
	BaseLab
}

func (l *CKADNetworkPolicyDenyAllLab) ID() string {
	return "ckad_network_policy_deny_all"
}

func (l *CKADNetworkPolicyDenyAllLab) Title() string {
	return "Create deny-all NetworkPolicy"
}

func (l *CKADNetworkPolicyDenyAllLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADNetworkPolicyDenyAllLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADNetworkPolicyDenyAllLab) Cert() Cert             { return CertCKAD }
func (l *CKADNetworkPolicyDenyAllLab) DomainWeight() int      { return 20 }
func (l *CKADNetworkPolicyDenyAllLab) EstimatedTime() int     { return 15 }
func (l *CKADNetworkPolicyDenyAllLab) Tags() []string {
	return []string{"network-policy", "deny-all", "security"}
}

func (l *CKADNetworkPolicyDenyAllLab) Description() string {
	return `A namespace needs to block all ingress traffic by default. Create a
NetworkPolicy that denies all incoming traffic to pods in the namespace.

Your task: Create a default deny-all ingress NetworkPolicy.`
}

func (l *CKADNetworkPolicyDenyAllLab) Hints() []string {
	return []string{
		"Use empty podSelector: {} to select all pods",
		"Set policyTypes to Ingress only",
		"Don't specify any ingress rules to deny all",
	}
}

func (l *CKADNetworkPolicyDenyAllLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADNetworkPolicyDenyAllLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: restricted`
	return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *CKADNetworkPolicyDenyAllLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "restricted",
		"-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return fmt.Errorf("failed to get networkpolicies: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no NetworkPolicy found")
	}
	if !strings.Contains(output, "deny-all") {
		return fmt.Errorf("deny-all policy not found")
	}
	return nil
}

func (l *CKADNetworkPolicyDenyAllLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create NetworkPolicy", Command: "Create NetworkPolicy with podSelector: {} and policyTypes: [Ingress]"},
		{Description: "Verify", Command: "kubectl get networkpolicies -n restricted"},
	}
}
