package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyEgressRestrictLab{})
}

type CKSNetworkPolicyEgressRestrictLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyEgressRestrictLab) ID() string             { return "cks_network_policy_egress_restrict" }
func (l *CKSNetworkPolicyEgressRestrictLab) Title() string          { return "Restrict Egress Traffic" }
func (l *CKSNetworkPolicyEgressRestrictLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSNetworkPolicyEgressRestrictLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkPolicyEgressRestrictLab) EstimatedTime() int     { return 20 }
func (l *CKSNetworkPolicyEgressRestrictLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyEgressRestrictLab) DomainWeight() int      { return 20 }
func (l *CKSNetworkPolicyEgressRestrictLab) Tags() []string {
	return []string{"cks", "network-policy", "egress", "microservice-vulns"}
}

func (l *CKSNetworkPolicyEgressRestrictLab) Description() string {
	return `The 'restricted-app' namespace allows unrestricted egress traffic. Pods can
connect to any external endpoint, which could lead to data exfiltration.

Your task: Create a NetworkPolicy that:
1. Denies all egress by default
2. Allows egress to internal pods in the same namespace
3. Allows egress to DNS (port 53)`
}

func (l *CKSNetworkPolicyEgressRestrictLab) Hints() []string {
	return []string{
		"Use policyTypes: [Egress] with no rules for default deny",
		"Add a separate policy for DNS egress",
		"Add a policy for same-namespace egress",
	}
}

func (l *CKSNetworkPolicyEgressRestrictLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyEgressRestrictLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: restricted-app
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: app-pod
  namespace: restricted-app
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSNetworkPolicyEgressRestrictLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "restricted-app", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "default-deny-egress") {
		return fmt.Errorf("default-deny-egress policy not found")
	}
	if !strings.Contains(output, "allow-dns") {
		return fmt.Errorf("allow-dns policy not found")
	}
	return nil
}

func (l *CKSNetworkPolicyEgressRestrictLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create default-deny egress", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-egress
  namespace: restricted-app
spec:
  podSelector: {}
  policyTypes:
  - Egress
EOF`},
		{Description: "Allow DNS egress", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns
  namespace: restricted-app
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
EOF`},
		{Description: "Verify", Command: "kubectl get networkpolicies -n restricted-app"},
	}
}
