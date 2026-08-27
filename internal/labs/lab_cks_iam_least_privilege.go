package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSIAMLeastPrivilegeLab{})
}

type CKSIAMLeastPrivilegeLab struct {
	BaseLab
}

func (l *CKSIAMLeastPrivilegeLab) ID() string             { return "cks_iam_least_privilege" }
func (l *CKSIAMLeastPrivilegeLab) Title() string          { return "Implement Least-Privilege IAM" }
func (l *CKSIAMLeastPrivilegeLab) Category() Category     { return CategorySystemHardening }
func (l *CKSIAMLeastPrivilegeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSIAMLeastPrivilegeLab) EstimatedTime() int     { return 20 }
func (l *CKSIAMLeastPrivilegeLab) Cert() Cert             { return CertCKS }
func (l *CKSIAMLeastPrivilegeLab) DomainWeight() int      { return 10 }
func (l *CKSIAMLeastPrivilegeLab) Tags() []string {
	return []string{"cks", "iam", "least-privilege", "rbac"}
}

func (l *CKSIAMLeastPrivilegeLab) Description() string {
	return `A ClusterRole 'full-access' grants wildcard permissions on all resources
in the cluster. This violates least-privilege principles.

Your task: Replace the ClusterRole 'full-access' with a ClusterRole that only
allows read-only access to pods, services, and configmaps across all namespaces.`
}

func (l *CKSIAMLeastPrivilegeLab) Hints() []string {
	return []string{
		"Check the current ClusterRole definition",
		"Replace wildcard resources with specific resources",
		"Use only get, list, watch verbs",
	}
}

func (l *CKSIAMLeastPrivilegeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSIAMLeastPrivilegeLab) Break(ctx context.Context, kubeconfigPath string) error {
	cr := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: full-access
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
`
	return kubectlApply(ctx, kubeconfigPath, cr)
}

func (l *CKSIAMLeastPrivilegeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "clusterrole", "full-access", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get clusterrole: %w", err)
	}
	if strings.Contains(output, "\"*\"") || strings.Contains(output, "verbs: [\"*\"]") {
		return fmt.Errorf("wildcard permissions still present")
	}
	if !strings.Contains(output, "pods") || !strings.Contains(output, "services") {
		return fmt.Errorf("specific resources not defined")
	}
	return nil
}

func (l *CKSIAMLeastPrivilegeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current ClusterRole", Command: "kubectl get clusterrole full-access -o yaml"},
		{Description: "Replace with least-privilege Role", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: full-access
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch"]
EOF`},
		{Description: "Verify", Command: "kubectl describe clusterrole full-access"},
	}
}
