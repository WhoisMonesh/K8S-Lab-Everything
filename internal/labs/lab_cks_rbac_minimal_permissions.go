package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSRBACMinimalPermissionsLab{})
}

type CKSRBACMinimalPermissionsLab struct {
	BaseLab
}

func (l *CKSRBACMinimalPermissionsLab) ID() string             { return "cks_rbac_minimal_permissions" }
func (l *CKSRBACMinimalPermissionsLab) Title() string          { return "Implement Least-Privilege RBAC" }
func (l *CKSRBACMinimalPermissionsLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSRBACMinimalPermissionsLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSRBACMinimalPermissionsLab) EstimatedTime() int     { return 25 }
func (l *CKSRBACMinimalPermissionsLab) Cert() Cert             { return CertCKS }
func (l *CKSRBACMinimalPermissionsLab) DomainWeight() int      { return 15 }
func (l *CKSRBACMinimalPermissionsLab) Tags() []string {
	return []string{"cks", "rbac", "least-privilege", "security"}
}

func (l *CKSRBACMinimalPermissionsLab) Description() string {
	return `A ServiceAccount 'deployer' in namespace 'app-team' has been granted
cluster-admin ClusterRoleBinding, giving it full cluster access.

Your task: Remove the cluster-admin ClusterRoleBinding and create a RoleBinding
that only allows the 'deployer' ServiceAccount to manage Deployments in
the 'app-team' namespace.`
}

func (l *CKSRBACMinimalPermissionsLab) Hints() []string {
	return []string{
		"Check existing ClusterRoleBindings with kubectl get clusterrolebindings",
		"Use RoleBinding instead of ClusterRoleBinding for namespace-scoped access",
		"Create a Role with only deployments permissions",
	}
}

func (l *CKSRBACMinimalPermissionsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRBACMinimalPermissionsLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: app-team
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	sa := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: deployer
  namespace: app-team
`
	if err := kubectlApply(ctx, kubeconfigPath, sa); err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}

	crb := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: deployer-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: deployer
  namespace: app-team
`
	return kubectlApply(ctx, kubeconfigPath, crb)
}

func (l *CKSRBACMinimalPermissionsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "clusterrolebindings", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get clusterrolebindings: %w", err)
	}
	if strings.Contains(output, "deployer-admin") {
		return fmt.Errorf("cluster-admin binding still exists for deployer")
	}

	rbOutput, err := kubectl(ctx, kubeconfigPath, "get", "rolebindings", "-n", "app-team", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get rolebindings: %w", err)
	}
	if !strings.Contains(rbOutput, "deployer") {
		return fmt.Errorf("rolebinding for deployer not created")
	}
	return nil
}

func (l *CKSRBACMinimalPermissionsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Remove cluster-admin binding", Command: "kubectl delete clusterrolebinding deployer-admin"},
		{Description: "Create deployment Role", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: deployment-manager
  namespace: app-team
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
EOF`},
		{Description: "Create RoleBinding", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: deployer-binding
  namespace: app-team
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: deployment-manager
subjects:
- kind: ServiceAccount
  name: deployer
  namespace: app-team
EOF`},
		{Description: "Verify", Command: "kubectl get rolebindings -n app-team"},
	}
}
