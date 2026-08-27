package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADNetworkPolicyAllowPodLab{})
}

type CKADNetworkPolicyAllowPodLab struct {
	BaseLab
}

func (l *CKADNetworkPolicyAllowPodLab) ID() string {
	return "ckad_network_policy_allow_pod"
}

func (l *CKADNetworkPolicyAllowPodLab) Title() string {
	return "Allow Traffic to Specific Pods"
}

func (l *CKADNetworkPolicyAllowPodLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADNetworkPolicyAllowPodLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADNetworkPolicyAllowPodLab) Cert() Cert             { return CertCKAD }
func (l *CKADNetworkPolicyAllowPodLab) DomainWeight() int      { return 20 }
func (l *CKADNetworkPolicyAllowPodLab) EstimatedTime() int     { return 25 }
func (l *CKADNetworkPolicyAllowPodLab) Tags() []string {
	return []string{"network-policy", "pod-selector", "ingress"}
}

func (l *CKADNetworkPolicyAllowPodLab) Description() string {
	return `Pods with label role=frontend should be able to communicate with pods
labeled role=backend on port 8080. All other traffic should be blocked.

Your task: Create a NetworkPolicy allowing frontend-to-backend traffic.`
}

func (l *CKADNetworkPolicyAllowPodLab) Hints() []string {
	return []string{
		"Use podSelector to match target pods (backend)",
		"Use from with podSelector to match source pods (frontend)",
		"Specify the port in the ingress rules",
	}
}

func (l *CKADNetworkPolicyAllowPodLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADNetworkPolicyAllowPodLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADNetworkPolicyAllowPodLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies",
		"-o", "jsonpath={.items[*].spec.ingress[*].from[*].podSelector.matchLabels.role}")
	if err != nil {
		return fmt.Errorf("failed to get networkpolicies: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no pod-based policy found")
	}
	if !strings.Contains(output, "frontend") {
		return fmt.Errorf("frontend selector not found")
	}
	return nil
}

func (l *CKADNetworkPolicyAllowPodLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create NetworkPolicy", Command: "Create policy with podSelector for backend and ingress from pods with role=frontend"},
		{Description: "Verify", Command: "kubectl get networkpolicies -o yaml"},
	}
}
