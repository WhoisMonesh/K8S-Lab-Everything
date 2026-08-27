package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyLabelsWrongLab{})
}

type NetworkPolicyLabelsWrongLab struct {
	BaseLab
}

func (l *NetworkPolicyLabelsWrongLab) ID() string {
	return "network_policy_labels_wrong"
}

func (l *NetworkPolicyLabelsWrongLab) Title() string {
	return "NetworkPolicy Label Selector Mismatch"
}

func (l *NetworkPolicyLabelsWrongLab) Category() Category {
	return CategoryNetworking
}

func (l *NetworkPolicyLabelsWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NetworkPolicyLabelsWrongLab) Description() string {
	return `A NetworkPolicy 'allow-web' is supposed to allow ingress traffic to
pods with label tier=web but the label selector uses tier=frontend.
Traffic is being blocked because the selector doesn't match.

Your task: Fix the NetworkPolicy label selector to match the correct pods.`
}

func (l *NetworkPolicyLabelsWrongLab) Hints() []string {
	return []string{
		"Check the NetworkPolicy podSelector",
		"Compare the selector labels with actual pod labels",
		"Update the selector to match the correct label",
	}
}

func (l *NetworkPolicyLabelsWrongLab) EstimatedTime() int {
	return 10
}

func (l *NetworkPolicyLabelsWrongLab) Tags() []string {
	return []string{"network-policy", "labels", "selectors", "networking"}
}

func (l *NetworkPolicyLabelsWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyLabelsWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: web-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	web := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: web-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-app
      tier: web
  template:
    metadata:
      labels:
        app: web-app
        tier: web
    spec:
      containers:
      - name: web
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, web); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	networkPolicy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-web
  namespace: web-ns
spec:
  podSelector:
    matchLabels:
      tier: frontend
  policyTypes:
  - Ingress
  ingress:
  - ports:
    - protocol: TCP
      port: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, networkPolicy); err != nil {
		return fmt.Errorf("creating network policy: %w", err)
	}

	return nil
}

func (l *NetworkPolicyLabelsWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *NetworkPolicyLabelsWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicy", "allow-web",
		"-n", "web-ns", "-o", "jsonpath={.spec.podSelector.matchLabels.tier}")
	if err != nil {
		return fmt.Errorf("failed to check network policy: %w", err)
	}

	if strings.TrimSpace(output) == "frontend" {
		return fmt.Errorf("podSelector still uses tier=frontend")
	}

	return nil
}

func (l *NetworkPolicyLabelsWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check NetworkPolicy selector",
			Command:     "kubectl get networkpolicy allow-web -n web-ns -o yaml | grep -A 3 podSelector",
			Notes:       "podSelector uses tier=frontend",
		},
		{
			Description: "Check pod labels",
			Command:     "kubectl get pods -n web-ns --show-labels",
			Notes:       "Pods have tier=web, not tier=frontend",
		},
		{
			Description: "Fix NetworkPolicy",
			Command:     "kubectl edit networkpolicy allow-web -n web-ns",
			Notes:       "Change tier=frontend to tier=web in podSelector",
		},
		{
			Description: "Verify policy is applied",
			Command:     "kubectl get networkpolicy allow-web -n web-ns -o yaml | grep -A 3 podSelector",
			Notes:       "Should now show tier=web",
		},
	}
}
