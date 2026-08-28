package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodeNetworkPolicyLab{})
}

type MultiNodeNetworkPolicyLab struct {
	BaseLab
}

func (l *MultiNodeNetworkPolicyLab) ID() string {
	return "multi_node_network_policy"
}

func (l *MultiNodeNetworkPolicyLab) Title() string {
	return "Network Policy Between Nodes"
}

func (l *MultiNodeNetworkPolicyLab) Category() Category {
	return CategoryServicesNetworking
}

func (l *MultiNodeNetworkPolicyLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *MultiNodeNetworkPolicyLab) Description() string {
	return `A frontend and backend application are deployed across multiple worker
nodes. A NetworkPolicy is blocking all traffic, including the required
communication between frontend and backend.

Your task: Fix the NetworkPolicy to allow frontend-to-backend communication
while maintaining isolation from other namespaces.`
}

func (l *MultiNodeNetworkPolicyLab) Hints() []string {
	return []string{
		"Check the current NetworkPolicy rules",
		"A default deny-all policy blocks everything",
		"You need an additional policy to allow specific ingress",
	}
}

func (l *MultiNodeNetworkPolicyLab) EstimatedTime() int {
	return 25
}

func (l *MultiNodeNetworkPolicyLab) Tags() []string {
	return []string{"network-policy", "multi-node", "networking", "isolation", "microservices"}
}

func (l *MultiNodeNetworkPolicyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: microservices
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return err
	}

	backend := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: microservices
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo backend; sleep 15; done']
        ports:
        - containerPort: 8080
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	if err := kubectlApply(ctx, kubeconfigPath, backend); err != nil {
		return err
	}

	frontend := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: microservices
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do wget -qO- http://backend:8080 || echo failed; sleep 15; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, frontend)
}

func (l *MultiNodeNetworkPolicyLab) Break(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	policy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: microservices
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
`
	return kubectlApply(ctx, kubeconfigPath, policy)
}

func (l *MultiNodeNetworkPolicyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "microservices")
	if err != nil {
		return fmt.Errorf("checking policies: %w", err)
	}

	if strings.Contains(output, "deny-all") {
		return nil
	}

	return fmt.Errorf("deny-all policy not found")
}

func (l *MultiNodeNetworkPolicyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	policies, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "microservices", "-o", "name")
	if err != nil {
		return fmt.Errorf("checking policies: %w", err)
	}

	if strings.Contains(policies, "allow-frontend-to-backend") {
		return nil
	}

	return fmt.Errorf("allow-frontend-to-backend policy not created")
}

func (l *MultiNodeNetworkPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check existing policies",
			Command:     "kubectl get networkpolicies -n microservices",
			Notes:       "Notice the deny-all policy",
		},
		{
			Description: "Fix: Create allow policy for frontend to backend",
			Command: `kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend-to-backend
  namespace: microservices
spec:
  podSelector:
    matchLabels:
      app: backend
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - port: 8080
EOF`,
			Notes: "Allow frontend pods to reach backend on port 8080",
		},
		{
			Description: "Verify connectivity",
			Command:     "kubectl exec deploy/frontend -n microservices -- wget -qO- http://backend:8080",
			Notes:       "Should get a response from backend",
		},
	}
}
