package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyLab{})
}

type NetworkPolicyLab struct {
	BaseLab
}

func (l *NetworkPolicyLab) ID() string {
	return "network_policy_blocking"
}

func (l *NetworkPolicyLab) Title() string {
	return "Network Policy Blocking Traffic"
}

func (l *NetworkPolicyLab) Category() Category {
	return CategoryNetworking
}

func (l *NetworkPolicyLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NetworkPolicyLab) Description() string {
	return `A frontend application cannot communicate with a backend service.
The pods are running, but network connectivity is blocked.

Your task: Fix the network policy to allow the frontend to communicate with the backend.`
}

func (l *NetworkPolicyLab) Hints() []string {
	return []string{
		"Check for NetworkPolicies in the namespace",
		"Review the pod labels on both frontend and backend",
		"NetworkPolicies use label selectors to allow/deny traffic",
		"Check both ingress and egress rules",
	}
}

func (l *NetworkPolicyLab) EstimatedTime() int {
	return 20
}

func (l *NetworkPolicyLab) Tags() []string {
	return []string{"networking", "network-policy", "labels", "selectors"}
}

func (l *NetworkPolicyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create a namespace for the lab
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: webapp
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	// Create backend deployment and service
	backend := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: webapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
      tier: backend
  template:
    metadata:
      labels:
        app: backend
        tier: backend
    spec:
      containers:
      - name: backend
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: webapp
spec:
  selector:
    app: backend
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, backend); err != nil {
		return fmt.Errorf("creating backend: %w", err)
	}

	// Create frontend deployment
	frontend := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: webapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
      tier: frontend
  template:
    metadata:
      labels:
        app: frontend
        tier: frontend
    spec:
      containers:
      - name: frontend
        image: busybox:1.28
        command: ['sh', '-c', 'while true; do sleep 3600; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, frontend); err != nil {
		return fmt.Errorf("creating frontend: %w", err)
	}

	// Create a restrictive network policy with wrong label selector
	networkPolicy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-policy
  namespace: webapp
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          tier: wrong-tier-label
    ports:
    - protocol: TCP
      port: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, networkPolicy); err != nil {
		return fmt.Errorf("creating network policy: %w", err)
	}

	return nil
}

func (l *NetworkPolicyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Wait for resources to be created
	time.Sleep(10 * time.Second)
	return nil
}

func (l *NetworkPolicyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if both pods are running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "webapp",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}

	if !strings.Contains(output, "Running") {
		return fmt.Errorf("not all pods are running yet")
	}

	// Get the frontend pod name
	frontendPod, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "webapp",
		"-l", "app=frontend", "-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return fmt.Errorf("failed to get frontend pod name: %w", err)
	}

	frontendPod = strings.TrimSpace(frontendPod)
	if frontendPod == "" {
		return fmt.Errorf("frontend pod not found")
	}

	// Test connectivity from frontend to backend
	// Use wget with timeout to test if backend is reachable
	output, err = kubectl(ctx, kubeconfigPath, "exec", "-n", "webapp", frontendPod,
		"--", "wget", "-O-", "--timeout=5", "-q", "http://backend")
	if err != nil {
		return fmt.Errorf("frontend cannot reach backend: %w", err)
	}

	// Check if we got nginx welcome page (or any response)
	if !strings.Contains(output, "nginx") && !strings.Contains(output, "Welcome") && len(output) < 10 {
		return fmt.Errorf("connectivity test did not return expected response")
	}

	return nil
}

func (l *NetworkPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the pods in the webapp namespace",
			Command:     "kubectl get pods -n webapp",
			Notes:       "Both frontend and backend pods should be running",
		},
		{
			Description: "Test connectivity from frontend to backend",
			Command:     "kubectl exec -n webapp -it deploy/frontend -- wget -O- --timeout=2 http://backend",
			Notes:       "This should timeout, confirming the connectivity issue",
		},
		{
			Description: "Check for NetworkPolicies",
			Command:     "kubectl get networkpolicies -n webapp",
			Notes:       "You should see a 'backend-policy' NetworkPolicy",
		},
		{
			Description: "Examine the NetworkPolicy",
			Command:     "kubectl describe networkpolicy backend-policy -n webapp",
			Notes:       "Look at the ingress rules and the pod selector",
		},
		{
			Description: "Check the frontend pod labels",
			Command:     "kubectl get pods -n webapp -l app=frontend --show-labels",
			Notes:       "Note the labels: app=frontend, tier=frontend",
		},
		{
			Description: "Identify the issue",
			Notes:       "The NetworkPolicy allows ingress from pods with 'tier=wrong-tier-label', but the frontend has 'tier=frontend'",
		},
		{
			Description: "Fix the NetworkPolicy",
			Command:     "kubectl edit networkpolicy backend-policy -n webapp",
			Notes:       "Change 'wrong-tier-label' to 'frontend' in the podSelector matchLabels",
		},
		{
			Description: "Verify connectivity is restored",
			Command:     "kubectl exec -n webapp -it deploy/frontend -- wget -O- --timeout=2 http://backend",
			Notes:       "This should now succeed and return the nginx welcome page",
		},
	}
}
