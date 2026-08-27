package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyDefaultDenyLab{})
}

type NetworkPolicyDefaultDenyLab struct {
	BaseLab
}

func (l *NetworkPolicyDefaultDenyLab) ID() string {
	return "network_policy_default_deny"
}

func (l *NetworkPolicyDefaultDenyLab) Title() string {
	return "Default Deny Blocking All Traffic"
}

func (l *NetworkPolicyDefaultDenyLab) Category() Category {
	return CategoryNetworking
}

func (l *NetworkPolicyDefaultDenyLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NetworkPolicyDefaultDenyLab) Description() string {
	return `A NetworkPolicy in the 'isolated' namespace blocks all ingress traffic
by default. This prevents pods from communicating with each other even
within the same namespace.

Your task: Create a NetworkPolicy that allows pods with label 'role=frontend'
to communicate with pods labeled 'role=backend' on port 8080.`
}

func (l *NetworkPolicyDefaultDenyLab) Hints() []string {
	return []string{
		"Check for NetworkPolicies in the namespace",
		"A default-deny policy blocks all traffic unless explicitly allowed",
		"Create an additional policy to allow specific traffic flows",
	}
}

func (l *NetworkPolicyDefaultDenyLab) EstimatedTime() int {
	return 20
}

func (l *NetworkPolicyDefaultDenyLab) Tags() []string {
	return []string{"network-policy", "default-deny", "networking"}
}

func (l *NetworkPolicyDefaultDenyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyDefaultDenyLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: isolated
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	defaultDeny := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: isolated
spec:
  podSelector: {}
  policyTypes:
  - Ingress
`
	if err := kubectlApply(ctx, kubeconfigPath, defaultDeny); err != nil {
		return fmt.Errorf("creating default-deny policy: %w", err)
	}

	backend := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: isolated
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
      role: backend
  template:
    metadata:
      labels:
        app: backend
        role: backend
    spec:
      containers:
      - name: backend
        image: nginx:alpine
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: isolated
spec:
  selector:
    app: backend
    role: backend
  ports:
  - port: 8080
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, backend); err != nil {
		return fmt.Errorf("creating backend: %w", err)
	}

	frontend := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: isolated
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
      role: frontend
  template:
    metadata:
      labels:
        app: frontend
        role: frontend
    spec:
      containers:
      - name: frontend
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do sleep 3600; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, frontend); err != nil {
		return fmt.Errorf("creating frontend: %w", err)
	}

	return nil
}

func (l *NetworkPolicyDefaultDenyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *NetworkPolicyDefaultDenyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "isolated",
		"-o", "name")
	if err != nil {
		return fmt.Errorf("failed to check network policies: %w", err)
	}

	if !strings.Contains(output, "allow-frontend-to-backend") {
		return fmt.Errorf("allow policy not created")
	}

	// Test connectivity
	frontendPod, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "isolated",
		"-l", "role=frontend", "-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return fmt.Errorf("failed to get frontend pod: %w", err)
	}

	frontendPod = strings.TrimSpace(frontendPod)
	_, err = kubectl(ctx, kubeconfigPath, "exec", "-n", "isolated", frontendPod,
		"--", "wget", "-O-", "--timeout=3", "-q", "http://backend:8080")
	if err != nil {
		return fmt.Errorf("frontend cannot reach backend: %w", err)
	}

	return nil
}

func (l *NetworkPolicyDefaultDenyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check existing NetworkPolicies",
			Command:     "kubectl get networkpolicies -n isolated",
			Notes:       "default-deny-ingress blocks all ingress traffic",
		},
		{
			Description: "Test connectivity (should fail)",
			Command:     "kubectl exec -n isolated deploy/frontend -- wget -O- --timeout=2 http://backend:8080",
			Notes:       "Connection should be refused",
		},
		{
			Description: "Create allow policy",
			Command:     "cat <<EOF | kubectl apply -f -\napiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n  name: allow-frontend-to-backend\n  namespace: isolated\nspec:\n  podSelector:\n    matchLabels:\n      role: backend\n  ingress:\n  - from:\n    - podSelector:\n        matchLabels:\n          role: frontend\n    ports:\n    - protocol: TCP\n      port: 8080\nEOF",
			Notes:       "Allow ingress from frontend to backend on port 8080",
		},
		{
			Description: "Verify connectivity",
			Command:     "kubectl exec -n isolated deploy/frontend -- wget -O- --timeout=3 http://backend:8080",
			Notes:       "Should now return nginx welcome page",
		},
	}
}
