package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NetworkPolicyIngressBlockLab{})
}

type NetworkPolicyIngressBlockLab struct {
	BaseLab
}

func (l *NetworkPolicyIngressBlockLab) ID() string {
	return "network_policy_ingress_block"
}

func (l *NetworkPolicyIngressBlockLab) Title() string {
	return "Ingress NetworkPolicy Too Restrictive"
}

func (l *NetworkPolicyIngressBlockLab) Category() Category {
	return CategoryNetworking
}

func (l *NetworkPolicyIngressBlockLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NetworkPolicyIngressBlockLab) Description() string {
	return `A NetworkPolicy 'restrict-access' is blocking all ingress traffic to
pods with label app=secure-api. The policy allows only pods with
label env=production but the client pods have env=staging.

Your task: Update the NetworkPolicy to allow traffic from staging pods.`
}

func (l *NetworkPolicyIngressBlockLab) Hints() []string {
	return []string{
		"Check the NetworkPolicy label selectors",
		"The policy allows only env=production but clients have env=staging",
		"Update the policy to include staging environment",
	}
}

func (l *NetworkPolicyIngressBlockLab) EstimatedTime() int {
	return 15
}

func (l *NetworkPolicyIngressBlockLab) Tags() []string {
	return []string{"network-policy", "ingress", "labels", "networking"}
}

func (l *NetworkPolicyIngressBlockLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyIngressBlockLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: api-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	secureApi := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: secure-api
  namespace: api-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: secure-api
  template:
    metadata:
      labels:
        app: secure-api
    spec:
      containers:
      - name: api
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: secure-api
  namespace: api-ns
spec:
  selector:
    app: secure-api
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, secureApi); err != nil {
		return fmt.Errorf("creating secure-api: %w", err)
	}

	client := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: staging-client
  namespace: api-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: staging-client
  template:
    metadata:
      labels:
        app: staging-client
        env: staging
    spec:
      containers:
      - name: client
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do sleep 3600; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, client); err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	networkPolicy := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: restrict-access
  namespace: api-ns
spec:
  podSelector:
    matchLabels:
      app: secure-api
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          env: production
    ports:
    - protocol: TCP
      port: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, networkPolicy); err != nil {
		return fmt.Errorf("creating network policy: %w", err)
	}

	return nil
}

func (l *NetworkPolicyIngressBlockLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *NetworkPolicyIngressBlockLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicy", "restrict-access",
		"-n", "api-ns", "-o", "jsonpath={.spec.ingress[0].from[0].podSelector.matchLabels.env}")
	if err != nil {
		return fmt.Errorf("failed to check network policy: %w", err)
	}

	if strings.TrimSpace(output) == "production" {
		return fmt.Errorf("network policy still only allows production")
	}

	// Test connectivity
	clientPod, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "api-ns",
		"-l", "app=staging-client", "-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return fmt.Errorf("failed to get client pod: %w", err)
	}

	clientPod = strings.TrimSpace(clientPod)
	_, err = kubectl(ctx, kubeconfigPath, "exec", "-n", "api-ns", clientPod,
		"--", "wget", "-O-", "--timeout=3", "-q", "http://secure-api")
	if err != nil {
		return fmt.Errorf("client cannot reach secure-api: %w", err)
	}

	return nil
}

func (l *NetworkPolicyIngressBlockLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check NetworkPolicy",
			Command:     "kubectl get networkpolicy restrict-access -n api-ns -o yaml",
			Notes:       "Ingress rule only allows env=production",
		},
		{
			Description: "Check client pod labels",
			Command:     "kubectl get pods -n api-ns --show-labels",
			Notes:       "Client has env=staging, not env=production",
		},
		{
			Description: "Fix NetworkPolicy",
			Command:     "kubectl edit networkpolicy restrict-access -n api-ns",
			Notes:       "Change env=production to env=staging in the ingress rule",
		},
		{
			Description: "Verify connectivity",
			Command:     "kubectl exec -n api-ns deploy/staging-client -- wget -O- --timeout=3 http://secure-api",
			Notes:       "Should now return nginx welcome page",
		},
	}
}
