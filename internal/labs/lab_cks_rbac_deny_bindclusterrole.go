package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSRBACDenyBindClusterRoleLab{})
}

type CKSRBACDenyBindClusterRoleLab struct {
	BaseLab
}

func (l *CKSRBACDenyBindClusterRoleLab) ID() string             { return "cks_rbac_deny_bindclusterrole" }
func (l *CKSRBACDenyBindClusterRoleLab) Title() string          { return "Deny Bind to cluster-admin" }
func (l *CKSRBACDenyBindClusterRoleLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSRBACDenyBindClusterRoleLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSRBACDenyBindClusterRoleLab) EstimatedTime() int     { return 25 }
func (l *CKSRBACDenyBindClusterRoleLab) Cert() Cert             { return CertCKS }
func (l *CKSRBACDenyBindClusterRoleLab) DomainWeight() int      { return 15 }
func (l *CKSRBACDenyBindClusterRoleLab) Tags() []string {
	return []string{"cks", "rbac", "cluster-admin", "binding"}
}

func (l *CKSRBACDenyBindClusterRoleLab) Description() string {
	return `A Role 'admin-creator' in namespace 'secure' grants bind permission on
the cluster-admin ClusterRole, allowing any user with this Role to escalate
to cluster-admin.

Your task: Remove the bind permission on cluster-admin from the admin-creator
Role and replace it with a RoleBinding that gives a more limited Role.`
}

func (l *CKSRBACDenyBindClusterRoleLab) Hints() []string {
	return []string{
		"Check the admin-creator Role rules",
		"Remove the clusterroles resource and bind verb",
		"Ensure the Role only manages namespace resources",
	}
}

func (l *CKSRBACDenyBindClusterRoleLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRBACDenyBindClusterRoleLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: secure
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	role := `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: admin-creator
  namespace: secure
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["clusterroles"]
  verbs: ["bind"]
  resourceNames: ["cluster-admin"]
`
	return kubectlApply(ctx, kubeconfigPath, role)
}

func (l *CKSRBACDenyBindClusterRoleLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "role", "admin-creator", "-n", "secure", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if strings.Contains(output, "clusterroles") {
		return fmt.Errorf("clusterroles resource still in role")
	}
	if strings.Contains(output, "bind") {
		return fmt.Errorf("bind verb still present")
	}
	return nil
}

func (l *CKSRBACDenyBindClusterRoleLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current Role rules", Command: "kubectl describe role admin-creator -n secure"},
		{Description: "Update Role to remove bind permission", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: admin-creator
  namespace: secure
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
EOF`},
		{Description: "Verify", Command: "kubectl describe role admin-creator -n secure"},
	}
}
