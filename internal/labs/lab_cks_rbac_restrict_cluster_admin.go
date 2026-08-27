package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSRBACRestrictClusterAdminLab{})
}

type CKSRBACRestrictClusterAdminLab struct {
	BaseLab
}

func (l *CKSRBACRestrictClusterAdminLab) ID() string             { return "cks_rbac_restrict_cluster_admin" }
func (l *CKSRBACRestrictClusterAdminLab) Title() string          { return "Restrict cluster-admin Access" }
func (l *CKSRBACRestrictClusterAdminLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSRBACRestrictClusterAdminLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSRBACRestrictClusterAdminLab) EstimatedTime() int     { return 25 }
func (l *CKSRBACRestrictClusterAdminLab) Cert() Cert             { return CertCKS }
func (l *CKSRBACRestrictClusterAdminLab) DomainWeight() int      { return 15 }
func (l *CKSRBACRestrictClusterAdminLab) Tags() []string {
	return []string{"cks", "rbac", "cluster-admin", "security"}
}

func (l *CKSRBACRestrictClusterAdminLab) Description() string {
	return `A user 'dev-user' has been granted cluster-admin via a ClusterRoleBinding.
This user should only have read-only access to pods and services in the
'development' namespace.

Your task: Remove the cluster-admin binding for dev-user and create a proper
RoleBinding with read-only namespace access.`
}

func (l *CKSRBACRestrictClusterAdminLab) Hints() []string {
	return []string{
		"Check ClusterRoleBindings for dev-user",
		"Delete the ClusterRoleBinding granting cluster-admin",
		"Create a RoleBinding with a read-only Role in the namespace",
	}
}

func (l *CKSRBACRestrictClusterAdminLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRBACRestrictClusterAdminLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: development
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	crb := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: dev-user-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: User
  name: dev-user
  apiGroup: rbac.authorization.k8s.io
`
	return kubectlApply(ctx, kubeconfigPath, crb)
}

func (l *CKSRBACRestrictClusterAdminLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "clusterrolebindings", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get clusterrolebindings: %w", err)
	}
	if strings.Contains(output, "dev-user-admin") {
		return fmt.Errorf("cluster-admin binding for dev-user still exists")
	}

	rbOutput, err := kubectl(ctx, kubeconfigPath, "get", "rolebindings", "-n", "development", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get rolebindings: %w", err)
	}
	if !strings.Contains(rbOutput, "dev-user") {
		return fmt.Errorf("rolebinding for dev-user not created in development namespace")
	}
	return nil
}

func (l *CKSRBACRestrictClusterAdminLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Delete cluster-admin binding", Command: "kubectl delete clusterrolebinding dev-user-admin"},
		{Description: "Create RoleBinding with view role", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: dev-user-view
  namespace: development
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
subjects:
- kind: User
  name: dev-user
  apiGroup: rbac.authorization.k8s.io
EOF`},
		{Description: "Verify", Command: "kubectl get rolebindings -n development"},
	}
}
