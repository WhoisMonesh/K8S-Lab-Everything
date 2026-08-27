package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&NetworkPolicyDefaultDenyCKALab{})
}

type NetworkPolicyDefaultDenyCKALab struct {
	BaseLab
}

func (l *NetworkPolicyDefaultDenyCKALab) ID() string {
	return "cka_network_policy_default_deny_v3"
}
func (l *NetworkPolicyDefaultDenyCKALab) Title() string {
	return "Create Default Deny NetworkPolicy"
}
func (l *NetworkPolicyDefaultDenyCKALab) Category() Category {
	return CategoryServicesNetworking
}
func (l *NetworkPolicyDefaultDenyCKALab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NetworkPolicyDefaultDenyCKALab) EstimatedTime() int     { return 15 }
func (l *NetworkPolicyDefaultDenyCKALab) Tags() []string {
	return []string{"network-policy", "default-deny", "security"}
}
func (l *NetworkPolicyDefaultDenyCKALab) Cert() Cert        { return CertCKA }
func (l *NetworkPolicyDefaultDenyCKALab) DomainWeight() int { return 20 }

func (l *NetworkPolicyDefaultDenyCKALab) Description() string {
	return `A namespace has no network policies. Create a default deny NetworkPolicy
that blocks all ingress traffic to pods in the namespace.`
}

func (l *NetworkPolicyDefaultDenyCKALab) Hints() []string {
	return []string{
		"Use an empty podSelector to select all pods",
		"Set policyTypes to Ingress only",
		"Do not specify any ingress rules to deny all",
	}
}

func (l *NetworkPolicyDefaultDenyCKALab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyDefaultDenyCKALab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *NetworkPolicyDefaultDenyCKALab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "deny-ns-3",
		"-o", "name")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "default-deny") {
		return fmt.Errorf("default deny policy not created")
	}
	return nil
}

func (l *NetworkPolicyDefaultDenyCKALab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create default deny policy", Command: "cat <<EOF | kubectl apply -f -\napiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n  name: default-deny\n  namespace: deny-ns-3\nspec:\n  podSelector: {}\n  policyTypes:\n  - Ingress\nEOF"},
		{Description: "Verify", Command: "kubectl get networkpolicies -n deny-ns-3"},
	}
}
