package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkPolicyAllowSpecificLab{})
}

type CKSNetworkPolicyAllowSpecificLab struct {
	BaseLab
}

func (l *CKSNetworkPolicyAllowSpecificLab) ID() string             { return "cks_network_policy_allow_specific" }
func (l *CKSNetworkPolicyAllowSpecificLab) Title() string          { return "Allow Specific Pod-to-Pod Traffic" }
func (l *CKSNetworkPolicyAllowSpecificLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSNetworkPolicyAllowSpecificLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNetworkPolicyAllowSpecificLab) EstimatedTime() int     { return 25 }
func (l *CKSNetworkPolicyAllowSpecificLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkPolicyAllowSpecificLab) DomainWeight() int      { return 20 }
func (l *CKSNetworkPolicyAllowSpecificLab) Tags() []string {
	return []string{"cks", "network-policy", "allow-traffic", "microservice-vulns"}
}

func (l *CKSNetworkPolicyAllowSpecificLab) Description() string {
	return `A default-deny policy in namespace 'microservices' blocks all traffic.
The 'frontend' pod needs to communicate with 'backend' on port 8080,
and 'backend' needs to communicate with 'database' on port 3306.

Your task: Create NetworkPolicies that allow:
1. frontend -> backend on port 8080
2. backend -> database on port 3306`
}

func (l *CKSNetworkPolicyAllowSpecificLab) Hints() []string {
	return []string{
		"Use podSelector with matchLabels for source pods",
		"Use to/podSelector for destination pods",
		"Specify port numbers in the ports section",
	}
}

func (l *CKSNetworkPolicyAllowSpecificLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkPolicyAllowSpecificLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSNetworkPolicyAllowSpecificLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "microservices", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "frontend-to-backend") {
		return fmt.Errorf("frontend-to-backend policy not created")
	}
	if !strings.Contains(output, "backend-to-database") {
		return fmt.Errorf("backend-to-database policy not created")
	}
	return nil
}

func (l *CKSNetworkPolicyAllowSpecificLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create frontend to backend policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: frontend-to-backend
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
    - protocol: TCP
      port: 8080
EOF`},
		{Description: "Create backend to database policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-to-database
  namespace: microservices
spec:
  podSelector:
    matchLabels:
      app: database
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 3306
EOF`},
		{Description: "Verify", Command: "kubectl get networkpolicies -n microservices"},
	}
}
