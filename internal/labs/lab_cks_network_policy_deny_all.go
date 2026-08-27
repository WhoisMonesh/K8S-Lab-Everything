package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyDenyAllLab{})
}

type CKSNetworkPolicyDenyAllLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyDenyAllLab) ID() string             { return "cks_network_policy_deny_all" }
func (l *CKSNetworkPolicyDenyAllLab) Title() string          { return "Default Deny All Cluster Traffic" }
func (l *CKSNetworkPolicyDenyAllLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSNetworkPolicyDenyAllLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKSNetworkPolicyDenyAllLab) EstimatedTime() int     { return 15 }
func (l *CKSNetworkPolicyDenyAllLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyDenyAllLab) DomainWeight() int      { return 15 }
func (l *CKSNetworkPolicyDenyAllLab) Tags() []string {
	return []string{"cks", "network-policy", "default-deny", "security"}
}

func (l *CKSNetworkPolicyDenyAllLab) Description() string {
	return `The 'restricted-ns' namespace has pods running but no default-deny NetworkPolicy.
Pods can freely communicate with each other and external endpoints.

Your task: Create a default-deny NetworkPolicy that blocks all ingress and egress
traffic in the 'restricted-ns' namespace.`
}

func (l *CKSNetworkPolicyDenyAllLab) Hints() []string {
	return []string{
		"Use an empty podSelector to apply to all pods",
		"Set policyTypes to both Ingress and Egress",
		"Do not define any ingress/egress rules to deny all",
	}
}

func (l *CKSNetworkPolicyDenyAllLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyDenyAllLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: restricted-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  namespace: restricted-ns
spec:
  replicas: 2
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKSNetworkPolicyDenyAllLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "restricted-ns", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "policyTypes") {
		return fmt.Errorf("network policy with policyTypes not found")
	}
	if !strings.Contains(output, "Ingress") || !strings.Contains(output, "Egress") {
		return fmt.Errorf("policy must include both Ingress and Egress types")
	}
	return nil
}

func (l *CKSNetworkPolicyDenyAllLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create default-deny NetworkPolicy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: restricted-ns
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
EOF`},
		{Description: "Verify policy", Command: "kubectl get networkpolicies -n restricted-ns"},
	}
}
