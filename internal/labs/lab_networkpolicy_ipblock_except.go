package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyIPBlockExceptLab{})
}

type NetworkPolicyIPBlockExceptLab struct {
	BaseLab
}

func (l *NetworkPolicyIPBlockExceptLab) ID() string {
	return "networkpolicy_ipblock_except"
}

func (l *NetworkPolicyIPBlockExceptLab) Title() string {
	return "NetworkPolicy with ipBlock and except"
}

func (l *NetworkPolicyIPBlockExceptLab) Category() Category {
	return CategoryServicesNetworking
}

func (l *NetworkPolicyIPBlockExceptLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *NetworkPolicyIPBlockExceptLab) Description() string {
	return `A NetworkPolicy needs to allow traffic from a CIDR range but with
specific exceptions. The current policy allows all traffic without
restrictions.

Your task: Configure the NetworkPolicy with ipBlock CIDR and except
fields to allow most traffic while blocking specific ranges.`
}

func (l *NetworkPolicyIPBlockExceptLab) Hints() []string {
	return []string{
		"Use ipBlock with cidr and except fields",
		"except specifies CIDRs to exclude from the allowed range",
		"Check the current NetworkPolicy rules",
	}
}

func (l *NetworkPolicyIPBlockExceptLab) EstimatedTime() int {
	return 20
}

func (l *NetworkPolicyIPBlockExceptLab) Tags() []string {
	return []string{"network-policy", "ipblock", "except", "cidr", "networking"}
}

func (l *NetworkPolicyIPBlockExceptLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: netpol-test
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return err
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: netpol-test
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo serving; sleep 15; done']
        ports:
        - containerPort: 8080
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *NetworkPolicyIPBlockExceptLab) Break(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	policy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all
  namespace: netpol-test
spec:
  podSelector: {}
  ingress:
  - {}
  policyTypes:
  - Ingress
`
	return kubectlApply(ctx, kubeconfigPath, policy)
}

func (l *NetworkPolicyIPBlockExceptLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicy", "allow-all",
		"-n", "netpol-test", "-o", "jsonpath={.spec.ingress[0]}")
	if err != nil {
		return nil
	}

	if strings.Contains(output, "ipBlock") {
		return fmt.Errorf("policy has ipBlock (expected empty ingress)")
	}

	return nil
}

func (l *NetworkPolicyIPBlockExceptLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicy", "allow-all",
		"-n", "netpol-test", "-o", "jsonpath={.spec.ingress[0]}")
	if err != nil {
		return fmt.Errorf("checking policy: %w", err)
	}

	if !strings.Contains(output, "ipBlock") || !strings.Contains(output, "except") {
		return fmt.Errorf("policy missing ipBlock with except")
	}

	return nil
}

func (l *NetworkPolicyIPBlockExceptLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check current NetworkPolicy",
			Command:     "kubectl get networkpolicy allow-all -n netpol-test -o yaml",
			Notes:       "Allows all ingress without restrictions",
		},
		{
			Description: "Fix: Add ipBlock with except",
			Command:     `kubectl patch networkpolicy allow-all -n netpol-test --type='merge' -p '{"spec":{"ingress":[{"from":[{"ipBlock":{"cidr":"10.0.0.0/8","except":["10.0.1.0/24","10.0.2.0/24"]}}]}]}}'`,
			Notes:       "Allow 10.0.0.0/8 but except 10.0.1.0/24 and 10.0.2.0/24",
		},
		{
			Description: "Verify ipBlock configuration",
			Command:     "kubectl get networkpolicy allow-all -n netpol-test -o yaml | grep -A 10 ipBlock",
			Notes:       "Should show cidr and except fields",
		},
	}
}
