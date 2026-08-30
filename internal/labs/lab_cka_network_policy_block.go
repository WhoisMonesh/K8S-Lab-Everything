package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyBlockLab{})
}

type NetworkPolicyBlockLab struct {
	BaseLab
}

func (l *NetworkPolicyBlockLab) ID() string { return "cka_network_policy_block" }
func (l *NetworkPolicyBlockLab) Title() string {
	return "Debug NetworkPolicy Blocking Traffic"
}
func (l *NetworkPolicyBlockLab) Category() Category     { return CategoryTroubleshooting }
func (l *NetworkPolicyBlockLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *NetworkPolicyBlockLab) EstimatedTime() int     { return 25 }
func (l *NetworkPolicyBlockLab) Tags() []string {
	return []string{"network-policy", "blocking", "troubleshooting"}
}
func (l *NetworkPolicyBlockLab) Cert() Cert        { return CertCKA }
func (l *NetworkPolicyBlockLab) DomainWeight() int { return 30 }

func (l *NetworkPolicyBlockLab) Description() string {
	return `Pods cannot communicate due to a NetworkPolicy blocking traffic.
Identify which NetworkPolicy is causing the block and create an
additional policy to allow the required traffic.`
}

func (l *NetworkPolicyBlockLab) Hints() []string {
	return []string{
		"List all NetworkPolicies in the namespace",
		"Check policy selectors and rules",
		"Create an additional allow policy",
	}
}

func (l *NetworkPolicyBlockLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyBlockLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: blocked-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	appDep := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: blocked-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, appDep); err != nil {
		return fmt.Errorf("creating app deployment: %w", err)
	}

	dbDep := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: db
  namespace: blocked-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: db
  template:
    metadata:
      labels:
        app: db
    spec:
      containers:
      - name: db
        image: nginx:alpine
        ports:
        - containerPort: 5432
`
	if err := kubectlApply(ctx, kubeconfigPath, dbDep); err != nil {
		return fmt.Errorf("creating db deployment: %w", err)
	}

	policy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-db
  namespace: blocked-ns
spec:
  podSelector:
    matchLabels:
      app: db
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: never-match
`
	if err := kubectlApply(ctx, kubeconfigPath, policy); err != nil {
		return fmt.Errorf("creating networkpolicy: %w", err)
	}

	return nil
}

func (l *NetworkPolicyBlockLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *NetworkPolicyBlockLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "blocked-ns",
		"-o", "name")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "allow-internal") {
		return fmt.Errorf("allow policy not created")
	}
	return nil
}

func (l *NetworkPolicyBlockLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "List policies", Command: "kubectl get networkpolicies -n blocked-ns"},
		{Description: "Describe policy", Command: "kubectl describe networkpolicy -n blocked-ns"},
		{Description: "Create allow policy", Command: "Create policy allowing required traffic"},
		{Description: "Test connectivity", Command: "kubectl exec -n blocked-ns <pod> -- wget -O- --timeout=2 http://service:80"},
	}
}
