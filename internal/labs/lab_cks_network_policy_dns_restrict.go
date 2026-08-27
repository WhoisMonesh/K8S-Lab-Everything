package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyDNSRestrictLab{})
}

type CKSNetworkPolicyDNSRestrictLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyDNSRestrictLab) ID() string             { return "cks_network_policy_dns_restrict" }
func (l *CKSNetworkPolicyDNSRestrictLab) Title() string          { return "Restrict DNS Traffic" }
func (l *CKSNetworkPolicyDNSRestrictLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSNetworkPolicyDNSRestrictLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkPolicyDNSRestrictLab) EstimatedTime() int     { return 20 }
func (l *CKSNetworkPolicyDNSRestrictLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyDNSRestrictLab) DomainWeight() int      { return 20 }
func (l *CKSNetworkPolicyDNSRestrictLab) Tags() []string {
	return []string{"cks", "network-policy", "dns", "microservice-vulns"}
}

func (l *CKSNetworkPolicyDNSRestrictLab) Description() string {
	return `The 'dns-controlled' namespace allows DNS resolution to all domains. Pods
can resolve any hostname, potentially accessing unauthorized services.

Your task: Create a NetworkPolicy that allows DNS resolution only to
'kube-dns.kube-system.svc.cluster.local' and blocks DNS to other domains.`
}

func (l *CKSNetworkPolicyDNSRestrictLab) Hints() []string {
	return []string{
		"Allow egress to kube-system namespace for DNS",
		"Use namespaceSelector for kube-system",
		"Restrict DNS port 53 to specific destination",
	}
}

func (l *CKSNetworkPolicyDNSRestrictLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyDNSRestrictLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: dns-controlled
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: dns-tester
  namespace: dns-controlled
spec:
  containers:
  - name: tester
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSNetworkPolicyDNSRestrictLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "dns-controlled", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "allow-dns-to-kube-system") {
		return fmt.Errorf("DNS restriction policy not found")
	}
	return nil
}

func (l *CKSNetworkPolicyDNSRestrictLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create DNS restriction policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns-to-kube-system
  namespace: dns-controlled
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
EOF`},
		{Description: "Verify", Command: "kubectl get networkpolicies -n dns-controlled"},
	}
}
