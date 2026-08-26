package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NetworkPolicyEgressBlocked{})
}

type NetworkPolicyEgressBlocked struct {
	BaseLab
}

func (l *NetworkPolicyEgressBlocked) ID() string             { return "network_policy_egress_blocked" }
func (l *NetworkPolicyEgressBlocked) Title() string          { return "NetworkPolicy Blocks Egress" }
func (l *NetworkPolicyEgressBlocked) Category() Category     { return CategoryNetworking }
func (l *NetworkPolicyEgressBlocked) Difficulty() Difficulty { return DifficultyMedium }
func (l *NetworkPolicyEgressBlocked) EstimatedTime() int     { return 20 }
func (l *NetworkPolicyEgressBlocked) Tags() []string {
	return []string{"networking", "egress", "policy"}
}

func (l *NetworkPolicyEgressBlocked) Description() string {
	return `A pod cannot reach external services because a NetworkPolicy is blocking egress traffic.
Fix the NetworkPolicy to allow required egress traffic.`
}

func (l *NetworkPolicyEgressBlocked) Hints() []string {
	return []string{
		"Check NetworkPolicies for egress rules",
		"Look at the podSelector and egress config",
		"Add an egress rule to allow DNS and external traffic",
	}
}

func (l *NetworkPolicyEgressBlocked) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyEgressBlocked) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: egress-test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: client
  namespace: egress-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: client
  template:
    metadata:
      labels:
        app: client
    spec:
      containers:
      - name: curl
        image: curlimages/curl
        command: ["sleep", "3600"]
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-egress
  namespace: egress-test
spec:
  podSelector:
    matchLabels:
      app: client
  policyTypes:
  - Egress
  egress: []`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *NetworkPolicyEgressBlocked) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "egress-test", "deploy/client",
		"--", "curl", "-s", "--max-time", "5", "http://httpbin.org/get")
	if err != nil {
		return fmt.Errorf("egress test failed: %w", err)
	}
	if output == "" {
		return fmt.Errorf("no response from external service")
	}
	return nil
}

func (l *NetworkPolicyEgressBlocked) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check NetworkPolicy", Command: "kubectl get networkpolicy -n egress-test -o yaml"},
		{Description: "Delete blocking policy", Command: "kubectl delete networkpolicy deny-egress -n egress-test"},
		{Description: "Or add egress rule", Command: "kubectl patch networkpolicy deny-egress -n egress-test --type='json' -p='[{\"op\": \"replace\", \"path\": \"/spec/egress\", \"value\": [{\"to\": [{\"namespaceSelector\": {\"matchLabels\": {\"name\": \"kube-system\"}}}],\"ports\": [{\"protocol\": \"UDP\", \"port\": 53}]}]}]'"},
	}
}
