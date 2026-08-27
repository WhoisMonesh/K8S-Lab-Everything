package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&NetworkPolicyIngressAllowLab{})
}

type NetworkPolicyIngressAllowLab struct {
	BaseLab
}

func (l *NetworkPolicyIngressAllowLab) ID() string {
	return "cka_network_policy_ingress_allow"
}
func (l *NetworkPolicyIngressAllowLab) Title() string {
	return "Allow Specific Ingress Traffic"
}
func (l *NetworkPolicyIngressAllowLab) Category() Category     { return CategoryServicesNetworking }
func (l *NetworkPolicyIngressAllowLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NetworkPolicyIngressAllowLab) EstimatedTime() int     { return 20 }
func (l *NetworkPolicyIngressAllowLab) Tags() []string {
	return []string{"network-policy", "ingress", "networking"}
}
func (l *NetworkPolicyIngressAllowLab) Cert() Cert        { return CertCKA }
func (l *NetworkPolicyIngressAllowLab) DomainWeight() int { return 20 }

func (l *NetworkPolicyIngressAllowLab) Description() string {
	return `A namespace has a default-deny ingress NetworkPolicy. Create an
additional policy that allows ingress traffic from pods with label
role=monitoring to all pods on port 9090 for metrics scraping.`
}

func (l *NetworkPolicyIngressAllowLab) Hints() []string {
	return []string{
		"Create a policy targeting pods in the namespace",
		"Allow ingress from pods with role=monitoring",
		"Specify port 9090 in the policy",
	}
}

func (l *NetworkPolicyIngressAllowLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyIngressAllowLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *NetworkPolicyIngressAllowLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "monitor-ns",
		"-o", "name")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "allow-monitoring") {
		return fmt.Errorf("monitoring ingress policy not created")
	}
	return nil
}

func (l *NetworkPolicyIngressAllowLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check existing policies", Command: "kubectl get networkpolicies -n monitor-ns"},
		{Description: "Create allow policy", Command: "cat <<EOF | kubectl apply -f -\napiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n  name: allow-monitoring\n  namespace: monitor-ns\nspec:\n  podSelector: {}\n  policyTypes:\n  - Ingress\n  ingress:\n  - from:\n    - podSelector:\n        matchLabels:\n          role: monitoring\n    ports:\n    - protocol: TCP\n      port: 9090\nEOF"},
		{Description: "Verify", Command: "kubectl get networkpolicies -n monitor-ns"},
	}
}
