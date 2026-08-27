package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyDefaultDenyPodLab{})
}

type CKSNetworkPolicyDefaultDenyPodLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyDefaultDenyPodLab) ID() string             { return "cks_network_policy_default_deny" }
func (l *CKSNetworkPolicyDefaultDenyPodLab) Title() string          { return "Default Deny All Pod Traffic" }
func (l *CKSNetworkPolicyDefaultDenyPodLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSNetworkPolicyDefaultDenyPodLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkPolicyDefaultDenyPodLab) EstimatedTime() int     { return 20 }
func (l *CKSNetworkPolicyDefaultDenyPodLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyDefaultDenyPodLab) DomainWeight() int      { return 20 }
func (l *CKSNetworkPolicyDefaultDenyPodLab) Tags() []string {
	return []string{"cks", "network-policy", "default-deny", "microservice-vulns"}
}

func (l *CKSNetworkPolicyDefaultDenyPodLab) Description() string {
	return `The 'microservices' namespace has multiple pods but no default-deny NetworkPolicy.
Pods can communicate freely, increasing the blast radius of any compromise.

Your task: Create a default-deny NetworkPolicy that blocks all ingress and egress
traffic in the 'microservices' namespace.`
}

func (l *CKSNetworkPolicyDefaultDenyPodLab) Hints() []string {
	return []string{
		"Use empty podSelector for all pods",
		"Set both Ingress and Egress in policyTypes",
		"Do not define ingress/egress rules to deny all",
	}
}

func (l *CKSNetworkPolicyDefaultDenyPodLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyDefaultDenyPodLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: microservices
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	for _, name := range []string{"frontend", "backend", "database"} {
		pod := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: microservices
  labels:
    app: %s
spec:
  containers:
  - name: %s
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`, name, name, name)
		if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
			return fmt.Errorf("creating pod %s: %w", name, err)
		}
	}
	return nil
}

func (l *CKSNetworkPolicyDefaultDenyPodLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "microservices", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "Ingress") || !strings.Contains(output, "Egress") {
		return fmt.Errorf("default-deny policy not created")
	}
	return nil
}

func (l *CKSNetworkPolicyDefaultDenyPodLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create default-deny policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: microservices
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
EOF`},
		{Description: "Verify", Command: "kubectl get networkpolicies -n microservices"},
	}
}
