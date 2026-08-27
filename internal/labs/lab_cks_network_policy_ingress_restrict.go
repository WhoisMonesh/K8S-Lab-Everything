package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyIngressRestrictLab{})
}

type CKSNetworkPolicyIngressRestrictLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyIngressRestrictLab) ID() string {
	return "cks_network_policy_ingress_restrict"
}
func (l *CKSNetworkPolicyIngressRestrictLab) Title() string          { return "Restrict Ingress Traffic" }
func (l *CKSNetworkPolicyIngressRestrictLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSNetworkPolicyIngressRestrictLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkPolicyIngressRestrictLab) EstimatedTime() int     { return 20 }
func (l *CKSNetworkPolicyIngressRestrictLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyIngressRestrictLab) DomainWeight() int      { return 20 }
func (l *CKSNetworkPolicyIngressRestrictLab) Tags() []string {
	return []string{"cks", "network-policy", "ingress", "microservice-vulns"}
}

func (l *CKSNetworkPolicyIngressRestrictLab) Description() string {
	return `The 'secure-backend' namespace accepts ingress traffic from all sources.
Any pod in the cluster can reach the backend pods.

Your task: Create a NetworkPolicy that only allows ingress from pods
with label 'role=frontend' on port 8443.`
}

func (l *CKSNetworkPolicyIngressRestrictLab) Hints() []string {
	return []string{
		"Use ingress.from with podSelector",
		"Match pods with role=frontend label",
		"Specify port 8443 in the ports section",
	}
}

func (l *CKSNetworkPolicyIngressRestrictLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyIngressRestrictLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: secure-backend
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	backend := `apiVersion: v1
kind: Pod
metadata:
  name: backend
  namespace: secure-backend
  labels:
    app: backend
spec:
  containers:
  - name: backend
    image: nginx:alpine
    ports:
    - containerPort: 8443
`
	if err := kubectlApply(ctx, kubeconfigPath, backend); err != nil {
		return fmt.Errorf("creating backend: %w", err)
	}

	Attacker := `apiVersion: v1
kind: Pod
metadata:
  name: attacker
  namespace: secure-backend
  labels:
    app: attacker
spec:
  containers:
  - name: attacker
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, Attacker)
}

func (l *CKSNetworkPolicyIngressRestrictLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "secure-backend", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "role: frontend") {
		return fmt.Errorf("ingress policy for frontend not found")
	}
	if !strings.Contains(output, "8443") {
		return fmt.Errorf("port 8443 not specified in policy")
	}
	return nil
}

func (l *CKSNetworkPolicyIngressRestrictLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create ingress restriction policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend-only
  namespace: secure-backend
spec:
  podSelector:
    matchLabels:
      app: backend
  ingress:
  - from:
    - podSelector:
        matchLabels:
          role: frontend
    ports:
    - protocol: TCP
      port: 8443
EOF`},
		{Description: "Verify", Command: "kubectl get networkpolicies -n secure-backend"},
	}
}
