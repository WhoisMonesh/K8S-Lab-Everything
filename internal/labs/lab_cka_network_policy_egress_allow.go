package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyEgressAllowLab{})
}

type NetworkPolicyEgressAllowLab struct {
	BaseLab
}

func (l *NetworkPolicyEgressAllowLab) ID() string { return "cka_network_policy_egress_allow" }
func (l *NetworkPolicyEgressAllowLab) Title() string {
	return "Allow Specific Egress Traffic"
}
func (l *NetworkPolicyEgressAllowLab) Category() Category     { return CategoryServicesNetworking }
func (l *NetworkPolicyEgressAllowLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NetworkPolicyEgressAllowLab) EstimatedTime() int     { return 20 }
func (l *NetworkPolicyEgressAllowLab) Tags() []string {
	return []string{"network-policy", "egress", "networking"}
}
func (l *NetworkPolicyEgressAllowLab) Cert() Cert        { return CertCKA }
func (l *NetworkPolicyEgressAllowLab) DomainWeight() int { return 20 }

func (l *NetworkPolicyEgressAllowLab) Description() string {
	return `A NetworkPolicy is blocking all egress traffic from pods. Create an
additional NetworkPolicy that allows egress to DNS (port 53) and to a
specific backend service on port 8080.`
}

func (l *NetworkPolicyEgressAllowLab) Hints() []string {
	return []string{
		"Check existing NetworkPolicies",
		"Allow egress to kube-system namespace for DNS",
		"Allow egress to backend pods on port 8080",
	}
}

func (l *NetworkPolicyEgressAllowLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyEgressAllowLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: egress-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	frontDep := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: egress-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: nginx:alpine
`
	if err := kubectlApply(ctx, kubeconfigPath, frontDep); err != nil {
		return fmt.Errorf("creating frontend deployment: %w", err)
	}

	backDep := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: egress-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: nginx:alpine
        ports:
        - containerPort: 8080
`
	if err := kubectlApply(ctx, kubeconfigPath, backDep); err != nil {
		return fmt.Errorf("creating backend deployment: %w", err)
	}

	policy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-egress
  namespace: egress-ns
spec:
  podSelector:
    matchLabels:
      app: frontend
  policyTypes:
  - Egress
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: never-match
`
	if err := kubectlApply(ctx, kubeconfigPath, policy); err != nil {
		return fmt.Errorf("creating networkpolicy: %w", err)
	}

	return nil
}

func (l *NetworkPolicyEgressAllowLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *NetworkPolicyEgressAllowLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "egress-ns",
		"-o", "name")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "allow-dns-egress") {
		return fmt.Errorf("DNS egress policy not created")
	}
	return nil
}

func (l *NetworkPolicyEgressAllowLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check existing policies", Command: "kubectl get networkpolicies -n egress-ns"},
		{Description: "Allow DNS egress", Command: "cat <<EOF | kubectl apply -f -\napiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n  name: allow-dns-egress\n  namespace: egress-ns\nspec:\n  podSelector: {}\n  policyTypes:\n  - Egress\n  egress:\n  - to:\n    - namespaceSelector:\n        matchLabels:\n          kubernetes.io/metadata.name: kube-system\n    ports:\n    - protocol: UDP\n      port: 53\n    - protocol: TCP\n      port: 53\nEOF"},
		{Description: "Allow backend egress", Command: "cat <<EOF | kubectl apply -f -\napiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n  name: allow-backend-egress\n  namespace: egress-ns\nspec:\n  podSelector:\n    matchLabels:\n      app: frontend\n  policyTypes:\n  - Egress\n  egress:\n  - to:\n    - podSelector:\n        matchLabels:\n          app: backend\n    ports:\n    - protocol: TCP\n      port: 8080\nEOF"},
	}
}
