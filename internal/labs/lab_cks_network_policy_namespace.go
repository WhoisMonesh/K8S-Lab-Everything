package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyNamespaceLab{})
}

type CKSNetworkPolicyNamespaceLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyNamespaceLab) ID() string { return "cks_network_policy_namespace" }
func (l *CKSNetworkPolicyNamespaceLab) Title() string {
	return "Restrict Namespace-to-Namespace Traffic"
}
func (l *CKSNetworkPolicyNamespaceLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSNetworkPolicyNamespaceLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkPolicyNamespaceLab) EstimatedTime() int     { return 25 }
func (l *CKSNetworkPolicyNamespaceLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyNamespaceLab) DomainWeight() int      { return 15 }
func (l *CKSNetworkPolicyNamespaceLab) Tags() []string {
	return []string{"cks", "network-policy", "namespace-isolation", "security"}
}

func (l *CKSNetworkPolicyNamespaceLab) Description() string {
	return `Two namespaces 'team-a' and 'team-b' exist in the cluster. Pods in team-b
can access pods in team-a without restriction.

Your task: Create a NetworkPolicy in namespace 'team-a' that only allows
ingress traffic from pods in the same namespace (namespace selector).`
}

func (l *CKSNetworkPolicyNamespaceLab) Hints() []string {
	return []string{
		"Use namespaceSelector in the ingress from rule",
		"Match pods with the same namespace label",
		"The policy should deny cross-namespace traffic",
	}
}

func (l *CKSNetworkPolicyNamespaceLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyNamespaceLab) Break(ctx context.Context, kubeconfigPath string) error {
	for _, ns := range []string{"team-a", "team-b"} {
		yaml := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    name: %s
`, ns, ns)
		if err := kubectlApply(ctx, kubeconfigPath, yaml); err != nil {
			return fmt.Errorf("creating namespace %s: %w", ns, err)
		}
	}

	deployA := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-a
  namespace: team-a
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app-a
  template:
    metadata:
      labels:
        app: app-a
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployA); err != nil {
		return fmt.Errorf("creating app-a: %w", err)
	}

	deployB := `apiVersion: v1
kind: Pod
metadata:
  name: pod-b
  namespace: team-b
spec:
  containers:
  - name: pod-b
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, deployB)
}

func (l *CKSNetworkPolicyNamespaceLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "team-a", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "namespaceSelector") {
		return fmt.Errorf("namespaceSelector not found in network policy")
	}
	return nil
}

func (l *CKSNetworkPolicyNamespaceLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create namespace isolation policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-same-namespace
  namespace: team-a
spec:
  podSelector: {}
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: team-a
EOF`},
		{Description: "Verify policy", Command: "kubectl get networkpolicies -n team-a"},
	}
}
