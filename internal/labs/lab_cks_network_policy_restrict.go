package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyRestrictLab{})
}

type CKSNetworkPolicyRestrictLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyRestrictLab) ID() string { return "cks_network_policy_restrict" }
func (l *CKSNetworkPolicyRestrictLab) Title() string {
	return "Restrict Cluster Access with NetworkPolicy"
}
func (l *CKSNetworkPolicyRestrictLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSNetworkPolicyRestrictLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkPolicyRestrictLab) EstimatedTime() int     { return 20 }
func (l *CKSNetworkPolicyRestrictLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyRestrictLab) DomainWeight() int      { return 15 }
func (l *CKSNetworkPolicyRestrictLab) Tags() []string {
	return []string{"cks", "network-policy", "cluster-setup", "security"}
}

func (l *CKSNetworkPolicyRestrictLab) Description() string {
	return `A multi-tier application in the 'secure-app' namespace has no network segmentation.
All pods can communicate with each other and external endpoints without restriction.

Your task: Create a NetworkPolicy that restricts pod-to-pod communication so that
only pods with label 'access=granted' can reach the backend pods on port 8443.`
}

func (l *CKSNetworkPolicyRestrictLab) Hints() []string {
	return []string{
		"Create a default-deny NetworkPolicy first",
		"Then create an allow policy for specific pods",
		"Use podSelector to match target pods",
	}
}

func (l *CKSNetworkPolicyRestrictLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyRestrictLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: secure-app
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: secure-app
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
        - containerPort: 8443
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating backend: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: attacker
  namespace: secure-app
  labels:
    app: attacker
spec:
  containers:
  - name: attacker
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSNetworkPolicyRestrictLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "secure-app", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to list network policies: %w", err)
	}
	if !strings.Contains(output, "default-deny") {
		return fmt.Errorf("default-deny policy not created")
	}
	if !strings.Contains(output, "allow-granted") {
		return fmt.Errorf("allow policy for access=granted pods not created")
	}
	return nil
}

func (l *CKSNetworkPolicyRestrictLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create default-deny policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
  namespace: secure-app
spec:
  podSelector: {}
  policyTypes:
  - Ingress
EOF`},
		{Description: "Create allow policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-granted
  namespace: secure-app
spec:
  podSelector:
    matchLabels:
      app: backend
  ingress:
  - from:
    - podSelector:
        matchLabels:
          access: granted
    ports:
    - protocol: TCP
      port: 8443
EOF`},
		{Description: "Verify policies exist", Command: "kubectl get networkpolicies -n secure-app"},
	}
}
