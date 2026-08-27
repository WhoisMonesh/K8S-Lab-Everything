package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyEgressAllowSpecificLab{})
}

type NetworkPolicyEgressAllowSpecificLab struct {
	BaseLab
}

func (l *NetworkPolicyEgressAllowSpecificLab) ID() string {
	return "network_policy_egress_allow_specific"
}

func (l *NetworkPolicyEgressAllowSpecificLab) Title() string {
	return "Egress Policy Too Restrictive"
}

func (l *NetworkPolicyEgressAllowSpecificLab) Category() Category {
	return CategoryNetworking
}

func (l *NetworkPolicyEgressAllowSpecificLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *NetworkPolicyEgressAllowSpecificLab) Description() string {
	return `A NetworkPolicy 'restrict-egress' blocks all egress traffic except to
a specific IP range (10.0.0.0/24). Pods need to reach external services
like DNS (kube-dns) and API endpoints.

Your task: Add egress rules to allow DNS resolution and traffic to
the internet.`
}

func (l *NetworkPolicyEgressAllowSpecificLab) Hints() []string {
	return []string{
		"Check the NetworkPolicy egress rules",
		"DNS traffic on port 53 needs to be allowed",
		"Allow traffic to kube-system namespace for DNS",
	}
}

func (l *NetworkPolicyEgressAllowSpecificLab) EstimatedTime() int {
	return 20
}

func (l *NetworkPolicyEgressAllowSpecificLab) Tags() []string {
	return []string{"network-policy", "egress", "dns", "networking"}
}

func (l *NetworkPolicyEgressAllowSpecificLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyEgressAllowSpecificLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: restricted
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	networkPolicy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: restrict-egress
  namespace: restricted
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to:
    - ipBlock:
        cidr: 10.0.0.0/24
`
	if err := kubectlApply(ctx, kubeconfigPath, networkPolicy); err != nil {
		return fmt.Errorf("creating network policy: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: restricted
  labels:
    app: test
spec:
  containers:
  - name: test
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do sleep 3600; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *NetworkPolicyEgressAllowSpecificLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *NetworkPolicyEgressAllowSpecificLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check DNS resolution works
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "restricted", "test-pod",
		"--", "nslookup", "kubernetes.default.svc.cluster.local")
	if err != nil {
		return fmt.Errorf("DNS resolution failed: %w", err)
	}

	if !strings.Contains(output, "Address") {
		return fmt.Errorf("DNS not resolving")
	}

	return nil
}

func (l *NetworkPolicyEgressAllowSpecificLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check NetworkPolicy egress rules",
			Command:     "kubectl get networkpolicy restrict-egress -n restricted -o yaml",
			Notes:       "Only allows egress to 10.0.0.0/24, no DNS",
		},
		{
			Description: "Test DNS (should fail)",
			Command:     "kubectl exec -n restricted test-pod -- nslookup kubernetes.default.svc.cluster.local",
			Notes:       "DNS resolution should fail",
		},
		{
			Description: "Add DNS egress rule",
			Command:     "kubectl edit networkpolicy restrict-egress -n restricted",
			Notes:       "Add egress rule for UDP/TCP port 53 to kube-system namespace",
		},
		{
			Description: "Verify DNS works",
			Command:     "kubectl exec -n restricted test-pod -- nslookup kubernetes.default.svc.cluster.local",
			Notes:       "Should now resolve successfully",
		},
	}
}
