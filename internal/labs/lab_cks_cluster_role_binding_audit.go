package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSClusterRoleBindingAuditLab{})
}

type CKSClusterRoleBindingAuditLab struct {
	BaseLab
}

func (l *CKSClusterRoleBindingAuditLab) ID() string             { return "cks_cluster_role_binding_audit" }
func (l *CKSClusterRoleBindingAuditLab) Title() string          { return "Audit ClusterRoleBindings" }
func (l *CKSClusterRoleBindingAuditLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSClusterRoleBindingAuditLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSClusterRoleBindingAuditLab) EstimatedTime() int     { return 20 }
func (l *CKSClusterRoleBindingAuditLab) Cert() Cert             { return CertCKS }
func (l *CKSClusterRoleBindingAuditLab) DomainWeight() int      { return 15 }
func (l *CKSClusterRoleBindingAuditLab) Tags() []string {
	return []string{"cks", "cluster-role-binding", "audit", "rbac"}
}

func (l *CKSClusterRoleBindingAuditLab) Description() string {
	return `Multiple ClusterRoleBindings grant cluster-admin access to various users
and service accounts. Some of these bindings are excessive.

Your task: Audit all ClusterRoleBindings and remove any that bind
non-system users or service accounts to the cluster-admin role.`
}

func (l *CKSClusterRoleBindingAuditLab) Hints() []string {
	return []string{
		"List all ClusterRoleBindings: kubectl get clusterrolebindings",
		"Check which bindings reference cluster-admin",
		"Remove bindings that are not required for system operation",
	}
}

func (l *CKSClusterRoleBindingAuditLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSClusterRoleBindingAuditLab) Break(ctx context.Context, kubeconfigPath string) error {
	crb1 := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: test-user-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: User
  name: test-user
  apiGroup: rbac.authorization.k8s.io
`
	crb2 := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: debug-sa-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: debug-sa
  namespace: default
`
	if err := kubectlApply(ctx, kubeconfigPath, crb1); err != nil {
		return fmt.Errorf("creating crb1: %w", err)
	}
	return kubectlApply(ctx, kubeconfigPath, crb2)
}

func (l *CKSClusterRoleBindingAuditLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "clusterrolebindings", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get clusterrolebindings: %w", err)
	}
	if strings.Contains(output, "test-user-admin") {
		return fmt.Errorf("test-user-admin binding still exists")
	}
	if strings.Contains(output, "debug-sa-admin") {
		return fmt.Errorf("debug-sa-admin binding still exists")
	}
	return nil
}

func (l *CKSClusterRoleBindingAuditLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "List all ClusterRoleBindings", Command: "kubectl get clusterrolebindings"},
		{Description: "Check bindings with cluster-admin", Command: "kubectl get clusterrolebindings -o json | jq '.items[] | select(.roleRef.name==\"cluster-admin\") | .metadata.name'"},
		{Description: "Remove excessive bindings", Command: "kubectl delete clusterrolebinding test-user-admin debug-sa-admin"},
		{Description: "Verify cleanup", Command: "kubectl get clusterrolebindings"},
	}
}
