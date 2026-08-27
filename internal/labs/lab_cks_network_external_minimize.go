package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkExternalMinimizeLab{})
}

type CKSNetworkExternalMinimizeLab struct {
	BaseLab
}

func (l *CKSNetworkExternalMinimizeLab) ID() string             { return "cks_network_external_minimize" }
func (l *CKSNetworkExternalMinimizeLab) Title() string          { return "Minimize External Network Access" }
func (l *CKSNetworkExternalMinimizeLab) Category() Category     { return CategorySystemHardening }
func (l *CKSNetworkExternalMinimizeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkExternalMinimizeLab) EstimatedTime() int     { return 20 }
func (l *CKSNetworkExternalMinimizeLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkExternalMinimizeLab) DomainWeight() int      { return 10 }
func (l *CKSNetworkExternalMinimizeLab) Tags() []string {
	return []string{"cks", "network", "external", "egress", "security"}
}

func (l *CKSNetworkExternalMinimizeLab) Description() string {
	return `Pods in the 'internal-app' namespace can access external endpoints freely.
This increases the risk of data exfiltration and supply chain attacks.

Your task: Create a NetworkPolicy that denies all egress traffic by default
in the 'internal-app' namespace, then allow DNS resolution only.`
}

func (l *CKSNetworkExternalMinimizeLab) Hints() []string {
	return []string{
		"Create a default-deny egress policy",
		"Allow UDP port 53 for DNS",
		"Use ipBlock with cidr 0.0.0.0/0 in egress rules",
	}
}

func (l *CKSNetworkExternalMinimizeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkExternalMinimizeLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: internal-app
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: web-fetcher
  namespace: internal-app
spec:
  containers:
  - name: fetcher
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSNetworkExternalMinimizeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "internal-app", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "Egress") {
		return fmt.Errorf("egress policy not found")
	}
	if !strings.Contains(output, "dns") {
		return fmt.Errorf("DNS allow policy not found")
	}
	return nil
}

func (l *CKSNetworkExternalMinimizeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create default-deny egress policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-egress
  namespace: internal-app
spec:
  podSelector: {}
  policyTypes:
  - Egress
EOF`},
		{Description: "Create DNS allow policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns
  namespace: internal-app
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to: []
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
EOF`},
		{Description: "Verify policies", Command: "kubectl get networkpolicies -n internal-app"},
	}
}
